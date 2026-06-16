package db

import (
	"context"
	"database/sql"
	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/ports"
	"path/filepath"
	"strings"
	"time"
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

	now := time.Now().Unix()
	query := `
		INSERT INTO tracks (
			path, filename, size, mod_time, added_at,
			artist_raw, artist_norm, album_raw, album_norm, title, year, track_num, 
			hash_partial, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		addedAt := t.AddedAt
		if addedAt == 0 {
			addedAt = now
		}
		_, err := stmt.Exec(
			t.Path, t.Filename, t.Size, t.ModTime, addedAt,
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

// ListTracks returns tracks for the library, filtered by the provided criteria.
func (a *Adapter) ListTracks(ctx context.Context, filter domain.TrackFilter) ([]domain.Track, int, error) {
	conditions := []string{"is_deleted = 0"}
	args := []interface{}{}

	// Root path filter
	if filter.Root != "" {
		// Normalize root so LIKE matches paths stored by scanner (no trailing slash).
		normalized := strings.TrimRight(filepath.Clean(filter.Root), "/\\")
		conditions = append(conditions, "(path = ? OR path LIKE ?)")
		args = append(args, normalized, normalized+"/%")
	}

	// Search query (searches across artist, album, title, filename)
	if filter.SearchQuery != "" {
		searchPattern := "%" + strings.ToLower(filter.SearchQuery) + "%"
		conditions = append(conditions,
			"(LOWER(artist_raw) LIKE ? OR LOWER(album_raw) LIKE ? OR LOWER(title) LIKE ? OR LOWER(filename) LIKE ?)")
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Year range filters
	if filter.YearMin != nil {
		conditions = append(conditions, "year >= ?")
		args = append(args, *filter.YearMin)
	}
	if filter.YearMax != nil {
		conditions = append(conditions, "year <= ?")
		args = append(args, *filter.YearMax)
	}

	// Size range filters
	if filter.SizeMin != nil {
		conditions = append(conditions, "size >= ?")
		args = append(args, *filter.SizeMin)
	}
	if filter.SizeMax != nil {
		conditions = append(conditions, "size <= ?")
		args = append(args, *filter.SizeMax)
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count query
	countQuery := "SELECT COUNT(*) FROM tracks " + whereClause
	var total int
	if err := a.Conn.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Select query with ID-based ordering for consistent pagination
	sel := "SELECT id, path, filename, size, mod_time, added_at, artist_raw, artist_norm, album_raw, album_norm, title, year, track_num, bitrate, hash_partial, hash_full, fingerprint, is_deleted, deleted_at, delete_reason, status FROM tracks " + whereClause + " ORDER BY id LIMIT ? OFFSET ?"
	args = append(args, filter.Limit, filter.Offset)
	rows, err := a.Conn.QueryContext(ctx, sel, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tracks []domain.Track
	for rows.Next() {
		var t domain.Track
		var bitrate sql.NullInt64
		var hashFull, fingerprint, deleteReason sql.NullString
		if err := rows.Scan(
			&t.ID, &t.Path, &t.Filename, &t.Size, &t.ModTime, &t.AddedAt,
			&t.ArtistRaw, &t.ArtistNorm, &t.AlbumRaw, &t.AlbumNorm, &t.Title, &t.Year, &t.TrackNum, &bitrate,
			&t.HashPartial, &hashFull, &fingerprint, &t.IsDeleted, &t.DeletedAt, &deleteReason, &t.Status,
		); err != nil {
			return nil, 0, err
		}
		if bitrate.Valid {
			t.Bitrate = int(bitrate.Int64)
		}
		if hashFull.Valid {
			t.HashFull = hashFull.String
		}
		if fingerprint.Valid {
			t.Fingerprint = fingerprint.String
		}
		if deleteReason.Valid {
			t.DeleteReason = deleteReason.String
		}
		tracks = append(tracks, t)
	}
	return tracks, total, nil
}

// scanTracks scans rows (SELECT id, path, filename, size, mod_time, added_at, artist_raw, artist_norm, album_raw, album_norm, title, year, track_num, bitrate, hash_partial, hash_full, fingerprint, is_deleted, deleted_at, delete_reason, status) into []domain.Track.
func scanTracks(rows *sql.Rows) ([]domain.Track, error) {
	var tracks []domain.Track
	for rows.Next() {
		var t domain.Track
		var bitrate sql.NullInt64
		var hashFull, fingerprint, deleteReason sql.NullString
		if err := rows.Scan(
			&t.ID, &t.Path, &t.Filename, &t.Size, &t.ModTime, &t.AddedAt,
			&t.ArtistRaw, &t.ArtistNorm, &t.AlbumRaw, &t.AlbumNorm, &t.Title, &t.Year, &t.TrackNum, &bitrate,
			&t.HashPartial, &hashFull, &fingerprint, &t.IsDeleted, &t.DeletedAt, &deleteReason, &t.Status,
		); err != nil {
			return nil, err
		}
		if bitrate.Valid {
			t.Bitrate = int(bitrate.Int64)
		}
		if hashFull.Valid {
			t.HashFull = hashFull.String
		}
		if fingerprint.Valid {
			t.Fingerprint = fingerprint.String
		}
		if deleteReason.Valid {
			t.DeleteReason = deleteReason.String
		}
		tracks = append(tracks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tracks, nil
}

// LatestAddedAtForRoot returns the most recent added_at (Unix) for tracks under root. Returns 0 if none.
func (a *Adapter) LatestAddedAtForRoot(ctx context.Context, root string) (int64, error) {
	normalized := strings.TrimRight(filepath.Clean(root), "/\\")
	query := `SELECT COALESCE(MAX(added_at), 0) FROM tracks WHERE is_deleted = 0 AND (path = ? OR path LIKE ?)`
	var latest int64
	err := a.Conn.QueryRowContext(ctx, query, normalized, normalized+"/%").Scan(&latest)
	return latest, err
}

// DeleteTracksForRoot hard-deletes all tracks under root (path = root or path LIKE root/%).
func (a *Adapter) DeleteTracksForRoot(ctx context.Context, root string) error {
	normalized := strings.TrimRight(filepath.Clean(root), "/\\")
	_, err := a.Conn.ExecContext(ctx, `DELETE FROM tracks WHERE path = ? OR path LIKE ?`, normalized, normalized+"/%")
	return err
}

// GetDuplicateCandidates returns tracks that share the same file size (potential duplicates).
// root restricts to tracks under that path; empty means all non-deleted tracks.
func (a *Adapter) GetDuplicateCandidates(ctx context.Context, root string) ([]domain.Track, error) {
	conditions := []string{"is_deleted = 0"}
	args := []interface{}{}
	if root != "" {
		normalized := strings.TrimRight(filepath.Clean(root), "/\\")
		conditions = append(conditions, "(path = ? OR path LIKE ?)")
		args = append(args, normalized, normalized+"/%")
	}
	where := "WHERE " + strings.Join(conditions, " AND ")
	// Subquery: sizes that appear more than once. Outer: all tracks with those sizes (same filter).
	query := `
		SELECT t1.id, t1.path, t1.filename, t1.size, t1.mod_time, t1.added_at,
		       t1.artist_raw, t1.artist_norm, t1.album_raw, t1.album_norm, t1.title, t1.year, t1.track_num, t1.bitrate,
		       t1.hash_partial, t1.hash_full, t1.fingerprint, t1.is_deleted, t1.deleted_at, t1.delete_reason, t1.status
		FROM tracks t1
		INNER JOIN (
			SELECT size FROM tracks ` + where + ` GROUP BY size HAVING COUNT(*) > 1
		) t2 ON t1.size = t2.size
		` + where + `
		ORDER BY t1.size DESC, t1.id
	`
	// Same args for subquery and outer WHERE
	allArgs := append(append([]interface{}{}, args...), args...)
	rows, err := a.Conn.QueryContext(ctx, query, allArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTracks(rows)
}

// StreamUniqueArtists streams distinct artist_norm for non-deleted tracks.
// root restricts to tracks under that path; empty means all.
func (a *Adapter) StreamUniqueArtists(ctx context.Context, root string) (<-chan string, error) {
	conditions := []string{"is_deleted = 0", "COALESCE(TRIM(artist_norm), '') != ''"}
	args := []interface{}{}
	if root != "" {
		normalized := strings.TrimRight(filepath.Clean(root), "/\\")
		conditions = append(conditions, "(path = ? OR path LIKE ?)")
		args = append(args, normalized, normalized+"/%")
	}
	query := `SELECT DISTINCT artist_norm FROM tracks WHERE ` + strings.Join(conditions, " AND ")
	rows, err := a.Conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	ch := make(chan string, 64)
	go func() {
		defer rows.Close()
		defer close(ch)
		for rows.Next() {
			var artist string
			if err := rows.Scan(&artist); err != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case ch <- artist:
			}
		}
	}()
	return ch, nil
}

func (a *Adapter) UpdateTrackPath(ctx context.Context, id uint64, newPath string) error {
	newFilename := filepath.Base(newPath)
	_, err := a.Conn.ExecContext(ctx,
		`UPDATE tracks SET path = ?, filename = ?, status = 'modified' WHERE id = ?`,
		newPath, newFilename, id,
	)
	return err
}

func (a *Adapter) UpdateTrackTags(ctx context.Context, id uint64, artist, title, album string, year, trackNum int) error {
	artistNorm := domain.NormalizeArtist(artist)
	albumNorm := domain.NormalizeArtist(album)
	_, err := a.Conn.ExecContext(ctx,
		`UPDATE tracks SET artist_raw = ?, artist_norm = ?, album_raw = ?, album_norm = ?, title = ?, year = ?, track_num = ?, status = 'modified' WHERE id = ?`,
		artist, artistNorm, album, albumNorm, title, year, trackNum, id,
	)
	return err
}

func (a *Adapter) GetTracks(ctx context.Context, ids []uint64) ([]domain.Track, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	query := `SELECT id, path, filename, size, mod_time, added_at, artist_raw, artist_norm, album_raw, album_norm, title, year, track_num, bitrate, hash_partial, hash_full, fingerprint, is_deleted, deleted_at, delete_reason, status FROM tracks WHERE id IN (` + placeholders + `)`
	rows, err := a.Conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTracks(rows)
}

func (a *Adapter) GetPathToIDMap(ctx context.Context) (map[string]uint64, error) {
	rows, err := a.Conn.QueryContext(ctx, `SELECT path, id FROM tracks WHERE is_deleted = 0`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]uint64)
	for rows.Next() {
		var path string
		var id uint64
		if err := rows.Scan(&path, &id); err != nil {
			return nil, err
		}
		m[path] = id
	}
	return m, nil
}

func (a *Adapter) SoftDelete(ctx context.Context, id uint64, reason string) error {
	now := time.Now().Unix()
	_, err := a.Conn.ExecContext(ctx,
		`UPDATE tracks SET is_deleted = 1, deleted_at = ?, delete_reason = ? WHERE id = ?`,
		now, reason, id,
	)
	return err
}

// Transaction methods — implemented in Phase 8 when the transactions table is added.

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
