package accounts

import (
	"bufio"
	"encoding/base64"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// The EVE launcher logs are an undocumented format, so everything here is defensive. They also
// contain SSO tokens in launch arguments: only the three extracted values (userID, charID,
// character ids) are ever kept; no other part of a line is stored, logged or returned.

var (
	launcherDataRe = regexp.MustCompile(`/LauncherData=([A-Za-z0-9+/=_%-]{8,1000})`)
	launcherValRe  = regexp.MustCompile(`^[A-Za-z0-9._-]+:[A-Za-z0-9._-]+::(\d{1,12}):(\d{1,12})$`)
	fetchingRe     = regexp.MustCompile(`\[esi\] Fetching details for (\d{1,3}) character\(s\)[^\[\]]*characterIds\W{0,4}\[([\d,\s"']{0,512})\]`)
	fetchedRe      = regexp.MustCompile(`\[esi\] Fetched (\d{1,3}) character details`)
	userIDRe       = regexp.MustCompile(`"?userId"?\s*[:=]\s*"?(\d{1,12})`)
	idRe           = regexp.MustCompile(`\d{1,12}`)
	stampRe        = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})[T ](\d{2}):(\d{2}):(\d{2})`)
)

const (
	maxLogLine  = 4 << 20
	maxLogFiles = 40
	maxLogSize  = 256 << 20
)

// LogsDir derives the launcher log folder from a settings profile directory: the profile lives
// under <...>/AppData/Local/CCP/EVE/<server>/settings_*, the logs under <...>/AppData/Roaming/EVE Online/logs
// (Windows and Proton prefixes). "" when the profile is not under an AppData folder.
func LogsDir(profileDir string) string {
	d := filepath.Clean(profileDir)
	for {
		parent := filepath.Dir(d)
		if parent == d {
			return ""
		}
		if filepath.Base(d) == "AppData" {
			return filepath.Join(d, "Roaming", "EVE Online", "logs")
		}
		d = parent
	}
}

// ScanLauncherLogs reads every eve-online-launcher-*.log of the profile's prefix (read-only) and
// returns the deterministic evidence found. Missing logs yield nil.
func ScanLauncherLogs(profileDir string) []Evidence {
	dir := LogsDir(profileDir)
	if dir == "" {
		return nil
	}
	names, _ := filepath.Glob(filepath.Join(dir, "eve-online-launcher-*.log"))
	sort.Strings(names)
	if len(names) > maxLogFiles {
		names = names[len(names)-maxLogFiles:]
	}
	var all []Evidence
	for _, n := range names {
		f, err := os.Open(n)
		if err != nil {
			continue
		}
		info, err := f.Stat()
		if err == nil && info.Size() <= maxLogSize {
			all = append(all, ParseLauncherLog(f, info.ModTime())...)
		}
		f.Close()
	}
	return dropMultiAccountESI(all)
}

// dropMultiAccountESI removes the esi evidence of characters that appear on more than one account.
func dropMultiAccountESI(ev []Evidence) []Evidence {
	users := map[string]map[string]bool{}
	for _, e := range ev {
		if e.Source == SourceLauncherESI {
			if users[e.CharID] == nil {
				users[e.CharID] = map[string]bool{}
			}
			users[e.CharID][e.UserID] = true
		}
	}
	out := ev[:0:0]
	for _, e := range ev {
		if e.Source == SourceLauncherESI && len(users[e.CharID]) > 1 {
			continue
		}
		out = append(out, e)
	}
	return out
}

// lineTime reads a leading ISO-like stamp (treated as UTC), else fallback.
func lineTime(line string, fallback time.Time) time.Time {
	head := line
	if len(head) > 64 {
		head = head[:64]
	}
	m := stampRe.FindStringSubmatch(head)
	if m == nil {
		return fallback
	}
	t, err := time.Parse("2006-01-02T15:04:05", m[1]+"-"+m[2]+"-"+m[3]+"T"+m[4]+":"+m[5]+":"+m[6])
	if err != nil {
		return fallback
	}
	return t
}

func decodeLauncherData(s string) (user, char string, ok bool) {
	if u, err := url.PathUnescape(s); err == nil {
		s = u
	}
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		b, err := enc.DecodeString(s)
		if err != nil {
			continue
		}
		if m := launcherValRe.FindStringSubmatch(string(b)); m != nil {
			return m[1], m[2], true
		}
	}
	return "", "", false
}

type pendingFetch struct {
	n   int
	ids []string
}

// ParseLauncherLog extracts evidence from one log. fallback stamps lines with no readable time.
// Source A: /LauncherData=<base64(product:tenant::userID:charID)>. Source B: a
// "Fetching details for N character(s)" line answered by "Fetched N character details {userId...}",
// paired only when exactly one fetch of that size is open (anything ambiguous is dropped).
func ParseLauncherLog(r io.Reader, fallback time.Time) []Evidence {
	var out []Evidence
	var pending []pendingFetch
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64<<10), maxLogLine)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.Contains(line, "/LauncherData="):
			if m := launcherDataRe.FindStringSubmatch(line); m != nil {
				if u, c, ok := decodeLauncherData(m[1]); ok {
					out = append(out, Evidence{CharID: c, UserID: u, Source: SourceLauncher, At: lineTime(line, fallback)})
				}
			}
		case strings.Contains(line, "[esi] Fetching details"):
			if m := fetchingRe.FindStringSubmatch(line); m != nil {
				n := atoi(m[1])
				ids := idRe.FindAllString(m[2], -1)
				if n > 0 && len(ids) == n {
					pending = append(pending, pendingFetch{n, ids})
				}
			}
		case strings.Contains(line, "[esi] Fetched"):
			m := fetchedRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			n := atoi(m[1])
			var same, rest []pendingFetch
			for _, p := range pending {
				if p.n == n {
					same = append(same, p)
				} else {
					rest = append(rest, p)
				}
			}
			pending = rest
			if len(same) != 1 {
				continue
			}
			us := userIDRe.FindAllStringSubmatch(line, -1)
			if len(us) == 0 {
				continue
			}
			user := us[0][1]
			for _, u := range us {
				if u[1] != user {
					user = ""
				}
			}
			if user == "" {
				continue
			}
			at := lineTime(line, fallback)
			for _, c := range same[0].ids {
				out = append(out, Evidence{CharID: c, UserID: user, Source: SourceLauncherESI, At: at})
			}
		}
	}
	return out
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
