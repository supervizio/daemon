//go:build linux

// Package linux provides Linux-specific metric collectors using /proc filesystem.
// It implements unified process metrics collection by combining CPU and memory collectors.
package linux

import (
	"context"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// ProcessCollector implements MetricsCollector by combining CPU and memory collectors.
// It provides a unified interface for collecting process-level metrics.
type ProcessCollector struct {
	cpu *CPUCollector
	mem *MemoryCollector
}

// NewProcessCollector creates a new combined process metrics collector.
//
// Returns:
//   - *ProcessCollector: configured collector with default /proc path
func NewProcessCollector() *ProcessCollector {
	// Create collector with default CPU and memory collectors
	return &ProcessCollector{
		cpu: NewCPUCollector(),
		mem: NewMemoryCollector(),
	}
}

// NewProcessCollectorWithPath creates a new collector with a custom proc path.
// This is useful for testing with mock /proc filesystems.
//
// Params:
//   - procPath: custom path to proc filesystem root for testing
//
// Returns:
//   - *ProcessCollector: configured collector with custom proc path
func NewProcessCollectorWithPath(procPath string) *ProcessCollector {
	// Create collector with custom proc path for testing scenarios
	return &ProcessCollector{
		cpu: NewCPUCollectorWithPath(procPath),
		mem: NewMemoryCollectorWithPath(procPath),
	}
}

// CollectCPU collects CPU metrics for a process.
//
// Params:
//   - ctx: context for cancellation and timeout control
//   - pid: process ID to collect metrics for
//
// Returns:
//   - metrics.ProcessCPU: process CPU metrics
//   - error: context cancellation or collection errors
func (c *ProcessCollector) CollectCPU(ctx context.Context, pid int) (metrics.ProcessCPU, error) {
	// Delegate to CPU collector for process metrics
	return c.cpu.CollectProcess(ctx, pid)
}

// CollectMemory collects memory metrics for a process.
//
// Params:
//   - ctx: context for cancellation and timeout control
//   - pid: process ID to collect metrics for
//
// Returns:
//   - metrics.ProcessMemory: process memory metrics
//   - error: context cancellation or collection errors
func (c *ProcessCollector) CollectMemory(ctx context.Context, pid int) (metrics.ProcessMemory, error) {
	// Delegate to memory collector for process metrics
	return c.mem.CollectProcess(ctx, pid)
}
