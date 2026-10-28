// Package marshaltest is a tiny TEST-ONLY encoder for the blue.marshal subset the decoder
// understands. It exists so committed fixtures are synthetic: nothing is copied from real
// settings files. The application itself never imports this package (and never writes the format).
package marshaltest

import (
	"encoding/binary"
	"math"

	"github.com/capsuleers-app/capsuleers.settings/internal/marshal"
)

// Long is emitted as a length-prefixed little-endian integer (opcode 0x2f), like the
// client's FILETIME timestamps.
type Long int64

// Shared wraps a value that gets a slot in the shared table; RefTo points back at it by Label.
type Shared struct {
	Label string
	V     any
}
type RefTo struct{ Label string }

// Opts tweaks the encoding.
type Opts struct {
	// ShareKeys encodes string dict keys as shared buffers (what the real client writes).
	ShareKeys bool
	// Reverse stores the shared-table permutation reversed, to exercise the index mapping.
	Reverse bool
}

type enc struct {
	b      []byte
	opts   Opts
	shared int
	labels map[string]int // label -> encounter index
	refs   []pendingRef
}

type pendingRef struct{ at, enc int }

// Encode builds a complete stream (header, objects, shared table).
func Encode(v any, o Opts) []byte {
	e := &enc{opts: o, labels: map[string]int{}}
	e.b = append(e.b, 0x7e, 0, 0, 0, 0)
	e.value(v)
	n := e.shared
	binary.LittleEndian.PutUint32(e.b[1:], uint32(n))
	for i := 0; i < n; i++ {
		m := i + 1
		if o.Reverse {
			m = n - i
		}
		e.b = binary.LittleEndian.AppendUint32(e.b, uint32(m))
	}
	// Patch references now that the permutation is known.
	for _, r := range e.refs {
		m := r.enc + 1
		if o.Reverse {
			m = n - r.enc
		}
		e.b[r.at] = byte(m) // refs stay below 255 in tests
	}
	return e.b
}

func (e *enc) length(n int) {
	if n < 255 {
		e.b = append(e.b, byte(n))
		return
	}
	e.b = append(e.b, 255)
	e.b = binary.LittleEndian.AppendUint32(e.b, uint32(n))
}

func (e *enc) op(code byte, shared bool) {
	if shared {
		code |= 0x40
		e.shared++
	}
	e.b = append(e.b, code)
}

func (e *enc) str(s string, shared bool) {
	e.op(0x13, shared)
	e.length(len(s))
	e.b = append(e.b, s...)
}

func (e *enc) value(v any) {
	e.valueS(v, false)
}

func (e *enc) valueS(v any, shared bool) {
	switch x := v.(type) {
	case nil:
		e.b = append(e.b, 0x01)
	case bool:
		if x {
			e.b = append(e.b, 0x1f)
		} else {
			e.b = append(e.b, 0x20)
		}
	case int:
		e.value(int64(x))
	case int64:
		switch {
		case x == -1:
			e.b = append(e.b, 0x07)
		case x == 0:
			e.b = append(e.b, 0x08)
		case x == 1:
			e.b = append(e.b, 0x09)
		case x >= math.MinInt8 && x <= math.MaxInt8:
			e.b = append(e.b, 0x06, byte(int8(x)))
		case x >= math.MinInt16 && x <= math.MaxInt16:
			e.b = append(e.b, 0x05)
			e.b = binary.LittleEndian.AppendUint16(e.b, uint16(int16(x)))
		case x >= math.MinInt32 && x <= math.MaxInt32:
			e.b = append(e.b, 0x04)
			e.b = binary.LittleEndian.AppendUint32(e.b, uint32(int32(x)))
		default:
			e.b = append(e.b, 0x03)
			e.b = binary.LittleEndian.AppendUint64(e.b, uint64(x))
		}
	case Long:
		e.op(0x2f, shared)
		e.b = append(e.b, 8)
		e.b = binary.LittleEndian.AppendUint64(e.b, uint64(x))
	case float64:
		if x == 0 {
			e.b = append(e.b, 0x0b)
		} else {
			e.b = append(e.b, 0x0a)
			e.b = binary.LittleEndian.AppendUint64(e.b, math.Float64bits(x))
		}
	case string:
		switch {
		case shared:
			e.str(x, true)
		case x == "":
			e.b = append(e.b, 0x0e)
		case len(x) == 1:
			e.b = append(e.b, 0x0f, x[0])
		case len(x) < 256:
			e.b = append(e.b, 0x10, byte(len(x)))
			e.b = append(e.b, x...)
		default:
			e.op(0x0d, false)
			e.length(len(x))
			e.b = append(e.b, x...)
		}
	case Shared:
		e.labels[x.Label] = e.shared
		e.valueS(x.V, true)
	case RefTo:
		e.b = append(e.b, 0x1b, 0)
		e.refs = append(e.refs, pendingRef{at: len(e.b) - 1, enc: e.labels[x.Label]})
	case *marshal.Tuple:
		switch len(x.Items) {
		case 0:
			e.b = append(e.b, 0x24)
		case 1:
			e.op(0x25, shared)
		case 2:
			e.op(0x2c, shared)
		default:
			e.op(0x14, shared)
			e.length(len(x.Items))
		}
		for _, it := range x.Items {
			e.value(it)
		}
	case *marshal.List:
		switch len(x.Items) {
		case 0:
			e.op(0x26, shared)
		case 1:
			e.op(0x27, shared)
		default:
			e.op(0x15, shared)
			e.length(len(x.Items))
		}
		for _, it := range x.Items {
			e.value(it)
		}
	case *marshal.Dict:
		e.op(0x16, shared)
		e.length(len(x.Keys))
		for i := range x.Keys { // counted dicts store the value before the key
			e.value(x.Vals[i])
			if s, ok := x.Keys[i].(string); ok && e.opts.ShareKeys {
				e.str(s, true)
			} else {
				e.value(x.Keys[i])
			}
		}
	case marshal.Global:
		e.op(0x02, shared)
		e.length(len(x))
		e.b = append(e.b, x...)
	case *marshal.Object:
		code := map[string]byte{"instance": 0x17, "newobj": 0x23, "reduce": 0x22}[x.Kind]
		e.op(code, shared)
		e.value(x.Head)
		if x.Kind == "instance" {
			e.value(x.State)
			return
		}
		for _, it := range x.Items {
			e.value(it)
		}
		e.b = append(e.b, 0x2d)
		for i := range x.Keys {
			e.value(x.Keys[i])
			e.value(x.Vals[i])
		}
		e.b = append(e.b, 0x2d)
	default:
		panic("marshaltest: unsupported value")
	}
}

// Helpers to build values tersely.
func T(items ...any) *marshal.Tuple { return &marshal.Tuple{Items: items} }
func L(items ...any) *marshal.List  { return &marshal.List{Items: items} }

// D builds a dict from alternating key, value arguments.
func D(kv ...any) *marshal.Dict {
	d := &marshal.Dict{}
	for i := 0; i+1 < len(kv); i += 2 {
		d.Keys = append(d.Keys, kv[i])
		d.Vals = append(d.Vals, kv[i+1])
	}
	return d
}
