package db

import (
	"context"
	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/ports"
	"path/filepath"
	"strings"
)

// Ensure Adapter implements ports.TrackRepository
var _ ports.TrackRepository = (*Adapter)(nil)

// BatchUpsert inserts or updates multiple tracks in a single transaction.
func (a *Adapter) BatchUpsert(tracks []domain.Track) error {
	tx, err := a.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO tracks (
			path, filename, size, mod_time, 
			artist_raw, artist_norm, album_raw, album_norm, title, year, track_num, 
			hash_partial, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			size = excluded.size,
			mod_time = excluded.mod_time,
			artist_raw = excluded.artist_raw,
			artist_norm = excluded.artist_norm,
			album_raw = excluded.album_raw,
			album_norm = excluded.album_norm,
			title = excluded.title,
			year = excluded.year,
			track_num = excluded.track_num,
			hash_partial = excluded.hash_partial,
			status = excluded.status
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range tracks {
		_, err := stmt.Exec(
			t.Path, t.Filename, t.Size, t.ModTime,
			t.ArtistRaw, t.ArtistNorm, t.AlbumRaw, t.AlbumNorm, t.Title, t.Year, t.TrackNum,
			t.HashPartial, t.Status,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetAllPathsModTime returns a map of Path -> ModTime for all active (non-deleted) tracks.
func (a *Adapter) GetAllPathsModTime() (map[string]int64, error) {
	rows, err := a.Conn.Query("SELECT path, mod_time FROM tracks WHERE is_deleted = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cache := make(map[string]int64)
	for rows.Next() {
		var path string
		var modTime int64
		if err := rows.Scan(&path, &modTime); err != nil {
			return nil, err
		}
		cache[path] = modTime
	}
	return cache, nil
}

// ListTracks returns tracks for the library, optionally filtered by root path prefix.
func (a *Adapter) ListTracks(ctx context.Context, root string, limit, offset int) ([]domain.Track, int, error) {
	baseCond := "WHERE is_deleted = 0"
	args := []interface{}{}
	if root != "" {
		// Normalize root so LIKE matches paths stored by scanner (no trailing slash).
		normalized := strings.TrimRight(filepath.Clean(root), "/\\")
		baseCond += " AND (path = ? OR path LIKE ?)"
		args = append(args, normalized, normalized+"/%")
	}

	countQuery := "SELECT COUNT(*) FROM tracks " + baseCond
	var total int
	if err := a.Conn.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sel := "SELECT id, path, filename, size, mod_time, artist_raw, artist_norm, album_raw, album_norm, title, year, track_num, bitrate, hash_partial, hash_full, fingerprint, is_deleted, deleted_at, delete_reason, status FROM tracks " + baseCond + " ORDER BY path LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := a.Conn.QueryContext(ctx, sel, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tracks []domain.Track
	for rows.Next() {
		var t domain.Track
		if err := rows.Scan(
			&t.ID, &t.Path, &t.Filename, &t.Size, &t.ModTime,
			&t.ArtistRaw, &t.ArtistNorm, &t.AlbumRaw, &t.AlbumNorm, &t.Title, &t.Year, &t.TrackNum, &t.Bitrate,
			&t.HashPartial, &t.HashFull, &t.Fingerprint, &t.IsDeleted, &t.DeletedAt, &t.DeleteReason, &t.Status,
		); err != nil {
			return nil, 0, err
		}
		tracks = append(tracks, t)
	}
	return tracks, total, nil
}

// Stub implementations for Phase 2 - will be implemented in later phases
func (a *Adapter) GetDuplicateCandidates() ([]domain.Track, error) {
	return nil, nil
}

func (a *Adapter) StreamUniqueArtists(ctx context.Context) (<-chan string, error) {
	return nil, nil
}

func (a *Adapter) UpdateTrackPath(ctx context.Context, id uint64, newPath string) error {
	return nil
}

func (a *Adapter) CreateTransaction(ctx context.Context, tx domain.Transaction) error {
	return nil
}

func (a *Adapter) UpdateTransactionStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	return nil
}

func (a *Adapter) UpdateStepStatus(ctx context.Context, txID string, stepIndex int, status domain.StepStatus, errMsg string) error {
	return nil
}

func (a *Adapter) GetPendingTransactions(ctx context.Context) ([]domain.Transaction, error) {
	return nil, nil
}

func (a *Adapter) SoftDelete(ctx context.Context, id uint64, reason string) error {
	return nil
}

func (a *Adapter) GetTracks(ctx context.Context, ids []uint64) ([]domain.Track, error) {
	return nil, nil
}

func (a *Adapter) GetPathToIDMap(ctx context.Context) (map[string]uint64, error) {
	return nil, nil
}
