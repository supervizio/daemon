// Package healthcheck_test provides black-box tests for the probe package.
package healthcheck_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/infrastructure/observability/healthcheck"
)

// TestNewGRPCProber tests gRPC prober creation.
func TestNewGRPCProber(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "standard_timeout",
			timeout: 5 * time.Second,
		},
		{
			name:    "short_timeout",
			timeout: 100 * time.Millisecond,
		},
		{
			name:    "zero_timeout",
			timeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create gRPC prober with specified timeout.
			prober := healthcheck.NewGRPCProber(tt.timeout)

			// Verify prober is created.
			require.NotNil(t, prober)
		})
	}
}

// TestNewGRPCProberSecure tests secure gRPC prober creation.
func TestNewGRPCProberSecure(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "standard_timeout",
			timeout: 5 * time.Second,
		},
		{
			name:    "short_timeout",
			timeout: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create secure gRPC prober.
			prober := healthcheck.NewGRPCProberSecure(tt.timeout)

			// Verify prober is created.
			require.NotNil(t, prober)
		})
	}
}

// TestGRPCProber_Type tests the Type method.
func TestGRPCProber_Type(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns_grpc",
			expected: "grpc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create gRPC prober.
			prober := healthcheck.NewGRPCProber(time.Second)

			// Verify type identifier.
			assert.Equal(t, tt.expected, prober.Type())
		})
	}
}

// startTestGRPCServer starts a gRPC server with health checking for testing.
// Returns the server address and a cleanup function.
func startTestGRPCServer(t *testing.T, serviceStatus grpc_health_v1.HealthCheckResponse_ServingStatus) (addr string, cleanup func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	healthServer := grpchealth.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	// Set the health status for the test service.
	healthServer.SetServingStatus("test.Service", serviceStatus)
	// Set overall server health.
	healthServer.SetServingStatus("", serviceStatus)

	go func() {
		_ = server.Serve(listener)
	}()

	cleanup = func() {
		server.GracefulStop()
	}

	addr = listener.Addr().String()
	return addr, cleanup
}

// TestGRPCProber_Probe tests gRPC probing functionality with real gRPC health server.
func TestGRPCProber_Probe(t *testing.T) {
	t.Run("successful_health_check", func(t *testing.T) {
		addr, cleanup := startTestGRPCServer(t, grpc_health_v1.HealthCheckResponse_SERVING)
		defer cleanup()

		prober := healthcheck.NewGRPCProber(5 * time.Second)
		ctx := context.Background()

		target := health.Target{
			Address: addr,
			Service: "",
		}

		result := prober.Probe(ctx, target)
		assert.True(t, result.Success)
		assert.NoError(t, result.Error)
		assert.Contains(t, result.Output, "serving")
		assert.Greater(t, result.Latency, time.Duration(0))
	})

	t.Run("successful_with_service_name", func(t *testing.T) {
		addr, cleanup := startTestGRPCServer(t, grpc_health_v1.HealthCheckResponse_SERVING)
		defer cleanup()

		prober := healthcheck.NewGRPCProber(5 * time.Second)
		ctx := context.Background()

		target := health.Target{
			Address: addr,
			Service: "test.Service",
		}

		result := prober.Probe(ctx, target)
		assert.True(t, result.Success)
		assert.NoError(t, result.Error)
		assert.Contains(t, result.Output, "test.Service")
		assert.Greater(t, result.Latency, time.Duration(0))
	})

	t.Run("not_serving", func(t *testing.T) {
		addr, cleanup := startTestGRPCServer(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		defer cleanup()

		prober := healthcheck.NewGRPCProber(5 * time.Second)
		ctx := context.Background()

		target := health.Target{
			Address: addr,
			Service: "test.Service",
		}

		result := prober.Probe(ctx, target)
		assert.False(t, result.Success)
		assert.ErrorIs(t, result.Error, healthcheck.ErrGRPCNotServing)
		assert.Contains(t, result.Output, "not serving")
	})

	t.Run("connection_refused", func(t *testing.T) {
		prober := healthcheck.NewGRPCProber(100 * time.Millisecond)
		ctx := context.Background()

		target := health.Target{
			Address: "127.0.0.1:1",
		}

		result := prober.Probe(ctx, target)
		assert.False(t, result.Success)
		assert.Greater(t, result.Latency, time.Duration(0))
	})

	t.Run("timeout_on_unreachable", func(t *testing.T) {
		prober := healthcheck.NewGRPCProber(50 * time.Millisecond)
		ctx := context.Background()

		target := health.Target{
			Address: "192.0.2.1:50051",
		}

		result := prober.Probe(ctx, target)
		assert.False(t, result.Success)
	})
}

// TestGRPCProber_Probe_ContextCancellation tests context cancellation.
func TestGRPCProber_Probe_ContextCancellation(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "cancelled_context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create prober with long timeout.
			prober := healthcheck.NewGRPCProber(10 * time.Second)

			// Create already cancelled context.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			target := health.Target{
				Address: "192.0.2.1:50051",
			}

			// Probe should fail due to cancelled context.
			result := prober.Probe(ctx, target)
			assert.False(t, result.Success)
		})
	}
}

// TestErrGRPCNotServing tests the exported error variable.
func TestErrGRPCNotServing(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "error_is_accessible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify error is not nil.
			assert.NotNil(t, healthcheck.ErrGRPCNotServing)
			assert.Contains(t, healthcheck.ErrGRPCNotServing.Error(), "serving")
		})
	}
}

// TestErrGRPCServiceUnknown tests the exported error variable.
func TestErrGRPCServiceUnknown(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "error_is_accessible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify error is not nil.
			assert.NotNil(t, healthcheck.ErrGRPCServiceUnknown)
			assert.Contains(t, healthcheck.ErrGRPCServiceUnknown.Error(), "unknown")
		})
	}
}
