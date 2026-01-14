//go:build linux

// Package proc provides Linux /proc filesystem adapters for metrics collection.
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
func NewProcessCollector() *ProcessCollector {
	return &ProcessCollector{
		cpu: NewCPUCollector(),
		mem: NewMemoryCollector(),
	}
}

// NewProcessCollectorWithPath creates a new collector with a custom proc path.
// This is useful for testing with mock /proc filesystems.
func NewProcessCollectorWithPath(procPath string) *ProcessCollector {
	return &ProcessCollector{
		cpu: NewCPUCollectorWithPath(procPath),
		mem: NewMemoryCollectorWithPath(procPath),
	}
}

// CollectCPU collects CPU metrics for a process.
func (c *ProcessCollector) CollectCPU(ctx context.Context, pid int) (metrics.ProcessCPU, error) {
	return c.cpu.CollectProcess(ctx, pid)
}

// CollectMemory collects memory metrics for a process.
func (c *ProcessCollector) CollectMemory(ctx context.Context, pid int) (metrics.ProcessMemory, error) {
	return c.mem.CollectProcess(ctx, pid)
}
