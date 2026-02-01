// Package config provides domain value objects for service configuration.
package config

// OpenRCDiscoveryConfig configures OpenRC service discovery.
// OpenRC discovery monitors services managed by OpenRC (Alpine Linux, Gentoo).
type OpenRCDiscoveryConfig struct {
	// Enabled activates OpenRC discovery.
	Enabled bool

	// Patterns are service name patterns to monitor.
	Patterns []string
}
