// Package health provides the application service for health monitoring.
package health

import (
	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/listener"
)

// ListenerProbe represents a listener with its probe configuration.
// It combines a network listener with its associated prober for health checks.
// The probe configuration comes from ProbeBinding (application layer).
type ListenerProbe struct {
	// Listener is the listener to probe.
	Listener *listener.Listener
	// Prober is the prober to use (created from binding).
	Prober domain.Prober
	// Binding is the probe binding configuration (application layer).
	Binding *ProbeBinding
}

// NewListenerProbe creates a new ListenerProbe with the given listener.
// The prober and binding can be set separately after creation.
//
// Params:
//   - l: the listener to associate with this probe.
//
// Returns:
//   - *ListenerProbe: a new ListenerProbe instance.
func NewListenerProbe(l *listener.Listener) *ListenerProbe {
	return &ListenerProbe{
		Listener: l,
	}
}

// NewListenerProbeWithBinding creates a new ListenerProbe with listener and binding.
//
// Params:
//   - l: the listener to associate with this probe.
//   - binding: the probe binding configuration.
//
// Returns:
//   - *ListenerProbe: a new ListenerProbe instance.
func NewListenerProbeWithBinding(l *listener.Listener, binding *ProbeBinding) *ListenerProbe {
	return &ListenerProbe{
		Listener: l,
		Binding:  binding,
	}
}

// HasProber returns true if this listener has a prober configured.
//
// Returns:
//   - bool: true if prober is configured.
func (lp *ListenerProbe) HasProber() bool {
	return lp.Prober != nil
}

// HasBinding returns true if this listener has a probe binding configured.
//
// Returns:
//   - bool: true if binding is configured.
func (lp *ListenerProbe) HasBinding() bool {
	return lp.Binding != nil
}

// ProbeAddress returns the address to probe.
// Uses binding target address if set, otherwise constructs from listener.
//
// Returns:
//   - string: the address to probe.
func (lp *ListenerProbe) ProbeAddress() string {
	// Use binding address if configured, otherwise fall back to listener address.
	if lp.Binding != nil && lp.Binding.Target.Address != "" {
		return lp.Binding.Target.Address
	}
	return lp.Listener.Address
}

// ProbeTarget returns the health target for this listener probe.
// Converts from application ProbeTarget to domain health.Target.
//
// Returns:
//   - domain.Target: the domain probe target.
func (lp *ListenerProbe) ProbeTarget() domain.Target {
	// Return minimal target when no binding configured.
	if lp.Binding == nil {
		return domain.Target{
			Address: lp.ProbeAddress(),
		}
	}
	// Return full target with all binding configuration fields.
	return domain.Target{
		Address:    lp.ProbeAddress(),
		Path:       lp.Binding.Target.Path,
		Service:    lp.Binding.Target.Service,
		Method:     lp.Binding.Target.Method,
		StatusCode: lp.Binding.Target.StatusCode,
		Command:    lp.Binding.Target.Command,
		Args:       lp.Binding.Target.Args,
	}
}

// ProbeConfig returns the health config for this listener probe.
// Converts from application ProbeConfig to domain health.CheckConfig.
//
// Returns:
//   - domain.CheckConfig: the domain probe config.
func (lp *ListenerProbe) ProbeConfig() domain.CheckConfig {
	// Return domain defaults when no binding configured.
	if lp.Binding == nil {
		return domain.CheckConfig{
			Interval:         domain.DefaultInterval,
			Timeout:          domain.DefaultTimeout,
			SuccessThreshold: domain.DefaultSuccessThreshold,
			FailureThreshold: domain.DefaultFailureThreshold,
		}
	}
	// Return config from binding.
	return domain.CheckConfig{
		Interval:         lp.Binding.Config.Interval,
		Timeout:          lp.Binding.Config.Timeout,
		SuccessThreshold: lp.Binding.Config.SuccessThreshold,
		FailureThreshold: lp.Binding.Config.FailureThreshold,
	}
}
