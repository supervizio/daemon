// Package config provides domain value objects for service configuration.
package config

// BSDRCDiscoveryConfig configures BSD rc.d service discovery.
// BSD rc.d discovery monitors services managed by BSD init system.
type BSDRCDiscoveryConfig struct {
	// Enabled activates BSD rc.d discovery.
	Enabled bool

	// Patterns are service name patterns to monitor.
	Patterns []string
}
