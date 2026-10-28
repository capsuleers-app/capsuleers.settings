package marshal_test

import (
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WilliamFalci/capsuleers.settings/internal/marshal"
	mt "github.com/WilliamFalci/capsuleers.settings/internal/marshal/marshaltest"
)

func sample() any {
	return mt.D(
		"flag", mt.T(mt.Long(134376261513464488), true),
		"num", mt.T(mt.Long(1), int64(-70000)),
		"small", mt.T(mt.Long(1), int64(300)),
		"big", mt.T(mt.Long(1), int64(1)<<40),
		"neg", mt.T(mt.Long(1), int64(-1)),
		"flt", mt.T(mt.Long(1), 0.9),
		"zero", mt.T(mt.Long(1), 0.0),
		"txt", mt.T(mt.Long(1), "hello"),
		"empty", mt.T(mt.Long(1), ""),
		"one", mt.T(mt.Long(1), "x"),
		"none", mt.T(mt.Long(1), nil),
		"list", mt.T(mt.Long(1), mt.L(int64(1), int64(2), int64(3))),
		"nested", mt.T(mt.Long(1), mt.D("a", mt.L(mt.L(), mt.T()), "b", mt.D())),
	)
}

func TestRoundTrip(t *testing.T) {
	for _, o := range []mt.Opts{{}, {ShareKeys: true}, {ShareKeys: true, Reverse: true}} {
		v, err := marshal.Decode(mt.Encode(sample(), o))
		if err != nil {
			t.Fatalf("%+v: %v", o, err)
		}
		d := v.(*marshal.Dict)
		pair := func(k string) any {
			x, ok := d.Get(k)
			if !ok {
				t.Fatalf("missing %s", k)
			}
			return x.(*marshal.Tuple).Items[1]
		}
		if pair("flag") != true || pair("num") != int64(-70000) || pair("small") != int64(300) || pair("big") != int64(1)<<40 || pair("neg") != int64(-1) {
			t.Fatalf("scalars wrong: %+v", o)
		}
		if pair("flt") != 0.9 || pair("zero") != 0.0 || pair("txt") != "hello" || pair("empty") != "" || pair("one") != "x" || pair("none") != nil {
			t.Fatalf("scalars wrong (2): %+v", o)
		}
		if l := pair("list").(*marshal.List); len(l.Items) != 3 || l.Items[2] != int64(3) {
			t.Fatal("list")
		}
		if ts := func() any { x, _ := d.Get("flag"); return x.(*marshal.Tuple).Items[0] }(); ts != int64(134376261513464488) {
			t.Fatalf("timestamp %v", ts)
		}
	}
}

func TestSharedRefs(t *testing.T) {
	for _, rev := range []bool{false, true} {
		v := mt.D("a", mt.Shared{Label: "x", V: mt.L(int64(7))}, "b", mt.RefTo{Label: "x"}, "c", mt.Shared{Label: "s", V: "shared text"}, "d", mt.RefTo{Label: "s"})
		got, err := marshal.Decode(mt.Encode(v, mt.Opts{ShareKeys: true, Reverse: rev}))
		if err != nil {
			t.Fatal(err)
		}
		d := got.(*marshal.Dict)
		a, _ := d.Get("a")
		b, _ := d.Get("b")
		if a != b {
			t.Fatalf("reverse=%v: ref must alias the shared list", rev)
		}
		if s, _ := d.Get("d"); s != "shared text" {
			t.Fatalf("string ref: %v", s)
		}
	}
}

func TestObjects(t *testing.T) {
	obj := &marshal.Object{Kind: "reduce", Head: mt.T(marshal.Global("collections.OrderedDict"), mt.T()), Items: []any{int64(1)}, Keys: []any{"k"}, Vals: []any{"v"}}
	inst := &marshal.Object{Kind: "instance", Head: "some.Class", State: mt.D("x", int64(1))}
	got, err := marshal.Decode(mt.Encode(mt.L(obj, inst, marshal.Global("a.b")), mt.Opts{ShareKeys: true}))
	if err != nil {
		t.Fatal(err)
	}
	l := got.(*marshal.List)
	o := l.Items[0].(*marshal.Object)
	if o.Kind != "reduce" || len(o.Items) != 1 || o.Keys[0] != "k" || o.Vals[0] != "v" {
		t.Fatalf("reduce: %+v", o)
	}
	i := l.Items[1].(*marshal.Object)
	if i.Head != "some.Class" || i.State.(*marshal.Dict).Keys[0] != "x" {
		t.Fatalf("instance: %+v", i)
	}
	if l.Items[2] != marshal.Global("a.b") {
		t.Fatal("global")
	}
}

func TestLongBig(t *testing.T) {
	// 9-byte positive value does not fit in int64.
	data := []byte{0x7e, 0, 0, 0, 0, 0x2f, 9, 1, 0, 0, 0, 0, 0, 0, 0, 1}
	v, err := marshal.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	want := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1))
	if b, ok := v.(*big.Int); !ok || b.Cmp(want) != 0 {
		t.Fatalf("got %v", v)
	}
}

func TestRejectsGarbage(t *testing.T) {
	good := mt.Encode(sample(), mt.Opts{ShareKeys: true})
	cases := map[string][]byte{
		"empty":       nil,
		"short":       {0x7e, 0, 0},
		"bad header":  {0x01, 0, 0, 0, 0, 0x08},
		"huge table":  {0x7e, 0xff, 0xff, 0xff, 0x7f, 0x08},
		"neg table":   {0x7e, 0xff, 0xff, 0xff, 0xff, 0x08},
		"bad map":     {0x7e, 1, 0, 0, 0, 0x53, 1, 'a', 9, 0, 0, 0},
		"no object":   {0x7e, 0, 0, 0, 0, 0x2c, 0x08},
		"bad opcode":  {0x7e, 0, 0, 0, 0, 0x30},
		"unsupported": {0x7e, 0, 0, 0, 0, 0x2a},
		"stray mark":  {0x7e, 0, 0, 0, 0, 0x2d},
		"ref no map":  {0x7e, 0, 0, 0, 0, 0x1b, 1},
		"huge tuple":  {0x7e, 0, 0, 0, 0, 0x14, 255, 0xff, 0xff, 0xff, 0x7f},
		"neg len":     {0x7e, 0, 0, 0, 0, 0x13, 255, 0xff, 0xff, 0xff, 0xff},
		"truncated":   good[:len(good)/2],
	}
	for name, data := range cases {
		if _, err := marshal.Decode(data); !errors.Is(err, marshal.ErrFormat) {
			t.Errorf("%s: want ErrFormat, got %v", name, err)
		}
	}
}

func TestDepthLimit(t *testing.T) {
	var v any = int64(1)
	for i := 0; i < marshal.MaxDepth+10; i++ {
		v = mt.L(v)
	}
	if _, err := marshal.Decode(mt.Encode(v, mt.Opts{})); !errors.Is(err, marshal.ErrFormat) {
		t.Fatalf("deep nesting must fail cleanly: %v", err)
	}
}

func TestInputLimit(t *testing.T) {
	if _, err := marshal.Decode(make([]byte, marshal.MaxInput+1)); !errors.Is(err, marshal.ErrFormat) {
		t.Fatal("oversized input accepted")
	}
}

// TestRealCorpus decodes every core_char_*.dat found under $CAPSULEERS_CORPUS (recursively).
// Skipped by default: real settings files are private and never committed.
func TestRealCorpus(t *testing.T) {
	root := os.Getenv("CAPSULEERS_CORPUS")
	if root == "" {
		t.Skip("CAPSULEERS_CORPUS not set")
	}
	n := 0
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(p) != ".dat" {
			return nil
		}
		b, _ := os.ReadFile(p)
		if _, err := marshal.Decode(b); err != nil {
			t.Errorf("%s: %v", filepath.Base(p), err)
		}
		n++
		return nil
	})
	t.Logf("decoded %d files", n)
}

func FuzzDecode(f *testing.F) {
	f.Add(mt.Encode(sample(), mt.Opts{}))
	f.Add(mt.Encode(sample(), mt.Opts{ShareKeys: true, Reverse: true}))
	f.Add(mt.Encode(mt.D("a", mt.Shared{Label: "x", V: mt.L(int64(7))}, "b", mt.RefTo{Label: "x"}), mt.Opts{ShareKeys: true}))
	f.Add(mt.Encode(mt.L(&marshal.Object{Kind: "newobj", Head: mt.T(mt.T()), Items: []any{int64(1)}, Keys: []any{"k"}, Vals: []any{"v"}}, &marshal.Object{Kind: "instance", Head: "c", State: nil}), mt.Opts{}))
	f.Add([]byte{0x7e, 0, 0, 0, 0, 0x2f, 9, 1, 0, 0, 0, 0, 0, 0, 0, 1})
	f.Add([]byte{0x7e, 0, 0, 0, 0, 0x2d})
	f.Add([]byte{0x7e, 1, 0, 0, 0, 0x53, 1, 'a', 1, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, err := marshal.Decode(data)
		if err != nil && !errors.Is(err, marshal.ErrFormat) {
			t.Fatalf("non-format error: %v", err)
		}
		// Decode converts panics into "internal error"; that means a bounds bug, not bad input.
		if err != nil && strings.Contains(err.Error(), "internal error") {
			t.Fatalf("decoder fault: %v", err)
		}
	})
}
