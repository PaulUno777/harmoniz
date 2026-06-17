package organize

import (
	"math"
	"sort"
	"strings"

	"github.com/xrash/smetrics"
	"harmoniz/internal/core/domain"
)

// ArtistEntry represents a canonical artist discovered across the library.
type ArtistEntry struct {
	Canonical  string         // display form: "Daft Punk"
	Norm       string         // normalized: "daftpunk"
	TrackCount int            // number of tracks with this artist in tag form
	Aliases    []string       // all observed raw forms
	Sources    map[string]int // "tag":800, "filename":47
}

// ArtistRegistry is a cross-directory artist index built from all tracks in the library.
type ArtistRegistry struct {
	byNorm map[string]*ArtistEntry
}

// BuildArtistRegistry makes one pass over all tracks to build the artist index.
// Tag values are given priority; filename-parsed values are secondary.
func BuildArtistRegistry(tracks []domain.Track) *ArtistRegistry {
	r := &ArtistRegistry{byNorm: make(map[string]*ArtistEntry)}
	for _, t := range tracks {
		if v := strings.TrimSpace(t.ArtistRaw); v != "" && !isPlaceholderArtist(v) {
			r.upsert(v, "tag")
		}
		parsed := parseFilename(t.Filename)
		if parsed.artist != "" && !isPlaceholderArtist(parsed.artist) {
			r.upsert(parsed.artist, "filename")
		}
	}
	r.mergeAliases()
	return r
}

// upsert inserts or updates an artist entry keyed by normStr(raw).
func (r *ArtistRegistry) upsert(raw, source string) {
	n := normStr(raw)
	if n == "" {
		return
	}
	e, ok := r.byNorm[n]
	if !ok {
		e = &ArtistEntry{
			Canonical: raw,
			Norm:      n,
			Sources:   make(map[string]int),
		}
		r.byNorm[n] = e
	}
	e.Sources[source]++
	if source == "tag" {
		e.TrackCount++
	}
	for _, a := range e.Aliases {
		if a == raw {
			return
		}
	}
	e.Aliases = append(e.Aliases, raw)
}

// mergeAliases performs a Jaro-Winkler pass (threshold 0.92) to merge near-duplicate
// tag-sourced entries (TrackCount ≥ 3) into a single canonical. Uses first-rune bucketing
// to keep it approximately O(n log n).
func (r *ArtistRegistry) mergeAliases() {
	buckets := make(map[rune][]*ArtistEntry)
	for _, e := range r.byNorm {
		if e.Sources["tag"] < 3 || len([]rune(e.Norm)) == 0 {
			continue
		}
		first := []rune(e.Norm)[0]
		buckets[first] = append(buckets[first], e)
	}
	for _, entries := range buckets {
		if len(entries) < 2 {
			continue
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].TrackCount > entries[j].TrackCount
		})
		for i := 0; i < len(entries); i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[j] == entries[i] {
					continue // already merged
				}
				score := smetrics.JaroWinkler(entries[i].Norm, entries[j].Norm, 0.1, 4)
				if score >= 0.92 {
					entries[i].TrackCount += entries[j].TrackCount
					for k, v := range entries[j].Sources {
						entries[i].Sources[k] += v
					}
					entries[i].Aliases = append(entries[i].Aliases, entries[j].Aliases...)
					r.byNorm[entries[j].Norm] = entries[i]
					entries[j] = entries[i] // prevent double-merges
				}
			}
		}
	}
}

// ExactLookup returns the registry entry for an already-normalized artist string.
// Confidence: base 0.80 + log-scaled frequency bonus, max 0.92.
func (r *ArtistRegistry) ExactLookup(norm string) (*ArtistEntry, float64) {
	if norm == "" || r == nil {
		return nil, 0
	}
	e, ok := r.byNorm[norm]
	if !ok {
		return nil, 0
	}
	base := 0.80
	freqBonus := math.Log1p(float64(e.TrackCount)) * 0.015
	return e, math.Min(base+freqBonus, 0.92)
}

// FuzzyLookup finds the closest registry entry for a candidate string using Jaro-Winkler.
// Requires corroboration (TrackCount > 50 or JW ≥ 0.93). Max confidence: 0.75.
func (r *ArtistRegistry) FuzzyLookup(candidate string) (*ArtistEntry, float64) {
	if candidate == "" || r == nil {
		return nil, 0
	}
	cn := normStr(candidate)
	if cn == "" {
		return nil, 0
	}

	var bestEntry *ArtistEntry
	var bestScore float64

	for _, e := range r.byNorm {
		if e.TrackCount < 2 {
			continue
		}
		score := smetrics.JaroWinkler(cn, e.Norm, 0.1, 4)
		if score > bestScore {
			bestScore = score
			bestEntry = e
		}
	}

	if bestEntry == nil || bestScore < 0.86 {
		return nil, 0
	}
	if bestEntry.TrackCount <= 50 && bestScore < 0.93 {
		return nil, 0 // insufficient corroboration
	}

	// Map to [0, 0.75] range
	conf := math.Min((bestScore-0.86)/(1.0-0.86)*0.75, 0.75)
	return bestEntry, conf
}

// Size returns the number of unique normalized entries in the registry.
func (r *ArtistRegistry) Size() int {
	if r == nil {
		return 0
	}
	return len(r.byNorm)
}
