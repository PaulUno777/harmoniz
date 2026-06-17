package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Adapter struct{}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) SafeMove(src, dst string) error {
	// 1. Create target directory
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 2. Try os.Rename (Atomic if on same device)
	var err error
	for i := 0; i < 3; i++ {
		err = os.Rename(src, dst)
		if err == nil {
			return nil
		}

		// If EXDEV, we must copy
		if isEXDEV(err) {
			return a.moveCrossDevice(src, dst)
		}

		// Retry for other errors (like file locked)
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("failed to rename after retries: %w", err)
}

func (a *Adapter) moveCrossDevice(src, dst string) error {
	// 1. Open source
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	// 2. Create destination
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	// 3. Copy
	if _, err := io.Copy(d, s); err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	// 4. Close handles before verification and deletion
	s.Close()
	d.Close()

	// 5. Verify (Hash check)
	srcHash, err := ComputeFullHash(src)
	if err != nil {
		return fmt.Errorf("failed to hash source: %w", err)
	}
	dstHash, err := ComputeFullHash(dst)
	if err != nil {
		return fmt.Errorf("failed to hash destination: %w", err)
	}

	if srcHash != dstHash {
		os.Remove(dst) // Cleanup corrupted copy
		return fmt.Errorf("integrity check failed: hash mismatch")
	}

	// 6. Delete original
	return os.Remove(src)
}

func isEXDEV(err error) bool {
	if err == nil {
		return false
	}
	// "cross-device link" or error code 18 (EXDEV)
	return strings.Contains(err.Error(), "cross-device link") || strings.Contains(err.Error(), "EXDEV")
}
