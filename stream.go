package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"harmoniz/internal/logger"
)

// MIME types for common audio formats
var audioMimeTypes = map[string]string{
	".mp3":  "audio/mpeg",
	".m4a":  "audio/mp4",
	".aac":  "audio/aac",
	".ogg":  "audio/ogg",
	".oga":  "audio/ogg",
	".wav":  "audio/wav",
	".flac": "audio/flac",
	".webm": "audio/webm",
}

func contentTypeForPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if mime, ok := audioMimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

// streamHandler handles GET /stream?path=... and streams the audio file with Range support for seeking.
func streamHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stream" {
			http.NotFound(w, r)
			return
		}
		
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := r.URL.Query().Get("path")
		if path == "" {
			http.Error(w, "missing path", http.StatusBadRequest)
			return
		}

		// Resolve to absolute path and ensure it's a regular file
		absPath, err := filepath.Abs(path)
		if err != nil {
			logger.Log.Warn("stream: resolve path", "path", path, "error", err)
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		info, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return
			}
			logger.Log.Warn("stream: stat", "path", absPath, "error", err)
			http.Error(w, "file error", http.StatusInternalServerError)
			return
		}
		if info.IsDir() {
			http.Error(w, "not a file", http.StatusBadRequest)
			return
		}

		f, err := os.Open(absPath)
		if err != nil {
			logger.Log.Warn("stream: open", "path", absPath, "error", err)
			http.Error(w, "file error", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		size := info.Size()
		contentType := contentTypeForPath(absPath)

		// Support Range requests for seeking
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Type", contentType)

		rangeHeader := r.Header.Get("Range")
		if rangeHeader == "" {
			w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
			w.WriteHeader(http.StatusOK)
			// Stream full file (copy in chunks to avoid loading entire file)
			buf := make([]byte, 32*1024)
			for {
				n, err := f.Read(buf)
				if n > 0 {
					if _, wErr := w.Write(buf[:n]); wErr != nil {
						return
					}
				}
				if err != nil {
					break
				}
			}
			return
		}

		// Parse "bytes=start-end"
		if !strings.HasPrefix(rangeHeader, "bytes=") {
			http.Error(w, "invalid range", http.StatusRequestedRangeNotSatisfiable)
			return
		}
		rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
		parts := strings.Split(rangeStr, "-")
		if len(parts) != 2 {
			http.Error(w, "invalid range", http.StatusRequestedRangeNotSatisfiable)
			return
		}
		start, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || start < 0 {
			http.Error(w, "invalid range", http.StatusRequestedRangeNotSatisfiable)
			return
		}
		var end int64 = size - 1
		if parts[1] != "" {
			end, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil || end >= size {
				end = size - 1
			}
		}
		if start > end {
			http.Error(w, "invalid range", http.StatusRequestedRangeNotSatisfiable)
			return
		}

		_, err = f.Seek(start, 0)
		if err != nil {
			http.Error(w, "seek error", http.StatusInternalServerError)
			return
		}

		contentLength := end - start + 1
		w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
		w.Header().Set("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(size, 10))
		w.WriteHeader(http.StatusPartialContent)

		toRead := contentLength
		buf := make([]byte, 32*1024)
		for toRead > 0 {
			n, err := f.Read(buf)
			if n > 0 {
				chunk := int64(n)
				if chunk > toRead {
					chunk = toRead
					n = int(toRead)
				}
				if _, wErr := w.Write(buf[:n]); wErr != nil {
					return
				}
				toRead -= chunk
			}
			if err != nil {
				break
			}
		}
	})
}
