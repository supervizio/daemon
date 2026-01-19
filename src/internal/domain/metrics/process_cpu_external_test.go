package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestProcessCPU_Total tests the Total method on ProcessCPU.
func TestProcessCPU_Total(t *testing.T) {
	tests := []struct {
		name     string
		proc     metrics.ProcessCPU
		expected uint64
	}{
		{
			name: "user_and_system",
			proc: metrics.ProcessCPU{
				PID:    1234,
				Name:   "test",
				User:   500,
				System: 300,
			},
			expected: 800,
		},
		{
			name: "only_user",
			proc: metrics.ProcessCPU{
				PID:    5678,
				Name:   "worker",
				User:   1000,
				System: 0,
			},
			expected: 1000,
		},
		{
			name: "zero_values",
			proc: metrics.ProcessCPU{
				PID:    9999,
				Name:   "idle",
				User:   0,
				System: 0,
			},
			expected: 0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate total.
			total := tt.proc.Total()

			// Verify expected value.
			assert.Equal(t, tt.expected, total)
		})
	}
}

// TestProcessCPU_TotalWithChildren tests the TotalWithChildren method on ProcessCPU.
func TestProcessCPU_TotalWithChildren(t *testing.T) {
	tests := []struct {
		name     string
		proc     metrics.ProcessCPU
		expected uint64
	}{
		{
			name: "includes_children_time",
			proc: metrics.ProcessCPU{
				PID:            1234,
				Name:           "test",
				User:           500,
				System:         300,
				ChildrenUser:   100,
				ChildrenSystem: 50,
			},
			expected: 950,
		},
		{
			name: "no_children",
			proc: metrics.ProcessCPU{
				PID:            5678,
				Name:           "worker",
				User:           1000,
				System:         500,
				ChildrenUser:   0,
				ChildrenSystem: 0,
			},
			expected: 1500,
		},
		{
			name: "only_children_time",
			proc: metrics.ProcessCPU{
				PID:            9999,
				Name:           "parent",
				User:           0,
				System:         0,
				ChildrenUser:   200,
				ChildrenSystem: 100,
			},
			expected: 300,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate total with children.
			total := tt.proc.TotalWithChildren()

			// Verify expected value.
			assert.Equal(t, tt.expected, total)
		})
	}
}

// TestProcessCPU_Fields tests all fields on ProcessCPU.
func TestProcessCPU_Fields(t *testing.T) {
	tests := []struct {
		name         string
		proc         metrics.ProcessCPU
		expectedPID  int
		expectedName string
		expectedUser uint64
	}{
		{
			name: "standard_process",
			proc: metrics.ProcessCPU{
				PID:          1234,
				Name:         "myprocess",
				User:         100,
				System:       50,
				StartTime:    12345678,
				UsagePercent: 5.5,
				Timestamp:    time.Now(),
			},
			expectedPID:  1234,
			expectedName: "myprocess",
			expectedUser: 100,
		},
		{
			name: "init_process",
			proc: metrics.ProcessCPU{
				PID:          1,
				Name:         "init",
				User:         50000,
				System:       25000,
				StartTime:    0,
				UsagePercent: 0.1,
				Timestamp:    time.Now(),
			},
			expectedPID:  1,
			expectedName: "init",
			expectedUser: 50000,
		},
		{
			name: "high_usage_process",
			proc: metrics.ProcessCPU{
				PID:          42,
				Name:         "compute",
				User:         999999,
				System:       111111,
				StartTime:    12345678,
				UsagePercent: 99.9,
				Timestamp:    time.Now(),
			},
			expectedPID:  42,
			expectedName: "compute",
			expectedUser: 999999,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify fields.
			assert.Equal(t, tt.expectedPID, tt.proc.PID)
			assert.Equal(t, tt.expectedName, tt.proc.Name)
			assert.Equal(t, tt.expectedUser, tt.proc.User)
		})
	}
}

// TestNewProcessCPU tests the NewProcessCPU constructor.
func TestNewProcessCPU(t *testing.T) {
	tests := []struct {
		name   string
		params *metrics.ProcessCPUParams
	}{
		{
			name: "standard_process",
			params: &metrics.ProcessCPUParams{
				PID:            1234,
				Name:           "myprocess",
				User:           500,
				System:         300,
				ChildrenUser:   100,
				ChildrenSystem: 50,
				StartTime:      12345678,
				UsagePercent:   5.5,
				Timestamp:      time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "zero_values",
			params: &metrics.ProcessCPUParams{
				PID:            0,
				Name:           "",
				User:           0,
				System:         0,
				ChildrenUser:   0,
				ChildrenSystem: 0,
				StartTime:      0,
				UsagePercent:   0,
				Timestamp:      time.Time{},
			},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ProcessCPU using constructor.
			proc := metrics.NewProcessCPU(tt.params)

			// Verify all fields are correctly set.
			assert.NotNil(t, proc)
			assert.Equal(t, tt.params.PID, proc.PID)
			assert.Equal(t, tt.params.Name, proc.Name)
			assert.Equal(t, tt.params.User, proc.User)
			assert.Equal(t, tt.params.System, proc.System)
			assert.Equal(t, tt.params.ChildrenUser, proc.ChildrenUser)
			assert.Equal(t, tt.params.ChildrenSystem, proc.ChildrenSystem)
			assert.Equal(t, tt.params.StartTime, proc.StartTime)
			assert.Equal(t, tt.params.UsagePercent, proc.UsagePercent)
			assert.Equal(t, tt.params.Timestamp, proc.Timestamp)
		})
	}
}
