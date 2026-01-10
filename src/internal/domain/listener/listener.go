// Package listener provides domain entities for network listeners.
package listener

import (
	"net"
	"strconv"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// Listener represents a network listener endpoint.
// It tracks the listener's state and associated probe configuration.
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

	// ProbeConfig contains the probe configuration for this listener.
	// May be nil if no probing is configured.
	ProbeConfig *probe.Config

	// ProbeType specifies the probe type for this listener.
	// Examples: "tcp", "http", "grpc".
	ProbeType string

	// ProbeTarget contains additional probe target configuration.
	// For HTTP: path, method, status code.
	// For gRPC: service name.
	ProbeTarget probe.Target
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
//   - *Listener: a new listener in Closed state.
func NewListener(name, protocol, address string, port int) *Listener {
	// Return new listener with Closed state.
	return &Listener{
		Name:     name,
		Protocol: protocol,
		Address:  address,
		Port:     port,
		State:    Closed,
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
//   - *Listener: a new TCP listener in Closed state.
func NewTCP(name, address string, port int) *Listener {
	// Return new TCP listener.
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
//   - *Listener: a new UDP listener in Closed state.
func NewUDP(name, address string, port int) *Listener {
	// Return new UDP listener.
	return NewListener(name, "udp", address, port)
}

// WithProbe configures probing for this listener.
//
// Params:
//   - probeType: the probe type (tcp, http, grpc).
//   - config: the probe configuration.
//   - target: the probe target configuration.
//
// Returns:
//   - *Listener: the listener with probe configured.
func (l *Listener) WithProbe(probeType string, config probe.Config, target probe.Target) *Listener {
	// Set probe configuration.
	l.ProbeType = probeType
	l.ProbeConfig = &config
	l.ProbeTarget = target
	// Return self for chaining.
	return l
}

// SetState transitions the listener to a new state.
//
// Params:
//   - state: the new state to transition to.
//
// Returns:
//   - bool: true if the transition was valid and applied.
func (l *Listener) SetState(state State) bool {
	// Check if transition is valid.
	if !l.State.CanTransitionTo(state) {
		// Return false if transition is invalid.
		return false
	}
	// Apply the new state.
	l.State = state
	// Return true for successful transition.
	return true
}

// MarkListening transitions the listener to Listening state.
//
// Returns:
//   - bool: true if transition was successful.
func (l *Listener) MarkListening() bool {
	// Transition to Listening state.
	return l.SetState(Listening)
}

// MarkReady transitions the listener to Ready state.
//
// Returns:
//   - bool: true if transition was successful.
func (l *Listener) MarkReady() bool {
	// Transition to Ready state.
	return l.SetState(Ready)
}

// MarkClosed transitions the listener to Closed state.
//
// Returns:
//   - bool: true if transition was successful.
func (l *Listener) MarkClosed() bool {
	// Transition to Closed state.
	return l.SetState(Closed)
}

// HasProbe returns true if probing is configured.
//
// Returns:
//   - bool: true if ProbeConfig is not nil.
func (l *Listener) HasProbe() bool {
	// Return true if probe config exists.
	return l.ProbeConfig != nil
}

// ProbeAddress returns the address for probing.
// It combines the listener's address and port for network probes.
// Uses net.JoinHostPort for proper IPv6 address handling.
//
// Returns:
//   - string: the address in host:port format.
func (l *Listener) ProbeAddress() string {
	// Normalize non-routable bind addresses to loopback for probing.
	addr := l.Address

	// Check if address is non-routable and needs normalization.
	switch addr {
	// Empty address defaults to IPv4 loopback.
	case "":
		addr = "127.0.0.1"
	// IPv4 any-address defaults to IPv4 loopback.
	case "0.0.0.0":
		addr = "127.0.0.1"
	// IPv6 any-address defaults to IPv6 loopback.
	case "::":
		addr = "::1"
	}

	// Use net.JoinHostPort for proper IPv6 address handling.
	return net.JoinHostPort(addr, formatPort(l.Port))
}

// formatPort converts a port number to string.
//
// Params:
//   - port: the port number to format.
//
// Returns:
//   - string: the port as a decimal string.
func formatPort(port int) string {
	// Use strconv for efficient integer to string conversion.
	return strconv.Itoa(port)
}
