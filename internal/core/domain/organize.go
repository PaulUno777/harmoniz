package domain

// FieldSuggestion is a suggested value for a single metadata field.
type FieldSuggestion struct {
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"` // 0.0–1.0
	Source     string  `json:"source"`     // "path", "filename", "neighbor"
}

// OrganizerSuggestion holds all inferred suggestions for a single track.
// Fields contains only keys where we have a meaningful suggestion.
type OrganizerSuggestion struct {
	TrackID uint64                    `json:"track_id"`
	Track   Track                     `json:"track"`
	Score   float64                   `json:"score"` // 0.0–100.0 overall confidence
	Fields  map[string]FieldSuggestion `json:"fields"`
	Issues  []string                  `json:"issues"`
}

// FilenamePreview holds the before/after filename pair for a template preview.
type FilenamePreview struct {
	TrackID uint64 `json:"track_id"`
	Before  string `json:"before"`
	After   string `json:"after"`
}
