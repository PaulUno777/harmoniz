package domain

// AppConfig holds user-configurable algorithm precision settings persisted to disk.
type AppConfig struct {
	Cleaner   CleanerAlgoConfig   `json:"cleaner"`
	Organizer OrganizerAlgoConfig `json:"organizer"`
}

// CleanerAlgoConfig controls the deduplication and artist-clustering algorithms.
type CleanerAlgoConfig struct {
	// SimilarityThreshold is the minimum Jaro-Winkler score to group similar titles/filenames (0.0–1.0).
	SimilarityThreshold float64 `json:"similarity_threshold"`
	// ArtistGroupThreshold is the minimum score to consider two artist names the same (0.0–1.0).
	ArtistGroupThreshold float64 `json:"artist_group_threshold"`
}

// OrganizerAlgoConfig controls the suggestion engine.
type OrganizerAlgoConfig struct {
	// FieldConfidenceThreshold is the minimum per-field confidence to auto-apply a suggestion (0.0–1.0).
	FieldConfidenceThreshold float64 `json:"field_confidence_threshold"`
}

// DefaultAppConfig returns the production defaults.
func DefaultAppConfig() AppConfig {
	return AppConfig{
		Cleaner: CleanerAlgoConfig{
			SimilarityThreshold:  0.88,
			ArtistGroupThreshold: 0.90,
		},
		Organizer: OrganizerAlgoConfig{
			FieldConfidenceThreshold: 0.75,
		},
	}
}
