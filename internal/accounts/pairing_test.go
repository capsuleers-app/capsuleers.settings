package accounts

import (
	"reflect"
	"testing"
	"time"
)

func snap(kv ...any) Snapshot {
	s := Snapshot{}
	for i := 0; i < len(kv); i += 2 {
		s[kv[i].(string)] = int64(kv[i+1].(int))
	}
	return s
}

func TestPairSingleUser(t *testing.T) {
	prev := snap("core_char_1.dat", 1, "core_char_2.dat", 1, "core_char_3.dat", 1, "core_user_10.dat", 1, "core_user_20.dat", 1)
	cur := snap("core_char_1.dat", 5, "core_char_2.dat", 5, "core_char_3.dat", 1, "core_user_10.dat", 5, "core_user_20.dat", 1,
		"core_char_4.dat", 5) // a new file counts as changed
	got := Pair(prev, cur)
	want := Outcome{Kind: PairMapped, UserID: "10", Chars: []string{"1", "2", "4"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestPairAmbiguousAndNone(t *testing.T) {
	prev := snap("core_char_1.dat", 1, "core_char_2.dat", 1, "core_user_10.dat", 1, "core_user_20.dat", 1)
	if o := Pair(prev, snap("core_char_1.dat", 2, "core_char_2.dat", 2, "core_user_10.dat", 2, "core_user_20.dat", 2)); o.Kind != PairAmbiguous || o.UserID != "" || len(o.Chars) != 0 {
		t.Fatalf("two users must be ambiguous: %+v", o)
	}
	// character changed, no account file: nothing
	if o := Pair(prev, snap("core_char_1.dat", 2, "core_char_2.dat", 1, "core_user_10.dat", 1, "core_user_20.dat", 1)); o.Kind != PairNone {
		t.Fatalf("no user: %+v", o)
	}
	// account file alone: nothing to map
	if o := Pair(prev, snap("core_char_1.dat", 1, "core_char_2.dat", 1, "core_user_10.dat", 3, "core_user_20.dat", 1)); o.Kind != PairNone {
		t.Fatalf("user only: %+v", o)
	}
	if o := Pair(prev, prev); o.Kind != PairNone {
		t.Fatalf("no change: %+v", o)
	}
}

func TestTrackerSettlesAndRemaps(t *testing.T) {
	base := snap("core_char_1.dat", 1, "core_char_2.dat", 1, "core_user_10.dat", 1, "core_user_20.dat", 1)
	tr := NewTracker(base, 3*time.Second)
	now := t0
	step := func(s Snapshot, d time.Duration) (Outcome, bool) { now = now.Add(d); return tr.Observe(s, now) }

	if _, ok := step(base, time.Second); ok {
		t.Fatal("no change, no outcome")
	}
	s1 := snap("core_char_1.dat", 2, "core_char_2.dat", 1, "core_user_10.dat", 2, "core_user_20.dat", 1)
	if _, ok := step(s1, time.Second); ok {
		t.Fatal("must wait for the changes to settle")
	}
	if _, ok := step(s1, 2*time.Second); ok {
		t.Fatal("only 2 s settled")
	}
	o, ok := step(s1, 2*time.Second)
	if !ok || o.Kind != PairMapped || o.UserID != "10" || !reflect.DeepEqual(o.Chars, []string{"1"}) {
		t.Fatalf("got %+v %v", o, ok)
	}
	if _, ok := step(s1, 10*time.Second); ok {
		t.Fatal("re-baselined: the same state must not fire again")
	}
	// the same character now logs in on the other account: remap
	s2 := snap("core_char_1.dat", 3, "core_char_2.dat", 1, "core_user_10.dat", 2, "core_user_20.dat", 3)
	step(s2, time.Second)
	o, ok = step(s2, 4*time.Second)
	if !ok || o.UserID != "20" || !reflect.DeepEqual(o.Chars, []string{"1"}) {
		t.Fatalf("remap: %+v %v", o, ok)
	}
	// two clients at once in one batch: ambiguous, then listening continues
	s3 := snap("core_char_1.dat", 4, "core_char_2.dat", 4, "core_user_10.dat", 4, "core_user_20.dat", 4)
	step(s3, time.Second)
	if o, ok = step(s3, 4*time.Second); !ok || o.Kind != PairAmbiguous {
		t.Fatalf("ambiguous: %+v %v", o, ok)
	}
}
