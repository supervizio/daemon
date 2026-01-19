// Package health provides the application service for health monitoring.
package health

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
)

// Creator creates probers based on type.
// It is the port that infrastructure adapters implement for prober creation.
type Creator interface {
	// Create creates a prober of the specified type.
	//
	// Params:
	//   - proberType: the type of prober to create.
	//   - timeout: the timeout for the prober.
	//
	// Returns:
	//   - health.Prober: the created prober.
	//   - error: if creation fails.
	Create(proberType string, timeout time.Duration) (health.Prober, error)
}
