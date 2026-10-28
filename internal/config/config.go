// Package config persists the small amount of user state the app keeps:
// manually added settings folders and the character -> account links learned from certain sources. Everything lives in os.UserConfigDir()/capsuleers.settings.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var digitsRe = regexp.MustCompile(`^\d+$`)

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
	// Accounts: legacy (older versions): profile dir -> character id -> account id confirmed by hand.
	// Read-only now; counts as undated guided evidence.
	Accounts map[string]map[string]string `json:"accounts,omitempty"`
	// AccountLearned: links read from the EVE launcher logs, stored so they outlive log rotation.
	AccountLearned map[string]map[string]Link `json:"accountLearned,omitempty"`
	// AccountGuided: links found by the guided identification (user logged characters in and out).
	AccountGuided map[string]map[string]Link `json:"accountGuided,omitempty"`
	// AccountAliases: profile dir -> account (user) id -> display name chosen by the user.
	AccountAliases map[string]map[string]string `json:"accountAliases,omitempty"`
	// Sources: profile dir -> character id last picked as source, restored at startup.
	Sources map[string]string `json:"sources,omitempty"`
}

// MaxAliasLen is the longest account alias, in characters.
const MaxAliasLen = 40

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

// AccountLinks returns the legacy links the user confirmed by hand in older versions (nil when none).
func (s *Store) AccountLinks(profileDir string) map[string]string {
	return s.Load().Accounts[filepath.Clean(profileDir)]
}

// Link is a character -> account link with when and from what it was learned.
type Link struct {
	UserID string    `json:"userId"`
	At     time.Time `json:"at"`
	Source string    `json:"source"`
}

func validLink(profileDir, charID string, l Link) bool {
	return filepath.IsAbs(profileDir) && digitsRe.MatchString(charID) && digitsRe.MatchString(l.UserID) && l.Source != ""
}

// AccountLearned returns the links learned from the launcher logs (kept after the logs rotate).
func (s *Store) AccountLearned(profileDir string) map[string]Link {
	return s.Load().AccountLearned[filepath.Clean(profileDir)]
}

// AccountGuided returns the links found by the guided identification.
func (s *Store) AccountGuided(profileDir string) map[string]Link {
	return s.Load().AccountGuided[filepath.Clean(profileDir)]
}

// SetAccountLearned replaces the learned links of a profile.
func (s *Store) SetAccountLearned(profileDir string, links map[string]Link) error {
	profileDir = filepath.Clean(profileDir)
	for c, l := range links {
		if !validLink(profileDir, c, l) {
			return errors.New("invalid account link")
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.load()
	if len(links) == 0 {
		delete(c.AccountLearned, profileDir)
	} else {
		if c.AccountLearned == nil {
			c.AccountLearned = map[string]map[string]Link{}
		}
		c.AccountLearned[profileDir] = links
	}
	return s.save(c)
}

// SetAccountGuided stores one link found by the guided identification.
func (s *Store) SetAccountGuided(profileDir, charID string, l Link) error {
	profileDir = filepath.Clean(profileDir)
	if !validLink(profileDir, charID, l) {
		return errors.New("invalid account link")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.load()
	if c.AccountGuided == nil {
		c.AccountGuided = map[string]map[string]Link{}
	}
	if c.AccountGuided[profileDir] == nil {
		c.AccountGuided[profileDir] = map[string]Link{}
	}
	c.AccountGuided[profileDir][charID] = l
	return s.save(c)
}

// SetAccountAlias stores the display name of one account in one profile; empty (after trim) removes it.
func (s *Store) SetAccountAlias(profileDir, userID, alias string) error {
	profileDir = filepath.Clean(profileDir)
	alias = strings.TrimSpace(alias)
	if !filepath.IsAbs(profileDir) || !digitsRe.MatchString(userID) || utf8.RuneCountInString(alias) > MaxAliasLen {
		return errors.New("invalid account alias")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.load()
	if alias == "" {
		delete(c.AccountAliases[profileDir], userID)
		if len(c.AccountAliases[profileDir]) == 0 {
			delete(c.AccountAliases, profileDir)
		}
		return s.save(c)
	}
	if c.AccountAliases == nil {
		c.AccountAliases = map[string]map[string]string{}
	}
	if c.AccountAliases[profileDir] == nil {
		c.AccountAliases[profileDir] = map[string]string{}
	}
	c.AccountAliases[profileDir][userID] = alias
	return s.save(c)
}

// AccountAliases returns the aliases of a profile (nil when none).
func (s *Store) AccountAliases(profileDir string) map[string]string {
	return s.Load().AccountAliases[filepath.Clean(profileDir)]
}

// SetSource remembers the character last picked as source in one profile.
func (s *Store) SetSource(profileDir, charID string) error {
	profileDir = filepath.Clean(profileDir)
	if !filepath.IsAbs(profileDir) || !digitsRe.MatchString(charID) {
		return errors.New("invalid source")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.load()
	if c.Sources[profileDir] == charID {
		return nil
	}
	if c.Sources == nil {
		c.Sources = map[string]string{}
	}
	c.Sources[profileDir] = charID
	return s.save(c)
}

// Source returns the character last picked as source in a profile ("" when none).
func (s *Store) Source(profileDir string) string {
	return s.Load().Sources[filepath.Clean(profileDir)]
}
