// Package metrics provides application services for process metrics tracking.
package probe

import (
	"context"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// ProcessTracker defines the interface for tracking process-level probe.
// It aggregates CPU and memory metrics per supervised process and publishes updates.
type ProcessTracker interface {
	// Track starts tracking metrics for a service with the given PID.
	Track(ctx context.Context, serviceName string, pid int) error
	// Untrack stops tracking metrics for a service.
	Untrack(serviceName string)
	// Get returns the current metrics for a service.
	Get(serviceName string) (probe.ProcessMetrics, bool)
	// GetAll returns metrics for all tracked services.
	GetAll() []probe.ProcessMetrics
	// Subscribe returns a channel that receives metrics updates.
	Subscribe() <-chan probe.ProcessMetrics
	// Unsubscribe removes a subscription channel.
	Unsubscribe(ch <-chan probe.ProcessMetrics)
}

// Collector abstracts the collection of process probe.
// It is implemented by infrastructure adapters (e.g., /proc readers).
type Collector interface {
	// CollectCPU collects CPU metrics for a process.
	CollectCPU(ctx context.Context, pid int) (probe.ProcessCPU, error)
	// CollectMemory collects memory metrics for a process.
	CollectMemory(ctx context.Context, pid int) (probe.ProcessMemory, error)
}
