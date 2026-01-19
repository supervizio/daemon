package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestSystemCPU_Total tests the Total method on SystemCPU.
func TestSystemCPU_Total(t *testing.T) {
	tests := []struct {
		name     string
		cpu      metrics.SystemCPU
		expected uint64
	}{
		{
			name: "all_fields_populated",
			cpu: metrics.SystemCPU{
				User:      1000,
				Nice:      100,
				System:    500,
				Idle:      8000,
				IOWait:    200,
				IRQ:       50,
				SoftIRQ:   50,
				Steal:     0,
				Guest:     0,
				GuestNice: 0,
			},
			expected: 9900,
		},
		{
			name: "only_user_and_system",
			cpu: metrics.SystemCPU{
				User:   5000,
				System: 2000,
			},
			expected: 7000,
		},
		{
			name: "zero_values",
			cpu: metrics.SystemCPU{
				User:   0,
				System: 0,
				Idle:   0,
			},
			expected: 0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate total.
			total := tt.cpu.Total()

			// Verify expected value.
			assert.Equal(t, tt.expected, total)
		})
	}
}

// TestSystemCPU_Active tests the Active method on SystemCPU.
func TestSystemCPU_Active(t *testing.T) {
	tests := []struct {
		name     string
		cpu      metrics.SystemCPU
		expected uint64
	}{
		{
			name: "excludes_idle_and_iowait",
			cpu: metrics.SystemCPU{
				User:   1000,
				Nice:   100,
				System: 500,
				Idle:   8000,
				IOWait: 200,
			},
			expected: 1000 + 100 + 500, // Total - Idle - IOWait
		},
		{
			name: "only_idle",
			cpu: metrics.SystemCPU{
				User:   0,
				System: 0,
				Idle:   10000,
				IOWait: 0,
			},
			expected: 0,
		},
		{
			name: "no_idle_time",
			cpu: metrics.SystemCPU{
				User:   3000,
				System: 2000,
				Idle:   0,
				IOWait: 0,
			},
			expected: 5000,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate active time.
			active := tt.cpu.Active()

			// Verify expected value.
			assert.Equal(t, tt.expected, active)
		})
	}
}

// TestSystemCPU_Timestamp tests the Timestamp field on SystemCPU.
func TestSystemCPU_Timestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp time.Time
	}{
		{
			name:      "current_time",
			timestamp: time.Now(),
		},
		{
			name:      "zero_time",
			timestamp: time.Time{},
		},
		{
			name:      "past_time",
			timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CPU with timestamp.
			cpu := metrics.SystemCPU{
				Timestamp: tt.timestamp,
			}

			// Verify timestamp.
			assert.Equal(t, tt.timestamp, cpu.Timestamp)
		})
	}
}

// TestNewSystemCPU tests the NewSystemCPU constructor.
func TestNewSystemCPU(t *testing.T) {
	tests := []struct {
		name   string
		params *metrics.SystemCPUParams
	}{
		{
			name: "all_fields_populated",
			params: &metrics.SystemCPUParams{
				User:         1000,
				Nice:         100,
				System:       500,
				Idle:         8000,
				IOWait:       200,
				IRQ:          50,
				SoftIRQ:      50,
				Steal:        10,
				Guest:        5,
				GuestNice:    1,
				UsagePercent: 25.5,
				Timestamp:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "zero_values",
			params: &metrics.SystemCPUParams{
				User:         0,
				Nice:         0,
				System:       0,
				Idle:         0,
				IOWait:       0,
				IRQ:          0,
				SoftIRQ:      0,
				Steal:        0,
				Guest:        0,
				GuestNice:    0,
				UsagePercent: 0,
				Timestamp:    time.Time{},
			},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create SystemCPU using constructor.
			cpu := metrics.NewSystemCPU(tt.params)

			// Verify all fields are correctly set.
			assert.NotNil(t, cpu)
			assert.Equal(t, tt.params.User, cpu.User)
			assert.Equal(t, tt.params.Nice, cpu.Nice)
			assert.Equal(t, tt.params.System, cpu.System)
			assert.Equal(t, tt.params.Idle, cpu.Idle)
			assert.Equal(t, tt.params.IOWait, cpu.IOWait)
			assert.Equal(t, tt.params.IRQ, cpu.IRQ)
			assert.Equal(t, tt.params.SoftIRQ, cpu.SoftIRQ)
			assert.Equal(t, tt.params.Steal, cpu.Steal)
			assert.Equal(t, tt.params.Guest, cpu.Guest)
			assert.Equal(t, tt.params.GuestNice, cpu.GuestNice)
			assert.Equal(t, tt.params.UsagePercent, cpu.UsagePercent)
			assert.Equal(t, tt.params.Timestamp, cpu.Timestamp)
		})
	}
}
