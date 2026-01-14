// Package health provides internal tests for listener_healthcheck.go.
// It tests internal implementation details using white-box testing.
package healthcheck

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/listener"
	"github.com/kodflow/daemon/internal/domain/healthcheck"
)

// testProber is a mock prober for internal testing.
//
// testProber provides a controllable prober implementation for
// testing internal ListenerProbe behavior.
type testProber struct {
	probeType string
	result    healthcheck.Result
}

// Probe returns the configured test result.
//
// Params:
//   - ctx: the context for cancellation.
//   - target: the probe target.
//
// Returns:
//   - healthcheck.Result: the configured test result.
func (p *testProber) Probe(_ context.Context, _ healthcheck.Target) healthcheck.Result {
	// Return the pre-configured result for testing.
	return p.result
}

// Type returns the prober type.
//
// Returns:
//   - string: the prober type identifier.
func (p *testProber) Type() string {
	// Return the configured prober type.
	return p.probeType
}

// Test_ListenerProbe_struct tests the ListenerProbe struct fields.
//
// Params:
//   - t: the testing context.
func Test_ListenerProbe_struct(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// listener is the listener to use.
		listener *listener.Listener
		// prober is the prober to use.
		prober healthcheck.Prober
		// config is the probe config.
		config healthcheck.Config
		// target is the probe target.
		target healthcheck.Target
	}{
		{
			name:     "empty_listener_probe",
			listener: nil,
			prober:   nil,
			config:   healthcheck.Config{},
			target:   healthcheck.Target{},
		},
		{
			name:     "listener_probe_with_listener",
			listener: listener.NewListener("test", "tcp", "localhost", 8080),
			prober:   nil,
			config:   healthcheck.Config{},
			target:   healthcheck.Target{},
		},
		{
			name:     "listener_probe_with_prober",
			listener: listener.NewListener("test", "tcp", "localhost", 8080),
			prober:   &testProber{probeType: "tcp"},
			config:   healthcheck.Config{},
			target:   healthcheck.Target{},
		},
		{
			name:     "listener_probe_with_all_fields",
			listener: listener.NewListener("test", "tcp", "localhost", 8080),
			prober:   &testProber{probeType: "tcp"},
			config: healthcheck.Config{
				SuccessThreshold: 1,
				FailureThreshold: 3,
			},
			target: healthcheck.Target{
				Address: "localhost:8080",
			},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			lp := &ListenerProbe{
				Listener: tt.listener,
				Prober:   tt.prober,
				Config:   tt.config,
				Target:   tt.target,
			}

			// Verify listener field.
			assert.Equal(t, tt.listener, lp.Listener)
			// Verify prober field.
			assert.Equal(t, tt.prober, lp.Prober)
			// Verify config field.
			assert.Equal(t, tt.config, lp.Config)
			// Verify target field.
			assert.Equal(t, tt.target, lp.Target)
		})
	}
}

// Test_ListenerProbe_HasProber_internal tests HasProber method internals.
//
// Params:
//   - t: the testing context.
func Test_ListenerProbe_HasProber_internal(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prober is the prober to assign.
		prober healthcheck.Prober
		// expected is the expected result.
		expected bool
	}{
		{
			name:     "returns_false_for_nil_prober",
			prober:   nil,
			expected: false,
		},
		{
			name:     "returns_true_for_non_nil_prober",
			prober:   &testProber{probeType: "tcp"},
			expected: true,
		},
		{
			name: "returns_true_for_prober_with_success_result",
			prober: &testProber{
				probeType: "http",
				result:    healthcheck.Result{Success: true},
			},
			expected: true,
		},
		{
			name: "returns_true_for_prober_with_failure_result",
			prober: &testProber{
				probeType: "tcp",
				result:    healthcheck.Result{Success: false},
			},
			expected: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			lp := &ListenerProbe{
				Listener: listener.NewListener("test", "tcp", "localhost", 8080),
				Prober:   tt.prober,
			}

			result := lp.HasProber()

			// Verify the expected result.
			assert.Equal(t, tt.expected, result)
		})
	}
}
