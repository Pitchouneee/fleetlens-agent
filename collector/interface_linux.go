//go:build linux

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
	case strings.HasPrefix(name, "wl"),
		strings.HasPrefix(name, "wlan"):
		return "wifi"
	case strings.HasPrefix(name, "veth"),
		strings.HasPrefix(name, "docker"),
		strings.HasPrefix(name, "br-"),
		strings.HasPrefix(name, "virbr"),
		strings.HasPrefix(name, "vbox"),
		strings.HasPrefix(name, "vmnet"),
		strings.HasPrefix(name, "tun"),
		strings.HasPrefix(name, "tap"):
		return "virtual"
	default:
		return "ethernet"
	}
}
