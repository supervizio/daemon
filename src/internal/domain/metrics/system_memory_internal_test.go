// Package metrics contains internal tests for system memory metrics.
package metrics

import (
	"testing"
)

// TestNewSystemMemory_calculations tests the internal percentage calculations.
func TestNewSystemMemory_calculations(t *testing.T) {
	tests := []struct {
		name                string
		input               *SystemMemoryInput
		wantUsed            uint64
		wantUsagePercent    float64
		wantTimestampNotNil bool
	}{
		{
			name: "typical_50_percent_usage",
			input: &SystemMemoryInput{
				Total:     1000,
				Available: 500,
				Free:      200,
				Cached:    200,
				Buffers:   100,
			},
			wantUsed:            500,
			wantUsagePercent:    50.0,
			wantTimestampNotNil: true,
		},
		{
			name: "zero_total_avoids_division_by_zero",
			input: &SystemMemoryInput{
				Total:     0,
				Available: 0,
			},
			wantUsed:            0,
			wantUsagePercent:    0.0,
			wantTimestampNotNil: true,
		},
		{
			name: "full_memory_100_percent",
			input: &SystemMemoryInput{
				Total:     1000,
				Available: 0,
			},
			wantUsed:            1000,
			wantUsagePercent:    100.0,
			wantTimestampNotNil: true,
		},
		{
			name: "high_usage_75_percent",
			input: &SystemMemoryInput{
				Total:     1000,
				Available: 250,
			},
			wantUsed:            750,
			wantUsagePercent:    75.0,
			wantTimestampNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewSystemMemory(tt.input)

			// Validate used memory calculation.
			if result.Used != tt.wantUsed {
				t.Errorf("Used = %v, want %v", result.Used, tt.wantUsed)
			}
			// Validate usage percentage calculation.
			if result.UsagePercent != tt.wantUsagePercent {
				t.Errorf("UsagePercent = %v, want %v", result.UsagePercent, tt.wantUsagePercent)
			}
			// Validate timestamp is set.
			if tt.wantTimestampNotNil && result.Timestamp.IsZero() {
				t.Error("Timestamp should not be zero")
			}
		})
	}
}
