// Package config provides domain value objects for service configuration.
package config

import (
	"github.com/kodflow/daemon/internal/domain/shared"
)

// Default monitoring values.
const (
	defaultMonitoringInterval int = 30 // seconds
	defaultMonitoringTimeout  int = 5  // seconds
	defaultMonitoringSuccess  int = 1
	defaultMonitoringFailure  int = 3
)

// MonitoringConfig defines configuration for external target monitoring.
// External targets are services, containers, or hosts that supervizio
// monitors but does not manage.
type MonitoringConfig struct {
	// Defaults specifies default values for all targets.
	Defaults MonitoringDefaults

	// Discovery configures auto-discovery for different platforms.
	Discovery DiscoveryConfig

	// Metrics configures granular metrics collection.
	Metrics MetricsConfig

	// Targets is the list of statically defined targets.
	Targets []TargetConfig
}

// NewMonitoringConfig creates a new monitoring configuration with defaults.
//
// Returns:
//   - MonitoringConfig: configuration with default values.
func NewMonitoringConfig() MonitoringConfig {
	// create monitoring config with defaults and no targets
	return MonitoringConfig{
		Defaults: DefaultMonitoringDefaults(),
		Metrics:  DefaultMetricsConfig(),
	}
}

// HasDiscoveryEnabled checks if any discovery is enabled.
//
// Returns:
//   - bool: true if any discovery is enabled.
func (c *MonitoringConfig) HasDiscoveryEnabled() bool {
	disc := &c.Discovery
	// check all discovery types: init systems, containers, orchestrators
	return disc.hasInitSystemDiscovery() ||
		disc.hasContainerDiscovery() ||
		disc.hasOrchestratorDiscovery()
}

// HasStaticTargets checks if any static targets are defined.
//
// Returns:
//   - bool: true if there are static targets.
func (c *MonitoringConfig) HasStaticTargets() bool {
	// check if targets slice is not empty
	return len(c.Targets) > 0
}

// IsEmpty checks if monitoring configuration is empty.
//
// Returns:
//   - bool: true if no targets or discovery is configured.
func (c *MonitoringConfig) IsEmpty() bool {
	// empty when no discovery or static targets
	return !c.HasDiscoveryEnabled() && !c.HasStaticTargets()
}

// DefaultMonitoringDefaults returns default values for monitoring.
//
// Returns:
//   - MonitoringDefaults: default monitoring timing and thresholds.
func DefaultMonitoringDefaults() MonitoringDefaults {
	// create defaults with standard intervals and thresholds
	return MonitoringDefaults{
		Interval:         shared.Seconds(defaultMonitoringInterval),
		Timeout:          shared.Seconds(defaultMonitoringTimeout),
		SuccessThreshold: defaultMonitoringSuccess,
		FailureThreshold: defaultMonitoringFailure,
	}
}
