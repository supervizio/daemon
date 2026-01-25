// Package health provides the application service for health monitoring.
package health

import (
	"time"

	domain "github.com/kodflow/daemon/internal/domain/health"
)

// HealthStateLogger is called when a health state transition occurs.
// This enables logging health state transitions for observability.
type HealthStateLogger func(listenerName string, prevState, newState domain.SubjectState, result domain.CheckResult)

// UnhealthyCallback is called when a service becomes unhealthy.
// This enables the supervisor to trigger restart on health failure.
type UnhealthyCallback func(listenerName string, reason string)

// HealthyCallback is called when a service becomes healthy.
// This enables the supervisor to emit healthy events for observability.
type HealthyCallback func(listenerName string)

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
	// OnStateChange is called when a health state transition occurs (optional).
	// This callback is invoked before the event is sent to the Events channel.
	OnStateChange HealthStateLogger
	// OnUnhealthy is called when a service becomes unhealthy (optional).
	// This callback enables the supervisor to trigger restart on health failure,
	// following the Kubernetes liveness probe pattern.
	OnUnhealthy UnhealthyCallback
	// OnHealthy is called when a service becomes healthy (optional).
	// This callback enables the supervisor to emit healthy events for observability.
	OnHealthy HealthyCallback
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
