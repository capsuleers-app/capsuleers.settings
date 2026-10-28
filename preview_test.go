package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/WilliamFalci/capsuleers.settings/internal/marshal"
	mt "github.com/WilliamFalci/capsuleers.settings/internal/marshal/marshaltest"
)

func fakeChar(t *testing.T, dir, id string, overview any) {
	t.Helper()
	wrap := func(v any) any { return mt.T(mt.Long(10), v) }
	flags := func(ids ...string) *marshal.Dict {
		d := &marshal.Dict{}
		for _, i := range ids {
			d.Keys, d.Vals = append(d.Keys, i), append(d.Vals, true)
		}
		return d
	}
	root := mt.D("windows", mt.D(
		"windowSizesAndPositions_1", wrap(mt.D("overview", overview, "fitting", mt.T(10, 10, 200, 200, 1920, 1080))),
		"openWindows", wrap(flags("overview", "fitting")),
	), "ui", mt.D())
	if err := os.WriteFile(filepath.Join(dir, "core_char_"+id+".dat"), mt.Encode(root, mt.Opts{ShareKeys: true}), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestPreviewCompares(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "settings_Default")
	_ = os.MkdirAll(dir, 0o700)
	fakeChar(t, dir, "1", mt.T(100, 100, 300, 300, 1920, 1080))
	fakeChar(t, dir, "2", mt.T(500, 100, 300, 300, 1920, 1080))
	_ = os.WriteFile(filepath.Join(dir, "core_char_3.dat"), []byte("not a settings file"), 0o600)
	s := &Service{}

	r := s.Preview(dir, "1", "2", "it")
	if !r.OK || r.Target == nil || !r.Source.Readable || !r.Target.Readable || len(r.Items) != 2 {
		t.Fatalf("%+v", r)
	}
	st := map[string]string{}
	for _, it := range r.Items {
		st[it.ID] = it.Status
	}
	if st["overview"] != "move" || st["fitting"] != "same" || r.Items[0].Label == "" {
		t.Fatalf("statuses %v", st)
	}

	if r := s.Preview(dir, "1", "", "en"); !r.OK || r.Target != nil || r.Items[0].Status != "" {
		t.Fatalf("no target: %+v", r)
	}
	if r := s.Preview(dir, "1", "3", "en"); r.Target == nil || r.Target.Readable || r.Target.Error != "unreadable" || len(r.Items) != 2 || r.Items[0].Status != "" {
		t.Fatalf("unreadable target: %+v", r.Target)
	}
	if r := s.Preview(dir, "3", "1", "en"); r.Source.Readable || r.Source.Error != "unreadable" {
		t.Fatalf("unreadable source: %+v", r.Source)
	}
	if r := s.Preview(dir, "99", "", "en"); r.Source.Error != "missing" {
		t.Fatalf("missing: %+v", r.Source)
	}
	for _, bad := range [][3]string{{"relative/settings_x", "1", ""}, {filepath.Join(dir, "..", "other"), "1", ""}, {dir, "../1", ""}, {dir, "1", "x"}} {
		if r := s.Preview(bad[0], bad[1], bad[2], "en"); r.Code != "bad_request" || r.OK {
			t.Fatalf("%v accepted", bad)
		}
	}
}
