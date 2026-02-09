// Package yaml provides YAML configuration loading infrastructure.
package yaml

import (
	"github.com/kodflow/daemon/internal/domain/config"
)

// MetricsConfigDTO is the YAML representation of metrics configuration.
// It provides granular control over metrics collection to reduce resource consumption.
type MetricsConfigDTO struct {
	Enabled     *bool                       `yaml:"enabled,omitempty"`
	CPU         *CPUMetricsConfigDTO        `yaml:"cpu,omitempty"`
	Memory      *MemoryMetricsConfigDTO     `yaml:"memory,omitempty"`
	Load        *LoadMetricsConfigDTO       `yaml:"load,omitempty"`
	Disk        *DiskMetricsConfigDTO       `yaml:"disk,omitempty"`
	Network     *NetworkMetricsConfigDTO    `yaml:"network,omitempty"`
	Connections *ConnectionMetricsConfigDTO `yaml:"connections,omitempty"`
	Thermal     *ThermalMetricsConfigDTO    `yaml:"thermal,omitempty"`
	Process     *ProcessMetricsConfigDTO    `yaml:"process,omitempty"`
	IO          *IOMetricsConfigDTO         `yaml:"io,omitempty"`
	Quota       *QuotaMetricsConfigDTO      `yaml:"quota,omitempty"`
	Container   *ContainerMetricsConfigDTO  `yaml:"container,omitempty"`
	Runtime     *RuntimeMetricsConfigDTO    `yaml:"runtime,omitempty"`
}

// CPUMetricsConfigDTO is the YAML representation of CPU metrics configuration.
type CPUMetricsConfigDTO struct {
	Enabled  *bool `yaml:"enabled,omitempty"`
	Pressure *bool `yaml:"pressure,omitempty"`
}

// MemoryMetricsConfigDTO is the YAML representation of memory metrics configuration.
type MemoryMetricsConfigDTO struct {
	Enabled  *bool `yaml:"enabled,omitempty"`
	Pressure *bool `yaml:"pressure,omitempty"`
}

// LoadMetricsConfigDTO is the YAML representation of load metrics configuration.
type LoadMetricsConfigDTO struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

// DiskMetricsConfigDTO is the YAML representation of disk metrics configuration.
type DiskMetricsConfigDTO struct {
	Enabled    *bool `yaml:"enabled,omitempty"`
	Partitions *bool `yaml:"partitions,omitempty"`
	Usage      *bool `yaml:"usage,omitempty"`
	IO         *bool `yaml:"io,omitempty"`
}

// NetworkMetricsConfigDTO is the YAML representation of network metrics configuration.
type NetworkMetricsConfigDTO struct {
	Enabled    *bool `yaml:"enabled,omitempty"`
	Interfaces *bool `yaml:"interfaces,omitempty"`
	Stats      *bool `yaml:"stats,omitempty"`
}

// ConnectionMetricsConfigDTO is the YAML representation of connection metrics configuration.
type ConnectionMetricsConfigDTO struct {
	Enabled        *bool `yaml:"enabled,omitempty"`
	TCPStats       *bool `yaml:"tcp_stats,omitempty"`
	TCPConnections *bool `yaml:"tcp_connections,omitempty"`
	UDPSockets     *bool `yaml:"udp_sockets,omitempty"`
	UnixSockets    *bool `yaml:"unix_sockets,omitempty"`
	ListeningPorts *bool `yaml:"listening_ports,omitempty"`
}

// ThermalMetricsConfigDTO is the YAML representation of thermal metrics configuration.
type ThermalMetricsConfigDTO struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

// ProcessMetricsConfigDTO is the YAML representation of process metrics configuration.
type ProcessMetricsConfigDTO struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

// IOMetricsConfigDTO is the YAML representation of I/O metrics configuration.
type IOMetricsConfigDTO struct {
	Enabled  *bool `yaml:"enabled,omitempty"`
	Pressure *bool `yaml:"pressure,omitempty"`
}

// QuotaMetricsConfigDTO is the YAML representation of quota metrics configuration.
type QuotaMetricsConfigDTO struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

// ContainerMetricsConfigDTO is the YAML representation of container metrics configuration.
type ContainerMetricsConfigDTO struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

// RuntimeMetricsConfigDTO is the YAML representation of runtime metrics configuration.
type RuntimeMetricsConfigDTO struct {
	Enabled *bool `yaml:"enabled,omitempty"`
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
	// Start with template as base configuration.
	result := resolveTemplate(template)

	// Apply global enabled override if specified.
	if m.Enabled != nil {
		result.Enabled = *m.Enabled
	}

	// Apply category overrides if specified.
	if m.CPU != nil {
		result.CPU = m.CPU.toDomain(result.CPU)
	}
	if m.Memory != nil {
		result.Memory = m.Memory.toDomain(result.Memory)
	}
	if m.Load != nil {
		result.Load = m.Load.toDomain(result.Load)
	}
	if m.Disk != nil {
		result.Disk = m.Disk.toDomain(result.Disk)
	}
	if m.Network != nil {
		result.Network = m.Network.toDomain(result.Network)
	}
	if m.Connections != nil {
		result.Connections = m.Connections.toDomain(result.Connections)
	}
	if m.Thermal != nil {
		result.Thermal = m.Thermal.toDomain(result.Thermal)
	}
	if m.Process != nil {
		result.Process = m.Process.toDomain(result.Process)
	}
	if m.IO != nil {
		result.IO = m.IO.toDomain(result.IO)
	}
	if m.Quota != nil {
		result.Quota = m.Quota.toDomain(result.Quota)
	}
	if m.Container != nil {
		result.Container = m.Container.toDomain(result.Container)
	}
	if m.Runtime != nil {
		result.Runtime = m.Runtime.toDomain(result.Runtime)
	}

	// Return merged configuration.
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
	// Override enabled if specified.
	if c.Enabled != nil {
		result.Enabled = *c.Enabled
	}
	// Override pressure if specified.
	if c.Pressure != nil {
		result.Pressure = *c.Pressure
	}
	// Return merged configuration.
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
	// Override enabled if specified.
	if m.Enabled != nil {
		result.Enabled = *m.Enabled
	}
	// Override pressure if specified.
	if m.Pressure != nil {
		result.Pressure = *m.Pressure
	}
	// Return merged configuration.
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
	// Override enabled if specified.
	if l.Enabled != nil {
		result.Enabled = *l.Enabled
	}
	// Return merged configuration.
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
	// Override enabled if specified.
	if d.Enabled != nil {
		result.Enabled = *d.Enabled
	}
	// Override partitions if specified.
	if d.Partitions != nil {
		result.Partitions = *d.Partitions
	}
	// Override usage if specified.
	if d.Usage != nil {
		result.Usage = *d.Usage
	}
	// Override IO if specified.
	if d.IO != nil {
		result.IO = *d.IO
	}
	// Return merged configuration.
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
	// Override enabled if specified.
	if n.Enabled != nil {
		result.Enabled = *n.Enabled
	}
	// Override interfaces if specified.
	if n.Interfaces != nil {
		result.Interfaces = *n.Interfaces
	}
	// Override stats if specified.
	if n.Stats != nil {
		result.Stats = *n.Stats
	}
	// Return merged configuration.
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
	// Override enabled if specified.
	if c.Enabled != nil {
		result.Enabled = *c.Enabled
	}
	// Override tcp_stats if specified.
	if c.TCPStats != nil {
		result.TCPStats = *c.TCPStats
	}
	// Override tcp_connections if specified.
	if c.TCPConnections != nil {
		result.TCPConnections = *c.TCPConnections
	}
	// Override udp_sockets if specified.
	if c.UDPSockets != nil {
		result.UDPSockets = *c.UDPSockets
	}
	// Override unix_sockets if specified.
	if c.UnixSockets != nil {
		result.UnixSockets = *c.UnixSockets
	}
	// Override listening_ports if specified.
	if c.ListeningPorts != nil {
		result.ListeningPorts = *c.ListeningPorts
	}
	// Return merged configuration.
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
	// Override enabled if specified.
	if t.Enabled != nil {
		result.Enabled = *t.Enabled
	}
	// Return merged configuration.
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
	// Override enabled if specified.
	if p.Enabled != nil {
		result.Enabled = *p.Enabled
	}
	// Return merged configuration.
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
	// Override enabled if specified.
	if i.Enabled != nil {
		result.Enabled = *i.Enabled
	}
	// Override pressure if specified.
	if i.Pressure != nil {
		result.Pressure = *i.Pressure
	}
	// Return merged configuration.
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
	// Override enabled if specified.
	if q.Enabled != nil {
		result.Enabled = *q.Enabled
	}
	// Return merged configuration.
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
	// Override enabled if specified.
	if c.Enabled != nil {
		result.Enabled = *c.Enabled
	}
	// Return merged configuration.
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
	// Override enabled if specified.
	if r.Enabled != nil {
		result.Enabled = *r.Enabled
	}
	// Return merged configuration.
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
	// Resolve template to configuration.
	switch template {
	case config.MetricsTemplateMinimal:
		return config.MinimalMetricsConfig()
	case config.MetricsTemplateFull:
		return config.FullMetricsConfig()
	case config.MetricsTemplateStandard, config.MetricsTemplateCustom:
		return config.StandardMetricsConfig()
	default:
		// Unknown template defaults to standard.
		return config.StandardMetricsConfig()
	}
}
