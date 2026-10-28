package accounts

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	mt "github.com/capsuleers-app/capsuleers.settings/internal/marshal/marshaltest"
)

var t0 = time.Date(2026, 6, 24, 21, 0, 0, 0, time.UTC)

func at(d time.Duration) time.Time { return t0.Add(d) }

func st(name string, d time.Duration) Stamp { return Stamp{Name: name, ModTime: at(d)} }

func TestSuggestRiftClusters(t *testing.T) {
	ms := time.Millisecond
	stamps := []Stamp{
		// live files, saved long apart (no mtime signal)
		st("core_char_11.dat", 10*time.Hour), st("core_char_12.dat", 20*time.Hour), st("core_char_21.dat", 30*time.Hour),
		st("core_char_31.dat", 40*time.Hour), st("core_char_41.dat", 50*time.Hour),
		st("core_user_100.dat", 1*time.Hour), st("core_user_200.dat", 2*time.Hour), st("core_user_300.dat", 3*time.Hour),
		// login of account 100: its two characters backed up in the same instant
		st("core_char_11_rift_backup_1.dat", 0), st("core_char_12_rift_backup_1.dat", 1*ms), st("core_user_100_rift_backup_1.dat", 2*ms),
		// login of account 200
		st("core_char_21_rift_backup_1.dat", time.Minute), st("core_user_200_rift_backup_1.dat", time.Minute+ms),
		// a mass backup of everything: ambiguous, must add nothing
		st("core_char_31_rift_backup_1.dat", time.Hour), st("core_char_11_rift_backup_2.dat", time.Hour),
		st("core_user_100_rift_backup_2.dat", time.Hour), st("core_user_300_rift_backup_1.dat", time.Hour+ms),
		// cluster with characters but no account: nothing
		st("core_char_41_rift_backup_1.dat", 2*time.Hour+30*time.Minute),
	}
	got := suggest(stamps, nil)
	want := map[string]Suggestion{
		"11": {"11", "100", SourceRift},
		"12": {"12", "100", SourceRift},
		"21": {"21", "200", SourceRift},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestSuggestRiftNeedsLiveFiles(t *testing.T) {
	got := suggest([]Stamp{
		st("core_char_11.dat", time.Hour),
		st("core_char_11_rift_backup_1.dat", 0), st("core_user_100_rift_backup_1.dat", 0), // no live core_user_100
	}, nil)
	if len(got) != 0 {
		t.Fatalf("suggested an account with no file: %+v", got)
	}
}

func TestSuggestStrongestWinsAndConflicts(t *testing.T) {
	ms := time.Millisecond
	stamps := []Stamp{
		// rift and portrait agree on 100; a live mtime pairing is never used
		st("core_char_1.dat", time.Hour), st("core_user_200.dat", time.Hour+time.Second), st("core_user_100.dat", 0),
		st("core_char_1_rift_backup_1.dat", 0), st("core_user_100_rift_backup_1.dat", ms),
		// two rift logins name char 2 on two accounts: no suggestion
		st("core_char_2.dat", 9*time.Hour), st("core_user_300.dat", 0),
		st("core_char_2_rift_backup_1.dat", 2*time.Hour), st("core_user_100_rift_backup_2.dat", 2*time.Hour),
		st("core_char_2_rift_backup_2.dat", 3*time.Hour), st("core_user_300_rift_backup_1.dat", 3*time.Hour),
	}
	got := suggest(stamps, map[string][]string{"100": {"1"}})
	if s := got["1"]; s.UserID != "100" || s.Reason != SourcePortrait {
		t.Fatalf("char 1: %+v", s)
	}
	if s, ok := got["2"]; ok {
		t.Fatalf("conflict must not suggest: %+v", s)
	}
}

func TestPortraitRefs(t *testing.T) {
	// shape seen in real files: ui/pendingPortraitGeneration = (FILETIME, [object((char, ...))])
	b := mt.Encode(mt.D(
		"ui", mt.D(
			"pendingPortraitGeneration", mt.T(mt.Long(134000000000000000), mt.L(mt.T(int64(2000000001), int64(5)))),
			"other", mt.T(mt.Long(1), int64(2000000002)),
		),
	), mt.Opts{ShareKeys: true})
	if got := PortraitRefs(b); !reflect.DeepEqual(got, []string{"2000000001"}) {
		t.Fatalf("got %v", got)
	}
	if PortraitRefs([]byte("junk")) != nil || PortraitRefs(mt.Encode(mt.D(), mt.Opts{})) != nil {
		t.Fatal("expected nil on junk / missing key")
	}
}

func TestSuggestReadsDir(t *testing.T) {
	dir := t.TempDir()
	write := func(name string, b []byte, mt time.Time) {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, mt, mt); err != nil {
			t.Fatal(err)
		}
	}
	user := mt.Encode(mt.D("ui", mt.D("pendingPortraitGeneration", mt.T(mt.Long(1), mt.L(int64(2000000001))))), mt.Opts{})
	write("core_char_2000000001.dat", []byte("x"), at(0))
	write("core_user_33.dat", user, at(10*time.Hour))
	write("core_char_('char', None, 'dat').dat", []byte("x"), at(0))
	got := Suggest(dir)
	if s := got["2000000001"]; s.UserID != "33" || s.Reason != SourcePortrait {
		t.Fatalf("got %+v", got)
	}
	if len(Suggest(filepath.Join(dir, "missing"))) != 0 {
		t.Fatal("missing dir")
	}
}

func TestMakePlan(t *testing.T) {
	acct := map[string]string{
		"1": "100", "2": "100", // source account
		"3": "200", "4": "200", "5": "200", // three targets, one account
		"6": "300",
		"7": "",    // explicitly unknown
		"9": "300", // not a target, but shares account 300
	}
	p := MakePlan("1", []string{"3", "4", "2", "6", "7", "8", "3", "1"}, acct)
	if p.SourceUser != "100" {
		t.Fatalf("source user %q", p.SourceUser)
	}
	want := []UserWrite{
		{UserID: "200", Targets: []string{"3", "4"}, Chars: []string{"3", "4", "5"}},
		{UserID: "300", Targets: []string{"6"}, Chars: []string{"6", "9"}},
	}
	if !reflect.DeepEqual(p.Writes, want) {
		t.Fatalf("writes %+v", p.Writes)
	}
	if !reflect.DeepEqual(p.Same, []string{"2"}) || !reflect.DeepEqual(p.Unknown, []string{"7", "8"}) {
		t.Fatalf("same=%v unknown=%v", p.Same, p.Unknown)
	}
	if p.Ready() {
		t.Fatal("unknown targets must block")
	}
	if !reflect.DeepEqual(p.Targets(), []string{"200", "300"}) {
		t.Fatalf("targets %v", p.Targets())
	}
	if q := MakePlan("1", []string{"3", "6"}, acct); !q.Ready() || len(q.Writes) != 2 {
		t.Fatalf("ready plan %+v", q)
	}
	if q := MakePlan("8", []string{"3"}, acct); q.Ready() || q.SourceUser != "" {
		t.Fatalf("unknown source must block: %+v", q)
	}
	if q := MakePlan("1", []string{"2"}, acct); !q.Ready() || len(q.Writes) != 0 || len(q.Same) != 1 {
		t.Fatalf("same account only: %+v", q)
	}
}
