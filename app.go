package main

import (
	"context"
	"fmt"
	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/ports"
	"harmoniz/internal/core/services/analysis"
	"harmoniz/internal/core/services/organize"
	"harmoniz/internal/core/services/scanner"
	"harmoniz/internal/logger"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	repo       ports.TrackRepository
	scanner    *scanner.Service
	clustering *analysis.ClusteringService
	deduper    *analysis.DeduplicationService
	renamer    *organize.RenameService
	metaWriter ports.MetadataWriter
}

// NewApp creates a new App application struct
func NewApp(
	scannerService *scanner.Service,
	repo ports.TrackRepository,
	fsPort ports.FilesystemPort,
	metaWriter ports.MetadataWriter,
) *App {
	return &App{
		repo:       repo,
		scanner:    scannerService,
		clustering: analysis.NewClusteringService(repo),
		deduper:    analysis.NewDeduplicationService(repo),
		renamer:    organize.NewRenameService(fsPort, repo),
		metaWriter: metaWriter,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logger.Log.Info("Wails App Started Successfully")
}

// ScanLibrary ensures the library is synced then scans. If the last sync for this root
// was more than 24h ago (or no data), all tracks for that root are removed and re-scanned from disk.
func (a *App) ScanLibrary(root string) error {
	logger.Log.Info("ScanLibrary called", "root", root)
	err := a.scanner.PrepareLibrarySync(a.ctx, root, 24*time.Hour)
	if err == nil && a.ctx != nil {
		runtime.EventsEmit(a.ctx, "scan:done", root)
	}
	return err
}

// OpenFolderDialog opens the native directory picker and returns the selected path (empty if cancelled).
func (a *App) OpenFolderDialog() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "Select Music Library Root",
		CanCreateDirectories: true,
	})
}

// ListTracks returns tracks for the library, filtered by the provided criteria.
// Supports text search (across artist, album, title, filename) and range filters (year, file size).
// Returns a single struct so Wails serializes it reliably to JS as { Tracks, Total }.
func (a *App) ListTracks(
	root string,
	searchQuery string,
	yearMin int,
	yearMax int,
	sizeMin int64,
	sizeMax int64,
	limit int,
	offset int,
) (*domain.ListTracksResult, error) {
	filter := domain.TrackFilter{
		Root:        root,
		SearchQuery: searchQuery,
		Limit:       limit,
		Offset:      offset,
	}

	// Set range filters only if provided (non-zero values)
	if yearMin > 0 {
		filter.YearMin = &yearMin
	}
	if yearMax > 0 {
		filter.YearMax = &yearMax
	}
	if sizeMin > 0 {
		filter.SizeMin = &sizeMin
	}
	if sizeMax > 0 {
		filter.SizeMax = &sizeMax
	}

	tracks, total, err := a.scanner.ListTracks(a.ctx, filter)
	if err != nil {
		logger.Log.Error("ListTracks failed", "error", err)
		return nil, err
	}
	logger.Log.Info("ListTracks returned", "tracks", len(tracks), "total", total, "filter", filter)
	return &domain.ListTracksResult{Tracks: tracks, Total: total}, nil
}

// AnalyzeArtists returns artist name clustering suggestions for the given library root.
// root is the current library path; use "" to analyze all tracks in the database.
func (a *App) AnalyzeArtists(root string) ([]domain.ArtistSuggestion, error) {
	if a.ctx == nil {
		a.ctx = context.Background()
	}
	return a.clustering.AnalyzeArtists(a.ctx, root)
}

// DetectDuplicates returns groups of byte-identical tracks under the given root.
// root is the current library path; use "" to scan the entire database.
func (a *App) DetectDuplicates(root string) ([][]domain.Track, error) {
	if a.ctx == nil {
		a.ctx = context.Background()
	}
	return a.deduper.DetectDuplicates(a.ctx, root)
}

// UpdateTrackTags writes new metadata to the audio file (best-effort, MP3 only in Phase 5)
// and updates all metadata fields in the database.
func (a *App) UpdateTrackTags(id uint64, artist, title, album string, year, trackNum int) error {
	if a.ctx == nil {
		a.ctx = context.Background()
	}

	tracks, err := a.repo.GetTracks(a.ctx, []uint64{id})
	if err != nil {
		return fmt.Errorf("failed to retrieve track: %w", err)
	}
	if len(tracks) == 0 {
		return fmt.Errorf("track %d not found", id)
	}
	track := tracks[0]

	// Write tags to file — best-effort; non-MP3 formats log a warning and continue.
	if writeErr := a.metaWriter.WriteMetadata(track.Path, ports.TrackMetadata{
		Artist:   artist,
		Title:    title,
		Album:    album,
		Year:     year,
		TrackNum: trackNum,
	}); writeErr != nil {
		logger.Log.Warn("Tag write to file skipped", "path", track.Path, "reason", writeErr)
	}

	return a.repo.UpdateTrackTags(a.ctx, id, artist, title, album, year, trackNum)
}

// RenameTrack renames the audio file on disk and updates the path in the database.
// newFilename is the bare filename (with or without extension); the directory is preserved.
// Returns the new absolute path on success.
func (a *App) RenameTrack(id uint64, newFilename string) (string, error) {
	if a.ctx == nil {
		a.ctx = context.Background()
	}

	tracks, err := a.repo.GetTracks(a.ctx, []uint64{id})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve track: %w", err)
	}
	if len(tracks) == 0 {
		return "", fmt.Errorf("track %d not found", id)
	}

	return a.renamer.RenameTrack(a.ctx, tracks[0], newFilename)
}
