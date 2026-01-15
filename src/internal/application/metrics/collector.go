// Package metrics provides application services for process metrics tracking.
package metrics

import (
	"context"

	domainmetrics "github.com/kodflow/daemon/internal/domain/metrics"
)

// ProcessTracker defines the interface for tracking process-level metrics.
// It aggregates CPU and memory metrics per supervised process and publishes updates.
type ProcessTracker interface {
	// Track starts tracking metrics for a service with the given PID.
	Track(serviceName string, pid int) error
	// Untrack stops tracking metrics for a service.
	Untrack(serviceName string)
	// Get returns the current metrics for a service.
	Get(serviceName string) (domainmetrics.ProcessMetrics, bool)
	// All returns metrics for all tracked services.
	All() []domainmetrics.ProcessMetrics
	// Subscribe returns a channel that receives metrics updates.
	Subscribe() <-chan domainmetrics.ProcessMetrics
	// Unsubscribe removes a subscription channel.
	Unsubscribe(ch <-chan domainmetrics.ProcessMetrics)
}

// Collector abstracts the collection of process metrics.
// It is implemented by infrastructure adapters (e.g., /proc readers).
type Collector interface {
	// CollectCPU collects CPU metrics for a process.
	CollectCPU(ctx context.Context, pid int) (domainmetrics.ProcessCPU, error)
	// CollectMemory collects memory metrics for a process.
	CollectMemory(ctx context.Context, pid int) (domainmetrics.ProcessMemory, error)
}
