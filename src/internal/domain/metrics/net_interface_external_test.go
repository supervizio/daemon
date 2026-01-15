// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestNetInterface tests NetInterface value object methods.
func TestNetInterface(t *testing.T) {
	tests := []struct {
		name         string
		iface        metrics.NetInterface
		wantUp       bool
		wantLoopback bool
	}{
		{
			name: "eth0_up",
			iface: metrics.NetInterface{
				Name:         "eth0",
				Index:        2,
				HardwareAddr: "00:11:22:33:44:55",
				MTU:          1500,
				Flags:        []string{"up", "broadcast", "multicast"},
				Addresses:    []string{"192.168.1.100/24"},
			},
			wantUp:       true,
			wantLoopback: false,
		},
		{
			name: "lo_loopback",
			iface: metrics.NetInterface{
				Name:         "lo",
				Index:        1,
				HardwareAddr: "",
				MTU:          65536,
				Flags:        []string{"up", "loopback"},
				Addresses:    []string{"127.0.0.1/8", "::1/128"},
			},
			wantUp:       true,
			wantLoopback: true,
		},
		{
			name: "eth1_down",
			iface: metrics.NetInterface{
				Name:         "eth1",
				Index:        3,
				HardwareAddr: "00:11:22:33:44:66",
				MTU:          1500,
				Flags:        []string{"broadcast", "multicast"},
				Addresses:    nil,
			},
			wantUp:       false,
			wantLoopback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantUp, tt.iface.IsUp())
			assert.Equal(t, tt.wantLoopback, tt.iface.IsLoopback())
		})
	}
}

// TestNewNetInterface tests the NewNetInterface constructor.
func TestNewNetInterface(t *testing.T) {
	tests := []struct {
		name      string
		ifaceName string
		index     int
	}{
		{
			name:      "eth0_interface",
			ifaceName: "eth0",
			index:     2,
		},
		{
			name:      "loopback_interface",
			ifaceName: "lo",
			index:     1,
		},
		{
			name:      "zero_index",
			ifaceName: "wlan0",
			index:     0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create NetInterface using constructor.
			iface := metrics.NewNetInterface(tt.ifaceName, tt.index)

			// Verify fields are correctly set.
			assert.Equal(t, tt.ifaceName, iface.Name)
			assert.Equal(t, tt.index, iface.Index)
		})
	}
}

// TestNetInterface_IsUp tests the IsUp method on NetInterface.
func TestNetInterface_IsUp(t *testing.T) {
	tests := []struct {
		name   string
		iface  metrics.NetInterface
		wantUp bool
	}{
		{
			name: "interface_up",
			iface: metrics.NetInterface{
				Name:  "eth0",
				Flags: []string{"up", "broadcast", "multicast"},
			},
			wantUp: true,
		},
		{
			name: "interface_down",
			iface: metrics.NetInterface{
				Name:  "eth1",
				Flags: []string{"broadcast", "multicast"},
			},
			wantUp: false,
		},
		{
			name: "empty_flags",
			iface: metrics.NetInterface{
				Name:  "eth2",
				Flags: []string{},
			},
			wantUp: false,
		},
		{
			name: "nil_flags",
			iface: metrics.NetInterface{
				Name:  "eth3",
				Flags: nil,
			},
			wantUp: false,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check IsUp result.
			assert.Equal(t, tt.wantUp, tt.iface.IsUp())
		})
	}
}

// TestNetInterface_IsLoopback tests the IsLoopback method on NetInterface.
func TestNetInterface_IsLoopback(t *testing.T) {
	tests := []struct {
		name         string
		iface        metrics.NetInterface
		wantLoopback bool
	}{
		{
			name: "loopback_interface",
			iface: metrics.NetInterface{
				Name:  "lo",
				Flags: []string{"up", "loopback"},
			},
			wantLoopback: true,
		},
		{
			name: "physical_interface",
			iface: metrics.NetInterface{
				Name:  "eth0",
				Flags: []string{"up", "broadcast", "multicast"},
			},
			wantLoopback: false,
		},
		{
			name: "empty_flags",
			iface: metrics.NetInterface{
				Name:  "eth1",
				Flags: []string{},
			},
			wantLoopback: false,
		},
		{
			name: "nil_flags",
			iface: metrics.NetInterface{
				Name:  "eth2",
				Flags: nil,
			},
			wantLoopback: false,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check IsLoopback result.
			assert.Equal(t, tt.wantLoopback, tt.iface.IsLoopback())
		})
	}
}
