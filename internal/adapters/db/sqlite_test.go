package db_test

import (
	"embed"
	"harmoniz/internal/adapters/db"
	"harmoniz/internal/logger"
	"log/slog"
	"os"
	"testing"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func TestDatabaseSetup(t *testing.T) {
	// Setup Logger
	logger.Init(slog.LevelDebug)

	dbPath := "test_library.db"
	os.Remove(dbPath)
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")
	defer func() {
		// cleanup
		os.Remove(dbPath)
		os.Remove(dbPath + "-wal")
		os.Remove(dbPath + "-shm")
	}()

	// 1. Init Adapter
	adapter, err := db.NewAdapter(dbPath)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	defer adapter.Close()

	// 2. Run Migrations - Use the embedded MigrationFS from db package
	err = adapter.RunMigrations(db.MigrationFS, "migrations")
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// 3. Verify Tables
	var tableName string
	err = adapter.Conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='tracks'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Tracks table not found: %v", err)
	}

	// 4. Verify Migration Version
	var version int
	err = adapter.Conn.QueryRow("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if err != nil {
		t.Fatalf("Migration version not found: %v", err)
	}
	if version != 2 {
		t.Errorf("Expected migration version 2, got %d", version)
	}

	// 5. Verify WAL Mode
	var mode string
	err = adapter.Conn.QueryRow("PRAGMA journal_mode").Scan(&mode)
	if err != nil {
		t.Fatalf("Failed to query journal mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("Expected WAL mode, got %s", mode)
	}

	// 6. Verify Constraints (Unique Path)
	_, err = adapter.Conn.Exec("INSERT INTO tracks (path, filename, size, mod_time) VALUES (?, ?, ?, ?)", "/tmp/test.mp3", "test.mp3", 1000, 1234567890)
	if err != nil {
		t.Fatalf("Failed to insert track: %v", err)
	}

	_, err = adapter.Conn.Exec("INSERT INTO tracks (path, filename, size, mod_time) VALUES (?, ?, ?, ?)", "/tmp/test.mp3", "test.mp3", 1000, 1234567890)
	if err == nil {
		t.Fatal("Expected unique constraint violation, got nil")
	}

	logger.Log.Info("Database test passed successfully")
}
