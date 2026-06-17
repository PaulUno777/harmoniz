package testutil

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"harmoniz/internal/adapters/db"
	"harmoniz/internal/core/ports"
	"harmoniz/internal/core/services/scanner"
	"harmoniz/internal/logger"
)

// FindRepoRoot walks up from the test working directory until it finds go.mod.
func FindRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (no go.mod found)")
		}
		dir = parent
	}
}

// TestSongsDir returns the path to test/songs inside the repository.
// The test is skipped automatically if the directory does not exist.
func TestSongsDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(FindRepoRoot(t), "test", "songs")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("test/songs directory not found — skipping integration test")
	}
	return dir
}

// CopyFile copies src to dst, creating any parent directories that do not exist.
func CopyFile(t *testing.T, src, dst string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(dst), err)
	}
	in, err := os.Open(src)
	if err != nil {
		t.Fatalf("open src %s: %v", src, err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		t.Fatalf("create dst %s: %v", dst, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		t.Fatalf("copy %s → %s: %v", src, dst, err)
	}
}

// CopyFileTo copies src into dstDir, keeping the original filename.
// Returns the full destination path.
func CopyFileTo(t *testing.T, src, dstDir string) string {
	t.Helper()
	dst := filepath.Join(dstDir, filepath.Base(src))
	CopyFile(t, src, dst)
	return dst
}

// CopyFileAs copies src to dst with a new filename (base name from dst).
func CopyFileAs(t *testing.T, src, dst string) {
	t.Helper()
	CopyFile(t, src, dst)
}

// BuildScanDB creates a temporary SQLite database, runs migrations, scans root with
// the real scanner, and returns the populated TrackRepository. The database file lives
// in t.TempDir() and is automatically removed when the test finishes.
func BuildScanDB(t *testing.T, root string) ports.TrackRepository {
	t.Helper()
	logger.Init(slog.LevelError) // suppress scanner log output in tests

	dbPath := filepath.Join(t.TempDir(), "test.db")
	adapter, err := db.NewAdapter(dbPath)
	if err != nil {
		t.Fatalf("NewAdapter: %v", err)
	}
	t.Cleanup(func() { adapter.Close() })

	if err := adapter.RunMigrations(db.MigrationFS, "migrations"); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	svc := scanner.NewService(adapter)
	if err := svc.Scan(context.Background(), root); err != nil {
		t.Fatalf("scanner.Scan: %v", err)
	}
	return adapter
}
