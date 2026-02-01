// Package config provides domain value objects for service configuration.
package config

// defaultExcludedSSHPort is the default SSH port excluded from port scan discovery.
const defaultExcludedSSHPort int = 22

// PortScanDiscoveryConfig configures port scan discovery.
// Port scan discovery monitors listening ports on network interfaces by
// reading /proc/net/tcp and /proc/net/tcp6 on Linux systems.
type PortScanDiscoveryConfig struct {
	// Enabled activates port scan discovery.
	Enabled bool

	// Interfaces filters discovery to specific network interfaces.
	// Empty list means all interfaces.
	// Example: ["eth0", "lo"].
	Interfaces []string

	// ExcludePorts are ports to exclude from discovery.
	// Default: [22] (SSH excluded by default).
	ExcludePorts []int

	// IncludePorts are specific ports to include.
	// If set, only these ports are discovered (ExcludePorts is ignored).
	// Empty list means all ports except excluded ones.
	IncludePorts []int
}

// NewPortScanDiscoveryConfig creates a new port scan discovery configuration.
// Default configuration excludes SSH (port 22) and disables discovery.
//
// Returns:
//   - *PortScanDiscoveryConfig: a new configuration with defaults.
func NewPortScanDiscoveryConfig() *PortScanDiscoveryConfig {
	// Return default config with SSH excluded and discovery disabled.
	return &PortScanDiscoveryConfig{
		Enabled:      false,
		Interfaces:   nil,
		ExcludePorts: []int{defaultExcludedSSHPort},
		IncludePorts: nil,
	}
}
