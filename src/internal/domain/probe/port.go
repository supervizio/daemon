// Package metrics provides domain types for system and process metrics collection.
package probe

import "context"

// CPUCollector defines the port interface for CPU metrics collection.
type CPUCollector interface {
	// CollectSystem collects system-wide CPU metrics.
	CollectSystem(ctx context.Context) (SystemCPU, error)
	// CollectProcess collects CPU metrics for a specific process.
	CollectProcess(ctx context.Context, pid int) (ProcessCPU, error)
	// CollectAllProcesses collects CPU metrics for all visible processes.
	CollectAllProcesses(ctx context.Context) ([]ProcessCPU, error)
	// CollectLoadAverage collects system load average.
	CollectLoadAverage(ctx context.Context) (LoadAverage, error)
	// CollectPressure collects CPU pressure metrics (PSI).
	CollectPressure(ctx context.Context) (CPUPressure, error)
}

// MemoryCollector defines the port interface for memory metrics collection.
type MemoryCollector interface {
	// CollectSystem collects system-wide memory metrics.
	CollectSystem(ctx context.Context) (SystemMemory, error)
	// CollectProcess collects memory metrics for a specific process.
	CollectProcess(ctx context.Context, pid int) (ProcessMemory, error)
	// CollectAllProcesses collects memory metrics for all visible processes.
	CollectAllProcesses(ctx context.Context) ([]ProcessMemory, error)
	// CollectPressure collects memory pressure metrics (PSI).
	CollectPressure(ctx context.Context) (MemoryPressure, error)
}

// DiskCollector defines the port interface for disk metrics collection.
type DiskCollector interface {
	// ListPartitions returns all mounted partitions.
	ListPartitions(ctx context.Context) ([]Partition, error)
	// CollectUsage collects disk usage for a specific path.
	CollectUsage(ctx context.Context, path string) (DiskUsage, error)
	// CollectAllUsage collects disk usage for all mounted partitions.
	CollectAllUsage(ctx context.Context) ([]DiskUsage, error)
	// CollectIO collects I/O statistics for all block devices.
	CollectIO(ctx context.Context) ([]DiskIOStats, error)
	// CollectDeviceIO collects I/O statistics for a specific device.
	CollectDeviceIO(ctx context.Context, device string) (DiskIOStats, error)
}

// NetworkCollector defines the port interface for network metrics collection.
type NetworkCollector interface {
	// ListInterfaces returns all network interfaces.
	ListInterfaces(ctx context.Context) ([]NetInterface, error)
	// CollectStats collects statistics for a specific interface.
	CollectStats(ctx context.Context, iface string) (NetStats, error)
	// CollectAllStats collects statistics for all interfaces.
	CollectAllStats(ctx context.Context) ([]NetStats, error)
}

// IOCollector defines the port interface for I/O metrics collection.
type IOCollector interface {
	// CollectStats collects system-wide I/O statistics.
	CollectStats(ctx context.Context) (IOStats, error)
	// CollectPressure collects I/O pressure metrics (PSI).
	CollectPressure(ctx context.Context) (IOPressure, error)
}

// SystemCollector aggregates all metrics collectors into a single interface.
// This provides a unified entry point for collecting all system metrics.
type SystemCollector interface {
	// CPU returns the CPU metrics collector.
	CPU() CPUCollector
	// Memory returns the memory metrics collector.
	Memory() MemoryCollector
	// Disk returns the disk metrics collector.
	Disk() DiskCollector
	// Network returns the network metrics collector.
	Network() NetworkCollector
	// IO returns the I/O metrics collector.
	IO() IOCollector
}
