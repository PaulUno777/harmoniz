package organize

import (
	"context"
	"math"
	"path/filepath"
	"regexp"
	"sort"
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
		sugg := s.buildSuggestion(t, dirInfos[dir])
		suggestions = append(suggestions, sugg)
	}
	return suggestions, nil
}

// dirMeta holds path-inferred and tag-voted metadata for a single directory.
type dirMeta struct {
	artistCandidate string
	albumCandidate  string
	pathConfidence  float64
	tagArtist       string  // plurality-vote artist from existing sibling tags
	tagAlbum        string
	tagConfidence   float64
	inferredArtist  string  // cross-file canonical artist inferred from filenames
	inferredArtConf float64
}

func (s *SuggestionService) computeDirInfos(root string, byDir map[string][]domain.Track) map[string]dirMeta {
	result := map[string]dirMeta{}
	cleanRoot := filepath.Clean(root)

	for dir, tracks := range byDir {
		info := dirMeta{}
		cleanDir := filepath.Clean(dir)

		rel, err := filepath.Rel(cleanRoot, cleanDir)
		if err == nil {
			if rel == "." {
				// Files directly in root: use the root folder name as artist hint
				info.artistCandidate = normalizeText(filepath.Base(cleanRoot))
				info.pathConfidence = 0.50
			} else {
				parts := splitPath(rel)
				switch len(parts) {
				case 1:
					info.artistCandidate = normalizeText(parts[0])
					info.pathConfidence = 0.65
				case 2:
					info.artistCandidate = normalizeText(parts[0])
					info.albumCandidate = normalizeText(parts[1])
					info.pathConfidence = 0.75
				default:
					info.artistCandidate = normalizeText(parts[len(parts)-2])
					info.albumCandidate = normalizeText(parts[len(parts)-1])
					info.pathConfidence = 0.65
				}
			}
		}

		// Plurality vote from existing sibling tags (accept at ≥15% rather than strict majority)
		artistVotes := map[string]int{}
		albumVotes := map[string]int{}
		for _, t := range tracks {
			if v := strings.TrimSpace(t.ArtistRaw); v != "" && !isPlaceholderArtist(v) {
				artistVotes[v]++
			}
			if v := strings.TrimSpace(t.AlbumRaw); v != "" && !isPlaceholderAlbum(v) {
				albumVotes[v]++
			}
		}
		total := len(tracks)

		if artist, count := topVote(artistVotes); count > 0 {
			thresh := math.Max(1, float64(total)*0.15)
			if float64(count) >= thresh {
				info.tagArtist = artist
				info.tagConfidence = math.Min(0.75+0.20*(float64(count)/float64(total)), 0.93)
			}
		}
		if album, count := topVote(albumVotes); float64(count) >= math.Max(1, float64(total)*0.30) {
			info.tagAlbum = album
			if info.tagConfidence == 0 {
				info.tagConfidence = math.Min(0.70+0.20*(float64(count)/float64(total)), 0.90)
			}
		}

		// Cross-file canonical artist inference from filename patterns
		info.inferredArtist, info.inferredArtConf = inferCanonicalArtist(tracks, info.artistCandidate)

		result[dir] = info
	}
	return result
}

// isPlaceholderArtist returns true for well-known placeholder values.
func isPlaceholderArtist(s string) bool {
	n := normStr(s)
	return n == "unknownartist" || n == "artisteinconnu" || n == "unknown" ||
		n == "inconnu" || n == "artiste" || n == "variousartists" || n == "va"
}

// inferCanonicalArtist finds the canonical artist name by analysing filename patterns
// across all tracks: it finds the longest normalized substring common to the most
// artist-portion variants while containing the folder hint.
func inferCanonicalArtist(tracks []domain.Track, folderHint string) (string, float64) {
	if len(tracks) == 0 {
		return "", 0
	}

	folderNorm := normStr(folderHint)

	// Collect unique normalized artist forms → shortest raw representative
	normToRaw := map[string]string{}
	normCount := map[string]int{}

	for _, t := range tracks {
		parsed := parseFilename(t.Filename)
		if parsed.artist == "" {
			continue
		}
		n := normStr(parsed.artist)
		if n == "" {
			continue
		}
		normCount[n]++
		// Keep shortest raw form (fewest extra tokens like album names)
		if prev, ok := normToRaw[n]; !ok || len(parsed.artist) < len(prev) {
			normToRaw[n] = parsed.artist
		}
	}

	if len(normCount) == 0 {
		return "", 0
	}

	total := len(tracks)
	uniqueNorms := make([]string, 0, len(normCount))
	for n := range normCount {
		uniqueNorms = append(uniqueNorms, n)
	}

	// For each norm, count how many OTHER unique norms contain it as a substring.
	// A norm with high containment count is a core component of many artist variants.
	type cand struct {
		norm         string
		superCount   int // how many unique norms contain this as substring
		ownFreq      float64
	}

	var candidates []cand
	for _, n := range uniqueNorms {
		if folderNorm != "" && !strings.Contains(n, folderNorm) {
			continue
		}
		sc := 0
		for _, other := range uniqueNorms {
			if strings.Contains(other, n) {
				sc++
			}
		}
		candidates = append(candidates, cand{
			norm:       n,
			superCount: sc,
			ownFreq:    float64(normCount[n]) / float64(total),
		})
	}

	if len(candidates) == 0 {
		// No candidates contain folder hint — fall back to most frequent
		best, count := topVote(normCount)
		if count == 0 {
			return "", 0
		}
		return normToRaw[best], math.Min(0.55+float64(count)/float64(total)*0.20, 0.72)
	}

	// Among candidates whose superCount meets the threshold (≥50% of unique norms),
	// prefer the LONGEST normalized form (most complete name without extra album tokens).
	threshold := math.Max(1, float64(len(uniqueNorms))*0.50)

	var qualified []cand
	for _, c := range candidates {
		if float64(c.superCount) >= threshold {
			qualified = append(qualified, c)
		}
	}
	if len(qualified) == 0 {
		// Loosen: take the highest superCount
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].superCount > candidates[j].superCount
		})
		qualified = candidates[:1]
	}

	// Sort qualified by length desc (prefer longer = more complete)
	sort.Slice(qualified, func(i, j int) bool {
		return len(qualified[i].norm) > len(qualified[j].norm)
	})
	best := qualified[0]

	subFreq := float64(best.superCount) / float64(len(uniqueNorms))
	conf := math.Min(0.65+best.ownFreq*0.15+subFreq*0.10, 0.88)
	if folderNorm != "" && strings.Contains(best.norm, folderNorm) {
		conf = math.Min(conf+0.05, 0.90)
	}
	return normToRaw[best.norm], conf
}

func (s *SuggestionService) buildSuggestion(t domain.Track, info dirMeta) domain.OrganizerSuggestion {
	fields := map[string]domain.FieldSuggestion{}
	issues := []string{}

	parsed := parseFilename(t.Filename)

	// --- Artist ---
	existingArtist := strings.TrimSpace(t.ArtistRaw)
	if existingArtist == "" || isPlaceholderArtist(existingArtist) {
		// Missing or placeholder artist
		if info.tagArtist != "" {
			fields["artist"] = domain.FieldSuggestion{Value: info.tagArtist, Confidence: info.tagConfidence, Source: "neighbor"}
		} else if info.inferredArtist != "" {
			fields["artist"] = domain.FieldSuggestion{Value: info.inferredArtist, Confidence: info.inferredArtConf, Source: "filename"}
		} else if parsed.artist != "" {
			conf := 0.65
			if info.artistCandidate != "" && strings.Contains(normStr(parsed.artist), normStr(info.artistCandidate)) {
				conf = 0.80
			}
			fields["artist"] = domain.FieldSuggestion{Value: parsed.artist, Confidence: conf, Source: "filename"}
		} else if info.artistCandidate != "" {
			fields["artist"] = domain.FieldSuggestion{Value: info.artistCandidate, Confidence: info.pathConfidence, Source: "path"}
		}
		issues = append(issues, "Missing artist")
	} else {
		// Artist exists — suggest normalization if it looks malformed
		normalized := normalizeExistingArtist(existingArtist, info.tagArtist)
		if normalized != existingArtist {
			fields["artist"] = domain.FieldSuggestion{Value: normalized, Confidence: 0.72, Source: "neighbor"}
		}
	}

	// --- Album ---
	existingAlbum := strings.TrimSpace(t.AlbumRaw)
	albumIsPlaceholder := existingAlbum != "" && isPlaceholderAlbum(existingAlbum)
	if existingAlbum == "" || albumIsPlaceholder {
		if info.tagAlbum != "" {
			fields["album"] = domain.FieldSuggestion{Value: info.tagAlbum, Confidence: info.tagConfidence, Source: "neighbor"}
		} else if parsed.album != "" {
			conf := 0.60
			if info.albumCandidate != "" && normStr(parsed.album) == normStr(info.albumCandidate) {
				conf = 0.80
			}
			fields["album"] = domain.FieldSuggestion{Value: parsed.album, Confidence: conf, Source: "filename"}
		} else if info.albumCandidate != "" {
			fields["album"] = domain.FieldSuggestion{Value: info.albumCandidate, Confidence: info.pathConfidence, Source: "path"}
		}
		if existingAlbum == "" {
			issues = append(issues, "Missing album")
		}
	}

	// --- Year: salvage from placeholder album timestamp e.g. "(09/10/2008 09:46:25)" ---
	if t.Year == 0 && albumIsPlaceholder {
		if m := yearInStringRe.FindString(existingAlbum); m != "" {
			if y, err := strconv.Atoi(m); err == nil && y >= 1960 && y <= 2030 {
				fields["year"] = domain.FieldSuggestion{Value: m, Confidence: 0.65, Source: "album_tag"}
			}
		}
	}

	// --- Title ---
	existingTitle := strings.TrimSpace(t.Title)
	titleIsPlaceholder := existingTitle != "" && isPlaceholderTitle(existingTitle)
	if existingTitle == "" || titleIsPlaceholder {
		// Missing or auto-generated placeholder (e.g. "Piste 2", "Track 01"):
		// derive the real title from the filename. The filename-parsed title already
		// has the artist prefix stripped (via --- or " - " split), so cleanTitle here
		// only needs to normalize. If parsed.title is also junk, we leave it.
		if parsed.title != "" && !isPlaceholderTitle(parsed.title) {
			clean := cleanTitle(parsed.title, info.tagArtist, info.inferredArtist, info.artistCandidate)
			if clean == "" {
				clean = normalizeText(parsed.title)
			}
			if clean != "" && normStr(clean) != normStr(existingTitle) {
				conf := 0.85
				if titleIsPlaceholder {
					conf = 0.80 // slightly lower: we're overriding an existing (bad) tag
				}
				fields["title"] = domain.FieldSuggestion{Value: clean, Confidence: conf, Source: "filename"}
			}
		}
		if existingTitle == "" {
			issues = append(issues, "Missing title")
		} else {
			issues = append(issues, "Placeholder title")
		}
	} else {
		// Title exists — suggest cleanup if it has audio extension, underscores, or artist prefix
		clean := cleanTitle(existingTitle, info.tagArtist, info.inferredArtist, info.artistCandidate)
		if clean != existingTitle {
			fields["title"] = domain.FieldSuggestion{Value: clean, Confidence: 0.75, Source: "filename"}
		}
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

// normalizeExistingArtist suggests a better form for a malformed artist string.
// If tagArtist is provided and looks better (longer/more complete), prefer it.
func normalizeExistingArtist(existing, tagArtist string) string {
	// If there's a consensus tag artist and it contains the same name, prefer it
	if tagArtist != "" && strings.Contains(normStr(tagArtist), normStr(existing)) {
		return tagArtist
	}
	// All-caps → Title Case
	if isAllCaps(existing) {
		return toTitleCase(strings.ReplaceAll(existing, "_", " "))
	}
	// Has underscores → replace
	if strings.Contains(existing, "_") {
		return normalizeText(existing)
	}
	return existing
}

// cleanTitle removes audio-extension suffixes, strips an artist-name prefix if present,
// and applies text normalization (underscores→spaces, Title Case).
// Each candidate is tried in full then word-by-word so a partial name like "AKOUALA"
// is stripped even when the canonical is "Sylvain Akouala".
func cleanTitle(title string, candidates ...string) string {
	// 1. Strip embedded audio extension (e.g. title stored as "Song.mp3")
	title = audioExtRe.ReplaceAllString(title, "")
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}

	tn := normStr(title)
	titleRunes := []rune(title)

	// 2. Try each candidate: full name first, then word by word
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		// Full candidate
		if res := tryStripArtistPrefix(titleRunes, tn, normStr(candidate)); res != "" {
			return res
		}
		// Individual words (handles "AKOUALA" prefix when canonical is "Sylvain Akouala")
		for _, word := range strings.Fields(candidate) {
			wn := normStr(word)
			if len([]rune(wn)) < 4 {
				continue // skip short particles like "Fr", "de", "la"
			}
			if res := tryStripArtistPrefix(titleRunes, tn, wn); res != "" {
				return res
			}
		}
	}

	// 3. Normalize underscores and casing
	return normalizeText(title)
}

// tryStripArtistPrefix attempts to remove prefixNorm from the beginning of title.
// Returns the normalized remainder, or "" if the prefix is not present.
func tryStripArtistPrefix(titleRunes []rune, tn, prefixNorm string) string {
	if prefixNorm == "" || !strings.HasPrefix(tn, prefixNorm) {
		return ""
	}
	cut := findCutPoint(string(titleRunes), prefixNorm)
	if cut > 0 && cut < len(titleRunes) {
		rest := strings.TrimLeft(string(titleRunes[cut:]), " -_—,.")
		if len([]rune(rest)) > 2 {
			return normalizeText(rest)
		}
	}
	return ""
}

// findCutPoint returns the rune index in title where the prefixNorm (letters/digits only) ends.
func findCutPoint(title string, prefixNorm string) int {
	prefixRunes := []rune(prefixNorm)
	titleRunes := []rune(title)
	pi := 0
	for ti, r := range titleRunes {
		if pi >= len(prefixRunes) {
			return ti
		}
		rl := unicode.ToLower(r)
		if unicode.IsLetter(rl) || unicode.IsDigit(rl) {
			if rl == prefixRunes[pi] {
				pi++
			} else {
				return 0 // mismatch in letter sequence
			}
		}
		// Non-letter/digit (underscore, hyphen, space): skip
	}
	if pi >= len(prefixRunes) {
		return len(titleRunes)
	}
	return 0
}

// computeScore returns a 0–100 confidence score for the suggestion set.
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

var (
	leadingNumRe       = regexp.MustCompile(`^(\d{1,3})[.\s\-_]+(.+)$`)
	multiHyphenRe      = regexp.MustCompile(`-{2,}`)
	hyphenWordRe       = regexp.MustCompile(`([^\s\-])-([^\s\-])`)
	audioExtRe         = regexp.MustCompile(`(?i)\.(mp3|flac|ogg|aac|m4a|wav|aiff|wma|opus|ape|wv)$`)
	placeholderTitleRe = regexp.MustCompile(`^(piste|track|titre|pista|traccia|untitled|unknown|inconnu)\d*$`)
	placeholderAlbumRe = regexp.MustCompile(`^(albuminconnu|unknownalbum|albumdesconhecido|noalbum|inconnu|unknown)`)
	timestampAlbumRe   = regexp.MustCompile(`\(\d{2}/\d{2}/\d{4}`)
	yearInStringRe     = regexp.MustCompile(`\b(19[6-9]\d|20[0-3]\d)\b`)
)

// isPlaceholderTitle returns true for generic auto-generated titles like "Piste 2", "Track 01".
func isPlaceholderTitle(s string) bool {
	return placeholderTitleRe.MatchString(normStr(s))
}

// isPlaceholderAlbum returns true for machine-generated album names like
// "Album inconnu (09/10/2008 09:46:25)" that carry no useful information.
func isPlaceholderAlbum(s string) bool {
	if placeholderAlbumRe.MatchString(normStr(s)) {
		return true
	}
	return timestampAlbumRe.MatchString(s)
}

func parseFilename(filename string) parsedFile {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	// Handle accidental double-extension stored in filename
	if ext2 := filepath.Ext(name); ext2 != "" && strings.EqualFold(ext2, ext) {
		name = strings.TrimSuffix(name, ext2)
	}
	name = strings.TrimSpace(name)
	result := parsedFile{}

	// Strategy 1: split on --- (2+ hyphens used as artist---title separator)
	parts2 := multiHyphenRe.Split(name, 2)
	if len(parts2) == 2 {
		left := strings.TrimSpace(parts2[0])
		right := strings.TrimSpace(parts2[1])
		if left != "" && right != "" {
			if num, rest, ok := extractLeadingNum(left); ok {
				result.trackNum = num
				result.title = normalizeText(right)
				_ = rest
			} else if num, rest, ok := extractLeadingNum(right); ok {
				result.artist = normalizeText(left)
				result.trackNum = num
				result.title = normalizeText(rest)
			} else {
				result.artist = normalizeText(left)
				result.title = normalizeText(right)
			}
			return result
		}
	}

	// Strategy 2: split on " - " (traditional convention)
	dashParts := splitOnDash(name)
	switch len(dashParts) {
	case 0:
	case 1:
		p := dashParts[0]
		if num, rest, ok := extractLeadingNum(p); ok {
			result.trackNum = num
			result.title = normalizeText(rest)
		} else {
			result.title = normalizeText(p)
		}
	case 2:
		p0, p1 := dashParts[0], dashParts[1]
		if num, _, ok := extractLeadingNum(p0); ok {
			result.trackNum = num
			result.title = normalizeText(p1)
		} else {
			result.artist = normalizeText(p0)
			result.title = normalizeText(p1)
		}
	case 3:
		p0, p1, p2 := dashParts[0], dashParts[1], dashParts[2]
		if num, _, ok := extractLeadingNum(p0); ok {
			result.trackNum = num
			result.album = normalizeText(p1)
			result.title = normalizeText(p2)
		} else if num, _, ok := extractLeadingNum(p1); ok {
			result.artist = normalizeText(p0)
			result.trackNum = num
			result.title = normalizeText(p2)
		} else {
			result.artist = normalizeText(p0)
			result.album = normalizeText(p1)
			result.title = normalizeText(p2)
		}
	default:
		result.artist = normalizeText(dashParts[0])
		result.album = normalizeText(dashParts[1])
		if num, _, ok := extractLeadingNum(dashParts[2]); ok {
			result.trackNum = num
		}
		result.title = normalizeText(dashParts[len(dashParts)-1])
	}

	return result
}

// normalizeText converts raw filename tokens to readable text:
// underscores → spaces, hyphens-as-word-separators → spaces, Title Case.
func normalizeText(s string) string {
	if s == "" {
		return s
	}
	s = strings.ReplaceAll(s, "_", " ")
	// Replace hyphens between non-hyphen, non-space chars with a space
	s = hyphenWordRe.ReplaceAllStringFunc(s, func(m string) string {
		r := []rune(m)
		return string(r[0]) + " " + string(r[2])
	})
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	return toTitleCase(s)
}

// toTitleCase capitalizes the first letter of each word, lowercasing the rest.
// French articles/prepositions after the first word are kept lowercase.
func toTitleCase(s string) string {
	if s == "" {
		return s
	}
	lowercase := map[string]bool{
		"de": true, "du": true, "des": true, "la": true, "le": true,
		"les": true, "et": true, "ou": true, "en": true, "au": true,
		"aux": true, "un": true, "une": true, "d": true, "l": true,
	}
	words := strings.Fields(s)
	for i, w := range words {
		runes := []rune(w)
		if len(runes) == 0 {
			continue
		}
		wl := strings.ToLower(w)
		if i > 0 && lowercase[wl] {
			words[i] = wl
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		for j := 1; j < len(runes); j++ {
			runes[j] = unicode.ToLower(runes[j])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// isAllCaps returns true if s contains only uppercase letters (and non-letters).
func isAllCaps(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if unicode.IsLower(r) {
				return false
			}
		}
	}
	return hasLetter && len([]rune(s)) > 3
}

// firstNonEmpty returns the first non-empty string from the arguments.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
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

// extractLeadingNum extracts a leading 1–3 digit number from s.
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
