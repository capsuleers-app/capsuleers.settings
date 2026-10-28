// Package accounts finds which EVE account (core_user_<userID>.dat) each character
// (core_char_<charID>.dat) of a profile belongs to, and plans an account-settings copy.
// ESI does not expose the character -> account link, so it comes only from CERTAIN evidence:
// the EVE launcher logs, a guided identification (see Tracker), the account file naming a
// character, or a client login backup. It only reads; learned links live in the app config.
package accounts

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/WilliamFalci/capsuleers.settings/internal/marshal"
)

// Evidence sources. Resolve ranks them (see tier).
const (
	SourceLauncher    = "launcher"     // /LauncherData= in the launcher log: userID:charID, certain
	SourceGuided      = "guided"       // the user's own identification (or a legacy confirmation)
	SourceLauncherESI = "launcher-esi" // strictly paired [esi] fetch lines
	SourcePortrait    = "portrait"     // core_user lists the character in ui/pendingPortraitGeneration
	SourceRift        = "rift"         // the client backed up one core_user with 1-3 core_char in the same instant
)

const (
	riftGap    = time.Second // rift backups of one login are written within a few ms
	maxPerAcct = 3           // EVE accounts hold at most 3 characters
	maxDecode  = 8 << 20
)

var fileRe = regexp.MustCompile(`^core_(char|user)_(\d+)(_rift_backup_\d+)?\.dat$`)

// Suggestion links one character to one account from the profile files alone.
type Suggestion struct {
	CharID string `json:"charId"`
	UserID string `json:"userId"`
	Reason string `json:"reason"`
}

// Stamp is one settings file as seen in the profile directory.
type Stamp struct {
	Name    string
	ModTime time.Time
}

// Suggest reads dir (never writes) and returns at most one deterministic suggestion per live character.
func Suggest(dir string) map[string]Suggestion {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return map[string]Suggestion{}
	}
	var stamps []Stamp
	refs := map[string][]string{}
	for _, e := range entries {
		m := fileRe.FindStringSubmatch(e.Name())
		if m == nil || !e.Type().IsRegular() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		stamps = append(stamps, Stamp{Name: e.Name(), ModTime: info.ModTime()})
		if m[1] == "user" && m[3] == "" && info.Size() <= maxDecode {
			if b, err := os.ReadFile(filepath.Join(dir, e.Name())); err == nil {
				refs[m[2]] = append(refs[m[2]], PortraitRefs(b)...)
			}
		}
	}
	return suggest(stamps, refs)
}

type file struct {
	kind, id string
	rift     bool
	t        time.Time
}

// suggest is the pure part of Suggest. refs maps a user id to the char ids its content names.
// Two accounts for one character = no suggestion.
func suggest(stamps []Stamp, refs map[string][]string) map[string]Suggestion {
	var files []file
	liveChar, liveUser := map[string]bool{}, map[string]bool{}
	for _, s := range stamps {
		m := fileRe.FindStringSubmatch(s.Name)
		if m == nil {
			continue
		}
		f := file{kind: m[1], id: m[2], rift: m[3] != "", t: s.ModTime}
		files = append(files, f)
		if !f.rift {
			if f.kind == "char" {
				liveChar[f.id] = true
			} else {
				liveUser[f.id] = true
			}
		}
	}
	type ev struct{ user, reason string }
	found := map[string][]ev{}
	add := func(char, user, reason string) {
		if liveChar[char] && liveUser[user] {
			found[char] = append(found[char], ev{user, reason})
		}
	}

	// 1. content: the account file names a character waiting for a portrait (created on it)
	for user, chars := range refs {
		for _, c := range chars {
			add(c, user, SourcePortrait)
		}
	}

	// 2. rift backups: one login backs up its core_user and the account's characters at once
	var rift []file
	for _, f := range files {
		if f.rift {
			rift = append(rift, f)
		}
	}
	sort.Slice(rift, func(i, j int) bool { return rift[i].t.Before(rift[j].t) })
	for i := 0; i < len(rift); {
		j := i + 1
		for j < len(rift) && rift[j].t.Sub(rift[j-1].t) <= riftGap {
			j++
		}
		users, chars := map[string]bool{}, map[string]bool{}
		for _, f := range rift[i:j] {
			if f.kind == "user" {
				users[f.id] = true
			} else {
				chars[f.id] = true
			}
		}
		if len(users) == 1 && len(chars) >= 1 && len(chars) <= maxPerAcct {
			for u := range users {
				for c := range chars {
					add(c, u, SourceRift)
				}
			}
		}
		i = j
	}

	out := map[string]Suggestion{}
	for c, list := range found {
		pick, conflict := list[0], false
		for _, e := range list[1:] {
			if e.user != pick.user {
				conflict = true
			} else if e.reason == SourcePortrait {
				pick = e
			}
		}
		if !conflict {
			out[c] = Suggestion{CharID: c, UserID: pick.user, Reason: pick.reason}
		}
	}
	return out
}

// PortraitRefs returns the character ids found under ui/pendingPortraitGeneration of a
// core_user file (the client queues portraits of characters created on that account).
// Undecodable input yields nil.
func PortraitRefs(data []byte) []string {
	v, err := marshal.Decode(data)
	if err != nil {
		return nil
	}
	root, _ := v.(*marshal.Dict)
	ui, _ := root.Get("ui")
	uid, _ := ui.(*marshal.Dict)
	p, ok := uid.Get("pendingPortraitGeneration")
	if !ok {
		return nil
	}
	var out []string
	seen := map[any]bool{}
	var walk func(v any, depth int)
	walk = func(v any, depth int) {
		if depth > 16 {
			return
		}
		switch x := v.(type) {
		case int64:
			// character ids: 90M+ (old) and 2.1G+ (new); FILETIME stamps are far larger
			if x >= 90_000_000 && x < 3_000_000_000 {
				out = append(out, itoa(x))
			}
		case *marshal.Tuple:
			if !seen[x] {
				seen[x] = true
				for _, it := range x.Items {
					walk(it, depth+1)
				}
			}
		case *marshal.List:
			if !seen[x] {
				seen[x] = true
				for _, it := range x.Items {
					walk(it, depth+1)
				}
			}
		case *marshal.Dict:
			if !seen[x] {
				seen[x] = true
				for i := range x.Keys {
					walk(x.Keys[i], depth+1)
					walk(x.Vals[i], depth+1)
				}
			}
		case *marshal.Object:
			if !seen[x] {
				seen[x] = true
				walk(x.Head, depth+1)
				walk(x.State, depth+1)
				for _, it := range x.Items {
					walk(it, depth+1)
				}
				for i := range x.Keys {
					walk(x.Keys[i], depth+1)
					walk(x.Vals[i], depth+1)
				}
			}
		}
	}
	walk(p, 0)
	return out
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// UserWrite is one account file the copy will overwrite.
type UserWrite struct {
	UserID  string   `json:"userId"`
	Targets []string `json:"targets"` // selected targets on this account
	Chars   []string `json:"chars"`   // every known character on this account (all share the file)
}

// Plan is what copying the source's account settings would do.
type Plan struct {
	SourceUser string      `json:"sourceUser"` // "" = unknown: nothing can be copied
	Writes     []UserWrite `json:"writes"`     // one per distinct target account, sorted
	Same       []string    `json:"same"`       // targets on the source account: no write needed
	Unknown    []string    `json:"unknown"`    // targets with no known account: the copy is refused
}

// MakePlan dedups targets by account. accountOf maps char id -> user id ("" or absent = unknown).
func MakePlan(sourceID string, targetIDs []string, accountOf map[string]string) Plan {
	p := Plan{SourceUser: accountOf[sourceID], Writes: []UserWrite{}, Same: []string{}, Unknown: []string{}}
	idx := map[string]int{}
	seen := map[string]bool{sourceID: true}
	for _, t := range targetIDs {
		if seen[t] {
			continue
		}
		seen[t] = true
		u := accountOf[t]
		switch {
		case u == "":
			p.Unknown = append(p.Unknown, t)
		case u == p.SourceUser:
			p.Same = append(p.Same, t)
		default:
			i, ok := idx[u]
			if !ok {
				i = len(p.Writes)
				idx[u] = i
				p.Writes = append(p.Writes, UserWrite{UserID: u, Targets: []string{}, Chars: []string{}})
			}
			p.Writes[i].Targets = append(p.Writes[i].Targets, t)
		}
	}
	for c, u := range accountOf {
		if i, ok := idx[u]; ok && u != "" {
			p.Writes[i].Chars = append(p.Writes[i].Chars, c)
		}
	}
	for i := range p.Writes {
		sort.Strings(p.Writes[i].Chars)
	}
	sort.Slice(p.Writes, func(i, j int) bool { return p.Writes[i].UserID < p.Writes[j].UserID })
	return p
}

// Ready reports whether the plan can be applied: source account known, no unknown targets.
func (p Plan) Ready() bool { return p.SourceUser != "" && len(p.Unknown) == 0 }

// Targets returns the user ids to overwrite.
func (p Plan) Targets() []string {
	out := make([]string, 0, len(p.Writes))
	for _, w := range p.Writes {
		out = append(out, w.UserID)
	}
	return out
}
