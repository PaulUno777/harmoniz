package organize

import (
	"fmt"
	"strings"

	"harmoniz/internal/core/domain"
)

// DefaultTemplate is the default filename naming template.
const DefaultTemplate = "{artist} - {title}.{ext}"

// ApplyTemplate generates a filename from a template and track metadata.
// Placeholders: {artist} {album} {title} {year} {track} {track:02d} {ext}
// Missing values are replaced with sensible defaults ("Unknown", "").
func ApplyTemplate(tmpl string, t domain.Track) string {
	ext := fileExt(t.Filename)

	result := tmpl
	result = strings.ReplaceAll(result, "{artist}", sanitizeComp(t.ArtistRaw, "Unknown Artist"))
	result = strings.ReplaceAll(result, "{album}", sanitizeComp(t.AlbumRaw, "Unknown Album"))
	result = strings.ReplaceAll(result, "{title}", sanitizeComp(t.Title, "Untitled"))
	result = strings.ReplaceAll(result, "{year}", yearStr(t.Year))
	result = strings.ReplaceAll(result, "{track:02d}", trackStr(t.TrackNum, true))
	result = strings.ReplaceAll(result, "{track}", trackStr(t.TrackNum, false))
	result = strings.ReplaceAll(result, "{ext}", ext)

	return result
}

func fileExt(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx >= 0 {
		return filename[idx+1:]
	}
	return ""
}

func sanitizeComp(s, fallback string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	return SanitizeFilename(s)
}

func yearStr(year int) string {
	if year == 0 {
		return ""
	}
	return fmt.Sprintf("%d", year)
}

func trackStr(num int, pad bool) string {
	if num == 0 {
		return ""
	}
	if pad {
		return fmt.Sprintf("%02d", num)
	}
	return fmt.Sprintf("%d", num)
}
