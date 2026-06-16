package analysis

import (
	"context"
	"fmt"
	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/ports"
	"harmoniz/internal/logger"
	"strings"
	"unicode"

	"github.com/xrash/smetrics"
)

// ClusteringService finds similar artist names using Jaro-Winkler and suggests normalization.
type ClusteringService struct {
	repo   ports.TrackRepository
	config ClusteringConfig
}

// NewClusteringService builds a clustering service with default config.
func NewClusteringService(repo ports.TrackRepository) *ClusteringService {
	return NewClusteringServiceWithConfig(repo, DefaultClusteringConfig())
}

// NewClusteringServiceWithConfig builds a clustering service with custom config.
func NewClusteringServiceWithConfig(repo ports.TrackRepository, cfg ClusteringConfig) *ClusteringService {
	return &ClusteringService{repo: repo, config: cfg}
}

// AnalyzeArtists streams unique artists, buckets by first rune, and returns similarity suggestions.
// root restricts analysis to tracks under that path; empty = all.
func (s *ClusteringService) AnalyzeArtists(ctx context.Context, root string) ([]domain.ArtistSuggestion, error) {
	artistCh, err := s.repo.StreamUniqueArtists(ctx, root)
	if err != nil {
		return nil, fmt.Errorf("stream unique artists: %w", err)
	}

	buckets := make(map[rune][]string)
	var total int
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case artist, ok := <-artistCh:
			if !ok {
				goto done
			}
			artist = trim(artist)
			if artist == "" {
				continue
			}
			runes := []rune(artist)
			if len(runes) == 0 {
				continue
			}
			first := runes[0]
			if !unicode.IsLetter(first) && !unicode.IsNumber(first) {
				first = '?' // group non-letter/number into one bucket
			}
			buckets[first] = append(buckets[first], artist)
			total++
		}
	}
done:

	logger.Log.Info("Clustering artists", "total", total, "buckets", len(buckets))

	var suggestions []domain.ArtistSuggestion
	threshold := s.config.JaroWinklerThreshold
	if threshold <= 0 {
		threshold = DefaultClusteringConfig().JaroWinklerThreshold
	}
	maxLenDiff := s.config.MaxLengthDiff
	if maxLenDiff <= 0 {
		maxLenDiff = DefaultClusteringConfig().MaxLengthDiff
	}
	maxBucket := s.config.MaxBucketSize
	if maxBucket <= 0 {
		maxBucket = DefaultClusteringConfig().MaxBucketSize
	}
	p := s.config.JaroWinklerP
	if p <= 0 {
		p = DefaultClusteringConfig().JaroWinklerP
	}
	maxPrefix := s.config.JaroWinklerMaxPrefix
	if maxPrefix <= 0 {
		maxPrefix = DefaultClusteringConfig().JaroWinklerMaxPrefix
	}

	for _, group := range buckets {
		if len(group) < 2 {
			continue
		}
		if len(group) > maxBucket {
			logger.Log.Warn("Skipping large artist bucket", "size", len(group), "max", maxBucket)
			continue
		}
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				a, b := group[i], group[j]
				if abs(len(a)-len(b)) > maxLenDiff {
					continue
				}
				score := smetrics.JaroWinkler(a, b, p, maxPrefix)
				if score >= threshold {
					suggestions = append(suggestions, domain.ArtistSuggestion{
						Original:        b,
						Suggested:       a,
						Score:           score,
						Reason:          "Similar Name",
						ConfidenceLevel: confidenceFromScore(score),
					})
				}
			}
		}
	}
	return suggestions, nil
}

func trim(s string) string {
	return strings.TrimSpace(s)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func confidenceFromScore(score float64) string {
	if score >= 0.95 {
		return "High"
	}
	if score >= 0.9 {
		return "Medium"
	}
	return "Low"
}
