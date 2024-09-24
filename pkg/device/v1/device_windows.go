//go:build windows
// +build windows

package device

import (
	"syscall"
)

func GetDeviceModel() (string, error) {
	var buf [1024]uint16
	var size uint32
	err := syscall.GetSystemFirmwareTable(0, 0, &buf[0], uint32(len(buf)))
	if err != nil {
		return "", err
	}
	model := syscall.UTF16ToString(buf[:])
	return model, nil
}
