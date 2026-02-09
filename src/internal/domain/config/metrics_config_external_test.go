package config_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/stretchr/testify/assert"
)

// TestDefaultMetricsConfig verifies the default metrics configuration.
func TestDefaultMetricsConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "default config enables all metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultMetricsConfig()

			// Verify global enabled
			assert.True(t, cfg.Enabled)

			// Verify all categories enabled
			assert.True(t, cfg.CPU.Enabled)
			assert.True(t, cfg.CPU.Pressure)
			assert.True(t, cfg.Memory.Enabled)
			assert.True(t, cfg.Memory.Pressure)
			assert.True(t, cfg.Load.Enabled)
			assert.True(t, cfg.Disk.Enabled)
			assert.True(t, cfg.Disk.Partitions)
			assert.True(t, cfg.Disk.Usage)
			assert.True(t, cfg.Disk.IO)
			assert.True(t, cfg.Network.Enabled)
			assert.True(t, cfg.Network.Interfaces)
			assert.True(t, cfg.Network.Stats)
			assert.True(t, cfg.Connections.Enabled)
			assert.True(t, cfg.Connections.TCPStats)
			assert.True(t, cfg.Connections.TCPConnections)
			assert.True(t, cfg.Connections.UDPSockets)
			assert.True(t, cfg.Connections.UnixSockets)
			assert.True(t, cfg.Connections.ListeningPorts)
			assert.True(t, cfg.Thermal.Enabled)
			assert.True(t, cfg.Process.Enabled)
			assert.True(t, cfg.IO.Enabled)
			assert.True(t, cfg.IO.Pressure)
			assert.True(t, cfg.Quota.Enabled)
			assert.True(t, cfg.Container.Enabled)
			assert.True(t, cfg.Runtime.Enabled)
		})
	}
}

// TestStandardMetricsConfig verifies the standard template.
func TestStandardMetricsConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "standard template matches default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std := config.StandardMetricsConfig()
			def := config.DefaultMetricsConfig()

			// Standard should match default
			assert.Equal(t, def, std)
		})
	}
}

// TestMinimalMetricsConfig verifies the minimal template.
func TestMinimalMetricsConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "minimal template enables only essentials"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.MinimalMetricsConfig()

			// Verify global enabled
			assert.True(t, cfg.Enabled)

			// Verify essential metrics enabled
			assert.True(t, cfg.CPU.Enabled)
			assert.False(t, cfg.CPU.Pressure) // No pressure in minimal
			assert.True(t, cfg.Memory.Enabled)
			assert.False(t, cfg.Memory.Pressure) // No pressure in minimal
			assert.True(t, cfg.Load.Enabled)

			// Verify all other metrics disabled
			assert.False(t, cfg.Disk.Enabled)
			assert.False(t, cfg.Network.Enabled)
			assert.False(t, cfg.Connections.Enabled)
			assert.False(t, cfg.Thermal.Enabled)
			assert.False(t, cfg.Process.Enabled)
			assert.False(t, cfg.IO.Enabled)
			assert.False(t, cfg.Quota.Enabled)
			assert.False(t, cfg.Container.Enabled)
			assert.False(t, cfg.Runtime.Enabled)
		})
	}
}

// TestFullMetricsConfig verifies the full template.
func TestFullMetricsConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "full template matches standard"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			full := config.FullMetricsConfig()
			std := config.StandardMetricsConfig()

			// Full should match standard for forward compatibility
			assert.Equal(t, std, full)
		})
	}
}

// TestNewMonitoringConfig verifies monitoring config includes metrics.
func TestNewMonitoringConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "new monitoring config includes default metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mon := config.NewMonitoringConfig()

			// Verify metrics config is initialized
			assert.True(t, mon.Metrics.Enabled)
			assert.Equal(t, config.DefaultMetricsConfig(), mon.Metrics)
		})
	}
}
