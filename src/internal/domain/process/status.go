// Package process provides domain entities and value objects for process lifecycle management.
package process

import "time"

// Status represents the status of a managed process.
//
// Status provides a snapshot of the current state of a managed process
// including identification, runtime metrics, and exit information.
type Status struct {
	// Name is the service name.
	Name string
	// State is the current process state.
	State State
	// PID is the process ID.
	PID int
	// Uptime is how long the process has been running.
	Uptime time.Duration
	// Restarts is the number of times the process has restarted.
	Restarts int
	// ExitCode is the last exit code.
	ExitCode int
}
