//go:build windows

package collector

import (
	"encoding/json"
	"log"
	"os/exec"
)

type powershellEntry struct {
	DisplayName    string `json:"DisplayName"`
	DisplayVersion string `json:"DisplayVersion"`
	Publisher      string `json:"Publisher"`
	InstallDate    string `json:"InstallDate"`
}

func collectSoftware() []Software {
	script := `
$paths = @(
  'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\*',
  'HKLM:\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*'
)
Get-ItemProperty $paths -ErrorAction SilentlyContinue |
  Where-Object { $_.DisplayName } |
  Select-Object DisplayName, DisplayVersion, Publisher, InstallDate |
  ConvertTo-Json -Compress
`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("WARNING: failed to collect installed software: %v", err)
		return nil
	}

	var entries []powershellEntry
	if err := json.Unmarshal(output, &entries); err != nil {
		// PowerShell returns a single object (not array) when there's only one result
		var single powershellEntry
		if err := json.Unmarshal(output, &single); err != nil {
			log.Printf("WARNING: failed to parse software list: %v", err)
			return nil
		}
		entries = []powershellEntry{single}
	}

	software := make([]Software, 0, len(entries))
	for _, e := range entries {
		if e.DisplayName == "" {
			continue
		}
		software = append(software, Software{
			Name:        e.DisplayName,
			Version:     e.DisplayVersion,
			Publisher:   e.Publisher,
			InstalledAt: e.InstallDate,
		})
	}

	return software
}
