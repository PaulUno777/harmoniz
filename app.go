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
	"strconv"
	"strings"
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
	suggester  *organize.SuggestionService
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
		suggester:  organize.NewSuggestionService(repo),
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

// AnalyzeForOrganize analyzes all tracks under root and returns per-track metadata suggestions.
func (a *App) AnalyzeForOrganize(root string) ([]domain.OrganizerSuggestion, error) {
	if a.ctx == nil {
		a.ctx = context.Background()
	}
	return a.suggester.AnalyzeRoot(a.ctx, root)
}

// ApplyOrganizerSuggestion applies selected metadata fields to a track (non-destructive).
// Empty string values for string fields mean "keep existing"; 0 for int fields means "keep existing".
// filenameTemplate: if non-empty, also rename the file using this template after updating tags.
// Returns the new absolute file path (unchanged if no rename was performed).
func (a *App) ApplyOrganizerSuggestion(trackID uint64, artist, title, album string, year, trackNum int, filenameTemplate string) (string, error) {
	if a.ctx == nil {
		a.ctx = context.Background()
	}

	tracks, err := a.repo.GetTracks(a.ctx, []uint64{trackID})
	if err != nil || len(tracks) == 0 {
		return "", fmt.Errorf("track %d not found", trackID)
	}
	t := tracks[0]

	// Merge: keep existing value when new value is empty/zero
	newArtist := mergeStr(artist, t.ArtistRaw)
	newTitle := mergeStr(title, t.Title)
	newAlbum := mergeStr(album, t.AlbumRaw)
	newYear := mergeInt(year, t.Year)
	newTrackNum := mergeInt(trackNum, t.TrackNum)

	// Write tags to file (best-effort; non-MP3 formats log a warning)
	if writeErr := a.metaWriter.WriteMetadata(t.Path, ports.TrackMetadata{
		Artist: newArtist, Title: newTitle, Album: newAlbum,
		Year: newYear, TrackNum: newTrackNum,
	}); writeErr != nil {
		logger.Log.Warn("Tag write skipped during organize", "path", t.Path, "reason", writeErr)
	}

	if err := a.repo.UpdateTrackTags(a.ctx, trackID, newArtist, newTitle, newAlbum, newYear, newTrackNum); err != nil {
		return "", fmt.Errorf("DB update failed: %w", err)
	}

	// Rename if template provided
	newPath := t.Path
	if strings.TrimSpace(filenameTemplate) != "" {
		updated := t
		updated.ArtistRaw = newArtist
		updated.Title = newTitle
		updated.AlbumRaw = newAlbum
		updated.Year = newYear
		updated.TrackNum = newTrackNum

		newFilename := organize.ApplyTemplate(filenameTemplate, updated)
		if newFilename != "" && newFilename != t.Filename {
			if renamedPath, renameErr := a.renamer.RenameTrack(a.ctx, updated, newFilename); renameErr == nil {
				newPath = renamedPath
			} else {
				logger.Log.Warn("Rename skipped during organize", "reason", renameErr)
			}
		}
	}

	return newPath, nil
}

// ApplyAllHighConfidence applies high-confidence suggestions (score >= minConfidence, fields >= 0.75)
// across the entire root. Returns count of successfully updated tracks.
func (a *App) ApplyAllHighConfidence(root string, minConfidence float64, filenameTemplate string) (int, error) {
	suggestions, err := a.AnalyzeForOrganize(root)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, s := range suggestions {
		if s.Score < minConfidence || len(s.Fields) == 0 {
			continue
		}
		artist := highConfStr(s.Fields, "artist", 0.75)
		title := highConfStr(s.Fields, "title", 0.75)
		album := highConfStr(s.Fields, "album", 0.75)
		year := highConfInt(s.Fields, "year", 0.75)
		trackNum := highConfInt(s.Fields, "track_num", 0.75)

		if artist == "" && title == "" && album == "" && year == 0 && trackNum == 0 {
			continue
		}

		if _, applyErr := a.ApplyOrganizerSuggestion(s.TrackID, artist, title, album, year, trackNum, filenameTemplate); applyErr != nil {
			logger.Log.Warn("Batch apply failed", "track_id", s.TrackID, "error", applyErr)
			continue
		}
		count++
	}

	runtime.EventsEmit(a.ctx, "organize:batch_done", count)
	return count, nil
}

// PreviewFilenameTemplate previews the filenames that would result from applying the template
// to all tracks under root. Returns only entries where the filename would change.
func (a *App) PreviewFilenameTemplate(root string, tmpl string) ([]domain.FilenamePreview, error) {
	if a.ctx == nil {
		a.ctx = context.Background()
	}
	if strings.TrimSpace(tmpl) == "" {
		return nil, nil
	}

	tracks, _, err := a.repo.ListTracks(a.ctx, domain.TrackFilter{Root: root, Limit: 50000})
	if err != nil {
		return nil, err
	}

	previews := make([]domain.FilenamePreview, 0)
	for _, t := range tracks {
		after := organize.ApplyTemplate(tmpl, t)
		if after != "" && after != t.Filename {
			previews = append(previews, domain.FilenamePreview{
				TrackID: t.ID,
				Before:  t.Filename,
				After:   after,
			})
		}
	}
	return previews, nil
}

func mergeStr(newVal, existing string) string {
	if strings.TrimSpace(newVal) == "" {
		return existing
	}
	return newVal
}

func mergeInt(newVal, existing int) int {
	if newVal == 0 {
		return existing
	}
	return newVal
}

func highConfStr(fields map[string]domain.FieldSuggestion, key string, threshold float64) string {
	if f, ok := fields[key]; ok && f.Confidence >= threshold {
		return f.Value
	}
	return ""
}

func highConfInt(fields map[string]domain.FieldSuggestion, key string, threshold float64) int {
	if f, ok := fields[key]; ok && f.Confidence >= threshold {
		n, _ := strconv.Atoi(f.Value)
		return n
	}
	return 0
}
