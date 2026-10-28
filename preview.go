package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/capsuleers-app/capsuleers.settings/internal/charsettings"
)

// Read-only preview of a character's client settings (never writes). A file that cannot be
// decoded gives an honest error; copying does not depend on any of this.

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
	Category  string   `json:"category"`
	Members   []string `json:"members"` // labels of the windows stacked here
	Locked    bool     `json:"locked"`
	Collapsed bool     `json:"collapsed"`
	Compact   bool     `json:"compact"`
	Minimized bool     `json:"minimized"`
}

// PreviewChannel is a chat channel.
type PreviewChannel struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

// PreviewSide describes the source file. Error is "", "missing" or "unreadable".
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
	Display  charsettings.Display `json:"display"`
	Windows  []PreviewWindow      `json:"windows"`
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

// Preview decodes the source character of one profile. lang is "it" or anything else for English.
func (s *Service) Preview(profileDir, sourceID, lang string) PreviewResult {
	dir := filepath.Clean(profileDir)
	if !filepath.IsAbs(dir) || !strings.HasPrefix(filepath.Base(dir), "settings_") || !idRe.MatchString(sourceID) {
		return PreviewResult{Code: "bad_request", Windows: []PreviewWindow{}, Channels: []PreviewChannel{}}
	}
	res := PreviewResult{OK: true, Windows: []PreviewWindow{}, Channels: []PreviewChannel{}, Display: charsettings.LoadDisplay(filepath.Join(dir, "core_public__.yaml"))}
	var src *charsettings.Settings
	res.Source, src = loadSide(dir, sourceID)
	if src == nil {
		return res
	}

	chans := map[string]charsettings.Channel{}
	for _, c := range src.Channels {
		chans[c.ID] = c
	}
	for i := range src.Windows {
		w := &src.Windows[i]
		labels := make([]string, 0, len(w.Members))
		for _, m := range w.Members {
			labels = append(labels, charsettings.WindowLabel(m, lang, chans, nil))
		}
		short := labels
		if len(short) > 3 {
			short = append(append([]string(nil), labels[:3]...), fmt.Sprintf("+%d", len(labels)-3))
		}
		res.Windows = append(res.Windows, PreviewWindow{ID: w.ID, Label: charsettings.WindowLabel(w.ID, lang, chans, short),
			Left: w.Left, Top: w.Top, Width: w.Width, Height: w.Height, ScreenW: w.ScreenW, ScreenH: w.ScreenH,
			Category: charsettings.WindowCategory(w.ID, w.Members), Members: labels, Locked: w.Locked, Collapsed: w.Collapsed, Compact: w.Compact, Minimized: w.Minimized})
	}
	for _, c := range src.Channels {
		res.Channels = append(res.Channels, PreviewChannel{ID: c.ID, Kind: c.Kind, Label: charsettings.ChannelLabel(c, lang)})
	}
	sort.SliceStable(res.Windows, func(i, j int) bool {
		return strings.ToLower(res.Windows[i].Label) < strings.ToLower(res.Windows[j].Label)
	})
	return res
}
