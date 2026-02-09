// Package yaml provides YAML configuration loading infrastructure.
// This file groups related metrics DTO types for configuration parsing.
// Grouping is permitted by dto tag convention (exception to KTN-STRUCT-ONEFILE).
package yaml

import (
	"github.com/kodflow/daemon/internal/domain/config"
)

// MetricsConfigDTO is the YAML representation of metrics configuration.
// It provides granular control over metrics collection to reduce resource consumption.
type MetricsConfigDTO struct {
	Enabled     *bool                       `yaml:"enabled,omitempty"`     // global metrics toggle
	CPU         *CPUMetricsConfigDTO        `yaml:"cpu,omitempty"`         // CPU metrics configuration
	Memory      *MemoryMetricsConfigDTO     `yaml:"memory,omitempty"`      // memory metrics configuration
	Load        *LoadMetricsConfigDTO       `yaml:"load,omitempty"`        // load average configuration
	Disk        *DiskMetricsConfigDTO       `yaml:"disk,omitempty"`        // disk metrics configuration
	Network     *NetworkMetricsConfigDTO    `yaml:"network,omitempty"`     // network metrics configuration
	Connections *ConnectionMetricsConfigDTO `yaml:"connections,omitempty"` // connection metrics configuration
	Thermal     *ThermalMetricsConfigDTO    `yaml:"thermal,omitempty"`     // thermal zone configuration
	Process     *ProcessMetricsConfigDTO    `yaml:"process,omitempty"`     // process metrics configuration
	IO          *IOMetricsConfigDTO         `yaml:"io,omitempty"`          // I/O metrics configuration
	Quota       *QuotaMetricsConfigDTO      `yaml:"quota,omitempty"`       // quota metrics configuration
	Container   *ContainerMetricsConfigDTO  `yaml:"container,omitempty"`   // container metrics configuration
	Runtime     *RuntimeMetricsConfigDTO    `yaml:"runtime,omitempty"`     // runtime metrics configuration
}

// CPUMetricsConfigDTO is the YAML representation of CPU metrics configuration.
// It controls CPU usage metrics and pressure stall information collection.
type CPUMetricsConfigDTO struct {
	Enabled  *bool `yaml:"enabled,omitempty"`  // enable CPU metrics collection
	Pressure *bool `yaml:"pressure,omitempty"` // enable PSI (pressure stall information)
}

// MemoryMetricsConfigDTO is the YAML representation of memory metrics configuration.
// It controls memory usage metrics and pressure stall information collection.
type MemoryMetricsConfigDTO struct {
	Enabled  *bool `yaml:"enabled,omitempty"`  // enable memory metrics collection
	Pressure *bool `yaml:"pressure,omitempty"` // enable PSI (pressure stall information)
}

// LoadMetricsConfigDTO is the YAML representation of load metrics configuration.
// It controls system load average collection (1min, 5min, 15min).
type LoadMetricsConfigDTO struct {
	Enabled *bool `yaml:"enabled,omitempty"` // enable load average metrics
}

// DiskMetricsConfigDTO is the YAML representation of disk metrics configuration.
// It controls disk partition, usage, and I/O statistics collection.
type DiskMetricsConfigDTO struct {
	Enabled    *bool `yaml:"enabled,omitempty"`    // enable disk metrics collection
	Partitions *bool `yaml:"partitions,omitempty"` // include partition information
	Usage      *bool `yaml:"usage,omitempty"`      // include filesystem usage
	IO         *bool `yaml:"io,omitempty"`         // include I/O statistics
}

// NetworkMetricsConfigDTO is the YAML representation of network metrics configuration.
// It controls network interface and statistics collection.
type NetworkMetricsConfigDTO struct {
	Enabled    *bool `yaml:"enabled,omitempty"`    // enable network metrics collection
	Interfaces *bool `yaml:"interfaces,omitempty"` // include interface details
	Stats      *bool `yaml:"stats,omitempty"`      // include transfer statistics
}

// ConnectionMetricsConfigDTO is the YAML representation of connection metrics configuration.
// It controls TCP, UDP, and Unix socket connection enumeration.
type ConnectionMetricsConfigDTO struct {
	Enabled        *bool `yaml:"enabled,omitempty"`         // enable connection metrics collection
	TCPStats       *bool `yaml:"tcp_stats,omitempty"`       // include TCP protocol statistics
	TCPConnections *bool `yaml:"tcp_connections,omitempty"` // enumerate TCP connections
	UDPSockets     *bool `yaml:"udp_sockets,omitempty"`     // enumerate UDP sockets
	UnixSockets    *bool `yaml:"unix_sockets,omitempty"`    // enumerate Unix domain sockets
	ListeningPorts *bool `yaml:"listening_ports,omitempty"` // include listening port information
}

// ThermalMetricsConfigDTO is the YAML representation of thermal metrics configuration.
// It controls thermal zone temperature collection.
type ThermalMetricsConfigDTO struct {
	Enabled *bool `yaml:"enabled,omitempty"` // enable thermal zone metrics
}

// ProcessMetricsConfigDTO is the YAML representation of process metrics configuration.
// It controls per-process metrics collection (CPU, memory, I/O per process).
type ProcessMetricsConfigDTO struct {
	Enabled *bool `yaml:"enabled,omitempty"` // enable per-process metrics
}

// IOMetricsConfigDTO is the YAML representation of I/O metrics configuration.
// It controls I/O statistics and pressure stall information collection.
type IOMetricsConfigDTO struct {
	Enabled  *bool `yaml:"enabled,omitempty"`  // enable I/O metrics collection
	Pressure *bool `yaml:"pressure,omitempty"` // enable PSI (pressure stall information)
}

// QuotaMetricsConfigDTO is the YAML representation of quota metrics configuration.
// It controls disk quota and limit collection.
type QuotaMetricsConfigDTO struct {
	Enabled *bool `yaml:"enabled,omitempty"` // enable quota metrics collection
}

// ContainerMetricsConfigDTO is the YAML representation of container metrics configuration.
// It controls container-specific metrics collection (cgroups, namespaces).
type ContainerMetricsConfigDTO struct {
	Enabled *bool `yaml:"enabled,omitempty"` // enable container metrics collection
}

// RuntimeMetricsConfigDTO is the YAML representation of runtime metrics configuration.
// It controls runtime environment detection (Docker, Kubernetes, etc.).
type RuntimeMetricsConfigDTO struct {
	Enabled *bool `yaml:"enabled,omitempty"` // enable runtime detection metrics
}

// ToDomain converts MetricsConfigDTO to domain MetricsConfig.
// It applies the template as a base, then overlays explicit configuration.
//
// Params:
//   - template: the metrics template to use as base configuration
//
// Returns:
//   - config.MetricsConfig: the converted domain metrics configuration
func (m *MetricsConfigDTO) ToDomain(template config.MetricsTemplate) config.MetricsConfig {
	// start with template as base configuration.
	result := resolveTemplate(template)

	// apply global enabled override if specified.
	if m.Enabled != nil {
		result.Enabled = *m.Enabled
	}

	// apply CPU category overrides if specified.
	if m.CPU != nil {
		result.CPU = m.CPU.toDomain(result.CPU)
	}
	// apply memory category overrides if specified.
	if m.Memory != nil {
		result.Memory = m.Memory.toDomain(result.Memory)
	}
	// apply load category overrides if specified.
	if m.Load != nil {
		result.Load = m.Load.toDomain(result.Load)
	}
	// apply disk category overrides if specified.
	if m.Disk != nil {
		result.Disk = m.Disk.toDomain(result.Disk)
	}
	// apply network category overrides if specified.
	if m.Network != nil {
		result.Network = m.Network.toDomain(result.Network)
	}
	// apply connections category overrides if specified.
	if m.Connections != nil {
		result.Connections = m.Connections.toDomain(result.Connections)
	}
	// apply thermal category overrides if specified.
	if m.Thermal != nil {
		result.Thermal = m.Thermal.toDomain(result.Thermal)
	}
	// apply process category overrides if specified.
	if m.Process != nil {
		result.Process = m.Process.toDomain(result.Process)
	}
	// apply I/O category overrides if specified.
	if m.IO != nil {
		result.IO = m.IO.toDomain(result.IO)
	}
	// apply quota category overrides if specified.
	if m.Quota != nil {
		result.Quota = m.Quota.toDomain(result.Quota)
	}
	// apply container category overrides if specified.
	if m.Container != nil {
		result.Container = m.Container.toDomain(result.Container)
	}
	// apply runtime category overrides if specified.
	if m.Runtime != nil {
		result.Runtime = m.Runtime.toDomain(result.Runtime)
	}

	// return merged configuration.
	return result
}

// toDomain converts CPUMetricsConfigDTO to domain CPUMetricsConfig.
// It overlays DTO values onto the base configuration.
//
// Params:
//   - base: the base configuration from template
//
// Returns:
//   - config.CPUMetricsConfig: the merged CPU metrics configuration
func (c *CPUMetricsConfigDTO) toDomain(base config.CPUMetricsConfig) config.CPUMetricsConfig {
	result := base
	// override enabled if specified.
	if c.Enabled != nil {
		result.Enabled = *c.Enabled
	}
	// override pressure if specified.
	if c.Pressure != nil {
		result.Pressure = *c.Pressure
	}
	// return merged configuration.
	return result
}

// toDomain converts MemoryMetricsConfigDTO to domain MemoryMetricsConfig.
// It overlays DTO values onto the base configuration.
//
// Params:
//   - base: the base configuration from template
//
// Returns:
//   - config.MemoryMetricsConfig: the merged memory metrics configuration
func (m *MemoryMetricsConfigDTO) toDomain(base config.MemoryMetricsConfig) config.MemoryMetricsConfig {
	result := base
	// override enabled if specified.
	if m.Enabled != nil {
		result.Enabled = *m.Enabled
	}
	// override pressure if specified.
	if m.Pressure != nil {
		result.Pressure = *m.Pressure
	}
	// return merged configuration.
	return result
}

// toDomain converts LoadMetricsConfigDTO to domain LoadMetricsConfig.
// It overlays DTO values onto the base configuration.
//
// Params:
//   - base: the base configuration from template
//
// Returns:
//   - config.LoadMetricsConfig: the merged load metrics configuration
func (l *LoadMetricsConfigDTO) toDomain(base config.LoadMetricsConfig) config.LoadMetricsConfig {
	result := base
	// override enabled if specified.
	if l.Enabled != nil {
		result.Enabled = *l.Enabled
	}
	// return merged configuration.
	return result
}

// toDomain converts DiskMetricsConfigDTO to domain DiskMetricsConfig.
// It overlays DTO values onto the base configuration.
//
// Params:
//   - base: the base configuration from template
//
// Returns:
//   - config.DiskMetricsConfig: the merged disk metrics configuration
func (d *DiskMetricsConfigDTO) toDomain(base config.DiskMetricsConfig) config.DiskMetricsConfig {
	result := base
	// override enabled if specified.
	if d.Enabled != nil {
		result.Enabled = *d.Enabled
	}
	// override partitions if specified.
	if d.Partitions != nil {
		result.Partitions = *d.Partitions
	}
	// override usage if specified.
	if d.Usage != nil {
		result.Usage = *d.Usage
	}
	// override IO if specified.
	if d.IO != nil {
		result.IO = *d.IO
	}
	// return merged configuration.
	return result
}

// toDomain converts NetworkMetricsConfigDTO to domain NetworkMetricsConfig.
// It overlays DTO values onto the base configuration.
//
// Params:
//   - base: the base configuration from template
//
// Returns:
//   - config.NetworkMetricsConfig: the merged network metrics configuration
func (n *NetworkMetricsConfigDTO) toDomain(base config.NetworkMetricsConfig) config.NetworkMetricsConfig {
	result := base
	// override enabled if specified.
	if n.Enabled != nil {
		result.Enabled = *n.Enabled
	}
	// override interfaces if specified.
	if n.Interfaces != nil {
		result.Interfaces = *n.Interfaces
	}
	// override stats if specified.
	if n.Stats != nil {
		result.Stats = *n.Stats
	}
	// return merged configuration.
	return result
}

// toDomain converts ConnectionMetricsConfigDTO to domain ConnectionMetricsConfig.
// It overlays DTO values onto the base configuration.
//
// Params:
//   - base: the base configuration from template
//
// Returns:
//   - config.ConnectionMetricsConfig: the merged connection metrics configuration
func (c *ConnectionMetricsConfigDTO) toDomain(base config.ConnectionMetricsConfig) config.ConnectionMetricsConfig {
	result := base
	// override enabled if specified.
	if c.Enabled != nil {
		result.Enabled = *c.Enabled
	}
	// override tcp_stats if specified.
	if c.TCPStats != nil {
		result.TCPStats = *c.TCPStats
	}
	// override tcp_connections if specified.
	if c.TCPConnections != nil {
		result.TCPConnections = *c.TCPConnections
	}
	// override udp_sockets if specified.
	if c.UDPSockets != nil {
		result.UDPSockets = *c.UDPSockets
	}
	// override unix_sockets if specified.
	if c.UnixSockets != nil {
		result.UnixSockets = *c.UnixSockets
	}
	// override listening_ports if specified.
	if c.ListeningPorts != nil {
		result.ListeningPorts = *c.ListeningPorts
	}
	// return merged configuration.
	return result
}

// toDomain converts ThermalMetricsConfigDTO to domain ThermalMetricsConfig.
// It overlays DTO values onto the base configuration.
//
// Params:
//   - base: the base configuration from template
//
// Returns:
//   - config.ThermalMetricsConfig: the merged thermal metrics configuration
func (t *ThermalMetricsConfigDTO) toDomain(base config.ThermalMetricsConfig) config.ThermalMetricsConfig {
	result := base
	// override enabled if specified.
	if t.Enabled != nil {
		result.Enabled = *t.Enabled
	}
	// return merged configuration.
	return result
}

// toDomain converts ProcessMetricsConfigDTO to domain ProcessMetricsConfig.
// It overlays DTO values onto the base configuration.
//
// Params:
//   - base: the base configuration from template
//
// Returns:
//   - config.ProcessMetricsConfig: the merged process metrics configuration
func (p *ProcessMetricsConfigDTO) toDomain(base config.ProcessMetricsConfig) config.ProcessMetricsConfig {
	result := base
	// override enabled if specified.
	if p.Enabled != nil {
		result.Enabled = *p.Enabled
	}
	// return merged configuration.
	return result
}

// toDomain converts IOMetricsConfigDTO to domain IOMetricsConfig.
// It overlays DTO values onto the base configuration.
//
// Params:
//   - base: the base configuration from template
//
// Returns:
//   - config.IOMetricsConfig: the merged I/O metrics configuration
func (i *IOMetricsConfigDTO) toDomain(base config.IOMetricsConfig) config.IOMetricsConfig {
	result := base
	// override enabled if specified.
	if i.Enabled != nil {
		result.Enabled = *i.Enabled
	}
	// override pressure if specified.
	if i.Pressure != nil {
		result.Pressure = *i.Pressure
	}
	// return merged configuration.
	return result
}

// toDomain converts QuotaMetricsConfigDTO to domain QuotaMetricsConfig.
// It overlays DTO values onto the base configuration.
//
// Params:
//   - base: the base configuration from template
//
// Returns:
//   - config.QuotaMetricsConfig: the merged quota metrics configuration
func (q *QuotaMetricsConfigDTO) toDomain(base config.QuotaMetricsConfig) config.QuotaMetricsConfig {
	result := base
	// override enabled if specified.
	if q.Enabled != nil {
		result.Enabled = *q.Enabled
	}
	// return merged configuration.
	return result
}

// toDomain converts ContainerMetricsConfigDTO to domain ContainerMetricsConfig.
// It overlays DTO values onto the base configuration.
//
// Params:
//   - base: the base configuration from template
//
// Returns:
//   - config.ContainerMetricsConfig: the merged container metrics configuration
func (c *ContainerMetricsConfigDTO) toDomain(base config.ContainerMetricsConfig) config.ContainerMetricsConfig {
	result := base
	// override enabled if specified.
	if c.Enabled != nil {
		result.Enabled = *c.Enabled
	}
	// return merged configuration.
	return result
}

// toDomain converts RuntimeMetricsConfigDTO to domain RuntimeMetricsConfig.
// It overlays DTO values onto the base configuration.
//
// Params:
//   - base: the base configuration from template
//
// Returns:
//   - config.RuntimeMetricsConfig: the merged runtime metrics configuration
func (r *RuntimeMetricsConfigDTO) toDomain(base config.RuntimeMetricsConfig) config.RuntimeMetricsConfig {
	result := base
	// override enabled if specified.
	if r.Enabled != nil {
		result.Enabled = *r.Enabled
	}
	// return merged configuration.
	return result
}

// resolveTemplate resolves a template name to a MetricsConfig.
// Unknown templates default to standard.
//
// Params:
//   - template: the template name to resolve
//
// Returns:
//   - config.MetricsConfig: the resolved metrics configuration
func resolveTemplate(template config.MetricsTemplate) config.MetricsConfig {
	// resolve template to configuration.
	switch template {
	// return minimal configuration for minimal template.
	case config.MetricsTemplateMinimal:
		// use minimal configuration factory.
		return config.MinimalMetricsConfig()
	// return full configuration for full template.
	case config.MetricsTemplateFull:
		// use full configuration factory.
		return config.FullMetricsConfig()
	// return standard configuration for standard/custom templates.
	case config.MetricsTemplateStandard, config.MetricsTemplateCustom:
		// use standard configuration factory.
		return config.StandardMetricsConfig()
	// default to standard configuration for unknown templates.
	default:
		// use standard configuration as safe fallback.
		return config.StandardMetricsConfig()
	}
}
