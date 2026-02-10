//go:build linux

package collector

import (
	"log"
	"os/exec"
	"strings"
)

func collectSoftware() []Software {
	// Try dpkg first (Debian/Ubuntu), then rpm (RHEL/Fedora/SUSE)
	if software := collectDpkg(); software != nil {
		return software
	}
	if software := collectRpm(); software != nil {
		return software
	}

	log.Printf("WARNING: no supported package manager found (tried dpkg, rpm)")
	return nil
}

func collectDpkg() []Software {
	cmd := exec.Command("dpkg-query", "-W", "-f=${Package}\\t${Version}\\t${Maintainer}\\n")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var software []Software
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.SplitN(line, "\t", 3)
		if len(fields) < 2 || fields[0] == "" {
			continue
		}

		s := Software{
			Name:    fields[0],
			Version: fields[1],
		}
		if len(fields) == 3 {
			s.Publisher = fields[2]
		}
		software = append(software, s)
	}

	return software
}

func collectRpm() []Software {
	cmd := exec.Command("rpm", "-qa", "--queryformat", "%{NAME}\\t%{VERSION}\\t%{VENDOR}\\t%{INSTALLTIME:date}\\n")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var software []Software
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.SplitN(line, "\t", 4)
		if len(fields) < 2 || fields[0] == "" {
			continue
		}

		s := Software{
			Name:    fields[0],
			Version: fields[1],
		}
		if len(fields) >= 3 {
			s.Publisher = fields[2]
		}
		if len(fields) >= 4 {
			s.InstalledAt = fields[3]
		}
		software = append(software, s)
	}

	return software
}
