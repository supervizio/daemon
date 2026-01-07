// Package supervisor provides the application service for orchestrating multiple services.
package supervisor

import domain "github.com/kodflow/daemon/internal/domain/process"

// ServiceInfo contains information about a managed service.
// It provides runtime details including the process state, PID, and uptime.
type ServiceInfo struct {
	// Name is the service name.
	Name string
	// State is the current process state.
	State domain.State
	// PID is the process ID.
	PID int
	// Uptime is the uptime in seconds.
	Uptime int64
}
