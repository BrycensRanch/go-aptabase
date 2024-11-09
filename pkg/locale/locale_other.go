//go:build !windows
// +build !windows

package locale

import (
	"os"
	"strings"
)

// getUnixLocale gets the locale for Unix-like systems (macOS, Linux, FreeBSD)
func GetLocale() string {
	locale := os.Getenv("LC_ALL") // First try LC_ALL
	if locale == "" {
		locale = os.Getenv("LANG") // Then fallback to LANG
	}
	if locale == "" || locale == "C" || locale == "C." || locale == "C.UTF-8" {
		return "en-US"
	}
	if strings.Contains(locale, ".") {
		locale = strings.Split(locale, ".")[0] // Remove any encoding like UTF-8
	}
	return locale
}
