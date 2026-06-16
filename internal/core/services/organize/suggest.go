package organize

import (
	"context"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/ports"
)

// SuggestionService infers metadata suggestions from path structure, filename patterns,
// and cross-file sibling context within the same directory.
type SuggestionService struct {
	repo ports.TrackRepository
}

func NewSuggestionService(repo ports.TrackRepository) *SuggestionService {
	return &SuggestionService{repo: repo}
}

// AnalyzeRoot analyzes all tracks under root and returns per-track suggestions.
func (s *SuggestionService) AnalyzeRoot(ctx context.Context, root string) ([]domain.OrganizerSuggestion, error) {
	tracks, _, err := s.repo.ListTracks(ctx, domain.TrackFilter{Root: root, Limit: 50000})
	if err != nil {
		return nil, err
	}

	byDir := map[string][]domain.Track{}
	for _, t := range tracks {
		dir := filepath.Dir(t.Path)
		byDir[dir] = append(byDir[dir], t)
	}

	dirInfos := s.computeDirInfos(root, byDir)

	suggestions := make([]domain.OrganizerSuggestion, 0, len(tracks))
	for _, t := range tracks {
		dir := filepath.Dir(t.Path)
		sugg := s.buildSuggestion(t, dirInfos[dir], byDir[dir])
		suggestions = append(suggestions, sugg)
	}
	return suggestions, nil
}

// dirMeta holds path-inferred and tag-voted metadata for a single directory.
type dirMeta struct {
	artistCandidate string
	albumCandidate  string
	pathConfidence  float64
	tagArtist       string  // majority-vote artist from existing sibling tags
	tagAlbum        string  // majority-vote album from existing sibling tags
	tagConfidence   float64 // confidence for tag-derived values
}

func (s *SuggestionService) computeDirInfos(root string, byDir map[string][]domain.Track) map[string]dirMeta {
	result := map[string]dirMeta{}
	cleanRoot := filepath.Clean(root)

	for dir, tracks := range byDir {
		info := dirMeta{}
		cleanDir := filepath.Clean(dir)

		rel, err := filepath.Rel(cleanRoot, cleanDir)
		if err == nil && rel != "." {
			parts := splitPath(rel)
			switch len(parts) {
			case 1:
				info.artistCandidate = parts[0]
				info.pathConfidence = 0.60
			case 2:
				info.artistCandidate = parts[0]
				info.albumCandidate = parts[1]
				info.pathConfidence = 0.72
			default:
				if len(parts) >= 2 {
					// Take the last two: assume artist/album nesting
					info.artistCandidate = parts[len(parts)-2]
					info.albumCandidate = parts[len(parts)-1]
					info.pathConfidence = 0.65
				}
			}
		}

		// Majority vote from existing sibling tags
		artistVotes := map[string]int{}
		albumVotes := map[string]int{}
		for _, t := range tracks {
			if v := strings.TrimSpace(t.ArtistRaw); v != "" {
				artistVotes[v]++
			}
			if v := strings.TrimSpace(t.AlbumRaw); v != "" {
				albumVotes[v]++
			}
		}
		total := len(tracks)

		if artist, count := topVote(artistVotes); count*2 > total {
			info.tagArtist = artist
			info.tagConfidence = 0.75 + 0.15*(float64(count)/float64(total))
		}
		if album, count := topVote(albumVotes); count*2 > total {
			info.tagAlbum = album
			if info.tagConfidence == 0 {
				info.tagConfidence = 0.75 + 0.15*(float64(count)/float64(total))
			}
		}

		result[dir] = info
	}
	return result
}

func (s *SuggestionService) buildSuggestion(t domain.Track, info dirMeta, _ []domain.Track) domain.OrganizerSuggestion {
	fields := map[string]domain.FieldSuggestion{}
	issues := []string{}

	parsed := parseFilename(t.Filename)

	// --- Artist ---
	if strings.TrimSpace(t.ArtistRaw) == "" {
		if parsed.artist != "" {
			conf := 0.70
			if info.tagArtist != "" && normStr(parsed.artist) == normStr(info.tagArtist) {
				conf = 0.88
			}
			fields["artist"] = domain.FieldSuggestion{Value: parsed.artist, Confidence: conf, Source: "filename"}
		} else if info.tagArtist != "" {
			fields["artist"] = domain.FieldSuggestion{Value: info.tagArtist, Confidence: info.tagConfidence, Source: "neighbor"}
		} else if info.artistCandidate != "" {
			fields["artist"] = domain.FieldSuggestion{Value: info.artistCandidate, Confidence: info.pathConfidence, Source: "path"}
		}
		issues = append(issues, "Missing artist")
	}

	// --- Album ---
	if strings.TrimSpace(t.AlbumRaw) == "" {
		if parsed.album != "" {
			conf := 0.65
			if info.tagAlbum != "" && normStr(parsed.album) == normStr(info.tagAlbum) {
				conf = 0.88
			}
			fields["album"] = domain.FieldSuggestion{Value: parsed.album, Confidence: conf, Source: "filename"}
		} else if info.tagAlbum != "" {
			fields["album"] = domain.FieldSuggestion{Value: info.tagAlbum, Confidence: info.tagConfidence, Source: "neighbor"}
		} else if info.albumCandidate != "" {
			fields["album"] = domain.FieldSuggestion{Value: info.albumCandidate, Confidence: info.pathConfidence, Source: "path"}
		}
		issues = append(issues, "Missing album")
	}

	// --- Title ---
	if strings.TrimSpace(t.Title) == "" {
		if parsed.title != "" {
			fields["title"] = domain.FieldSuggestion{Value: parsed.title, Confidence: 0.82, Source: "filename"}
		}
		issues = append(issues, "Missing title")
	}

	// --- Track number ---
	if t.TrackNum == 0 && parsed.trackNum > 0 {
		fields["track_num"] = domain.FieldSuggestion{
			Value:      strconv.Itoa(parsed.trackNum),
			Confidence: 0.88,
			Source:     "filename",
		}
	}

	score := computeScore(fields)

	return domain.OrganizerSuggestion{
		TrackID: t.ID,
		Track:   t,
		Score:   score,
		Fields:  fields,
		Issues:  issues,
	}
}

// computeScore returns a 0-100 confidence score for the suggestion set.
// Returns 0 if no fields are suggested, 100 if no issues exist (already complete).
func computeScore(fields map[string]domain.FieldSuggestion) float64 {
	if len(fields) == 0 {
		return 0
	}
	total := 0.0
	for _, f := range fields {
		total += f.Confidence
	}
	return math.Min((total/float64(len(fields)))*100, 100)
}

// --- Filename parser ---

type parsedFile struct {
	title    string
	artist   string
	album    string
	trackNum int
}

var leadingNumRe = regexp.MustCompile(`^(\d{1,3})[.\s\-]+(.*)$`)

func parseFilename(filename string) parsedFile {
	ext := filepath.Ext(filename)
	name := strings.TrimSpace(strings.TrimSuffix(filename, ext))
	result := parsedFile{}

	parts := splitOnDash(name)

	switch len(parts) {
	case 0:
	case 1:
		p := parts[0]
		if num, rest, ok := extractLeadingNum(p); ok {
			result.trackNum = num
			result.title = rest
		} else {
			result.title = p
		}
	case 2:
		p0, p1 := parts[0], parts[1]
		if num, _, ok := extractLeadingNum(p0); ok {
			result.trackNum = num
			result.title = p1
		} else {
			result.artist = p0
			result.title = p1
		}
	case 3:
		p0, p1, p2 := parts[0], parts[1], parts[2]
		if num, _, ok := extractLeadingNum(p0); ok {
			result.trackNum = num
			result.album = p1
			result.title = p2
		} else if num, _, ok := extractLeadingNum(p1); ok {
			result.artist = p0
			result.trackNum = num
			result.title = p2
		} else {
			result.artist = p0
			result.album = p1
			result.title = p2
		}
	default:
		// 4+ parts: Artist - Album - NN - Title
		result.artist = parts[0]
		result.album = parts[1]
		if num, _, ok := extractLeadingNum(parts[2]); ok {
			result.trackNum = num
		}
		result.title = parts[len(parts)-1]
	}

	return result
}

// splitOnDash splits on " - " separator, trimming each part.
func splitOnDash(s string) []string {
	raw := strings.Split(s, " - ")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// extractLeadingNum extracts a leading 1-3 digit number from the start of s.
func extractLeadingNum(s string) (int, string, bool) {
	m := leadingNumRe.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, s, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n == 0 || n > 999 {
		return 0, s, false
	}
	return n, strings.TrimSpace(m[2]), true
}

// --- Helpers ---

func splitPath(rel string) []string {
	raw := strings.Split(rel, string(filepath.Separator))
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		if p != "" && p != "." {
			out = append(out, p)
		}
	}
	return out
}

func topVote(votes map[string]int) (string, int) {
	best, bestCount := "", 0
	for k, v := range votes {
		if v > bestCount {
			bestCount = v
			best = k
		}
	}
	return best, bestCount
}

func normStr(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
