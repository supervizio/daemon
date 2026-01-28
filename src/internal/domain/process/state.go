// Package process provides domain entities and value objects for process lifecycle management.
package process

// State represents the current state of a process.
// It is used to track the lifecycle of a managed process through its various phases.
type State int

// Process state constants define the possible lifecycle states.
const (
	// StateStopped indicates the process is not running and has been stopped normally.
	StateStopped State = iota
	// StateStarting indicates the process is in the process of starting up.
	StateStarting
	// StateRunning indicates the process is currently executing normally.
	StateRunning
	// StateStopping indicates the process is in the process of shutting down gracefully.
	StateStopping
	// StateFailed indicates the process has terminated with an error or non-zero exit code.
	StateFailed
)

// String returns the string representation of the State.
//
// Returns:
//   - string: human-readable state name
func (s State) String() string {
	switch s {
	case StateStopped:
		return "stopped"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// IsTerminal returns true if the state is a terminal state (stopped or failed).
//
// Returns:
//   - bool: true if the process has reached a terminal state
func (s State) IsTerminal() bool {
	return s == StateStopped || s == StateFailed
}

// IsActive returns true if the process is starting or running.
//
// Returns:
//   - bool: true if the process is currently active
func (s State) IsActive() bool {
	return s == StateStarting || s == StateRunning
}

// IsRunning returns true if the process is in running state.
//
// Returns:
//   - bool: true if the process is currently running
func (s State) IsRunning() bool {
	return s == StateRunning
}

// IsStopping returns true if the process is stopping.
//
// Returns:
//   - bool: true if the process is currently stopping
func (s State) IsStopping() bool {
	return s == StateStopping
}

// IsStarting returns true if the process is starting.
//
// Returns:
//   - bool: true if the process is currently starting
func (s State) IsStarting() bool {
	return s == StateStarting
}

// IsFailed returns true if the process has failed.
//
// Returns:
//   - bool: true if the process is in failed state
func (s State) IsFailed() bool {
	return s == StateFailed
}

// IsStopped returns true if the process is stopped.
//
// Returns:
//   - bool: true if the process is in stopped state
func (s State) IsStopped() bool {
	return s == StateStopped
}
