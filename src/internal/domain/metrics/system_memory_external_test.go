package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestNewSystemMemory tests the NewSystemMemory constructor.
func TestNewSystemMemory(t *testing.T) {
	tests := []struct {
		name            string
		input           *metrics.SystemMemoryInput
		expectedUsed    uint64
		expectedPercent float64
	}{
		{
			name: "typical_system_50_percent",
			input: &metrics.SystemMemoryInput{
				Total:     16 * 1024 * 1024 * 1024, // 16 GB
				Available: 8 * 1024 * 1024 * 1024,  // 8 GB
				Free:      2 * 1024 * 1024 * 1024,
				Cached:    4 * 1024 * 1024 * 1024,
				Buffers:   1 * 1024 * 1024 * 1024,
				SwapTotal: 4 * 1024 * 1024 * 1024,
				SwapUsed:  1 * 1024 * 1024 * 1024,
				SwapFree:  3 * 1024 * 1024 * 1024,
				Shared:    512 * 1024 * 1024,
			},
			expectedUsed:    8 * 1024 * 1024 * 1024,
			expectedPercent: 50.0,
		},
		{
			name: "high_usage_75_percent",
			input: &metrics.SystemMemoryInput{
				Total:     4 * 1024 * 1024 * 1024, // 4 GB
				Available: 1 * 1024 * 1024 * 1024, // 1 GB
				Free:      500 * 1024 * 1024,
			},
			expectedUsed:    3 * 1024 * 1024 * 1024,
			expectedPercent: 75.0,
		},
		{
			name: "zero_total_memory",
			input: &metrics.SystemMemoryInput{
				Total:     0,
				Available: 0,
			},
			expectedUsed:    0,
			expectedPercent: 0.0,
		},
		{
			name: "full_usage_100_percent",
			input: &metrics.SystemMemoryInput{
				Total:     1000,
				Available: 0,
			},
			expectedUsed:    1000,
			expectedPercent: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := metrics.NewSystemMemory(tt.input)

			assert.Equal(t, tt.expectedUsed, mem.Used)
			assert.InDelta(t, tt.expectedPercent, mem.UsagePercent, 0.01)
			assert.Equal(t, tt.input.Total, mem.Total)
			assert.Equal(t, tt.input.Available, mem.Available)
			assert.Equal(t, tt.input.Free, mem.Free)
			assert.Equal(t, tt.input.Cached, mem.Cached)
			assert.Equal(t, tt.input.Buffers, mem.Buffers)
			assert.Equal(t, tt.input.SwapTotal, mem.SwapTotal)
			assert.Equal(t, tt.input.SwapUsed, mem.SwapUsed)
			assert.Equal(t, tt.input.SwapFree, mem.SwapFree)
			assert.Equal(t, tt.input.Shared, mem.Shared)
			assert.False(t, mem.Timestamp.IsZero())
		})
	}
}

// TestSystemMemory_SwapUsagePercent tests the SwapUsagePercent method.
func TestSystemMemory_SwapUsagePercent(t *testing.T) {
	tests := []struct {
		name     string
		mem      metrics.SystemMemory
		expected float64
	}{
		{
			name: "50_percent_swap_usage",
			mem: metrics.SystemMemory{
				SwapTotal: 1000,
				SwapUsed:  500,
			},
			expected: 50.0,
		},
		{
			name: "no_swap",
			mem: metrics.SystemMemory{
				SwapTotal: 0,
				SwapUsed:  0,
			},
			expected: 0.0,
		},
		{
			name: "100_percent_swap_usage",
			mem: metrics.SystemMemory{
				SwapTotal: 1000,
				SwapUsed:  1000,
			},
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mem.SwapUsagePercent()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSystemMemory_Fields tests all fields on SystemMemory.
func TestSystemMemory_Fields(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name              string
		mem               metrics.SystemMemory
		expectedTotal     uint64
		expectedAvailable uint64
		expectedUsed      uint64
		expectedPercent   float64
	}{
		{
			name: "typical_16gb_system",
			mem: metrics.SystemMemory{
				Total:        16 * 1024 * 1024 * 1024, // 16 GB
				Available:    8 * 1024 * 1024 * 1024,  // 8 GB
				Used:         8 * 1024 * 1024 * 1024,  // 8 GB
				Free:         2 * 1024 * 1024 * 1024,  // 2 GB
				Cached:       4 * 1024 * 1024 * 1024,  // 4 GB
				Buffers:      1 * 1024 * 1024 * 1024,  // 1 GB
				SwapTotal:    4 * 1024 * 1024 * 1024,  // 4 GB
				SwapUsed:     1 * 1024 * 1024 * 1024,  // 1 GB
				SwapFree:     3 * 1024 * 1024 * 1024,  // 3 GB
				Shared:       512 * 1024 * 1024,       // 512 MB
				UsagePercent: 50.0,
				Timestamp:    now,
			},
			expectedTotal:     16 * 1024 * 1024 * 1024,
			expectedAvailable: 8 * 1024 * 1024 * 1024,
			expectedUsed:      8 * 1024 * 1024 * 1024,
			expectedPercent:   50.0,
		},
		{
			name: "low_memory_system",
			mem: metrics.SystemMemory{
				Total:        4 * 1024 * 1024 * 1024, // 4 GB
				Available:    1 * 1024 * 1024 * 1024, // 1 GB
				Used:         3 * 1024 * 1024 * 1024, // 3 GB
				UsagePercent: 75.0,
				Timestamp:    now,
			},
			expectedTotal:     4 * 1024 * 1024 * 1024,
			expectedAvailable: 1 * 1024 * 1024 * 1024,
			expectedUsed:      3 * 1024 * 1024 * 1024,
			expectedPercent:   75.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedTotal, tt.mem.Total)
			assert.Equal(t, tt.expectedAvailable, tt.mem.Available)
			assert.Equal(t, tt.expectedUsed, tt.mem.Used)
			assert.Equal(t, tt.expectedPercent, tt.mem.UsagePercent)
		})
	}
}
