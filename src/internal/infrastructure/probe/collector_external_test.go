//go:build cgo

package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCollector verifies collector creation.
func TestNewCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewCollector()
			require.NotNil(t, collector)
		})
	}
}

// TestCollector_Cpu verifies CPU collector access.
func TestCollector_Cpu(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns non-nil CPU collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewCollector()
			cpu := collector.Cpu()
			assert.NotNil(t, cpu)
		})
	}
}

// TestCollector_Memory verifies memory collector access.
func TestCollector_Memory(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns non-nil memory collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewCollector()
			memory := collector.Memory()
			assert.NotNil(t, memory)
		})
	}
}

// TestCollector_Disk verifies disk collector access.
func TestCollector_Disk(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns non-nil disk collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewCollector()
			disk := collector.Disk()
			assert.NotNil(t, disk)
		})
	}
}

// TestCollector_Network verifies network collector access.
func TestCollector_Network(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns non-nil network collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewCollector()
			network := collector.Network()
			assert.NotNil(t, network)
		})
	}
}

// TestCollector_Io verifies I/O collector access.
func TestCollector_Io(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns non-nil I/O collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewCollector()
			io := collector.Io()
			assert.NotNil(t, io)
		})
	}
}

// TestCollector_Connection verifies connection collector access.
func TestCollector_Connection(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns non-nil connection collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewCollector()
			conn := collector.Connection()
			assert.NotNil(t, conn)
		})
	}
}

// TestCollector_ImplementsInterface verifies the collector implements SystemCollector.
func TestCollector_ImplementsInterface(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "all sub-collectors are accessible"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewCollector()

			// Verify sub-collectors are not nil
			assert.NotNil(t, collector.Cpu())
			assert.NotNil(t, collector.Memory())
			assert.NotNil(t, collector.Disk())
			assert.NotNil(t, collector.Network())
			assert.NotNil(t, collector.Io())
		})
	}
}
