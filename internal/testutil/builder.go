package testutil

import (
	"path/filepath"

	"harmoniz/internal/core/domain"
)

// TrackBuilder provides a fluent interface for constructing test domain.Track values.
type TrackBuilder struct {
	t domain.Track
}

// NewTrack initialises a builder with the given ID and path.
// Filename is derived automatically from the path base.
func NewTrack(id uint64, path string) *TrackBuilder {
	return &TrackBuilder{t: domain.Track{
		ID:       id,
		Path:     path,
		Filename: filepath.Base(path),
		Status:   domain.StatusClean,
	}}
}

// WithArtist sets ArtistRaw and computes ArtistNorm via domain.NormalizeArtist.
func (b *TrackBuilder) WithArtist(raw string) *TrackBuilder {
	b.t.ArtistRaw = raw
	b.t.ArtistNorm = domain.NormalizeArtist(raw)
	return b
}

// WithAlbum sets AlbumRaw.
func (b *TrackBuilder) WithAlbum(raw string) *TrackBuilder {
	b.t.AlbumRaw = raw
	return b
}

// WithTitle sets Title.
func (b *TrackBuilder) WithTitle(title string) *TrackBuilder {
	b.t.Title = title
	return b
}

// WithYear sets Year.
func (b *TrackBuilder) WithYear(y int) *TrackBuilder {
	b.t.Year = y
	return b
}

// WithTrackNum sets TrackNum.
func (b *TrackBuilder) WithTrackNum(n int) *TrackBuilder {
	b.t.TrackNum = n
	return b
}

// WithBitrate sets Bitrate.
func (b *TrackBuilder) WithBitrate(br int) *TrackBuilder {
	b.t.Bitrate = br
	return b
}

// WithSize sets the file size in bytes.
func (b *TrackBuilder) WithSize(size int64) *TrackBuilder {
	b.t.Size = size
	return b
}

// WithHashes sets HashPartial and HashFull.
// When HashFull is pre-populated, DeduplicationService skips file I/O — safe for unit tests.
func (b *TrackBuilder) WithHashes(partial, full string) *TrackBuilder {
	b.t.HashPartial = partial
	b.t.HashFull = full
	return b
}

// WithFilename overrides the filename (defaults to filepath.Base(path)).
func (b *TrackBuilder) WithFilename(name string) *TrackBuilder {
	b.t.Filename = name
	return b
}

// Build returns the constructed Track.
func (b *TrackBuilder) Build() domain.Track {
	return b.t
}

// DuplicateSet is a convenience builder for a set of byte-identical tracks used in
// deduplication tests. All tracks share the same Size, HashPartial, and HashFull.
func DuplicateSet(size int64, partial, full string, tracks ...*TrackBuilder) []domain.Track {
	result := make([]domain.Track, 0, len(tracks))
	for _, tb := range tracks {
		tb.WithSize(size).WithHashes(partial, full)
		result = append(result, tb.Build())
	}
	return result
}
