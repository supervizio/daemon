// Package metrics contains internal tests for process memory metrics.
package metrics

import (
	"testing"
)

// TestNewProcessMemory_calculations tests the internal percentage calculations.
func TestNewProcessMemory_calculations(t *testing.T) {
	tests := []struct {
		name                string
		input               *ProcessMemoryInput
		wantUsagePercent    float64
		wantTimestampNotNil bool
	}{
		{
			name: "typical_2_5_percent_usage",
			input: &ProcessMemoryInput{
				PID:               1234,
				Name:              "test",
				RSS:               100,
				TotalSystemMemory: 4000,
			},
			wantUsagePercent:    2.5,
			wantTimestampNotNil: true,
		},
		{
			name: "zero_total_avoids_division_by_zero",
			input: &ProcessMemoryInput{
				PID:               1,
				Name:              "init",
				RSS:               50,
				TotalSystemMemory: 0,
			},
			wantUsagePercent:    0.0,
			wantTimestampNotNil: true,
		},
		{
			name: "full_memory_100_percent",
			input: &ProcessMemoryInput{
				PID:               9999,
				Name:              "memhog",
				RSS:               1000,
				TotalSystemMemory: 1000,
			},
			wantUsagePercent:    100.0,
			wantTimestampNotNil: true,
		},
		{
			name: "low_usage_0_1_percent",
			input: &ProcessMemoryInput{
				PID:               2,
				Name:              "kthreadd",
				RSS:               1,
				TotalSystemMemory: 1000,
			},
			wantUsagePercent:    0.1,
			wantTimestampNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewProcessMemory(tt.input)

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
