//go:build linux

package collector

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strings"
)

func collectUsers() []User {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		log.Printf("WARNING: failed to read /etc/passwd: %v", err)
		return nil
	}
	defer f.Close()

	admins := getAdminUsers()
	lastLogins := getLastLogins()

	var users []User
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) < 7 {
			continue
		}

		username := fields[0]
		gecos := fields[4]
		shell := fields[6]

		// Skip system accounts (nologin/false shells)
		if strings.Contains(shell, "nologin") || strings.Contains(shell, "/false") {
			continue
		}

		displayName := username
		if gecos != "" {
			// GECOS field: full name is the first comma-separated value
			displayName = strings.SplitN(gecos, ",", 2)[0]
		}

		users = append(users, User{
			Username:    username,
			DisplayName: displayName,
			IsAdmin:     admins[username],
			LastLogin:   lastLogins[username],
		})
	}

	return users
}

func getAdminUsers() map[string]bool {
	admins := make(map[string]bool)

	f, err := os.Open("/etc/group")
	if err != nil {
		return admins
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) < 4 {
			continue
		}
		group := fields[0]
		if group != "sudo" && group != "wheel" {
			continue
		}
		for _, user := range strings.Split(fields[3], ",") {
			user = strings.TrimSpace(user)
			if user != "" {
				admins[user] = true
			}
		}
	}

	// Also check root
	admins["root"] = true

	return admins
}

func getLastLogins() map[string]string {
	logins := make(map[string]string)

	cmd := exec.Command("lastlog")
	output, err := cmd.Output()
	if err != nil {
		return logins
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		username := fields[0]
		if len(fields) >= 5 && fields[1] != "**Never" {
			// Format: Username Port From Latest
			logins[username] = strings.Join(fields[3:], " ")
		}
	}

	return logins
}
