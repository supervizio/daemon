// Package listener_test provides black-box tests for the listener package.
package listener_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/listener"
)

// TestState_String tests the String method.
func TestState_String(t *testing.T) {
	tests := []struct {
		name     string
		state    listener.State
		expected string
	}{
		{
			name:     "closed",
			state:    listener.StateClosed,
			expected: "closed",
		},
		{
			name:     "listening",
			state:    listener.StateListening,
			expected: "listening",
		},
		{
			name:     "ready",
			state:    listener.StateReady,
			expected: "ready",
		},
		{
			name:     "unknown",
			state:    listener.State(99),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify string representation.
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

// TestState_IsClosed tests the IsClosed method.
func TestState_IsClosed(t *testing.T) {
	tests := []struct {
		name     string
		state    listener.State
		expected bool
	}{
		{
			name:     "closed",
			state:    listener.StateClosed,
			expected: true,
		},
		{
			name:     "listening",
			state:    listener.StateListening,
			expected: false,
		},
		{
			name:     "ready",
			state:    listener.StateReady,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify IsClosed.
			assert.Equal(t, tt.expected, tt.state.IsClosed())
		})
	}
}

// TestState_IsListening tests the IsListening method.
func TestState_IsListening(t *testing.T) {
	tests := []struct {
		name     string
		state    listener.State
		expected bool
	}{
		{
			name:     "closed",
			state:    listener.StateClosed,
			expected: false,
		},
		{
			name:     "listening",
			state:    listener.StateListening,
			expected: true,
		},
		{
			name:     "ready",
			state:    listener.StateReady,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify IsListening.
			assert.Equal(t, tt.expected, tt.state.IsListening())
		})
	}
}

// TestState_IsReady tests the IsReady method.
func TestState_IsReady(t *testing.T) {
	tests := []struct {
		name     string
		state    listener.State
		expected bool
	}{
		{
			name:     "closed",
			state:    listener.StateClosed,
			expected: false,
		},
		{
			name:     "listening",
			state:    listener.StateListening,
			expected: false,
		},
		{
			name:     "ready",
			state:    listener.StateReady,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify IsReady.
			assert.Equal(t, tt.expected, tt.state.IsReady())
		})
	}
}

// TestState_CanTransitionTo tests the CanTransitionTo method.
func TestState_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     listener.State
		to       listener.State
		expected bool
	}{
		{
			name:     "closed_to_listening",
			from:     listener.StateClosed,
			to:       listener.StateListening,
			expected: true,
		},
		{
			name:     "closed_to_ready",
			from:     listener.StateClosed,
			to:       listener.StateReady,
			expected: false,
		},
		{
			name:     "listening_to_ready",
			from:     listener.StateListening,
			to:       listener.StateReady,
			expected: true,
		},
		{
			name:     "listening_to_closed",
			from:     listener.StateListening,
			to:       listener.StateClosed,
			expected: true,
		},
		{
			name:     "ready_to_listening",
			from:     listener.StateReady,
			to:       listener.StateListening,
			expected: true,
		},
		{
			name:     "ready_to_closed",
			from:     listener.StateReady,
			to:       listener.StateClosed,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify transition.
			assert.Equal(t, tt.expected, tt.from.CanTransitionTo(tt.to))
		})
	}
}
