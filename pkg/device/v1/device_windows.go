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

	// Step 1: First call to get the size
	ret, _, err := procGetSystemFirmwareTable.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("ACPI"))), // Example provider signature
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("SMBIOS"))), // Example table ID
		0,
		0,
		0,
	)

	if ret == 0 {
		fmt.Printf("Error getting table size: %v\n", err)
		return
	}

	// Step 2: Prepare buffer based on returned size
	bufSize := uint32(ret)
	buf := make([]uint16, bufSize/2) // Each UTF16 character is 2 bytes

	// Step 3: Call again to get the actual data
	ret, _, err = procGetSystemFirmwareTable.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("ACPI"))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("SMBIOS"))),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(bufSize),
		0,
	)

	if ret == 0 {
		fmt.Printf("Error getting firmware table: %v\n", err)
		return "", err
	}

	// Convert the buffer to a string
	model := windows.UTF16ToString(buf)
	fmt.Println("Model:", model)
	return model, nil
}
