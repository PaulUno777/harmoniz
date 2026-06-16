package domain

// ArtistSuggestion represents a suggested normalization for similar artist names
// (e.g. "guns n roses" → "Guns N' Roses").
type ArtistSuggestion struct {
	Original        string  `json:"original"`
	Suggested       string  `json:"suggested"`
	Score           float64 `json:"score"`             // Jaro-Winkler score (0.0–1.0)
	Reason          string  `json:"reason"`            // e.g. "Similar Name"
	ConfidenceLevel string  `json:"confidence_level"`   // "High", "Medium", "Low"
}
