package charsettings_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	cs "github.com/capsuleers-app/capsuleers.settings/internal/charsettings"
	"github.com/capsuleers-app/capsuleers.settings/internal/marshal"
	mt "github.com/capsuleers-app/capsuleers.settings/internal/marshal/marshaltest"
)

// wrap stores a value the way the client does: (timestamp, value).
func wrap(ts int64, v any) any { return mt.T(mt.Long(ts), v) }

func win(l, t, w, h, sw, sh int) any { return mt.T(l, t, w, h, sw, sh) }

type layout struct {
	pos       map[string]any // id -> tuple
	open      []string
	stacks    map[string]any // id -> parent id or nil
	locked    []string
	collapsed []string
	compact   []string
	channels  []any
	order     map[string]any
}

func (l layout) bytes(suffixes map[string]int64) []byte {
	flagsOf := func(ids []string, all []string) *marshal.Dict {
		d := &marshal.Dict{}
		for _, id := range all {
			d.Keys = append(d.Keys, id)
			d.Vals = append(d.Vals, false)
		}
		for _, id := range ids {
			for i, k := range d.Keys {
				if k == id {
					d.Vals[i] = true
				}
			}
		}
		return d
	}
	var all []string
	posD := &marshal.Dict{}
	for id, v := range l.pos {
		all = append(all, id)
		posD.Keys = append(posD.Keys, id)
		posD.Vals = append(posD.Vals, v)
	}
	stacks := &marshal.Dict{}
	for id, p := range l.stacks {
		stacks.Keys = append(stacks.Keys, id)
		stacks.Vals = append(stacks.Vals, p)
	}
	windows := mt.D(
		"openWindows", wrap(5, flagsOf(l.open, all)),
		"lockedWindows", wrap(5, flagsOf(l.locked, all)),
		"collapsedWindows", wrap(5, flagsOf(l.collapsed, all)),
		"compactWindows", wrap(5, flagsOf(l.compact, all)),
		"stacksWindows", wrap(5, stacks),
		"preferredIdxInStack3", wrap(5, orEmpty(l.order)),
	)
	if suffixes == nil {
		suffixes = map[string]int64{"1": 100}
	}
	for s, ts := range suffixes {
		windows.Keys = append(windows.Keys, "windowSizesAndPositions_"+s)
		windows.Vals = append(windows.Vals, wrap(ts, posD))
	}
	root := mt.D(
		"windows", windows,
		"ui", mt.D("chatchannels", wrap(5, mt.L(l.channels...)), "somenumber", wrap(5, int64(3)), "flag", wrap(5, true)),
		"generic", mt.D(),
	)
	return mt.Encode(root, mt.Opts{ShareKeys: true})
}

func orEmpty(m map[string]any) *marshal.Dict {
	d := &marshal.Dict{}
	for k, v := range m {
		d.Keys = append(d.Keys, k)
		d.Vals = append(d.Vals, v)
	}
	return d
}

func base() layout {
	return layout{
		pos: map[string]any{
			"overview":                  win(1480, 160, 420, 470, 1920, 1080),
			"selecteditemview":          win(1480, 40, 420, 110, 1920, 1080),
			"ChatWindowStack":           win(54, 700, 420, 360, 1920, 1080),
			"chatchannel_local":         win(54, 700, 420, 360, 1920, 1080),
			"chatchannel_corp":          win(54, 700, 420, 360, 1920, 1080),
			"chatchannel_player_abc123": win(54, 700, 420, 360, 1920, 1080),
			"probeScannerWindow":        win(484, 790, 520, 270, 2133, 1200),
			"closedthing":               win(1, 1, 100, 100, 1920, 1080),
			"ChatInvitation_42":         win(1, 1, 100, 100, 1920, 1080),
			"badtuple":                  mt.T(1, 2),
		},
		open:   []string{"overview", "selecteditemview", "ChatWindowStack", "chatchannel_local", "chatchannel_corp", "chatchannel_player_abc123", "probeScannerWindow", "ChatInvitation_42"},
		stacks: map[string]any{"chatchannel_local": "ChatWindowStack", "chatchannel_corp": "ChatWindowStack", "chatchannel_player_abc123": "ChatWindowStack", "overview": nil},
		locked: []string{"overview"}, collapsed: []string{"selecteditemview"}, compact: []string{"probeScannerWindow"},
		channels: []any{mt.T("local", "local_100", "Local"), mt.T("player_abc123", "player_abc123", "Test Channel")},
		order:    map[string]any{"ChatWindowStack": mt.D("chatchannel_corp", int64(0), "chatchannel_local", int64(1))},
	}
}

func TestParseLayout(t *testing.T) {
	s, err := cs.Parse(base().bytes(nil))
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]cs.Window{}
	for _, w := range s.Windows {
		ids[w.ID] = w
	}
	if len(ids) != 4 {
		t.Fatalf("want overview, selecteditemview, ChatWindowStack, probeScannerWindow; got %v", ids)
	}
	if w := ids["overview"]; w.Left != 1480 || w.Top != 160 || w.Width != 420 || w.Height != 470 || !w.Locked || w.ScreenW != 1920 {
		t.Fatalf("overview: %+v", w)
	}
	if !ids["selecteditemview"].Collapsed || !ids["probeScannerWindow"].Compact {
		t.Fatal("collapsed/compact flags")
	}
	m := ids["ChatWindowStack"].Members
	if len(m) != 3 || m[0] != "chatchannel_corp" || m[1] != "chatchannel_local" || m[2] != "chatchannel_player_abc123" {
		t.Fatalf("stack members/order: %v", m)
	}
	if s.ScreenW != 1920 || s.ScreenH != 1080 || s.Suffix != 1 {
		t.Fatalf("screen/suffix: %dx%d _%d", s.ScreenW, s.ScreenH, s.Suffix)
	}
}

func TestPicksNewestSuffix(t *testing.T) {
	a := base()
	s, err := cs.Parse(a.bytes(map[string]int64{"1": 100, "2": 900, "x": 5000}))
	if err != nil {
		t.Fatal(err)
	}
	if s.Suffix != 2 {
		t.Fatalf("suffix %d", s.Suffix)
	}
}

func TestChannelsAndKeys(t *testing.T) {
	s, _ := cs.Parse(base().bytes(nil))
	kinds := map[string]cs.Channel{}
	for _, c := range s.Channels {
		kinds[c.ID] = c
	}
	if len(kinds) != 3 || kinds["chatchannel_local"].Name != "Local" || kinds["chatchannel_player_abc123"].Kind != "player" || kinds["chatchannel_corp"].Kind != "corp" {
		t.Fatalf("channels: %+v", s.Channels)
	}
	var sawFlag, sawMap bool
	for _, k := range s.Keys {
		if k.Group == "ui" && k.Name == "flag" && k.Type == "flag" {
			sawFlag = true
		}
		if k.Group == "windows" && k.Name == "openWindows" && k.Type == "map" && k.Len > 0 {
			sawMap = true
		}
	}
	if !sawFlag || !sawMap {
		t.Fatal("keys not described")
	}
}

func TestLabels(t *testing.T) {
	ch := map[string]cs.Channel{"chatchannel_player_abc123": {ID: "chatchannel_player_abc123", Kind: "player", Name: "Test Channel"}}
	cases := []struct {
		id, lang, want string
		members        []string
	}{
		{"overview", "en", "Overview", nil},
		{"overview", "it", "Overview", nil},
		{"watchlistpanel", "it", "Watch List", nil},
		{"overview_3", "it", "Overview 3", nil},
		{"droneview", "en", "Drones", nil},
		{"overview_2", "en", "Overview 2", nil},
		{"chatchannel_local", "it", "Local", nil},
		{"chatchannel_private_zzz", "it", "Canale privato", nil},
		{"chatchannel_player_abc123", "en", "Test Channel", nil},
		{"chatchannel_private_zzz", "en", "Private channel", nil},
		{"65", "en", "Local · Corp", []string{"Local", "Corp"}},
		{"65", "en", "Window group 65", nil},
		{"mysteryWindow", "en", "mysteryWindow", nil},
	}
	for _, c := range cases {
		if got := cs.WindowLabel(c.id, c.lang, ch, c.members); got != c.want {
			t.Errorf("%s/%s: %q != %q", c.id, c.lang, got, c.want)
		}
	}
}

func TestUnreadable(t *testing.T) {
	for name, data := range map[string][]byte{
		"empty":   nil,
		"garbage": []byte("not a marshal stream at all"),
		"notdict": mt.Encode(mt.L(int64(1)), mt.Opts{}),
	} {
		if _, err := cs.Parse(data); !errors.Is(err, cs.ErrUnreadable) {
			t.Errorf("%s: %v", name, err)
		}
	}
	// A valid stream without the expected groups is readable but empty.
	s, err := cs.Parse(mt.Encode(mt.D("generic", mt.D()), mt.Opts{ShareKeys: true}))
	if err != nil || len(s.Windows) != 0 || s.Suffix != -1 {
		t.Fatalf("empty settings: %v %+v", err, s)
	}
}

func TestLoadAndDisplay(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "core_char_1.dat")
	if err := os.WriteFile(p, base().bytes(nil), 0o600); err != nil {
		t.Fatal(err)
	}
	if s, err := cs.Load(p); err != nil || len(s.Windows) == 0 {
		t.Fatalf("load: %v", err)
	}
	if _, err := cs.Load(filepath.Join(dir, "missing.dat")); !os.IsNotExist(err) {
		t.Fatalf("missing: %v", err)
	}
	yaml := "audio:\n  masterVolume: [1, 0.7]\ndevice:\n  FullScreenSettings:\n  - 1\n  - {adapter: 0, height: 1080, presentInterval: 1, width: 1920}\n  UIScaleFullscreen: [2, 0.9]\n  UIScaleWindowed: [3, 1.0]\n  WindowedSettings:\n  - 1\n  - {height: 900, width: 1400}\nui: {}\n"
	y := filepath.Join(dir, "core_public__.yaml")
	_ = os.WriteFile(y, []byte(yaml), 0o600)
	d := cs.LoadDisplay(y)
	if !d.Known || d.ScreenW != 1920 || d.ScreenH != 1080 || d.ScaleFullscreen != 0.9 || d.ScaleWindowed != 1.0 {
		t.Fatalf("display: %+v", d)
	}
	if cs.LoadDisplay(filepath.Join(dir, "nope.yaml")).Known {
		t.Fatal("missing yaml must be unknown")
	}
}

// TestRealCorpus parses every core_char_*.dat under $CAPSULEERS_CORPUS (private, never committed).
func TestRealCorpus(t *testing.T) {
	root := os.Getenv("CAPSULEERS_CORPUS")
	if root == "" {
		t.Skip("CAPSULEERS_CORPUS not set")
	}
	n, withWindows := 0, 0
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(p) != ".dat" {
			return nil
		}
		n++
		s, err := cs.Load(p)
		if err != nil {
			t.Errorf("%s: %v", filepath.Base(p), err)
			return nil
		}
		if len(s.Windows) > 0 {
			withWindows++
			if s.ScreenW < 640 || s.ScreenH < 480 {
				t.Errorf("%s: implausible screen %dx%d", filepath.Base(p), s.ScreenW, s.ScreenH)
			}
			for _, w := range s.Windows {
				if w.Width <= 0 || w.Height <= 0 || w.Width > 4*w.ScreenW || w.Height > 4*w.ScreenH {
					t.Errorf("%s: implausible window %s %+v", filepath.Base(p), w.ID, w)
				}
			}
		}
		return nil
	})
	t.Logf("parsed %d files, %d with a window layout", n, withWindows)
}

func TestWindowCategory(t *testing.T) {
	cases := []struct {
		id      string
		members []string
		want    string
	}{
		{"overview", nil, "combat"}, {"overview_3", nil, "combat"}, {"droneview", nil, "combat"},
		{"directionalScannerWindow", nil, "combat"}, {"directionalScannerFoo", nil, "combat"}, {"probeScannerWindow", nil, "combat"},
		{"chatchannel_local", nil, "chat"}, {"ChatWindowStack", nil, "chat"}, {"invitestack", nil, "chat"},
		{"fleetwindow", nil, "fleet"}, {"watchlistpanel", nil, "fleet"}, {"StoreFleetSetupWnd", nil, "other"},
		{"inventory", nil, "ship"}, {"ActiveShipCargo", nil, "ship"}, {"corpHangarWnd", nil, "ship"}, {"fittingWnd", nil, "ship"},
		{"marketbuyaction", nil, "market"}, {"SellItemsWindow", nil, "market"}, {"walletWindow", nil, "market"}, {"MyMarketOrders", nil, "market"},
		{"('myPlaces', 1)", nil, "navigation"}, {"primary_map_panel", nil, "navigation"}, {"starMapFoo", nil, "navigation"},
		{"mail", nil, "other"}, {"telecom", nil, "other"}, {"neocom", nil, "other"}, {"lobbyWnd", nil, "other"},
		{"SkillPlanner", nil, "other"}, {"careerPortal", nil, "other"}, {"TaskConversationWindow", nil, "other"},
		{"FriendsList", nil, "other"}, {"65", nil, "other"},
		{"65", []string{"mail", "chatchannel_corp", "overview"}, "chat"},
		{"65", []string{"mail", "telecom"}, "other"},
	}
	for _, c := range cases {
		if got := cs.WindowCategory(c.id, c.members); got != c.want {
			t.Errorf("%s %v: %q != %q", c.id, c.members, got, c.want)
		}
	}
}
