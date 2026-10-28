package accounts

import (
	"os"
	"sort"
	"time"
)

// Snapshot maps each live settings file name (core_char_<id>.dat, core_user_<id>.dat; never rift
// backups) to its modification time in nanoseconds.
type Snapshot map[string]int64

// ReadSnapshot lists dir (read-only) and records the mtimes of the live character/account files.
func ReadSnapshot(dir string) (Snapshot, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	s := Snapshot{}
	for _, e := range entries {
		m := fileRe.FindStringSubmatch(e.Name())
		if m == nil || m[3] != "" || !e.Type().IsRegular() {
			continue
		}
		if info, err := e.Info(); err == nil {
			s[e.Name()] = info.ModTime().UnixNano()
		}
	}
	return s, nil
}

// Pairing outcomes.
const (
	PairNone      = "none"      // nothing usable changed
	PairMapped    = "mapped"    // one account file changed: its characters are identified
	PairAmbiguous = "ambiguous" // several account files changed: nothing can be told
)

// Outcome is what the changes between two snapshots say.
type Outcome struct {
	Kind   string
	UserID string   // PairMapped
	Chars  []string // PairMapped: characters whose file changed, sorted
}

// Pair compares two snapshots. Exactly one core_user changed together with at least one core_char:
// every changed character belongs to that account. More than one core_user: ambiguous. Characters
// with no account change (or none changed at all): nothing.
func Pair(prev, cur Snapshot) Outcome {
	var users, chars []string
	for name, t := range cur {
		if old, ok := prev[name]; ok && old == t {
			continue
		}
		m := fileRe.FindStringSubmatch(name)
		if m == nil || m[3] != "" {
			continue
		}
		if m[1] == "user" {
			users = append(users, m[2])
		} else {
			chars = append(chars, m[2])
		}
	}
	sort.Strings(chars)
	switch {
	case len(users) > 1:
		return Outcome{Kind: PairAmbiguous}
	case len(users) == 1 && len(chars) > 0:
		return Outcome{Kind: PairMapped, UserID: users[0], Chars: chars}
	}
	return Outcome{Kind: PairNone}
}

// Tracker turns successive snapshots into outcomes: it waits until the changes settle (no new
// change for `settle`), pairs everything changed since the last baseline, then re-baselines.
type Tracker struct {
	settle     time.Duration
	base, last Snapshot
	lastChange time.Time
}

// NewTracker starts from baseline.
func NewTracker(baseline Snapshot, settle time.Duration) *Tracker {
	return &Tracker{settle: settle, base: baseline, last: baseline}
}

// Observe feeds the current snapshot; ok is true when a settled batch produced an outcome other than none.
func (t *Tracker) Observe(cur Snapshot, now time.Time) (Outcome, bool) {
	if !equalSnap(cur, t.last) {
		t.last = cur
		t.lastChange = now
		return Outcome{}, false
	}
	if equalSnap(cur, t.base) || now.Sub(t.lastChange) < t.settle {
		return Outcome{}, false
	}
	out := Pair(t.base, cur)
	t.base = cur
	return out, out.Kind != PairNone
}

func equalSnap(a, b Snapshot) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || w != v {
			return false
		}
	}
	return true
}
