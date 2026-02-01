// Package monitoring provides the application service for external target monitoring.
package monitoring

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/target"
)

// Default values for monitor configuration.
const (
	// DefaultInterval is the default probe interval.
	DefaultInterval time.Duration = 30 * time.Second

	// DefaultTimeout is the default probe timeout.
	DefaultTimeout time.Duration = 5 * time.Second

	// DefaultSuccessThreshold is the default consecutive successes for healthy.
	DefaultSuccessThreshold int = 1

	// DefaultFailureThreshold is the default consecutive failures for unhealthy.
	DefaultFailureThreshold int = 3

	// DefaultDiscoveryInterval is the default interval for re-discovery.
	DefaultDiscoveryInterval time.Duration = 60 * time.Second
)

// Config contains configuration for the ExternalMonitor.
// It includes defaults, discovery settings, callbacks, and the prober factory.
type Config struct {
	// Defaults are the default timing and threshold values.
	Defaults DefaultsConfig

	// Discovery configures automatic target discovery.
	Discovery DiscoveryModeConfig

	// Events is the channel for sending target events.
	// Optional - if nil, events are not sent.
	Events chan<- target.Event

	// Factory creates probers.
	// Required for probe execution.
	Factory Creator

	// OnHealthChange is called when target health changes.
	OnHealthChange HealthCallback

	// OnUnhealthy is called when a target becomes unhealthy.
	OnUnhealthy UnhealthyCallback

	// OnHealthy is called when a target becomes healthy.
	OnHealthy HealthyCallback
}

// NewConfig creates a new configuration with default values.
//
// Returns:
//   - Config: a new configuration with defaults applied.
func NewConfig() Config {
	// construct config with default values
	return Config{
		Defaults:  NewDefaultsConfig(),
		Discovery: NewDiscoveryModeConfig(),
	}
}

// WithFactory sets the prober factory.
//
// Params:
//   - factory: the prober factory.
//
// Returns:
//   - Config: the config for method chaining.
func (c Config) WithFactory(factory Creator) Config {
	c.Factory = factory
	// return updated config
	return c
}

// WithEvents sets the events channel.
//
// Params:
//   - events: the events channel.
//
// Returns:
//   - Config: the config for method chaining.
func (c Config) WithEvents(events chan<- target.Event) Config {
	c.Events = events
	// return updated config
	return c
}

// WithCallbacks sets the health callbacks.
//
// Params:
//   - onChange: called when health state changes.
//   - onUnhealthy: called when target becomes unhealthy.
//   - onHealthy: called when target becomes healthy.
//
// Returns:
//   - Config: the config for method chaining.
func (c Config) WithCallbacks(onChange HealthCallback, onUnhealthy UnhealthyCallback, onHealthy HealthyCallback) Config {
	c.OnHealthChange = onChange
	c.OnUnhealthy = onUnhealthy
	c.OnHealthy = onHealthy
	// return updated config
	return c
}

// WithDiscoverers adds discoverers for auto-discovery.
//
// Params:
//   - discoverers: the discoverer adapters.
//
// Returns:
//   - Config: the config for method chaining.
func (c Config) WithDiscoverers(discoverers ...target.Discoverer) Config {
	c.Discovery.Enabled = len(discoverers) > 0
	c.Discovery.Discoverers = discoverers
	// return updated config
	return c
}

// WithWatchers adds watchers for real-time updates.
//
// Params:
//   - watchers: the watcher adapters.
//
// Returns:
//   - Config: the config for method chaining.
func (c Config) WithWatchers(watchers ...target.Watcher) Config {
	c.Discovery.Watchers = watchers
	// return updated config
	return c
}

// WithDiscoveryInterval sets the discovery refresh interval.
//
// Params:
//   - interval: the discovery interval.
//
// Returns:
//   - Config: the config for method chaining.
func (c Config) WithDiscoveryInterval(interval time.Duration) Config {
	c.Discovery.Interval = interval
	// return updated config
	return c
}

// GetInterval returns the effective probe interval.
//
// Params:
//   - override: optional override interval.
//
// Returns:
//   - time.Duration: the effective interval.
func (c Config) GetInterval(override time.Duration) time.Duration {
	// check if override is set
	if override > 0 {
		// return override value
		return override
	}
	// check if default is set
	if c.Defaults.Interval > 0 {
		// return default value
		return c.Defaults.Interval
	}
	// return package default
	return DefaultInterval
}

// GetTimeout returns the effective probe timeout.
//
// Params:
//   - override: optional override timeout.
//
// Returns:
//   - time.Duration: the effective timeout.
func (c Config) GetTimeout(override time.Duration) time.Duration {
	// check if override is set
	if override > 0 {
		// return override value
		return override
	}
	// check if default is set
	if c.Defaults.Timeout > 0 {
		// return default value
		return c.Defaults.Timeout
	}
	// return package default
	return DefaultTimeout
}

// GetSuccessThreshold returns the effective success threshold.
//
// Params:
//   - override: optional override threshold.
//
// Returns:
//   - int: the effective threshold.
func (c Config) GetSuccessThreshold(override int) int {
	// check if override is set
	if override > 0 {
		// return override value
		return override
	}
	// check if default is set
	if c.Defaults.SuccessThreshold > 0 {
		// return default value
		return c.Defaults.SuccessThreshold
	}
	// return package default
	return DefaultSuccessThreshold
}

// GetFailureThreshold returns the effective failure threshold.
//
// Params:
//   - override: optional override threshold.
//
// Returns:
//   - int: the effective threshold.
func (c Config) GetFailureThreshold(override int) int {
	// check if override is set
	if override > 0 {
		// return override value
		return override
	}
	// check if default is set
	if c.Defaults.FailureThreshold > 0 {
		// return default value
		return c.Defaults.FailureThreshold
	}
	// return package default
	return DefaultFailureThreshold
}
