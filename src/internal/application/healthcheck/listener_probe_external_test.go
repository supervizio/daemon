// Package health_test provides black-box tests for the health package.
package healthcheck_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apphc "github.com/kodflow/daemon/internal/application/healthcheck"
	"github.com/kodflow/daemon/internal/domain/healthcheck"
	"github.com/kodflow/daemon/internal/domain/listener"
)

// mockProber is a mock implementation of healthcheck.Prober for testing.
type mockProber struct {
	probeType string
	result    healthcheck.Result
}

// Probe returns the configured mock result.
//
// Params:
//   - ctx: the context for the probe operation.
//   - target: the target to healthcheck.
//
// Returns:
//   - healthcheck.Result: the configured mock result.
func (m *mockProber) Probe(_ context.Context, _ healthcheck.Target) healthcheck.Result {
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
		lp       apphc.ListenerProbe
		expected bool
	}{
		{
			name: "without_prober",
			lp: apphc.ListenerProbe{
				Listener: listener.NewListener("test", "tcp", "localhost", 8080),
				Prober:   nil,
			},
			expected: false,
		},
		{
			name: "with_prober",
			lp: apphc.ListenerProbe{
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
			lp := apphc.NewListenerProbe(tt.listener)

			// Verify listener probe was created.
			require.NotNil(t, lp)
			// Verify listener is set.
			assert.Equal(t, tt.listener, lp.Listener)
			// Verify prober is nil by default.
			assert.Nil(t, lp.Prober)
		})
	}
}
