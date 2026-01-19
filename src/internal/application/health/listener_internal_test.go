// Package health provides internal tests for listener.go.
// It tests internal implementation details using white-box testing.
package health

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/listener"
)

// testProber is a mock prober for internal testing.
//
// testProber provides a controllable prober implementation for
// testing internal ListenerProbe behavior.
type testProber struct {
	probeType string
	result    domain.CheckResult
}

// Probe returns the configured test result.
//
// Params:
//   - ctx: the context for cancellation.
//   - target: the probe target.
//
// Returns:
//   - domain.CheckResult: the configured test result.
func (p *testProber) Probe(_ context.Context, _ domain.Target) domain.CheckResult {
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
		prober domain.Prober
		// binding is the probe binding.
		binding *ProbeBinding
	}{
		{
			name:     "empty_listener_probe",
			listener: nil,
			prober:   nil,
			binding:  nil,
		},
		{
			name:     "listener_probe_with_listener",
			listener: listener.NewListener("test", "tcp", "localhost", 8080),
			prober:   nil,
			binding:  nil,
		},
		{
			name:     "listener_probe_with_prober",
			listener: listener.NewListener("test", "tcp", "localhost", 8080),
			prober:   &testProber{probeType: "tcp"},
			binding:  nil,
		},
		{
			name:     "listener_probe_with_binding",
			listener: listener.NewListener("test", "tcp", "localhost", 8080),
			prober:   &testProber{probeType: "tcp"},
			binding: &ProbeBinding{
				ListenerName: "test",
				Type:         ProbeTCP,
				Target: ProbeTarget{
					Address: "localhost:8080",
				},
				Config: ProbeConfig{
					Interval:         10 * time.Second,
					Timeout:          5 * time.Second,
					SuccessThreshold: 1,
					FailureThreshold: 3,
				},
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
				Binding:  tt.binding,
			}

			// Verify listener field.
			assert.Equal(t, tt.listener, lp.Listener)
			// Verify prober field.
			assert.Equal(t, tt.prober, lp.Prober)
			// Verify binding field.
			assert.Equal(t, tt.binding, lp.Binding)
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
		prober domain.Prober
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
				result:    domain.CheckResult{Success: true},
			},
			expected: true,
		},
		{
			name: "returns_true_for_prober_with_failure_result",
			prober: &testProber{
				probeType: "tcp",
				result:    domain.CheckResult{Success: false},
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

// Test_ListenerProbe_HasBinding_internal tests HasBinding method internals.
//
// Params:
//   - t: the testing context.
func Test_ListenerProbe_HasBinding_internal(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// binding is the binding to assign.
		binding *ProbeBinding
		// expected is the expected result.
		expected bool
	}{
		{
			name:     "returns_false_for_nil_binding",
			binding:  nil,
			expected: false,
		},
		{
			name: "returns_true_for_non_nil_binding",
			binding: &ProbeBinding{
				ListenerName: "test",
				Type:         ProbeTCP,
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
				Binding:  tt.binding,
			}

			result := lp.HasBinding()

			// Verify the expected result.
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_ListenerProbe_ProbeAddress_internal tests ProbeAddress method.
//
// Params:
//   - t: the testing context.
func Test_ListenerProbe_ProbeAddress_internal(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// listenerAddr is the listener address.
		listenerAddr string
		// binding is the probe binding.
		binding *ProbeBinding
		// expected is the expected address.
		expected string
	}{
		{
			name:         "returns_listener_address_without_binding",
			listenerAddr: "localhost",
			binding:      nil,
			expected:     "localhost",
		},
		{
			name:         "returns_binding_address_when_set",
			listenerAddr: "localhost",
			binding: &ProbeBinding{
				Target: ProbeTarget{
					Address: "127.0.0.1:8080",
				},
			},
			expected: "127.0.0.1:8080",
		},
		{
			name:         "returns_listener_address_when_binding_address_empty",
			listenerAddr: "localhost",
			binding: &ProbeBinding{
				Target: ProbeTarget{
					Address: "",
				},
			},
			expected: "localhost",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			l := listener.NewListener("test", "tcp", tt.listenerAddr, 8080)
			lp := &ListenerProbe{
				Listener: l,
				Binding:  tt.binding,
			}

			result := lp.ProbeAddress()

			// Verify the expected address.
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_ListenerProbe_ProbeConfig_internal tests ProbeConfig method.
//
// Params:
//   - t: the testing context.
func Test_ListenerProbe_ProbeConfig_internal(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// binding is the probe binding.
		binding *ProbeBinding
		// expectedTimeout is the expected timeout.
		expectedTimeout time.Duration
		// expectedInterval is the expected interval.
		expectedInterval time.Duration
	}{
		{
			name:             "returns_defaults_without_binding",
			binding:          nil,
			expectedTimeout:  domain.DefaultTimeout,
			expectedInterval: domain.DefaultInterval,
		},
		{
			name: "returns_binding_config_when_set",
			binding: &ProbeBinding{
				Config: ProbeConfig{
					Timeout:  10 * time.Second,
					Interval: 30 * time.Second,
				},
			},
			expectedTimeout:  10 * time.Second,
			expectedInterval: 30 * time.Second,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			lp := &ListenerProbe{
				Listener: listener.NewListener("test", "tcp", "localhost", 8080),
				Binding:  tt.binding,
			}

			config := lp.ProbeConfig()

			// Verify the expected config.
			assert.Equal(t, tt.expectedTimeout, config.Timeout)
			assert.Equal(t, tt.expectedInterval, config.Interval)
		})
	}
}
