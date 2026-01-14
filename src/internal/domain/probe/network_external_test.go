// Package metrics_test provides black-box tests for the metrics package.
package probe_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// TestNetInterface tests NetInterface value object methods.
func TestNetInterface(t *testing.T) {
	tests := []struct {
		name         string
		iface        probe.NetInterface
		wantUp       bool
		wantLoopback bool
	}{
		{
			name: "eth0_up",
			iface: probe.NetInterface{
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
			iface: probe.NetInterface{
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
			iface: probe.NetInterface{
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

// TestNetStats tests NetStats value object methods.
func TestNetStats(t *testing.T) {
	tests := []struct {
		name            string
		stats           probe.NetStats
		wantTotalBytes  uint64
		wantTotalPkts   uint64
		wantTotalErrors uint64
		wantTotalDrops  uint64
	}{
		{
			name: "eth0_stats",
			stats: probe.NetStats{
				Interface:   "eth0",
				BytesSent:   1024 * 1024 * 100, // 100MB sent
				BytesRecv:   1024 * 1024 * 200, // 200MB received
				PacketsSent: 10000,
				PacketsRecv: 20000,
				ErrorsIn:    5,
				ErrorsOut:   3,
				DropsIn:     2,
				DropsOut:    1,
				Timestamp:   time.Now(),
			},
			wantTotalBytes:  1024 * 1024 * 300,
			wantTotalPkts:   30000,
			wantTotalErrors: 8,
			wantTotalDrops:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantTotalBytes, tt.stats.TotalBytes())
			assert.Equal(t, tt.wantTotalPkts, tt.stats.TotalPackets())
			assert.Equal(t, tt.wantTotalErrors, tt.stats.TotalErrors())
			assert.Equal(t, tt.wantTotalDrops, tt.stats.TotalDrops())
		})
	}
}

// TestBandwidth tests Bandwidth value object methods.
func TestBandwidth(t *testing.T) {
	tests := []struct {
		name              string
		bandwidth         probe.Bandwidth
		wantTotalBytesPS  float64
		wantTotalPktsPS   float64
		wantTxBitsPerSec  float64
		wantRxBitsPerSec  float64
	}{
		{
			name: "eth0_bandwidth",
			bandwidth: probe.Bandwidth{
				Interface:       "eth0",
				TxBytesPerSec:   1024 * 1024,     // 1MB/s tx
				RxBytesPerSec:   1024 * 1024 * 2, // 2MB/s rx
				TxPacketsPerSec: 1000,
				RxPacketsPerSec: 2000,
				Duration:        time.Second,
				Timestamp:       time.Now(),
			},
			wantTotalBytesPS: 1024 * 1024 * 3,
			wantTotalPktsPS:  3000,
			wantTxBitsPerSec: 1024 * 1024 * 8,
			wantRxBitsPerSec: 1024 * 1024 * 2 * 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantTotalBytesPS, tt.bandwidth.TotalBytesPerSec())
			assert.Equal(t, tt.wantTotalPktsPS, tt.bandwidth.TotalPacketsPerSec())
			assert.Equal(t, tt.wantTxBitsPerSec, tt.bandwidth.TxBitsPerSec())
			assert.Equal(t, tt.wantRxBitsPerSec, tt.bandwidth.RxBitsPerSec())
		})
	}
}

// TestCalculateBandwidth tests the CalculateBandwidth function.
func TestCalculateBandwidth(t *testing.T) {
	tests := []struct {
		name       string
		prev       probe.NetStats
		curr       probe.NetStats
		wantTxRate float64
		wantRxRate float64
	}{
		{
			name: "calculate_from_samples",
			prev: probe.NetStats{
				Interface:   "eth0",
				BytesSent:   1000,
				BytesRecv:   2000,
				PacketsSent: 100,
				PacketsRecv: 200,
				Timestamp:   time.Now(),
			},
			curr: probe.NetStats{
				Interface:   "eth0",
				BytesSent:   2000,
				BytesRecv:   4000,
				PacketsSent: 200,
				PacketsRecv: 400,
				Timestamp:   time.Now().Add(time.Second),
			},
			wantTxRate: 1000, // 1000 bytes/sec
			wantRxRate: 2000, // 2000 bytes/sec
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bw := probe.CalculateBandwidth(tt.prev, tt.curr)
			assert.InDelta(t, tt.wantTxRate, bw.TxBytesPerSec, 1)
			assert.InDelta(t, tt.wantRxRate, bw.RxBytesPerSec, 1)
		})
	}
}

// TestCalculateBandwidth_ZeroDuration tests zero duration handling.
func TestCalculateBandwidth_ZeroDuration(t *testing.T) {
	now := time.Now()
	prev := probe.NetStats{
		Interface: "eth0",
		BytesSent: 1000,
		BytesRecv: 2000,
		Timestamp: now,
	}
	curr := probe.NetStats{
		Interface: "eth0",
		BytesSent: 2000,
		BytesRecv: 4000,
		Timestamp: now, // Same timestamp
	}

	bw := probe.CalculateBandwidth(prev, curr)
	assert.Equal(t, float64(0), bw.TxBytesPerSec)
	assert.Equal(t, float64(0), bw.RxBytesPerSec)
}
