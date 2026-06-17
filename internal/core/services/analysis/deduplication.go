package analysis

import (
	"context"
	"fmt"
	"regexp"

	"harmoniz/internal/adapters/fs"
	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/ports"
	"harmoniz/internal/logger"
)

// copyRe matches paths where the filename ends with a copy indicator just before the extension.
// Requires a SPACE (not underscore) before "(N)" so YouTube filenames like "_(0).mp3" are
// not falsely penalized. Examples that match: "song (1).mp3", "song copy.mp3", "song copy 2.mp3".
var copyRe = regexp.MustCompile(`(?i)( copy\s*\d*| \(\d+\))\.\w+$`)

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

// DeduplicationService finds exact duplicate files via size → partial hash → full hash funnel.
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

// DetectDuplicates returns groups of byte-identical tracks with a recommended track to keep.
// root restricts to tracks under that path; empty = all.
func (s *DeduplicationService) DetectDuplicates(ctx context.Context, root string) ([]domain.DuplicateGroup, error) {
	candidates, err := s.repo.GetDuplicateCandidates(ctx, root)
	if err != nil {
		return nil, fmt.Errorf("get duplicate candidates: %w", err)
	}

	logger.Log.Info("Deduplication: candidates found", "count", len(candidates), "root", root)

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

	var duplicates []domain.DuplicateGroup
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
				var bestScore = -999.0
				for _, t := range fullGroup {
					if sc := QualityScore(t); sc > bestScore {
						bestScore = sc
						bestID = t.ID
					}
				}
				duplicates = append(duplicates, domain.DuplicateGroup{
					Tracks:            fullGroup,
					RecommendedKeepID: bestID,
				})
			}
		}
	}
	return duplicates, nil
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
