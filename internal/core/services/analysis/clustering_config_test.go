package analysis_test

// Tests that JaroWinklerThreshold from ClusteringConfig is actually consulted
// during AnalyzeArtists — same regression check as deduplication_config_test.go.

import (
	"context"
	"testing"

	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/services/analysis"
	"harmoniz/internal/testutil"
)

func artistTracks() (domain.Track, domain.Track) {
	return domain.Track{
		ID:         1,
		Path:       "/music/a.mp3",
		ArtistNorm: "The Beatles",
	}, domain.Track{
		ID:         2,
		Path:       "/music/b.mp3",
		ArtistNorm: "The Beatlez", // one char off
	}
}

// TestClustering_DefaultThreshold_SimilarArtistsGrouped verifies that two
// similar artist names produce a suggestion at the default threshold.
func TestClustering_DefaultThreshold_SimilarArtistsGrouped(t *testing.T) {
	a, b := artistTracks()
	repo := testutil.NewFakeRepo(a, b)

	svc := analysis.NewClusteringServiceWithConfig(repo, analysis.DefaultClusteringConfig())
	suggestions, err := svc.AnalyzeArtists(context.Background(), "")
	if err != nil {
		t.Fatalf("AnalyzeArtists: %v", err)
	}
	if len(suggestions) == 0 {
		t.Error("want ≥1 suggestion at default threshold, got 0")
	}
}

// TestClustering_StrictThreshold_SimilarArtistsNotGrouped verifies that
// raising the threshold to 1.0 suppresses the suggestion — the config IS read.
func TestClustering_StrictThreshold_SimilarArtistsNotGrouped(t *testing.T) {
	a, b := artistTracks()
	repo := testutil.NewFakeRepo(a, b)

	cfg := analysis.DefaultClusteringConfig()
	cfg.JaroWinklerThreshold = 1.0

	svc := analysis.NewClusteringServiceWithConfig(repo, cfg)
	suggestions, err := svc.AnalyzeArtists(context.Background(), "")
	if err != nil {
		t.Fatalf("AnalyzeArtists: %v", err)
	}
	if len(suggestions) != 0 {
		t.Errorf("want 0 suggestions at threshold=1.0, got %d", len(suggestions))
	}
}

// TestClustering_ThresholdDeterminesGrouping is the core regression guard:
// identical input, different threshold → opposite outcomes.
func TestClustering_ThresholdDeterminesGrouping(t *testing.T) {
	a, b := artistTracks()

	loose := analysis.DefaultClusteringConfig()
	loose.JaroWinklerThreshold = 0.90 // default

	strict := analysis.DefaultClusteringConfig()
	strict.JaroWinklerThreshold = 1.0

	repoLoose := testutil.NewFakeRepo(a, b)
	repoStrict := testutil.NewFakeRepo(a, b)

	sugLoose, err := analysis.NewClusteringServiceWithConfig(repoLoose, loose).AnalyzeArtists(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	sugStrict, err := analysis.NewClusteringServiceWithConfig(repoStrict, strict).AnalyzeArtists(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}

	if len(sugLoose) == 0 {
		t.Error("loose threshold (0.90): expected at least 1 suggestion")
	}
	if len(sugStrict) != 0 {
		t.Errorf("strict threshold (1.0): expected 0 suggestions, got %d", len(sugStrict))
	}
	if len(sugLoose) == len(sugStrict) {
		t.Error("JaroWinklerThreshold had no effect — config is not being used")
	}
}
