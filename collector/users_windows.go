//go:build windows

package collector

import (
	"encoding/json"
	"log"
	"os/exec"
	"strings"
)

type powershellUser struct {
	Name      string `json:"Name"`
	FullName  string `json:"FullName"`
	LastLogon string `json:"LastLogon"`
}

func collectUsers() []User {
	script := `
$admins = (Get-LocalGroupMember -Group "Administrators" -ErrorAction SilentlyContinue).Name | ForEach-Object { ($_ -split '\\')[-1] }
Get-LocalUser | Where-Object { $_.Enabled } | ForEach-Object {
  [PSCustomObject]@{
    Name      = $_.Name
    FullName  = $_.FullName
    IsAdmin   = ($admins -contains $_.Name)
    LastLogon = if ($_.LastLogon) { $_.LastLogon.ToString("yyyy-MM-dd HH:mm:ss") } else { "" }
  }
} | ConvertTo-Json -Compress
`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("WARNING: failed to collect users: %v", err)
		return nil
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return nil
	}

	type psUser struct {
		Name      string `json:"Name"`
		FullName  string `json:"FullName"`
		IsAdmin   bool   `json:"IsAdmin"`
		LastLogon string `json:"LastLogon"`
	}

	var entries []psUser
	if err := json.Unmarshal([]byte(trimmed), &entries); err != nil {
		var single psUser
		if err := json.Unmarshal([]byte(trimmed), &single); err != nil {
			log.Printf("WARNING: failed to parse users: %v", err)
			return nil
		}
		entries = []psUser{single}
	}

	users := make([]User, 0, len(entries))
	for _, e := range entries {
		users = append(users, User{
			Username:    e.Name,
			DisplayName: e.FullName,
			IsAdmin:     e.IsAdmin,
			LastLogin:   e.LastLogon,
		})
	}

	return users
}
