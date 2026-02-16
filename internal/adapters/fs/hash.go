package fs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const (
	PartialHashLimit = 2 * 1024 * 1024 // 2MB
	ChunkSize        = 1 * 1024 * 1024 // 1MB
)

// ComputePartialHash calculates a hash based on file size strategy.
// - If size < 2MB: Computes Full SHA256 (to avoid overlap issues).
// - If size >= 2MB: Computes SHA256 of (First 1MB + Last 1MB).
func ComputePartialHash(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	size := info.Size()

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()

	// Strategy Check
	if size < PartialHashLimit {
		// SMALL FILE: Full Hash
		if _, err := io.Copy(hash, file); err != nil {
			return "", err
		}
	} else {
		// LARGE FILE: Head + Tail

		// 1. Read Head (First 1MB)
		head := make([]byte, ChunkSize)
		if _, err := io.ReadFull(file, head); err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return "", fmt.Errorf("failed to read head: %w", err)
		}
		hash.Write(head)

		// 2. Seek to Tail (Size - 1MB)
		if _, err := file.Seek(-ChunkSize, io.SeekEnd); err != nil {
			return "", fmt.Errorf("failed to seek to tail: %w", err)
		}

		// 3. Read Tail (Last 1MB)
		tail := make([]byte, ChunkSize)
		if _, err := io.ReadFull(file, tail); err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return "", fmt.Errorf("failed to read tail: %w", err)
		}
		hash.Write(tail)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// ComputeFullHash computes the standard SHA256 of the entire file.
func ComputeFullHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
