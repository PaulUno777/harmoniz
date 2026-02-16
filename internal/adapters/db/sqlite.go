package db

import (
	"database/sql"
	"embed"
	"fmt"
	"harmoniz/internal/logger"
	"io/fs"
	"sort"
	"strings"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

//go:embed migrations/*.sql
var MigrationFS embed.FS

// ListEmbeddedFiles is a helper to debug embedded files (can be removed later)
func ListEmbeddedFiles() error {
	entries, err := fs.ReadDir(MigrationFS, "migrations")
	if err != nil {
		return err
	}
	logger.Log.Info("Embedded migration files:")
	for _, entry := range entries {
		logger.Log.Info("  Found", "name", entry.Name(), "isDir", entry.IsDir())
	}
	return nil
}

type Adapter struct {
	Conn *sql.DB
}

func NewAdapter(dbPath string) (*Adapter, error) {
	// Connect to SQLite
	// _pragma=journal_mode=WAL enables Write-Ahead Logging for concurrency
	// _pragma=foreign_keys=ON ensures data integrity
	dsn := fmt.Sprintf("%s?_pragma=journal_mode=WAL&_pragma=foreign_keys=ON&_pragma=busy_timeout=5000", dbPath)

	logger.Log.Info("Opening database connection", "dsn", dsn)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Adapter{Conn: conn}, nil
}

func (a *Adapter) Close() error {
	return a.Conn.Close()
}

// EnsureSchemaMigrationsTable creates the migration tracking table if not exists
func (a *Adapter) EnsureSchemaMigrationsTable() error {
	query := `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		dirty BOOLEAN NOT NULL
	);`
	_, err := a.Conn.Exec(query)
	return err
}

// RunMigrations applies SQL migrations from the provided FS
func (a *Adapter) RunMigrations(migrationFS fs.FS, dir string) error {
	if err := a.EnsureSchemaMigrationsTable(); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	// Read migration files
	logger.Log.Info("Reading migrations directory", "dir", dir)
	entries, err := fs.ReadDir(migrationFS, dir)
	if err != nil {
		logger.Log.Error("Failed to read migrations directory", "dir", dir, "error", err)
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrations = append(migrations, entry.Name())
			logger.Log.Info("Found migration file", "file", entry.Name())
		}
	}
	sort.Strings(migrations)

	if len(migrations) == 0 {
		logger.Log.Warn("No migration files found", "dir", dir)
		return fmt.Errorf("no migration files found in directory: %s", dir)
	}

	logger.Log.Info("Found migrations", "count", len(migrations), "files", migrations)

	// Get current version
	var currentVersion int
	var dirty bool
	err = a.Conn.QueryRow("SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&currentVersion, &dirty)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state from version %d", currentVersion)
	}

	logger.Log.Info("Current DB Version", "version", currentVersion)

	for _, migFile := range migrations {
		// Extract version from filename "001_init.sql" -> 1
		var version int
		_, err := fmt.Sscanf(migFile, "%d_", &version)
		if err != nil {
			continue // Skip files not matching pattern
		}

		if version <= currentVersion {
			continue
		}

		logger.Log.Info("Applying migration", "file", migFile)

		// Mark dirty
		_, err = a.Conn.Exec("INSERT INTO schema_migrations (version, dirty) VALUES (?, ?)", version, true)
		if err != nil {
			return fmt.Errorf("failed to record dirty migration: %w", err)
		}

		content, err := fs.ReadFile(migrationFS, dir+"/"+migFile)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migFile, err)
		}

		_, err = a.Conn.Exec(string(content))
		if err != nil {
			logger.Log.Error("Migration execution failed", "file", migFile, "error", err)
			return fmt.Errorf("failed to execute migration %s: %w", migFile, err)
		}

		logger.Log.Info("Migration executed successfully", "file", migFile, "version", version)

		// Mark clean
		_, err = a.Conn.Exec("UPDATE schema_migrations SET dirty = ? WHERE version = ?", false, version)
		if err != nil {
			return fmt.Errorf("failed to mark migration as clean: %w", err)
		}
	}

	logger.Log.Info("All migrations completed successfully")
	return nil
}
