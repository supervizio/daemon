// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestNewBandwidth tests the NewBandwidth constructor.
func TestNewBandwidth(t *testing.T) {
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
			bw := metrics.NewBandwidth(tt.iface, tt.timestamp)

			assert.Equal(t, tt.iface, bw.Interface)
			assert.Equal(t, tt.timestamp, bw.Timestamp)
		})
	}
}

// TestBandwidth_TotalBytesPerSec tests the TotalBytesPerSec method.
func TestBandwidth_TotalBytesPerSec(t *testing.T) {
	tests := []struct {
		name      string
		bandwidth metrics.Bandwidth
		expected  float64
	}{
		{
			name: "typical_bandwidth",
			bandwidth: metrics.Bandwidth{
				TxBytesPerSec: 1024 * 1024,     // 1MB/s tx
				RxBytesPerSec: 1024 * 1024 * 2, // 2MB/s rx
			},
			expected: 1024 * 1024 * 3,
		},
		{
			name: "no_traffic",
			bandwidth: metrics.Bandwidth{
				TxBytesPerSec: 0,
				RxBytesPerSec: 0,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.bandwidth.TotalBytesPerSec())
		})
	}
}

// TestBandwidth_TotalPacketsPerSec tests the TotalPacketsPerSec method.
func TestBandwidth_TotalPacketsPerSec(t *testing.T) {
	tests := []struct {
		name      string
		bandwidth metrics.Bandwidth
		expected  float64
	}{
		{
			name: "typical_traffic",
			bandwidth: metrics.Bandwidth{
				TxPacketsPerSec: 1000,
				RxPacketsPerSec: 2000,
			},
			expected: 3000,
		},
		{
			name: "no_traffic",
			bandwidth: metrics.Bandwidth{
				TxPacketsPerSec: 0,
				RxPacketsPerSec: 0,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.bandwidth.TotalPacketsPerSec())
		})
	}
}

// TestBandwidth_TxBitsPerSec tests the TxBitsPerSec method.
func TestBandwidth_TxBitsPerSec(t *testing.T) {
	tests := []struct {
		name      string
		bandwidth metrics.Bandwidth
		expected  float64
	}{
		{
			name: "typical_bandwidth",
			bandwidth: metrics.Bandwidth{
				TxBytesPerSec: 1024 * 1024, // 1MB/s
			},
			expected: 1024 * 1024 * 8, // 8Mbps
		},
		{
			name: "no_traffic",
			bandwidth: metrics.Bandwidth{
				TxBytesPerSec: 0,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.bandwidth.TxBitsPerSec())
		})
	}
}

// TestBandwidth_RxBitsPerSec tests the RxBitsPerSec method.
func TestBandwidth_RxBitsPerSec(t *testing.T) {
	tests := []struct {
		name      string
		bandwidth metrics.Bandwidth
		expected  float64
	}{
		{
			name: "typical_bandwidth",
			bandwidth: metrics.Bandwidth{
				RxBytesPerSec: 1024 * 1024 * 2, // 2MB/s
			},
			expected: 1024 * 1024 * 2 * 8, // 16Mbps
		},
		{
			name: "no_traffic",
			bandwidth: metrics.Bandwidth{
				RxBytesPerSec: 0,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.bandwidth.RxBitsPerSec())
		})
	}
}

// TestCalculateBandwidth tests the CalculateBandwidth function.
func TestCalculateBandwidth(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		prev       metrics.NetStats
		curr       metrics.NetStats
		wantTxRate float64
		wantRxRate float64
	}{
		{
			name: "calculate_from_samples",
			prev: metrics.NetStats{
				Interface:   "eth0",
				BytesSent:   1000,
				BytesRecv:   2000,
				PacketsSent: 100,
				PacketsRecv: 200,
				Timestamp:   now,
			},
			curr: metrics.NetStats{
				Interface:   "eth0",
				BytesSent:   2000,
				BytesRecv:   4000,
				PacketsSent: 200,
				PacketsRecv: 400,
				Timestamp:   now.Add(time.Second),
			},
			wantTxRate: 1000, // 1000 bytes/sec
			wantRxRate: 2000, // 2000 bytes/sec
		},
		{
			name: "zero_duration",
			prev: metrics.NetStats{
				Interface: "eth0",
				BytesSent: 1000,
				BytesRecv: 2000,
				Timestamp: now,
			},
			curr: metrics.NetStats{
				Interface: "eth0",
				BytesSent: 2000,
				BytesRecv: 4000,
				Timestamp: now, // Same timestamp
			},
			wantTxRate: 0, // Zero duration
			wantRxRate: 0, // Zero duration
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bw := metrics.CalculateBandwidth(&tt.prev, &tt.curr)
			assert.InDelta(t, tt.wantTxRate, bw.TxBytesPerSec, 1)
			assert.InDelta(t, tt.wantRxRate, bw.RxBytesPerSec, 1)
		})
	}
}
