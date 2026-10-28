package config

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAccountLinks(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	prof := filepath.Join(dir, "settings_Default")
	at := time.Date(2026, 6, 24, 21, 0, 0, 0, time.UTC)
	if err := s.SetAccountGuided(prof+"/", "9000000001", Link{"90000003", at, "guided"}); err != nil { // path cleaned
		t.Fatal(err)
	}
	if err := s.SetAccountLearned(prof, map[string]Link{"90000004": {"9000005", at, "launcher"}}); err != nil {
		t.Fatal(err)
	}
	r := NewStore(dir)
	if g := r.AccountGuided(prof)["9000000001"]; g.UserID != "90000003" || !g.At.Equal(at) {
		t.Fatalf("guided %+v", g)
	}
	if l := r.AccountLearned(prof)["90000004"]; l.UserID != "9000005" || l.Source != "launcher" {
		t.Fatalf("learned %+v", l)
	}
	for _, bad := range [][3]string{{"rel/settings_x", "1", "2"}, {prof, "../1", "2"}, {prof, "1", "x2"}} {
		if s.SetAccountGuided(bad[0], bad[1], Link{bad[2], at, "guided"}) == nil {
			t.Errorf("accepted %v", bad)
		}
	}
	if s.SetAccountLearned(prof, map[string]Link{"1": {"2", at, ""}}) == nil {
		t.Error("accepted a link with no source")
	}
	if err := s.SetAccountLearned(prof, nil); err != nil || len(NewStore(dir).AccountLearned(prof)) != 0 {
		t.Fatal("clearing learned links failed")
	}
}

func TestSetAccountAlias(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	prof := filepath.Join(dir, "settings_Default")
	if err := s.SetAccountAlias(prof, "9000005", "  Main  "); err != nil {
		t.Fatal(err)
	}
	if got := NewStore(dir).AccountAliases(prof)["9000005"]; got != "Main" {
		t.Fatalf("alias %q", got)
	}
	if err := s.SetAccountAlias(prof, "9000005", "   "); err != nil {
		t.Fatal(err)
	}
	if got := s.AccountAliases(prof); len(got) != 0 {
		t.Fatalf("not removed: %v", got)
	}
	long := strings.Repeat("a", MaxAliasLen+1)
	for _, bad := range [][3]string{{"rel", "1", "x"}, {prof, "x1", "x"}, {prof, "1", long}} {
		if s.SetAccountAlias(bad[0], bad[1], bad[2]) == nil {
			t.Errorf("accepted %v", bad)
		}
	}
	if s.SetAccountAlias(prof, "1", strings.Repeat("à", MaxAliasLen)) != nil {
		t.Error("40 runes rejected")
	}
}
