package analysis

import (
	"context"
	"fmt"
	"harmoniz/internal/adapters/fs"
	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/ports"
	"harmoniz/internal/logger"
)

// DeduplicationService finds exact duplicate files via size → partial hash → full hash funnel.
type DeduplicationService struct {
	repo   ports.TrackRepository
	config DedupConfig
}

// NewDeduplicationService builds a dedup service with default config.
func NewDeduplicationService(repo ports.TrackRepository) *DeduplicationService {
	return NewDeduplicationServiceWithConfig(repo, DefaultDedupConfig())
}

// NewDeduplicationServiceWithConfig builds a dedup service with custom config.
func NewDeduplicationServiceWithConfig(repo ports.TrackRepository, cfg DedupConfig) *DeduplicationService {
	return &DeduplicationService{repo: repo, config: cfg}
}

// DetectDuplicates returns groups of tracks that are byte-identical (same full hash).
// root restricts to tracks under that path; empty = all.
func (s *DeduplicationService) DetectDuplicates(ctx context.Context, root string) ([][]domain.Track, error) {
	candidates, err := s.repo.GetDuplicateCandidates(ctx, root)
	if err != nil {
		return nil, fmt.Errorf("get duplicate candidates: %w", err)
	}

	logger.Log.Info("Deduplication: candidates found", "count", len(candidates), "root", root)

	bySize := make(map[int64][]domain.Track)
	for _, t := range candidates {
		if t.HashPartial == "" {
			continue
		}
		bySize[t.Size] = append(bySize[t.Size], t)
	}

	maxLog := s.config.MaxSizeGroupLog
	if maxLog <= 0 {
		maxLog = DefaultDedupConfig().MaxSizeGroupLog
	}

	var duplicates [][]domain.Track
	for size, group := range bySize {
		if len(group) > maxLog {
			logger.Log.Warn("Large size group", "size", size, "count", len(group))
		}
		byPartial := make(map[string][]domain.Track)
		for _, t := range group {
			byPartial[t.HashPartial] = append(byPartial[t.HashPartial], t)
		}
		for _, partialGroup := range byPartial {
			if len(partialGroup) < 2 {
				continue
			}
			byFull, err := s.groupByFullHash(ctx, partialGroup)
			if err != nil {
				logger.Log.Error("Group by full hash failed", "error", err)
				continue
			}
			for _, fullGroup := range byFull {
				if len(fullGroup) > 1 {
					duplicates = append(duplicates, fullGroup)
				}
			}
		}
	}
	return duplicates, nil
}

// groupByFullHash computes full hash for tracks that don't have it, then groups by hash.
func (s *DeduplicationService) groupByFullHash(ctx context.Context, group []domain.Track) (map[string][]domain.Track, error) {
	byFull := make(map[string][]domain.Track)
	for i := range group {
		t := &group[i]
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if t.HashFull == "" {
			full, err := fs.ComputeFullHash(t.Path)
			if err != nil {
				logger.Log.Error("Compute full hash failed", "path", t.Path, "error", err)
				continue
			}
			t.HashFull = full
		}
		byFull[t.HashFull] = append(byFull[t.HashFull], *t)
	}
	return byFull, nil
}
