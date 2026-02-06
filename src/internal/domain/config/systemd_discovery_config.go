// Package config provides domain value objects for service configuration.
package config

// SystemdDiscoveryConfig configures systemd service discovery.
// Systemd discovery monitors services managed by systemd on Linux systems.
type SystemdDiscoveryConfig struct {
	// Enabled activates systemd discovery.
	Enabled bool

	// Patterns are unit name patterns to monitor.
	// Supports glob patterns (e.g., "nginx.service", "*.target").
	Patterns []string
}
