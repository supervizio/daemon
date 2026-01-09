// Package probe provides internal tests for UDP prober.
package probe

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	domainprobe "github.com/kodflow/daemon/internal/domain/probe"
)

// TestUDPProber_internalFields tests internal struct fields.
func TestUDPProber_internalFields(t *testing.T) {
	tests := []struct {
		name            string
		timeout         time.Duration
		expectedTimeout time.Duration
	}{
		{
			name:            "timeout_is_stored",
			timeout:         5 * time.Second,
			expectedTimeout: 5 * time.Second,
		},
		{
			name:            "zero_timeout_is_stored",
			timeout:         0,
			expectedTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create UDP prober.
			prober := NewUDPProber(tt.timeout)

			// Verify internal timeout field.
			assert.Equal(t, tt.expectedTimeout, prober.timeout)
		})
	}
}

// TestUDPProber_internalPayload tests internal payload field.
func TestUDPProber_internalPayload(t *testing.T) {
	tests := []struct {
		name            string
		payload         []byte
		expectedPayload []byte
	}{
		{
			name:            "default_payload",
			payload:         nil,
			expectedPayload: defaultUDPPayload,
		},
		{
			name:            "custom_payload",
			payload:         []byte("CUSTOM"),
			expectedPayload: []byte("CUSTOM"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var prober *UDPProber
			if tt.payload == nil {
				// Use default constructor.
				prober = NewUDPProber(time.Second)
			} else {
				// Use custom payload constructor.
				prober = NewUDPProberWithPayload(time.Second, tt.payload)
			}

			// Verify internal payload field.
			assert.Equal(t, tt.expectedPayload, prober.payload)
		})
	}
}

// TestProberTypeUDP_constant tests the constant value.
func TestProberTypeUDP_constant(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "constant_value",
			expected: "udp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify constant matches expected value.
			assert.Equal(t, tt.expected, proberTypeUDP)
		})
	}
}

// TestDefaultUDPPayload tests the default payload value.
func TestDefaultUDPPayload(t *testing.T) {
	tests := []struct {
		name     string
		expected []byte
	}{
		{
			name:     "default_is_ping",
			expected: []byte("PING"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify default payload.
			assert.Equal(t, tt.expected, defaultUDPPayload)
		})
	}
}

// Test_UDPProber_dialUDP tests the internal dialUDP method.
func Test_UDPProber_dialUDP(t *testing.T) {
	tests := []struct {
		name          string
		target        domainprobe.Target
		timeout       time.Duration
		expectConn    bool
		expectSuccess bool
	}{
		{
			name: "valid_address",
			target: domainprobe.Target{
				Address: "127.0.0.1:12345",
			},
			timeout:       time.Second,
			expectConn:    true,
			expectSuccess: true,
		},
		{
			name: "valid_address_with_network",
			target: domainprobe.Target{
				Address: "127.0.0.1:12345",
				Network: "udp",
			},
			timeout:       time.Second,
			expectConn:    true,
			expectSuccess: true,
		},
		{
			name: "invalid_address_format",
			target: domainprobe.Target{
				Address: "invalid:address:format:extra",
			},
			timeout:       time.Second,
			expectConn:    false,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create UDP prober.
			prober := NewUDPProber(tt.timeout)
			start := time.Now()

			// Call internal method.
			conn, result := prober.dialUDP(tt.target, start)

			// Verify result.
			if tt.expectConn {
				// Expect connection to be returned.
				assert.NotNil(t, conn)
				// Clean up connection.
				if conn != nil {
					_ = conn.Close()
				}
			} else {
				// Expect nil connection with failure result.
				assert.Nil(t, conn)
				assert.False(t, result.Success)
			}
		})
	}
}

// Test_UDPProber_sendAndReceive tests the internal sendAndReceive method.
// This test uses an echo server to verify UDP communication.
// The goroutine terminates when the server connection is closed via defer.
func Test_UDPProber_sendAndReceive(t *testing.T) {
	// Start a UDP echo server.
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	// Check if address resolution failed.
	if err != nil {
		t.Fatalf("failed to resolve address: %v", err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	// Check if server creation failed.
	if err != nil {
		t.Fatalf("failed to create UDP server: %v", err)
	}
	// Close server when test completes.
	defer func() {
		// Cleanup server.
		_ = serverConn.Close()
	}()

	// Echo server goroutine.
	// Goroutine terminates when serverConn.ReadFromUDP returns error on Close.
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, remoteAddr, readErr := serverConn.ReadFromUDP(buffer)
			// Check if server was closed.
			if readErr != nil {
				// Server closed, terminate goroutine.
				return
			}
			// Echo back the data.
			_, _ = serverConn.WriteToUDP(buffer[:n], remoteAddr)
		}
	}()

	tests := []struct {
		name          string
		timeout       time.Duration
		expectSuccess bool
	}{
		{
			name:          "successful_echo",
			timeout:       time.Second,
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create UDP prober.
			prober := NewUDPProber(tt.timeout)

			// Create client connection.
			clientAddr, err := net.ResolveUDPAddr("udp", serverConn.LocalAddr().String())
			// Check if address resolution failed.
			if err != nil {
				t.Fatalf("failed to resolve client address: %v", err)
			}

			clientConn, err := net.DialUDP("udp", nil, clientAddr)
			// Check if client connection failed.
			if err != nil {
				t.Fatalf("failed to dial UDP: %v", err)
			}
			// Close client when test case completes.
			defer func() {
				// Cleanup client.
				_ = clientConn.Close()
			}()

			// Set deadline.
			deadline := time.Now().Add(tt.timeout)
			// Check if deadline configuration fails.
			if err := clientConn.SetDeadline(deadline); err != nil {
				t.Fatalf("failed to set deadline: %v", err)
			}

			start := time.Now()

			// Call internal method.
			result := prober.sendAndReceive(clientConn, serverConn.LocalAddr().String(), start)

			// Verify result.
			if tt.expectSuccess {
				assert.True(t, result.Success)
				assert.Contains(t, result.Output, "received")
			} else {
				assert.False(t, result.Success)
			}
		})
	}
}

// Test_UDPProber_handleReadResult tests the internal handleReadResult method.
func Test_UDPProber_handleReadResult(t *testing.T) {
	// Define a timeout error for testing.
	timeoutErr := &net.OpError{
		Op:  "read",
		Err: &timeoutError{},
	}

	tests := []struct {
		name          string
		err           error
		bytesRead     int
		address       string
		latency       time.Duration
		expectSuccess bool
		expectOutput  string
	}{
		{
			name:          "successful_read",
			err:           nil,
			bytesRead:     10,
			address:       "127.0.0.1:1234",
			latency:       100 * time.Millisecond,
			expectSuccess: true,
			expectOutput:  "received 10 bytes",
		},
		{
			name:          "timeout_is_success",
			err:           timeoutErr,
			bytesRead:     0,
			address:       "127.0.0.1:1234",
			latency:       100 * time.Millisecond,
			expectSuccess: true,
			expectOutput:  "no response within timeout",
		},
		{
			name:          "other_error_is_failure",
			err:           errors.New("connection refused"),
			bytesRead:     0,
			address:       "127.0.0.1:1234",
			latency:       100 * time.Millisecond,
			expectSuccess: false,
			expectOutput:  "failed to read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create UDP prober.
			prober := NewUDPProber(time.Second)

			// Call internal method.
			result := prober.handleReadResult(tt.err, tt.bytesRead, tt.address, tt.latency)

			// Verify result.
			assert.Equal(t, tt.expectSuccess, result.Success)
			assert.Contains(t, result.Output, tt.expectOutput)
			assert.Equal(t, tt.latency, result.Latency)
		})
	}
}

// timeoutError implements net.Error for testing timeout handling.
type timeoutError struct{}

// Error returns the error message.
func (e *timeoutError) Error() string {
	// Return timeout error message.
	return "i/o timeout"
}

// Timeout returns true indicating this is a timeout error.
func (e *timeoutError) Timeout() bool {
	// This is a timeout error.
	return true
}

// Temporary returns false.
func (e *timeoutError) Temporary() bool {
	// This is not a temporary error.
	return false
}
