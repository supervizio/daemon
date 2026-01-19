// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestNewNetStats tests the NewNetStats constructor.
func TestNewNetStats(t *testing.T) {
	tests := []struct {
		name      string
		iface     string
		timestamp time.Time
	}{
		{
			name:      "eth0_interface",
			iface:     "eth0",
			timestamp: time.Now(),
		},
		{
			name:      "lo_interface",
			iface:     "lo",
			timestamp: time.Now().Add(-time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := metrics.NewNetStats(tt.iface, tt.timestamp)

			assert.Equal(t, tt.iface, stats.Interface)
			assert.Equal(t, tt.timestamp, stats.Timestamp)
		})
	}
}

// TestNetStats_TotalBytes tests the TotalBytes method.
func TestNetStats_TotalBytes(t *testing.T) {
	tests := []struct {
		name     string
		stats    metrics.NetStats
		expected uint64
	}{
		{
			name: "typical_traffic",
			stats: metrics.NetStats{
				BytesSent: 1024 * 1024 * 100, // 100MB sent
				BytesRecv: 1024 * 1024 * 200, // 200MB received
			},
			expected: 1024 * 1024 * 300,
		},
		{
			name: "no_traffic",
			stats: metrics.NetStats{
				BytesSent: 0,
				BytesRecv: 0,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.stats.TotalBytes())
		})
	}
}

// TestNetStats_TotalPackets tests the TotalPackets method.
func TestNetStats_TotalPackets(t *testing.T) {
	tests := []struct {
		name     string
		stats    metrics.NetStats
		expected uint64
	}{
		{
			name: "typical_traffic",
			stats: metrics.NetStats{
				PacketsSent: 10000,
				PacketsRecv: 20000,
			},
			expected: 30000,
		},
		{
			name: "no_traffic",
			stats: metrics.NetStats{
				PacketsSent: 0,
				PacketsRecv: 0,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.stats.TotalPackets())
		})
	}
}

// TestNetStats_TotalErrors tests the TotalErrors method.
func TestNetStats_TotalErrors(t *testing.T) {
	tests := []struct {
		name     string
		stats    metrics.NetStats
		expected uint64
	}{
		{
			name: "with_errors",
			stats: metrics.NetStats{
				ErrorsIn:  5,
				ErrorsOut: 3,
			},
			expected: 8,
		},
		{
			name: "no_errors",
			stats: metrics.NetStats{
				ErrorsIn:  0,
				ErrorsOut: 0,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.stats.TotalErrors())
		})
	}
}

// TestNetStats_TotalDrops tests the TotalDrops method.
func TestNetStats_TotalDrops(t *testing.T) {
	tests := []struct {
		name     string
		stats    metrics.NetStats
		expected uint64
	}{
		{
			name: "with_drops",
			stats: metrics.NetStats{
				DropsIn:  2,
				DropsOut: 1,
			},
			expected: 3,
		},
		{
			name: "no_drops",
			stats: metrics.NetStats{
				DropsIn:  0,
				DropsOut: 0,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.stats.TotalDrops())
		})
	}
}
