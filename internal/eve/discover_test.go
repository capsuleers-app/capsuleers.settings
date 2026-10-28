package eve

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, p, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestParseIDs(t *testing.T) {
	cases := map[string]bool{
		"core_char_2117694563.dat": true, "core_char_.dat": false, "core_char_12a.dat": false,
		"core_char_12.dat.bak": false, "core_char_12_1.dat": false, "core_user_12.dat": false,
	}
	for n, want := range cases {
		if _, ok := ParseCharID(n); ok != want {
			t.Errorf("ParseCharID(%q)=%v want %v", n, ok, want)
		}
	}
	if id, ok := ParseUserID("core_user_8.dat"); !ok || id != "8" {
		t.Errorf("ParseUserID failed: %q %v", id, ok)
	}
	if _, ok := ParseUserID("core_user_x.dat"); ok {
		t.Error("non-numeric user id accepted")
	}
}

func TestDiscoverProtonAndLibrary(t *testing.T) {
	home := t.TempDir()
	steam := filepath.Join(home, ".local/share/Steam")
	lib := filepath.Join(home, "games-lib")
	write(t, filepath.Join(steam, "steamapps/libraryfolders.vdf"),
		"\"libraryfolders\"\n{\n\t\"0\"\n\t{\n\t\t\"path\"\t\t\""+steam+"\"\n\t}\n\t\"1\"\n\t{\n\t\t\"path\"\t\t\""+lib+"\"\n\t}\n}\n")
	root := filepath.Join(lib, protonSuffix)
	prof := filepath.Join(root, "c_ccp_eve_online_tq_tranquility", "settings_Default")
	write(t, filepath.Join(prof, "core_char_111.dat"), "a")
	write(t, filepath.Join(prof, "core_char_222.dat"), "bb")
	write(t, filepath.Join(prof, "core_char_abc.dat"), "x")
	write(t, filepath.Join(prof, "core_user_9.dat"), "u")
	write(t, filepath.Join(root, "c_ccp_eve_online_sisi_singularity", "settings_PvP", "core_char_5.dat"), "z")
	// a symlink to the same root must not duplicate it
	if err := os.MkdirAll(filepath.Join(home, ".steam"), 0o755); err == nil {
		_ = os.Symlink(steam, filepath.Join(home, ".steam/steam"))
	}
	extra := t.TempDir()
	write(t, filepath.Join(extra, "settings_Old", "core_char_7.dat"), "q")

	got := Discover(Env{GOOS: "linux", Home: home}, []string{extra})
	if len(got) != 2 {
		t.Fatalf("installs=%d: %+v", len(got), got)
	}
	inst := got[0]
	if inst.Kind != "Proton" || len(inst.Servers) != 2 {
		t.Fatalf("bad install %+v", inst)
	}
	var tq Server
	for _, s := range inst.Servers {
		if s.Name == "Tranquility" {
			tq = s
		}
	}
	if len(tq.Profiles) != 1 || len(tq.Profiles[0].Chars) != 2 || len(tq.Profiles[0].Users) != 1 {
		t.Fatalf("bad tranquility %+v", tq)
	}
	if !got[1].Manual || got[1].Servers[0].Profiles[0].Name != "settings_Old" {
		t.Fatalf("manual folder not found: %+v", got[1])
	}
}

func TestCountProcCmdlines(t *testing.T) {
	proc := t.TempDir()
	write(t, filepath.Join(proc, "100", "cmdline"), "Z:\\steam\\EVE\\bin64\\ExeFile.exe\x00")
	write(t, filepath.Join(proc, "101", "cmdline"), "firefox\x00")
	write(t, filepath.Join(proc, "self", "cmdline"), "exefile.exe")
	if n := CountProcCmdlines(proc, 0); n != 1 {
		t.Fatalf("count=%d", n)
	}
	if n := CountProcCmdlines(proc, 100); n != 0 {
		t.Fatalf("skip pid failed: %d", n)
	}
}
