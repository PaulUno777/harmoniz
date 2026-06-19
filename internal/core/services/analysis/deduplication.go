package analysis

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/xrash/smetrics"

	"harmoniz/internal/adapters/fs"
	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/ports"
	"harmoniz/internal/logger"
)

// copyRe matches paths where the filename ends with a copy indicator just before the extension.
// Requires a SPACE (not underscore) before "(N)" so YouTube filenames like "_(0).mp3" are
// not falsely penalized. Examples that match: "song (1).mp3", "song copy.mp3", "song copy 2.mp3".
var copyRe = regexp.MustCompile(`(?i)( copy\s*\d*| \(\d+\))\.\w+$`)

// trackNumPrefixRe strips leading track-number patterns like "01 ", "01. ", "01 - ", "1-".
var trackNumPrefixRe = regexp.MustCompile(`^\d{1,3}[\s.\-_]+`)

// trailAnnotRe strips trailing parenthesized/bracketed annotations like "(Remastered)", "[Live]", "(2020)".
var trailAnnotRe = regexp.MustCompile(`\s*[\(\[][^\)\]]{1,30}[\)\]]\s*$`)

// QualityScore rates a track for duplicate resolution. Higher = prefer to keep.
func QualityScore(t domain.Track) float64 {
	var s float64
	if t.Title != "" {
		s += 10
	}
	if t.ArtistRaw != "" {
		s += 10
	}
	if t.AlbumRaw != "" {
		s += 10
	}
	if t.Year > 0 {
		s += 5
	}
	if t.TrackNum > 0 {
		s += 5
	}
	s += float64(t.Bitrate) / 100
	s -= float64(len(t.Path)) * 0.01
	if copyRe.MatchString(t.Path) {
		s -= 5
	}
	return s
}

// normSimilar prepares a string for fuzzy comparison:
// lowercases, strips leading track numbers and trailing annotations,
// keeps only letters/digits/spaces, strips leading articles, collapses spaces.
func normSimilar(s string) string {
	s = strings.ToLower(s)
	s = trackNumPrefixRe.ReplaceAllString(s, "")
	s = trailAnnotRe.ReplaceAllString(s, "")
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	s = strings.Join(strings.Fields(b.String()), " ")
	for _, art := range []string{"the ", "a ", "an "} {
		if strings.HasPrefix(s, art) {
			s = s[len(art):]
			break
		}
	}
	return strings.TrimSpace(s)
}

// filenameStem strips extension then normalizes.
func filenameStem(filename string) string {
	return normSimilar(strings.TrimSuffix(filename, filepath.Ext(filename)))
}

// DeduplicationService finds exact and similar duplicate files.
type DeduplicationService struct {
	repo   ports.TrackRepository
	config DedupConfig
}

// NewDeduplicationService builds a dedup service with default config.
func NewDeduplicationService(repo ports.TrackRepository) *DeduplicationService {
	return NewDeduplicationServiceWithConfig(repo, DefaultDedupConfig())
}

// NewDeduplicationServiceWithConfig builds a dedup service with custom config.
func NewDeduplicationServiceWithConfig(repo ports.TrackRepository, cfg DedupConfig) *DeduplicationService {
	return &DeduplicationService{repo: repo, config: cfg}
}

// DetectDuplicates returns groups of duplicate tracks: byte-identical (kind="exact"),
// similar title tags (kind="similar_title"), and similar filename stems (kind="similar_filename").
// root restricts to tracks under that path; empty = all.
func (s *DeduplicationService) DetectDuplicates(ctx context.Context, root string) ([]domain.DuplicateGroup, error) {
	candidates, err := s.repo.GetDuplicateCandidates(ctx, root)
	if err != nil {
		return nil, fmt.Errorf("get duplicate candidates: %w", err)
	}

	logger.Log.Info("Deduplication: candidates found", "count", len(candidates), "root", root)

	exactGroups, exactIDs := s.detectExact(ctx, candidates)

	allTracks, _, err := s.repo.ListTracks(ctx, domain.TrackFilter{Root: root})
	if err != nil {
		return nil, fmt.Errorf("list tracks for similarity: %w", err)
	}

	similarGroups := s.detectSimilarGroups(ctx, allTracks, exactIDs)

	return append(exactGroups, similarGroups...), nil
}

// detectExact runs the size → partial hash → full hash funnel and returns exact-duplicate groups.
func (s *DeduplicationService) detectExact(ctx context.Context, candidates []domain.Track) ([]domain.DuplicateGroup, map[uint64]bool) {
	bySize := make(map[int64][]domain.Track)
	for _, t := range candidates {
		if t.HashPartial == "" {
			continue
		}
		bySize[t.Size] = append(bySize[t.Size], t)
	}

	maxLog := s.config.MaxSizeGroupLog
	if maxLog <= 0 {
		maxLog = DefaultDedupConfig().MaxSizeGroupLog
	}

	var groups []domain.DuplicateGroup
	exactIDs := make(map[uint64]bool)
	for size, group := range bySize {
		if len(group) > maxLog {
			logger.Log.Warn("Large size group", "size", size, "count", len(group))
		}
		byPartial := make(map[string][]domain.Track)
		for _, t := range group {
			byPartial[t.HashPartial] = append(byPartial[t.HashPartial], t)
		}
		for _, partialGroup := range byPartial {
			if len(partialGroup) < 2 {
				continue
			}
			byFull, err := s.groupByFullHash(ctx, partialGroup)
			if err != nil {
				logger.Log.Error("Group by full hash failed", "error", err)
				continue
			}
			for _, fullGroup := range byFull {
				if len(fullGroup) < 2 {
					continue
				}
				var bestID uint64
				bestScore := -999.0
				for _, t := range fullGroup {
					if sc := QualityScore(t); sc > bestScore {
						bestScore = sc
						bestID = t.ID
					}
					exactIDs[t.ID] = true
				}
				groups = append(groups, domain.DuplicateGroup{
					Kind:              domain.DuplicateKindExact,
					Tracks:            fullGroup,
					RecommendedKeepID: bestID,
				})
			}
		}
	}
	return groups, exactIDs
}

// detectSimilarGroups runs title-based and filename-based similarity passes.
// Tracks already in exact groups are excluded to avoid redundant reporting.
func (s *DeduplicationService) detectSimilarGroups(ctx context.Context, all []domain.Track, exactIDs map[uint64]bool) []domain.DuplicateGroup {
	titleEntries := make([]similarEntry, 0, len(all))
	filenameEntries := make([]similarEntry, 0, len(all))

	inTitleGroup := make(map[uint64]bool)

	for _, t := range all {
		if exactIDs[t.ID] {
			continue
		}
		if t.Title != "" {
			norm := normSimilar(t.Title)
			if len([]rune(norm)) >= s.cfg().MinSimilarKeyLen {
				titleEntries = append(titleEntries, similarEntry{track: t, norm: norm})
			}
		}
	}

	titleGroups := s.detectSimilarByKey(ctx, titleEntries, domain.DuplicateKindSimilarTitle)
	for _, g := range titleGroups {
		for _, t := range g.Tracks {
			inTitleGroup[t.ID] = true
		}
	}

	for _, t := range all {
		if exactIDs[t.ID] || inTitleGroup[t.ID] || t.Filename == "" {
			continue
		}
		norm := filenameStem(t.Filename)
		if len([]rune(norm)) >= s.cfg().MinSimilarKeyLen {
			filenameEntries = append(filenameEntries, similarEntry{track: t, norm: norm})
		}
	}

	filenameGroups := s.detectSimilarByKey(ctx, filenameEntries, domain.DuplicateKindSimilarFilename)

	return append(titleGroups, filenameGroups...)
}

// cfg returns config with zero-values replaced by defaults.
func (s *DeduplicationService) cfg() DedupConfig {
	d := DefaultDedupConfig()
	c := s.config
	if c.SimilarityThreshold <= 0 {
		c.SimilarityThreshold = d.SimilarityThreshold
	}
	if c.JaroWinklerP <= 0 {
		c.JaroWinklerP = d.JaroWinklerP
	}
	if c.JaroWinklerMaxPrefix <= 0 {
		c.JaroWinklerMaxPrefix = d.JaroWinklerMaxPrefix
	}
	if c.MaxSimilarBucketSize <= 0 {
		c.MaxSimilarBucketSize = d.MaxSimilarBucketSize
	}
	if c.MinSimilarKeyLen <= 0 {
		c.MinSimilarKeyLen = d.MinSimilarKeyLen
	}
	return c
}

type similarEntry struct {
	track domain.Track
	norm  string
}

// detectSimilarByKey groups entries whose norm keys are JW-similar, returns DuplicateGroups of given kind.
func (s *DeduplicationService) detectSimilarByKey(ctx context.Context, entries []similarEntry, kind string) []domain.DuplicateGroup {
	if len(entries) == 0 {
		return nil
	}
	cfg := s.cfg()

	// Bucket by first rune to keep O(N²) within manageable bounds.
	buckets := make(map[rune][]int)
	for i, e := range entries {
		r := []rune(e.norm)[0]
		buckets[r] = append(buckets[r], i)
	}

	// Union-find over entry indices.
	parent := make([]int, len(entries))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(x, y int) {
		px, py := find(x), find(y)
		if px != py {
			parent[px] = py
		}
	}

	for _, idxs := range buckets {
		if len(idxs) > cfg.MaxSimilarBucketSize {
			logger.Log.Warn("Similar-name bucket too large, skipping", "kind", kind, "size", len(idxs))
			continue
		}
		for i := range len(idxs) {
			for j := i + 1; j < len(idxs); j++ {
				select {
				case <-ctx.Done():
					return nil
				default:
				}
				a, b := entries[idxs[i]], entries[idxs[j]]
				score := smetrics.JaroWinkler(a.norm, b.norm, cfg.JaroWinklerP, cfg.JaroWinklerMaxPrefix)
				if score >= cfg.SimilarityThreshold {
					union(idxs[i], idxs[j])
				}
			}
		}
	}

	// Collect groups.
	groupMap := make(map[int][]domain.Track)
	for i, e := range entries {
		groupMap[find(i)] = append(groupMap[find(i)], e.track)
	}

	var groups []domain.DuplicateGroup
	for _, tracks := range groupMap {
		if len(tracks) < 2 {
			continue
		}
		var bestID uint64
		bestScore := -999.0
		for _, t := range tracks {
			if sc := QualityScore(t); sc > bestScore {
				bestScore = sc
				bestID = t.ID
			}
		}
		groups = append(groups, domain.DuplicateGroup{
			Kind:              kind,
			Tracks:            tracks,
			RecommendedKeepID: bestID,
		})
	}
	return groups
}

// groupByFullHash computes full hash for tracks that don't have it, then groups by hash.
func (s *DeduplicationService) groupByFullHash(ctx context.Context, group []domain.Track) (map[string][]domain.Track, error) {
	byFull := make(map[string][]domain.Track)
	for i := range group {
		t := &group[i]
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if t.HashFull == "" {
			full, err := fs.ComputeFullHash(t.Path)
			if err != nil {
				logger.Log.Error("Compute full hash failed", "path", t.Path, "error", err)
				continue
			}
			t.HashFull = full
		}
		byFull[t.HashFull] = append(byFull[t.HashFull], *t)
	}
	return byFull, nil
}
