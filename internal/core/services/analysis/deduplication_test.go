package analysis

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"harmoniz/internal/core/domain"
	"harmoniz/internal/logger"
	"harmoniz/internal/testutil"
)

func TestMain(m *testing.M) {
	logger.Init(slog.LevelError)
	os.Exit(m.Run())
}

// ============================================================
// QualityScore — pure function, no repo needed
// ============================================================

func TestQualityScore_RichTagsWin(t *testing.T) {
	rich := domain.Track{
		Path:      "/music/song.mp3",
		Title:     "Song",
		ArtistRaw: "Artist",
		AlbumRaw:  "Album",
		Year:      2020,
		TrackNum:  1,
	}
	bare := domain.Track{Path: "/music/other.mp3"}

	if QualityScore(rich) <= QualityScore(bare) {
		t.Errorf("rich track (%.2f) should score higher than bare track (%.2f)",
			QualityScore(rich), QualityScore(bare))
	}
}

func TestQualityScore_CopyPathPenalized(t *testing.T) {
	clean := domain.Track{Path: "/music/song.mp3", Title: "Song"}
	copy1 := domain.Track{Path: "/music/song (1).mp3", Title: "Song"}

	if QualityScore(copy1) >= QualityScore(clean) {
		t.Errorf("copy track (%.2f) should score lower than clean track (%.2f)",
			QualityScore(copy1), QualityScore(clean))
	}
	diff := QualityScore(clean) - QualityScore(copy1)
	if diff < 4.5 {
		t.Errorf("penalty should be close to 5, got diff=%.2f", diff)
	}
}

func TestQualityScore_HighBitrateWins(t *testing.T) {
	hi := domain.Track{Path: "/music/hi.mp3", Bitrate: 320}
	lo := domain.Track{Path: "/music/lo.mp3", Bitrate: 128}

	if QualityScore(hi) <= QualityScore(lo) {
		t.Errorf("high bitrate (%.2f) should beat low bitrate (%.2f)",
			QualityScore(hi), QualityScore(lo))
	}
}

func TestQualityScore_ShortPathWins(t *testing.T) {
	short := domain.Track{Path: "/music/song.mp3"}
	long := domain.Track{Path: "/music/deeply/nested/subfolder/collection/song.mp3"}

	if QualityScore(short) <= QualityScore(long) {
		t.Errorf("short path (%.2f) should beat long path (%.2f)",
			QualityScore(short), QualityScore(long))
	}
}

func TestQualityScore_CopyRegexVariants(t *testing.T) {
	// All of these must be penalized (-5 pts) relative to the clean version.
	cleanPath := "/music/song.mp3"
	cleanScore := QualityScore(domain.Track{Path: cleanPath})

	copyPaths := []string{
		"/music/001 Recois l'adoration copy.mp3",   // " copy" suffix
		"/music/01 PÈLERIN copy.mp3",               // " copy" suffix, accented
		"/music/amen (1).mp3",                      // " (1)" suffix
		"/music/AKOUALA- Aide-moi Seigneur! copy.mp3", // " copy" after "!"
	}
	for _, p := range copyPaths {
		score := QualityScore(domain.Track{Path: p})
		if score >= cleanScore {
			t.Errorf("path %q (%.2f) should be penalized vs clean (%.2f)", p, score, cleanScore)
		}
	}
}

func TestQualityScore_FrenchCopieNotPenalized(t *testing.T) {
	// "Copie" (French noun meaning "copy") embedded mid-name must NOT trigger the penalty.
	// The actual copy is "... - Copie copy.mp3"; the original is "... - Copie.mp3".
	original := domain.Track{Path: "/music/AKOUALA- Connais-tu - Copie.mp3"}
	copy_ := domain.Track{Path: "/music/AKOUALA- Connais-tu - Copie copy.mp3"}

	scoreOrig := QualityScore(original)
	scoreCopy := QualityScore(copy_)

	if scoreCopy >= scoreOrig {
		t.Errorf("copy variant (%.2f) should score lower than original with 'Copie' in name (%.2f)",
			scoreCopy, scoreOrig)
	}
}

func TestQualityScore_MidWordParenthesisNotPenalized(t *testing.T) {
	// "_(1)_" embedded mid-filename must NOT be penalized as a copy indicator.
	// The only score difference should be the tiny path-length penalty (0.01 per char).
	// A copy penalty is -5, far larger than any path-length difference between these paths.
	track := domain.Track{Path: "/music/Symphony_(1)_of_life.mp3"}
	clean := domain.Track{Path: "/music/Symphony_of_life.mp3"}
	diff := QualityScore(clean) - QualityScore(track)
	if diff > 1.0 {
		t.Errorf("mid-word _(1)_ must not trigger -5 copy penalty: score diff=%.2f (expected ≤ 1.0 for path-length only)", diff)
	}
}

func TestQualityScore_YouTubeUnderscoreSuffixNotPenalized(t *testing.T) {
	// YouTube downloads append "_(0)" with an underscore prefix — these are originals.
	// The copy detector must not penalize them.
	youtube := domain.Track{Path: "/music/Alex_Kimunga_chante___À_TES_PIEDS_(0).mp3"}
	clean := domain.Track{Path: "/music/Alex_Kimunga_chante___À_TES_PIEDS.mp3"}
	diff := QualityScore(clean) - QualityScore(youtube)
	if diff > 1.0 {
		t.Errorf("YouTube _(0) suffix must not trigger copy penalty: score diff=%.2f", diff)
	}
}

// ============================================================
// DetectDuplicates — via FakeRepo with pre-populated HashFull
// ============================================================

const (
	hashPartial = "abcdef123"
	hashFull    = "deadbeef0011223344"
	fileSize    = int64(5_000_000)
)

func trkDup(id uint64, path string, extra ...func(*testutil.TrackBuilder)) domain.Track {
	b := testutil.NewTrack(id, path).
		WithSize(fileSize).
		WithHashes(hashPartial, hashFull)
	for _, fn := range extra {
		fn(b)
	}
	return b.Build()
}

func TestDetectDuplicates_ByteIdenticalDetected(t *testing.T) {
	tracks := []domain.Track{
		trkDup(1, "/lib/a.mp3"),
		trkDup(2, "/lib/b.mp3"),
	}
	svc := NewDeduplicationService(testutil.NewFakeRepo(tracks...))
	groups, err := svc.DetectDuplicates(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("want 1 group, got %d", len(groups))
	}
	if len(groups[0].Tracks) != 2 {
		t.Errorf("want 2 tracks in group, got %d", len(groups[0].Tracks))
	}
}

func TestDetectDuplicates_DifferentSizeNotGrouped(t *testing.T) {
	tracks := []domain.Track{
		testutil.NewTrack(1, "/lib/a.mp3").WithSize(1000).WithHashes(hashPartial, hashFull).Build(),
		testutil.NewTrack(2, "/lib/b.mp3").WithSize(2000).WithHashes(hashPartial, hashFull).Build(),
	}
	svc := NewDeduplicationService(testutil.NewFakeRepo(tracks...))
	groups, err := svc.DetectDuplicates(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 0 {
		t.Errorf("different sizes must not group as duplicates, got %d groups", len(groups))
	}
}

func TestDetectDuplicates_SameSizeDifferentPartialHash(t *testing.T) {
	tracks := []domain.Track{
		testutil.NewTrack(1, "/lib/a.mp3").WithSize(fileSize).WithHashes("abc", "full1").Build(),
		testutil.NewTrack(2, "/lib/b.mp3").WithSize(fileSize).WithHashes("xyz", "full2").Build(),
	}
	svc := NewDeduplicationService(testutil.NewFakeRepo(tracks...))
	groups, err := svc.DetectDuplicates(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 0 {
		t.Errorf("different partial hash must not group, got %d groups", len(groups))
	}
}

func TestDetectDuplicates_SameSizeSamePartialDifferentFull(t *testing.T) {
	// Same size and partial hash but different full hash → NOT duplicates.
	tracks := []domain.Track{
		testutil.NewTrack(1, "/lib/a.mp3").WithSize(fileSize).WithHashes(hashPartial, "full-A").Build(),
		testutil.NewTrack(2, "/lib/b.mp3").WithSize(fileSize).WithHashes(hashPartial, "full-B").Build(),
	}
	svc := NewDeduplicationService(testutil.NewFakeRepo(tracks...))
	groups, err := svc.DetectDuplicates(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 0 {
		t.Errorf("different full hash must not group, got %d groups", len(groups))
	}
}

func TestDetectDuplicates_RecommendedKeepHigherQuality(t *testing.T) {
	// Track A: full tags, clean path → higher quality score.
	// Track B: no tags, copy suffix in path → lower quality score.
	trackA := testutil.NewTrack(1, "/music/song.mp3").
		WithSize(fileSize).WithHashes(hashPartial, hashFull).
		WithArtist("Artist").WithTitle("Song").WithAlbum("Album").
		WithYear(2020).WithTrackNum(1).WithBitrate(320).
		Build()
	trackB := testutil.NewTrack(2, "/music/song (1).mp3").
		WithSize(fileSize).WithHashes(hashPartial, hashFull).
		Build()

	svc := NewDeduplicationService(testutil.NewFakeRepo(trackA, trackB))
	groups, err := svc.DetectDuplicates(context.Background(), "/music")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("want 1 group, got %d", len(groups))
	}
	if groups[0].RecommendedKeepID != 1 {
		t.Errorf("want recommended_keep_id=1, got %d", groups[0].RecommendedKeepID)
	}
}

func TestDetectDuplicates_ThreeCopiesPicksBest(t *testing.T) {
	// Three identical files; C has highest bitrate → recommended keeper.
	a := testutil.NewTrack(1, "/music/song.mp3").WithSize(fileSize).WithHashes(hashPartial, hashFull).Build()
	b := testutil.NewTrack(2, "/music/song copy.mp3").WithSize(fileSize).WithHashes(hashPartial, hashFull).Build()
	c := testutil.NewTrack(3, "/music/hq/song.mp3").WithSize(fileSize).WithHashes(hashPartial, hashFull).
		WithBitrate(320).WithTitle("Song").WithArtist("Artist").WithAlbum("Album").Build()

	svc := NewDeduplicationService(testutil.NewFakeRepo(a, b, c))
	groups, err := svc.DetectDuplicates(context.Background(), "/music")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("want 1 group, got %d", len(groups))
	}
	if groups[0].RecommendedKeepID != 3 {
		t.Errorf("want recommended_keep_id=3 (highest quality), got %d", groups[0].RecommendedKeepID)
	}
}

func TestDetectDuplicates_TwoSeparateGroups(t *testing.T) {
	// Two independent pairs of duplicates.
	pair1 := []domain.Track{
		testutil.NewTrack(1, "/lib/a1.mp3").WithSize(1000).WithHashes("hash1partial", "hash1full").Build(),
		testutil.NewTrack(2, "/lib/a2.mp3").WithSize(1000).WithHashes("hash1partial", "hash1full").Build(),
	}
	pair2 := []domain.Track{
		testutil.NewTrack(3, "/lib/b1.mp3").WithSize(2000).WithHashes("hash2partial", "hash2full").Build(),
		testutil.NewTrack(4, "/lib/b2.mp3").WithSize(2000).WithHashes("hash2partial", "hash2full").Build(),
	}
	svc := NewDeduplicationService(testutil.NewFakeRepo(append(pair1, pair2...)...))
	groups, err := svc.DetectDuplicates(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Errorf("want 2 groups, got %d", len(groups))
	}
}

func TestDetectDuplicates_EmptyHashPartialSkipped(t *testing.T) {
	// A track with no HashPartial must be excluded from duplicate analysis.
	tracks := []domain.Track{
		testutil.NewTrack(1, "/lib/a.mp3").WithSize(fileSize).Build(), // no partial hash
		testutil.NewTrack(2, "/lib/b.mp3").WithSize(fileSize).Build(), // no partial hash
	}
	svc := NewDeduplicationService(testutil.NewFakeRepo(tracks...))
	groups, err := svc.DetectDuplicates(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 0 {
		t.Errorf("tracks without HashPartial must not form groups, got %d", len(groups))
	}
}

func TestDetectDuplicates_RootFilter(t *testing.T) {
	// Track A and B are byte-identical but in different roots.
	tA := trkDup(1, "/lib/music/a.mp3")
	tB := trkDup(2, "/lib/backup/a.mp3")
	svc := NewDeduplicationService(testutil.NewFakeRepo(tA, tB))

	// Filter to /lib/music → only 1 track matches → 0 groups
	groups, err := svc.DetectDuplicates(context.Background(), "/lib/music")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 0 {
		t.Errorf("root filter to /lib/music: want 0 groups, got %d", len(groups))
	}

	// No filter → both tracks found → 1 group
	groups, err = svc.DetectDuplicates(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Errorf("no root filter: want 1 group, got %d", len(groups))
	}
}

func TestDetectDuplicates_EmptyRepository(t *testing.T) {
	svc := NewDeduplicationService(testutil.NewFakeRepo())
	groups, err := svc.DetectDuplicates(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 0 {
		t.Errorf("want 0 groups from empty repo, got %d", len(groups))
	}
}
