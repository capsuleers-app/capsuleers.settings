package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/WilliamFalci/capsuleers.settings/internal/accounts"
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
	mu      sync.Mutex // serializes account evidence/config writes
	ident   *identSession
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
	OK       bool     `json:"ok"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Backup   string   `json:"backup"`
	Updated  []string `json:"updated"`
	Accounts []string `json:"accounts"`
	Taken    string   `json:"taken"`
}

func result(r *backup.ApplyResult, err error) Result {
	out := Result{}
	if r != nil {
		out.Backup, out.Updated, out.Accounts, out.Taken = r.Backup, r.Updated, r.Accounts, r.Taken
	}
	if out.Updated == nil {
		out.Updated = []string{}
	}
	if out.Accounts == nil {
		out.Accounts = []string{}
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

// Apply backs up the targets and copies the source character file over each of them. With
// CopyUser it also copies the source account's core_user over the targets' accounts, planned
// here from the stored/suggested links; any target without a known account refuses the apply.
func (s *Service) Apply(req backup.ApplyRequest) Result {
	if !req.CopyUser {
		return result(s.backups.Apply(req, nil))
	}
	plan := s.AccountPlan(req)
	if !plan.Ready() {
		return Result{Code: "unknown_account", Updated: []string{}, Accounts: []string{}}
	}
	return result(s.backups.Apply(req, &backup.UserCopy{Source: plan.SourceUser, Targets: plan.Targets()}))
}

// AccountRow is the account of one character. UserID "" = unknown; Source is where the
// link comes from (launcher, launcher-esi, guided, portrait, rift).
type AccountRow struct {
	CharID string `json:"charId"`
	UserID string `json:"userId"`
	Source string `json:"source"`
}

// AccountsResult lists the profile's account files and each character's account.
type AccountsResult struct {
	Users []string `json:"users"`
	// Aliases: account id -> name chosen by the user (absent = none).
	Aliases map[string]string `json:"aliases"`
	Chars   []AccountRow      `json:"chars"`
}

// effective resolves the account of each character of a profile from certain evidence only:
// guided identifications and launcher logs (newest wins), strictly paired esi lines, then the
// profile files themselves. Links read from the launcher logs are stored in the config so they
// survive log rotation. EVE files are only read.
func (s *Service) effective(profileDir string) (map[string]accounts.Link, eve.Profile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := eve.ScanProfile(profileDir)
	users, chars := map[string]bool{}, map[string]bool{}
	for _, u := range p.Users {
		users[u.ID] = true
	}
	for _, c := range p.Chars {
		chars[c.ID] = true
	}
	now := time.Now()
	var ev, learned []accounts.Evidence
	for c, u := range s.cfg.AccountLinks(profileDir) { // legacy hand confirmations, undated
		if u != "" {
			ev = append(ev, accounts.Evidence{CharID: c, UserID: u, Source: accounts.SourceGuided})
		}
	}
	for c, l := range s.cfg.AccountGuided(profileDir) {
		ev = append(ev, accounts.Evidence{CharID: c, UserID: l.UserID, Source: accounts.SourceGuided, At: l.At})
	}
	for c, l := range s.cfg.AccountLearned(profileDir) {
		learned = append(learned, accounts.Evidence{CharID: c, UserID: l.UserID, Source: l.Source, At: l.At})
	}
	for _, e := range accounts.ScanLauncherLogs(profileDir) {
		if e.At.After(now) {
			e.At = now
		}
		learned = append(learned, e)
	}
	// persist what the logs taught (only changes are written)
	stored := s.cfg.AccountLearned(profileDir)
	next := map[string]config.Link{}
	for c, l := range accounts.Resolve(learned, nil, chars) {
		next[c] = config.Link{UserID: l.UserID, At: l.At, Source: l.Source}
	}
	if !sameLinks(stored, next) {
		_ = s.cfg.SetAccountLearned(profileDir, next)
	}
	ev = append(ev, learned...)
	for _, sg := range accounts.Suggest(profileDir) {
		ev = append(ev, accounts.Evidence{CharID: sg.CharID, UserID: sg.UserID, Source: sg.Reason})
	}
	return accounts.Resolve(ev, users, chars), p
}

func sameLinks(a, b map[string]config.Link) bool {
	if len(a) != len(b) {
		return false
	}
	for c, l := range a {
		if m, ok := b[c]; !ok || m.UserID != l.UserID || m.Source != l.Source || !m.At.Equal(l.At) {
			return false
		}
	}
	return true
}

// Accounts returns the character -> account mapping of a profile.
func (s *Service) Accounts(profileDir string) AccountsResult {
	links, p := s.effective(profileDir)
	out := AccountsResult{Users: []string{}, Chars: []AccountRow{}, Aliases: map[string]string{}}
	for u, a := range s.cfg.AccountAliases(profileDir) {
		out.Aliases[u] = a
	}
	for _, u := range p.Users {
		out.Users = append(out.Users, u.ID)
	}
	for _, c := range p.Chars {
		row := AccountRow{CharID: c.ID}
		if l, ok := links[c.ID]; ok {
			row.UserID, row.Source = l.UserID, l.Source
		}
		out.Chars = append(out.Chars, row)
	}
	return out
}

// SetAccountAlias names an account of the profile ("" removes the name). App config only.
func (s *Service) SetAccountAlias(profileDir, userID, alias string) error {
	for _, u := range eve.ScanProfile(profileDir).Users {
		if u.ID == userID {
			return s.cfg.SetAccountAlias(profileDir, userID, alias)
		}
	}
	return errors.New("unknown account")
}

// SetLastSource remembers the source picked in a profile, restored by LastSource at startup.
func (s *Service) SetLastSource(profileDir, charID string) error {
	for _, c := range eve.ScanProfile(profileDir).Chars {
		if c.ID == charID {
			return s.cfg.SetSource(profileDir, charID)
		}
	}
	return errors.New("unknown character")
}

// LastSource returns the remembered source of a profile, "" when none or no longer present.
func (s *Service) LastSource(profileDir string) string {
	id := s.cfg.Source(profileDir)
	for _, c := range eve.ScanProfile(profileDir).Chars {
		if c.ID == id {
			return id
		}
	}
	return ""
}

// AccountPlan shows which account files copying the source's core_user would overwrite.
func (s *Service) AccountPlan(req backup.ApplyRequest) accounts.Plan {
	acct := map[string]string{}
	for _, r := range s.Accounts(req.ProfileDir).Chars {
		acct[r.CharID] = r.UserID
	}
	return accounts.MakePlan(req.SourceID, req.TargetIDs, acct)
}

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
