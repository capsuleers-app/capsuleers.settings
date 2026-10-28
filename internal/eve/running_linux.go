//go:build linux

package eve

import "os"

// RunningCount returns the number of running EVE client processes.
func RunningCount() int { return CountProcCmdlines("/proc", os.Getpid()) }
