package collector

import (
	"log"
	"net"
	"os"
	"runtime"
)

type SystemInfo struct {
	Hostname        string   `json:"hostname"`
	IPAddresses     []string `json:"ip_addresses"`
	SerialNumber    string   `json:"serial_number"`
	OperatingSystem string   `json:"operating_system"`
	Architecture    string   `json:"architecture"`
}

func Collect() SystemInfo {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("WARNING: failed to get hostname: %v", err)
	}

	return SystemInfo{
		Hostname:        hostname,
		IPAddresses:     collectIPAddresses(),
		SerialNumber:    getSerialNumber(),
		OperatingSystem: runtime.GOOS,
		Architecture:    runtime.GOARCH,
	}
}

func collectIPAddresses() []string {
	var ips []string

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("WARNING: failed to get network interfaces: %v", err)
		return ips
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}

	return ips
}
