//go:build linux

package collector

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getKernelVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		wantEmpty      bool
		wantUnknown    bool
	}{
		{
			name:        "returns non-empty kernel version on Linux",
			wantEmpty:   false,
			wantUnknown: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			version := getKernelVersion()

			if tt.wantEmpty {
				assert.Empty(t, version, "kernel version should be empty")
			} else {
				assert.NotEmpty(t, version, "kernel version should not be empty on Linux")
			}

			if tt.wantUnknown {
				assert.Equal(t, unknownValue, version, "kernel version should be unknown")
			} else {
				assert.NotEqual(t, unknownValue, version, "kernel version should not be unknown on Linux")
			}
		})
	}
}

func Test_isValidInterface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		iface    net.Interface
		expected bool
	}{
		{
			name: "loopback should be invalid",
			iface: net.Interface{
				Name:  "lo",
				Flags: net.FlagLoopback | net.FlagUp,
			},
			expected: false,
		},
		{
			name: "down interface should be invalid",
			iface: net.Interface{
				Name:  "eth0",
				Flags: 0, // Not up
			},
			expected: false,
		},
		{
			name: "up non-loopback should be valid",
			iface: net.Interface{
				Name:  "eth0",
				Flags: net.FlagUp,
			},
			expected: true,
		},
		{
			name: "up broadcast should be valid",
			iface: net.Interface{
				Name:  "eth0",
				Flags: net.FlagUp | net.FlagBroadcast,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isValidInterface(tt.iface)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// mockAddrser implements Addrser for testing.
type mockAddrser struct {
	addrs []net.Addr
	err   error
}

func (m *mockAddrser) Addrs() ([]net.Addr, error) {
	return m.addrs, m.err
}

func Test_getIPv4FromInterface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		addrser  Addrser
		expected string
	}{
		{
			name: "returns first IPv4",
			addrser: &mockAddrser{
				addrs: []net.Addr{
					&net.IPNet{IP: net.ParseIP("192.168.1.100"), Mask: net.CIDRMask(24, 32)},
				},
			},
			expected: "192.168.1.100",
		},
		{
			name: "skips loopback",
			addrser: &mockAddrser{
				addrs: []net.Addr{
					&net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(8, 32)},
					&net.IPNet{IP: net.ParseIP("192.168.1.100"), Mask: net.CIDRMask(24, 32)},
				},
			},
			expected: "192.168.1.100",
		},
		{
			name: "skips IPv6",
			addrser: &mockAddrser{
				addrs: []net.Addr{
					&net.IPNet{IP: net.ParseIP("fe80::1"), Mask: net.CIDRMask(64, 128)},
					&net.IPNet{IP: net.ParseIP("192.168.1.100"), Mask: net.CIDRMask(24, 32)},
				},
			},
			expected: "192.168.1.100",
		},
		{
			name: "handles IPAddr type",
			addrser: &mockAddrser{
				addrs: []net.Addr{
					&net.IPAddr{IP: net.ParseIP("192.168.1.100")},
				},
			},
			expected: "192.168.1.100",
		},
		{
			name: "returns empty when no IPv4",
			addrser: &mockAddrser{
				addrs: []net.Addr{
					&net.IPNet{IP: net.ParseIP("fe80::1"), Mask: net.CIDRMask(64, 128)},
				},
			},
			expected: "",
		},
		{
			name: "returns empty on error",
			addrser: &mockAddrser{
				err: assert.AnError,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := getIPv4FromInterface(tt.addrser)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_getPrimaryIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		wantEmpty bool
	}{
		{
			name:      "returns non-empty IP or unknown",
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ip := getPrimaryIP()

			if tt.wantEmpty {
				assert.Empty(t, ip, "primary IP should be empty")
				return
			}

			assert.NotEmpty(t, ip, "primary IP should not be empty")

			// Validate IP format if not unknown.
			if ip != unknownValue {
				parsedIP := net.ParseIP(ip)
				assert.NotNil(t, parsedIP, "primary IP should be valid IP format")
			}
		})
	}
}

func Test_getOutboundIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "returns empty or valid non-loopback IP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ip := getOutboundIP()

			// If not empty, should be valid IP.
			if ip != "" {
				parsedIP := net.ParseIP(ip)
				assert.NotNil(t, parsedIP, "outbound IP should be valid IP format")
				assert.False(t, parsedIP.IsLoopback(), "outbound IP should not be loopback")
			}
		})
	}
}

// Test_getIPFromInterfaces tests the getIPFromInterfaces function.
// It verifies that IP address retrieval from interfaces works correctly.
//
// Params:
//   - t: the testing context.
func Test_getIPFromInterfaces(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "returns_ip_or_unknown",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call function.
			ip := getIPFromInterfaces()

			// Result should be non-empty.
			assert.NotEmpty(t, ip)

			// If not unknown, should be valid IP format.
			if ip != unknownValue {
				parsedIP := net.ParseIP(ip)
				assert.NotNil(t, parsedIP, "IP should be valid format")
			}
		})
	}
}
