// Package healthcheck provides internal tests for ICMP prober.
package healthcheck

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestICMPProber_internalFields tests internal struct fields.
func TestICMPProber_internalFields(t *testing.T) {
	tests := []struct {
		name                   string
		timeout                time.Duration
		tcpPort                int
		expectedTimeout        time.Duration
		expectedUseTCPFallback bool
		expectedTCPPort        int
		useWithTCPFallback     bool
	}{
		{
			name:                   "default_prober",
			timeout:                5 * time.Second,
			expectedTimeout:        5 * time.Second,
			expectedUseTCPFallback: true,
			expectedTCPPort:        defaultTCPFallbackPort,
			useWithTCPFallback:     false,
		},
		{
			name:                   "prober_with_custom_port",
			timeout:                5 * time.Second,
			tcpPort:                443,
			expectedTimeout:        5 * time.Second,
			expectedUseTCPFallback: true,
			expectedTCPPort:        443,
			useWithTCPFallback:     true,
		},
		{
			name:                   "zero_timeout",
			timeout:                0,
			expectedTimeout:        0,
			expectedUseTCPFallback: true,
			expectedTCPPort:        defaultTCPFallbackPort,
			useWithTCPFallback:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var prober *ICMPProber
			if tt.useWithTCPFallback {
				// Create ICMP prober with TCP fallback.
				prober = NewICMPProberWithTCPFallback(tt.timeout, tt.tcpPort)
			} else {
				// Create default ICMP prober.
				prober = NewICMPProber(tt.timeout)
			}

			// Verify internal fields.
			assert.Equal(t, tt.expectedTimeout, prober.timeout)
			assert.Equal(t, tt.expectedUseTCPFallback, prober.useTCPFallback)
			assert.Equal(t, tt.expectedTCPPort, prober.tcpPort)
		})
	}
}

// TestICMPProber_tcpPing tests the internal tcpPing method.
// This test spawns a background goroutine to accept connections.
// The goroutine terminates when the listener is closed via defer.
func TestICMPProber_tcpPing(t *testing.T) {
	// Start a test TCP server.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
		return
	}
	defer func() { _ = listener.Close() }()

	// Get the port number.
	_, portStr, err := net.SplitHostPort(listener.Addr().String())
	// Check if port extraction failed.
	if err != nil {
		t.Fatalf("failed to extract port: %v", err)
	}
	port := parsePort(portStr)

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
		host          string
		tcpPort       int
		timeout       time.Duration
		expectSuccess bool
	}{
		{
			name:          "successful_ping",
			host:          "127.0.0.1",
			tcpPort:       port,
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name:          "failure_unreachable_port",
			host:          "127.0.0.1",
			tcpPort:       1,
			timeout:       100 * time.Millisecond,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober.
			prober := NewICMPProberWithTCPFallback(tt.timeout, tt.tcpPort)
			ctx := context.Background()
			start := time.Now()

			// Call internal method.
			result := prober.tcpPing(ctx, tt.host, start)

			// Verify result.
			if tt.expectSuccess {
				assert.True(t, result.Success)
			} else {
				assert.False(t, result.Success)
			}
		})
	}
}

// TestProberTypeICMP_constant tests the constant value.
func TestProberTypeICMP_constant(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "constant_value",
			expected: "icmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify constant matches expected value.
			assert.Equal(t, tt.expected, proberTypeICMP)
		})
	}
}

// parsePort parses a port string to int.
func parsePort(portStr string) int {
	var port int
	for _, c := range portStr {
		port = port*10 + int(c-'0')
	}
	return port
}
