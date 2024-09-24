//go:build windows
// +build windows

package device

import (
	"golang.org/x/sys/windows"
	"unsafe"
)

func GetDeviceModel() (string, error) {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	procGetSystemFirmwareTable := kernel32.NewProc("GetSystemFirmwareTable")

	var buf [1024]uint16
	bufSize := uint32(len(buf))

	ret, _, err := procGetSystemFirmwareTable.Call(0, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(bufSize))
	if ret == 0 {
		return "", err
	}

	model := windows.UTF16ToString(buf[:])
	return model, nil
}
