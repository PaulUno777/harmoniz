package scanner

import (
	"context"
	"fmt"
	"harmoniz/internal/adapters/fs"
	"harmoniz/internal/core/domain"
	"harmoniz/internal/core/ports"
	"harmoniz/internal/logger"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dhowden/tag"
)

type Service struct {
	repo ports.TrackRepository
}

func NewService(repo ports.TrackRepository) *Service {
	return &Service{repo: repo}
}

// PrepareLibrarySync ensures the DB for this root is fresh. If the most recent added_at
// is older than maxAge (or there are no tracks), all tracks under root are deleted and
// a full scan is run. Otherwise nothing is done. Use this when opening/browsing a library.
func (s *Service) PrepareLibrarySync(ctx context.Context, root string, maxAge time.Duration) error {
	latest, err := s.repo.LatestAddedAtForRoot(ctx, root)
	if err != nil {
		return fmt.Errorf("checking library staleness: %w", err)
	}
	needResync := latest == 0 || time.Since(time.Unix(latest, 0)) > maxAge
	if !needResync {
		logger.Log.Info("Library still fresh, skipping re-sync", "root", root)
		return nil
	}
	logger.Log.Info("Library stale or empty, re-syncing from disk", "root", root)
	if err := s.repo.DeleteTracksForRoot(ctx, root); err != nil {
		return fmt.Errorf("deleting stale tracks: %w", err)
	}
	return s.Scan(ctx, root)
}

func (s *Service) Scan(ctx context.Context, root string) error {
	logger.Log.Info("Starting scan", "root", root)
	startTime := time.Now()

	// 1. Load Cache
	cache, err := s.repo.GetAllPathsModTime()
	if err != nil {
		return fmt.Errorf("failed to load cache: %w", err)
	}
	logger.Log.Info("Cache loaded", "entries", len(cache))

	// Channels
	pathChan := make(chan string, 1000)
	trackChan := make(chan domain.Track, 500)
	errChan := make(chan error, 1)

	// WaitGroups
	var workerWg sync.WaitGroup
	var writerWg sync.WaitGroup

	// 2. Spawn Workers
	numWorkers := runtime.NumCPU() * 4
	logger.Log.Info("Spawning workers", "count", numWorkers)

	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for path := range pathChan {
				select {
				case <-ctx.Done():
					return
				default:
					s.processPath(ctx, path, cache, trackChan)
				}
			}
		}()
	}

	// 3. Spawn Writer
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		if err := s.writerLoop(ctx, trackChan); err != nil {
			// Non-blocking send
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	// 4. Start Walking
	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Log.Error("Walk error", "path", path, "error", err)
			return nil // Skip unreadable
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".mp3" && ext != ".flac" && ext != ".m4a" && ext != ".ogg" && ext != ".wav" {
			return nil
		}

		select {
		case <-ctx.Done():
			return filepath.SkipAll
		case pathChan <- path:
		}
		return nil
	})

	close(pathChan)
	workerWg.Wait()
	close(trackChan)
	writerWg.Wait()

	if walkErr != nil {
		return walkErr
	}

	select {
	case err := <-errChan:
		return err
	default:
	}

	logger.Log.Info("Scan completed", "duration", time.Since(startTime).String())
	return nil
}

func (s *Service) processPath(ctx context.Context, path string, cache map[string]int64, out chan<- domain.Track) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	// Cache Check
	if cachedModTime, ok := cache[path]; ok {
		if info.ModTime().Unix() == cachedModTime {
			// Hit! Skip heavy processing
			return
		}
	}

	// Process File
	track, err := s.analyzeFile(path, info)
	if err != nil {
		logger.Log.Warn("Failed to analyze file", "path", path, "error", err)
		return
	}

	select {
	case <-ctx.Done():
		return
	case out <- track:
	}
}

func (s *Service) analyzeFile(path string, info os.FileInfo) (domain.Track, error) {
	f, err := os.Open(path)
	if err != nil {
		return domain.Track{}, err
	}
	defer f.Close()

	// Metadata
	m, err := tag.ReadFrom(f)
	artist := ""
	title := ""
	album := ""
	year := 0
	trackNum := 0

	if err == nil {
		artist = m.Artist()
		title = m.Title()
		album = m.Album()
		year = m.Year()
		trackNum, _ = m.Track()
	}

	// Hash
	hash, err := fs.ComputePartialHash(path)
	if err != nil {
		return domain.Track{}, err
	}

	// Create Track
	// Use path as fallback if metadata empty
	if title == "" {
		title = filepath.Base(path)
	}

	return domain.Track{
		Path:        path,
		Filename:    filepath.Base(path),
		Size:        info.Size(),
		ModTime:     info.ModTime().Unix(),
		ArtistRaw:   artist,
		ArtistNorm:  domain.NormalizeArtist(artist),
		AlbumRaw:    album,
		AlbumNorm:   domain.NormalizeArtist(album),
		Title:       title,
		Year:        year,
		TrackNum:    trackNum,
		HashPartial: hash,
		Status:      domain.StatusClean,
	}, nil
}

func (s *Service) writerLoop(ctx context.Context, in <-chan domain.Track) error {
	batchSize := 500
	buffer := make([]domain.Track, 0, batchSize)

	flush := func() error {
		if len(buffer) == 0 {
			return nil
		}
		if err := s.repo.BatchUpsert(buffer); err != nil {
			return err
		}
		buffer = buffer[:0]
		return nil
	}

	for track := range in {
		buffer = append(buffer, track)
		if len(buffer) >= batchSize {
			if err := flush(); err != nil {
				return err
			}
		}
	}

	return flush()
}

// ListTracks returns tracks for the library, optionally filtered by root path.
func (s *Service) ListTracks(ctx context.Context, root string, limit, offset int) ([]domain.Track, int, error) {
	return s.repo.ListTracks(ctx, root, limit, offset)
}
