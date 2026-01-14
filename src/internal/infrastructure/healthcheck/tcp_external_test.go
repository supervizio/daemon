// Package healthcheck_test provides black-box tests for the probe package.
package healthcheck_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainhealthcheck "github.com/kodflow/daemon/internal/domain/healthcheck"
	"github.com/kodflow/daemon/internal/infrastructure/healthcheck"
)

// TestNewTCPProber tests TCP prober creation.
func TestNewTCPProber(t *testing.T) {
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
			// Create TCP prober with specified timeout.
			prober := healthcheck.NewTCPProber(tt.timeout)

			// Verify prober is created.
			require.NotNil(t, prober)
		})
	}
}

// TestTCPProber_Type tests the Type method.
func TestTCPProber_Type(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns_tcp",
			expected: "tcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create TCP prober.
			prober := healthcheck.NewTCPProber(time.Second)

			// Verify type identifier.
			assert.Equal(t, tt.expected, prober.Type())
		})
	}
}

// TestTCPProber_Probe tests TCP probing functionality.
// This test spawns a background goroutine to accept connections.
// The goroutine terminates when the listener is closed via defer.
func TestTCPProber_Probe(t *testing.T) {
	// Start a test TCP server.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
		return
	}
	defer func() { _ = listener.Close() }()

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
		target        domainhealthcheck.Target
		timeout       time.Duration
		expectSuccess bool
	}{
		{
			name: "successful_connection",
			target: domainhealthcheck.Target{
				Address: listener.Addr().String(),
			},
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "successful_with_explicit_network",
			target: domainhealthcheck.Target{
				Address: listener.Addr().String(),
				Network: "tcp",
			},
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "connection_refused",
			target: domainhealthcheck.Target{
				Address: "127.0.0.1:1",
			},
			timeout:       100 * time.Millisecond,
			expectSuccess: false,
		},
		{
			name: "invalid_address",
			target: domainhealthcheck.Target{
				Address: "invalid:address:format",
			},
			timeout:       100 * time.Millisecond,
			expectSuccess: false,
		},
		{
			name: "timeout_on_unreachable",
			target: domainhealthcheck.Target{
				Address: "192.0.2.1:80",
			},
			timeout:       50 * time.Millisecond,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create TCP prober.
			prober := healthcheck.NewTCPProber(tt.timeout)
			ctx := context.Background()

			// Perform healthcheck.
			result := prober.Probe(ctx, tt.target)

			// Verify result based on expected outcome.
			if tt.expectSuccess {
				assert.True(t, result.Success)
				assert.NoError(t, result.Error)
				assert.Contains(t, result.Output, "connected")
			} else {
				assert.False(t, result.Success)
			}

			// Latency should always be measured.
			assert.Greater(t, result.Latency, time.Duration(0))
		})
	}
}

// TestTCPProber_Probe_ContextCancellation tests context cancellation.
func TestTCPProber_Probe_ContextCancellation(t *testing.T) {
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
			prober := healthcheck.NewTCPProber(10 * time.Second)

			// Create already cancelled context.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			// Create target for unreachable address.
			target := domainhealthcheck.Target{
				Address: "192.0.2.1:80",
			}

			// Probe should fail due to cancelled context.
			result := prober.Probe(ctx, target)
			assert.False(t, result.Success)
		})
	}
}
