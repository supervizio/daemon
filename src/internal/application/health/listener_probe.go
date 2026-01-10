// Package health provides the application service for health monitoring.
package health

import (
	"github.com/kodflow/daemon/internal/domain/listener"
	"github.com/kodflow/daemon/internal/domain/probe"
)

// ListenerProbe represents a listener with its probe configuration.
// It combines a network listener with its associated prober for health checks.
type ListenerProbe struct {
	// Listener is the listener to probe.
	Listener *listener.Listener
	// Prober is the prober to use.
	Prober probe.Prober
	// Config is the probe configuration.
	Config probe.Config
	// Target is the probe target.
	Target probe.Target
}

// NewListenerProbe creates a new ListenerProbe with the given listener.
// The prober, config, and target can be set separately after creation.
//
// Params:
//   - l: the listener to associate with this probe.
//
// Returns:
//   - *ListenerProbe: a new ListenerProbe instance.
func NewListenerProbe(l *listener.Listener) *ListenerProbe {
	// Return a new ListenerProbe with the listener set.
	return &ListenerProbe{
		Listener: l,
	}
}

// HasProber returns true if this listener has a prober configured.
//
// Returns:
//   - bool: true if prober is configured.
func (lp *ListenerProbe) HasProber() bool {
	// Check if a prober has been assigned to enable health probing.
	return lp.Prober != nil
}
