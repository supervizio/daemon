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

// setupTestGRPCServer configures and starts a gRPC health server on the given listener.
// Returns a cleanup function that must be deferred by the caller.
//
// Goroutine lifecycle: The spawned goroutine runs server.Serve(listener) which
// blocks until the returned cleanup function calls server.GracefulStop().
func setupTestGRPCServer(listener net.Listener, serviceStatus grpc_health_v1.HealthCheckResponse_ServingStatus) func() {
	server := grpc.NewServer()

	healthServer := grpchealth.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	// Set the health status for the test service.
	healthServer.SetServingStatus("test.Service", serviceStatus)
	// Set overall server health.
	healthServer.SetServingStatus("", serviceStatus)

	// Start server in goroutine. Terminates when GracefulStop() is called.
	go func() {
		_ = server.Serve(listener)
	}()

	return func() {
		server.GracefulStop()
	}
}

// TestGRPCProber_Probe tests gRPC probing functionality with real gRPC health server.
func TestGRPCProber_Probe(t *testing.T) {
	tests := []struct {
		name           string
		serverStatus   grpc_health_v1.HealthCheckResponse_ServingStatus
		useServer      bool
		timeout        time.Duration
		targetAddress  string
		targetService  string
		expectSuccess  bool
		expectError    error
		outputContains string
	}{
		{
			name:           "successful_health_check",
			serverStatus:   grpc_health_v1.HealthCheckResponse_SERVING,
			useServer:      true,
			timeout:        5 * time.Second,
			targetService:  "",
			expectSuccess:  true,
			outputContains: "serving",
		},
		{
			name:           "successful_with_service_name",
			serverStatus:   grpc_health_v1.HealthCheckResponse_SERVING,
			useServer:      true,
			timeout:        5 * time.Second,
			targetService:  "test.Service",
			expectSuccess:  true,
			outputContains: "test.Service",
		},
		{
			name:           "not_serving",
			serverStatus:   grpc_health_v1.HealthCheckResponse_NOT_SERVING,
			useServer:      true,
			timeout:        5 * time.Second,
			targetService:  "test.Service",
			expectSuccess:  false,
			expectError:    healthcheck.ErrGRPCNotServing,
			outputContains: "not serving",
		},
		{
			name:          "connection_refused",
			useServer:     false,
			timeout:       100 * time.Millisecond,
			targetAddress: "127.0.0.1:1",
			expectSuccess: false,
		},
		{
			name:          "timeout_on_unreachable",
			useServer:     false,
			timeout:       50 * time.Millisecond,
			targetAddress: "192.0.2.1:50051",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addr string

			// Start test server if needed.
			if tt.useServer {
				listener, err := net.Listen("tcp", "127.0.0.1:0")
				require.NoError(t, err)
				defer func() { _ = listener.Close() }()

				addr = listener.Addr().String()
				cleanup := setupTestGRPCServer(listener, tt.serverStatus)
				defer cleanup()
			} else {
				addr = tt.targetAddress
			}

			// Create prober and target.
			prober := healthcheck.NewGRPCProber(tt.timeout)
			ctx := context.Background()

			target := health.Target{
				Address: addr,
				Service: tt.targetService,
			}

			// Execute probe.
			result := prober.Probe(ctx, target)

			// Verify success status.
			assert.Equal(t, tt.expectSuccess, result.Success)

			// Verify error if expected.
			if tt.expectError != nil {
				assert.ErrorIs(t, result.Error, tt.expectError)
			}

			// Verify output contains expected string.
			if tt.outputContains != "" {
				assert.Contains(t, result.Output, tt.outputContains)
			}

			// Verify latency is recorded.
			assert.Greater(t, result.Latency, time.Duration(0))
		})
	}
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
