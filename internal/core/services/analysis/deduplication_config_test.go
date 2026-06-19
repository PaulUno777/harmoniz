package analysis

// Tests in this file prove that the SimilarityThreshold config field is
// actually consulted during detection — addressing the "set values but they
// are not used" regression.

import (
	"context"
	"testing"

	"harmoniz/internal/core/domain"
	"harmoniz/internal/testutil"
)

// twoSimilarTitleTracks returns two tracks with similar (but non-identical)
// titles that have no HashPartial, so exact detection is skipped entirely.
// They are expected to group at the default threshold (0.88) and above their
// JW score, but not at threshold = 1.0.
func twoSimilarTitleTracks() (domain.Track, domain.Track) {
	return domain.Track{
		ID:       1,
		Path:     "/music/yesterday.mp3",
		Filename: "yesterday.mp3",
		Title:    "Yesterday",
	}, domain.Track{
		ID:       2,
		Path:     "/music/yesterdaay.mp3",
		Filename: "yesterdaay.mp3",
		Title:    "Yesterdaay", // one repeated 'a' — similar but not identical
	}
}

// TestDedup_DefaultThreshold_SimilarTitlesGrouped verifies that two tracks
// with similar-but-not-identical titles are flagged at the default threshold.
func TestDedup_DefaultThreshold_SimilarTitlesGrouped(t *testing.T) {
	a, b := twoSimilarTitleTracks()
	repo := testutil.NewFakeRepo(a, b)

	svc := NewDeduplicationServiceWithConfig(repo, DefaultDedupConfig())
	groups, err := svc.DetectDuplicates(context.Background(), "")
	if err != nil {
		t.Fatalf("DetectDuplicates: %v", err)
	}

	titleGroups := filterKind(groups, domain.DuplicateKindSimilarTitle)
	if len(titleGroups) != 1 {
		t.Fatalf("want 1 similar-title group at default threshold, got %d", len(titleGroups))
	}
	if len(titleGroups[0].Tracks) != 2 {
		t.Errorf("want 2 tracks in group, got %d", len(titleGroups[0].Tracks))
	}
}

// TestDedup_StrictThreshold_SimilarTitlesNotGrouped proves that raising
// SimilarityThreshold to 1.0 suppresses the same pair — the config IS read.
func TestDedup_StrictThreshold_SimilarTitlesNotGrouped(t *testing.T) {
	a, b := twoSimilarTitleTracks()
	repo := testutil.NewFakeRepo(a, b)

	cfg := DefaultDedupConfig()
	cfg.SimilarityThreshold = 1.0 // only byte-identical titles would match (impossible after normSimilar)

	svc := NewDeduplicationServiceWithConfig(repo, cfg)
	groups, err := svc.DetectDuplicates(context.Background(), "")
	if err != nil {
		t.Fatalf("DetectDuplicates: %v", err)
	}

	titleGroups := filterKind(groups, domain.DuplicateKindSimilarTitle)
	if len(titleGroups) != 0 {
		t.Errorf("want 0 similar-title groups at threshold=1.0, got %d", len(titleGroups))
	}
}

// TestDedup_ThresholdDeterminesGrouping is the key regression test:
// it runs the same pair through two services with different thresholds and
// asserts opposite outcomes, proving that SimilarityThreshold from DedupConfig
// is the actual gate, not a hardcoded constant.
func TestDedup_ThresholdDeterminesGrouping(t *testing.T) {
	a, b := twoSimilarTitleTracks()

	loose := DefaultDedupConfig()
	loose.SimilarityThreshold = 0.88 // default — should group

	strict := DefaultDedupConfig()
	strict.SimilarityThreshold = 1.0 // maximum — should not group

	repoLoose := testutil.NewFakeRepo(a, b)
	repoStrict := testutil.NewFakeRepo(a, b)

	groupsLoose, err := NewDeduplicationServiceWithConfig(repoLoose, loose).DetectDuplicates(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	groupsStrict, err := NewDeduplicationServiceWithConfig(repoStrict, strict).DetectDuplicates(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}

	looseCount := len(filterKind(groupsLoose, domain.DuplicateKindSimilarTitle))
	strictCount := len(filterKind(groupsStrict, domain.DuplicateKindSimilarTitle))

	if looseCount == 0 {
		t.Error("loose threshold (0.88): expected at least 1 similar-title group")
	}
	if strictCount != 0 {
		t.Errorf("strict threshold (1.0): expected 0 groups, got %d", strictCount)
	}
	if looseCount == strictCount {
		t.Error("threshold had no effect — config is not being used")
	}
}

func filterKind(groups []domain.DuplicateGroup, kind string) []domain.DuplicateGroup {
	var out []domain.DuplicateGroup
	for _, g := range groups {
		if g.Kind == kind {
			out = append(out, g)
		}
	}
	return out
}
