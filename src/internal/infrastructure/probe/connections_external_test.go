//go:build cgo

package probe_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionCollector_Collection(t *testing.T) {
	err := probe.Init()
	require.NoError(t, err)
	defer probe.Shutdown()

	collector := probe.NewConnectionCollector()
	ctx := context.Background()

	tests := []struct {
		name     string
		collect  func() (any, error)
		validate func(t *testing.T, result any)
	}{
		{
			name: "CollectTCP",
			collect: func() (any, error) {
				return collector.CollectTCP(ctx)
			},
			validate: func(t *testing.T, result any) {
				connections := result.([]probe.TcpConnection)
				t.Logf("Found %d TCP connections", len(connections))
				for _, conn := range connections {
					assert.True(t, conn.Family == probe.AddressFamilyIPv4 || conn.Family == probe.AddressFamilyIPv6)
					assert.NotEmpty(t, conn.LocalAddr)
				}
			},
		},
		{
			name: "CollectUDP",
			collect: func() (any, error) {
				return collector.CollectUDP(ctx)
			},
			validate: func(t *testing.T, result any) {
				connections := result.([]probe.UdpConnection)
				t.Logf("Found %d UDP sockets", len(connections))
			},
		},
		{
			name: "CollectUnix",
			collect: func() (any, error) {
				return collector.CollectUnix(ctx)
			},
			validate: func(t *testing.T, result any) {
				sockets := result.([]probe.UnixSocket)
				t.Logf("Found %d Unix sockets", len(sockets))
			},
		},
		{
			name: "CollectTCPStats",
			collect: func() (any, error) {
				return collector.CollectTCPStats(ctx)
			},
			validate: func(t *testing.T, result any) {
				stats := result.(probe.TcpStats)
				t.Logf("TCP Stats: %+v", stats)
				t.Logf("Total connections: %d", stats.Total())
			},
		},
		{
			name: "CollectListeningPorts",
			collect: func() (any, error) {
				return collector.CollectListeningPorts(ctx)
			},
			validate: func(t *testing.T, result any) {
				listening := result.([]probe.TcpConnection)
				t.Logf("Found %d listening ports", len(listening))
				for _, conn := range listening {
					assert.Equal(t, probe.SocketStateListen, conn.State)
					t.Logf("  Port %d (%s) - PID %d (%s)",
						conn.LocalPort, conn.LocalAddr, conn.PID, conn.ProcessName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.collect()
			require.NoError(t, err)
			tt.validate(t, result)
		})
	}
}

func TestSocketState_String(t *testing.T) {
	tests := []struct {
		state    probe.SocketState
		expected string
	}{
		{probe.SocketStateEstablished, "ESTABLISHED"},
		{probe.SocketStateListen, "LISTEN"},
		{probe.SocketStateTimeWait, "TIME_WAIT"},
		{probe.SocketStateUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestAddressFamily_String(t *testing.T) {
	tests := []struct {
		family   probe.AddressFamily
		expected string
	}{
		{probe.AddressFamilyIPv4, "IPv4"},
		{probe.AddressFamilyIPv6, "IPv6"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.family.String())
		})
	}
}
