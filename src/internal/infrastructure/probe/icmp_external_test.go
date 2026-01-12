// Package probe_test provides black-box tests for the probe package.
package probe_test

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainprobe "github.com/kodflow/daemon/internal/domain/probe"
	"github.com/kodflow/daemon/internal/infrastructure/probe"
)

// TestNewICMPProber tests ICMP prober creation.
func TestNewICMPProber(t *testing.T) {
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
			// Create ICMP prober with specified timeout.
			prober := probe.NewICMPProber(tt.timeout)

			// Verify prober is created.
			require.NotNil(t, prober)
		})
	}
}

// TestNewICMPProberWithTCPFallback tests ICMP prober creation with custom TCP port.
func TestNewICMPProberWithTCPFallback(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		tcpPort int
	}{
		{
			name:    "port_80",
			timeout: time.Second,
			tcpPort: 80,
		},
		{
			name:    "port_443",
			timeout: time.Second,
			tcpPort: 443,
		},
		{
			name:    "custom_port",
			timeout: time.Second,
			tcpPort: 8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober with TCP fallback.
			prober := probe.NewICMPProberWithTCPFallback(tt.timeout, tt.tcpPort)

			// Verify prober is created.
			require.NotNil(t, prober)
		})
	}
}

// TestICMPProber_Type tests the Type method.
func TestICMPProber_Type(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns_icmp",
			expected: "icmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober.
			prober := probe.NewICMPProber(time.Second)

			// Verify type identifier.
			assert.Equal(t, tt.expected, prober.Type())
		})
	}
}

// TestICMPProber_Probe tests ICMP probing functionality (TCP fallback).
// This test spawns a background goroutine to accept connections.
// The goroutine terminates when the listener is closed via defer.
func TestICMPProber_Probe(t *testing.T) {
	// Start a test TCP server to simulate reachable host.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
		return
	}
	defer listener.Close() //nolint:errcheck // cleanup in test

	// Get the port number.
	_, portStr, err := net.SplitHostPort(listener.Addr().String())
	// Check if port extraction failed.
	if err != nil {
		t.Fatalf("failed to extract port: %v", err)
	}

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
		tcpPort       int
		timeout       time.Duration
		expectSuccess bool
	}{
		{
			name: "successful_tcp_fallback",
			target: domainprobe.Target{
				Address: "127.0.0.1",
			},
			tcpPort:       mustParsePort(portStr),
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "successful_with_port_in_address",
			target: domainprobe.Target{
				Address: listener.Addr().String(),
			},
			tcpPort:       mustParsePort(portStr),
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "failure_unreachable_port",
			target: domainprobe.Target{
				Address: "127.0.0.1",
			},
			tcpPort:       1, // Port 1 should be unreachable.
			timeout:       100 * time.Millisecond,
			expectSuccess: false,
		},
		{
			name: "failure_unreachable_host",
			target: domainprobe.Target{
				Address: "192.0.2.1",
			},
			tcpPort:       80,
			timeout:       50 * time.Millisecond,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober with specific TCP port.
			prober := probe.NewICMPProberWithTCPFallback(tt.timeout, tt.tcpPort)
			ctx := context.Background()

			// Perform probe.
			result := prober.Probe(ctx, tt.target)

			// Verify result based on expected outcome.
			if tt.expectSuccess {
				assert.True(t, result.Success)
				assert.NoError(t, result.Error)
				assert.Contains(t, result.Output, "ping")
			} else {
				assert.False(t, result.Success)
			}

			// Latency should always be measured.
			assert.Greater(t, result.Latency, time.Duration(0))
		})
	}
}

// TestICMPProber_Probe_ContextCancellation tests context cancellation.
func TestICMPProber_Probe_ContextCancellation(t *testing.T) {
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
			prober := probe.NewICMPProber(10 * time.Second)

			// Create already cancelled context.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			target := domainprobe.Target{
				Address: "192.0.2.1",
			}

			// Probe should fail due to cancelled context.
			result := prober.Probe(ctx, target)
			assert.False(t, result.Success)
		})
	}
}

// mustParsePort parses a port string and panics on error.
func mustParsePort(portStr string) int {
	// Try to resolve named port (e.g., "http" -> 80).
	if port, err := net.LookupPort("tcp", portStr); err == nil {
		// Return resolved port number.
		return port
	}

	// Fall back to numeric parsing.
	port, err := strconv.Atoi(portStr)
	// Panic if parsing fails to catch test configuration errors.
	if err != nil {
		panic(err)
	}
	// Return parsed port number.
	return port
}
