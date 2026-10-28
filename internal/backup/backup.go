// Package backup copies one core_char file over many targets, always after zipping the files
// it is about to overwrite, and restores such zips. Nothing is written outside the profile dir
// recorded in the backup manifest / chosen by the caller.
package backup

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const KeepLast = 20

var (
	ErrRunning     = errors.New("client_running")
	ErrBadRequest  = errors.New("bad_request")
	ErrBadBackup   = errors.New("bad_backup")
	zipNameRe      = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{6}(-\d+)?\.zip$`)
	datNameRe      = regexp.MustCompile(`^core_(char|user)_\d+\.dat$`)
	charNameRe     = regexp.MustCompile(`^core_char_(\d+)\.dat$`)
	profileDirName = "settings_"
)

type Entry struct {
	Zip      string `json:"zip"`      // entry name inside the zip
	Original string `json:"original"` // absolute path of the file that was backed up
	CharID   string `json:"charId"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
}

type Manifest struct {
	Version    int     `json:"version"`
	Kind       string  `json:"kind"` // apply | restore
	CreatedAt  string  `json:"createdAt"`
	SourceID   string  `json:"sourceId"`
	SourceName string  `json:"sourceName"`
	ProfileDir string  `json:"profileDir"`
	Profile    string  `json:"profile"`
	Files      []Entry `json:"files"`
}

// Info is the list-view summary of a stored backup.
type Info struct {
	Name       string  `json:"name"`
	Taken      string  `json:"taken"`
	Kind       string  `json:"kind"`
	SourceName string  `json:"sourceName"`
	Profile    string  `json:"profile"`
	ProfileDir string  `json:"profileDir"`
	Size       int64   `json:"size"`
	Files      []Entry `json:"files"`
}

// Manager owns a backups directory. Running reports live EVE clients; NameOf is optional.
type Manager struct {
	Dir     string
	Running func() int
	NameOf  func(charID string) string
	Now     func() time.Time
}

type ApplyRequest struct {
	ProfileDir string   `json:"profileDir"`
	SourceID   string   `json:"sourceId"`
	TargetIDs  []string `json:"targetIds"`
}

type ApplyResult struct {
	Backup  string   `json:"backup"`
	Updated []string `json:"updated"`
	Taken   string   `json:"taken"`
}

func (m *Manager) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now()
}

func (m *Manager) name(id string) string {
	if m.NameOf != nil {
		return m.NameOf(id)
	}
	return ""
}

func (m *Manager) checkRunning() error {
	if m.Running != nil && m.Running() > 0 {
		return ErrRunning
	}
	return nil
}

func validProfileDir(dir string) bool {
	return filepath.IsAbs(dir) && strings.HasPrefix(filepath.Base(dir), profileDirName)
}

// Apply backs up every target core_char file and then replaces each one with the source file.
func (m *Manager) Apply(req ApplyRequest) (*ApplyResult, error) {
	if err := m.checkRunning(); err != nil {
		return nil, err
	}
	dir := filepath.Clean(req.ProfileDir)
	if !validProfileDir(dir) || req.SourceID == "" || len(req.TargetIDs) == 0 {
		return nil, ErrBadRequest
	}
	srcPath := filepath.Join(dir, "core_char_"+req.SourceID+".dat")
	if !charNameRe.MatchString(filepath.Base(srcPath)) {
		return nil, ErrBadRequest
	}
	if st, err := os.Stat(srcPath); err != nil || !st.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: source file missing", ErrBadRequest)
	}
	seen := map[string]bool{req.SourceID: true}
	var targets []string
	for _, id := range req.TargetIDs {
		if seen[id] {
			continue
		}
		seen[id] = true
		p := filepath.Join(dir, "core_char_"+id+".dat")
		if !charNameRe.MatchString(filepath.Base(p)) {
			return nil, ErrBadRequest
		}
		if st, err := os.Stat(p); err != nil || !st.Mode().IsRegular() {
			return nil, fmt.Errorf("%w: target %s missing", ErrBadRequest, id)
		}
		targets = append(targets, p)
	}
	if len(targets) == 0 {
		return nil, ErrBadRequest
	}
	zipName, err := m.write("apply", req.SourceID, dir, targets)
	if err != nil {
		return nil, err
	}
	res := &ApplyResult{Backup: zipName, Taken: m.now().UTC().Format(time.RFC3339)}
	for _, t := range targets {
		if err := copyAtomic(srcPath, t); err != nil {
			return res, fmt.Errorf("copy to %s: %w", filepath.Base(t), err)
		}
		id, _ := charID(t)
		res.Updated = append(res.Updated, id)
	}
	m.prune()
	return res, nil
}

// Restore puts the files of a backup back. The current files are backed up first.
func (m *Manager) Restore(zipName string) (*ApplyResult, error) {
	if err := m.checkRunning(); err != nil {
		return nil, err
	}
	zr, mf, err := m.open(zipName)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	dir := filepath.Clean(mf.ProfileDir)
	if !validProfileDir(dir) {
		return nil, ErrBadBackup
	}
	byName := map[string]*zip.File{}
	for _, f := range zr.File {
		byName[f.Name] = f
	}
	// validate everything before touching disk
	var current []string
	for _, e := range mf.Files {
		o := filepath.Clean(e.Original)
		if filepath.Dir(o) != dir || !datNameRe.MatchString(filepath.Base(o)) || byName[e.Zip] == nil {
			return nil, ErrBadBackup
		}
		if st, err := os.Stat(o); err == nil && st.Mode().IsRegular() {
			current = append(current, o)
		}
	}
	res := &ApplyResult{Taken: m.now().UTC().Format(time.RFC3339)}
	if len(current) > 0 {
		if res.Backup, err = m.write("restore", "", dir, current); err != nil {
			return nil, err
		}
	}
	for _, e := range mf.Files {
		if err := extractAtomic(byName[e.Zip], filepath.Clean(e.Original)); err != nil {
			return res, err
		}
		res.Updated = append(res.Updated, e.CharID)
	}
	m.prune()
	return res, nil
}

// List returns stored backups, newest first. Unreadable zips are skipped.
func (m *Manager) List() []Info {
	names := m.zipNames()
	out := []Info{}
	for i := len(names) - 1; i >= 0; i-- {
		zr, mf, err := m.open(names[i])
		if err != nil {
			continue
		}
		zr.Close()
		st, _ := os.Stat(filepath.Join(m.Dir, names[i]))
		var size int64
		if st != nil {
			size = st.Size()
		}
		out = append(out, Info{Name: names[i], Taken: mf.CreatedAt, Kind: mf.Kind, SourceName: mf.SourceName,
			Profile: mf.Profile, ProfileDir: mf.ProfileDir, Size: size, Files: mf.Files})
	}
	return out
}

func (m *Manager) Delete(zipName string) error {
	if !zipNameRe.MatchString(zipName) {
		return ErrBadRequest
	}
	return os.Remove(filepath.Join(m.Dir, zipName))
}

func (m *Manager) zipNames() []string {
	entries, _ := os.ReadDir(m.Dir)
	var names []string
	for _, e := range entries {
		if !e.IsDir() && zipNameRe.MatchString(e.Name()) {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}

func (m *Manager) prune() {
	names := m.zipNames()
	for len(names) > KeepLast {
		_ = os.Remove(filepath.Join(m.Dir, names[0]))
		names = names[1:]
	}
}

func (m *Manager) open(zipName string) (*zip.ReadCloser, *Manifest, error) {
	if !zipNameRe.MatchString(zipName) {
		return nil, nil, ErrBadRequest
	}
	zr, err := zip.OpenReader(filepath.Join(m.Dir, zipName))
	if err != nil {
		return nil, nil, ErrBadBackup
	}
	for _, f := range zr.File {
		if f.Name != "manifest.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			break
		}
		var mf Manifest
		err = json.NewDecoder(io.LimitReader(rc, 1<<20)).Decode(&mf)
		rc.Close()
		if err != nil {
			break
		}
		return zr, &mf, nil
	}
	zr.Close()
	return nil, nil, ErrBadBackup
}

// write zips the given existing files with a manifest and returns the zip file name.
func (m *Manager) write(kind, sourceID, profileDir string, files []string) (string, error) {
	if err := os.MkdirAll(m.Dir, 0o700); err != nil {
		return "", err
	}
	t := m.now()
	base := t.Format("2006-01-02_150405")
	name := base + ".zip"
	for i := 2; fileExists(filepath.Join(m.Dir, name)); i++ {
		name = base + "-" + strconv.Itoa(i) + ".zip"
	}
	final := filepath.Join(m.Dir, name)
	tmp := final + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return "", err
	}
	ok := false
	defer func() {
		if !ok {
			out.Close()
			os.Remove(tmp)
		}
	}()
	zw := zip.NewWriter(out)
	mf := Manifest{Version: 1, Kind: kind, CreatedAt: t.UTC().Format(time.RFC3339), SourceID: sourceID,
		SourceName: m.name(sourceID), ProfileDir: profileDir, Profile: filepath.Base(profileDir)}
	for i, f := range files {
		st, err := os.Stat(f)
		if err != nil {
			return "", err
		}
		id, _ := charID(f)
		entry := fmt.Sprintf("files/%03d_%s", i+1, filepath.Base(f))
		w, err := zw.Create(entry)
		if err != nil {
			return "", err
		}
		src, err := os.Open(f)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(w, src)
		src.Close()
		if err != nil {
			return "", err
		}
		mf.Files = append(mf.Files, Entry{Zip: entry, Original: f, CharID: id, Name: m.name(id), Size: st.Size()})
	}
	w, err := zw.Create("manifest.json")
	if err != nil {
		return "", err
	}
	if err := json.NewEncoder(w).Encode(mf); err != nil {
		return "", err
	}
	if err := zw.Close(); err != nil {
		return "", err
	}
	if err := out.Sync(); err != nil {
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, final); err != nil {
		return "", err
	}
	ok = true
	return name, nil
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

// charID extracts the numeric id of a core_char_/core_user_ path (any kind).
func charID(p string) (string, bool) {
	b := strings.TrimSuffix(filepath.Base(p), ".dat")
	i := strings.LastIndex(b, "_")
	if i < 0 {
		return "", false
	}
	return b[i+1:], true
}

func copyAtomic(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	return writeAtomic(dst, in)
}

func extractAtomic(zf *zip.File, dst string) error {
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	return writeAtomic(dst, rc)
}

// writeAtomic writes r to a temp file next to dst, then renames it over dst.
func writeAtomic(dst string, r io.Reader) error {
	dir := filepath.Dir(dst)
	mode := os.FileMode(0o644)
	if st, err := os.Stat(dst); err == nil {
		mode = st.Mode().Perm()
	}
	tmp, err := os.CreateTemp(dir, ".capsuleers-*.tmp")
	if err != nil {
		return err
	}
	tn := tmp.Name()
	ok := false
	defer func() {
		if !ok {
			tmp.Close()
			os.Remove(tn)
		}
	}()
	if _, err := io.Copy(tmp, r); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	_ = os.Chmod(tn, mode)
	if err := os.Rename(tn, dst); err != nil {
		return err
	}
	ok = true
	return nil
}
