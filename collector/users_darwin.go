//go:build darwin

package collector

import (
	"log"
	"os/exec"
	"strings"
)

func collectUsers() []User {
	cmd := exec.Command("dscl", ".", "-list", "/Users")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("WARNING: failed to list users: %v", err)
		return nil
	}

	admins := getDarwinAdmins()
	lastLogins := getDarwinLastLogins()

	var users []User
	for _, username := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		username = strings.TrimSpace(username)
		if username == "" || strings.HasPrefix(username, "_") {
			continue
		}

		displayName := dsclRead(username, "RealName")
		if displayName == "" {
			displayName = username
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

func dsclRead(username, key string) string {
	cmd := exec.Command("dscl", ".", "-read", "/Users/"+username, key)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Output format: "Key: Value" or multiline with value on next line
	lines := strings.SplitN(string(output), "\n", 3)
	if len(lines) >= 1 {
		parts := strings.SplitN(lines[0], ":", 2)
		if len(parts) == 2 {
			val := strings.TrimSpace(parts[1])
			if val != "" {
				return val
			}
		}
		// Value may be on the next line
		if len(lines) >= 2 {
			return strings.TrimSpace(lines[1])
		}
	}

	return ""
}

func getDarwinAdmins() map[string]bool {
	admins := make(map[string]bool)

	cmd := exec.Command("dscl", ".", "-read", "/Groups/admin", "GroupMembership")
	output, err := cmd.Output()
	if err != nil {
		return admins
	}

	parts := strings.SplitN(string(output), ":", 2)
	if len(parts) == 2 {
		for _, user := range strings.Fields(parts[1]) {
			admins[user] = true
		}
	}

	return admins
}

func getDarwinLastLogins() map[string]string {
	logins := make(map[string]string)

	cmd := exec.Command("last")
	output, err := cmd.Output()
	if err != nil {
		return logins
	}

	seen := make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		username := fields[0]
		if seen[username] || username == "reboot" || username == "shutdown" {
			continue
		}
		seen[username] = true
		// `last` output: user tty host date-fields
		logins[username] = strings.Join(fields[2:], " ")
	}

	return logins
}
