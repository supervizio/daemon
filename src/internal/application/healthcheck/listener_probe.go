// Package healthcheck provides the application service for health monitoring.
package healthcheck

import (
	"github.com/kodflow/daemon/internal/domain/healthcheck"
	"github.com/kodflow/daemon/internal/domain/listener"
)

// ListenerProbe represents a listener with its probe configuration.
// It combines a network listener with its associated prober for health checks.
type ListenerProbe struct {
	// Listener is the listener to healthcheck.
	Listener *listener.Listener
	// Prober is the prober to use.
	Prober healthcheck.Prober
	// Config is the probe configuration.
	Config healthcheck.Config
	// Target is the probe target.
	Target healthcheck.Target
}

// NewListenerProbe creates a new ListenerProbe with the given listener.
// The prober, config, and target can be set separately after creation.
//
// Params:
//   - l: the listener to associate with this healthcheck.
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
