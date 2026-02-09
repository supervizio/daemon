// Package config provides domain value objects for service configuration.
package config

// MetricsTemplate defines preset configurations for common use cases.
type MetricsTemplate string

const (
	// MetricsTemplateMinimal enables only essential metrics (CPU, memory, load).
	// Provides 70-80% allocation reduction compared to standard.
	MetricsTemplateMinimal MetricsTemplate = "minimal"

	// MetricsTemplateStandard enables all metrics (default behavior).
	// Equivalent to existing behavior with all collection enabled.
	MetricsTemplateStandard MetricsTemplate = "standard"

	// MetricsTemplateFull enables all metrics explicitly.
	// Identical to standard, provided for forward compatibility.
	MetricsTemplateFull MetricsTemplate = "full"

	// MetricsTemplateCustom indicates user-defined granular configuration.
	// Used when explicit metrics config overrides template.
	MetricsTemplateCustom MetricsTemplate = "custom"
)

// MetricsConfig defines granular control over metrics collection.
// Each category can be independently enabled/disabled to reduce resource consumption.
type MetricsConfig struct {
	// Enabled is the global metrics toggle.
	// When false, no metrics are collected regardless of category settings.
	Enabled bool

	// CPU configures CPU metrics collection.
	CPU CPUMetricsConfig

	// Memory configures memory metrics collection.
	Memory MemoryMetricsConfig

	// Load configures load average metrics collection.
	Load LoadMetricsConfig

	// Disk configures disk metrics collection.
	Disk DiskMetricsConfig

	// Network configures network metrics collection.
	Network NetworkMetricsConfig

	// Connections configures connection metrics collection.
	// This is the most expensive category on busy systems.
	Connections ConnectionMetricsConfig

	// Thermal configures thermal sensor metrics collection (Linux only).
	Thermal ThermalMetricsConfig

	// Process configures process metrics collection.
	Process ProcessMetricsConfig

	// IO configures I/O metrics collection.
	IO IOMetricsConfig

	// Quota configures resource quota metrics collection.
	Quota QuotaMetricsConfig

	// Container configures container detection metrics.
	Container ContainerMetricsConfig

	// Runtime configures runtime detection metrics.
	Runtime RuntimeMetricsConfig
}

// CPUMetricsConfig defines CPU metrics collection settings.
type CPUMetricsConfig struct {
	// Enabled controls CPU metrics collection.
	Enabled bool

	// Pressure controls PSI (Pressure Stall Information) collection.
	Pressure bool
}

// MemoryMetricsConfig defines memory metrics collection settings.
type MemoryMetricsConfig struct {
	// Enabled controls memory metrics collection.
	Enabled bool

	// Pressure controls PSI (Pressure Stall Information) collection.
	Pressure bool
}

// LoadMetricsConfig defines load average metrics collection settings.
type LoadMetricsConfig struct {
	// Enabled controls load average collection.
	Enabled bool
}

// DiskMetricsConfig defines disk metrics collection settings.
type DiskMetricsConfig struct {
	// Enabled controls disk metrics collection.
	Enabled bool

	// Partitions controls partition enumeration.
	Partitions bool

	// Usage controls disk usage collection per partition.
	Usage bool

	// IO controls disk I/O statistics collection.
	IO bool
}

// NetworkMetricsConfig defines network metrics collection settings.
type NetworkMetricsConfig struct {
	// Enabled controls network metrics collection.
	Enabled bool

	// Interfaces controls interface enumeration.
	Interfaces bool

	// Stats controls per-interface statistics collection.
	Stats bool
}

// ConnectionMetricsConfig defines connection metrics collection settings.
// This is the most expensive category, especially on systems with many connections.
type ConnectionMetricsConfig struct {
	// Enabled controls connection metrics collection.
	Enabled bool

	// TCPStats controls aggregated TCP statistics collection.
	TCPStats bool

	// TCPConnections controls individual TCP connection enumeration.
	// This is expensive on busy systems with thousands of connections.
	TCPConnections bool

	// UDPSockets controls UDP socket enumeration.
	UDPSockets bool

	// UnixSockets controls Unix socket enumeration.
	UnixSockets bool

	// ListeningPorts controls listening port enumeration.
	ListeningPorts bool
}

// ThermalMetricsConfig defines thermal metrics collection settings.
type ThermalMetricsConfig struct {
	// Enabled controls thermal sensor collection (Linux only).
	Enabled bool
}

// ProcessMetricsConfig defines process metrics collection settings.
type ProcessMetricsConfig struct {
	// Enabled controls process metrics collection.
	Enabled bool
}

// IOMetricsConfig defines I/O metrics collection settings.
type IOMetricsConfig struct {
	// Enabled controls I/O metrics collection.
	Enabled bool

	// Pressure controls PSI (Pressure Stall Information) collection.
	Pressure bool
}

// QuotaMetricsConfig defines resource quota metrics collection settings.
type QuotaMetricsConfig struct {
	// Enabled controls resource quota collection.
	Enabled bool
}

// ContainerMetricsConfig defines container detection metrics settings.
type ContainerMetricsConfig struct {
	// Enabled controls container detection.
	Enabled bool
}

// RuntimeMetricsConfig defines runtime detection metrics settings.
type RuntimeMetricsConfig struct {
	// Enabled controls runtime detection.
	Enabled bool
}

// DefaultMetricsConfig returns the standard template configuration.
// All metrics are enabled, matching existing default behavior.
//
// Returns:
//   - MetricsConfig: standard configuration with all metrics enabled.
func DefaultMetricsConfig() MetricsConfig {
	// Return standard template.
	return StandardMetricsConfig()
}

// StandardMetricsConfig returns the standard template configuration.
// All metrics are enabled, matching existing default behavior.
//
// Returns:
//   - MetricsConfig: standard configuration with all metrics enabled.
func StandardMetricsConfig() MetricsConfig {
	// Enable all categories and sub-features.
	return newMetricsConfig(true, true)
}

// MinimalMetricsConfig returns the minimal template configuration.
// Only essential metrics (CPU, memory, load) are enabled.
// Provides 70-80% allocation reduction compared to standard.
//
// Returns:
//   - MetricsConfig: minimal configuration for low resource consumption.
func MinimalMetricsConfig() MetricsConfig {
	// Enable only essential metrics without expensive sub-features.
	return newMetricsConfig(false, false)
}

// newMetricsConfig creates a MetricsConfig with essential metrics always enabled.
// When allCategories is true, all categories are enabled (standard/full template).
// When false, only CPU, memory, and load are enabled (minimal template).
// The pressure parameter controls PSI collection for CPU/memory.
func newMetricsConfig(allCategories, pressure bool) MetricsConfig {
	return MetricsConfig{
		Enabled:     true,
		CPU:         CPUMetricsConfig{Enabled: true, Pressure: pressure},
		Memory:      MemoryMetricsConfig{Enabled: true, Pressure: pressure},
		Load:        LoadMetricsConfig{Enabled: true},
		Disk:        DiskMetricsConfig{Enabled: allCategories, Partitions: allCategories, Usage: allCategories, IO: allCategories},
		Network:     NetworkMetricsConfig{Enabled: allCategories, Interfaces: allCategories, Stats: allCategories},
		Connections: ConnectionMetricsConfig{Enabled: allCategories, TCPStats: allCategories, TCPConnections: allCategories, UDPSockets: allCategories, UnixSockets: allCategories, ListeningPorts: allCategories},
		Thermal:     ThermalMetricsConfig{Enabled: allCategories},
		Process:     ProcessMetricsConfig{Enabled: allCategories},
		IO:          IOMetricsConfig{Enabled: allCategories, Pressure: allCategories && pressure},
		Quota:       QuotaMetricsConfig{Enabled: allCategories},
		Container:   ContainerMetricsConfig{Enabled: allCategories},
		Runtime:     RuntimeMetricsConfig{Enabled: allCategories},
	}
}

// FullMetricsConfig returns the full template configuration.
// Identical to StandardMetricsConfig, provided for forward compatibility.
//
// Returns:
//   - MetricsConfig: full configuration with all metrics enabled.
func FullMetricsConfig() MetricsConfig {
	// Full template is identical to standard.
	return StandardMetricsConfig()
}
