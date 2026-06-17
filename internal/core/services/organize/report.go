package organize

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// OrganizeReport holds stats from a single AnalyzeRoot run for debug output.
type OrganizeReport struct {
	RunID      string      `json:"run_id"`
	Root       string      `json:"root"`
	At         string      `json:"at"`
	DurationMs int64       `json:"duration_ms"`
	Stats      ReportStats `json:"stats"`
	LowConfidenceWarnings []LowConfWarn `json:"low_confidence_warnings,omitempty"`
}

// ReportStats holds aggregate counters from an organize run.
type ReportStats struct {
	TracksAnalyzed       int `json:"tracks_analyzed"`
	SuggestionsGenerated int `json:"suggestions_generated"`
	HighConfidence       int `json:"high_confidence"`
	RegistryArtists      int `json:"registry_artists"`
}

// LowConfWarn marks a low-confidence suggestion for investigation.
type LowConfWarn struct {
	TrackID    uint64  `json:"track_id"`
	Field      string  `json:"field"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// IsDebugMode returns true when HARMONIZ_DEBUG=1 is set in the environment.
func IsDebugMode() bool {
	return os.Getenv("HARMONIZ_DEBUG") == "1"
}

// WriteDebugReport writes report to ~/.harmoniz/debug/organize-<ts>.json.
// Returns the path written, or an error.
func WriteDebugReport(report OrganizeReport) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	dir := filepath.Join(home, ".harmoniz", "debug")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", dir, err)
	}
	ts := time.Now().Format("20060102-150405")
	path := filepath.Join(dir, fmt.Sprintf("organize-%s.json", ts))
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, data, 0o644)
}
