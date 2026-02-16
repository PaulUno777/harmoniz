package domain

import (
	"strings"
	"unicode"
)

type TrackStatus string

const (
	StatusClean    TrackStatus = "clean"
	StatusModified TrackStatus = "modified"
	StatusError    TrackStatus = "error"
)

type Track struct {
	ID       uint64 `json:"id"`
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	ModTime  int64  `json:"mod_time"` // Unix timestamp

	// Metadata
	ArtistRaw  string `json:"artist_raw"`
	ArtistNorm string `json:"artist_norm"`
	AlbumRaw   string `json:"album_raw"`
	AlbumNorm  string `json:"album_norm"`
	Title      string `json:"title"`
	Year       int    `json:"year"`
	TrackNum   int    `json:"track_num"`
	Bitrate    int    `json:"bitrate"`

	// Identity
	HashPartial string `json:"hash_partial"`
	HashFull    string `json:"hash_full"`
	Fingerprint string `json:"fingerprint"`

	// Status
	IsDeleted    bool        `json:"is_deleted"`
	DeletedAt    *int64      `json:"deleted_at"`    // Unix timestamp
	DeleteReason string      `json:"delete_reason"` // e.g., "PRUNE", "USER"
	Status       TrackStatus `json:"status"`
}

// ListTracksResult is the single return value for ListTracks (Wails serializes it to JS).
type ListTracksResult struct {
	Tracks []Track `json:"Tracks"`
	Total  int     `json:"Total"`
}

// NormalizeArtist removes accents, punctuation, and converts to lowercase
// Example: "Guns N' Roses" -> "guns n roses"
func NormalizeArtist(input string) string {
	return normalize(input)
}

func normalize(input string) string {
	// Simple normalization for now, can be improved with advanced unicode handling
	// 1. Lowercase
	s := strings.ToLower(input)

	// 2. Remove accents (simplified approach, better to use transform/norm in robust impl)
	// For now, let's just keep alphanumeric and spaces
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}

	// 3. Collapse spaces
	return strings.Join(strings.Fields(b.String()), " ")
}
