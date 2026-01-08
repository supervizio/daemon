// Package listener provides domain entities for network listeners.
// It defines listener state and transitions for health tracking.
package listener

// State represents the state of a network listener.
// It tracks the lifecycle from closed to ready.
type State int

// Listener states.
const (
	// Closed indicates the listener port is not open.
	// This is the initial state before the service starts.
	Closed State = iota

	// Listening indicates the port is open and accepting connections.
	// The service has started but health checks may not have passed yet.
	Listening

	// Ready indicates health checks have passed.
	// The listener is fully operational and ready for traffic.
	Ready
)

// String returns the string representation of the state.
//
// Returns:
//   - string: the human-readable state name.
func (s State) String() string {
	// Switch on state value to return the corresponding string.
	switch s {
	// Case for closed state.
	case Closed:
		// Return the closed state string.
		return "closed"
	// Case for listening state.
	case Listening:
		// Return the listening state string.
		return "listening"
	// Case for ready state.
	case Ready:
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
	// Return true if state equals Closed.
	return s == Closed
}

// IsListening returns true if the listener is open.
//
// Returns:
//   - bool: true if state is Listening or Ready.
func (s State) IsListening() bool {
	// Return true if state is Listening or Ready.
	return s == Listening || s == Ready
}

// IsReady returns true if the listener is ready.
//
// Returns:
//   - bool: true if state is Ready.
func (s State) IsReady() bool {
	// Return true if state equals Ready.
	return s == Ready
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
	// From Closed, can go to Listening.
	case Closed:
		// Return true only for Listening.
		return target == Listening
	// From Listening, can go to Ready or back to Closed.
	case Listening:
		// Return true for Ready or Closed.
		return target == Ready || target == Closed
	// From Ready, can go back to Listening or Closed.
	case Ready:
		// Return true for Listening or Closed.
		return target == Listening || target == Closed
	// Default case for unknown states.
	default:
		// Return false for unknown states.
		return false
	}
}
