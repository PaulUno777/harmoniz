package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/services/analysis"
	"harmoniz/internal/testutil"
)

// pickSong returns an absolute path to one MP3 in test/songs, skipping if not found.
func pickSong(t *testing.T, name string) string {
	t.Helper()
	songs := testutil.TestSongsDir(t)
	p := filepath.Join(songs, name)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Skipf("test song %q not found", name)
	}
	return p
}

// pickAnyMP3 returns the path of the first scannable audio file in test/songs.
// The scanner only processes mp3, flac, m4a, ogg, wav — not wma.
func pickAnyMP3(t *testing.T) string {
	t.Helper()
	songs := testutil.TestSongsDir(t)
	entries, err := os.ReadDir(songs)
	if err != nil {
		t.Fatalf("read songs dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext == ".mp3" || ext == ".MP3" || ext == ".flac" || ext == ".m4a" {
			return filepath.Join(songs, e.Name())
		}
	}
	t.Skip("no scannable MP3/FLAC/M4A files in test/songs")
	return ""
}

// ============================================================
// " copy" suffix — the most common duplicate pattern in the library
// ============================================================

func TestCleanerIntegration_SpaceCopySuffix(t *testing.T) {
	src := pickSong(t, "Nathan Epenge - Macpéla.mp3")
	tmp := t.TempDir()

	orig := filepath.Join(tmp, "Nathan Epenge - Macpéla.mp3")
	copy_ := filepath.Join(tmp, "Nathan Epenge - Macpéla copy.mp3")
	testutil.CopyFile(t, src, orig)
	testutil.CopyFile(t, src, copy_)

	repo := testutil.BuildScanDB(t, tmp)
	svc := analysis.NewDeduplicationService(repo)
	groups, err := svc.DetectDuplicates(context.Background(), tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("want 1 duplicate group, got %d", len(groups))
	}
	if len(groups[0].Tracks) != 2 {
		t.Errorf("want 2 tracks in group, got %d", len(groups[0].Tracks))
	}
	// The non-copy track should be recommended.
	winner := trackByID(groups[0].Tracks, groups[0].RecommendedKeepID)
	if winner == nil {
		t.Fatal("recommended_keep_id not found in group tracks")
	}
	if filepath.Base(winner.Path) != "Nathan Epenge - Macpéla.mp3" {
		t.Errorf("expected clean file to be recommended, got %q", winner.Path)
	}
}

// ============================================================
// " (1)" suffix
// ============================================================

func TestCleanerIntegration_ParenthesisOneSuffix(t *testing.T) {
	src := pickSong(t, "Nathan Epenge - Macpéla.mp3")
	tmp := t.TempDir()

	orig := filepath.Join(tmp, "song.mp3")
	copy1 := filepath.Join(tmp, "song (1).mp3")
	testutil.CopyFile(t, src, orig)
	testutil.CopyFile(t, src, copy1)

	repo := testutil.BuildScanDB(t, tmp)
	svc := analysis.NewDeduplicationService(repo)
	groups, err := svc.DetectDuplicates(context.Background(), tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("want 1 group, got %d", len(groups))
	}
	winner := trackByID(groups[0].Tracks, groups[0].RecommendedKeepID)
	if winner == nil {
		t.Fatal("winner not found")
	}
	if filepath.Base(winner.Path) != "song.mp3" {
		t.Errorf("expected 'song.mp3' to be recommended, got %q", filepath.Base(winner.Path))
	}
}

// ============================================================
// French "Copie" mid-name must NOT be penalized; trailing " copy" must be
// ============================================================

func TestCleanerIntegration_FrenchCopieInNameNotPenalized(t *testing.T) {
	// Use any WMA file as binary content; we rename it to simulate the real pattern.
	src := pickAnyMP3(t)
	tmp := t.TempDir()

	// "Original" file: name contains " - Copie" (French noun, not a copy indicator)
	original := filepath.Join(tmp, "AKOUALA- Connais-tu - Copie.mp3")
	// Duplicate: same content, with trailing " copy"
	copyFile := filepath.Join(tmp, "AKOUALA- Connais-tu - Copie copy.mp3")
	testutil.CopyFile(t, src, original)
	testutil.CopyFile(t, src, copyFile)

	repo := testutil.BuildScanDB(t, tmp)
	svc := analysis.NewDeduplicationService(repo)
	groups, err := svc.DetectDuplicates(context.Background(), tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("want 1 group, got %d", len(groups))
	}
	winner := trackByID(groups[0].Tracks, groups[0].RecommendedKeepID)
	if winner == nil {
		t.Fatal("winner not found")
	}
	// The file WITHOUT trailing " copy" must be recommended.
	if filepath.Base(winner.Path) != "AKOUALA- Connais-tu - Copie.mp3" {
		t.Errorf("expected 'Copie.mp3' (original) to be recommended, got %q", filepath.Base(winner.Path))
	}
}

// ============================================================
// Three copies — cleanest path wins
// ============================================================

func TestCleanerIntegration_ThreeCopiesPicksCleanest(t *testing.T) {
	src := pickSong(t, "Nathan Epenge - Macpéla.mp3")
	tmp := t.TempDir()

	// Three byte-identical copies with different paths/names.
	orig := filepath.Join(tmp, "original.mp3")
	in1 := filepath.Join(tmp, "original (1).mp3")
	sub := filepath.Join(tmp, "subfolder", "original.mp3")
	testutil.CopyFile(t, src, orig)
	testutil.CopyFile(t, src, in1)
	testutil.CopyFile(t, src, sub)

	repo := testutil.BuildScanDB(t, tmp)
	svc := analysis.NewDeduplicationService(repo)
	groups, err := svc.DetectDuplicates(context.Background(), tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("want 1 group for 3 copies, got %d", len(groups))
	}
	if len(groups[0].Tracks) != 3 {
		t.Errorf("want 3 tracks in group, got %d", len(groups[0].Tracks))
	}
	winner := trackByID(groups[0].Tracks, groups[0].RecommendedKeepID)
	if winner == nil {
		t.Fatal("winner not found in group")
	}
	// "original.mp3" at root has shortest path AND no copy indicator.
	if filepath.Base(winner.Path) == "original (1).mp3" {
		t.Error("copy-suffix file must not be recommended")
	}
}

// ============================================================
// All distinct songs → 0 groups
// ============================================================

func TestCleanerIntegration_AllDistinctNoDuplicates(t *testing.T) {
	songs := testutil.TestSongsDir(t)
	entries, err := os.ReadDir(songs)
	if err != nil {
		t.Fatal(err)
	}

	tmp := t.TempDir()
	copied := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		testutil.CopyFileTo(t, filepath.Join(songs, e.Name()), tmp)
		copied++
	}
	if copied == 0 {
		t.Skip("no root-level audio files in test/songs")
	}

	repo := testutil.BuildScanDB(t, tmp)
	svc := analysis.NewDeduplicationService(repo)
	groups, err := svc.DetectDuplicates(context.Background(), tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 0 {
		t.Errorf("distinct files must produce 0 groups, got %d", len(groups))
	}
}

// ============================================================
// Cross-directory duplicate
// ============================================================

func TestCleanerIntegration_CrossDirectoryDuplicate(t *testing.T) {
	src := pickAnyMP3(t)
	tmp := t.TempDir()

	music := filepath.Join(tmp, "music")
	backup := filepath.Join(tmp, "backup")
	testutil.CopyFileAs(t, src, filepath.Join(music, "song.mp3"))
	testutil.CopyFileAs(t, src, filepath.Join(backup, "song.mp3"))

	repo := testutil.BuildScanDB(t, tmp)
	svc := analysis.NewDeduplicationService(repo)

	// Detect across both dirs → 1 group
	groups, err := svc.DetectDuplicates(context.Background(), tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Errorf("want 1 cross-directory group, got %d", len(groups))
	}

	// Filter to music/ only → 0 groups (only 1 track in that root)
	groupsFiltered, err := svc.DetectDuplicates(context.Background(), music)
	if err != nil {
		t.Fatal(err)
	}
	if len(groupsFiltered) != 0 {
		t.Errorf("root filter to music/: want 0 groups, got %d", len(groupsFiltered))
	}
}

// ============================================================
// Root scoping isolates two independent pairs
// ============================================================

func TestCleanerIntegration_RootScopingFiltersCorrectly(t *testing.T) {
	srcA := pickSong(t, "Nathan Epenge - Macpéla.mp3")
	srcB := pickAnyMP3(t)
	tmp := t.TempDir()

	// Pair in /lib/A
	testutil.CopyFileAs(t, srcA, filepath.Join(tmp, "A", "song.mp3"))
	testutil.CopyFileAs(t, srcA, filepath.Join(tmp, "A", "song copy.mp3"))
	// Pair in /lib/B
	testutil.CopyFileAs(t, srcB, filepath.Join(tmp, "B", "song.mp3"))
	testutil.CopyFileAs(t, srcB, filepath.Join(tmp, "B", "song (1).mp3"))

	repo := testutil.BuildScanDB(t, tmp)
	svc := analysis.NewDeduplicationService(repo)

	groupsA, err := svc.DetectDuplicates(context.Background(), filepath.Join(tmp, "A"))
	if err != nil {
		t.Fatal(err)
	}
	if len(groupsA) != 1 {
		t.Errorf("root A: want 1 group, got %d", len(groupsA))
	}

	groupsB, err := svc.DetectDuplicates(context.Background(), filepath.Join(tmp, "B"))
	if err != nil {
		t.Fatal(err)
	}
	if len(groupsB) != 1 {
		t.Errorf("root B: want 1 group, got %d", len(groupsB))
	}

	// Full scan: both pairs detected
	all, err := svc.DetectDuplicates(context.Background(), tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("full scan: want 2 groups, got %d", len(all))
	}
}

// ============================================================
// Helper
// ============================================================

func trackByID(tracks []domain.Track, id uint64) *domain.Track {
	for i := range tracks {
		if tracks[i].ID == id {
			return &tracks[i]
		}
	}
	return nil
}
