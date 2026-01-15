// Package metrics_test provides external tests for the metrics domain package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
)

// TestNewProcessMetrics tests the NewProcessMetrics constructor.
func TestNewProcessMetrics(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []struct {
		name   string
		params *metrics.ProcessMetricsParams
	}{
		{
			name: "all_fields_populated",
			params: &metrics.ProcessMetricsParams{
				ServiceName:  "test-service",
				PID:          1234,
				State:        process.StateRunning,
				Healthy:      true,
				CPU:          metrics.ProcessCPU{User: 100, System: 50},
				Memory:       metrics.ProcessMemory{RSS: 1024 * 1024},
				StartTime:    now,
				Uptime:       5 * time.Minute,
				RestartCount: 2,
				LastError:    "previous failure",
				Timestamp:    now,
			},
		},
		{
			name: "stopped_process",
			params: &metrics.ProcessMetricsParams{
				ServiceName:  "stopped-service",
				PID:          0,
				State:        process.StateStopped,
				Healthy:      false,
				CPU:          metrics.ProcessCPU{},
				Memory:       metrics.ProcessMemory{},
				StartTime:    time.Time{},
				Uptime:       0,
				RestartCount: 0,
				LastError:    "",
				Timestamp:    now,
			},
		},
		{
			name: "failed_process",
			params: &metrics.ProcessMetricsParams{
				ServiceName:  "failed-service",
				PID:          0,
				State:        process.StateFailed,
				Healthy:      false,
				CPU:          metrics.ProcessCPU{},
				Memory:       metrics.ProcessMemory{},
				StartTime:    time.Time{},
				Uptime:       0,
				RestartCount: 5,
				LastError:    "process exited with code 1",
				Timestamp:    now,
			},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create ProcessMetrics using constructor.
			m := metrics.NewProcessMetrics(tt.params)

			// Verify pointer is not nil.
			require.NotNil(t, m)

			// Verify all fields are correctly set.
			assert.Equal(t, tt.params.ServiceName, m.ServiceName)
			assert.Equal(t, tt.params.PID, m.PID)
			assert.Equal(t, tt.params.State, m.State)
			assert.Equal(t, tt.params.Healthy, m.Healthy)
			assert.Equal(t, tt.params.CPU.User, m.CPU.User)
			assert.Equal(t, tt.params.CPU.System, m.CPU.System)
			assert.Equal(t, tt.params.Memory.RSS, m.Memory.RSS)
			assert.Equal(t, tt.params.StartTime, m.StartTime)
			assert.Equal(t, tt.params.Uptime, m.Uptime)
			assert.Equal(t, tt.params.RestartCount, m.RestartCount)
			assert.Equal(t, tt.params.LastError, m.LastError)
			assert.Equal(t, tt.params.Timestamp, m.Timestamp)
		})
	}
}

// TestProcessMetrics_IsRunning tests the IsRunning method on ProcessMetrics.
func TestProcessMetrics_IsRunning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state process.State
		want  bool
	}{
		{"running", process.StateRunning, true},
		{"stopped", process.StateStopped, false},
		{"failed", process.StateFailed, false},
		{"starting", process.StateStarting, false},
		{"stopping", process.StateStopping, false},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create ProcessMetrics with the specified state.
			m := metrics.ProcessMetrics{State: tt.state}
			// Check IsRunning result.
			assert.Equal(t, tt.want, m.IsRunning())
		})
	}
}

// TestProcessMetrics_IsTerminal tests the IsTerminal method on ProcessMetrics.
func TestProcessMetrics_IsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state process.State
		want  bool
	}{
		{"running", process.StateRunning, false},
		{"stopped", process.StateStopped, true},
		{"failed", process.StateFailed, true},
		{"starting", process.StateStarting, false},
		{"stopping", process.StateStopping, false},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create ProcessMetrics with the specified state.
			m := metrics.ProcessMetrics{State: tt.state}
			// Check IsTerminal result.
			assert.Equal(t, tt.want, m.IsTerminal())
		})
	}
}

// TestProcessMetrics_Fields tests that ProcessMetrics fields are accessible.
func TestProcessMetrics_Fields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []struct {
		name         string
		metrics      metrics.ProcessMetrics
		wantService  string
		wantPID      int
		wantState    process.State
		wantHealthy  bool
		wantCPUUser  uint64
		wantMemRSS   uint64
		wantUptime   time.Duration
		wantRestarts int
		wantError    string
	}{
		{
			name: "running_service",
			metrics: metrics.ProcessMetrics{
				ServiceName:  "test-service",
				PID:          1234,
				State:        process.StateRunning,
				Healthy:      true,
				CPU:          metrics.ProcessCPU{User: 100, System: 50},
				Memory:       metrics.ProcessMemory{RSS: 1024 * 1024},
				StartTime:    now,
				Uptime:       5 * time.Minute,
				RestartCount: 2,
				LastError:    "previous failure",
				Timestamp:    now,
			},
			wantService:  "test-service",
			wantPID:      1234,
			wantState:    process.StateRunning,
			wantHealthy:  true,
			wantCPUUser:  100,
			wantMemRSS:   1024 * 1024,
			wantUptime:   5 * time.Minute,
			wantRestarts: 2,
			wantError:    "previous failure",
		},
		{
			name: "stopped_service",
			metrics: metrics.ProcessMetrics{
				ServiceName:  "stopped-service",
				PID:          0,
				State:        process.StateStopped,
				Healthy:      false,
				CPU:          metrics.ProcessCPU{User: 0, System: 0},
				Memory:       metrics.ProcessMemory{RSS: 0},
				StartTime:    time.Time{},
				Uptime:       0,
				RestartCount: 0,
				LastError:    "",
				Timestamp:    now,
			},
			wantService:  "stopped-service",
			wantPID:      0,
			wantState:    process.StateStopped,
			wantHealthy:  false,
			wantCPUUser:  0,
			wantMemRSS:   0,
			wantUptime:   0,
			wantRestarts: 0,
			wantError:    "",
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Verify all field values.
			assert.Equal(t, tt.wantService, tt.metrics.ServiceName)
			assert.Equal(t, tt.wantPID, tt.metrics.PID)
			assert.Equal(t, tt.wantState, tt.metrics.State)
			assert.Equal(t, tt.wantHealthy, tt.metrics.Healthy)
			assert.Equal(t, tt.wantCPUUser, tt.metrics.CPU.User)
			assert.Equal(t, tt.wantMemRSS, tt.metrics.Memory.RSS)
			assert.Equal(t, tt.wantUptime, tt.metrics.Uptime)
			assert.Equal(t, tt.wantRestarts, tt.metrics.RestartCount)
			assert.Equal(t, tt.wantError, tt.metrics.LastError)
		})
	}
}
