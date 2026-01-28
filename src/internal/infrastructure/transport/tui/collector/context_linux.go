//go:build linux

// Package collector provides system metrics collection.
package collector

import (
	"context"
	"net"
	"syscall"
	"time"
)

const (
	// unameReleaseBufferSize is the buffer size for uname Release field.
	unameReleaseBufferSize int = 65
)

// getKernelVersion returns the kernel version on Linux.
//
// Returns:
//   - string: kernel version string or "unknown" on error
func getKernelVersion() string {
	var uname syscall.Utsname
	// Try to get uname info.
	if err := syscall.Uname(&uname); err != nil {
		// Syscall failed.
		return unknownValue
	}

	// Convert [65]int8 to string.
	release := make([]byte, 0, unameReleaseBufferSize)
	// Iterate through release bytes.
	for _, b := range uname.Release {
		// Check for null terminator.
		if b == 0 {
			// End of string.
			break
		}
		release = append(release, byte(b))
	}

	// Return kernel version string.
	return string(release)
}

// getPrimaryIP returns the primary non-loopback IP address.
//
// Returns:
//   - string: primary IP address or "unknown" if not found
func getPrimaryIP() string {
	// Try to get the IP used for outbound connections.
	if ip := getOutboundIP(); ip != "" {
		// Outbound IP found.
		return ip
	}

	// Fallback: iterate interfaces.
	return getIPFromInterfaces()
}

// getOutboundIP returns the local IP used for outbound connections.
//
// Returns:
//   - string: outbound IP address or empty string if not found
func getOutboundIP() string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "udp", "8.8.8.8:80")
	// Check if connection succeeded.
	if err != nil {
		// Cannot establish outbound connection.
		return ""
	}
	defer func() { _ = conn.Close() }()

	// Try to get local address.
	if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		// Return outbound IP.
		return addr.IP.String()
	}

	// Could not determine outbound IP.
	return ""
}

// getIPFromInterfaces returns the first non-loopback IPv4 from interfaces.
//
// Returns:
//   - string: IP address or "unknown" if not found
func getIPFromInterfaces() string {
	interfaces, err := net.Interfaces()
	// Handle interface listing error.
	if err != nil {
		// Cannot list interfaces.
		return unknownValue
	}

	// Process each interface.
	for _, iface := range interfaces {
		// Skip loopback and down interfaces.
		if !isValidInterface(iface) {
			// Skip invalid interface.
			continue
		}

		// Try to get IPv4 from this interface.
		if ip := getIPv4FromInterface(&iface); ip != "" {
			// Found valid IPv4.
			return ip
		}
	}

	// No suitable IP found.
	return unknownValue
}

// isValidInterface checks if interface is suitable for IP discovery.
//
// Params:
//   - iface: network interface to check
//
// Returns:
//   - bool: true if interface is up and not loopback
func isValidInterface(iface net.Interface) bool {
	// Check for loopback flag.
	if iface.Flags&net.FlagLoopback != 0 {
		// Skip loopback.
		return false
	}
	// Check for up flag.
	if iface.Flags&net.FlagUp == 0 {
		// Skip down interface.
		return false
	}
	// Interface is valid.
	return true
}

// getIPv4FromInterface extracts the first IPv4 address from an interface.
//
// Params:
//   - iface: network interface to scan (uses Addrser interface)
//
// Returns:
//   - string: IPv4 address or empty string if not found
func getIPv4FromInterface(iface Addrser) string {
	addrs, err := iface.Addrs()
	// Handle address retrieval error.
	if err != nil {
		// Skip interface on error.
		return ""
	}

	// Process each address.
	for _, addr := range addrs {
		var ip net.IP
		// Extract IP from address type.
		switch val := addr.(type) {
		// Handle IPNet type.
		case *net.IPNet:
			ip = val.IP
		// Handle IPAddr type.
		case *net.IPAddr:
			ip = val.IP
		}

		// Prefer IPv4.
		// Check for valid non-loopback IPv4.
		if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
			// Found valid IPv4.
			return ip.String()
		}
	}

	// No IPv4 found on this interface.
	return ""
}
