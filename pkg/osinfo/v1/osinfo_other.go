// osinfo_other.go
//go:build !windows
// +build !windows

package osinfo

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

func ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
func OpenFile(filePath string) (*os.File, error) {
	return os.Open(filePath)
}
func Exec(cmd ...string) (*exec.Cmd) {
	command := exec.Command(cmd[0], cmd[1:]...)

	return command
}

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
		return runtime.GOOS, ""
	}
}

func getLinuxDistroFromProcVersion() (string, string) {
	// Open /proc/version for reading
	// On Firejail's default profile, this is allowed. :)
	// This *WILL* misreport when the program is ran under something like Docker that masquerades as another Linux distribution but uses the same kernel for better performance.
	file, err := OpenFile("/proc/version")
	if err != nil {
		return fallbackToLinuxVersion()
	}
	defer file.Close()

	// Read the contents of the file
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "Ubuntu") {
			// There is no way of getting the Ubuntu version from /proc/version, give up.
			// While this may be viewed as misreporting, this is *still* useful information.
			// Why? Because Ubuntu Noble uses Kernel version 6.8.* anyone who knows Ubuntu knows this.
			return "Ubuntu", getKernelVersion()
		}
		if strings.Contains(line, "WSL") {
			if strings.Contains(line, "WSL2") {
				return "WSL", "2"
			} else {
				// Guess. /shrug
				return "WSL", "1"
			}
		}

		// Use a regex to capture the distribution information
		re := regexp.MustCompile(`Linux version ([^\s]+) ([^\s]+).*?(\w+)\s+(\d+\.\d+\.\d+.*)`)
		matches := re.FindStringSubmatch(line)
		kernel := matches[1]
		// kernelBuilder := matches[2]

		distName, distVersion := getDistributionInfo(kernel)
		if distName != "" {
			return distName, distVersion
		}

		return fallbackToLinuxVersion()
	} else {
		return fallbackToLinuxVersion()
	}
}

func getDistributionInfo(kernelVersion string) (string, string) {
	if strings.Contains(kernelVersion, "fc") {
		parts := strings.Split(kernelVersion, "-")
		for _, part := range parts {
			if strings.HasPrefix(part, "fc") {
				return "Fedora Linux", part
			}
		}
	} else if strings.Contains(kernelVersion, "mga") {
		// Older versions of Mageia Linux ie 6 did not suffix with a number. Likely doesn't matter because this module is go1.22+
		parts := strings.Split(kernelVersion, "-")
		re := regexp.MustCompile("[^0-9]")
		versionNumber := re.ReplaceAllString(parts[2], "")
		return "Mageia", versionNumber

	} else if strings.Contains(kernelVersion, "-arch") {
		return "Arch Linux", "rolling"
	}

	return "", ""
}

func getKernelVersion() string {
	var uname unix.Utsname
	err := unix.Uname(&uname)
	if err != nil {
		return ""
	}
	return string(uname.Release[:])
}

func fallbackToLinuxVersion() (string, string) {
	return "Linux", getKernelVersion()
}

func parseLSBReleaseOrFallback() (string, string) {
	// Execute the lsb_release command with the -a option
	// The actual file is flaky to exist across Linux distros so we use the command instead.
	cmd := Exec("lsb_release", "-a")
	output, err := cmd.Output()
	if err != nil {
		return getLinuxDistroFromProcVersion()
	}

	// Read the output line by line
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var distro, version string
	for scanner.Scan() {
		line := scanner.Text()

		// Extract the Distribution (Distributor ID) and Version
		if strings.HasPrefix(line, "Distributor ID:") {
			distro = strings.TrimSpace(strings.Split(line, ":")[1])
			if distro == "Fedora" {
				// Consistency with the code parsing /etc/os-release
				distro = "Fedora Linux"
			}
		}
		if strings.HasPrefix(line, "Release:") {
			version = strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return getLinuxDistroFromProcVersion()
	}

	// Return the parsed distribution name and version
	return distro, version
}

// https://www.freedesktop.org/software/systemd/man/latest/os-release.html
// getLinuxInfo reads the OS release information directly from the filesystem.
func getLinuxInfo() (string, string) {
	// Under firejail, access to /etc/os-release is denied.
	data, err := ReadFile("/etc/os-release")
	if err != nil {
		return parseLSBReleaseOrFallback()
	}

	lines := strings.Split(string(data), "\n")
	var name, version string

	for _, line := range lines {
		if strings.HasPrefix(line, "NAME=") {
			name = strings.Trim(strings.SplitN(line, "=", 2)[1], "\"")
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			version = strings.Trim(strings.SplitN(line, "=", 2)[1], "\"")
		} else if strings.HasPrefix(line, "VERSION=") && version == "" {
			version = strings.Trim(strings.SplitN(line, "=", 2)[1], "\"")
		} else if strings.HasPrefix(line, "VERSION_CODENAME=") && version == "" {
			version = strings.Trim(strings.SplitN(line, "=", 2)[1], "\"")
		}
	}

	return name, version
}

// untested!
// getMacOSVersion retrieves the macOS version from the software version command.
func getMacOSVersion() string {
	cmd := Exec("sw_vers", "--productVersion")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	content, err := io.ReadAll(stdout)
	if err == nil {
		return string(bytes.TrimSpace(content))
	}
	return ""
}

// getFreeBSDVersion retrieves the FreeBSD version from the uname command.
func getFreeBSDVersion() string {
	// Attempt to read the version from /etc/freebsd-version
	data, err := ReadFile("/etc/freebsd-version")
	if err == nil {
		return strings.TrimSpace(string(data))
	}

	// Fallback to using freebsd-version command if the file is not available
	output, err := exec.Command("freebsd-version").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
