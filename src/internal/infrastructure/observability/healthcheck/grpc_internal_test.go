// Package healthcheck provides internal tests for gRPC prober.
package healthcheck

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"github.com/kodflow/daemon/internal/domain/health"
)

// TestGRPCProber_internalFields tests internal struct fields.
func TestGRPCProber_internalFields(t *testing.T) {
	tests := []struct {
		name             string
		timeout          time.Duration
		expectedTimeout  time.Duration
		expectedInsecure bool
		useSecure        bool
	}{
		{
			name:             "insecure_prober",
			timeout:          5 * time.Second,
			expectedTimeout:  5 * time.Second,
			expectedInsecure: true,
			useSecure:        false,
		},
		{
			name:             "secure_prober",
			timeout:          5 * time.Second,
			expectedTimeout:  5 * time.Second,
			expectedInsecure: false,
			useSecure:        true,
		},
		{
			name:             "zero_timeout_insecure",
			timeout:          0,
			expectedTimeout:  0,
			expectedInsecure: true,
			useSecure:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var prober *GRPCProber
			if tt.useSecure {
				// Create secure gRPC prober.
				prober = NewGRPCProberSecure(tt.timeout)
			} else {
				// Create insecure gRPC prober.
				prober = NewGRPCProber(tt.timeout)
			}

			// Verify internal fields.
			assert.Equal(t, tt.expectedTimeout, prober.timeout)
			assert.Equal(t, tt.expectedInsecure, prober.insecureMode)
		})
	}
}

// TestProberTypeGRPC_constant tests the constant value.
func TestProberTypeGRPC_constant(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "constant_value",
			expected: "grpc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify constant matches expected value.
			assert.Equal(t, tt.expected, proberTypeGRPC)
		})
	}
}

// setupInternalTestGRPCServer configures and starts a gRPC health server on the given listener.
// Returns a cleanup function that must be deferred by the caller.
//
// Goroutine lifecycle: The spawned goroutine runs server.Serve(listener) which
// blocks until the returned cleanup function calls server.GracefulStop().
func setupInternalTestGRPCServer(listener net.Listener, serviceStatus grpc_health_v1.HealthCheckResponse_ServingStatus) func() {
	server := grpc.NewServer()
	healthServer := grpchealth.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	// Set service status.
	healthServer.SetServingStatus("test.Service", serviceStatus)
	healthServer.SetServingStatus("", serviceStatus)

	// Start server in goroutine. Terminates when GracefulStop() is called.
	go func() {
		_ = server.Serve(listener)
	}()

	return func() {
		server.GracefulStop()
	}
}

// TestGRPCProber_connect tests the connect method.
func TestGRPCProber_connect(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		insecureMode  bool
		useServer     bool
		targetAddress string
		expectError   bool
	}{
		{
			name:         "successful_connection",
			timeout:      5 * time.Second,
			insecureMode: true,
			useServer:    true,
			expectError:  false,
		},
		{
			name:          "connection_refused",
			timeout:       100 * time.Millisecond,
			insecureMode:  true,
			useServer:     false,
			targetAddress: "127.0.0.1:1",
			expectError:   true,
		},
		{
			name:          "timeout_on_unreachable",
			timeout:       50 * time.Millisecond,
			insecureMode:  true,
			useServer:     false,
			targetAddress: "192.0.2.1:50051",
			expectError:   true,
		},
		{
			name:         "secure_mode_connection",
			timeout:      100 * time.Millisecond,
			insecureMode: false,
			useServer:    true,
			expectError:  true,
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
				cleanup := setupInternalTestGRPCServer(listener, grpc_health_v1.HealthCheckResponse_SERVING)
				defer cleanup()
			} else {
				addr = tt.targetAddress
			}

			// Create prober with internal configuration.
			prober := &GRPCProber{
				timeout:      tt.timeout,
				insecureMode: tt.insecureMode,
			}

			// Test connect method.
			ctx := context.Background()
			conn, err := prober.connect(ctx, addr)

			// Verify result.
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, conn)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, conn)
				_ = conn.Close()
			}
		})
	}
}

// TestGRPCProber_checkHealth tests the checkHealth method.
func TestGRPCProber_checkHealth(t *testing.T) {
	tests := []struct {
		name         string
		serverStatus grpc_health_v1.HealthCheckResponse_ServingStatus
		service      string
		expectError  bool
		expectStatus grpc_health_v1.HealthCheckResponse_ServingStatus
	}{
		{
			name:         "serving_empty_service",
			serverStatus: grpc_health_v1.HealthCheckResponse_SERVING,
			service:      "",
			expectError:  false,
			expectStatus: grpc_health_v1.HealthCheckResponse_SERVING,
		},
		{
			name:         "serving_named_service",
			serverStatus: grpc_health_v1.HealthCheckResponse_SERVING,
			service:      "test.Service",
			expectError:  false,
			expectStatus: grpc_health_v1.HealthCheckResponse_SERVING,
		},
		{
			name:         "not_serving",
			serverStatus: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
			service:      "test.Service",
			expectError:  false,
			expectStatus: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		},
		{
			name:         "unknown_service",
			serverStatus: grpc_health_v1.HealthCheckResponse_SERVING,
			service:      "nonexistent.Service",
			expectError:  true,
			expectStatus: grpc_health_v1.HealthCheckResponse_UNKNOWN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create and start test server.
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			defer func() { _ = listener.Close() }()

			cleanup := setupInternalTestGRPCServer(listener, tt.serverStatus)
			defer cleanup()

			// Create prober and connect.
			prober := NewGRPCProber(5 * time.Second)
			ctx := context.Background()

			conn, err := prober.connect(ctx, listener.Addr().String())
			require.NoError(t, err)
			defer func() { _ = conn.Close() }()

			// Test checkHealth method.
			resp, err := prober.checkHealth(ctx, conn, tt.service)

			// Verify result.
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectStatus, resp.Status)
			}
		})
	}
}

// TestGRPCProber_handleRPCError tests the handleRPCError method.
func TestGRPCProber_handleRPCError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		latency        time.Duration
		service        string
		expectSuccess  bool
		expectError    error
		outputContains string
	}{
		{
			name:           "not_found_error",
			err:            status.Error(codes.NotFound, "service not found"),
			latency:        10 * time.Millisecond,
			service:        "test.Service",
			expectSuccess:  false,
			expectError:    ErrGRPCServiceUnknown,
			outputContains: "unknown",
		},
		{
			name:           "deadline_exceeded",
			err:            status.Error(codes.DeadlineExceeded, "deadline exceeded"),
			latency:        5 * time.Second,
			service:        "test.Service",
			expectSuccess:  false,
			expectError:    nil,
			outputContains: "timeout",
		},
		{
			name:           "unavailable_error",
			err:            status.Error(codes.Unavailable, "service unavailable"),
			latency:        100 * time.Millisecond,
			service:        "test.Service",
			expectSuccess:  false,
			expectError:    nil,
			outputContains: "service unavailable",
		},
		{
			name:           "internal_error",
			err:            status.Error(codes.Internal, "internal error"),
			latency:        50 * time.Millisecond,
			service:        "test.Service",
			expectSuccess:  false,
			expectError:    nil,
			outputContains: "internal error",
		},
		{
			name:           "non_grpc_error",
			err:            errors.New("network error"),
			latency:        25 * time.Millisecond,
			service:        "test.Service",
			expectSuccess:  false,
			expectError:    nil,
			outputContains: "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create prober.
			prober := NewGRPCProber(5 * time.Second)

			// Test handleRPCError method.
			result := prober.handleRPCError(tt.err, tt.latency, tt.service)

			// Verify result.
			assert.Equal(t, tt.expectSuccess, result.Success)
			assert.Equal(t, tt.latency, result.Latency)
			assert.Contains(t, result.Output, tt.outputContains)

			// Verify specific error if expected.
			if tt.expectError != nil {
				assert.ErrorIs(t, result.Error, tt.expectError)
			}
		})
	}
}

// TestGRPCProber_handleHealthStatus tests the handleHealthStatus method.
func TestGRPCProber_handleHealthStatus(t *testing.T) {
	tests := []struct {
		name           string
		status         grpc_health_v1.HealthCheckResponse_ServingStatus
		latency        time.Duration
		target         health.Target
		expectSuccess  bool
		expectError    error
		outputContains string
	}{
		{
			name:    "serving_with_service",
			status:  grpc_health_v1.HealthCheckResponse_SERVING,
			latency: 10 * time.Millisecond,
			target: health.Target{
				Address: "localhost:50051",
				Service: "test.Service",
			},
			expectSuccess:  true,
			expectError:    nil,
			outputContains: "test.Service",
		},
		{
			name:    "serving_empty_service",
			status:  grpc_health_v1.HealthCheckResponse_SERVING,
			latency: 10 * time.Millisecond,
			target: health.Target{
				Address: "localhost:50051",
				Service: "",
			},
			expectSuccess:  true,
			expectError:    nil,
			outputContains: "(server)",
		},
		{
			name:    "not_serving",
			status:  grpc_health_v1.HealthCheckResponse_NOT_SERVING,
			latency: 15 * time.Millisecond,
			target: health.Target{
				Address: "localhost:50051",
				Service: "test.Service",
			},
			expectSuccess:  false,
			expectError:    ErrGRPCNotServing,
			outputContains: "not serving",
		},
		{
			name:    "service_unknown",
			status:  grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN,
			latency: 20 * time.Millisecond,
			target: health.Target{
				Address: "localhost:50051",
				Service: "unknown.Service",
			},
			expectSuccess:  false,
			expectError:    ErrGRPCServiceUnknown,
			outputContains: "unknown",
		},
		{
			name:    "unknown_status_value",
			status:  grpc_health_v1.HealthCheckResponse_ServingStatus(999),
			latency: 5 * time.Millisecond,
			target: health.Target{
				Address: "localhost:50051",
				Service: "test.Service",
			},
			expectSuccess:  false,
			expectError:    ErrGRPCUnknownStatus,
			outputContains: "status unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create prober.
			prober := NewGRPCProber(5 * time.Second)

			// Build response.
			resp := &grpc_health_v1.HealthCheckResponse{
				Status: tt.status,
			}

			// Test handleHealthStatus method.
			result := prober.handleHealthStatus(resp, tt.latency, tt.target)

			// Verify result.
			assert.Equal(t, tt.expectSuccess, result.Success)
			assert.Equal(t, tt.latency, result.Latency)
			assert.Contains(t, result.Output, tt.outputContains)

			// Verify specific error if expected.
			if tt.expectError != nil {
				assert.ErrorIs(t, result.Error, tt.expectError)
			}
		})
	}
}
