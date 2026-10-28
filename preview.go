package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/WilliamFalci/capsuleers.settings/internal/charsettings"
)

// Read-only preview of a character's client settings (never writes). A file that cannot be
// decoded gives an honest per-side error; copying does not depend on any of this.

var idRe = regexp.MustCompile(`^\d{1,20}$`)

// PreviewWindow is one window or stack of windows as saved.
type PreviewWindow struct {
	ID        string   `json:"id"`
	Label     string   `json:"label"`
	Left      int      `json:"left"`
	Top       int      `json:"top"`
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	ScreenW   int      `json:"screenW"`
	ScreenH   int      `json:"screenH"`
	Members   []string `json:"members"` // labels of the windows stacked here
	Locked    bool     `json:"locked"`
	Collapsed bool     `json:"collapsed"`
	Compact   bool     `json:"compact"`
	Minimized bool     `json:"minimized"`
}

// PreviewItem pairs a window in the source with the same id in the target. Status is
// same, move, add (only in source) or remove (only in target); empty without a target.
type PreviewItem struct {
	ID     string         `json:"id"`
	Label  string         `json:"label"`
	Status string         `json:"status"`
	Source *PreviewWindow `json:"source"`
	Target *PreviewWindow `json:"target"`
}

// PreviewChannel is a chat channel; Status is same, add, remove, or empty without a target.
type PreviewChannel struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Label  string `json:"label"`
	Status string `json:"status"`
}

// PreviewSide describes one file. Error is "", "missing" or "unreadable".
type PreviewSide struct {
	CharID   string             `json:"charId"`
	Readable bool               `json:"readable"`
	Error    string             `json:"error"`
	Windows  int                `json:"windows"`
	Stacks   int                `json:"stacks"`
	Locked   int                `json:"locked"`
	Channels int                `json:"channels"`
	ScreenW  int                `json:"screenW"` // screen size most windows were saved at
	ScreenH  int                `json:"screenH"`
	Keys     []charsettings.Key `json:"keys"`
}

// PreviewResult is what the Preview view shows. Code is "bad_request" when the ids or folder are not valid.
type PreviewResult struct {
	OK       bool                 `json:"ok"`
	Code     string               `json:"code"`
	Source   PreviewSide          `json:"source"`
	Target   *PreviewSide         `json:"target"`
	Display  charsettings.Display `json:"display"`
	Items    []PreviewItem        `json:"items"`
	Channels []PreviewChannel     `json:"channels"`
}

func charPath(dir, id string) (string, bool) {
	if !idRe.MatchString(id) {
		return "", false
	}
	return filepath.Join(dir, "core_char_"+id+".dat"), true
}

func loadSide(dir, id string) (PreviewSide, *charsettings.Settings) {
	side := PreviewSide{CharID: id, Keys: []charsettings.Key{}}
	p, ok := charPath(dir, id)
	if !ok {
		side.Error = "missing"
		return side, nil
	}
	s, err := charsettings.Load(p)
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		side.Error = "missing"
		return side, nil
	default:
		side.Error = "unreadable"
		return side, nil
	}
	side.Readable = true
	side.Windows, side.Channels, side.ScreenW, side.ScreenH = len(s.Windows), len(s.Channels), s.ScreenW, s.ScreenH
	for _, w := range s.Windows {
		if len(w.Members) > 0 {
			side.Stacks++
		}
		if w.Locked {
			side.Locked++
		}
	}
	if s.Keys != nil {
		side.Keys = s.Keys
	}
	return side, s
}

// Preview decodes the source (and optionally a target) of one profile and compares them.
// lang is "it" or anything else for English. targetID may be empty.
func (s *Service) Preview(profileDir, sourceID, targetID, lang string) PreviewResult {
	dir := filepath.Clean(profileDir)
	if !filepath.IsAbs(dir) || !strings.HasPrefix(filepath.Base(dir), "settings_") || !idRe.MatchString(sourceID) || (targetID != "" && !idRe.MatchString(targetID)) {
		return PreviewResult{Code: "bad_request", Items: []PreviewItem{}, Channels: []PreviewChannel{}}
	}
	res := PreviewResult{OK: true, Items: []PreviewItem{}, Channels: []PreviewChannel{}, Display: charsettings.LoadDisplay(filepath.Join(dir, "core_public__.yaml"))}
	var src, tgt *charsettings.Settings
	res.Source, src = loadSide(dir, sourceID)
	if targetID != "" {
		var side PreviewSide
		side, tgt = loadSide(dir, targetID)
		res.Target = &side
	}
	if src == nil {
		return res
	}

	chans := map[string]charsettings.Channel{}
	for _, c := range src.Channels {
		chans[c.ID] = c
	}
	if tgt != nil {
		for _, c := range tgt.Channels {
			if _, ok := chans[c.ID]; !ok {
				chans[c.ID] = c
			}
		}
	}
	toWin := func(w *charsettings.Window) *PreviewWindow {
		if w == nil {
			return nil
		}
		labels := make([]string, 0, len(w.Members))
		for _, m := range w.Members {
			labels = append(labels, charsettings.WindowLabel(m, lang, chans, nil))
		}
		short := labels
		if len(short) > 3 {
			short = append(append([]string(nil), labels[:3]...), fmt.Sprintf("+%d", len(labels)-3))
		}
		return &PreviewWindow{ID: w.ID, Label: charsettings.WindowLabel(w.ID, lang, chans, short),
			Left: w.Left, Top: w.Top, Width: w.Width, Height: w.Height, ScreenW: w.ScreenW, ScreenH: w.ScreenH,
			Members: labels, Locked: w.Locked, Collapsed: w.Collapsed, Compact: w.Compact, Minimized: w.Minimized}
	}

	if tgt != nil {
		for _, d := range charsettings.DiffWindows(src, tgt) {
			it := PreviewItem{ID: d.ID, Status: d.Status, Source: toWin(d.Source), Target: toWin(d.Target)}
			it.Label = firstLabel(it)
			res.Items = append(res.Items, it)
		}
		for _, d := range charsettings.DiffChannels(src, tgt) {
			res.Channels = append(res.Channels, PreviewChannel{ID: d.Channel.ID, Kind: d.Channel.Kind, Label: charsettings.ChannelLabel(d.Channel, lang), Status: d.Status})
		}
	} else {
		for i := range src.Windows {
			it := PreviewItem{ID: src.Windows[i].ID, Source: toWin(&src.Windows[i])}
			it.Label = firstLabel(it)
			res.Items = append(res.Items, it)
		}
		for _, c := range src.Channels {
			res.Channels = append(res.Channels, PreviewChannel{ID: c.ID, Kind: c.Kind, Label: charsettings.ChannelLabel(c, lang)})
		}
	}
	sort.SliceStable(res.Items, func(i, j int) bool { return strings.ToLower(res.Items[i].Label) < strings.ToLower(res.Items[j].Label) })
	return res
}

func firstLabel(it PreviewItem) string {
	if it.Source != nil {
		return it.Source.Label
	}
	return it.Target.Label
}
