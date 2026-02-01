// Package metrics provides domain value objects for system and process metrics.
package metrics

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/process"
)

// ProcessMetrics aggregates CPU, memory, and I/O metrics for a supervised process.
//
// This value object provides a unified view of resource usage correlated with
// lifecycle state for monitoring supervised processes.
type ProcessMetrics struct {
	// ServiceName is the name from the service configuration.
	ServiceName string
	// PID is the current process ID (0 if not running).
	PID int
	// State is the current lifecycle state.
	State process.State
	// Healthy indicates the overall health status.
	Healthy bool
	// CPU contains CPU time metrics for the process.
	CPU ProcessCPU
	// Memory contains memory usage metrics for the process.
	Memory ProcessMemory
	// NumFDs is the number of open file descriptors for the process.
	NumFDs uint32
	// ReadBytesPerSec is the disk read rate in bytes per second.
	ReadBytesPerSec uint64
	// WriteBytesPerSec is the disk write rate in bytes per second.
	WriteBytesPerSec uint64
	// StartTime is when the current process instance started.
	StartTime time.Time
	// Uptime is the duration since StartTime.
	Uptime time.Duration
	// RestartCount is the number of times this service has been restarted.
	RestartCount int
	// LastError contains the last error message if State is failed.
	LastError string
	// Timestamp is when these metrics were collected.
	Timestamp time.Time
}

// NewProcessMetrics creates a new ProcessMetrics instance.
//
// Params:
//   - params: ProcessMetricsParams containing all process metrics
//
// Returns:
//   - *ProcessMetrics: the created ProcessMetrics instance
func NewProcessMetrics(params *ProcessMetricsParams) *ProcessMetrics {
	// initialize with all process metrics fields
	return &ProcessMetrics{
		ServiceName:      params.ServiceName,
		PID:              params.PID,
		State:            params.State,
		Healthy:          params.Healthy,
		CPU:              params.CPU,
		Memory:           params.Memory,
		NumFDs:           params.NumFDs,
		ReadBytesPerSec:  params.ReadBytesPerSec,
		WriteBytesPerSec: params.WriteBytesPerSec,
		StartTime:        params.StartTime,
		Uptime:           params.Uptime,
		RestartCount:     params.RestartCount,
		LastError:        params.LastError,
		Timestamp:        params.Timestamp,
	}
}

// IsRunning returns true if the process is currently running.
//
// Returns:
//   - bool: true if the process is in a running state.
func (m *ProcessMetrics) IsRunning() bool {
	// delegate to state's IsRunning method
	return m.State.IsRunning()
}

// IsTerminal returns true if the process is in a terminal state (stopped or failed).
//
// Returns:
//   - bool: true if the process is stopped or failed.
func (m *ProcessMetrics) IsTerminal() bool {
	// delegate to state's IsTerminal method
	return m.State.IsTerminal()
}
