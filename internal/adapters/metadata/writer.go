package metadata

import (
	"fmt"
	"harmoniz/internal/core/ports"
	"path/filepath"
	"strconv"
	"strings"

	id3v2 "github.com/bogem/id3v2/v2"
)

// Adapter implements ports.MetadataWriter.
// Phase 5: MP3 (ID3v2) only. Other formats update DB only (see App.UpdateTrackTags).
type Adapter struct{}

func NewAdapter() *Adapter { return &Adapter{} }

// Ensure Adapter implements ports.MetadataWriter
var _ ports.MetadataWriter = (*Adapter)(nil)

func (a *Adapter) WriteMetadata(path string, meta ports.TrackMetadata) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp3":
		return writeMP3(path, meta)
	default:
		return fmt.Errorf("tag writing not yet supported for %s files (DB updated)", ext)
	}
}

func writeMP3(path string, meta ports.TrackMetadata) error {
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("open MP3 for tagging: %w", err)
	}
	defer tag.Close()

	if meta.Artist != "" {
		tag.SetArtist(meta.Artist)
	}
	if meta.Title != "" {
		tag.SetTitle(meta.Title)
	}
	if meta.Album != "" {
		tag.SetAlbum(meta.Album)
	}
	if meta.Year > 0 {
		tag.SetYear(strconv.Itoa(meta.Year))
	}
	if meta.TrackNum > 0 {
		tag.AddTextFrame(
			tag.CommonID("Track number"),
			tag.DefaultEncoding(),
			strconv.Itoa(meta.TrackNum),
		)
	}

	return tag.Save()
}
