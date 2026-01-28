// Package health_test provides black-box tests for the health package.
package health_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apphealth "github.com/kodflow/daemon/internal/application/health"
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/listener"
)

// mockProber is a mock implementation of health.Prober for testing.
type mockProber struct {
	probeType string
	result    health.CheckResult
}

// Probe returns the configured mock result.
//
// Params:
//   - ctx: the context for the probe operation.
//   - target: the target to healthcheck.
//
// Returns:
//   - health.CheckResult: the configured mock result.
func (m *mockProber) Probe(_ context.Context, _ health.Target) health.CheckResult {
	return m.result
}

// Type returns the prober type.
//
// Returns:
//   - string: the prober type identifier.
func (m *mockProber) Type() string {
	return m.probeType
}

// TestListenerProbe_HasProber tests the HasProber method.
func TestListenerProbe_HasProber(t *testing.T) {
	tests := []struct {
		name     string
		lp       apphealth.ListenerProbe
		expected bool
	}{
		{
			name: "without_prober",
			lp: apphealth.ListenerProbe{
				Listener: listener.NewListener("test", "tcp", "localhost", 8080),
				Prober:   nil,
			},
			expected: false,
		},
		{
			name: "with_prober",
			lp: apphealth.ListenerProbe{
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
			lp := apphealth.NewListenerProbe(tt.listener)

			// Verify listener probe was created.
			require.NotNil(t, lp)
			// Verify listener is set.
			assert.Equal(t, tt.listener, lp.Listener)
			// Verify prober is nil by default.
			assert.Nil(t, lp.Prober)
		})
	}
}

// TestNewListenerProbeWithBinding tests the NewListenerProbeWithBinding constructor.
func TestNewListenerProbeWithBinding(t *testing.T) {
	tests := []struct {
		name     string
		listener *listener.Listener
		binding  *apphealth.ProbeBinding
	}{
		{
			name:     "creates_with_nil_binding",
			listener: listener.NewListener("test", "tcp", "localhost", 8080),
			binding:  nil,
		},
		{
			name:     "creates_with_valid_binding",
			listener: listener.NewListener("test", "tcp", "localhost", 8080),
			binding: apphealth.NewProbeBinding("test", apphealth.ProbeTCP, apphealth.ProbeTarget{
				Address: "localhost:8080",
			}),
		},
		{
			name:     "creates_with_http_binding",
			listener: listener.NewListener("http-test", "tcp", "0.0.0.0", 9090),
			binding: apphealth.NewProbeBinding("http-test", apphealth.ProbeHTTP, apphealth.ProbeTarget{
				Address:    "localhost:9090",
				Path:       "/health",
				Method:     "GET",
				StatusCode: 200,
			}),
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			lp := apphealth.NewListenerProbeWithBinding(tt.listener, tt.binding)

			// Verify listener probe was created.
			require.NotNil(t, lp)
			// Verify listener is set.
			assert.Equal(t, tt.listener, lp.Listener)
			// Verify binding is set correctly.
			assert.Equal(t, tt.binding, lp.Binding)
			// Verify prober is nil by default (set separately).
			assert.Nil(t, lp.Prober)
		})
	}
}

// TestListenerProbe_HasBinding tests the HasBinding method.
func TestListenerProbe_HasBinding(t *testing.T) {
	tests := []struct {
		name     string
		lp       *apphealth.ListenerProbe
		expected bool
	}{
		{
			name: "without_binding",
			lp: apphealth.NewListenerProbe(
				listener.NewListener("test", "tcp", "localhost", 8080),
			),
			expected: false,
		},
		{
			name: "with_binding",
			lp: apphealth.NewListenerProbeWithBinding(
				listener.NewListener("test", "tcp", "localhost", 8080),
				apphealth.NewProbeBinding("test", apphealth.ProbeTCP, apphealth.ProbeTarget{
					Address: "localhost:8080",
				}),
			),
			expected: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Verify HasBinding returns expected value.
			assert.Equal(t, tt.expected, tt.lp.HasBinding())
		})
	}
}

// TestListenerProbe_ProbeAddress tests the ProbeAddress method.
func TestListenerProbe_ProbeAddress(t *testing.T) {
	tests := []struct {
		name     string
		lp       *apphealth.ListenerProbe
		expected string
	}{
		{
			name: "without_binding_uses_listener_address",
			lp: apphealth.NewListenerProbe(
				listener.NewListener("test", "tcp", "localhost", 8080),
			),
			expected: "localhost",
		},
		{
			name: "with_binding_empty_address_uses_listener",
			lp: apphealth.NewListenerProbeWithBinding(
				listener.NewListener("test", "tcp", "localhost", 8080),
				apphealth.NewProbeBinding("test", apphealth.ProbeTCP, apphealth.ProbeTarget{
					Address: "",
				}),
			),
			expected: "localhost",
		},
		{
			name: "with_binding_address_uses_binding",
			lp: apphealth.NewListenerProbeWithBinding(
				listener.NewListener("test", "tcp", "localhost", 8080),
				apphealth.NewProbeBinding("test", apphealth.ProbeTCP, apphealth.ProbeTarget{
					Address: "custom:9999",
				}),
			),
			expected: "custom:9999",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Verify ProbeAddress returns expected value.
			assert.Equal(t, tt.expected, tt.lp.ProbeAddress())
		})
	}
}

// TestListenerProbe_ProbeTarget tests the ProbeTarget method.
func TestListenerProbe_ProbeTarget(t *testing.T) {
	tests := []struct {
		name     string
		lp       *apphealth.ListenerProbe
		expected health.Target
	}{
		{
			name: "without_binding_returns_minimal_target",
			lp: apphealth.NewListenerProbe(
				listener.NewListener("test", "tcp", "localhost", 8080),
			),
			expected: health.Target{
				Address: "localhost",
			},
		},
		{
			name: "with_tcp_binding",
			lp: apphealth.NewListenerProbeWithBinding(
				listener.NewListener("test", "tcp", "localhost", 8080),
				apphealth.NewProbeBinding("test", apphealth.ProbeTCP, apphealth.ProbeTarget{
					Address: "custom:9999",
				}),
			),
			expected: health.Target{
				Address: "custom:9999",
			},
		},
		{
			name: "with_http_binding",
			lp: apphealth.NewListenerProbeWithBinding(
				listener.NewListener("http-test", "tcp", "0.0.0.0", 9090),
				apphealth.NewProbeBinding("http-test", apphealth.ProbeHTTP, apphealth.ProbeTarget{
					Address:    "localhost:9090",
					Path:       "/health",
					Method:     "GET",
					StatusCode: 200,
				}),
			),
			expected: health.Target{
				Address:    "localhost:9090",
				Path:       "/health",
				Method:     "GET",
				StatusCode: 200,
			},
		},
		{
			name: "with_grpc_binding",
			lp: apphealth.NewListenerProbeWithBinding(
				listener.NewListener("grpc-test", "tcp", "localhost", 50051),
				apphealth.NewProbeBinding("grpc-test", apphealth.ProbeGRPC, apphealth.ProbeTarget{
					Address: "localhost:50051",
					Service: "grpc.health.v1.Health",
				}),
			),
			expected: health.Target{
				Address: "localhost:50051",
				Service: "grpc.health.v1.Health",
			},
		},
		{
			name: "with_exec_binding",
			lp: apphealth.NewListenerProbeWithBinding(
				listener.NewListener("exec-test", "tcp", "localhost", 8080),
				apphealth.NewProbeBinding("exec-test", apphealth.ProbeExec, apphealth.ProbeTarget{
					Command: "/bin/check",
					Args:    []string{"--status"},
				}),
			),
			expected: health.Target{
				Address: "localhost",
				Command: "/bin/check",
				Args:    []string{"--status"},
			},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Verify ProbeTarget returns expected value.
			target := tt.lp.ProbeTarget()
			assert.Equal(t, tt.expected.Address, target.Address)
			assert.Equal(t, tt.expected.Path, target.Path)
			assert.Equal(t, tt.expected.Service, target.Service)
			assert.Equal(t, tt.expected.Method, target.Method)
			assert.Equal(t, tt.expected.StatusCode, target.StatusCode)
			assert.Equal(t, tt.expected.Command, target.Command)
			assert.Equal(t, tt.expected.Args, target.Args)
		})
	}
}

// TestListenerProbe_ProbeConfig tests the ProbeConfig method.
func TestListenerProbe_ProbeConfig(t *testing.T) {
	tests := []struct {
		name     string
		lp       *apphealth.ListenerProbe
		expected health.CheckConfig
	}{
		{
			name: "without_binding_returns_defaults",
			lp: apphealth.NewListenerProbe(
				listener.NewListener("test", "tcp", "localhost", 8080),
			),
			expected: health.CheckConfig{
				Interval:         health.DefaultInterval,
				Timeout:          health.DefaultTimeout,
				SuccessThreshold: health.DefaultSuccessThreshold,
				FailureThreshold: health.DefaultFailureThreshold,
			},
		},
		{
			name: "with_binding_returns_config",
			lp: apphealth.NewListenerProbeWithBinding(
				listener.NewListener("test", "tcp", "localhost", 8080),
				apphealth.NewProbeBinding("test", apphealth.ProbeTCP, apphealth.ProbeTarget{
					Address: "localhost:8080",
				}).WithConfig(apphealth.ProbeConfig{
					Interval:         apphealth.DefaultProbeConfig().Interval,
					Timeout:          apphealth.DefaultProbeConfig().Timeout,
					SuccessThreshold: apphealth.DefaultProbeConfig().SuccessThreshold,
					FailureThreshold: apphealth.DefaultProbeConfig().FailureThreshold,
				}),
			),
			expected: health.CheckConfig{
				Interval:         apphealth.DefaultProbeConfig().Interval,
				Timeout:          apphealth.DefaultProbeConfig().Timeout,
				SuccessThreshold: apphealth.DefaultProbeConfig().SuccessThreshold,
				FailureThreshold: apphealth.DefaultProbeConfig().FailureThreshold,
			},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Verify ProbeConfig returns expected value.
			config := tt.lp.ProbeConfig()
			assert.Equal(t, tt.expected.Interval, config.Interval)
			assert.Equal(t, tt.expected.Timeout, config.Timeout)
			assert.Equal(t, tt.expected.SuccessThreshold, config.SuccessThreshold)
			assert.Equal(t, tt.expected.FailureThreshold, config.FailureThreshold)
		})
	}
}
