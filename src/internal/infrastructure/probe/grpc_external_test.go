// Package probe_test provides black-box tests for the probe package.
package probe_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainprobe "github.com/kodflow/daemon/internal/domain/probe"
	"github.com/kodflow/daemon/internal/infrastructure/probe"
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
			prober := probe.NewGRPCProber(tt.timeout)

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
			prober := probe.NewGRPCProberSecure(tt.timeout)

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
			prober := probe.NewGRPCProber(time.Second)

			// Verify type identifier.
			assert.Equal(t, tt.expected, prober.Type())
		})
	}
}

// TestGRPCProber_Probe tests gRPC probing functionality.
// This test spawns a background goroutine to accept connections.
// The goroutine terminates when the listener is closed via defer.
func TestGRPCProber_Probe(t *testing.T) {
	// Start a test TCP server to simulate gRPC endpoint.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
		return
	}
	defer listener.Close() //nolint:errcheck // cleanup in test

	// Accept connections in background goroutine.
	// Goroutine terminates when listener.Accept returns error on Close.
	go func() {
		for {
			conn, acceptErr := listener.Accept()
			// Check if listener was closed.
			if acceptErr != nil {
				// Listener closed, terminate goroutine.
				return
			}
			// Close accepted connection.
			_ = conn.Close()
		}
	}()

	tests := []struct {
		name          string
		target        domainprobe.Target
		timeout       time.Duration
		expectSuccess bool
	}{
		{
			name: "successful_connection",
			target: domainprobe.Target{
				Address: listener.Addr().String(),
			},
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "successful_with_service",
			target: domainprobe.Target{
				Address: listener.Addr().String(),
				Service: "health.v1.Health",
			},
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "connection_refused",
			target: domainprobe.Target{
				Address: "127.0.0.1:1",
			},
			timeout:       100 * time.Millisecond,
			expectSuccess: false,
		},
		{
			name: "timeout_on_unreachable",
			target: domainprobe.Target{
				Address: "192.0.2.1:50051",
			},
			timeout:       50 * time.Millisecond,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create gRPC prober.
			prober := probe.NewGRPCProber(tt.timeout)
			ctx := context.Background()

			// Perform probe.
			result := prober.Probe(ctx, tt.target)

			// Verify result based on expected outcome.
			if tt.expectSuccess {
				assert.True(t, result.Success)
				assert.NoError(t, result.Error)
				assert.Contains(t, result.Output, "gRPC")
			} else {
				assert.False(t, result.Success)
			}

			// Latency should always be measured.
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
			prober := probe.NewGRPCProber(10 * time.Second)

			// Create already cancelled context.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			target := domainprobe.Target{
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
			assert.NotNil(t, probe.ErrGRPCNotServing)
			assert.Contains(t, probe.ErrGRPCNotServing.Error(), "serving")
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
			assert.NotNil(t, probe.ErrGRPCServiceUnknown)
			assert.Contains(t, probe.ErrGRPCServiceUnknown.Error(), "unknown")
		})
	}
}
