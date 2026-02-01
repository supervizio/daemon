//go:build linux

package discovery

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseHexPort verifies hex port parsing.
func TestParseHexPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{
			name:     "port 80",
			input:    "0050",
			expected: 80,
			wantErr:  false,
		},
		{
			name:     "port 443",
			input:    "01BB",
			expected: 443,
			wantErr:  false,
		},
		{
			name:     "port 22",
			input:    "0016",
			expected: 22,
			wantErr:  false,
		},
		{
			name:    "invalid hex",
			input:   "ZZZZ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := parseHexPort(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestParseIPv4Bytes verifies IPv4 byte parsing.
func TestParseIPv4Bytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []byte
		expected string
		wantErr  bool
	}{
		{
			name:     "localhost reversed",
			input:    []byte{127, 0, 0, 1},
			expected: "1.0.0.127",
			wantErr:  false,
		},
		{
			name:    "wrong length",
			input:   []byte{127, 0, 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := parseIPv4Bytes(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestParseIPv6Bytes verifies IPv6 byte parsing.
func TestParseIPv6Bytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "valid length",
			input:   make([]byte, 16),
			wantErr: false,
		},
		{
			name:    "wrong length",
			input:   []byte{0, 0, 0, 0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := parseIPv6Bytes(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestPortScanDiscoverer_shouldIncludePort verifies port filtering logic.
func TestPortScanDiscoverer_shouldIncludePort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		includePorts []int
		excludePorts []int
		port         int
		expected     bool
	}{
		{
			name:         "include specific port",
			includePorts: []int{80, 443},
			port:         80,
			expected:     true,
		},
		{
			name:         "exclude specific port",
			excludePorts: []int{22},
			port:         22,
			expected:     false,
		},
		{
			name:         "port not in include list",
			includePorts: []int{80, 443},
			port:         22,
			expected:     false,
		},
		{
			name:         "port not in exclude list",
			excludePorts: []int{22},
			port:         80,
			expected:     true,
		},
		{
			name:     "no filters",
			port:     8080,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.PortScanDiscoveryConfig{
				IncludePorts: tt.includePorts,
				ExcludePorts: tt.excludePorts,
			}
			discoverer := NewPortScanDiscoverer(cfg)

			result := discoverer.shouldIncludePort(tt.port)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// mockAddrser is a mock implementation of the Addrser interface.
type mockAddrser struct {
	addrs []net.Addr
	err   error
}

// Addrs returns the mock addresses.
func (m *mockAddrser) Addrs() ([]net.Addr, error) {
	return m.addrs, m.err
}

// TestPortScanDiscoverer_interfaceContainsIP verifies interface IP checking.
func TestPortScanDiscoverer_interfaceContainsIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		iface     Addrser
		targetIP  string
		wantMatch bool
	}{
		{
			name:      "returns false when address listing fails",
			iface:     &mockAddrser{err: assert.AnError},
			targetIP:  "127.0.0.1",
			wantMatch: false,
		},
		{
			name:      "returns false when no addresses found",
			iface:     &mockAddrser{addrs: []net.Addr{}},
			targetIP:  "127.0.0.1",
			wantMatch: false,
		},
		{
			name: "returns true when IP matches",
			iface: &mockAddrser{
				addrs: []net.Addr{
					&net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(8, 32)},
				},
			},
			targetIP:  "127.0.0.1",
			wantMatch: true,
		},
		{
			name: "returns false when IP does not match",
			iface: &mockAddrser{
				addrs: []net.Addr{
					&net.IPNet{IP: net.ParseIP("192.168.1.1"), Mask: net.CIDRMask(24, 32)},
				},
			},
			targetIP:  "127.0.0.1",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.PortScanDiscoveryConfig{}
			discoverer := NewPortScanDiscoverer(cfg)

			result := discoverer.interfaceContainsIP(tt.iface, net.ParseIP(tt.targetIP))

			assert.Equal(t, tt.wantMatch, result)
		})
	}
}

// TestPortScanDiscoverer_parseNetTCPLine verifies single line parsing.
func TestPortScanDiscoverer_parseNetTCPLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		line     string
		protocol string
		wantOK   bool
		wantPort int
	}{
		{
			name:     "empty line returns false",
			line:     "",
			protocol: "tcp4",
			wantOK:   false,
		},
		{
			name:     "whitespace only line returns false",
			line:     "   ",
			protocol: "tcp4",
			wantOK:   false,
		},
		{
			name:     "line with insufficient fields returns false",
			line:     "0: 00000000:0016",
			protocol: "tcp4",
			wantOK:   false,
		},
		{
			name:     "non-listening state returns false",
			line:     "   0: 00000000:0050 00000000:0000 01 00000000:00000000 00:00000000 00000000     0        0 12345 1",
			protocol: "tcp4",
			wantOK:   false,
		},
		{
			name:     "listening state returns true",
			line:     "   0: 00000000:0050 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 12345 1",
			protocol: "tcp4",
			wantOK:   true,
			wantPort: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.PortScanDiscoveryConfig{}
			discoverer := NewPortScanDiscoverer(cfg)

			port, ok := discoverer.parseNetTCPLine(tt.line, tt.protocol)

			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantPort, port.LocalPort)
			}
		})
	}
}

// TestPortScanDiscoverer_parseAddress verifies hex address parsing.
func TestPortScanDiscoverer_parseAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hexAddr  string
		protocol string
		wantIP   string
		wantPort int
		wantErr  bool
	}{
		{
			name:     "valid IPv4 localhost port 80",
			hexAddr:  "0100007F:0050",
			protocol: "tcp4",
			wantIP:   "127.0.0.1",
			wantPort: 80,
			wantErr:  false,
		},
		{
			name:     "valid IPv4 port 443",
			hexAddr:  "00000000:01BB",
			protocol: "tcp4",
			wantIP:   "0.0.0.0",
			wantPort: 443,
			wantErr:  false,
		},
		{
			name:     "invalid format no colon",
			hexAddr:  "01000000050",
			protocol: "tcp4",
			wantErr:  true,
		},
		{
			name:     "invalid format multiple colons",
			hexAddr:  "01:00:00:7F:0050",
			protocol: "tcp4",
			wantErr:  true,
		},
		{
			name:     "invalid port hex",
			hexAddr:  "0100007F:ZZZZ",
			protocol: "tcp4",
			wantErr:  true,
		},
		{
			name:     "invalid IP hex",
			hexAddr:  "ZZZZZZZZ:0050",
			protocol: "tcp4",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.PortScanDiscoveryConfig{}
			discoverer := NewPortScanDiscoverer(cfg)

			ip, port, err := discoverer.parseAddress(tt.hexAddr, tt.protocol)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantIP, ip)
				assert.Equal(t, tt.wantPort, port)
			}
		})
	}
}

// Test_parseHexIP verifies hex IP parsing for both IPv4 and IPv6.
func Test_parseHexIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ipHex    string
		protocol string
		wantIP   string
		wantErr  bool
	}{
		{
			name:     "valid IPv4 localhost",
			ipHex:    "0100007F",
			protocol: "tcp4",
			wantIP:   "127.0.0.1",
			wantErr:  false,
		},
		{
			name:     "valid IPv4 all interfaces",
			ipHex:    "00000000",
			protocol: "tcp4",
			wantIP:   "0.0.0.0",
			wantErr:  false,
		},
		{
			name:     "valid IPv6 localhost",
			ipHex:    "00000000000000000000000001000000",
			protocol: "tcp6",
			wantIP:   "::100:0",
			wantErr:  false,
		},
		{
			name:     "invalid hex string",
			ipHex:    "ZZZZZZZZ",
			protocol: "tcp4",
			wantErr:  true,
		},
		{
			name:     "wrong length for IPv4",
			ipHex:    "0100",
			protocol: "tcp4",
			wantErr:  true,
		},
		{
			name:     "wrong length for IPv6",
			ipHex:    "00000000",
			protocol: "tcp6",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ip, err := parseHexIP(tt.ipHex, tt.protocol)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantIP, ip)
			}
		})
	}
}

// TestPortScanDiscoverer_collectAllListeningPorts verifies collection from proc files.
func TestPortScanDiscoverer_collectAllListeningPorts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "collects ports from proc files",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.PortScanDiscoveryConfig{}
			discoverer := NewPortScanDiscoverer(cfg)

			ports, err := discoverer.collectAllListeningPorts()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// May return empty or populated, but should not error.
				assert.NoError(t, err)
				assert.NotNil(t, ports)
			}
		})
	}
}

// TestPortScanDiscoverer_filterAndConvertPorts verifies port filtering and target conversion.
func TestPortScanDiscoverer_filterAndConvertPorts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		ports        []listeningPort
		includePorts []int
		excludePorts []int
		wantCount    int
	}{
		{
			name:      "empty ports returns empty",
			ports:     []listeningPort{},
			wantCount: 0,
		},
		{
			name: "filters duplicate ports",
			ports: []listeningPort{
				{Protocol: "tcp4", LocalAddr: "0.0.0.0", LocalPort: 80},
				{Protocol: "tcp4", LocalAddr: "0.0.0.0", LocalPort: 80},
			},
			wantCount: 1,
		},
		{
			name: "excludes ports in exclude list",
			ports: []listeningPort{
				{Protocol: "tcp4", LocalAddr: "0.0.0.0", LocalPort: 22},
				{Protocol: "tcp4", LocalAddr: "0.0.0.0", LocalPort: 80},
			},
			excludePorts: []int{22},
			wantCount:    1,
		},
		{
			name: "includes only ports in include list",
			ports: []listeningPort{
				{Protocol: "tcp4", LocalAddr: "0.0.0.0", LocalPort: 22},
				{Protocol: "tcp4", LocalAddr: "0.0.0.0", LocalPort: 80},
				{Protocol: "tcp4", LocalAddr: "0.0.0.0", LocalPort: 443},
			},
			includePorts: []int{80, 443},
			wantCount:    2,
		},
		{
			name: "different addresses same port are unique",
			ports: []listeningPort{
				{Protocol: "tcp4", LocalAddr: "0.0.0.0", LocalPort: 80},
				{Protocol: "tcp4", LocalAddr: "127.0.0.1", LocalPort: 80},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.PortScanDiscoveryConfig{
				IncludePorts: tt.includePorts,
				ExcludePorts: tt.excludePorts,
			}
			discoverer := NewPortScanDiscoverer(cfg)

			targets := discoverer.filterAndConvertPorts(tt.ports)

			assert.Len(t, targets, tt.wantCount)
		})
	}
}

// TestPortScanDiscoverer_parseNetTCP verifies parsing of /proc/net/tcp files.
func TestPortScanDiscoverer_parseNetTCP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		protocol string
		wantErr  bool
	}{
		{
			name:     "non-existent file returns nil without error",
			path:     "/nonexistent/path/tcp",
			protocol: "tcp4",
			wantErr:  false,
		},
		{
			name:     "parses /proc/net/tcp if exists",
			path:     "/proc/net/tcp",
			protocol: "tcp4",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.PortScanDiscoveryConfig{}
			discoverer := NewPortScanDiscoverer(cfg)

			ports, err := discoverer.parseNetTCP(tt.path, tt.protocol)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Result may be nil for non-existent file or populated.
				_ = ports
			}
		})
	}
}

// TestPortScanDiscoverer_scanNetTCPFile verifies file scanning.
func TestPortScanDiscoverer_scanNetTCPFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		protocol  string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty file returns empty list",
			content:   "",
			protocol:  "tcp4",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "header only returns empty list",
			content:   "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode",
			protocol:  "tcp4",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "parses listening ports",
			content: `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:0050 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 12345 1`,
			protocol:  "tcp4",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "skips non-listening sockets",
			content: `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:0050 00000000:0000 01 00000000:00000000 00:00000000 00000000     0        0 12345 1`,
			protocol:  "tcp4",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp file with content using t.TempDir().
			tmpDir := t.TempDir()
			tmpPath := filepath.Join(tmpDir, "tcp_test")
			err := os.WriteFile(tmpPath, []byte(tt.content), 0600)
			require.NoError(t, err)

			tmpFile, err := os.Open(tmpPath)
			require.NoError(t, err)
			defer func() { _ = tmpFile.Close() }()

			cfg := &config.PortScanDiscoveryConfig{}
			discoverer := NewPortScanDiscoverer(cfg)

			ports, err := discoverer.scanNetTCPFile(tmpFile, tt.protocol)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, ports, tt.wantCount)
			}
		})
	}
}

// TestPortScanDiscoverer_matchesInterface verifies interface IP matching.
func TestPortScanDiscoverer_matchesInterface(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		interfaces []string
		addr       string
		wantMatch  bool
	}{
		{
			name:       "invalid IP returns false",
			interfaces: []string{"lo"},
			addr:       "invalid",
			wantMatch:  false,
		},
		{
			name:       "empty interfaces returns false for any valid IP",
			interfaces: []string{},
			addr:       "127.0.0.1",
			wantMatch:  false, // No interfaces configured means no match
		},
		{
			name:       "unknown interface returns false",
			interfaces: []string{"nonexistent-if"},
			addr:       "127.0.0.1",
			wantMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.PortScanDiscoveryConfig{
				Interfaces: tt.interfaces,
			}
			discoverer := NewPortScanDiscoverer(cfg)

			result := discoverer.matchesInterface(tt.addr)

			assert.Equal(t, tt.wantMatch, result)
		})
	}
}

// TestPortScanDiscoverer_portToTarget verifies port to target conversion.
func TestPortScanDiscoverer_portToTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		port          listeningPort
		wantID        string
		wantName      string
		wantType      target.Type
		wantProbeType string
		wantLabels    map[string]string
	}{
		{
			name: "converts TCP4 port",
			port: listeningPort{
				Protocol:  "tcp4",
				LocalAddr: "0.0.0.0",
				LocalPort: 80,
				State:     "0A",
			},
			wantID:        "portscan:0.0.0.0:80",
			wantName:      "tcp4:80",
			wantType:      target.TypeCustom,
			wantProbeType: "tcp",
			wantLabels: map[string]string{
				"portscan.protocol": "tcp4",
				"portscan.port":     "80",
				"portscan.address":  "0.0.0.0",
			},
		},
		{
			name: "converts TCP6 port",
			port: listeningPort{
				Protocol:  "tcp6",
				LocalAddr: "::1",
				LocalPort: 443,
				State:     "0A",
			},
			wantID:        "portscan:::1:443",
			wantName:      "tcp6:443",
			wantType:      target.TypeCustom,
			wantProbeType: "tcp",
			wantLabels: map[string]string{
				"portscan.protocol": "tcp6",
				"portscan.port":     "443",
				"portscan.address":  "::1",
			},
		},
		{
			name: "converts localhost port",
			port: listeningPort{
				Protocol:  "tcp4",
				LocalAddr: "127.0.0.1",
				LocalPort: 8080,
				State:     "0A",
			},
			wantID:        "portscan:127.0.0.1:8080",
			wantName:      "tcp4:8080",
			wantType:      target.TypeCustom,
			wantProbeType: "tcp",
			wantLabels: map[string]string{
				"portscan.protocol": "tcp4",
				"portscan.port":     "8080",
				"portscan.address":  "127.0.0.1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.PortScanDiscoveryConfig{}
			discoverer := NewPortScanDiscoverer(cfg)

			tgt := discoverer.portToTarget(tt.port)

			assert.Equal(t, tt.wantID, tgt.ID)
			assert.Equal(t, tt.wantName, tgt.Name)
			assert.Equal(t, tt.wantType, tgt.Type)
			assert.Equal(t, tt.wantProbeType, tgt.ProbeType)
			assert.Equal(t, target.SourceDiscovered, tgt.Source)

			for key, value := range tt.wantLabels {
				assert.Equal(t, value, tgt.Labels[key], "label %s mismatch", key)
			}

			// Verify probe target is configured.
			assert.NotNil(t, tgt.ProbeTarget)
			assert.True(t, strings.Contains(tgt.ProbeTarget.Address, ":"))
		})
	}
}
