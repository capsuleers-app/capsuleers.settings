package eve

import (
	"os"
	"path/filepath"
	"strings"
)

// CountProcCmdlines counts processes under a /proc-like root with an argument whose file name is
// exefile.exe (the EVE client under Wine/Proton). A mere mention inside a longer argument (a shell
// script, grep, an editor) does not count. skipPID is excluded.
func CountProcCmdlines(procRoot string, skipPID int) int {
	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		name := e.Name()
		if name == "" || name[0] < '0' || name[0] > '9' {
			continue
		}
		if skipPID > 0 && name == itoa(skipPID) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(procRoot, name, "cmdline"))
		if err != nil {
			continue
		}
		if isClientCmdline(b) {
			n++
		}
	}
	return n
}

func isClientCmdline(b []byte) bool {
	for _, arg := range strings.Split(string(b), "\x00") {
		arg = strings.ToLower(strings.TrimSpace(arg))
		if i := strings.LastIndexAny(arg, `/\`); i >= 0 {
			arg = arg[i+1:]
		}
		if arg == "exefile.exe" {
			return true
		}
	}
	return false
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for ; i > 0; i /= 10 {
		b = append([]byte{byte('0' + i%10)}, b...)
	}
	return string(b)
}
