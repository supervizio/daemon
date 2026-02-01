// Package monitoring provides the application service for external target monitoring.
package monitoring

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
)

// Creator creates probers based on type.
// This is the same interface as health.Creator, re-exported for clarity.
// Infrastructure adapters implement this for prober creation.
type Creator interface {
	// Create creates a prober of the specified type.
	//
	// Params:
	//   - proberType: the type of prober to create (tcp, http, icmp, etc.).
	//   - timeout: the timeout for the prober.
	//
	// Returns:
	//   - health.Prober: the created prober.
	//   - error: if creation fails.
	Create(proberType string, timeout time.Duration) (health.Prober, error)
}

// HealthCallback is called when a target's health state changes.
type HealthCallback func(targetID string, previousState, newState string)

// UnhealthyCallback is called when a target becomes unhealthy.
type UnhealthyCallback func(targetID string, reason string)

// HealthyCallback is called when a target becomes healthy.
type HealthyCallback func(targetID string)
