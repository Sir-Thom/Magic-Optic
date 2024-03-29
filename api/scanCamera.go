package api

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// VideoDevice represents a video device
type VideoDevice struct {
	Name        string   `json:"name"`
	DevicePaths []string `json:"devicePaths"`
}

func listVideoDevices() ([]VideoDevice, error) {
	cmd := exec.Command("/usr/bin/v4l2-ctl", "--list-devices")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running v4l2-ctl: %v", err)
	}

	// Parse the output and create a list of VideoDevice objects
	devices := parseVideoDeviceOutput(string(output))

	return devices, nil
}

func parseVideoDeviceOutput(output string) []VideoDevice {
	var devices []VideoDevice

	lines := strings.Split(output, "\n")
	var currentDevice VideoDevice

	devPathRegex := regexp.MustCompile(`^\s*/dev/\S+`)

	for _, line := range lines {
		if matches := devPathRegex.FindStringSubmatch(line); len(matches) > 0 {
			// Lines starting with "/dev", add them to the current device

			path := strings.TrimSpace(matches[0])

			currentDevice.DevicePaths = append(currentDevice.DevicePaths, path)
		} else if strings.HasSuffix(line, ":") {
			// New device
			if currentDevice.Name != "" {
				devices = append(devices, currentDevice)
			}
			currentDevice = VideoDevice{Name: strings.TrimSpace(line), DevicePaths: []string{}}
		}
	}

	// Add the last device
	if currentDevice.Name != "" {
		devices = append(devices, currentDevice)
	}

	return devices
}
