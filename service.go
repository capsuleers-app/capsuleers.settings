package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/WilliamFalci/capsuleers.settings/internal/backup"
	"github.com/WilliamFalci/capsuleers.settings/internal/config"
	"github.com/WilliamFalci/capsuleers.settings/internal/esi"
	"github.com/WilliamFalci/capsuleers.settings/internal/eve"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Service is the Go surface exposed to the frontend (bindings are generated from it).
type Service struct {
	cfg     *config.Store
	names   *esi.Resolver
	backups *backup.Manager
}

func NewService() (*Service, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}
	s := &Service{cfg: config.NewStore(dir), names: esi.NewResolver(dir)}
	s.backups = &backup.Manager{
		Dir:     filepath.Join(dir, "backups"),
		Running: eve.RunningCount,
		NameOf:  func(id string) string { return s.names.Cached([]string{id})[id] },
	}
	return s, nil
}

// Meta is static app info for the footer.
type Meta struct {
	Version string `json:"version"`
	Repo    string `json:"repo"`
}

func (s *Service) Meta() Meta { return Meta{Version: esi.AppVersion, Repo: esi.RepoURL} }

// ClientStatus returns how many EVE client processes run right now.
func (s *Service) ClientStatus() int { return eve.RunningCount() }

// Installs scans well-known roots plus manually added folders.
func (s *Service) Installs() []eve.Install {
	got := eve.Discover(eve.CurrentEnv(runtime.GOOS), s.cfg.Load().ExtraFolders)
	if got == nil {
		got = []eve.Install{}
	}
	return got
}

// Profile re-reads one profile directory (after external changes or an apply).
func (s *Service) Profile(dir string) eve.Profile { return eve.ScanProfile(dir) }

// AddFolder asks for a folder and stores it. Empty string means the dialog was cancelled.
func (s *Service) AddFolder() (string, error) {
	p, err := application.Get().Dialog.OpenFile().CanChooseDirectories(true).CanChooseFiles(false).
		SetTitle("EVE settings folder").PromptForSingleSelection()
	if err != nil || p == "" {
		return "", nil
	}
	return p, s.cfg.AddFolder(p)
}

func (s *Service) RemoveFolder(path string) error { return s.cfg.RemoveFolder(path) }

// CachedNames returns known names without network access.
func (s *Service) CachedNames(ids []string) map[string]string { return s.names.Cached(ids) }

// ResolveNames fetches missing names from ESI (offline: returns what the cache has).
func (s *Service) ResolveNames(ids []string) map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return s.names.Resolve(ctx, ids)
}

// Result is the outcome of Apply and Restore; Code is machine readable when OK is false.
type Result struct {
	OK      bool     `json:"ok"`
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Backup  string   `json:"backup"`
	Updated []string `json:"updated"`
	Taken   string   `json:"taken"`
}

func result(r *backup.ApplyResult, err error) Result {
	out := Result{}
	if r != nil {
		out.Backup, out.Updated, out.Taken = r.Backup, r.Updated, r.Taken
	}
	if out.Updated == nil {
		out.Updated = []string{}
	}
	switch {
	case err == nil:
		out.OK = true
	case errors.Is(err, backup.ErrRunning):
		out.Code = "client_running"
	case errors.Is(err, backup.ErrBadRequest):
		out.Code, out.Message = "bad_request", strings.TrimPrefix(strings.TrimPrefix(err.Error(), backup.ErrBadRequest.Error()), ": ")
	case errors.Is(err, backup.ErrBadBackup):
		out.Code = "bad_backup"
	default:
		out.Code, out.Message = "io", err.Error()
	}
	return out
}

// Apply backs up the targets and copies the source character file over each of them.
func (s *Service) Apply(req backup.ApplyRequest) Result { return result(s.backups.Apply(req)) }

func (s *Service) Backups() []backup.Info { return s.backups.List() }

func (s *Service) Restore(name string) Result { return result(s.backups.Restore(name)) }

func (s *Service) DeleteBackup(name string) error { return s.backups.Delete(name) }

// OpenBackupsFolder shows the backups directory in the file manager.
func (s *Service) OpenBackupsFolder() error {
	if err := os.MkdirAll(s.backups.Dir, 0o700); err != nil {
		return err
	}
	return application.Get().Browser.OpenFile(s.backups.Dir)
}
