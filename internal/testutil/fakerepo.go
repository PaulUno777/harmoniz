package testutil

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"harmoniz/internal/core/domain"
)

// FakeRepo is an in-memory TrackRepository for unit tests.
// It is safe for concurrent use within a single test.
type FakeRepo struct {
	mu     sync.Mutex
	Tracks []domain.Track
	nextID uint64
}

// NewFakeRepo creates a FakeRepo pre-loaded with the given tracks.
// Tracks with ID=0 are assigned auto-increment IDs starting at 1.
func NewFakeRepo(tracks ...domain.Track) *FakeRepo {
	r := &FakeRepo{}
	atomic.StoreUint64(&r.nextID, 1)
	for _, t := range tracks {
		r.addTrack(t)
	}
	return r
}

func (r *FakeRepo) addTrack(t domain.Track) {
	if t.ID == 0 {
		t.ID = atomic.AddUint64(&r.nextID, 1) - 1
	}
	if t.Filename == "" && t.Path != "" {
		t.Filename = filepath.Base(t.Path)
	}
	if t.Status == "" {
		t.Status = domain.StatusClean
	}
	r.Tracks = append(r.Tracks, t)
}

// ListTracks returns all non-deleted tracks matching the filter.
func (r *FakeRepo) ListTracks(_ context.Context, filter domain.TrackFilter) ([]domain.Track, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []domain.Track
	for _, t := range r.Tracks {
		if t.IsDeleted {
			continue
		}
		if filter.Root != "" {
			cleanRoot := filepath.Clean(filter.Root)
			cleanPath := filepath.Clean(t.Path)
			if cleanPath != cleanRoot && !strings.HasPrefix(cleanPath, cleanRoot+string(filepath.Separator)) {
				continue
			}
		}
		result = append(result, t)
	}
	total := len(result)
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, total, nil
}

// GetDuplicateCandidates returns non-deleted tracks under root with a non-empty
// HashPartial where the same Size appears in ≥2 tracks (mirrors the real SQL query).
func (r *FakeRepo) GetDuplicateCandidates(_ context.Context, root string) ([]domain.Track, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var under []domain.Track
	for _, t := range r.Tracks {
		if t.IsDeleted || t.HashPartial == "" {
			continue
		}
		if root != "" {
			cleanRoot := filepath.Clean(root)
			cleanPath := filepath.Clean(t.Path)
			if cleanPath != cleanRoot && !strings.HasPrefix(cleanPath, cleanRoot+string(filepath.Separator)) {
				continue
			}
		}
		under = append(under, t)
	}

	// Only return tracks whose size is shared by ≥2 candidates.
	sizeCounts := make(map[int64]int, len(under))
	for _, t := range under {
		sizeCounts[t.Size]++
	}
	result := under[:0]
	for _, t := range under {
		if sizeCounts[t.Size] >= 2 {
			result = append(result, t)
		}
	}
	return result, nil
}

// StreamUniqueArtists streams unique non-empty ArtistNorm values for tracks under root.
func (r *FakeRepo) StreamUniqueArtists(_ context.Context, root string) (<-chan string, error) {
	r.mu.Lock()
	tracks := make([]domain.Track, len(r.Tracks))
	copy(tracks, r.Tracks)
	r.mu.Unlock()

	ch := make(chan string, len(tracks))
	go func() {
		defer close(ch)
		seen := make(map[string]bool)
		for _, t := range tracks {
			if t.IsDeleted {
				continue
			}
			if root != "" {
				cleanRoot := filepath.Clean(root)
				cleanPath := filepath.Clean(t.Path)
				if cleanPath != cleanRoot && !strings.HasPrefix(cleanPath, cleanRoot+string(filepath.Separator)) {
					continue
				}
			}
			if t.ArtistNorm != "" && !seen[t.ArtistNorm] {
				seen[t.ArtistNorm] = true
				ch <- t.ArtistNorm
			}
		}
	}()
	return ch, nil
}

// BatchUpsert stores tracks: inserts new ones (by Path) and updates existing ones.
func (r *FakeRepo) BatchUpsert(tracks []domain.Track) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	pathIndex := make(map[string]int, len(r.Tracks))
	for i, t := range r.Tracks {
		pathIndex[t.Path] = i
	}
	for _, t := range tracks {
		if t.Filename == "" && t.Path != "" {
			t.Filename = filepath.Base(t.Path)
		}
		if idx, ok := pathIndex[t.Path]; ok {
			r.Tracks[idx] = t
		} else {
			if t.ID == 0 {
				t.ID = atomic.AddUint64(&r.nextID, 1) - 1
			}
			r.Tracks = append(r.Tracks, t)
			pathIndex[t.Path] = len(r.Tracks) - 1
		}
	}
	return nil
}

// GetAllPathsModTime returns path→modTime for all active tracks.
func (r *FakeRepo) GetAllPathsModTime() (map[string]int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := make(map[string]int64, len(r.Tracks))
	for _, t := range r.Tracks {
		if !t.IsDeleted {
			m[t.Path] = t.ModTime
		}
	}
	return m, nil
}

// UpdateTrackPath is a no-op stub (tests verify logic, not persistence).
func (r *FakeRepo) UpdateTrackPath(_ context.Context, id uint64, newPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, t := range r.Tracks {
		if t.ID == id {
			r.Tracks[i].Path = newPath
			r.Tracks[i].Filename = filepath.Base(newPath)
			return nil
		}
	}
	return nil
}

// UpdateTrackTags is a no-op stub.
func (r *FakeRepo) UpdateTrackTags(_ context.Context, _ uint64, _, _, _ string, _, _ int) error {
	return nil
}

// SoftDelete marks a track as deleted.
func (r *FakeRepo) SoftDelete(_ context.Context, id uint64, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, t := range r.Tracks {
		if t.ID == id {
			r.Tracks[i].IsDeleted = true
			r.Tracks[i].DeleteReason = reason
			return nil
		}
	}
	return nil
}

// GetTracks returns tracks by IDs.
func (r *FakeRepo) GetTracks(_ context.Context, ids []uint64) ([]domain.Track, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	idSet := make(map[uint64]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}
	var result []domain.Track
	for _, t := range r.Tracks {
		if idSet[t.ID] {
			result = append(result, t)
		}
	}
	return result, nil
}

// LatestAddedAtForRoot returns 0 (tests don't need staleness checks).
func (r *FakeRepo) LatestAddedAtForRoot(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

// DeleteTracksForRoot is a no-op stub.
func (r *FakeRepo) DeleteTracksForRoot(_ context.Context, _ string) error { return nil }

// GetPathToIDMap returns path→ID for all active tracks.
func (r *FakeRepo) GetPathToIDMap(_ context.Context) (map[string]uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := make(map[string]uint64, len(r.Tracks))
	for _, t := range r.Tracks {
		if !t.IsDeleted {
			m[t.Path] = t.ID
		}
	}
	return m, nil
}

// Transaction stubs — tests don't exercise the transaction log.
func (r *FakeRepo) CreateTransaction(_ context.Context, _ domain.Transaction) error { return nil }
func (r *FakeRepo) UpdateTransactionStatus(_ context.Context, _ string, _ domain.TransactionStatus) error {
	return nil
}
func (r *FakeRepo) UpdateStepStatus(_ context.Context, _ string, _ int, _ domain.StepStatus, _ string) error {
	return nil
}
func (r *FakeRepo) GetPendingTransactions(_ context.Context) ([]domain.Transaction, error) {
	return nil, nil
}
