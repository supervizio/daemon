//go:build freebsd || openbsd || netbsd || dragonfly

package collector

import (
	"net"
	"os"
	"strings"
)

// getKernelVersion returns the kernel version on BSD.
// Uses /etc/os-release or falls back to a simple version string.
func getKernelVersion() string {
	// Try to read from /etc/os-release.
	if content, err := os.ReadFile("/etc/os-release"); err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			if strings.HasPrefix(line, "VERSION_ID=") {
				ver := strings.TrimPrefix(line, "VERSION_ID=")
				ver = strings.Trim(ver, "\"")
				if ver != "" {
					return ver
				}
			}
		}
	}

	// Fallback: just return "BSD".
	return "BSD"
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
