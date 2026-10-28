package accounts

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func ld(v string) string { return base64.StdEncoding.EncodeToString([]byte(v)) }

// Synthetic log: ids are made up; the SSO token stands in for the secrets real lines carry.
func fixture() string {
	return strings.Join([]string{
		"2026-06-24T21:00:00.100Z [info] unrelated line",
		"2026-06-24T21:00:01.000Z [info] start exefile.exe /server:Tranquility /ssoToken=SECRET-TOKEN-1 /LauncherData=" + ld("eve-online:tranquility::7001:9001") + " /refresh=SECRET-2",
		"2026-06-24T21:00:02.000Z [info] bad /LauncherData=" + ld("garbage"),
		"2026-06-24T21:00:03.000Z [info] bad /LauncherData=!!!notbase64!!!",
		"2026-06-24T21:10:00.000Z [esi] Fetching details for 2 character(s) {characterIds:[9101,9102]}",
		"2026-06-24T21:10:01.000Z [esi] Fetched 2 character details {userId:7002, names:[a,b]}",
		// two open fetches of the same size, one answer: ambiguous, dropped
		"2026-06-24T21:20:00.000Z [esi] Fetching details for 1 character(s) {characterIds:[9201]}",
		"2026-06-24T21:20:00.500Z [esi] Fetching details for 1 character(s) {characterIds:[9202]}",
		"2026-06-24T21:20:01.000Z [esi] Fetched 1 character details {userId:7003}",
		// id count differing from N: malformed, dropped
		"2026-06-24T21:30:00.000Z [esi] Fetching details for 3 character(s) {characterIds:[9301,9302]}",
		"2026-06-24T21:30:01.000Z [esi] Fetched 3 character details {userId:7004}",
		"2026-06-24T21:40:00.000Z [esi] Fetched 1 character details {userId:7005}", // answer with no request
	}, "\n")
}

func TestParseLauncherLog(t *testing.T) {
	fb := at(0)
	got := ParseLauncherLog(strings.NewReader(fixture()), fb)
	if len(got) != 3 {
		t.Fatalf("got %+v", got)
	}
	if e := got[0]; e.Source != SourceLauncher || e.CharID != "9001" || e.UserID != "7001" || !e.At.Equal(time.Date(2026, 6, 24, 21, 0, 1, 0, time.UTC)) {
		t.Fatalf("A: %+v", e)
	}
	for i, c := range []string{"9101", "9102"} {
		if e := got[1+i]; e.Source != SourceLauncherESI || e.CharID != c || e.UserID != "7002" {
			t.Fatalf("B %d: %+v", i, e)
		}
	}
	// never keeps anything but ids
	for _, e := range got {
		if strings.Contains(e.CharID+e.UserID+e.Source, "SECRET") {
			t.Fatal("token leaked")
		}
	}
}

func TestParseLauncherLogDefensive(t *testing.T) {
	long := "x" + strings.Repeat("y", 100000) + " /LauncherData=" + ld("eve-online:tranquility::1:2")
	cases := []string{"", "\x00\xff\xfe", "/LauncherData=", "[esi] Fetching details for 99999 character(s)", long,
		"[esi] Fetched 1 character details {userId:1} {userId:2}"}
	for _, c := range cases {
		_ = ParseLauncherLog(strings.NewReader(c), at(0)) // must not panic
	}
	// no timestamp: falls back to the file time; url-safe unpadded base64 is accepted
	raw := base64.RawURLEncoding.EncodeToString([]byte("eve-online:tranquility::31:41"))
	got := ParseLauncherLog(strings.NewReader("start /LauncherData="+raw), at(time.Hour))
	if len(got) != 1 || got[0].UserID != "31" || got[0].CharID != "41" || !got[0].At.Equal(at(time.Hour)) {
		t.Fatalf("got %+v", got)
	}
}

func TestLogsDir(t *testing.T) {
	p := filepath.Join("/pfx/drive_c/users/steamuser/AppData/Local/CCP/EVE/c_ccp_eve_online_tq_tranquility/settings_Default")
	want := filepath.Join("/pfx/drive_c/users/steamuser/AppData/Roaming/EVE Online/logs")
	if got := LogsDir(p); got != want {
		t.Fatalf("got %q", got)
	}
	if LogsDir("/home/u/Library/Application Support/CCP/EVE/x/settings_Default") != "" {
		t.Fatal("no AppData: unknown")
	}
}

func TestScanLauncherLogsDropsMultiAccountESI(t *testing.T) {
	root := t.TempDir()
	prof := filepath.Join(root, "AppData", "Local", "CCP", "EVE", "srv", "settings_Default")
	logs := filepath.Join(root, "AppData", "Roaming", "EVE Online", "logs")
	for _, d := range []string{prof, logs} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	one := "[esi] Fetching details for 1 character(s) {characterIds:[9500]}\n[esi] Fetched 1 character details {userId:7100}\n"
	two := "[esi] Fetching details for 1 character(s) {characterIds:[9500]}\n[esi] Fetched 1 character details {userId:7200}\n" +
		"[esi] Fetching details for 1 character(s) {characterIds:[9600]}\n[esi] Fetched 1 character details {userId:7200}\n"
	for n, c := range map[string]string{"eve-online-launcher-1.log": one, "eve-online-launcher-2.log": two, "other.log": one} {
		if err := os.WriteFile(filepath.Join(logs, n), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got := ScanLauncherLogs(prof)
	if len(got) != 1 || got[0].CharID != "9600" || got[0].UserID != "7200" {
		t.Fatalf("char on two accounts must be dropped: %+v", got)
	}
	if ScanLauncherLogs(filepath.Join(root, "nowhere")) != nil {
		t.Fatal("no logs dir")
	}
}
