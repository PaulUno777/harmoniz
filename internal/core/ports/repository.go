package ports

import (
	"context"
	"harmoniz/internal/core/domain"
)

type TrackRepository interface {
	// BatchUpsert inserts or updates multiple tracks in a single transaction.
	// It should use INSERT ... ON CONFLICT DO UPDATE on the 'path' unique key.
	BatchUpsert(tracks []domain.Track) error

	// GetAllPathsModTime returns a map of Path -> ModTime for all active (non-deleted) tracks.
	// This is used to pre-load the cache for the scanner to avoid N+1 DB queries.
	GetAllPathsModTime() (map[string]int64, error)

	// GetDuplicateCandidates returns tracks that share the same size (potential duplicates).
	// It uses an optimized JOIN query to filter only relevant tracks.
	GetDuplicateCandidates() ([]domain.Track, error)

	// StreamUniqueArtists streams unique normalized artist names to a channel.
	StreamUniqueArtists(ctx context.Context) (<-chan string, error)

	// UpdateTrackPath updates the file path of a track.
	UpdateTrackPath(ctx context.Context, id uint64, newPath string) error

	// Transaction methods
	CreateTransaction(ctx context.Context, tx domain.Transaction) error
	UpdateTransactionStatus(ctx context.Context, id string, status domain.TransactionStatus) error
	UpdateStepStatus(ctx context.Context, txID string, stepIndex int, status domain.StepStatus, errMsg string) error
	GetPendingTransactions(ctx context.Context) ([]domain.Transaction, error)

	// SoftDelete marks a track as deleted with a reason and timestamp.
	SoftDelete(ctx context.Context, id uint64, reason string) error

	// GetTracks returns tracks by IDs.
	GetTracks(ctx context.Context, ids []uint64) ([]domain.Track, error)

	// ListTracks returns tracks for the library, optionally filtered by root path prefix.
	// If root is empty, all non-deleted tracks are returned. total is the full count matching the filter.
	ListTracks(ctx context.Context, root string, limit, offset int) ([]domain.Track, int, error)

	// LatestAddedAtForRoot returns the most recent added_at (Unix timestamp) among tracks under root.
	// Returns 0 if no tracks. Used for staleness check before full re-sync.
	LatestAddedAtForRoot(ctx context.Context, root string) (int64, error)

	// DeleteTracksForRoot hard-deletes all tracks under root (path = root OR path LIKE root/%).
	// Used when re-syncing from disk after data is considered stale.
	DeleteTracksForRoot(ctx context.Context, root string) error

	// GetPathToIDMap returns a map of Path -> ID for all active tracks.
	GetPathToIDMap(ctx context.Context) (map[string]uint64, error)
}
