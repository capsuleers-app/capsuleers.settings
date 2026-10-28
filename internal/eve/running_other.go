//go:build !linux && !windows

package eve

import (
	"os/exec"
	"strings"
)

// RunningCount counts EVE client processes on macOS via ps.
func RunningCount() int {
	out, err := exec.Command("ps", "-axo", "comm=").Output()
	if err != nil {
		return 0
	}
	n := 0
	for _, l := range strings.Split(string(out), "\n") {
		l = strings.ToLower(l)
		if strings.Contains(l, "exefile") || strings.HasSuffix(strings.TrimSpace(l), "/eve online") {
			n++
		}
	}
	return n
}
