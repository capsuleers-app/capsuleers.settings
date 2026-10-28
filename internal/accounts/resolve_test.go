package accounts

import (
	"testing"
	"time"
)

func TestResolveTiersAndNewest(t *testing.T) {
	ev := []Evidence{
		{"1", "100", SourcePortrait, time.Time{}},
		{"1", "200", SourceLauncherESI, at(time.Hour)},
		{"2", "100", SourceLauncher, at(time.Hour)},
		{"2", "300", SourceGuided, at(2 * time.Hour)}, // newer certain evidence: transfer
		{"2", "200", SourceLauncherESI, at(9 * time.Hour)},
		{"3", "100", SourceRift, time.Time{}},
		{"3", "200", SourcePortrait, time.Time{}}, // same tier, undated, two accounts: unknown
		{"4", "999", SourceLauncher, at(0)},       // account not in the profile
		{"5", "100", SourceLauncher, at(0)},       // character not in the profile
	}
	users := map[string]bool{"100": true, "200": true, "300": true}
	chars := map[string]bool{"1": true, "2": true, "3": true, "4": true}
	got := Resolve(ev, users, chars)
	if got["1"].UserID != "200" || got["2"].UserID != "300" || got["2"].Source != SourceGuided {
		t.Fatalf("got %+v", got)
	}
	if _, ok := got["3"]; ok {
		t.Fatalf("conflict must be unknown: %+v", got["3"])
	}
	if _, ok := got["4"]; ok {
		t.Fatal("unknown account file")
	}
	if _, ok := got["5"]; ok {
		t.Fatal("foreign character")
	}
	if all := Resolve(ev, nil, nil); all["4"].UserID != "999" {
		t.Fatalf("nil filters keep everything: %+v", all["4"])
	}
}
