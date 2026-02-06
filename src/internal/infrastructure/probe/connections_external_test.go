//go:build cgo

package probe_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnectionCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "ReturnsNonNilCollector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewConnectionCollector()
			assert.NotNil(t, collector)
		})
	}
}

func TestConnectionCollector_CollectTCP(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "CollectsConnections"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewConnectionCollector()
			ctx := context.Background()

			connections, err := collector.CollectTCP(ctx)
			require.NoError(t, err)
			t.Logf("Found %d TCP connections", len(connections))
			for _, conn := range connections {
				assert.True(t, conn.Family == probe.AddressFamilyIPv4 || conn.Family == probe.AddressFamilyIPv6)
			}
		})
	}
}

func TestConnectionCollector_CollectUDP(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "CollectsSockets"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewConnectionCollector()
			ctx := context.Background()

			connections, err := collector.CollectUDP(ctx)
			require.NoError(t, err)
			t.Logf("Found %d UDP sockets", len(connections))
		})
	}
}

func TestConnectionCollector_CollectUnix(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "CollectsSockets"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewConnectionCollector()
			ctx := context.Background()

			sockets, err := collector.CollectUnix(ctx)
			require.NoError(t, err)
			t.Logf("Found %d Unix sockets", len(sockets))
		})
	}
}

func TestConnectionCollector_CollectTCPStats(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "CollectsStats"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewConnectionCollector()
			ctx := context.Background()

			stats, err := collector.CollectTCPStats(ctx)
			require.NoError(t, err)
			t.Logf("TCP Stats: %+v", stats)
			t.Logf("Total connections: %d", stats.Total())
		})
	}
}

func TestConnectionCollector_FindProcessByPort(t *testing.T) {
	tests := []struct {
		name string
		port uint16
		tcp  bool
	}{
		{name: "SearchesForPort", port: 80, tcp: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewConnectionCollector()
			ctx := context.Background()

			pid, err := collector.FindProcessByPort(ctx, tt.port, tt.tcp)
			if err != nil {
				t.Logf("Port %d not found: %v", tt.port, err)
			} else {
				t.Logf("Port %d owned by PID %d", tt.port, pid)
			}
		})
	}
}

func TestConnectionCollector_CollectListeningPorts(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "CollectsListeningPorts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			}
		})
	}
}

func TestConnectionCollector_CollectEstablishedConnections(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "CollectsEstablishedConnections"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewConnectionCollector()
			ctx := context.Background()

			established, err := collector.CollectEstablishedConnections(ctx)
			require.NoError(t, err)
			t.Logf("Found %d established connections", len(established))
		})
	}
}

func TestConnectionCollector_CollectProcessConnections(t *testing.T) {
	tests := []struct {
		name string
		pid  int32
	}{
		{name: "CollectsProcessConnections", pid: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewConnectionCollector()
			ctx := context.Background()

			tcp, udp, err := collector.CollectProcessConnections(ctx, tt.pid)
			if err != nil {
				t.Logf("Process %d connections error: %v", tt.pid, err)
			} else {
				t.Logf("Process %d: %d TCP, %d UDP connections", tt.pid, len(tcp), len(udp))
			}
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
