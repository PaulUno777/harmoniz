package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"harmoniz/internal/core/domain"
)

// Store persists AppConfig as JSON in the application data directory.
type Store struct {
	path string
}

// NewStore creates a Store that persists config to <dataDir>/config.json.
func NewStore(dataDir string) *Store {
	return &Store{path: filepath.Join(dataDir, "config.json")}
}

// Load reads config from disk. Returns DefaultAppConfig if the file doesn't exist yet.
func (s *Store) Load() (domain.AppConfig, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return domain.DefaultAppConfig(), nil
	}
	if err != nil {
		return domain.DefaultAppConfig(), err
	}
	cfg := domain.DefaultAppConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return domain.DefaultAppConfig(), err
	}
	return cfg, nil
}

// Save atomically writes cfg to disk.
func (s *Store) Save(cfg domain.AppConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
