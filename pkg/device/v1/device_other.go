//go:build !windows
// +build !windows

package device

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

func GetDeviceModel() (string, error) {
	switch runtime.GOOS {
	case "darwin": // macOS
		return getMacDeviceModel()
	case "linux":
		return getFreeBSDDeviceModel()
	case "freebsd":
		return getFreeBSDDeviceModel()
	default:
		return "Unknown Device", fmt.Errorf("Unsupported Platform, running on %s", runtime.GOOS)
	}
}

func getMacDeviceModel() (string, error) {
	// Run the sysctl command to get the hardware model
	cmd := exec.Command("sysctl", "-n", "hw.model")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Printf("Error running sysctl: %v", err)
		return "", err
	}

	// Process the output, trim it to clean up extra newlines
	modelIdentifier := strings.TrimSpace(out.String())
	if modelIdentifier == "" {
		return "", nil
	}
	return modelIdentifier, nil
}
func getLinuxDeviceModel() (string, error) {
	data, err := ioutil.ReadFile("/sys/class/dmi/id/product_name")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func getFreeBSDDeviceModel() (string, error) {
	// Run the kenv command to get the smbios.system.product value
	cmd := exec.Command("kenv", "smbios.system.product")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Printf("Error running kenv: %v", err)
		return "", err
	}

	// Process the output, trim it to clean up extra newlines
	deviceModel := strings.TrimSpace(out.String())
	if deviceModel == "" {
		return "", nil
	}
	return deviceModel, nil
}

func main() {
	println(GetDeviceModel())
}
