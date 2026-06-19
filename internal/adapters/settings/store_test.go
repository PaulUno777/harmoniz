package settings_test

import (
	"os"
	"path/filepath"
	"testing"

	"harmoniz/internal/adapters/settings"
	"harmoniz/internal/core/domain"
)

func TestStore_DefaultsWhenFileAbsent(t *testing.T) {
	store := settings.NewStore(t.TempDir())
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	def := domain.DefaultAppConfig()
	if cfg.Cleaner.SimilarityThreshold != def.Cleaner.SimilarityThreshold {
		t.Errorf("SimilarityThreshold: want %v, got %v", def.Cleaner.SimilarityThreshold, cfg.Cleaner.SimilarityThreshold)
	}
	if cfg.Cleaner.ArtistGroupThreshold != def.Cleaner.ArtistGroupThreshold {
		t.Errorf("ArtistGroupThreshold: want %v, got %v", def.Cleaner.ArtistGroupThreshold, cfg.Cleaner.ArtistGroupThreshold)
	}
	if cfg.Organizer.FieldConfidenceThreshold != def.Organizer.FieldConfidenceThreshold {
		t.Errorf("FieldConfidenceThreshold: want %v, got %v", def.Organizer.FieldConfidenceThreshold, cfg.Organizer.FieldConfidenceThreshold)
	}
}

func TestStore_RoundTrip(t *testing.T) {
	store := settings.NewStore(t.TempDir())

	want := domain.AppConfig{
		Cleaner: domain.CleanerAlgoConfig{
			SimilarityThreshold:  0.72,
			ArtistGroupThreshold: 0.85,
		},
		Organizer: domain.OrganizerAlgoConfig{
			FieldConfidenceThreshold: 0.60,
		},
	}
	if err := store.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Cleaner.SimilarityThreshold != want.Cleaner.SimilarityThreshold {
		t.Errorf("SimilarityThreshold: want %v, got %v", want.Cleaner.SimilarityThreshold, got.Cleaner.SimilarityThreshold)
	}
	if got.Cleaner.ArtistGroupThreshold != want.Cleaner.ArtistGroupThreshold {
		t.Errorf("ArtistGroupThreshold: want %v, got %v", want.Cleaner.ArtistGroupThreshold, got.Cleaner.ArtistGroupThreshold)
	}
	if got.Organizer.FieldConfidenceThreshold != want.Organizer.FieldConfidenceThreshold {
		t.Errorf("FieldConfidenceThreshold: want %v, got %v", want.Organizer.FieldConfidenceThreshold, got.Organizer.FieldConfidenceThreshold)
	}
}

func TestStore_OverwritePreservesNewValues(t *testing.T) {
	dir := t.TempDir()
	store := settings.NewStore(dir)

	first := domain.AppConfig{
		Cleaner:   domain.CleanerAlgoConfig{SimilarityThreshold: 0.80, ArtistGroupThreshold: 0.80},
		Organizer: domain.OrganizerAlgoConfig{FieldConfidenceThreshold: 0.80},
	}
	if err := store.Save(first); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	second := domain.AppConfig{
		Cleaner:   domain.CleanerAlgoConfig{SimilarityThreshold: 0.95, ArtistGroupThreshold: 0.95},
		Organizer: domain.OrganizerAlgoConfig{FieldConfidenceThreshold: 0.95},
	}
	if err := store.Save(second); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Cleaner.SimilarityThreshold != 0.95 {
		t.Errorf("want 0.95, got %v", got.Cleaner.SimilarityThreshold)
	}
}

func TestStore_CorruptFileFallsBackToDefaults(t *testing.T) {
	dir := t.TempDir()
	// Write garbage JSON directly to the config path.
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("not-json{{{"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := settings.NewStore(dir)
	// Load returns (DefaultAppConfig, non-nil error) for corrupt JSON.
	// The caller (app.go) ignores the error; we verify the returned config is safe defaults.
	cfg, _ := store.Load()
	def := domain.DefaultAppConfig()
	if cfg.Cleaner.SimilarityThreshold != def.Cleaner.SimilarityThreshold {
		t.Errorf("corrupt file: want default SimilarityThreshold %v, got %v", def.Cleaner.SimilarityThreshold, cfg.Cleaner.SimilarityThreshold)
	}
	if cfg.Organizer.FieldConfidenceThreshold != def.Organizer.FieldConfidenceThreshold {
		t.Errorf("corrupt file: want default FieldConfidenceThreshold %v, got %v", def.Organizer.FieldConfidenceThreshold, cfg.Organizer.FieldConfidenceThreshold)
	}
}
