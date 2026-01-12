// Package metrics provides domain value objects for system and process metrics.
package metrics

import "time"

// ProcessState represents the current lifecycle state of a process.
type ProcessState string

// Process state constants.
const (
	ProcessStateRunning  ProcessState = "running"
	ProcessStateStopped  ProcessState = "stopped"
	ProcessStateFailed   ProcessState = "failed"
	ProcessStateStarting ProcessState = "starting"
	ProcessStateStopping ProcessState = "stopping"
)

// ProcessMetrics aggregates CPU and memory metrics for a supervised process.
// It provides a unified view of resource usage correlated with lifecycle state.
type ProcessMetrics struct {
	// ServiceName is the name from the service configuration.
	ServiceName string
	// PID is the current process ID (0 if not running).
	PID int
	// State is the current lifecycle state.
	State ProcessState
	// Healthy indicates the overall health status.
	Healthy bool
	// CPU contains CPU time metrics for the process.
	CPU ProcessCPU
	// Memory contains memory usage metrics for the process.
	Memory ProcessMemory
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

// IsRunning returns true if the process is currently running.
func (m ProcessMetrics) IsRunning() bool {
	return m.State == ProcessStateRunning
}

// IsTerminal returns true if the process is in a terminal state (stopped or failed).
func (m ProcessMetrics) IsTerminal() bool {
	return m.State == ProcessStateStopped || m.State == ProcessStateFailed
}
