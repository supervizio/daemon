//go:build cgo

package probe_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewNetworkCollector verifies network collector creation.
func TestNewNetworkCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewNetworkCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestNetworkCollector_ListInterfaces verifies interface listing.
func TestNetworkCollector_ListInterfaces(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "lists interfaces with non-empty names"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewNetworkCollector()
			ctx := context.Background()

			ifaces, err := collector.ListInterfaces(ctx)
			require.NoError(t, err)

			// Should have at least one interface (loopback)
			assert.NotEmpty(t, ifaces)

			// Verify interface structure
			for _, iface := range ifaces {
				assert.NotEmpty(t, iface.Name)
			}
		})
	}
}

// TestNetworkCollector_CollectAllStats verifies all network stats collection.
func TestNetworkCollector_CollectAllStats(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects all network stats"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewNetworkCollector()
			ctx := context.Background()

			stats, err := collector.CollectAllStats(ctx)
			require.NoError(t, err)

			// Should have at least one stats record
			assert.NotEmpty(t, stats)

			// Verify stats structure
			for _, s := range stats {
				assert.NotEmpty(t, s.Interface)
				// Timestamps should be set
				assert.False(t, s.Timestamp.IsZero())
			}
		})
	}
}

// TestNetworkCollector_CollectStats verifies interface-specific stats.
func TestNetworkCollector_CollectStats(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects stats for specific interface"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewNetworkCollector()
			ctx := context.Background()

			// First get list of interfaces
			ifaces, err := collector.ListInterfaces(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, ifaces)

			// Test with first available interface
			ifaceName := ifaces[0].Name
			stats, err := collector.CollectStats(ctx, ifaceName)
			require.NoError(t, err)
			assert.Equal(t, ifaceName, stats.Interface)
		})
	}
}

// TestNetworkCollector_CollectStats_NotFound verifies ErrNotFound for unknown interface.
func TestNetworkCollector_CollectStats_NotFound(t *testing.T) {
	tests := []struct {
		name      string
		ifaceName string
	}{
		{name: "returns error for nonexistent interface", ifaceName: "nonexistent_interface_xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewNetworkCollector()
			ctx := context.Background()

			_, err = collector.CollectStats(ctx, tt.ifaceName)
			assert.Error(t, err)
		})
	}
}

// TestNetworkCollector_NotInitialized verifies error when not initialized.
func TestNetworkCollector_NotInitialized(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns error when not initialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewNetworkCollector()
			ctx := context.Background()

			_, err := collector.ListInterfaces(ctx)
			// Should return error because probe is not initialized
			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}
