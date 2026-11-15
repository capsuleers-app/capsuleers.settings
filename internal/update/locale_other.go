//go:build !windows

package update

import "os"

// SystemLocale returns the first locale set in the usual environment variables.
func SystemLocale() string {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANGUAGE", "LANG"} {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}
