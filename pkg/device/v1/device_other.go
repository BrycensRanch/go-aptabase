//go:build !windows
// +build !windows

package device

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"
)

// DMI structure definitions based on the DMI specification
const (
	DMI_TYPE_SYSTEM = 1
)

type DmiHeader struct {
	Type   uint8
	Length uint8
	Handle uint16
}

type DmiSystem struct {
	Manufacturer string
	ProductName  string
}

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
	return "Unknown", nil
}
func getLinuxDeviceModel() (string, error) {
	data, err := ioutil.ReadFile("/sys/class/dmi/id/product_name")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func getFreeBSDDeviceModel() (string, error) {
	data, err := ioutil.ReadFile("/sys/firmware/dmi/tables/DMI")
	if err != nil {
		return "", err
	}

	// Parse the binary data
	var productName string

	offset := 0
	for offset < len(data) {
		header := DmiHeader{
			Type:   data[offset],
			Length: data[offset+1],
			Handle: binary.BigEndian.Uint16(data[offset+2 : offset+4]),
		}

		// If this is the system type, extract the product name
		if header.Type == DMI_TYPE_SYSTEM {
			// Read manufacturer and product name
			start := offset + 5                   // Start after header
			end := start + int(header.Length) - 5 // Exclude header

			if end <= len(data) {
				info := data[start:end]
				// Convert to string and split by null character
				parts := strings.Split(string(info), "\x00")
				if len(parts) > 1 {
					productName = parts[1] // Product name is the second part
					break
				}
			}
		}
		offset += int(header.Length)
	}

	if productName == "" {
		return "Unknown Device", nil
	}
	return productName, nil
}

func main() {
	println(GetDeviceModel())
}
