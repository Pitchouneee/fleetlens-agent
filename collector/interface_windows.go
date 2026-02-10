//go:build windows

package collector

import (
	"net"
	"strings"
)

func getInterfaceType(iface net.Interface) string {
	if iface.Flags&net.FlagLoopback != 0 {
		return "loopback"
	}

	name := strings.ToLower(iface.Name)

	switch {
	case strings.Contains(name, "wi-fi"),
		strings.Contains(name, "wifi"),
		strings.Contains(name, "wireless"),
		strings.Contains(name, "wlan"):
		return "wifi"
	case strings.Contains(name, "vethernet"),
		strings.Contains(name, "vmware"),
		strings.Contains(name, "virtualbox"),
		strings.Contains(name, "hyper-v"),
		strings.Contains(name, "vbox"),
		strings.Contains(name, "docker"),
		strings.Contains(name, "wsl"):
		return "virtual"
	default:
		return "ethernet"
	}
}
