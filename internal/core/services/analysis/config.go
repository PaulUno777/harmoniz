package analysis

// ClusteringConfig holds tunable parameters for artist clustering.
type ClusteringConfig struct {
	// JaroWinklerThreshold: pairs with score >= this are considered similar (default 0.9).
	JaroWinklerThreshold float64
	// MaxLengthDiff: skip pair if |len(a)-len(b)| > this (default 3).
	MaxLengthDiff int
	// MaxBucketSize: skip buckets with more artists than this to avoid O(N²) blow-up (default 500).
	MaxBucketSize int
	// JaroWinklerP: prefix scaling factor (default 0.7).
	JaroWinklerP float64
	// JaroWinklerMaxPrefix: max prefix length for bonus (default 4).
	JaroWinklerMaxPrefix int
}

// DefaultClusteringConfig returns production defaults.
func DefaultClusteringConfig() ClusteringConfig {
	return ClusteringConfig{
		JaroWinklerThreshold: 0.9,
		MaxLengthDiff:        3,
		MaxBucketSize:        500,
		JaroWinklerP:         0.7,
		JaroWinklerMaxPrefix: 4,
	}
}

// DedupConfig holds tunable parameters for duplicate detection.
type DedupConfig struct {
	// MaxSizeGroupLog: log warning when a size-group exceeds this (default 500).
	MaxSizeGroupLog int
}

// DefaultDedupConfig returns production defaults.
func DefaultDedupConfig() DedupConfig {
	return DedupConfig{MaxSizeGroupLog: 500}
}
