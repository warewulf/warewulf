package config

import (
	"fmt"
	"net"
)

func BoolP(p *bool) bool {
	return p != nil && *p
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "192.0.2.1:80")
	if err != nil {
		return nil
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func GetIPNetForIP(ip net.IP) (*net.IPNet, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	for _, iface := range interfaces {
		// skip interfaces that are down or not relevant
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue // try next interface
		}

		for _, addr := range addrs {
			// We expect addr to be of type *net.IPNet.
			var ipNet *net.IPNet
			switch v := addr.(type) {
			case *net.IPNet:
				ipNet = v
			case *net.IPAddr:
				// Wrap net.IPAddr in an IPNet with its default mask.
				ipNet = &net.IPNet{IP: v.IP, Mask: v.IP.DefaultMask()}
			}

			if ipNet == nil {
				continue
			}

			// Check if the IP matches (for IPv4, Equal works well)
			if ipNet.IP.Equal(ip) {
				return ipNet, nil
			}
		}
	}

	return nil, fmt.Errorf("could not find IPNet for IP %v", ip)
}
