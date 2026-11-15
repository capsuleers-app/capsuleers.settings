// Package update implements the silent self-update: look up the latest GitHub release, download the
// right asset next to the file it replaces, verify its SHA-256 and swap it in. The Wails wiring
// (dialog, shutdown hook) lives in the main package; everything here is testable on its own.
package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// LatestURL is the GitHub API endpoint of the latest release.
	LatestURL = "https://api.github.com/repos/capsuleers-app/capsuleers.settings/releases/latest"

	AssetLinux   = "capsuleers.settings-linux-amd64.AppImage"
	AssetWindows = "capsuleers.settings-windows-amd64.exe"

	// MaxSize caps the download (the AppImage is ~100 MB).
	MaxSize = 400 << 20

	tmpPrefix = ".capsuleers-update-"
	tmpSuffix = ".tmp"
)

// ParseVersion parses "<major>.<minor>.<run>" (optional leading "v"); anything else is not a release build.
func ParseVersion(s string) ([3]int, bool) {
	var v [3]int
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(s), "v"), ".")
	if len(parts) != 3 {
		return v, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || p != strconv.Itoa(n) {
			return v, false
		}
		v[i] = n
	}
	return v, true
}

// Newer reports whether latest is strictly newer than current (numeric, per component).
func Newer(current, latest string) bool {
	c, ok1 := ParseVersion(current)
	l, ok2 := ParseVersion(latest)
	if !ok1 || !ok2 {
		return false
	}
	for i := range c {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

// Target decides what can be updated: the asset to fetch and the file to replace. ok=false means
// "do nothing" (non-AppImage Linux, macOS). appImage is the $APPIMAGE env value, exe the running binary.
func Target(goos, appImage, exe string) (asset, file string, ok bool) {
	switch goos {
	case "linux":
		if appImage == "" {
			return "", "", false
		}
		return AssetLinux, appImage, true
	case "windows":
		if exe == "" {
			return "", "", false
		}
		if r, err := filepath.EvalSymlinks(exe); err == nil {
			exe = r
		}
		return AssetWindows, exe, true
	}
	return "", "", false
}

// Asset is the part of a GitHub release asset we use.
type Asset struct {
	Name   string `json:"name"`
	URL    string `json:"browser_download_url"`
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
}

// Release is the part of the GitHub release response we use.
type Release struct {
	Tag    string  `json:"tag_name"`
	Assets []Asset `json:"assets"`
}

// Version is the tag without the leading "v".
func (r Release) Version() string { return strings.TrimPrefix(r.Tag, "v") }

// FindAsset returns the asset with the given name.
func (r Release) FindAsset(name string) (Asset, bool) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a, true
		}
	}
	return Asset{}, false
}

// Pending is a verified download waiting to be swapped in.
type Pending struct {
	Version string
	Tmp     string // verified file, same folder as File
	File    string // file to replace
	Windows bool
}

// Updater checks and downloads. The zero value is not usable: set Current, UserAgent and Client.
type Updater struct {
	Current   string
	UserAgent string
	APIURL    string // default LatestURL
	Client    *http.Client
	MaxSize   int64 // default MaxSize
}

func (u *Updater) get(ctx context.Context, url, accept string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", u.UserAgent)
	req.Header.Set("Accept", accept)
	resp, err := u.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("update: %s: HTTP %d", url, resp.StatusCode)
	}
	return resp, nil
}

// Check looks for a newer release and, if there is one, downloads and verifies the asset.
// (nil, nil) means nothing to do.
func (u *Updater) Check(ctx context.Context, goos, appImage, exe string) (*Pending, error) {
	assetName, file, ok := Target(goos, appImage, exe)
	if !ok {
		return nil, nil
	}
	api := u.APIURL
	if api == "" {
		api = LatestURL
	}
	resp, err := u.get(ctx, api, "application/vnd.github+json")
	if err != nil {
		return nil, err
	}
	var rel Release
	err = json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&rel)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if !Newer(u.Current, rel.Version()) {
		return nil, nil
	}
	asset, ok := rel.FindAsset(assetName)
	if !ok || asset.URL == "" {
		return nil, errors.New("update: asset not in release")
	}
	want, ok := strings.CutPrefix(strings.ToLower(asset.Digest), "sha256:")
	if !ok || len(want) != 64 {
		return nil, errors.New("update: missing sha256 digest")
	}
	tmp, err := u.download(ctx, asset, want, filepath.Dir(file))
	if err != nil {
		return nil, err
	}
	return &Pending{Version: rel.Version(), Tmp: tmp, File: file, Windows: goos == "windows"}, nil
}

func (u *Updater) download(ctx context.Context, a Asset, wantSum, dir string) (string, error) {
	max := u.MaxSize
	if max <= 0 {
		max = MaxSize
	}
	if a.Size > max {
		return "", errors.New("update: asset too large")
	}
	resp, err := u.get(ctx, a.URL, "application/octet-stream")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	f, err := os.CreateTemp(dir, tmpPrefix+"*"+tmpSuffix)
	if err != nil {
		return "", err
	}
	name := f.Name()
	fail := func(err error) (string, error) {
		f.Close()
		os.Remove(name)
		return "", err
	}
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, h), io.LimitReader(resp.Body, max+1))
	if err != nil {
		return fail(err)
	}
	if n > max {
		return fail(errors.New("update: download exceeds size limit"))
	}
	if hex.EncodeToString(h.Sum(nil)) != wantSum {
		return fail(errors.New("update: sha256 mismatch"))
	}
	if err := f.Close(); err != nil {
		os.Remove(name)
		return "", err
	}
	return name, nil
}

// Apply swaps the verified file in. Linux: chmod 0755 + atomic rename. Windows: the running exe is
// renamed to <name>.old (removed on next start), then the new file takes its place.
func (p *Pending) Apply() error {
	return Swap(p.Tmp, p.File, p.Windows)
}

// Discard removes the downloaded file.
func (p *Pending) Discard() { os.Remove(p.Tmp) }

// Swap replaces target with src (same folder, so the rename is atomic).
func Swap(src, target string, windows bool) error {
	if err := os.Chmod(src, 0o755); err != nil {
		return err
	}
	if !windows {
		return os.Rename(src, target)
	}
	old := target + ".old"
	_ = os.Remove(old)
	if err := os.Rename(target, old); err != nil {
		return err
	}
	if err := os.Rename(src, target); err != nil {
		_ = os.Rename(old, target) // put the old one back
		return err
	}
	return nil
}

// Cleanup removes a leftover <file>.old (Windows) and orphaned temp downloads next to file.
func Cleanup(file string) {
	_ = os.Remove(file + ".old")
	dir := filepath.Dir(file)
	matches, _ := filepath.Glob(filepath.Join(dir, tmpPrefix+"*"+tmpSuffix))
	for _, m := range matches {
		_ = os.Remove(m)
	}
}

// Strings are the dialog texts.
type Strings struct {
	Title, Message, Detail, Later, Restart string
}

// Texts returns the dialog strings for a language ("it" or anything else = English).
func Texts(lang, version string) Strings {
	if lang == "it" {
		return Strings{
			Title:   "Aggiornamento disponibile",
			Message: fmt.Sprintf("La versione %s è stata scaricata.", version),
			Detail:  "Vuoi riavviare ora per installarla? Puoi anche farlo più tardi: verrà applicata alla prossima chiusura.",
			Later:   "Più tardi",
			Restart: "Riavvia e installa",
		}
	}
	return Strings{
		Title:   "Update available",
		Message: fmt.Sprintf("Version %s has been downloaded.", version),
		Detail:  "Restart now to install it? You can also do it later: it will be applied on next quit.",
		Later:   "Later",
		Restart: "Restart & install",
	}
}

// Lang maps a locale string (en_US.UTF-8, it-IT, ...) to "it" or "en".
func Lang(locale string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(locale)), "it") {
		return "it"
	}
	return "en"
}

// CheckTimeout bounds the whole check + download.
const CheckTimeout = 10 * time.Minute
