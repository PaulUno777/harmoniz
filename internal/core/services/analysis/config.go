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

	// SimilarityThreshold: minimum Jaro-Winkler score to consider two titles/filenames similar (default 0.88).
	SimilarityThreshold float64
	// JaroWinklerP: prefix scaling factor for similarity detection (default 0.7).
	JaroWinklerP float64
	// JaroWinklerMaxPrefix: max prefix length for JW bonus in similarity detection (default 4).
	JaroWinklerMaxPrefix int
	// MaxSimilarBucketSize: skip buckets with more tracks than this to avoid O(N²) blow-up (default 500).
	MaxSimilarBucketSize int
	// MinSimilarKeyLen: skip keys shorter than this to avoid false positives (default 3).
	MinSimilarKeyLen int
}

// DefaultDedupConfig returns production defaults.
func DefaultDedupConfig() DedupConfig {
	return DedupConfig{
		MaxSizeGroupLog:      500,
		SimilarityThreshold:  0.88,
		JaroWinklerP:         0.7,
		JaroWinklerMaxPrefix: 4,
		MaxSimilarBucketSize: 500,
		MinSimilarKeyLen:     3,
	}
}
