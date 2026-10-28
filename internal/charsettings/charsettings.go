// Package charsettings extracts a read-only summary (window layout, chat channels, key list)
// from a core_char_<id>.dat file, and compares two of them. It never writes: the app copies
// whole files, this is only for the preview.
package charsettings

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/capsuleers-app/capsuleers.settings/internal/marshal"
)

// maxFile bounds what Load reads (real files are ~25 KB).
const maxFile = 8 << 20

// ErrUnreadable means the file exists but is not a decodable settings file.
var ErrUnreadable = errors.New("unreadable settings file")

// Window is one top-level window or window stack as saved by the client.
type Window struct {
	ID        string   `json:"id"`
	Left      int      `json:"left"`
	Top       int      `json:"top"`
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	ScreenW   int      `json:"screenW"` // screen size when the window was saved
	ScreenH   int      `json:"screenH"`
	Members   []string `json:"members"` // ids stacked in this window (tabs), in tab order
	Locked    bool     `json:"locked"`
	Collapsed bool     `json:"collapsed"`
	Compact   bool     `json:"compact"`
	Minimized bool     `json:"minimized"`
}

// Channel is a chat channel the character has configured.
type Channel struct {
	ID   string `json:"id"`   // chatchannel_<...>
	Kind string `json:"kind"` // local, corp, alliance, fleet, private, player, other
	Name string `json:"name"` // as saved by the client, may be empty
}

// Key is one entry of a top-level settings group.
type Key struct {
	Group string `json:"group"`
	Name  string `json:"name"`
	Type  string `json:"type"` // flag, number, text, list, map, tuple, empty, object
	Len   int    `json:"len"`  // items for list/map/tuple, characters for text
}

// Settings is the extracted summary of one file.
type Settings struct {
	Windows  []Window // visible top-level windows, sorted by id
	ScreenW  int      // most common screen size among the windows (0 when none)
	ScreenH  int
	Suffix   int       // n of the windowSizesAndPositions_<n> entry used (-1 when none)
	Channels []Channel // sorted by id
	Keys     []Key     // sorted by group, name
}

// Load reads and parses a settings file.
func Load(path string) (*Settings, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() || st.Size() > maxFile {
		return nil, ErrUnreadable
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

// Parse decodes the marshal stream and extracts the summary. Any decoding problem is ErrUnreadable.
func Parse(data []byte) (*Settings, error) {
	v, err := marshal.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnreadable, err)
	}
	root, ok := v.(*marshal.Dict)
	if !ok {
		return nil, fmt.Errorf("%w: top level is not a dict", ErrUnreadable)
	}
	s := &Settings{Suffix: -1}
	x := extractor{root: root}
	s.Keys = x.keys()
	s.Windows, s.Suffix = x.windows()
	s.ScreenW, s.ScreenH = commonScreen(s.Windows)
	s.Channels = x.channels()
	return s, nil
}

// entry is a stored setting: the client wraps every value as (timestamp, value).
type entry struct {
	ts  int64
	val any
}

type extractor struct{ root *marshal.Dict }

func (x *extractor) group(name string) *marshal.Dict {
	g, _ := x.root.Get(name)
	d, _ := g.(*marshal.Dict)
	return d
}

func unwrap(v any) (entry, bool) {
	t, ok := v.(*marshal.Tuple)
	if !ok || len(t.Items) != 2 {
		return entry{}, false
	}
	ts, _ := toInt(t.Items[0])
	return entry{ts: int64(ts), val: t.Items[1]}, true
}

func (x *extractor) setting(group, key string) (entry, bool) {
	g := x.group(group)
	if g == nil {
		return entry{}, false
	}
	v, ok := g.Get(key)
	if !ok {
		return entry{}, false
	}
	return unwrap(v)
}

func (x *extractor) dict(group, key string) *marshal.Dict {
	e, ok := x.setting(group, key)
	if !ok {
		return nil
	}
	d, _ := e.val.(*marshal.Dict)
	return d
}

func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int64:
		if n > 1<<40 || n < -(1<<40) {
			return 0, false
		}
		return int(n), true
	case float64:
		if n != n || n > 1e9 || n < -1e9 {
			return 0, false
		}
		return int(n), true
	}
	return 0, false
}

func flags(d *marshal.Dict) map[string]bool {
	m := map[string]bool{}
	if d == nil {
		return m
	}
	for i, k := range d.Keys {
		if s, ok := k.(string); ok {
			if b, ok := d.Vals[i].(bool); ok {
				m[s] = b
			}
		}
	}
	return m
}

const posPrefix = "windowSizesAndPositions_"

// windows builds the list of visible top-level windows. The saved position set is the
// windowSizesAndPositions_<n> entry with the newest timestamp (every file seen so far only has _1).
func (x *extractor) windows() ([]Window, int) {
	g := x.group("windows")
	if g == nil {
		return nil, -1
	}
	bestN, bestTS := -1, int64(-1)
	var pos *marshal.Dict
	for i, k := range g.Keys {
		name, ok := k.(string)
		if !ok || !strings.HasPrefix(name, posPrefix) {
			continue
		}
		n, err := strconv.Atoi(name[len(posPrefix):])
		if err != nil {
			continue
		}
		e, ok := unwrap(g.Vals[i])
		if !ok {
			continue
		}
		d, ok := e.val.(*marshal.Dict)
		if !ok {
			continue
		}
		if e.ts > bestTS || (e.ts == bestTS && n < bestN) {
			bestN, bestTS, pos = n, e.ts, d
		}
	}
	if pos == nil {
		return nil, -1
	}
	openMap := x.dict("windows", "openWindows")
	open := flags(openMap)
	locked, collapsed := flags(x.dict("windows", "lockedWindows")), flags(x.dict("windows", "collapsedWindows"))
	compact, minimized := flags(x.dict("windows", "compactWindows")), flags(x.dict("windows", "minimizedWindows"))

	parent := map[string]string{}
	if sd := x.dict("windows", "stacksWindows"); sd != nil {
		for i, k := range sd.Keys {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			switch p := sd.Vals[i].(type) {
			case string:
				if p != "" && p != ks {
					parent[ks] = p
				}
			case int64:
				parent[ks] = strconv.FormatInt(p, 10)
			}
		}
	}
	order := x.dict("windows", "preferredIdxInStack3")

	isOpen := func(id string) bool {
		if openMap == nil {
			return true
		}
		return open[id]
	}

	// Positions of every window, then group visible leaves under their root stack.
	type rect struct{ l, t, w, h, sw, sh int }
	rects := map[string]rect{}
	for i, k := range pos.Keys {
		id, ok := k.(string)
		if !ok || len(id) > 200 {
			continue
		}
		t, ok := pos.Vals[i].(*marshal.Tuple)
		if !ok || len(t.Items) < 6 {
			continue
		}
		var n [6]int
		good := true
		for j := 0; j < 6; j++ {
			if n[j], ok = toInt(t.Items[j]); !ok {
				good = false
			}
		}
		if !good || n[2] <= 0 || n[3] <= 0 || n[4] <= 0 || n[5] <= 0 || n[2] > 100000 || n[3] > 100000 {
			continue
		}
		rects[id] = rect{n[0], n[1], n[2], n[3], n[4], n[5]}
	}

	root := func(id string) string { // follow stack parents, bounded to survive cycles
		for i := 0; i < 8; i++ {
			p, ok := parent[id]
			if !ok {
				return id
			}
			id = p
		}
		return id
	}
	members := map[string][]string{}
	for id := range rects {
		if isTransient(id) || !isOpen(id) {
			continue
		}
		if r := root(id); r != id {
			members[r] = append(members[r], id)
		}
	}
	isStack := map[string]bool{}
	for _, p := range parent {
		isStack[root(p)] = true
	}

	var out []Window
	for id, r := range rects {
		if isTransient(id) || !isOpen(id) {
			continue
		}
		if _, nested := parent[id]; nested {
			continue
		}
		w := Window{ID: id, Left: r.l, Top: r.t, Width: r.w, Height: r.h, ScreenW: r.sw, ScreenH: r.sh,
			Locked: locked[id], Collapsed: collapsed[id], Compact: compact[id], Minimized: minimized[id]}
		w.Members = orderMembers(members[id], order, id)
		if (isStack[id] || numericRe.MatchString(id)) && len(w.Members) == 0 {
			continue // an empty stack: its tabs are closed or transient (numeric ids are stack containers)
		}
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, bestN
}

func orderMembers(ids []string, order *marshal.Dict, stack string) []string {
	idx := map[string]int{}
	if order != nil {
		if v, ok := order.Get(stack); ok {
			if d, ok := v.(*marshal.Dict); ok {
				for i, k := range d.Keys {
					if s, ok := k.(string); ok {
						if n, ok := toInt(d.Vals[i]); ok {
							idx[s] = n
						}
					}
				}
			}
		}
	}
	out := append([]string(nil), ids...)
	sort.Slice(out, func(i, j int) bool {
		a, aok := idx[out[i]]
		b, bok := idx[out[j]]
		switch {
		case aok && bok && a != b:
			return a < b
		case aok != bok:
			return aok
		}
		return out[i] < out[j]
	})
	return out
}

var transientRe = regexp.MustCompile(`(?i)(^ChatInvitation_|dialog(ue)?$|popup$|messagebox$)`)

// Short-lived dialogs the client also remembers; they are not part of a layout worth previewing.
var transientIDs = map[string]bool{
	"message": true, "setNewName": true, "NewMessageWindow": true, "CrateRedeemedWindow": true,
	"GammaSliderWindow": true, "NewFeatureNotifyWnd": true, "CtrlTabWindow": true, "EngineTools": true,
}

func isTransient(id string) bool { return transientIDs[id] || transientRe.MatchString(id) }

// commonScreen returns the screen size most windows were saved at (ties: the larger one).
func commonScreen(ws []Window) (int, int) {
	type k struct{ w, h int }
	cnt := map[k]int{}
	for _, w := range ws {
		cnt[k{w.ScreenW, w.ScreenH}]++
	}
	var best k
	bc := 0
	for s, c := range cnt {
		if c > bc || (c == bc && s.w*s.h > best.w*best.h) {
			best, bc = s, c
		}
	}
	return best.w, best.h
}

var channelRe = regexp.MustCompile(`^chatchannel_(.+)$`)

func channelKind(id string) string {
	rest := strings.TrimPrefix(id, "chatchannel_")
	switch {
	case rest == "local" || rest == "corp" || rest == "alliance" || rest == "fleet":
		return rest
	case strings.HasPrefix(rest, "private_"):
		return "private"
	case strings.HasPrefix(rest, "player_"):
		return "player"
	}
	return "other"
}

// channels merges the client's saved channel list (ui/chatchannels) with the chat windows that are open.
func (x *extractor) channels() []Channel {
	byID := map[string]*Channel{}
	add := func(id, name string) {
		if len(id) > 200 || !channelRe.MatchString(id) {
			return
		}
		c := byID[id]
		if c == nil {
			c = &Channel{ID: id, Kind: channelKind(id)}
			byID[id] = c
		}
		if name != "" && c.Name == "" {
			c.Name = name
		}
	}
	if e, ok := x.setting("ui", "chatchannels"); ok {
		if l, ok := e.val.(*marshal.List); ok {
			for _, it := range l.Items {
				t, ok := it.(*marshal.Tuple)
				if !ok || len(t.Items) < 3 {
					continue
				}
				suffix, ok1 := t.Items[0].(string)
				name, _ := t.Items[2].(string)
				if ok1 {
					add("chatchannel_"+suffix, name)
				}
			}
		}
	}
	if op := x.dict("windows", "openWindows"); op != nil {
		for i, k := range op.Keys {
			if s, ok := k.(string); ok && strings.HasPrefix(s, "chatchannel_") && op.Vals[i] == true {
				add(s, "")
			}
		}
	}
	out := make([]Channel, 0, len(byID))
	for _, c := range byID {
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

const maxKeys = 5000

func (x *extractor) keys() []Key {
	var out []Key
	for gi, gk := range x.root.Keys {
		gname, ok := gk.(string)
		if !ok {
			continue
		}
		gd, ok := x.root.Vals[gi].(*marshal.Dict)
		if !ok {
			out = append(out, describe("", gname, x.root.Vals[gi]))
			continue
		}
		for i, k := range gd.Keys {
			if len(out) >= maxKeys {
				break
			}
			name := fmt.Sprint(k)
			if len(name) > 200 {
				name = name[:200]
			}
			v := gd.Vals[i]
			if e, ok := unwrap(v); ok {
				v = e.val
			}
			out = append(out, describe(gname, name, v))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Group != out[j].Group {
			return out[i].Group < out[j].Group
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func describe(group, name string, v any) Key {
	k := Key{Group: group, Name: name}
	switch t := v.(type) {
	case nil:
		k.Type = "empty"
	case bool:
		k.Type = "flag"
	case int64, float64:
		k.Type = "number"
	case string:
		k.Type, k.Len = "text", len(t)
	case *marshal.List:
		k.Type, k.Len = "list", len(t.Items)
	case *marshal.Tuple:
		k.Type, k.Len = "tuple", len(t.Items)
	case *marshal.Dict:
		k.Type, k.Len = "map", len(t.Keys)
	default:
		k.Type = "object"
	}
	return k
}
