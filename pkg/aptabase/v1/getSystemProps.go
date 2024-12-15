package aptabase

import (
	_ "embed"
	"fmt"
	"github.com/brycensranch/go-aptabase/pkg/device/v1"
	"github.com/brycensranch/go-aptabase/pkg/locale"
	"github.com/brycensranch/go-aptabase/pkg/osinfo/v1"
	"runtime"
)

//go:embed VERSION

var SDKVersion string // This will hold the version at build time

// GetVersion retrieves the SDK version for use in the application.
func GetVersion() string {
	if SDKVersion == "" {
		return "unknown" // In case SDKVersion is not set
	}
	return SDKVersion
}

// systemProps retrieves system information using the osinfo package,
// and includes Client-specific details like AppVersion, AppBuildNumber, and DebugMode.
func (c *Client) systemProps() (map[string]interface{}, error) {
	osName, osVersion := osinfo.GetOSInfo()
	deviceModel, err := device.GetDeviceModel()
	if err != nil && c.DebugMode {
		c.Logger.Printf("WARNING got error trying to get device model: %v", err)
	}

	props := map[string]interface{}{
		"isDebug":        c.DebugMode,
		"osName":         osName,
		"osVersion":      osVersion,
		"engineName":     "go",
		"engineVersion":  runtime.Version(),
		"locale":         locale.GetLocale(),
		"appVersion":     c.AppVersion,
		"appBuildNumber": fmt.Sprintf("%v", c.AppBuildNumber),
		"deviceModel":    deviceModel,
		"sdkVersion":     fmt.Sprintf("go-aptabase@%s", GetVersion()),
	}
	if c.DebugMode {
		c.Logger.Printf("systemProps: %v", props)
	}

	return props, nil
}
