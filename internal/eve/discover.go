// Package eve discovers EVE Online client settings on disk and detects a running client.
package eve

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	charFileRe = regexp.MustCompile(`^core_char_(\d+)\.dat$`)
	userFileRe = regexp.MustCompile(`^core_user_(\d+)\.dat$`)
	vdfPathRe  = regexp.MustCompile(`"path"\s+"([^"]+)"`)
)

// File is one core_char_<id>.dat or core_user_<id>.dat file.
type File struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	Saved string `json:"saved"` // RFC3339
}

type Profile struct {
	Name  string `json:"name"` // settings_Default
	Dir   string `json:"dir"`
	Chars []File `json:"chars"`
	Users []File `json:"users"`
	Saved string `json:"saved"` // newest file mtime, RFC3339
}

type Server struct {
	Name     string    `json:"name"` // Tranquility
	Dir      string    `json:"dir"`  // c_ccp_eve_online_tq_tranquility
	Profiles []Profile `json:"profiles"`
}

type Install struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"` // Steam, Windows, macOS, folder name
	Kind    string   `json:"kind"` // Proton, Native, Manual
	Path    string   `json:"path"`
	Manual  bool     `json:"manual"`
	Servers []Server `json:"servers"`
}

// Env is the host description, injectable for tests.
type Env struct {
	GOOS         string
	Home         string
	LocalAppData string
}

// CurrentEnv reads the real environment.
func CurrentEnv(goos string) Env {
	home, _ := os.UserHomeDir()
	return Env{GOOS: goos, Home: home, LocalAppData: os.Getenv("LOCALAPPDATA")}
}

// ParseCharID returns the numeric id of core_char_<digits>.dat, ignoring other variants.
func ParseCharID(name string) (string, bool) { return match(charFileRe, name) }

// ParseUserID returns the numeric id of core_user_<digits>.dat.
func ParseUserID(name string) (string, bool) { return match(userFileRe, name) }

func match(re *regexp.Regexp, name string) (string, bool) {
	m := re.FindStringSubmatch(name)
	if m == nil {
		return "", false
	}
	return m[1], true
}

type candidate struct{ name, kind, path string }

const protonSuffix = "steamapps/compatdata/8500/pfx/drive_c/users/steamuser/AppData/Local/CCP/EVE"

// Candidates lists the well-known settings roots for the host (existing or not).
func Candidates(env Env) []candidate {
	var out []candidate
	switch env.GOOS {
	case "windows":
		if env.LocalAppData != "" {
			out = append(out, candidate{"Windows", "Native", filepath.Join(env.LocalAppData, "CCP", "EVE")})
		}
	case "darwin":
		out = append(out, candidate{"macOS", "Native", filepath.Join(env.Home, "Library", "Application Support", "CCP", "EVE")})
	default:
		steams := []candidate{
			{"Steam", "Proton", filepath.Join(env.Home, ".local/share/Steam")},
			{"Steam", "Proton", filepath.Join(env.Home, ".steam/steam")},
			{"Steam (Flatpak)", "Proton", filepath.Join(env.Home, ".var/app/com.valvesoftware.Steam/data/Steam")},
		}
		for _, s := range steams {
			out = append(out, candidate{s.name, s.kind, filepath.Join(s.path, filepath.FromSlash(protonSuffix))})
			for _, lib := range steamLibraries(s.path) {
				out = append(out, candidate{s.name, s.kind, filepath.Join(lib, filepath.FromSlash(protonSuffix))})
			}
		}
	}
	return out
}

func steamLibraries(steamRoot string) []string {
	b, err := os.ReadFile(filepath.Join(steamRoot, "steamapps", "libraryfolders.vdf"))
	if err != nil {
		return nil
	}
	var out []string
	for _, m := range vdfPathRe.FindAllStringSubmatch(string(b), -1) {
		out = append(out, strings.ReplaceAll(m[1], `\\`, `\`))
	}
	return out
}

// Discover scans the well-known roots plus manually added folders. Roots that do not exist or
// hold no profile are skipped; duplicates (symlinks, repeated libraries) are collapsed.
func Discover(env Env, extra []string) []Install {
	seen := map[string]bool{}
	var out []Install
	add := func(name, kind, path string, manual bool) {
		real, err := filepath.EvalSymlinks(path)
		if err != nil || seen[real] {
			return
		}
		seen[real] = true
		servers := ScanRoot(real)
		if len(servers) == 0 {
			return
		}
		out = append(out, Install{ID: real, Name: name, Kind: kind, Path: path, Manual: manual, Servers: servers})
	}
	for _, c := range Candidates(env) {
		add(c.name, c.kind, c.path, false)
	}
	for _, p := range extra {
		add(filepath.Base(p), "Manual", p, true)
	}
	return out
}

// ScanRoot lists server dirs under root. A folder that directly holds settings_* profiles is
// accepted as a single server (handy for manually added folders).
func ScanRoot(root string) []Server {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var servers []Server
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), "settings_") {
			continue
		}
		dir := filepath.Join(root, e.Name())
		if profs := scanProfiles(dir); len(profs) > 0 {
			servers = append(servers, Server{Name: serverName(e.Name()), Dir: dir, Profiles: profs})
		}
	}
	if len(servers) == 0 {
		if profs := scanProfiles(root); len(profs) > 0 {
			servers = append(servers, Server{Name: serverName(filepath.Base(root)), Dir: root, Profiles: profs})
		}
	}
	sort.Slice(servers, func(i, j int) bool { return servers[i].Name < servers[j].Name })
	return servers
}

func serverName(dir string) string {
	n := dir
	if i := strings.LastIndex(dir, "_"); i >= 0 && i < len(dir)-1 {
		n = dir[i+1:]
	}
	if n == "" {
		return dir
	}
	return strings.ToUpper(n[:1]) + n[1:]
}

func scanProfiles(serverDir string) []Profile {
	entries, err := os.ReadDir(serverDir)
	if err != nil {
		return nil
	}
	var out []Profile
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "settings_") {
			continue
		}
		out = append(out, ScanProfile(filepath.Join(serverDir, e.Name())))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ScanProfile reads the core_char / core_user files of one settings_* directory.
func ScanProfile(dir string) Profile {
	p := Profile{Name: filepath.Base(dir), Dir: dir, Chars: []File{}, Users: []File{}}
	entries, _ := os.ReadDir(dir)
	var newest time.Time
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		f := File{Path: filepath.Join(dir, e.Name()), Size: info.Size(), Saved: info.ModTime().UTC().Format(time.RFC3339)}
		if id, ok := ParseCharID(e.Name()); ok {
			f.ID = id
			p.Chars = append(p.Chars, f)
		} else if id, ok := ParseUserID(e.Name()); ok {
			f.ID = id
			p.Users = append(p.Users, f)
		} else {
			continue
		}
		if info.ModTime().After(newest) {
			newest = info.ModTime()
		}
	}
	sort.Slice(p.Chars, func(i, j int) bool { return p.Chars[i].ID < p.Chars[j].ID })
	sort.Slice(p.Users, func(i, j int) bool { return p.Users[i].ID < p.Users[j].ID })
	if !newest.IsZero() {
		p.Saved = newest.UTC().Format(time.RFC3339)
	}
	return p
}
