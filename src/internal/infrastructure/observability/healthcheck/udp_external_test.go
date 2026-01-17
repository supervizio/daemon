// Package healthcheck_test provides black-box tests for the probe package.
package healthcheck_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/infrastructure/observability/healthcheck"
)

// TestNewUDPProber tests UDP prober creation.
func TestNewUDPProber(t *testing.T) {
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
			// Create UDP prober with specified timeout.
			prober := healthcheck.NewUDPProber(tt.timeout)

			// Verify prober is created.
			require.NotNil(t, prober)
		})
	}
}

// TestNewUDPProberWithPayload tests UDP prober creation with custom payload.
func TestNewUDPProberWithPayload(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		payload []byte
	}{
		{
			name:    "custom_payload",
			timeout: time.Second,
			payload: []byte("HELLO"),
		},
		{
			name:    "empty_payload",
			timeout: time.Second,
			payload: nil,
		},
		{
			name:    "binary_payload",
			timeout: time.Second,
			payload: []byte{0x00, 0x01, 0x02},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create UDP prober with custom payload.
			prober := healthcheck.NewUDPProberWithPayload(tt.timeout, tt.payload)

			// Verify prober is created.
			require.NotNil(t, prober)
		})
	}
}

// TestUDPProber_Type tests the Type method.
func TestUDPProber_Type(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns_udp",
			expected: "udp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create UDP prober.
			prober := healthcheck.NewUDPProber(time.Second)

			// Verify type identifier.
			assert.Equal(t, tt.expected, prober.Type())
		})
	}
}

// TestUDPProber_Probe tests UDP probing functionality.
// This test spawns a background goroutine to echo UDP packets.
// The goroutine terminates when the connection is closed via defer.
func TestUDPProber_Probe(t *testing.T) {
	// Start a test UDP server that echoes back.
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	require.NoError(t, err)

	conn, err := net.ListenUDP("udp", serverAddr)
	require.NoError(t, err)
	defer func() {
		// Cleanup connection.
		_ = conn.Close()
	}()

	// Echo server in background goroutine.
	// Goroutine terminates when conn.ReadFromUDP returns error on Close.
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, addr, readErr := conn.ReadFromUDP(buffer)
			// Check if connection was closed.
			if readErr != nil {
				// Connection closed, terminate goroutine.
				return
			}
			// Echo back the data.
			_, _ = conn.WriteToUDP(buffer[:n], addr)
		}
	}()

	tests := []struct {
		name          string
		target        health.Target
		timeout       time.Duration
		expectSuccess bool
	}{
		{
			name: "successful_with_response",
			target: health.Target{
				Address: conn.LocalAddr().String(),
			},
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "successful_with_explicit_network",
			target: health.Target{
				Address: conn.LocalAddr().String(),
				Network: "udp",
			},
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "unreachable_port",
			target: health.Target{
				Address: "127.0.0.1:1",
			},
			timeout:       50 * time.Millisecond,
			expectSuccess: false, // Port 1 typically causes immediate connection refused
		},
		{
			name: "invalid_address",
			target: health.Target{
				Address: "invalid:address:format",
			},
			timeout:       100 * time.Millisecond,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create UDP prober.
			prober := healthcheck.NewUDPProber(tt.timeout)
			ctx := context.Background()

			// Perform healthcheck.
			result := prober.Probe(ctx, tt.target)

			// Verify result based on expected outcome.
			if tt.expectSuccess {
				assert.True(t, result.Success)
			} else {
				assert.False(t, result.Success)
			}

			// Latency should always be measured.
			assert.Greater(t, result.Latency, time.Duration(0))
		})
	}
}

// TestUDPProber_Probe_WithCustomPayload tests probing with custom payload.
// This test spawns a background goroutine to echo UDP packets.
// The goroutine terminates when the connection is closed via defer.
func TestUDPProber_Probe_WithCustomPayload(t *testing.T) {
	// Start a test UDP server.
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	require.NoError(t, err)

	conn, err := net.ListenUDP("udp", serverAddr)
	require.NoError(t, err)
	defer func() {
		// Cleanup connection.
		_ = conn.Close()
	}()

	// Echo server in background goroutine.
	// Goroutine terminates when conn.ReadFromUDP returns error on Close.
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, addr, readErr := conn.ReadFromUDP(buffer)
			// Check if connection was closed.
			if readErr != nil {
				// Connection closed, terminate goroutine.
				return
			}
			// Echo back the data.
			_, _ = conn.WriteToUDP(buffer[:n], addr)
		}
	}()

	tests := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "custom_text_payload",
			payload: []byte("HEALTH_CHECK"),
		},
		{
			name:    "single_byte_payload",
			payload: []byte{0x01},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create UDP prober with custom payload.
			prober := healthcheck.NewUDPProberWithPayload(time.Second, tt.payload)

			target := health.Target{
				Address: conn.LocalAddr().String(),
			}

			// Perform healthcheck.
			result := prober.Probe(context.Background(), target)

			// Should succeed with response.
			assert.True(t, result.Success)
		})
	}
}
