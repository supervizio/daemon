// Package listener_test provides black-box tests for the listener package.
package listener_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/listener"
	"github.com/kodflow/daemon/internal/domain/probe"
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
			expectedState:    listener.Closed,
			expectedProtocol: "tcp",
		},
		{
			name:             "udp_listener",
			listenerName:     "dns",
			protocol:         "udp",
			address:          "0.0.0.0",
			port:             53,
			expectedState:    listener.Closed,
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

// TestListener_WithProbe tests probe configuration.
func TestListener_WithProbe(t *testing.T) {
	tests := []struct {
		name      string
		probeType string
		config    probe.Config
		target    probe.Target
	}{
		{
			name:      "tcp_probe",
			probeType: "tcp",
			config:    probe.NewConfig(),
			target:    probe.NewTCPTarget("localhost:8080"),
		},
		{
			name:      "http_probe",
			probeType: "http",
			config:    probe.NewConfig().WithTimeout(10 * time.Second),
			target:    probe.NewHTTPTarget("http://localhost:8080/health", "GET", 200),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener with probe.
			l := listener.NewTCP("http", "localhost", 8080)
			l.WithProbe(tt.probeType, tt.config, tt.target)

			// Verify probe configuration.
			assert.True(t, l.HasProbe())
			assert.Equal(t, tt.probeType, l.ProbeType)
			require.NotNil(t, l.ProbeConfig)
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
			initialState:  listener.Closed,
			targetState:   listener.Listening,
			shouldSucceed: true,
		},
		{
			name:          "closed_to_ready_invalid",
			initialState:  listener.Closed,
			targetState:   listener.Ready,
			shouldSucceed: false,
		},
		{
			name:          "listening_to_ready",
			initialState:  listener.Listening,
			targetState:   listener.Ready,
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
			initialState:  listener.Closed,
			shouldSucceed: true,
			expectedState: listener.Listening,
		},
		{
			name:          "from_ready_valid",
			initialState:  listener.Ready,
			shouldSucceed: true,
			expectedState: listener.Listening,
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
			initialState:  listener.Listening,
			shouldSucceed: true,
			expectedState: listener.Ready,
		},
		{
			name:          "from_closed_invalid",
			initialState:  listener.Closed,
			shouldSucceed: false,
			expectedState: listener.Closed,
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
			initialState:  listener.Listening,
			shouldSucceed: true,
			expectedState: listener.Closed,
		},
		{
			name:          "from_ready",
			initialState:  listener.Ready,
			shouldSucceed: true,
			expectedState: listener.Closed,
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

// TestListener_HasProbe tests HasProbe method.
func TestListener_HasProbe(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*listener.Listener)
		expected bool
	}{
		{
			name: "without_probe",
			setup: func(_ *listener.Listener) {
				// No setup needed.
			},
			expected: false,
		},
		{
			name: "with_probe",
			setup: func(l *listener.Listener) {
				l.WithProbe("tcp", probe.NewConfig(), probe.NewTCPTarget("localhost:8080"))
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener and apply setup.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			tt.setup(l)

			// Verify HasProbe.
			assert.Equal(t, tt.expected, l.HasProbe())
		})
	}
}

// TestListener_ProbeAddress tests ProbeAddress method.
func TestListener_ProbeAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		port     int
		expected string
	}{
		{
			name:     "localhost",
			address:  "localhost",
			port:     8080,
			expected: "localhost:8080",
		},
		{
			name:     "empty_address",
			address:  "",
			port:     9090,
			expected: "127.0.0.1:9090",
		},
		{
			name:     "any_address",
			address:  "0.0.0.0",
			port:     80,
			expected: "127.0.0.1:80",
		},
		{
			name:     "ip_address",
			address:  "192.168.1.1",
			port:     443,
			expected: "192.168.1.1:443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener.
			l := listener.NewListener("test", "tcp", tt.address, tt.port)

			// Verify probe address.
			assert.Equal(t, tt.expected, l.ProbeAddress())
		})
	}
}
