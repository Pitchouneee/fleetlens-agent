//go:build darwin

package collector

import (
	"log"
	"os/exec"
	"strings"
)

func getSerialNumber() string {
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("WARNING: failed to get serial number: %v", err)
		return ""
	}

	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "IOPlatformSerialNumber") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				serial := strings.TrimSpace(parts[1])
				serial = strings.Trim(serial, "\"")
				return serial
			}
		}
	}

	log.Printf("WARNING: serial number not found in ioreg output")
	return ""
}
