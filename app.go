package main

import (
	"context"
	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/services/scanner"
	"harmoniz/internal/logger"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	scanner *scanner.Service
}

// NewApp creates a new App application struct
func NewApp(scannerService *scanner.Service) *App {
	return &App{
		scanner: scannerService,
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
