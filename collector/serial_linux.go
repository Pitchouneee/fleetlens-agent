//go:build linux

package collector

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

func getSerialNumber() string {
	data, err := os.ReadFile("/sys/class/dmi/id/product_serial")
	if err == nil {
		serial := strings.TrimSpace(string(data))
		if serial != "" {
			return serial
		}
	}

	cmd := exec.Command("dmidecode", "-s", "system-serial-number")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("WARNING: failed to get serial number: %v", err)
		return ""
	}

	return strings.TrimSpace(string(output))
}
