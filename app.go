package main

import (
	"context"
	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/services/scanner"
	"harmoniz/internal/logger"

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

// ScanLibrary scans a music library folder and imports tracks into the database
func (a *App) ScanLibrary(root string) error {
	logger.Log.Info("ScanLibrary called", "root", root)
	err := a.scanner.Scan(a.ctx, root)
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

// ListTracks returns tracks for the library, filtered by root path only.
// Scope: only tracks under root (that path and its subfolders) are returned; parent and unrelated paths are excluded.
// Returns a single struct so Wails serializes it reliably to JS as { Tracks, Total }.
func (a *App) ListTracks(root string, limit, offset int) (*domain.ListTracksResult, error) {
	tracks, total, err := a.scanner.ListTracks(a.ctx, root, limit, offset)
	if err != nil {
		return nil, err
	}
	return &domain.ListTracksResult{Tracks: tracks, Total: total}, nil
}
