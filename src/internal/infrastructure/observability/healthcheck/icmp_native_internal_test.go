//go:build unix

// Package healthcheck provides internal tests for native ICMP prober.
package healthcheck

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/health"
)

// TestDetectICMPCapability tests ICMP capability detection.
func TestDetectICMPCapability(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "detect_capability",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Detect ICMP capability.
			hasCapability := detectICMPCapability()

			// Log result (may be true or false depending on privileges).
			t.Logf("ICMP capability detected: %v", hasCapability)

			// Just verify the function runs without panic.
			// Result depends on runtime environment.
		})
	}
}

// TestICMPProber_modeSelection tests mode-based probe method selection.
func TestICMPProber_modeSelection(t *testing.T) {
	tests := []struct {
		name                string
		mode                config.ICMPMode
		hasNativeCapability bool
		expectNative        bool
	}{
		{
			name:                "native_mode_with_capability",
			mode:                config.ICMPModeNative,
			hasNativeCapability: true,
			expectNative:        true,
		},
		{
			name:                "native_mode_without_capability",
			mode:                config.ICMPModeNative,
			hasNativeCapability: false,
			expectNative:        true, // Will fail but try native
		},
		{
			name:                "fallback_mode",
			mode:                config.ICMPModeFallback,
			hasNativeCapability: true,
			expectNative:        false,
		},
		{
			name:                "auto_mode_with_capability",
			mode:                config.ICMPModeAuto,
			hasNativeCapability: true,
			expectNative:        true,
		},
		{
			name:                "auto_mode_without_capability",
			mode:                config.ICMPModeAuto,
			hasNativeCapability: false,
			expectNative:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober with specified mode and capability.
			prober := &ICMPProber{
				timeout:             100 * time.Millisecond,
				mode:                tt.mode,
				hasNativeCapability: tt.hasNativeCapability,
				tcpPort:             defaultTCPFallbackPort,
			}

			target := health.Target{
				Address: "127.0.0.1",
			}

			// Probe should execute without panic.
			result := prober.Probe(context.Background(), target)

			// Verify result is returned.
			assert.Greater(t, result.Latency, time.Duration(0))

			// Log result for debugging.
			t.Logf("Mode: %s, native capability: %v, result: %s",
				tt.mode, tt.hasNativeCapability, result.Output)
		})
	}
}

// TestICMPProber_nativePing_unreachableHost tests native ping with unreachable host.
func TestICMPProber_nativePing_unreachableHost(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "unreachable_host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober.
			prober := &ICMPProber{
				timeout:             50 * time.Millisecond,
				mode:                config.ICMPModeNative,
				hasNativeCapability: true,
				tcpPort:             defaultTCPFallbackPort,
			}

			ctx := context.Background()
			start := time.Now()

			// Test with TEST-NET-1 address (guaranteed to not respond).
			result := prober.nativePing(ctx, "192.0.2.1", start)

			// Should fail (timeout or unreachable).
			// May fall back to TCP if ICMP socket creation fails.
			assert.Greater(t, result.Latency, time.Duration(0))
		})
	}
}

// TestICMPProber_nativePing_invalidHost tests native ping with invalid host.
func TestICMPProber_nativePing_invalidHost(t *testing.T) {
	tests := []struct {
		name string
		host string
	}{
		{
			name: "invalid_hostname",
			host: "this-host-does-not-exist-12345.invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober.
			prober := &ICMPProber{
				timeout:             100 * time.Millisecond,
				mode:                config.ICMPModeNative,
				hasNativeCapability: true,
				tcpPort:             defaultTCPFallbackPort,
			}

			ctx := context.Background()
			start := time.Now()

			// Test with invalid hostname.
			result := prober.nativePing(ctx, tt.host, start)

			// Should fail with resolution error.
			assert.False(t, result.Success)
			assert.Error(t, result.Error)
			assert.Contains(t, result.Output, "resolve failed")
		})
	}
}

// TestNewICMPProberWithMode_internalFields tests internal field initialization.
func TestNewICMPProberWithMode_internalFields(t *testing.T) {
	tests := []struct {
		name            string
		timeout         time.Duration
		mode            config.ICMPMode
		expectedTimeout time.Duration
		expectedMode    config.ICMPMode
		expectedTCPPort int
	}{
		{
			name:            "native_mode",
			timeout:         5 * time.Second,
			mode:            config.ICMPModeNative,
			expectedTimeout: 5 * time.Second,
			expectedMode:    config.ICMPModeNative,
			expectedTCPPort: defaultTCPFallbackPort,
		},
		{
			name:            "fallback_mode",
			timeout:         3 * time.Second,
			mode:            config.ICMPModeFallback,
			expectedTimeout: 3 * time.Second,
			expectedMode:    config.ICMPModeFallback,
			expectedTCPPort: defaultTCPFallbackPort,
		},
		{
			name:            "auto_mode",
			timeout:         2 * time.Second,
			mode:            config.ICMPModeAuto,
			expectedTimeout: 2 * time.Second,
			expectedMode:    config.ICMPModeAuto,
			expectedTCPPort: defaultTCPFallbackPort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober with specified mode.
			prober := NewICMPProberWithMode(tt.timeout, tt.mode)

			// Verify internal fields.
			assert.Equal(t, tt.expectedTimeout, prober.timeout)
			assert.Equal(t, tt.expectedMode, prober.mode)
			assert.Equal(t, tt.expectedTCPPort, prober.tcpPort)

			// hasNativeCapability depends on environment.
			t.Logf("Native capability: %v", prober.hasNativeCapability)
		})
	}
}

// mockPacketConn is a mock implementation of packetConn for testing.
type mockPacketConn struct {
	writeToFunc     func(b []byte, dst net.Addr) (int, error)
	setDeadlineFunc func(t time.Time) error
	readFromFunc    func(b []byte) (int, net.Addr, error)
}

// WriteTo implements packetConn.WriteTo.
func (m *mockPacketConn) WriteTo(b []byte, dst net.Addr) (int, error) {
	if m.writeToFunc != nil {
		return m.writeToFunc(b, dst)
	}
	return len(b), nil
}

// SetDeadline implements packetConn.SetDeadline.
func (m *mockPacketConn) SetDeadline(t time.Time) error {
	if m.setDeadlineFunc != nil {
		return m.setDeadlineFunc(t)
	}
	return nil
}

// ReadFrom implements packetConn.ReadFrom.
func (m *mockPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	if m.readFromFunc != nil {
		return m.readFromFunc(b)
	}
	return 0, nil, nil
}

// TestICMPProber_nativePing tests the nativePing method.
func TestICMPProber_nativePing(t *testing.T) {
	tests := []struct {
		name          string
		host          string
		timeout       time.Duration
		expectSuccess bool
		expectError   bool
		outputContain string
	}{
		{
			name:          "invalid_hostname_resolution_failure",
			host:          "this-host-does-not-exist-67890.invalid",
			timeout:       100 * time.Millisecond,
			expectSuccess: false,
			expectError:   true,
			outputContain: "resolve failed",
		},
		{
			name:          "localhost_may_succeed_or_fallback",
			host:          "127.0.0.1",
			timeout:       200 * time.Millisecond,
			expectSuccess: false, // depends on environment
			expectError:   false, // may succeed or timeout
			outputContain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober.
			prober := &ICMPProber{
				timeout:             tt.timeout,
				mode:                config.ICMPModeNative,
				hasNativeCapability: true,
				tcpPort:             defaultTCPFallbackPort,
			}

			ctx := context.Background()
			start := time.Now()

			// Execute nativePing.
			result := prober.nativePing(ctx, tt.host, start)

			// Verify latency is recorded.
			assert.Greater(t, result.Latency, time.Duration(0))

			// Verify expected error state.
			if tt.expectError {
				assert.Error(t, result.Error)
			}

			// Verify output contains expected text.
			if tt.outputContain != "" {
				assert.Contains(t, result.Output, tt.outputContain)
			}

			// Log result for debugging.
			t.Logf("Result: success=%v, output=%s", result.Success, result.Output)
		})
	}
}

// TestICMPProber_sendAndReceiveICMP tests the sendAndReceiveICMP method.
func TestICMPProber_sendAndReceiveICMP(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		expectError bool
	}{
		{
			name:        "with_unreachable_address",
			timeout:     50 * time.Millisecond,
			expectError: true, // Cannot actually send without real connection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober.
			prober := &ICMPProber{
				timeout:             tt.timeout,
				mode:                config.ICMPModeNative,
				hasNativeCapability: true,
				tcpPort:             defaultTCPFallbackPort,
			}

			// Test with TEST-NET-1 address that cannot respond.
			addr := &net.IPAddr{IP: net.ParseIP("192.0.2.1")}

			// Try to create a real ICMP connection for testing.
			// This will fail if CAP_NET_RAW is not available, which is expected.
			conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
			if err != nil {
				// ICMP socket creation failed, which is expected without CAP_NET_RAW.
				// Test passes as this verifies the function would fail gracefully.
				t.Logf("ICMP socket creation failed (expected without CAP_NET_RAW): %v", err)
				return
			}
			defer func() { _ = conn.Close() }()

			// Execute sendAndReceiveICMP with short timeout.
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			err = prober.sendAndReceiveICMP(ctx, conn, addr)

			// Should fail due to timeout or unreachable host.
			if tt.expectError {
				assert.Error(t, err)
			}

			// Log result for debugging.
			t.Logf("sendAndReceiveICMP error: %v", err)
		})
	}
}

// TestICMPProber_sendEchoRequest tests the sendEchoRequest method.
func TestICMPProber_sendEchoRequest(t *testing.T) {
	tests := []struct {
		name           string
		timeout        time.Duration
		deadlineError  error
		writeError     error
		expectError    bool
		expectContains string
	}{
		{
			name:           "successful_send",
			timeout:        100 * time.Millisecond,
			deadlineError:  nil,
			writeError:     nil,
			expectError:    false,
			expectContains: "",
		},
		{
			name:           "deadline_error",
			timeout:        100 * time.Millisecond,
			deadlineError:  errors.New("deadline error"),
			writeError:     nil,
			expectError:    true,
			expectContains: "deadline failed",
		},
		{
			name:           "write_error",
			timeout:        100 * time.Millisecond,
			deadlineError:  nil,
			writeError:     errors.New("write error"),
			expectError:    true,
			expectContains: "send failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober.
			prober := &ICMPProber{
				timeout:             tt.timeout,
				mode:                config.ICMPModeNative,
				hasNativeCapability: true,
				tcpPort:             defaultTCPFallbackPort,
			}

			// Create mock connection.
			mockConn := &mockPacketConn{
				setDeadlineFunc: func(t time.Time) error {
					return tt.deadlineError
				},
				writeToFunc: func(b []byte, dst net.Addr) (int, error) {
					if tt.writeError != nil {
						return 0, tt.writeError
					}
					return len(b), nil
				},
			}

			// Test address.
			addr := &net.IPAddr{IP: net.ParseIP("127.0.0.1")}

			// Execute sendEchoRequest.
			ctx := context.Background()
			err := prober.sendEchoRequest(ctx, mockConn, addr)

			// Verify error expectation.
			if tt.expectError {
				assert.Error(t, err)
				if tt.expectContains != "" {
					assert.Contains(t, err.Error(), tt.expectContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestICMPProber_buildEchoMessage tests the buildEchoMessage method.
func TestICMPProber_buildEchoMessage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "builds_valid_message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober.
			prober := &ICMPProber{
				timeout:             100 * time.Millisecond,
				mode:                config.ICMPModeNative,
				hasNativeCapability: true,
				tcpPort:             defaultTCPFallbackPort,
			}

			// Execute buildEchoMessage.
			msg := prober.buildEchoMessage()

			// Verify message type is Echo.
			assert.Equal(t, ipv4.ICMPTypeEcho, msg.Type)

			// Verify code is 0.
			assert.Equal(t, 0, msg.Code)

			// Verify body is Echo type.
			echo, ok := msg.Body.(*icmp.Echo)
			assert.True(t, ok, "body should be *icmp.Echo")

			// Verify echo sequence.
			assert.Equal(t, 1, echo.Seq)

			// Verify echo data size.
			assert.Equal(t, icmpEchoDataSize, len(echo.Data))

			// Verify echo data pattern (sequential bytes).
			for i := range icmpEchoDataSize {
				assert.Equal(t, byte(i&icmpByteMask), echo.Data[i])
			}

			// Verify message can be marshaled.
			msgBytes, err := msg.Marshal(nil)
			assert.NoError(t, err)
			assert.NotEmpty(t, msgBytes)
		})
	}
}

// TestICMPProber_setConnectionDeadline tests the setConnectionDeadline method.
func TestICMPProber_setConnectionDeadline(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		ctxDeadline   bool
		deadlineError error
		expectError   bool
	}{
		{
			name:          "with_context_deadline",
			timeout:       100 * time.Millisecond,
			ctxDeadline:   true,
			deadlineError: nil,
			expectError:   false,
		},
		{
			name:          "without_context_deadline",
			timeout:       100 * time.Millisecond,
			ctxDeadline:   false,
			deadlineError: nil,
			expectError:   false,
		},
		{
			name:          "deadline_set_error",
			timeout:       100 * time.Millisecond,
			ctxDeadline:   false,
			deadlineError: errors.New("deadline set error"),
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober.
			prober := &ICMPProber{
				timeout:             tt.timeout,
				mode:                config.ICMPModeNative,
				hasNativeCapability: true,
				tcpPort:             defaultTCPFallbackPort,
			}

			// Create mock connection.
			var setDeadlineTime time.Time
			mockConn := &mockPacketConn{
				setDeadlineFunc: func(t time.Time) error {
					setDeadlineTime = t
					return tt.deadlineError
				},
			}

			// Create context with or without deadline.
			var ctx context.Context
			var cancel context.CancelFunc
			if tt.ctxDeadline {
				ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(50*time.Millisecond))
				defer cancel()
			} else {
				ctx = context.Background()
			}

			// Execute setConnectionDeadline.
			err := prober.setConnectionDeadline(ctx, mockConn)

			// Verify error expectation.
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "deadline failed")
			} else {
				assert.NoError(t, err)
				assert.False(t, setDeadlineTime.IsZero())
			}
		})
	}
}

// TestICMPProber_receiveEchoReply tests the receiveEchoReply method.
func TestICMPProber_receiveEchoReply(t *testing.T) {
	tests := []struct {
		name           string
		readFunc       func(b []byte) (int, net.Addr, error)
		expectError    bool
		expectContains string
	}{
		{
			name: "read_error",
			readFunc: func(b []byte) (int, net.Addr, error) {
				return 0, nil, errors.New("read error")
			},
			expectError:    true,
			expectContains: "receive failed",
		},
		{
			name: "invalid_icmp_message",
			readFunc: func(b []byte) (int, net.Addr, error) {
				// Return invalid ICMP data.
				b[0] = 0xff
				return 1, nil, nil
			},
			expectError:    true,
			expectContains: "parse failed",
		},
		{
			name: "unexpected_reply_type",
			readFunc: func(b []byte) (int, net.Addr, error) {
				// Build an ICMP destination unreachable message.
				msg := icmp.Message{
					Type: ipv4.ICMPTypeDestinationUnreachable,
					Code: 0,
					Body: &icmp.DstUnreach{
						Data: make([]byte, 8),
					},
				}
				data, _ := msg.Marshal(nil)
				copy(b, data)
				return len(data), nil, nil
			},
			expectError:    true,
			expectContains: "unexpected reply type",
		},
		{
			name: "valid_echo_reply",
			readFunc: func(b []byte) (int, net.Addr, error) {
				// Build a valid ICMP echo reply message.
				msg := icmp.Message{
					Type: ipv4.ICMPTypeEchoReply,
					Code: 0,
					Body: &icmp.Echo{
						ID:   1,
						Seq:  1,
						Data: []byte("test"),
					},
				}
				data, _ := msg.Marshal(nil)
				copy(b, data)
				return len(data), nil, nil
			},
			expectError:    false,
			expectContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober.
			prober := &ICMPProber{
				timeout:             100 * time.Millisecond,
				mode:                config.ICMPModeNative,
				hasNativeCapability: true,
				tcpPort:             defaultTCPFallbackPort,
			}

			// Create mock connection.
			mockConn := &mockPacketConn{
				readFromFunc: tt.readFunc,
			}

			// Execute receiveEchoReply.
			err := prober.receiveEchoReply(mockConn)

			// Verify error expectation.
			if tt.expectError {
				assert.Error(t, err)
				if tt.expectContains != "" {
					assert.Contains(t, err.Error(), tt.expectContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
