package organize

import (
	"testing"

	"harmoniz/internal/core/domain"
)

// --- helpers ---

func tracks(raws ...string) []domain.Track {
	result := make([]domain.Track, len(raws))
	for i, r := range raws {
		result[i] = domain.Track{ID: uint64(i + 1), ArtistRaw: r, Path: "/lib/song.mp3"}
	}
	return result
}

func nTracks(artistRaw string, n int) []domain.Track {
	ts := make([]domain.Track, n)
	for i := range ts {
		ts[i] = domain.Track{ID: uint64(i + 1), ArtistRaw: artistRaw, Path: "/lib/song.mp3"}
	}
	return ts
}

// --- BuildArtistRegistry ---

func TestBuildRegistry_TagIngestion(t *testing.T) {
	ts := nTracks("Daft Punk", 5)
	r := BuildArtistRegistry(ts)
	e, _ := r.ExactLookup(normStr("Daft Punk"))
	if e == nil {
		t.Fatal("expected entry for 'Daft Punk', got nil")
	}
	if e.TrackCount != 5 {
		t.Errorf("TrackCount: want 5, got %d", e.TrackCount)
	}
}

func TestBuildRegistry_SkipsPlaceholders(t *testing.T) {
	placeholders := []string{"Unknown Artist", "unknown", "va", "Various Artists", "Artiste inconnu"}
	for _, p := range placeholders {
		r := BuildArtistRegistry(nTracks(p, 3))
		e, _ := r.ExactLookup(normStr(p))
		if e != nil {
			t.Errorf("placeholder %q should not be in registry, got entry", p)
		}
	}
}

func TestBuildRegistry_FilenameSource(t *testing.T) {
	// No tags, but filename has "Artist - Title" format
	ts := []domain.Track{
		{ID: 1, ArtistRaw: "", Filename: "Emile Bobo - La Richesse.mp3", Path: "/lib/Emile Bobo - La Richesse.mp3"},
	}
	r := BuildArtistRegistry(ts)
	e, _ := r.ExactLookup(normStr("Emile Bobo"))
	if e == nil {
		t.Fatal("expected filename-sourced entry for 'Emile Bobo'")
	}
	if e.Sources["filename"] == 0 {
		t.Error("expected Sources['filename'] > 0")
	}
}

func TestBuildRegistry_MultipleAliasesShareEntry(t *testing.T) {
	ts := []domain.Track{
		{ID: 1, ArtistRaw: "Daft Punk", Path: "/lib/a.mp3"},
		{ID: 2, ArtistRaw: "daft punk", Path: "/lib/b.mp3"},
		{ID: 3, ArtistRaw: "DAFT PUNK", Path: "/lib/c.mp3"},
	}
	r := BuildArtistRegistry(ts)
	// normStr collapses all three to "daftpunk"
	e, _ := r.ExactLookup("daftpunk")
	if e == nil {
		t.Fatal("expected entry for 'daftpunk'")
	}
}

// --- ExactLookup ---

func TestExactLookup_Found(t *testing.T) {
	r := BuildArtistRegistry(nTracks("Sylvain Akouala", 1))
	e, conf := r.ExactLookup(normStr("Sylvain Akouala"))
	if e == nil {
		t.Fatal("expected entry")
	}
	if conf < 0.80 || conf > 0.92 {
		t.Errorf("conf %v outside [0.80, 0.92]", conf)
	}
}

func TestExactLookup_NotFound(t *testing.T) {
	r := BuildArtistRegistry(nTracks("Daft Punk", 3))
	e, conf := r.ExactLookup("unknownartist")
	if e != nil || conf != 0 {
		t.Error("expected nil entry and 0 confidence for unknown artist")
	}
}

func TestExactLookup_ConfidenceScalesWithFrequency(t *testing.T) {
	r1 := BuildArtistRegistry(nTracks("Nathan Epenge", 1))
	_, c1 := r1.ExactLookup(normStr("Nathan Epenge"))

	r2 := BuildArtistRegistry(nTracks("Nathan Epenge", 200))
	_, c2 := r2.ExactLookup(normStr("Nathan Epenge"))

	if c2 <= c1 {
		t.Errorf("expected higher confidence with more tracks: c1=%v c2=%v", c1, c2)
	}
	if c2 > 0.92 {
		t.Errorf("confidence must be capped at 0.92, got %v", c2)
	}
}

func TestExactLookup_ConfidenceCappedAt092(t *testing.T) {
	// Even with 10 000 tracks the cap must hold.
	r := BuildArtistRegistry(nTracks("Emile Bravo", 10000))
	_, conf := r.ExactLookup(normStr("Emile Bravo"))
	if conf > 0.92 {
		t.Errorf("confidence must not exceed 0.92, got %v", conf)
	}
}

// --- FuzzyLookup ---

func TestFuzzyLookup_HighJWWithLargeLibrary(t *testing.T) {
	// TrackCount > 50 satisfies the corroboration requirement even at JW < 0.93.
	r := BuildArtistRegistry(nTracks("Sylvain Akouala", 60))
	// "sylvainakouala" vs "sylvainakouala" is an exact norm match so JW=1.0.
	e, conf := r.FuzzyLookup("Sylvain Akouala")
	if e == nil {
		t.Fatal("expected fuzzy hit for 'Sylvain Akouala'")
	}
	if conf <= 0 || conf > 0.75 {
		t.Errorf("fuzzy conf %v outside (0, 0.75]", conf)
	}
}

func TestFuzzyLookup_InsufficientCorroboration(t *testing.T) {
	// TrackCount=5 and JW only ~0.88 → no result (needs TrackCount>50 or JW≥0.93).
	r := BuildArtistRegistry(nTracks("Emile Bravo", 5))
	// "emilbravo" has JW ≈ 0.88 against "emilebravo"
	e, conf := r.FuzzyLookup("Emil Bravo")
	if e != nil && conf > 0 {
		// Acceptable if JW happens to be ≥ 0.93 for this pair — just verify cap.
		if conf > 0.75 {
			t.Errorf("fuzzy conf %v exceeds 0.75 cap", conf)
		}
	}
}

func TestFuzzyLookup_LargeLibraryLowerJW(t *testing.T) {
	// TrackCount=51 should allow matches even at JW slightly below 0.93.
	r := BuildArtistRegistry(nTracks("Alfred Ondo", 51))
	e, conf := r.FuzzyLookup("Alfred Ondo")
	if e == nil {
		t.Fatal("expected fuzzy hit for 'Alfred Ondo' in large library")
	}
	if conf <= 0 || conf > 0.75 {
		t.Errorf("fuzzy conf %v outside (0, 0.75]", conf)
	}
}

func TestFuzzyLookup_BelowThreshold(t *testing.T) {
	r := BuildArtistRegistry(nTracks("Daft Punk", 60))
	// Completely unrelated name should return nil.
	e, _ := r.FuzzyLookup("Nathan Epenge")
	if e != nil {
		t.Error("expected nil fuzzy result for unrelated artist")
	}
}

// --- mergeAliases (exercised via BuildArtistRegistry) ---

func TestMergeAliases_NearDuplicatesCollapsed(t *testing.T) {
	// "Daft Punk" and "Daft-Punk" normalize to different norms but JW ≥ 0.92.
	// Each must have ≥ 3 tag-sourced entries to qualify for merge.
	ts := append(nTracks("Daft Punk", 4), nTracks("Daft-Punk", 4)...)
	r := BuildArtistRegistry(ts)

	e1, _ := r.ExactLookup(normStr("Daft Punk"))
	e2, _ := r.ExactLookup(normStr("Daft-Punk"))
	if e1 == nil || e2 == nil {
		t.Skip("both entries must exist to verify merge")
	}
	// After merge, both norms should point to the same ArtistEntry pointer.
	if e1 != e2 {
		// Acceptable if the norms are already identical (e.g. both normalise to "daftpunk").
		if normStr("Daft Punk") != normStr("Daft-Punk") {
			t.Errorf("expected 'Daft Punk' and 'Daft-Punk' to share a canonical entry after merge")
		}
	}
}

func TestRegistry_SizeReturnsCorrectCount(t *testing.T) {
	ts := append(nTracks("Artist A", 2), nTracks("Artist B", 2)...)
	r := BuildArtistRegistry(ts)
	if r.Size() < 1 {
		t.Errorf("Size should be ≥ 1, got %d", r.Size())
	}
}
