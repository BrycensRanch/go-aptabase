//go:build windows
// +build windows

package osinfo

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func GetOSInfo() (string, string) {
	return "Windows", getWindowsVersion()
}

// getWindowsVersion retrieves the Windows version information.
func getWindowsVersion() string {
	MajorVersion, MinorVersion, BuildNumber := windows.RtlGetNtVersionNumbers()

	// Construct version string (major.minor.build)
	return formatVersion(MajorVersion, MinorVersion, BuildNumber)
}

// formatVersion formats the version number into a string.
func formatVersion(major, minor, build uint32) string {
	return fmt.Sprintf("%d.%d.%d", major, minor, build)
}
