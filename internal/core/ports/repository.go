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
	// root restricts to tracks under that path (path = root OR path LIKE root/%); empty = all.
	GetDuplicateCandidates(ctx context.Context, root string) ([]domain.Track, error)

	// StreamUniqueArtists streams unique normalized artist names. root restricts scope; empty = all.
	StreamUniqueArtists(ctx context.Context, root string) (<-chan string, error)

	// UpdateTrackPath updates the file path and filename of a track after a rename.
	UpdateTrackPath(ctx context.Context, id uint64, newPath string) error

	// UpdateTrackTags writes new metadata fields into the DB for the given track.
	// Re-computes artist_norm and album_norm. Sets status = 'modified'.
	UpdateTrackTags(ctx context.Context, id uint64, artist, title, album string, year, trackNum int) error

	// Transaction methods
	CreateTransaction(ctx context.Context, tx domain.Transaction) error
	UpdateTransactionStatus(ctx context.Context, id string, status domain.TransactionStatus) error
	UpdateStepStatus(ctx context.Context, txID string, stepIndex int, status domain.StepStatus, errMsg string) error
	GetPendingTransactions(ctx context.Context) ([]domain.Transaction, error)

	// SoftDelete marks a track as deleted with a reason and timestamp.
	SoftDelete(ctx context.Context, id uint64, reason string) error

	// GetTracks returns tracks by IDs.
	GetTracks(ctx context.Context, ids []uint64) ([]domain.Track, error)

	// ListTracks returns tracks for the library, filtered by the provided criteria.
	// If filter.Root is empty, all non-deleted tracks are returned. total is the full count matching the filter.
	ListTracks(ctx context.Context, filter domain.TrackFilter) ([]domain.Track, int, error)

	// LatestAddedAtForRoot returns the most recent added_at (Unix timestamp) among tracks under root.
	// Returns 0 if no tracks. Used for staleness check before full re-sync.
	LatestAddedAtForRoot(ctx context.Context, root string) (int64, error)

	// DeleteTracksForRoot hard-deletes all tracks under root (path = root OR path LIKE root/%).
	// Used when re-syncing from disk after data is considered stale.
	DeleteTracksForRoot(ctx context.Context, root string) error

	// GetPathToIDMap returns a map of Path -> ID for all active tracks.
	GetPathToIDMap(ctx context.Context) (map[string]uint64, error)
}
