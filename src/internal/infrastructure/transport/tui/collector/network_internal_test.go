// Package collector provides internal tests for network.go.
// It tests internal implementation details using white-box testing.
package collector

import (
	"errors"
	"net"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// mockAddrserNetwork implements Addrser for network testing.
type mockAddrserNetwork struct {
	addrs []net.Addr
	err   error
}

// Addrs returns the mock addresses.
//
// Returns:
//   - []net.Addr: mock addresses
//   - error: mock error if set
func (m *mockAddrserNetwork) Addrs() ([]net.Addr, error) {
	return m.addrs, m.err
}

// Test_netStats tests the netStats struct.
// It verifies that the struct fields work correctly.
//
// Params:
//   - t: the testing context.
func Test_netStats(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// rxBytes is the RX bytes value.
		rxBytes uint64
		// txBytes is the TX bytes value.
		txBytes uint64
	}{
		{
			name:    "zero_stats",
			rxBytes: 0,
			txBytes: 0,
		},
		{
			name:    "small_values",
			rxBytes: 1024,
			txBytes: 512,
		},
		{
			name:    "large_values",
			rxBytes: 1024 * 1024 * 1024,
			txBytes: 512 * 1024 * 1024,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create stats struct.
			stats := netStats{
				rxBytes: tt.rxBytes,
				txBytes: tt.txBytes,
			}

			// Verify fields.
			assert.Equal(t, tt.rxBytes, stats.rxBytes)
			assert.Equal(t, tt.txBytes, stats.txBytes)
		})
	}
}

// Test_NetworkCollector_getInterfaceIPv4 tests the getInterfaceIPv4 method.
// It verifies that IPv4 addresses are correctly extracted.
//
// Params:
//   - t: the testing context.
func Test_NetworkCollector_getInterfaceIPv4(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// addrser is the mock Addrser.
		addrser Addrser
		// wantIP is the expected IP.
		wantIP string
	}{
		{
			name: "returns_ipv4_from_ipnet",
			addrser: &mockAddrserNetwork{
				addrs: []net.Addr{
					&net.IPNet{
						IP:   net.ParseIP("192.168.1.100"),
						Mask: net.CIDRMask(24, 32),
					},
				},
			},
			wantIP: "192.168.1.100",
		},
		{
			name: "returns_empty_for_ipv6_only",
			addrser: &mockAddrserNetwork{
				addrs: []net.Addr{
					&net.IPNet{
						IP:   net.ParseIP("::1"),
						Mask: net.CIDRMask(128, 128),
					},
				},
			},
			wantIP: "",
		},
		{
			name: "returns_first_ipv4",
			addrser: &mockAddrserNetwork{
				addrs: []net.Addr{
					&net.IPNet{
						IP:   net.ParseIP("::1"),
						Mask: net.CIDRMask(128, 128),
					},
					&net.IPNet{
						IP:   net.ParseIP("10.0.0.1"),
						Mask: net.CIDRMask(8, 32),
					},
				},
			},
			wantIP: "10.0.0.1",
		},
		{
			name: "returns_empty_for_no_addrs",
			addrser: &mockAddrserNetwork{
				addrs: nil,
			},
			wantIP: "",
		},
		{
			name: "returns_empty_on_error",
			addrser: &mockAddrserNetwork{
				err: errors.New("mock error"),
			},
			wantIP: "",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collector.
			c := NewNetworkCollector()

			// Call method.
			result := c.getInterfaceIPv4(tt.addrser)

			// Verify result.
			assert.Equal(t, tt.wantIP, result)
		})
	}
}

// Test_NetworkCollector_calculateRates tests the calculateRates method.
// It verifies that RX/TX rates are correctly calculated.
//
// Params:
//   - t: the testing context.
func Test_NetworkCollector_calculateRates(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// prevRxBytes is previous RX bytes.
		prevRxBytes uint64
		// prevTxBytes is previous TX bytes.
		prevTxBytes uint64
		// currRxBytes is current RX bytes.
		currRxBytes uint64
		// currTxBytes is current TX bytes.
		currTxBytes uint64
		// hasPrev indicates if previous stats exist.
		hasPrev bool
		// wantRxRate is expected RX rate.
		wantRxRate uint64
		// wantTxRate is expected TX rate.
		wantTxRate uint64
	}{
		{
			name:        "no_previous_stats",
			currRxBytes: 1000,
			currTxBytes: 500,
			hasPrev:     false,
			wantRxRate:  0,
			wantTxRate:  0,
		},
		{
			name:        "with_previous_stats",
			prevRxBytes: 1000,
			prevTxBytes: 500,
			currRxBytes: 2000,
			currTxBytes: 1500,
			hasPrev:     true,
			wantRxRate:  1000,
			wantTxRate:  1000,
		},
		{
			name:        "counter_wrap_protection",
			prevRxBytes: 2000,
			prevTxBytes: 1500,
			currRxBytes: 1000,
			currTxBytes: 500,
			hasPrev:     true,
			wantRxRate:  0,
			wantTxRate:  0,
		},
		{
			name:        "no_change",
			prevRxBytes: 1000,
			prevTxBytes: 500,
			currRxBytes: 1000,
			currTxBytes: 500,
			hasPrev:     true,
			wantRxRate:  0,
			wantTxRate:  0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collector.
			c := NewNetworkCollector()

			// Set up previous stats if needed.
			if tt.hasPrev {
				c.prevStats["eth0"] = netStats{
					rxBytes: tt.prevRxBytes,
					txBytes: tt.prevTxBytes,
				}
			}

			// Create network interface.
			ni := &model.NetworkInterface{
				Name: "eth0",
			}

			// Call method.
			c.calculateRates(ni, "eth0", tt.currRxBytes, tt.currTxBytes)

			// Verify rates.
			assert.Equal(t, tt.wantRxRate, ni.RxBytesPerSec)
			assert.Equal(t, tt.wantTxRate, ni.TxBytesPerSec)
		})
	}
}

// Test_NetworkCollector_collectInterface tests the collectInterface method.
// It verifies that interface data is correctly collected.
//
// Params:
//   - t: the testing context.
func Test_NetworkCollector_collectInterface(t *testing.T) {
	t.Parallel()

	// Get a real interface for testing.
	interfaces, err := net.Interfaces()
	if err != nil || len(interfaces) == 0 {
		// No network interfaces available for testing.
		return
	}

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// ifaceIdx is the index of interface to use.
		ifaceIdx int
	}{
		{
			name:     "collects_first_interface",
			ifaceIdx: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collector.
			c := NewNetworkCollector()

			// Get interface.
			iface := interfaces[tt.ifaceIdx]

			// Call method.
			result := c.collectInterface(iface)

			// Verify result.
			assert.Equal(t, iface.Name, result.Name)
			assert.Equal(t, iface.Flags&net.FlagUp != 0, result.IsUp)
			assert.Equal(t, iface.Flags&net.FlagLoopback != 0, result.IsLoopback)
		})
	}
}

// Test_NetworkCollector_prevStats_persistence tests prevStats map persistence.
// It verifies that stats are stored for subsequent rate calculations.
//
// Params:
//   - t: the testing context.
func Test_NetworkCollector_prevStats_persistence(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "stores_stats_after_gather",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collector.
			c := NewNetworkCollector()

			// Create snapshot.
			snap := &model.Snapshot{}

			// Gather first time.
			_ = c.Gather(snap)

			// Verify prevStats was populated (if interfaces exist).
			if len(snap.Network) > 0 {
				assert.NotEmpty(t, c.prevStats)
			}
		})
	}
}

// Test_typicalInterfaceCount tests the typicalInterfaceCount constant.
// It verifies that the constant is a reasonable value.
//
// Params:
//   - t: the testing context.
func Test_typicalInterfaceCount(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// wantValue is the expected value.
		wantValue int
	}{
		{
			name:      "has_reasonable_value",
			wantValue: 8,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify constant value.
			assert.Equal(t, tt.wantValue, typicalInterfaceCount)
		})
	}
}

// Test_bitsPerByte tests the bitsPerByte constant.
// It verifies that the constant has the correct value.
//
// Params:
//   - t: the testing context.
func Test_bitsPerByte(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// wantValue is the expected value.
		wantValue uint64
	}{
		{
			name:      "has_correct_value",
			wantValue: 8,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify constant value.
			assert.Equal(t, tt.wantValue, bitsPerByte)
		})
	}
}
