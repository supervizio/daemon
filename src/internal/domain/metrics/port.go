// Package metrics provides domain types for system and process metrics collection.
package metrics

import "context"

// CPUCollector defines the port interface for CPU metrics collection.
type CPUCollector interface {
	// CollectSystem collects system-wide CPU metrics.
	CollectSystem(ctx context.Context) (SystemCPU, error)
	// CollectProcess collects CPU metrics for a specific process.
	CollectProcess(ctx context.Context, pid int) (ProcessCPU, error)
	// CollectAllProcesses collects CPU metrics for all visible processes.
	CollectAllProcesses(ctx context.Context) ([]ProcessCPU, error)
}

// MemoryCollector defines the port interface for memory metrics collection.
type MemoryCollector interface {
	// CollectSystem collects system-wide memory metrics.
	CollectSystem(ctx context.Context) (SystemMemory, error)
	// CollectProcess collects memory metrics for a specific process.
	CollectProcess(ctx context.Context, pid int) (ProcessMemory, error)
	// CollectAllProcesses collects memory metrics for all visible processes.
	CollectAllProcesses(ctx context.Context) ([]ProcessMemory, error)
}
