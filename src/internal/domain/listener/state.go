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
	// match state to string representation
	switch s {
	// closed state
	case StateClosed:
		// return closed state name
		return "closed"
	// listening state
	case StateListening:
		// return listening state name
		return "listening"
	// ready state
	case StateReady:
		// return ready state name
		return "ready"
	// unknown state
	default:
		// return unknown for unmapped states
		return "unknown"
	}
}

// IsClosed returns true if the listener is closed.
//
// Returns:
//   - bool: true if state is Closed.
func (s State) IsClosed() bool {
	// return closed state check
	return s == StateClosed
}

// IsListening returns true if the listener is open.
//
// Returns:
//   - bool: true if state is Listening or Ready.
func (s State) IsListening() bool {
	// return listening or ready state check
	return s == StateListening || s == StateReady
}

// IsReady returns true if the listener is ready.
//
// Returns:
//   - bool: true if state is Ready.
func (s State) IsReady() bool {
	// return ready state check
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
	// check valid transitions based on current state
	switch s {
	// transitions from closed state
	case StateClosed:
		// from closed can only go to listening
		return target == StateListening
	// transitions from listening state
	case StateListening:
		// from listening can go to ready or closed
		return target == StateReady || target == StateClosed
	// transitions from ready state
	case StateReady:
		// from ready can go to listening or closed
		return target == StateListening || target == StateClosed
	// transitions from unknown state
	default:
		// unknown states cannot transition
		return false
	}
}
