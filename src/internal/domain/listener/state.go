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
	switch s {
	case StateClosed:
		return "closed"
	case StateListening:
		return "listening"
	case StateReady:
		return "ready"
	default:
		return "unknown"
	}
}

// IsClosed returns true if the listener is closed.
//
// Returns:
//   - bool: true if state is Closed.
func (s State) IsClosed() bool {
	return s == StateClosed
}

// IsListening returns true if the listener is open.
//
// Returns:
//   - bool: true if state is Listening or Ready.
func (s State) IsListening() bool {
	return s == StateListening || s == StateReady
}

// IsReady returns true if the listener is ready.
//
// Returns:
//   - bool: true if state is Ready.
func (s State) IsReady() bool {
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
	switch s {
	case StateClosed:
		return target == StateListening
	case StateListening:
		return target == StateReady || target == StateClosed
	case StateReady:
		return target == StateListening || target == StateClosed
	default:
		return false
	}
}
