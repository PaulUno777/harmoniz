package organize

import (
	"context"
	"fmt"
	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/ports"
	"harmoniz/internal/logger"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// RenameService handles atomic single-track file renames and DB path updates.
type RenameService struct {
	fs   ports.FilesystemPort
	repo ports.TrackRepository
}

func NewRenameService(fs ports.FilesystemPort, repo ports.TrackRepository) *RenameService {
	return &RenameService{fs: fs, repo: repo}
}

// RenameTrack renames track.Path to newFilename (within the same directory).
// It sanitizes the filename, preserves the extension if omitted, checks for collisions,
// performs an atomic rename, and updates the DB. Returns the new absolute path.
func (s *RenameService) RenameTrack(ctx context.Context, track domain.Track, newFilename string) (string, error) {
	newFilename = SanitizeFilename(newFilename)
	if newFilename == "" {
		return "", fmt.Errorf("filename is empty after sanitization")
	}

	// Preserve original extension if caller omitted it
	origExt := filepath.Ext(track.Path)
	if filepath.Ext(newFilename) == "" {
		newFilename += origExt
	}

	dir := filepath.Dir(track.Path)
	newPath := filepath.Join(dir, newFilename)

	// No-op if name unchanged
	if newPath == track.Path {
		return track.Path, nil
	}

	// Collision check
	if _, err := os.Stat(newPath); err == nil {
		return "", fmt.Errorf("a file already exists with that name: %s", newFilename)
	}

	logger.Log.Info("Renaming track", "from", track.Path, "to", newPath)

	// os.Rename is atomic within the same filesystem (same directory always qualifies).
	if err := os.Rename(track.Path, newPath); err != nil {
		return "", fmt.Errorf("rename failed: %w", err)
	}

	// Update DB — if this fails the file is already renamed, log and surface the error
	// so the caller can decide (the UI will show the new path optimistically).
	if err := s.repo.UpdateTrackPath(ctx, track.ID, newPath); err != nil {
		logger.Log.Warn("File renamed but DB update failed", "newPath", newPath, "error", err)
		return newPath, fmt.Errorf("file renamed but DB update failed: %w", err)
	}

	return newPath, nil
}

// SanitizeFilename strips OS-illegal characters and trims whitespace.
// Exported so it can be used for template-based batch rename (Phase 7).
func SanitizeFilename(name string) string {
	const illegal = `/\:*?"<>|` + "\x00"
	var b strings.Builder
	for _, r := range name {
		if strings.ContainsRune(illegal, r) || !unicode.IsPrint(r) {
			b.WriteRune('-')
		} else {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
