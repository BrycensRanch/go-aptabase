//go:build windows
// +build windows

package locale

import (
	"os/exec"
	"strings"
	"syscall"
)

// getWindowsLocale gets the user's locale for Windows
func GetLocale() string {
	// Exec powershell Get-Culture on Windows.
	cmd := exec.Command("powershell", "Get-Culture | select -exp Name")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // No flash bangs
	output, err := cmd.Output()
	if err == nil {
		return strings.Trim(string(output), "\r\n")
	}
	return "en-US"
}
