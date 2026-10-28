package charsettings

import (
	"math"
	"sort"
)

// Status of one item when the source layout is copied over a target.
const (
	Same   = "same"
	Move   = "move"
	Add    = "add"
	Remove = "remove"
)

// WindowDiff pairs a window in the source and in the target (either may be nil).
type WindowDiff struct {
	ID     string
	Status string
	Source *Window
	Target *Window
}

// ChannelDiff is the same for chat channels.
type ChannelDiff struct {
	Channel Channel
	Status  string // same, add (only in source), remove (only in target)
}

// tolerance for comparing positions saved on screens of different size: share of the screen.
const tolerance = 0.004

func frac(v, of int) float64 { return float64(v) / float64(of) }

func sameRect(a, b *Window) bool {
	if a.ScreenW == b.ScreenW && a.ScreenH == b.ScreenH {
		return a.Left == b.Left && a.Top == b.Top && a.Width == b.Width && a.Height == b.Height
	}
	near := func(x, y float64) bool { return math.Abs(x-y) <= tolerance }
	return near(frac(a.Left, a.ScreenW), frac(b.Left, b.ScreenW)) &&
		near(frac(a.Top, a.ScreenH), frac(b.Top, b.ScreenH)) &&
		near(frac(a.Width, a.ScreenW), frac(b.Width, b.ScreenW)) &&
		near(frac(a.Height, a.ScreenH), frac(b.Height, b.ScreenH))
}

// DiffWindows compares by window id: what copying src over tgt would change. Results are sorted by id.
func DiffWindows(src, tgt *Settings) []WindowDiff {
	ids := map[string]*WindowDiff{}
	for i := range src.Windows {
		w := &src.Windows[i]
		ids[w.ID] = &WindowDiff{ID: w.ID, Source: w}
	}
	for i := range tgt.Windows {
		w := &tgt.Windows[i]
		d := ids[w.ID]
		if d == nil {
			d = &WindowDiff{ID: w.ID}
			ids[w.ID] = d
		}
		d.Target = w
	}
	out := make([]WindowDiff, 0, len(ids))
	for _, d := range ids {
		switch {
		case d.Source != nil && d.Target == nil:
			d.Status = Add
		case d.Source == nil:
			d.Status = Remove
		case sameRect(d.Source, d.Target):
			d.Status = Same
		default:
			d.Status = Move
		}
		out = append(out, *d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// DiffChannels compares chat channels by id.
func DiffChannels(src, tgt *Settings) []ChannelDiff {
	in := map[string]bool{}
	for _, c := range tgt.Channels {
		in[c.ID] = true
	}
	var out []ChannelDiff
	seen := map[string]bool{}
	for _, c := range src.Channels {
		seen[c.ID] = true
		st := Add
		if in[c.ID] {
			st = Same
		}
		out = append(out, ChannelDiff{Channel: c, Status: st})
	}
	for _, c := range tgt.Channels {
		if !seen[c.ID] {
			out = append(out, ChannelDiff{Channel: c, Status: Remove})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Channel.ID < out[j].Channel.ID })
	return out
}
