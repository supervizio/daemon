// Package metrics provides domain value objects for system and process metrics.
package metrics

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/process"
)

// ProcessMetricsParams contains parameters for creating ProcessMetrics instances.
// This struct groups all process metrics to avoid excessive constructor parameters.
type ProcessMetricsParams struct {
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
