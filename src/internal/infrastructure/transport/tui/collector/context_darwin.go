//go:build darwin

package collector

import (
	"net"
	"os"
	"strings"
)

// getKernelVersion returns the kernel version on macOS.
// Uses /System/Library/CoreServices/SystemVersion.plist or simple fallback.
func getKernelVersion() string {
	// Try to read macOS version from SystemVersion.plist.
	if content, err := os.ReadFile("/System/Library/CoreServices/SystemVersion.plist"); err == nil {
		// Simple parsing for ProductVersion.
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.Contains(line, "ProductVersion") && i+1 < len(lines) {
				ver := strings.TrimSpace(lines[i+1])
				ver = strings.TrimPrefix(ver, "<string>")
				ver = strings.TrimSuffix(ver, "</string>")
				if ver != "" {
					return ver
				}
			}
		}
	}

	// Fallback: just return "macOS".
	return "macOS"
}

// getPrimaryIP returns the primary non-loopback IP address.
func getPrimaryIP() string {
	// Try to get the IP used for outbound connections.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer func() { _ = conn.Close() }()
		if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			return addr.IP.String()
		}
	}

	// Fallback: iterate interfaces.
	interfaces, err := net.Interfaces()
	if err != nil {
		return unknownValue
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces.
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
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Prefer IPv4.
			if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
				return ip.String()
			}
		}
	}

	return unknownValue
}
