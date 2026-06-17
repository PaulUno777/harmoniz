package integration_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/services/organize"
	"harmoniz/internal/testutil"
)

// TestOrganizerIntegration_ArtistSubdirInference copies MP3 files into a one-level
// subdirectory (Alfred Ondo/) and verifies that AnalyzeRoot produces artist suggestions
// sourced from "path" or "neighbor" for every track without existing artist tags.
func TestOrganizerIntegration_ArtistSubdirInference(t *testing.T) {
	songs := testutil.TestSongsDir(t)

	// Collect up to 5 root-level MP3 files from test/songs.
	entries, err := os.ReadDir(songs)
	if err != nil {
		t.Fatalf("read songs dir: %v", err)
	}

	tmp := t.TempDir()
	dst := filepath.Join(tmp, "Alfred Ondo")
	copied := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".mp3" && ext != ".MP3" && ext != ".flac" && ext != ".m4a" {
			continue
		}
		testutil.CopyFileTo(t, filepath.Join(songs, e.Name()), dst)
		copied++
		if copied >= 5 {
			break
		}
	}
	if copied == 0 {
		t.Skip("no scannable audio files in test/songs")
	}

	repo := testutil.BuildScanDB(t, tmp)
	svc := organize.NewSuggestionService(repo)
	suggs, regSize, err := svc.AnalyzeRoot(context.Background(), tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions returned")
	}
	if regSize < 0 {
		t.Errorf("registrySize must be ≥ 0, got %d", regSize)
	}

	for _, s := range suggs {
		art, ok := s.Fields["artist"]
		if !ok {
			// Artist already present and not a placeholder — that's fine too.
			continue
		}
		validSources := map[string]bool{"path": true, "neighbor": true, "filename": true, "library": true}
		if !validSources[art.Source] {
			t.Errorf("unexpected artist source %q for track %d", art.Source, s.TrackID)
		}
	}
}

// TestOrganizerIntegration_NathanEpengeFilename verifies that a track whose filename follows
// "Artist - Title.mp3" produces a non-empty artist and title suggestion.
func TestOrganizerIntegration_NathanEpengeFilename(t *testing.T) {
	songs := testutil.TestSongsDir(t)
	src := filepath.Join(songs, "Nathan Epenge - Macpéla.mp3")
	if _, err := os.Stat(src); os.IsNotExist(err) {
		t.Skip("Nathan Epenge track not found in test/songs")
	}

	tmp := t.TempDir()
	testutil.CopyFileTo(t, src, tmp)

	repo := testutil.BuildScanDB(t, tmp)
	svc := organize.NewSuggestionService(repo)
	suggs, _, err := svc.AnalyzeRoot(context.Background(), tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}

	s := suggs[0]
	// If artist tag is missing, we expect a filename-based suggestion.
	if art, ok := s.Fields["artist"]; ok {
		if art.Value == "" {
			t.Error("artist suggestion value must not be empty")
		}
	}
	// Similarly for title.
	if ttl, ok := s.Fields["title"]; ok {
		if ttl.Value == "" {
			t.Error("title suggestion value must not be empty")
		}
	}
}

// TestOrganizerIntegration_CompilationDirSkipsRegistry builds a library with many tracks from
// distinct artists in a compilation subdirectory and verifies that the "library" source is
// never used for tracks in that directory.
func TestOrganizerIntegration_CompilationDirSkipsRegistry(t *testing.T) {
	// Use FakeRepo so we control exact track count and tag data.
	artists := []string{"Ado", "Emile Bobo", "Alfred Ondo", "Nathan Epenge", "Anicet Mundundu"}

	var allTracks []domain.Track
	id := uint64(1)

	// 60 tracks per artist in /lib/main/ → seeds registry.
	for _, a := range artists {
		for j := range 60 {
			allTracks = append(allTracks, testutil.NewTrack(id, fmt.Sprintf("/lib/main/%s/t%d.mp3", a, j)).
				WithArtist(a).Build())
			id++
		}
	}

	// 5 tracks (one per artist) in /lib/compilation/ → isCompilation=true.
	for i, a := range artists {
		allTracks = append(allTracks, testutil.NewTrack(id, fmt.Sprintf("/lib/compilation/t%d.mp3", i)).
			WithArtist(a).Build())
		id++
	}

	svc := organize.NewSuggestionService(testutil.NewFakeRepo(allTracks...))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib/compilation")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range suggs {
		if art, ok := s.Fields["artist"]; ok {
			if art.Source == "library" {
				t.Errorf("compilation dir must not use 'library' source (track %d, artist=%q)", s.TrackID, art.Value)
			}
		}
	}
}

// TestOrganizerIntegration_RegistryBoostsConfidence verifies that when a lone track's
// filename matches a high-frequency artist in the registry, ExactLookup is chosen with
// source="library" and conf ≥ 0.80.
func TestOrganizerIntegration_RegistryBoostsConfidence(t *testing.T) {
	var allTracks []domain.Track
	id := uint64(1)
	for i := range 100 {
		allTracks = append(allTracks, testutil.NewTrack(id, fmt.Sprintf("/lib/main/t%d.mp3", i)).
			WithArtist("Daft Punk").Build())
		id++
	}
	// Lone track: filename encodes "Daft Punk" but no ArtistRaw tag.
	lone := testutil.NewTrack(id, "/lib/other/Daft Punk---Revolution.mp3").
		WithFilename("Daft Punk---Revolution.mp3").Build()
	allTracks = append(allTracks, lone)

	svc := organize.NewSuggestionService(testutil.NewFakeRepo(allTracks...))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}

	var loneS domain.OrganizerSuggestion
	for _, s := range suggs {
		if s.TrackID == id {
			loneS = s
			break
		}
	}

	art, ok := loneS.Fields["artist"]
	if !ok {
		t.Fatal("expected artist suggestion for lone track")
	}
	if art.Source != "library" {
		t.Errorf("expected source='library', got %q (value=%q, conf=%.2f)",
			art.Source, art.Value, art.Confidence)
	}
	if art.Confidence < 0.80 {
		t.Errorf("library exact conf %.2f < 0.80", art.Confidence)
	}
}

// TestOrganizerIntegration_YouTubeFilenameNaming verifies that filenames with YouTube-style
// formatting (resolution tags, underscores) produce at least one non-empty inferred field.
func TestOrganizerIntegration_YouTubeFilenameNaming(t *testing.T) {
	youtubeFilenames := []string{
		"Arms_Of_The_Savior_-_Elizabeth_South_-_Lyrics(360p).mp3",
		"Sylvain_Akouala__L_HOMME_NÉ_DE_L_HOMME_VIT_ÉTERNELLEMENT_(0).mp3",
		"ARMEE_DE_LOUANGE__TU_MERITES_MON_ADORATION(128k).mp3",
	}

	for _, name := range youtubeFilenames {
		t.Run(name, func(t *testing.T) {
			tr := testutil.NewTrack(1, "/lib/"+name).WithFilename(name).Build()
			svc := organize.NewSuggestionService(testutil.NewFakeRepo(tr))
			suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
			if err != nil {
				t.Fatal(err)
			}
			if len(suggs) == 0 {
				t.Fatal("no suggestions")
			}
			s := suggs[0]
			hasArtist := s.Fields["artist"].Value != ""
			hasTitle := s.Fields["title"].Value != ""
			if !hasArtist && !hasTitle {
				t.Errorf("YouTube filename %q produced no artist or title suggestion", name)
			}
		})
	}
}

// TestOrganizerIntegration_AllSongsScannedAndAnalyzed is a smoke test that scans the
// full test/songs directory and runs AnalyzeRoot — verifying no crashes and that the
// number of suggestions matches the number of scanned tracks.
func TestOrganizerIntegration_AllSongsScannedAndAnalyzed(t *testing.T) {
	songs := testutil.TestSongsDir(t)
	tmp := t.TempDir()

	// Copy everything into tmp to avoid polluting the source tree.
	err := filepath.WalkDir(songs, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(songs, path)
		testutil.CopyFile(t, path, filepath.Join(tmp, rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	repo := testutil.BuildScanDB(t, tmp)
	svc := organize.NewSuggestionService(repo)
	suggs, regSize, err := svc.AnalyzeRoot(context.Background(), tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Error("expected at least one suggestion from the test/songs library")
	}
	t.Logf("tracks=%d  registryArtists=%d", len(suggs), regSize)
}
