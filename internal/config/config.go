// Package config persists the small amount of user state the app keeps:
// manually added settings folders. Everything lives in os.UserConfigDir()/capsuleers.settings.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

const AppDir = "capsuleers.settings"

// Dir returns the app config directory (created on demand by writers).
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, AppDir), nil
}

type Config struct {
	ExtraFolders []string `json:"extraFolders"`
}

// Store reads and writes config.json under a directory.
type Store struct {
	mu   sync.Mutex
	path string
}

func NewStore(dir string) *Store { return &Store{path: filepath.Join(dir, "config.json")} }

func (s *Store) Load() Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load()
}

func (s *Store) load() Config {
	var c Config
	b, err := os.ReadFile(s.path)
	if err == nil {
		_ = json.Unmarshal(b, &c)
	}
	return c
}

func (s *Store) save(c Config) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// AddFolder stores an absolute, existing directory once.
func (s *Store) AddFolder(p string) error {
	p = filepath.Clean(p)
	if !filepath.IsAbs(p) {
		return errors.New("folder must be an absolute path")
	}
	if st, err := os.Stat(p); err != nil || !st.IsDir() {
		return errors.New("not a folder")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.load()
	for _, f := range c.ExtraFolders {
		if f == p {
			return nil
		}
	}
	c.ExtraFolders = append(c.ExtraFolders, p)
	return s.save(c)
}

func (s *Store) RemoveFolder(p string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.load()
	out := c.ExtraFolders[:0]
	for _, f := range c.ExtraFolders {
		if f != p {
			out = append(out, f)
		}
	}
	c.ExtraFolders = out
	return s.save(c)
}
