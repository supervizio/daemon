// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestIOStats tests IOStats value object methods.
func TestIOStats(t *testing.T) {
	tests := []struct {
		name           string
		stats          metrics.IOStats
		wantTotalBytes uint64
		wantTotalOps   uint64
	}{
		{
			name: "system_io_stats",
			stats: metrics.IOStats{
				ReadBytesTotal:  1024 * 1024 * 100, // 100MB read
				WriteBytesTotal: 1024 * 1024 * 50,  // 50MB written
				ReadOpsTotal:    10000,
				WriteOpsTotal:   5000,
				Timestamp:       time.Now(),
			},
			wantTotalBytes: 1024 * 1024 * 150,
			wantTotalOps:   15000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantTotalBytes, tt.stats.TotalBytes())
			assert.Equal(t, tt.wantTotalOps, tt.stats.TotalOps())
		})
	}
}

// TestNewIOStats tests the NewIOStats constructor.
func TestNewIOStats(t *testing.T) {
	tests := []struct {
		name       string
		readBytes  uint64
		writeBytes uint64
		readOps    uint64
		writeOps   uint64
		timestamp  time.Time
	}{
		{
			name:       "all_fields_populated",
			readBytes:  1024 * 1024 * 100,
			writeBytes: 1024 * 1024 * 50,
			readOps:    10000,
			writeOps:   5000,
			timestamp:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:       "zero_values",
			readBytes:  0,
			writeBytes: 0,
			readOps:    0,
			writeOps:   0,
			timestamp:  time.Time{},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create IOStats using constructor.
			stats := metrics.NewIOStats(tt.readBytes, tt.writeBytes, tt.readOps, tt.writeOps, tt.timestamp)

			// Verify all fields are correctly set.
			assert.Equal(t, tt.readBytes, stats.ReadBytesTotal)
			assert.Equal(t, tt.writeBytes, stats.WriteBytesTotal)
			assert.Equal(t, tt.readOps, stats.ReadOpsTotal)
			assert.Equal(t, tt.writeOps, stats.WriteOpsTotal)
			assert.Equal(t, tt.timestamp, stats.Timestamp)
		})
	}
}

// TestIOStats_TotalBytes tests the TotalBytes method on IOStats.
func TestIOStats_TotalBytes(t *testing.T) {
	tests := []struct {
		name           string
		stats          metrics.IOStats
		wantTotalBytes uint64
	}{
		{
			name: "read_and_write",
			stats: metrics.IOStats{
				ReadBytesTotal:  1024,
				WriteBytesTotal: 2048,
			},
			wantTotalBytes: 3072,
		},
		{
			name: "only_reads",
			stats: metrics.IOStats{
				ReadBytesTotal:  5000,
				WriteBytesTotal: 0,
			},
			wantTotalBytes: 5000,
		},
		{
			name: "only_writes",
			stats: metrics.IOStats{
				ReadBytesTotal:  0,
				WriteBytesTotal: 8000,
			},
			wantTotalBytes: 8000,
		},
		{
			name: "zero_values",
			stats: metrics.IOStats{
				ReadBytesTotal:  0,
				WriteBytesTotal: 0,
			},
			wantTotalBytes: 0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate total bytes.
			assert.Equal(t, tt.wantTotalBytes, tt.stats.TotalBytes())
		})
	}
}

// TestIOStats_TotalOps tests the TotalOps method on IOStats.
func TestIOStats_TotalOps(t *testing.T) {
	tests := []struct {
		name         string
		stats        metrics.IOStats
		wantTotalOps uint64
	}{
		{
			name: "read_and_write_ops",
			stats: metrics.IOStats{
				ReadOpsTotal:  100,
				WriteOpsTotal: 50,
			},
			wantTotalOps: 150,
		},
		{
			name: "only_read_ops",
			stats: metrics.IOStats{
				ReadOpsTotal:  200,
				WriteOpsTotal: 0,
			},
			wantTotalOps: 200,
		},
		{
			name: "only_write_ops",
			stats: metrics.IOStats{
				ReadOpsTotal:  0,
				WriteOpsTotal: 300,
			},
			wantTotalOps: 300,
		},
		{
			name: "zero_values",
			stats: metrics.IOStats{
				ReadOpsTotal:  0,
				WriteOpsTotal: 0,
			},
			wantTotalOps: 0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate total ops.
			assert.Equal(t, tt.wantTotalOps, tt.stats.TotalOps())
		})
	}
}
