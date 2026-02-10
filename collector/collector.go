package collector

import (
	"log"
	"net"
	"os"
	"runtime"
)

type NetworkInterface struct {
	Name       string   `json:"name"`
	IPAdresses []string `json:"ip_addresses"`
	MACAddress string   `json:"mac_address"`
	Type       string   `json:"type"`
}

type SystemInfo struct {
	Hostname          string             `json:"hostname"`
	IPAddresses       []string           `json:"ip_addresses"`
	SerialNumber      string             `json:"serial_number"`
	OperatingSystem   string             `json:"operating_system"`
	Architecture      string             `json:"architecture"`
	NetworkInterfaces []NetworkInterface `json:"network_interfaces"`
}

func Collect() SystemInfo {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("WARNING: failed to get hostname: %v", err)
	}

	return SystemInfo{
		Hostname:          hostname,
		IPAddresses:       collectIPAddresses(),
		SerialNumber:      getSerialNumber(),
		OperatingSystem:   runtime.GOOS,
		Architecture:      runtime.GOARCH,
		NetworkInterfaces: collectNetworkInterfaces(),
	}
}

func collectNetworkInterfaces() []NetworkInterface {
	var interfaces []NetworkInterface

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("WARNING: failed to get network interfaces: %v", err)
		return interfaces
	}

	for _, iface := range ifaces {
		var ips []string
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok {
					ips = append(ips, ipNet.IP.String())
				}
			}
		}

		interfaces = append(interfaces, NetworkInterface{
			Name:       iface.Name,
			IPAdresses: ips,
			MACAddress: iface.HardwareAddr.String(),
			Type:       getInterfaceType(iface),
		})
	}

	return interfaces
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
