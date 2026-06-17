package analysis_test

import (
	"harmoniz/internal/core/services/analysis"
	"testing"
)

func TestClusteringConfigDefaults(t *testing.T) {
	cfg := analysis.DefaultClusteringConfig()
	if cfg.JaroWinklerThreshold <= 0 || cfg.JaroWinklerThreshold > 1 {
		t.Errorf("invalid default JaroWinklerThreshold: %f", cfg.JaroWinklerThreshold)
	}
	if cfg.MaxBucketSize <= 0 {
		t.Errorf("invalid default MaxBucketSize: %d", cfg.MaxBucketSize)
	}
	if cfg.MaxLengthDiff < 0 {
		t.Errorf("invalid default MaxLengthDiff: %d", cfg.MaxLengthDiff)
	}
}

func TestDedupConfigDefaults(t *testing.T) {
	cfg := analysis.DefaultDedupConfig()
	if cfg.MaxSizeGroupLog <= 0 {
		t.Errorf("invalid default MaxSizeGroupLog: %d", cfg.MaxSizeGroupLog)
	}
}
