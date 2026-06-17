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

// OrganizeTrack moves a track to newRelPath within libraryRoot, creating subdirectories
// as needed and avoiding filename collisions by appending a counter suffix.
// newRelPath may include path separators for subfolder organization
// (e.g. "Artist/Artist - Title.mp3").
func (s *RenameService) OrganizeTrack(ctx context.Context, track domain.Track, newRelPath, libraryRoot string) (string, error) {
	// Sanitize each path component individually (preserve '/' as separator)
	parts := strings.Split(filepath.ToSlash(newRelPath), "/")
	sanitized := make([]string, 0, len(parts))
	for _, p := range parts {
		if c := SanitizePathComponent(p); c != "" {
			sanitized = append(sanitized, c)
		}
	}
	if len(sanitized) == 0 {
		return "", fmt.Errorf("path is empty after sanitization")
	}

	newPath := filepath.Join(append([]string{libraryRoot}, sanitized...)...)

	// Preserve extension if the template produced none
	if filepath.Ext(newPath) == "" {
		newPath += filepath.Ext(track.Path)
	}

	// Avoid collision: if target already exists and is not the same file, add counter
	newPath = avoidCollision(newPath, track.Path)

	if newPath == track.Path {
		return track.Path, nil
	}

	logger.Log.Info("Organizing track", "from", track.Path, "to", newPath)

	if err := s.fs.SafeMove(track.Path, newPath); err != nil {
		return "", fmt.Errorf("move failed: %w", err)
	}

	if err := s.repo.UpdateTrackPath(ctx, track.ID, newPath); err != nil {
		logger.Log.Warn("File moved but DB update failed", "newPath", newPath, "error", err)
		return newPath, fmt.Errorf("file moved but DB update failed: %w", err)
	}

	return newPath, nil
}

// avoidCollision appends _2, _3 … until an unused path is found.
// If targetPath == srcPath the file is the same, return as-is.
func avoidCollision(targetPath, srcPath string) string {
	if targetPath == srcPath {
		return srcPath
	}
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return targetPath
	}
	ext := filepath.Ext(targetPath)
	base := strings.TrimSuffix(targetPath, ext)
	for i := 2; i <= 999; i++ {
		candidate := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	return targetPath
}

// SanitizePathComponent sanitizes a single path component (no '/' allowed).
func SanitizePathComponent(name string) string {
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

// SanitizeFilename strips OS-illegal characters and trims whitespace.
// Exported so it can be used for template-based batch rename (Phase 7).
func SanitizeFilename(name string) string {
	return SanitizePathComponent(name)
}
