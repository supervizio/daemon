//go:build linux

package collector

import (
	"context"
	"net"
	"syscall"
	"time"
)

// getKernelVersion returns the kernel version on Linux.
func getKernelVersion() string {
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return unknownValue
	}

	// Convert [65]int8 to string.
	release := make([]byte, 0, 65)
	for _, b := range uname.Release {
		if b == 0 {
			break
		}
		release = append(release, byte(b))
	}

	return string(release)
}

// getPrimaryIP returns the primary non-loopback IP address.
func getPrimaryIP() string {
	// Try to get the IP used for outbound connections.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(ctx, "udp", "8.8.8.8:80")
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
