//go:build darwin

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
	case strings.HasPrefix(name, "awdl"),
		strings.HasPrefix(name, "wlan"):
		return "wifi"
	case strings.HasPrefix(name, "utun"),
		strings.HasPrefix(name, "bridge"),
		strings.HasPrefix(name, "vmnet"),
		strings.HasPrefix(name, "vbox"),
		strings.HasPrefix(name, "docker"):
		return "virtual"
	case name == "en0":
		return "wifi"
	default:
		return "ethernet"
	}
}
