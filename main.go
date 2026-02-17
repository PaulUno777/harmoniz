package main

import (
	"embed"
	"harmoniz/internal/adapters/db"
	"harmoniz/internal/core/services/scanner"
	"harmoniz/internal/logger"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Initialize logger
	logger.Init(slog.LevelDebug)

	// Setup DB
	userHome, err := os.UserHomeDir()
	if err != nil {
		logger.Log.Error("Failed to get user home directory", "error", err)
		panic(err)
	}
	appDir := filepath.Join(userHome, ".harmoniz")
	os.MkdirAll(appDir, 0755)
	dbPath := filepath.Join(appDir, "library.db")

	dbAdapter, err := db.NewAdapter(dbPath)
	if err != nil {
		logger.Log.Error("Failed to create database adapter", "error", err)
		panic(err)
	}
	defer dbAdapter.Close()

	// Debug: List embedded files
	if err := db.ListEmbeddedFiles(); err != nil {
		logger.Log.Warn("Failed to list embedded files", "error", err)
	}

	// Run migrations
	if err := dbAdapter.RunMigrations(db.MigrationFS, "migrations"); err != nil {
		logger.Log.Error("Failed to run migrations", "error", err)
		panic(err)
	}

	logger.Log.Info("Database initialized successfully", "path", dbPath)

	// Setup Services
	scannerService := scanner.NewService(dbAdapter)

	// Create an instance of the app structure
	app := NewApp(scannerService)

	// Create application with options
	err = wails.Run(&options.App{
		Title:     "Harmoniz",
		Width:     1366,
		Height:    768,

		MinHeight: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:   true,
			DisableWebViewDrop: false,
		},
	})

	if err != nil {
		logger.Log.Error("Failed to start Wails app", "error", err)
		println("Error:", err.Error())
	}
}
