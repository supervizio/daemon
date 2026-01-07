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
	// Map each state constant to its corresponding string representation.
	switch s {
	// Handle the stopped state.
	case StateStopped:
		// Return the string for stopped state.
		return "stopped"
	// Handle the starting state.
	case StateStarting:
		// Return the string for starting state.
		return "starting"
	// Handle the running state.
	case StateRunning:
		// Return the string for running state.
		return "running"
	// Handle the stopping state.
	case StateStopping:
		// Return the string for stopping state.
		return "stopping"
	// Handle the failed state.
	case StateFailed:
		// Return the string for failed state.
		return "failed"
	// Handle any unknown or invalid state values.
	default:
		// Return unknown for unrecognized states.
		return "unknown"
	}
}

// IsTerminal returns true if the state is a terminal state (stopped or failed).
//
// Returns:
//   - bool: true if the process has reached a terminal state
func (s State) IsTerminal() bool {
	// Check if the state is either stopped or failed, both are terminal states.
	return s == StateStopped || s == StateFailed
}

// IsActive returns true if the process is starting or running.
//
// Returns:
//   - bool: true if the process is currently active
func (s State) IsActive() bool {
	// Check if the state indicates the process is actively running or starting up.
	return s == StateStarting || s == StateRunning
}
