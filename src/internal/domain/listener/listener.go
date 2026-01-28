// Package listener provides domain entities for network listeners.
package listener

// Listener represents a network listener endpoint.
// It tracks the listener's state using a finite state machine.
type Listener struct {
	// Name is the unique identifier for this listener.
	// Example: "http", "admin", "grpc".
	Name string

	// Protocol is the network protocol.
	// Supported values: "tcp", "udp".
	Protocol string

	// Address is the bind address.
	// Example: "0.0.0.0", "127.0.0.1", "".
	Address string

	// Port is the listen port number.
	Port int

	// State is the current listener state.
	State State
}

// NewListener creates a new listener with default state.
//
// Params:
//   - name: the unique identifier for this listener.
//   - protocol: the network protocol (tcp, udp).
//   - address: the bind address.
//   - port: the listen port number.
//
// Returns:
//   - *Listener: a new listener in StateClosed state.
func NewListener(name, protocol, address string, port int) *Listener {
	return &Listener{
		Name:     name,
		Protocol: protocol,
		Address:  address,
		Port:     port,
		State:    StateClosed,
	}
}

// NewTCP creates a new TCP listener.
//
// Params:
//   - name: the unique identifier for this listener.
//   - address: the bind address.
//   - port: the listen port number.
//
// Returns:
//   - *Listener: a new TCP listener in StateClosed state.
func NewTCP(name, address string, port int) *Listener {
	return NewListener(name, "tcp", address, port)
}

// NewUDP creates a new UDP listener.
//
// Params:
//   - name: the unique identifier for this listener.
//   - address: the bind address.
//   - port: the listen port number.
//
// Returns:
//   - *Listener: a new UDP listener in StateClosed state.
func NewUDP(name, address string, port int) *Listener {
	return NewListener(name, "udp", address, port)
}

// SetState transitions the listener to a new state.
//
// Params:
//   - state: the new state to transition to.
//
// Returns:
//   - bool: true if the transition was valid and applied.
func (l *Listener) SetState(state State) bool {
	if !l.State.CanTransitionTo(state) {
		return false
	}
	l.State = state
	return true
}

// MarkListening transitions the listener to StateListening state.
//
// Returns:
//   - bool: true if transition was successful.
func (l *Listener) MarkListening() bool {
	// Transition to StateListening state.
	return l.SetState(StateListening)
}

// MarkReady transitions the listener to StateReady state.
//
// Returns:
//   - bool: true if transition was successful.
func (l *Listener) MarkReady() bool {
	// Transition to StateReady state.
	return l.SetState(StateReady)
}

// MarkClosed transitions the listener to StateClosed state.
//
// Returns:
//   - bool: true if transition was successful.
func (l *Listener) MarkClosed() bool {
	// Transition to StateClosed state.
	return l.SetState(StateClosed)
}
