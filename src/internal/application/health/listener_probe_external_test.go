// Package health_test provides black-box tests for the health package.
package health_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/application/health"
	"github.com/kodflow/daemon/internal/domain/listener"
	"github.com/kodflow/daemon/internal/domain/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProber is a mock implementation of probe.Prober for testing.
type mockProber struct {
	probeType string
	result    probe.Result
}

// Probe returns the configured mock result.
//
// Params:
//   - ctx: the context for the probe operation.
//   - target: the target to probe.
//
// Returns:
//   - probe.Result: the configured mock result.
func (m *mockProber) Probe(_ context.Context, _ probe.Target) probe.Result {
	// Return the pre-configured result for testing.
	return m.result
}

// Type returns the prober type.
//
// Returns:
//   - string: the prober type identifier.
func (m *mockProber) Type() string {
	// Return the configured prober type.
	return m.probeType
}

// TestListenerProbe_HasProber tests the HasProber method.
func TestListenerProbe_HasProber(t *testing.T) {
	tests := []struct {
		name     string
		lp       health.ListenerProbe
		expected bool
	}{
		{
			name: "without_prober",
			lp: health.ListenerProbe{
				Listener: listener.NewListener("test", "tcp", "localhost", 8080),
				Prober:   nil,
			},
			expected: false,
		},
		{
			name: "with_prober",
			lp: health.ListenerProbe{
				Listener: listener.NewListener("test", "tcp", "localhost", 8080),
				Prober:   &mockProber{probeType: "tcp"},
			},
			expected: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Verify HasProber.
			assert.Equal(t, tt.expected, tt.lp.HasProber())
		})
	}
}

// TestNewListenerProbe tests the NewListenerProbe constructor.
func TestNewListenerProbe(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// listener is the listener to use.
		listener *listener.Listener
	}{
		{
			name:     "creates_with_nil_listener",
			listener: nil,
		},
		{
			name:     "creates_with_valid_listener",
			listener: listener.NewListener("test", "tcp", "localhost", 8080),
		},
		{
			name:     "creates_with_udp_listener",
			listener: listener.NewListener("udp-test", "udp", "0.0.0.0", 5353),
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			lp := health.NewListenerProbe(tt.listener)

			// Verify listener probe was created.
			require.NotNil(t, lp)
			// Verify listener is set.
			assert.Equal(t, tt.listener, lp.Listener)
			// Verify prober is nil by default.
			assert.Nil(t, lp.Prober)
		})
	}
}
