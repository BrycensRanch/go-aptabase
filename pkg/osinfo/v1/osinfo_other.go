// osinfo_other.go
//go:build !windows
// +build !windows

package osinfo

import (
	"io/ioutil"
	"os/exec"
	"runtime"
	"strings"
)

// GetOSInfo retrieves the OS name and version based on the operating system.
func GetOSInfo() (string, string) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxInfo()
	case "darwin":
		return "macOS", getMacOSVersion()
	case "freebsd":
		return "FreeBSD", getFreeBSDVersion()
	default:
		return runtime.GOOS, "0.0.0"
	}
}

// getLinuxInfo reads the OS release information directly from the filesystem.
func getLinuxInfo() (string, string) {
	data, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		return "Linux", "0.0.0"
	}

	lines := strings.Split(string(data), "\n")
	var name, version string

	for _, line := range lines {
		if strings.HasPrefix(line, "NAME=") {
			name = strings.Trim(strings.SplitN(line, "=", 2)[1], "\"")
		} else if strings.HasPrefix(line, "VERSION=") {
			version = strings.Trim(strings.SplitN(line, "=", 2)[1], "\"")
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			version = strings.Trim(strings.SplitN(line, "=", 2)[1], "\"")
		} else if strings.HasPrefix(line, "VERSION_CODENAME=") {
			version = strings.Trim(strings.SplitN(line, "=", 2)[1], "\"")
		}
	}

	return name, version
}

// getMacOSVersion retrieves the macOS version from the system version file.
func getMacOSVersion() string {
	data, err := ioutil.ReadFile("/System/Library/CoreServices/SystemVersion.plist")
	if err != nil {
		return "0.0.0"
	}

	version := string(data)
	// Extract version from plist data (this is simplified)
	if strings.Contains(version, "ProductVersion") {
		parts := strings.Split(version, "ProductVersion")
		if len(parts) > 1 {
			return strings.TrimSpace(strings.Split(parts[1], "<")[0])
		}
	}

	return "0.0.0"
}

// getFreeBSDVersion retrieves the FreeBSD version from the uname command.
func getFreeBSDVersion() string {
	// Attempt to read the version from /etc/freebsd-version
	data, err := ioutil.ReadFile("/etc/freebsd-version")
	if err == nil {
		return strings.TrimSpace(string(data))
	}

	// Fallback to using freebsd-version command if the file is not available
	output, err := exec.Command("freebsd-version").Output()
	if err != nil {
		return "0.0.0"
	}
	return strings.TrimSpace(string(output))
}
