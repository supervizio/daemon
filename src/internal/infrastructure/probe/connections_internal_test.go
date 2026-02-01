//go:build cgo

package probe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSocketState_String verifies SocketState string conversion.
func TestSocketState_String_Internal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		state    SocketState
		expected string
	}{
		{SocketStateUnknown, "UNKNOWN"},
		{SocketStateEstablished, "ESTABLISHED"},
		{SocketStateSynSent, "SYN_SENT"},
		{SocketStateSynRecv, "SYN_RECV"},
		{SocketStateFinWait1, "FIN_WAIT1"},
		{SocketStateFinWait2, "FIN_WAIT2"},
		{SocketStateTimeWait, "TIME_WAIT"},
		{SocketStateClose, "CLOSE"},
		{SocketStateCloseWait, "CLOSE_WAIT"},
		{SocketStateLastAck, "LAST_ACK"},
		{SocketStateListen, "LISTEN"},
		{SocketStateClosing, "CLOSING"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

// TestSocketState_String_Unknown verifies unknown state handling.
func TestSocketState_String_Unknown(t *testing.T) {
	t.Parallel()

	unknown := SocketState(255)
	assert.Equal(t, "UNKNOWN", unknown.String())
}

// TestAddressFamily_String_Internal verifies AddressFamily string conversion.
func TestAddressFamily_String_Internal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		family   AddressFamily
		expected string
	}{
		{AddressFamilyIPv4, "IPv4"},
		{AddressFamilyIPv6, "IPv6"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.family.String())
		})
	}
}

// TestAddressFamily_String_Unknown verifies unknown family handling.
func TestAddressFamily_String_Unknown(t *testing.T) {
	t.Parallel()

	unknown := AddressFamily(99)
	assert.Equal(t, "Unknown", unknown.String())
}

// TestTcpStats_Total verifies TcpStats total calculation.
func TestTcpStats_Total(t *testing.T) {
	t.Parallel()

	stats := &TcpStats{
		Established: 10,
		SynSent:     1,
		SynRecv:     2,
		FinWait1:    3,
		FinWait2:    4,
		TimeWait:    5,
		Close:       6,
		CloseWait:   7,
		LastAck:     8,
		Listen:      9,
		Closing:     0,
	}

	expected := uint32(10 + 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 0)
	assert.Equal(t, expected, stats.Total())
}

// TestNewTcpStats verifies TcpStats constructor.
func TestNewTcpStats(t *testing.T) {
	t.Parallel()

	stats := NewTcpStats()
	assert.NotNil(t, stats)
	assert.Equal(t, uint32(0), stats.Total())
}

// TestNewConnectionCollector verifies ConnectionCollector constructor.
func TestNewConnectionCollector(t *testing.T) {
	t.Parallel()

	collector := NewConnectionCollector()
	assert.NotNil(t, collector)
}

// TestNotFoundPIDConstant verifies notFoundPID constant.
func TestNotFoundPIDConstant(t *testing.T) {
	t.Parallel()

	assert.Equal(t, int32(-1), notFoundPID)
}

// TestSocketStateNames verifies all states are mapped.
func TestSocketStateNames(t *testing.T) {
	t.Parallel()

	states := []SocketState{
		SocketStateUnknown,
		SocketStateEstablished,
		SocketStateSynSent,
		SocketStateSynRecv,
		SocketStateFinWait1,
		SocketStateFinWait2,
		SocketStateTimeWait,
		SocketStateClose,
		SocketStateCloseWait,
		SocketStateLastAck,
		SocketStateListen,
		SocketStateClosing,
	}

	for _, state := range states {
		_, ok := socketStateNames[state]
		assert.True(t, ok, "state %d should be mapped", state)
	}
}

// TestConnectionCollector_CollectTCP verifies TCP connection collection.
func TestConnectionCollector_CollectTCP(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewConnectionCollector()
			ctx := context.Background()

			conns, err := collector.CollectTCP(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// TCP connections may be empty, just verify no error
				_ = conns
			}
		})
	}
}

// TestConnectionCollector_CollectUDP verifies UDP connection collection.
func TestConnectionCollector_CollectUDP(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewConnectionCollector()
			ctx := context.Background()

			conns, err := collector.CollectUDP(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// UDP connections may be empty, just verify no error
				_ = conns
			}
		})
	}
}

// TestConnectionCollector_CollectUnix verifies Unix socket collection.
func TestConnectionCollector_CollectUnix(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewConnectionCollector()
			ctx := context.Background()

			socks, err := collector.CollectUnix(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// Unix sockets may be empty, just verify no error
				_ = socks
			}
		})
	}
}

// TestConnectionCollector_CollectTCPStats verifies TCP stats collection.
func TestConnectionCollector_CollectTCPStats(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewConnectionCollector()
			ctx := context.Background()

			stats, err := collector.CollectTCPStats(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, stats)
			}
		})
	}
}

// TestConnectionCollector_CollectListeningPorts verifies listening port collection.
func TestConnectionCollector_CollectListeningPorts(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewConnectionCollector()
			ctx := context.Background()

			ports, err := collector.CollectListeningPorts(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// Listening ports may be empty, just verify no error
				_ = ports
			}
		})
	}
}

// TestConnectionCollector_CollectEstablishedConnections verifies established connection collection.
func TestConnectionCollector_CollectEstablishedConnections(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewConnectionCollector()
			ctx := context.Background()

			conns, err := collector.CollectEstablishedConnections(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// Established connections may be empty
				_ = conns
			}
		})
	}
}

// TestConnectionCollector_CollectProcessConnections verifies process connection collection.
func TestConnectionCollector_CollectProcessConnections(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		pid         int32
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			pid:         1,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			pid:         1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewConnectionCollector()
			ctx := context.Background()

			tcp, udp, err := collector.CollectProcessConnections(ctx, tt.pid)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// Process connections may be empty
				_ = tcp
				_ = udp
			}
		})
	}
}

// TestConnectionCollector_FindProcessByPort verifies process port lookup.
func TestConnectionCollector_FindProcessByPort(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		port        uint16
		tcp         bool
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			port:        80,
			tcp:         true,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			port:        80,
			tcp:         true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewConnectionCollector()
			ctx := context.Background()

			pid, err := collector.FindProcessByPort(ctx, tt.port, tt.tcp)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Port 80 may not be in use, just verify no panic
				_ = pid
			}
		})
	}
}
