package charsettings

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Display is what the profile's core_public__.yaml says about the screen. It is profile-wide.
type Display struct {
	Known           bool    `json:"known"`
	ScreenW         int     `json:"screenW"` // full screen settings
	ScreenH         int     `json:"screenH"`
	ScaleFullscreen float64 `json:"scaleFullscreen"` // 0 when not stored
	ScaleWindowed   float64 `json:"scaleWindowed"`
}

var (
	yamlScaleRe  = regexp.MustCompile(`^\s{2}(UIScaleFullscreen|UIScaleWindowed):\s*\[\s*\d+\s*,\s*([0-9.]+)\s*\]`)
	yamlWidthRe  = regexp.MustCompile(`\bwidth:\s*(\d+)`)
	yamlHeightRe = regexp.MustCompile(`\bheight:\s*(\d+)`)
)

// LoadDisplay reads the few device keys needed; missing or odd files give Known=false.
// It is a line scanner, not a YAML parser: the file is machine written in a fixed shape.
func LoadDisplay(path string) Display {
	st, err := os.Stat(path)
	if err != nil || !st.Mode().IsRegular() || st.Size() > 1<<20 {
		return Display{}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return Display{}
	}
	var d Display
	inDevice, inFull := false, false
	var full strings.Builder
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimRight(line, "\r")
		if line != "" && line[0] != ' ' {
			inDevice = line == "device:"
			inFull = false
			continue
		}
		if !inDevice {
			continue
		}
		if m := yamlScaleRe.FindStringSubmatch(line); m != nil {
			f, _ := strconv.ParseFloat(m[2], 64)
			if m[1] == "UIScaleFullscreen" {
				d.ScaleFullscreen = f
			} else {
				d.ScaleWindowed = f
			}
			d.Known = true
		}
		if strings.HasPrefix(line, "  FullScreenSettings:") {
			inFull = true
			continue
		}
		if inFull {
			if len(line) > 2 && line[2] != ' ' && line[2] != '-' { // next key at the same indent
				inFull = false
				continue
			}
			full.WriteString(line)
			full.WriteByte(' ')
		}
	}
	if m := yamlWidthRe.FindStringSubmatch(full.String()); m != nil {
		d.ScreenW, _ = strconv.Atoi(m[1])
	}
	if m := yamlHeightRe.FindStringSubmatch(full.String()); m != nil {
		d.ScreenH, _ = strconv.Atoi(m[1])
	}
	if d.ScreenW > 0 && d.ScreenH > 0 {
		d.Known = true
	}
	return d
}
