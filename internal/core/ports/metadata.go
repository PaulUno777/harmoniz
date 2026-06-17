package ports

// TrackMetadata holds the writable metadata fields for an audio file.
type TrackMetadata struct {
	Artist   string
	Title    string
	Album    string
	Year     int
	TrackNum int
}

// MetadataWriter writes metadata tags to audio files on disk.
// Implementations are format-specific (e.g. ID3v2 for MP3).
type MetadataWriter interface {
	WriteMetadata(path string, meta TrackMetadata) error
}
