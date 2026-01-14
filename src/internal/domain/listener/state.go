// Package listener provides domain entities for network listeners.
// It defines listener state and transitions for health tracking.
package listener

// State represents the state of a network listener.
// It tracks the lifecycle from closed to ready.
type State int

// Listener states.
const (
	// StateClosed indicates the listener port is not open.
	// This is the initial state before the service starts.
	StateClosed State = iota

	// StateListening indicates the port is open and accepting connections.
	// The service has started but health checks may not have passed yet.
	StateListening

	// StateReady indicates health checks have passed.
	// The listener is fully operational and ready for traffic.
	StateReady
)

// String returns the string representation of the state.
//
// Returns:
//   - string: the human-readable state name.
func (s State) String() string {
	// Switch on state value to return the corresponding string.
	switch s {
	// Case for closed state.
	case StateClosed:
		// Return the closed state string.
		return "closed"
	// Case for listening state.
	case StateListening:
		// Return the listening state string.
		return "listening"
	// Case for ready state.
	case StateReady:
		// Return the ready state string.
		return "ready"
	// Default case for unknown states.
	default:
		// Return unknown for unrecognized values.
		return "unknown"
	}
}

// IsClosed returns true if the listener is closed.
//
// Returns:
//   - bool: true if state is Closed.
func (s State) IsClosed() bool {
	// Return true if state equals StateClosed.
	return s == StateClosed
}

// IsListening returns true if the listener is open.
//
// Returns:
//   - bool: true if state is Listening or Ready.
func (s State) IsListening() bool {
	// Return true if state is StateListening or StateReady.
	return s == StateListening || s == StateReady
}

// IsReady returns true if the listener is ready.
//
// Returns:
//   - bool: true if state is Ready.
func (s State) IsReady() bool {
	// Return true if state equals StateReady.
	return s == StateReady
}

// CanTransitionTo checks if a transition to the target state is valid.
//
// Params:
//   - target: the target state to transition to.
//
// Returns:
//   - bool: true if the transition is valid.
func (s State) CanTransitionTo(target State) bool {
	// Define valid transitions.
	switch s {
	// From StateClosed, can go to StateListening.
	case StateClosed:
		// Return true only for StateListening.
		return target == StateListening
	// From StateListening, can go to StateReady or back to StateClosed.
	case StateListening:
		// Return true for StateReady or StateClosed.
		return target == StateReady || target == StateClosed
	// From StateReady, can go back to StateListening or StateClosed.
	case StateReady:
		// Return true for StateListening or StateClosed.
		return target == StateListening || target == StateClosed
	// Default case for unknown states.
	default:
		// Return false for unknown states.
		return false
	}
}
