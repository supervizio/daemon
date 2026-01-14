// Package health provides the application service for health monitoring.
package healthcheck

import (
	"time"

	domain "github.com/kodflow/daemon/internal/domain/health"
)

// ProbeMonitorConfig contains configuration for ProbeMonitor.
// It provides all necessary dependencies for creating a new ProbeMonitor.
type ProbeMonitorConfig struct {
	// Factory creates probers based on type.
	Factory Creator
	// Events channel for health events.
	Events chan<- domain.Event
	// DefaultTimeout for probes when not specified per-listener.
	DefaultTimeout time.Duration
	// DefaultInterval between probes when not specified per-listener.
	DefaultInterval time.Duration
}

// NewProbeMonitorConfig creates a new ProbeMonitorConfig with the given factory.
// The events channel and timeouts can be set separately after creation.
//
// Params:
//   - factory: the prober factory to use for creating probers.
//
// Returns:
//   - ProbeMonitorConfig: a new configuration instance.
func NewProbeMonitorConfig(factory Creator) ProbeMonitorConfig {
	// Return config with factory set and default timeouts to be resolved by ProbeMonitor.
	return ProbeMonitorConfig{
		Factory: factory,
	}
}
