// Package health provides health monitoring for services.
package health

// ProbeBinding associates a listener with its probe configuration.
// This is the application-level wiring between listeners and health checks.
type ProbeBinding struct {
	// ListenerName is the name of the listener to probe.
	ListenerName string
	// Type is the probe type.
	Type ProbeType
	// Target is the probe target configuration.
	Target ProbeTarget
	// Config is the probe timing configuration.
	Config ProbeConfig
}

// NewProbeBinding creates a new probe binding.
//
// Params:
//   - listenerName: the name of the listener to probe.
//   - probeType: the type of probe to execute (TCP, HTTP, etc.).
//   - target: the probe target configuration.
//
// Returns:
//   - *ProbeBinding: new binding with default config.
func NewProbeBinding(listenerName string, probeType ProbeType, target ProbeTarget) *ProbeBinding {
	// Return new binding with provided parameters and default configuration.
	return &ProbeBinding{
		ListenerName: listenerName,
		Type:         probeType,
		Target:       target,
		Config:       DefaultProbeConfig(),
	}
}

// WithConfig sets a custom probe configuration.
//
// Params:
//   - config: the custom probe configuration to apply.
//
// Returns:
//   - *ProbeBinding: self for method chaining.
func (b *ProbeBinding) WithConfig(config ProbeConfig) *ProbeBinding {
	// Update config and return self for fluent API.
	b.Config = config
	// Return self to enable method chaining.
	return b
}
