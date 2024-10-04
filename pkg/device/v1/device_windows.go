//go:build windows
// +build windows

package device

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

func GetDeviceModel() (string, error) {
	// Open the registry key for the computer system
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\BIOS`, registry.READ)
	if err != nil {
		return "", fmt.Errorf("failed to open registry key: %v", err)
	}
	defer k.Close()

	// Read the "SystemBiosVersion" value
	model, _, err := k.GetStringValue("SystemProductName")
	if err != nil {
		return "", fmt.Errorf("failed to get SystemProductName: %v", err)
	}

	return model, nil
}
