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
		// Parse each line looking for VERSION_ID.
		for _, line := range strings.Split(string(content), "\n") {
			// Check for VERSION_ID prefix.
			if strings.HasPrefix(line, "VERSION_ID=") {
				ver := strings.TrimPrefix(line, "VERSION_ID=")
				ver = strings.Trim(ver, "\"")
				// Validate version string.
				if ver != "" {
					// Return parsed version.
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
	// Check if connection succeeded.
	if err == nil {
		defer func() { _ = conn.Close() }()
		// Try to get local address.
		if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			// Return outbound IP.
			return addr.IP.String()
		}
	}

	// Fallback: iterate interfaces.
	interfaces, err := net.Interfaces()
	// Handle interface listing error.
	if err != nil {
		// Cannot list interfaces.
		return unknownValue
	}

	// Process each interface.
	for _, iface := range interfaces {
		// Skip loopback and down interfaces.
		// Check for loopback flag.
		if iface.Flags&net.FlagLoopback != 0 {
			// Skip loopback.
			continue
		}
		// Check for up flag.
		if iface.Flags&net.FlagUp == 0 {
			// Skip down interface.
			continue
		}

		addrs, err := iface.Addrs()
		// Handle address retrieval error.
		if err != nil {
			// Skip interface on error.
			continue
		}

		// Process each address.
		for _, addr := range addrs {
			var ip net.IP
			// Extract IP from address type.
			switch v := addr.(type) {
			// Handle IPNet type.
			case *net.IPNet:
				ip = v.IP
			// Handle IPAddr type.
			case *net.IPAddr:
				ip = v.IP
			}

			// Prefer IPv4.
			// Check for valid non-loopback IPv4.
			if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
				// Found valid IPv4.
				return ip.String()
			}
		}
	}

	// No suitable IP found.
	return unknownValue
}
