// Package metrics_test provides external tests for the metrics domain package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

func TestProcessMetrics_IsRunning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state metrics.ProcessState
		want  bool
	}{
		{"running", metrics.ProcessStateRunning, true},
		{"stopped", metrics.ProcessStateStopped, false},
		{"failed", metrics.ProcessStateFailed, false},
		{"starting", metrics.ProcessStateStarting, false},
		{"stopping", metrics.ProcessStateStopping, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := metrics.ProcessMetrics{State: tt.state}
			assert.Equal(t, tt.want, m.IsRunning())
		})
	}
}

func TestProcessMetrics_IsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state metrics.ProcessState
		want  bool
	}{
		{"running", metrics.ProcessStateRunning, false},
		{"stopped", metrics.ProcessStateStopped, true},
		{"failed", metrics.ProcessStateFailed, true},
		{"starting", metrics.ProcessStateStarting, false},
		{"stopping", metrics.ProcessStateStopping, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := metrics.ProcessMetrics{State: tt.state}
			assert.Equal(t, tt.want, m.IsTerminal())
		})
	}
}

func TestProcessMetrics_Fields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	m := metrics.ProcessMetrics{
		ServiceName:  "test-service",
		PID:          1234,
		State:        metrics.ProcessStateRunning,
		Healthy:      true,
		CPU:          metrics.ProcessCPU{User: 100, System: 50},
		Memory:       metrics.ProcessMemory{RSS: 1024 * 1024},
		StartTime:    now,
		Uptime:       5 * time.Minute,
		RestartCount: 2,
		LastError:    "previous failure",
		Timestamp:    now,
	}

	assert.Equal(t, "test-service", m.ServiceName)
	assert.Equal(t, 1234, m.PID)
	assert.Equal(t, metrics.ProcessStateRunning, m.State)
	assert.True(t, m.Healthy)
	assert.Equal(t, uint64(100), m.CPU.User)
	assert.Equal(t, uint64(1024*1024), m.Memory.RSS)
	assert.Equal(t, now, m.StartTime)
	assert.Equal(t, 5*time.Minute, m.Uptime)
	assert.Equal(t, 2, m.RestartCount)
	assert.Equal(t, "previous failure", m.LastError)
}
