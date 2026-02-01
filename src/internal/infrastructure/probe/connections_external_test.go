//go:build cgo

package probe_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionCollector_CollectTCP(t *testing.T) {
	err := probe.Init()
	require.NoError(t, err)
	defer probe.Shutdown()

	collector := probe.NewConnectionCollector()
	ctx := context.Background()

	connections, err := collector.CollectTCP(ctx)
	require.NoError(t, err)

	// We should have at least some TCP connections on any system
	t.Logf("Found %d TCP connections", len(connections))

	// Verify structure of connections
	for _, conn := range connections {
		assert.True(t, conn.Family == probe.AddressFamilyIPv4 || conn.Family == probe.AddressFamilyIPv6)
		assert.NotEmpty(t, conn.LocalAddr)
	}
}

func TestConnectionCollector_CollectUDP(t *testing.T) {
	err := probe.Init()
	require.NoError(t, err)
	defer probe.Shutdown()

	collector := probe.NewConnectionCollector()
	ctx := context.Background()

	connections, err := collector.CollectUDP(ctx)
	require.NoError(t, err)

	t.Logf("Found %d UDP sockets", len(connections))
}

func TestConnectionCollector_CollectUnix(t *testing.T) {
	err := probe.Init()
	require.NoError(t, err)
	defer probe.Shutdown()

	collector := probe.NewConnectionCollector()
	ctx := context.Background()

	sockets, err := collector.CollectUnix(ctx)
	require.NoError(t, err)

	t.Logf("Found %d Unix sockets", len(sockets))
}

func TestConnectionCollector_CollectTCPStats(t *testing.T) {
	err := probe.Init()
	require.NoError(t, err)
	defer probe.Shutdown()

	collector := probe.NewConnectionCollector()
	ctx := context.Background()

	stats, err := collector.CollectTCPStats(ctx)
	require.NoError(t, err)

	t.Logf("TCP Stats: %+v", stats)
	t.Logf("Total connections: %d", stats.Total())
}

func TestConnectionCollector_CollectListeningPorts(t *testing.T) {
	err := probe.Init()
	require.NoError(t, err)
	defer probe.Shutdown()

	collector := probe.NewConnectionCollector()
	ctx := context.Background()

	listening, err := collector.CollectListeningPorts(ctx)
	require.NoError(t, err)

	t.Logf("Found %d listening ports", len(listening))

	for _, conn := range listening {
		assert.Equal(t, probe.SocketStateListen, conn.State)
		t.Logf("  Port %d (%s) - PID %d (%s)",
			conn.LocalPort, conn.LocalAddr, conn.PID, conn.ProcessName)
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
	assert.Equal(t, "IPv4", probe.AddressFamilyIPv4.String())
	assert.Equal(t, "IPv6", probe.AddressFamilyIPv6.String())
}
