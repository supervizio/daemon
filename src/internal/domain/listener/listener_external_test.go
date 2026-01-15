// Package listener_test provides black-box tests for the listener package.
package listener_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/listener"
)

// TestNewListener tests listener creation.
func TestNewListener(t *testing.T) {
	tests := []struct {
		name             string
		listenerName     string
		protocol         string
		address          string
		port             int
		expectedState    listener.State
		expectedProtocol string
	}{
		{
			name:             "tcp_listener",
			listenerName:     "http",
			protocol:         "tcp",
			address:          "localhost",
			port:             8080,
			expectedState:    listener.StateClosed,
			expectedProtocol: "tcp",
		},
		{
			name:             "udp_listener",
			listenerName:     "dns",
			protocol:         "udp",
			address:          "0.0.0.0",
			port:             53,
			expectedState:    listener.StateClosed,
			expectedProtocol: "udp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener.
			l := listener.NewListener(tt.listenerName, tt.protocol, tt.address, tt.port)

			// Verify fields.
			require.NotNil(t, l)
			assert.Equal(t, tt.listenerName, l.Name)
			assert.Equal(t, tt.expectedProtocol, l.Protocol)
			assert.Equal(t, tt.address, l.Address)
			assert.Equal(t, tt.port, l.Port)
			assert.Equal(t, tt.expectedState, l.State)
		})
	}
}

// TestNewTCP tests TCP listener creation.
func TestNewTCP(t *testing.T) {
	tests := []struct {
		name         string
		listenerName string
		address      string
		port         int
	}{
		{
			name:         "localhost",
			listenerName: "http",
			address:      "localhost",
			port:         8080,
		},
		{
			name:         "any_address",
			listenerName: "admin",
			address:      "0.0.0.0",
			port:         9090,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create TCP listener.
			l := listener.NewTCP(tt.listenerName, tt.address, tt.port)

			// Verify fields.
			require.NotNil(t, l)
			assert.Equal(t, "tcp", l.Protocol)
			assert.Equal(t, tt.listenerName, l.Name)
		})
	}
}

// TestNewUDP tests UDP listener creation.
func TestNewUDP(t *testing.T) {
	tests := []struct {
		name         string
		listenerName string
		address      string
		port         int
	}{
		{
			name:         "dns",
			listenerName: "dns",
			address:      "localhost",
			port:         53,
		},
		{
			name:         "ntp",
			listenerName: "ntp",
			address:      "0.0.0.0",
			port:         123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create UDP listener.
			l := listener.NewUDP(tt.listenerName, tt.address, tt.port)

			// Verify fields.
			require.NotNil(t, l)
			assert.Equal(t, "udp", l.Protocol)
			assert.Equal(t, tt.listenerName, l.Name)
		})
	}
}

// TestListener_SetState tests state transitions.
func TestListener_SetState(t *testing.T) {
	tests := []struct {
		name          string
		initialState  listener.State
		targetState   listener.State
		shouldSucceed bool
	}{
		{
			name:          "closed_to_listening",
			initialState:  listener.StateClosed,
			targetState:   listener.StateListening,
			shouldSucceed: true,
		},
		{
			name:          "closed_to_ready_invalid",
			initialState:  listener.StateClosed,
			targetState:   listener.StateReady,
			shouldSucceed: false,
		},
		{
			name:          "listening_to_ready",
			initialState:  listener.StateListening,
			targetState:   listener.StateReady,
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener and set initial state.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			l.State = tt.initialState

			// Attempt transition.
			result := l.SetState(tt.targetState)

			// Verify result.
			assert.Equal(t, tt.shouldSucceed, result)
			if tt.shouldSucceed {
				assert.Equal(t, tt.targetState, l.State)
			} else {
				assert.Equal(t, tt.initialState, l.State)
			}
		})
	}
}

// TestListener_MarkListening tests MarkListening method.
func TestListener_MarkListening(t *testing.T) {
	tests := []struct {
		name          string
		initialState  listener.State
		shouldSucceed bool
		expectedState listener.State
	}{
		{
			name:          "from_closed",
			initialState:  listener.StateClosed,
			shouldSucceed: true,
			expectedState: listener.StateListening,
		},
		{
			name:          "from_ready_valid",
			initialState:  listener.StateReady,
			shouldSucceed: true,
			expectedState: listener.StateListening,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener with initial state.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			l.State = tt.initialState

			// Mark as listening.
			result := l.MarkListening()

			// Verify result.
			assert.Equal(t, tt.shouldSucceed, result)
			assert.Equal(t, tt.expectedState, l.State)
		})
	}
}

// TestListener_MarkReady tests MarkReady method.
func TestListener_MarkReady(t *testing.T) {
	tests := []struct {
		name          string
		initialState  listener.State
		shouldSucceed bool
		expectedState listener.State
	}{
		{
			name:          "from_listening",
			initialState:  listener.StateListening,
			shouldSucceed: true,
			expectedState: listener.StateReady,
		},
		{
			name:          "from_closed_invalid",
			initialState:  listener.StateClosed,
			shouldSucceed: false,
			expectedState: listener.StateClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener with initial state.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			l.State = tt.initialState

			// Mark as ready.
			result := l.MarkReady()

			// Verify result.
			assert.Equal(t, tt.shouldSucceed, result)
			assert.Equal(t, tt.expectedState, l.State)
		})
	}
}

// TestListener_MarkClosed tests MarkClosed method.
func TestListener_MarkClosed(t *testing.T) {
	tests := []struct {
		name          string
		initialState  listener.State
		shouldSucceed bool
		expectedState listener.State
	}{
		{
			name:          "from_listening",
			initialState:  listener.StateListening,
			shouldSucceed: true,
			expectedState: listener.StateClosed,
		},
		{
			name:          "from_ready",
			initialState:  listener.StateReady,
			shouldSucceed: true,
			expectedState: listener.StateClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener with initial state.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			l.State = tt.initialState

			// Mark as closed.
			result := l.MarkClosed()

			// Verify result.
			assert.Equal(t, tt.shouldSucceed, result)
			assert.Equal(t, tt.expectedState, l.State)
		})
	}
}
