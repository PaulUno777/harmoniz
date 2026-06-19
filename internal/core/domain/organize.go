package domain

// TraceStep records one inference step in the artist decision trace.
type TraceStep struct {
	Step       string  `json:"step"`
	Input      string  `json:"input"`
	Result     string  `json:"result"`
	Confidence float64 `json:"confidence"`
	Rejected   bool    `json:"rejected,omitempty"`
	Reason     string  `json:"reason,omitempty"`
}

// FieldSuggestion is a suggested value for a single metadata field.
type FieldSuggestion struct {
	Value      string      `json:"value"`
	Confidence float64     `json:"confidence"` // 0.0–1.0
	Source     string      `json:"source"`     // "path", "filename", "neighbor", "library"
	Trace      []TraceStep `json:"trace,omitempty"`
}

// OrganizerSuggestion holds all inferred suggestions for a single track.
// Fields contains only keys where we have a meaningful suggestion.
type OrganizerSuggestion struct {
	TrackID uint64                     `json:"track_id"`
	Track   Track                      `json:"track"`
	Score   float64                    `json:"score"` // 0.0–100.0 overall confidence
	Fields  map[string]FieldSuggestion `json:"fields"`
	Issues  []string                   `json:"issues"`
}

// FilenamePreview holds the before/after filename pair for a template preview.
type FilenamePreview struct {
	TrackID uint64 `json:"track_id"`
	Before  string `json:"before"`
	After   string `json:"after"`
}

// DuplicateGroup holds a set of tracks that are identical or similar, with a recommended track to keep.
type DuplicateGroup struct {
	Kind              string  `json:"kind"`               // "exact", "similar_title", "similar_filename"
	Tracks            []Track `json:"tracks"`
	RecommendedKeepID uint64  `json:"recommended_keep_id"`
}

const (
	DuplicateKindExact           = "exact"
	DuplicateKindSimilarTitle    = "similar_title"
	DuplicateKindSimilarFilename = "similar_filename"
)
