package backup

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setup(t *testing.T) (*Manager, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "settings_Default")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for id, c := range map[string]string{"1": "SOURCE", "2": "two", "3": "three", "4": "four"} {
		if err := os.WriteFile(filepath.Join(dir, "core_char_"+id+".dat"), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	clock := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	m := &Manager{Dir: filepath.Join(t.TempDir(), "backups"), Running: func() int { return 0 },
		NameOf: func(id string) string { return "char" + id }, Now: func() time.Time { clock = clock.Add(time.Second); return clock }}
	return m, dir
}

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestApplyRestoreRoundtrip(t *testing.T) {
	m, dir := setup(t)
	res, err := m.Apply(ApplyRequest{ProfileDir: dir, SourceID: "1", TargetIDs: []string{"2", "3", "1"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Updated) != 2 {
		t.Fatalf("updated=%v", res.Updated)
	}
	for _, id := range []string{"2", "3"} {
		if got := read(t, filepath.Join(dir, "core_char_"+id+".dat")); got != "SOURCE" {
			t.Errorf("target %s=%q", id, got)
		}
	}
	if got := read(t, filepath.Join(dir, "core_char_4.dat")); got != "four" {
		t.Errorf("untouched file changed: %q", got)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temp file left: %s", e.Name())
		}
	}
	list := m.List()
	if len(list) != 1 || len(list[0].Files) != 2 || list[0].SourceName != "char1" || list[0].Profile != "settings_Default" {
		t.Fatalf("list=%+v", list)
	}
	rr, err := m.Restore(res.Backup)
	if err != nil {
		t.Fatal(err)
	}
	if rr.Backup == "" || len(m.List()) != 2 {
		t.Fatalf("restore must back up current files first: %+v", rr)
	}
	if read(t, filepath.Join(dir, "core_char_2.dat")) != "two" || read(t, filepath.Join(dir, "core_char_3.dat")) != "three" {
		t.Fatal("restore did not bring originals back")
	}
	if err := m.Delete(res.Backup); err != nil || len(m.List()) != 1 {
		t.Fatalf("delete: %v", err)
	}
	if m.Delete("../x.zip") == nil {
		t.Fatal("path traversal in delete accepted")
	}
}

func TestRefuseWhenRunning(t *testing.T) {
	m, dir := setup(t)
	m.Running = func() int { return 1 }
	if _, err := m.Apply(ApplyRequest{ProfileDir: dir, SourceID: "1", TargetIDs: []string{"2"}}); !errors.Is(err, ErrRunning) {
		t.Fatalf("apply err=%v", err)
	}
	if read(t, filepath.Join(dir, "core_char_2.dat")) != "two" || len(m.List()) != 0 {
		t.Fatal("running client must leave files and backups untouched")
	}
	m.Running = func() int { return 0 }
	res, _ := m.Apply(ApplyRequest{ProfileDir: dir, SourceID: "1", TargetIDs: []string{"2"}})
	m.Running = func() int { return 2 }
	if _, err := m.Restore(res.Backup); !errors.Is(err, ErrRunning) {
		t.Fatalf("restore err=%v", err)
	}
}

func TestRejectsBadRequests(t *testing.T) {
	m, dir := setup(t)
	bad := []ApplyRequest{
		{ProfileDir: filepath.Dir(dir), SourceID: "1", TargetIDs: []string{"2"}},   // not a settings_ dir
		{ProfileDir: dir, SourceID: "../1", TargetIDs: []string{"2"}},              // traversal
		{ProfileDir: dir, SourceID: "1", TargetIDs: []string{"../../etc/passwd"}},  // traversal
		{ProfileDir: dir, SourceID: "1", TargetIDs: []string{"99"}},                // missing target
		{ProfileDir: dir, SourceID: "1", TargetIDs: []string{"1"}},                 // only the source
		{ProfileDir: dir, SourceID: "1", TargetIDs: nil},
	}
	for i, r := range bad {
		if _, err := m.Apply(r); err == nil {
			t.Errorf("case %d accepted", i)
		}
	}
}

func TestRestoreRejectsTamperedManifest(t *testing.T) {
	m, dir := setup(t)
	res, _ := m.Apply(ApplyRequest{ProfileDir: dir, SourceID: "1", TargetIDs: []string{"2"}})
	// rewrite the zip with a manifest pointing outside the profile dir
	p := filepath.Join(m.Dir, res.Backup)
	out := filepath.Join(m.Dir, "2030-01-01_000000.zip")
	f, _ := os.Create(out)
	zw := zip.NewWriter(f)
	w, _ := zw.Create("files/001_core_char_2.dat")
	w.Write([]byte("evil"))
	w, _ = zw.Create("manifest.json")
	w.Write([]byte(`{"version":1,"profileDir":"` + filepath.ToSlash(dir) + `","files":[{"zip":"files/001_core_char_2.dat","original":"` + filepath.ToSlash(filepath.Join(filepath.Dir(dir), "core_char_2.dat")) + `"}]}`))
	zw.Close()
	f.Close()
	_ = p
	if _, err := m.Restore("2030-01-01_000000.zip"); !errors.Is(err, ErrBadBackup) {
		t.Fatalf("err=%v", err)
	}
}

func TestKeepsLast20(t *testing.T) {
	m, dir := setup(t)
	for i := 0; i < KeepLast+3; i++ {
		if _, err := m.Apply(ApplyRequest{ProfileDir: dir, SourceID: "1", TargetIDs: []string{"2"}}); err != nil {
			t.Fatal(err)
		}
	}
	if n := len(m.List()); n != KeepLast {
		t.Fatalf("kept %d", n)
	}
}
