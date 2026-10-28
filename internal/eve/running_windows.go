//go:build windows

package eve

import (
	"os/exec"
	"strings"
	"syscall"
)

// RunningCount counts exefile.exe processes via tasklist.
func RunningCount() int {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq exefile.exe", "/NH", "/FO", "CSV")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	return strings.Count(strings.ToLower(string(out)), `"exefile.exe"`)
}
