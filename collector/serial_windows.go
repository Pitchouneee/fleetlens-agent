//go:build windows

package collector

import (
	"log"
	"os/exec"
	"strings"
)

func getSerialNumber() string {
	cmd := exec.Command("wmic", "bios", "get", "serialnumber")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("WARNING: failed to get serial number: %v", err)
		return ""
	}

	lines := strings.Split(strings.ReplaceAll(string(output), "\r", ""), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.EqualFold(trimmed, "SerialNumber") {
			return trimmed
		}
	}

	return ""
}
