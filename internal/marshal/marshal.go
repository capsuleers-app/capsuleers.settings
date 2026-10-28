// Package marshal is a READ-ONLY decoder for CCP "blue.marshal" streams (the format of the
// EVE client's core_char_*.dat settings files) into generic Go values.
//
// Opcode reference: github.com/ntt/reverence, src/blue/marshal.c. The package never encodes:
// the app copies whole files and must never rewrite this format.
//
// Value mapping: None -> nil, bool, int64 (or *big.Int when it does not fit), float64,
// string (byte strings, buffers and unicode all become Go strings), *Tuple, *List, *Dict,
// Global (a class/function name), *Object (instance, newobj, reduce).
// The decoder is bounds-checked, iterative (no recursion), limited in depth and input
// size, and returns an error instead of panicking on malformed data.
package marshal

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"unicode/utf16"
)

const (
	protocolID = 0x7e
	sharedFlag = 0x40

	// Limits: settings files are ~25 KB; these only stop pathological input.
	MaxInput    = 16 << 20
	MaxDepth    = 128
	maxPrealloc = 1024
)

// Opcodes (see reverence marshal.h).
const (
	tNone     = 0x01
	tGlobal   = 0x02
	tInt64    = 0x03
	tInt32    = 0x04
	tInt16    = 0x05
	tInt8     = 0x06
	tMinusOne = 0x07
	tZero     = 0x08
	tOne      = 0x09
	tFloat    = 0x0a
	tFloat0   = 0x0b
	tStringL  = 0x0d
	tString0  = 0x0e
	tString1  = 0x0f
	tString   = 0x10
	tStringR  = 0x11
	tUnicode  = 0x12
	tBuffer   = 0x13
	tTuple    = 0x14
	tList     = 0x15
	tDict     = 0x16
	tInstance = 0x17
	tRef      = 0x1b
	tChecksum = 0x1c
	tTrue     = 0x1f
	tFalse    = 0x20
	tReduce   = 0x22
	tNewObj   = 0x23
	tTuple0   = 0x24
	tTuple1   = 0x25
	tList0    = 0x26
	tList1    = 0x27
	tUnicode0 = 0x28
	tUnicode1 = 0x29
	tStream   = 0x2b
	tTuple2   = 0x2c
	tMark     = 0x2d
	tUTF8     = 0x2e
	tLong     = 0x2f
)

// Tuple, List and Dict are pointer types so shared references (and cycles) are preserved.
type Tuple struct{ Items []any }
type List struct{ Items []any }

// Dict keeps insertion order; keys may be any value.
type Dict struct {
	Keys, Vals []any
	idx        map[string]int
}

// Get returns the value stored under a string key.
func (d *Dict) Get(key string) (any, bool) {
	if d == nil {
		return nil, false
	}
	if d.idx == nil {
		d.idx = make(map[string]int, len(d.Keys))
		for i, k := range d.Keys {
			if s, ok := k.(string); ok {
				if _, dup := d.idx[s]; !dup {
					d.idx[s] = i
				}
			}
		}
	}
	i, ok := d.idx[key]
	if !ok {
		return nil, false
	}
	return d.Vals[i], true
}

// Global is a by-name reference to a class or function.
type Global string

// Object is an instance / newobj / reduce record. Head is the class name (instance) or the
// constructor tuple; State is only set for instances; Items and Keys/Vals are the list and
// dict iterator payloads of newobj/reduce.
type Object struct {
	Kind       string // "instance", "newobj", "reduce"
	Head       any
	State      any
	Items      []any
	Keys, Vals []any
}

type mark struct{}

type frame struct {
	kind     byte // tTuple, tList, tDict, tInstance, tNewObj, tReduce
	left     int
	tup      *Tuple
	list     *List
	dict     *Dict
	obj      *Object
	phase    int // objects: 0 head, 1 state (instance) / items, 2 pairs
	slot     int // reserved shared slot of an object (-1: none)
	half     any
	haveHalf bool
}

type decoder struct {
	data   []byte
	pos    int
	end    int
	slots  []any
	filled []bool
	smap   []int
	scount int
	stack  []frame
	root   any
	done   bool
}

// ErrFormat wraps every decoding failure.
var ErrFormat = errors.New("marshal: invalid stream")

func fail(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrFormat, fmt.Sprintf(format, a...))
}

// Decode parses a complete marshal stream. It never panics: any internal fault is returned as an error.
func Decode(data []byte) (v any, err error) {
	defer func() {
		if r := recover(); r != nil {
			v, err = nil, fail("internal error: %v", r)
		}
	}()
	if len(data) > MaxInput {
		return nil, fail("input too large (%d bytes)", len(data))
	}
	if len(data) < 6 || data[0] != protocolID {
		return nil, fail("missing protocol header")
	}
	n := int(int32(le32(data[1:])))
	if n < 0 || n > (len(data)-5)/4 {
		return nil, fail("shared table does not fit the stream")
	}
	d := &decoder{data: data, pos: 5, end: len(data) - 4*n, slots: make([]any, n), filled: make([]bool, n), smap: make([]int, n)}
	for i := 0; i < n; i++ {
		m := int(int32(le32(data[d.end+4*i:])))
		if m < 1 || m > n {
			return nil, fail("bad shared table entry")
		}
		d.smap[i] = m - 1
	}
	if err := d.run(); err != nil {
		return nil, err
	}
	return d.root, nil
}

func le32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func (d *decoder) need(n int) error {
	if n < 0 || n > d.end-d.pos {
		return fail("unexpected end of data at %d", d.pos)
	}
	return nil
}

func (d *decoder) readLen() (int, error) {
	if err := d.need(1); err != nil {
		return 0, err
	}
	l := int(d.data[d.pos])
	d.pos++
	if l == 255 {
		if err := d.need(4); err != nil {
			return 0, err
		}
		l = int(int32(le32(d.data[d.pos:])))
		d.pos += 4
		if l < 0 {
			return 0, fail("negative length")
		}
	}
	return l, nil
}

func (d *decoder) take(n int) ([]byte, error) {
	if err := d.need(n); err != nil {
		return nil, err
	}
	b := d.data[d.pos : d.pos+n]
	d.pos += n
	return b, nil
}

// reserve takes the next shared slot (encounter order) when the opcode carries the shared flag.
func (d *decoder) reserve(shared bool) (int, error) {
	if !shared {
		return -1, nil
	}
	if d.scount >= len(d.smap) {
		return 0, fail("shared object table overflow")
	}
	s := d.smap[d.scount]
	d.scount++
	return s, nil
}

func (d *decoder) fill(slot int, v any) {
	if slot >= 0 {
		d.slots[slot], d.filled[slot] = v, true
	}
}

func (d *decoder) store(shared bool, v any) error {
	s, err := d.reserve(shared)
	if err != nil {
		return err
	}
	d.fill(s, v)
	return nil
}

func (d *decoder) push(f frame) error {
	if len(d.stack) >= MaxDepth {
		return fail("nesting too deep")
	}
	d.stack = append(d.stack, f)
	return nil
}

func prealloc(n int) int {
	if n > maxPrealloc {
		return maxPrealloc
	}
	return n
}

func (d *decoder) run() error {
	for !d.done {
		if d.pos >= d.end {
			return fail("not enough objects in stream")
		}
		t := d.data[d.pos]
		d.pos++
		shared := t&sharedFlag != 0
		t &^= sharedFlag

		var obj any
		switch t {
		case tNone:
			obj = nil
		case tTrue:
			obj = true
		case tFalse:
			obj = false
		case tMinusOne:
			obj = int64(-1)
		case tZero:
			obj = int64(0)
		case tOne:
			obj = int64(1)
		case tFloat0:
			obj = float64(0)
		case tString0, tUnicode0:
			obj = ""
		case tTuple0:
			obj = &Tuple{}
		case tInt8:
			b, err := d.take(1)
			if err != nil {
				return err
			}
			obj = int64(int8(b[0]))
		case tInt16:
			b, err := d.take(2)
			if err != nil {
				return err
			}
			obj = int64(int16(uint16(b[0]) | uint16(b[1])<<8))
		case tInt32:
			b, err := d.take(4)
			if err != nil {
				return err
			}
			obj = int64(int32(le32(b)))
		case tInt64:
			b, err := d.take(8)
			if err != nil {
				return err
			}
			obj = int64(le64(b))
		case tFloat:
			b, err := d.take(8)
			if err != nil {
				return err
			}
			obj = math.Float64frombits(le64(b))
		case tLong:
			l, err := d.readLen()
			if err != nil {
				return err
			}
			b, err := d.take(l)
			if err != nil {
				return err
			}
			obj = bigFromLE(b)
			if err := d.store(shared, obj); err != nil {
				return err
			}
		case tChecksum:
			// Read-only decoder: the adler32 of the rest of the stream is not verified.
			if _, err := d.take(4); err != nil {
				return err
			}
			continue
		case tStringR:
			l, err := d.readLen()
			if err != nil {
				return err
			}
			obj = fmt.Sprintf("<strtab:%d>", l) // the client's string table is not available
		case tString:
			if err := d.need(1); err != nil {
				return err
			}
			l := int(d.data[d.pos])
			d.pos++
			b, err := d.take(l)
			if err != nil {
				return err
			}
			obj = string(b)
		case tString1:
			b, err := d.take(1)
			if err != nil {
				return err
			}
			obj = string(b)
		case tStringL, tBuffer, tStream:
			l, err := d.readLen()
			if err != nil {
				return err
			}
			b, err := d.take(l)
			if err != nil {
				return err
			}
			obj = string(b)
			if err := d.store(shared, obj); err != nil {
				return err
			}
		case tUnicode1:
			b, err := d.take(2)
			if err != nil {
				return err
			}
			obj = string(utf16.Decode([]uint16{uint16(b[0]) | uint16(b[1])<<8}))
		case tUnicode:
			l, err := d.readLen()
			if err != nil {
				return err
			}
			if l > (d.end-d.pos)/2 {
				return fail("unexpected end of data at %d", d.pos)
			}
			b, _ := d.take(l * 2)
			u := make([]uint16, l)
			for i := range u {
				u[i] = uint16(b[2*i]) | uint16(b[2*i+1])<<8
			}
			obj = string(utf16.Decode(u))
		case tUTF8:
			l, err := d.readLen()
			if err != nil {
				return err
			}
			b, err := d.take(l)
			if err != nil {
				return err
			}
			obj = strings.ToValidUTF8(string(b), "�")
		case tTuple1, tTuple2, tTuple:
			n := 1
			switch t {
			case tTuple2:
				n = 2
			case tTuple:
				var err error
				if n, err = d.readLen(); err != nil {
					return err
				}
			}
			if err := d.need(n); err != nil {
				return err
			}
			tp := &Tuple{Items: make([]any, 0, prealloc(n))}
			if n == 0 {
				if err := d.store(shared, tp); err != nil {
					return err
				}
				obj = tp
				break
			}
			if err := d.store(shared, tp); err != nil {
				return err
			}
			if err := d.push(frame{kind: tTuple, left: n, tup: tp, slot: -1}); err != nil {
				return err
			}
			continue
		case tList0:
			obj = &List{}
			if err := d.store(shared, obj); err != nil {
				return err
			}
		case tList1, tList:
			n := 1
			if t == tList {
				var err error
				if n, err = d.readLen(); err != nil {
					return err
				}
			}
			if err := d.need(n); err != nil {
				return err
			}
			l := &List{Items: make([]any, 0, prealloc(n))}
			if err := d.store(shared, l); err != nil {
				return err
			}
			if n == 0 {
				obj = l
				break
			}
			if err := d.push(frame{kind: tList, left: n, list: l, slot: -1}); err != nil {
				return err
			}
			continue
		case tDict:
			n, err := d.readLen()
			if err != nil {
				return err
			}
			if err := d.need(n * 2); err != nil { // n <= int32 max, no overflow on 64-bit
				return err
			}
			dc := &Dict{Keys: make([]any, 0, prealloc(n)), Vals: make([]any, 0, prealloc(n))}
			if err := d.store(shared, dc); err != nil {
				return err
			}
			if n == 0 {
				obj = dc
				break
			}
			if err := d.push(frame{kind: tDict, left: n * 2, dict: dc, slot: -1}); err != nil {
				return err
			}
			continue
		case tRef:
			l, err := d.readLen()
			if err != nil {
				return err
			}
			if l < 1 || l > len(d.slots) {
				return fail("shared reference out of range")
			}
			if !d.filled[l-1] {
				return fail("shared reference to an object not yet built")
			}
			obj = d.slots[l-1]
		case tGlobal:
			l, err := d.readLen()
			if err != nil {
				return err
			}
			b, err := d.take(l)
			if err != nil {
				return err
			}
			obj = Global(b)
			if err := d.store(shared, obj); err != nil {
				return err
			}
		case tInstance, tNewObj, tReduce:
			slot, err := d.reserve(shared)
			if err != nil {
				return err
			}
			kind := map[byte]string{tInstance: "instance", tNewObj: "newobj", tReduce: "reduce"}[t]
			if err := d.push(frame{kind: t, obj: &Object{Kind: kind}, slot: slot}); err != nil {
				return err
			}
			continue
		case tMark:
			obj = mark{}
		default:
			return fail("unsupported opcode 0x%02x at %d", t, d.pos-1)
		}
		if err := d.deliver(obj); err != nil {
			return err
		}
	}
	return nil
}

// deliver hands a finished value to the innermost open container, closing containers that fill up.
func (d *decoder) deliver(obj any) error {
	for {
		if len(d.stack) == 0 {
			if _, isMark := obj.(mark); isMark {
				return fail("stray mark")
			}
			d.root, d.done = obj, true
			return nil
		}
		f := &d.stack[len(d.stack)-1]
		_, isMark := obj.(mark)
		finished := false
		var value any
		switch f.kind {
		case tTuple:
			if isMark {
				return fail("mark inside tuple")
			}
			f.tup.Items = append(f.tup.Items, obj)
			f.left--
			finished, value = f.left == 0, f.tup
		case tList:
			if isMark {
				return fail("mark inside list")
			}
			f.list.Items = append(f.list.Items, obj)
			f.left--
			finished, value = f.left == 0, f.list
		case tDict:
			if isMark {
				return fail("mark inside dict")
			}
			if f.haveHalf { // counted dicts store value first, then key
				f.dict.Keys = append(f.dict.Keys, obj)
				f.dict.Vals = append(f.dict.Vals, f.half)
				f.haveHalf, f.half = false, nil
			} else {
				f.half, f.haveHalf = obj, true
			}
			f.left--
			finished, value = f.left == 0, f.dict
		default: // instance, newobj, reduce
			o := f.obj
			switch {
			case f.phase == 0:
				if isMark {
					return fail("mark as object head")
				}
				o.Head = obj
				d.fill(f.slot, o)
				f.phase = 1
			case f.kind == tInstance:
				if isMark {
					return fail("mark as object state")
				}
				o.State = obj
				finished = true
			case f.phase == 1:
				if isMark {
					f.phase = 2
				} else {
					o.Items = append(o.Items, obj)
				}
			default: // phase 2: key, value pairs until the closing mark
				if isMark {
					if f.haveHalf {
						return fail("dangling key in object")
					}
					finished = true
				} else if f.haveHalf {
					o.Keys = append(o.Keys, f.half)
					o.Vals = append(o.Vals, obj)
					f.haveHalf, f.half = false, nil
				} else {
					f.half, f.haveHalf = obj, true
				}
			}
			value = o
		}
		if !finished {
			return nil
		}
		obj = value
		d.stack = d.stack[:len(d.stack)-1]
	}
}

func le64(b []byte) uint64 {
	return uint64(le32(b)) | uint64(le32(b[4:]))<<32
}

// bigFromLE converts a little-endian two's complement integer; values that fit int64 stay int64.
func bigFromLE(b []byte) any {
	if len(b) == 0 {
		return int64(0)
	}
	neg := b[len(b)-1]&0x80 != 0
	be := make([]byte, len(b))
	for i := range b {
		be[len(b)-1-i] = b[i]
	}
	if neg {
		for i := range be {
			be[i] = ^be[i]
		}
	}
	n := new(big.Int).SetBytes(be)
	if neg {
		n.Add(n, big.NewInt(1))
		n.Neg(n)
	}
	if n.IsInt64() {
		return n.Int64()
	}
	return n
}
