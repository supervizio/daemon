package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestNewProcessMemory tests the NewProcessMemory constructor.
func TestNewProcessMemory(t *testing.T) {
	tests := []struct {
		name            string
		input           *metrics.ProcessMemoryInput
		expectedPercent float64
	}{
		{
			name: "typical_process_2_5_percent",
			input: &metrics.ProcessMemoryInput{
				PID:               1234,
				Name:              "myprocess",
				RSS:               100 * 1024 * 1024, // 100 MB
				VMS:               500 * 1024 * 1024,
				Shared:            20 * 1024 * 1024,
				Swap:              5 * 1024 * 1024,
				Data:              50 * 1024 * 1024,
				Stack:             8 * 1024 * 1024,
				TotalSystemMemory: 4 * 1024 * 1024 * 1024, // 4 GB
			},
			expectedPercent: 2.44140625, // 100MB / 4GB * 100
		},
		{
			name: "init_process_low_usage",
			input: &metrics.ProcessMemoryInput{
				PID:               1,
				Name:              "init",
				RSS:               10 * 1024 * 1024, // 10 MB
				VMS:               50 * 1024 * 1024,
				TotalSystemMemory: 16 * 1024 * 1024 * 1024, // 16 GB
			},
			expectedPercent: 0.06103515625, // 10MB / 16GB * 100
		},
		{
			name: "zero_total_system_memory",
			input: &metrics.ProcessMemoryInput{
				PID:               9999,
				Name:              "orphan",
				RSS:               50 * 1024 * 1024,
				TotalSystemMemory: 0,
			},
			expectedPercent: 0.0,
		},
		{
			name: "full_memory_usage",
			input: &metrics.ProcessMemoryInput{
				PID:               2,
				Name:              "memhog",
				RSS:               1000,
				TotalSystemMemory: 1000,
			},
			expectedPercent: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := metrics.NewProcessMemory(tt.input)

			assert.InDelta(t, tt.expectedPercent, mem.UsagePercent, 0.001)
			assert.Equal(t, tt.input.PID, mem.PID)
			assert.Equal(t, tt.input.Name, mem.Name)
			assert.Equal(t, tt.input.RSS, mem.RSS)
			assert.Equal(t, tt.input.VMS, mem.VMS)
			assert.Equal(t, tt.input.Shared, mem.Shared)
			assert.Equal(t, tt.input.Swap, mem.Swap)
			assert.Equal(t, tt.input.Data, mem.Data)
			assert.Equal(t, tt.input.Stack, mem.Stack)
			assert.False(t, mem.Timestamp.IsZero())
		})
	}
}

// TestProcessMemory_TotalResident tests the TotalResident method.
func TestProcessMemory_TotalResident(t *testing.T) {
	tests := []struct {
		name     string
		proc     metrics.ProcessMemory
		expected uint64
	}{
		{
			name: "rss_and_swap",
			proc: metrics.ProcessMemory{
				PID:  1234,
				Name: "test",
				RSS:  100 * 1024 * 1024, // 100 MB
				Swap: 10 * 1024 * 1024,  // 10 MB
			},
			expected: 110 * 1024 * 1024,
		},
		{
			name: "no_swap",
			proc: metrics.ProcessMemory{
				PID:  5678,
				Name: "worker",
				RSS:  50 * 1024 * 1024, // 50 MB
				Swap: 0,
			},
			expected: 50 * 1024 * 1024,
		},
		{
			name: "only_swap",
			proc: metrics.ProcessMemory{
				PID:  9999,
				Name: "swapped",
				RSS:  0,
				Swap: 25 * 1024 * 1024, // 25 MB
			},
			expected: 25 * 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := tt.proc.TotalResident()
			assert.Equal(t, tt.expected, total)
		})
	}
}

// TestProcessMemory_Fields tests all fields on ProcessMemory.
func TestProcessMemory_Fields(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		proc         metrics.ProcessMemory
		expectedPID  int
		expectedName string
		expectedRSS  uint64
		expectedVMS  uint64
	}{
		{
			name: "standard_process",
			proc: metrics.ProcessMemory{
				PID:          1234,
				Name:         "myprocess",
				RSS:          100 * 1024 * 1024,
				VMS:          500 * 1024 * 1024,
				Shared:       20 * 1024 * 1024,
				Swap:         5 * 1024 * 1024,
				Data:         50 * 1024 * 1024,
				Stack:        8 * 1024 * 1024,
				UsagePercent: 2.5,
				Timestamp:    now,
			},
			expectedPID:  1234,
			expectedName: "myprocess",
			expectedRSS:  100 * 1024 * 1024,
			expectedVMS:  500 * 1024 * 1024,
		},
		{
			name: "init_process",
			proc: metrics.ProcessMemory{
				PID:          1,
				Name:         "init",
				RSS:          10 * 1024 * 1024,
				VMS:          50 * 1024 * 1024,
				UsagePercent: 0.1,
				Timestamp:    now,
			},
			expectedPID:  1,
			expectedName: "init",
			expectedRSS:  10 * 1024 * 1024,
			expectedVMS:  50 * 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedPID, tt.proc.PID)
			assert.Equal(t, tt.expectedName, tt.proc.Name)
			assert.Equal(t, tt.expectedRSS, tt.proc.RSS)
			assert.Equal(t, tt.expectedVMS, tt.proc.VMS)
		})
	}
}
