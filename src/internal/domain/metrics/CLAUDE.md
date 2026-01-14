# Metrics Domain Package

Domain types and port interfaces for system and process metrics collection.

## Structure

```
metrics/
├── cpu.go              # SystemCPU, ProcessCPU value objects
├── memory.go           # SystemMemory, ProcessMemory value objects
├── disk.go             # Partition, DiskUsage, DiskIOStats value objects
├── network.go          # NetInterface, NetStats, Bandwidth value objects
├── io.go               # IOStats, IOPressure, MemoryPressure, CPUPressure, LoadAverage
├── process.go          # ProcessMetrics (uses process.State from domain/process)
└── collector.go        # Collector interfaces
```

## Value Objects

### CPU Metrics

| Type | Description |
|------|-------------|
| `SystemCPU` | System-wide CPU metrics (user, system, idle, iowait, etc.) |
| `ProcessCPU` | Per-process CPU metrics (utime, stime, start time) |
| `CPUPressure` | CPU pressure metrics from PSI |
| `LoadAverage` | System load average (1, 5, 15 minutes) |

### Memory Metrics

| Type | Description |
|------|-------------|
| `SystemMemory` | System-wide memory (total, available, cached, swap) |
| `ProcessMemory` | Per-process memory (RSS, VMS, swap, shared) |
| `MemoryPressure` | Memory pressure metrics from PSI |

### Disk Metrics

| Type | Description |
|------|-------------|
| `Partition` | Disk partition info (device, mountpoint, fstype) |
| `DiskUsage` | Disk space usage (total, used, free, inodes) |
| `DiskIOStats` | Block device I/O statistics |

### Network Metrics

| Type | Description |
|------|-------------|
| `NetInterface` | Network interface info (name, MAC, MTU, addresses) |
| `NetStats` | Interface statistics (bytes, packets, errors, drops) |
| `Bandwidth` | Calculated bandwidth from two samples |

### I/O Metrics

| Type | Description |
|------|-------------|
| `IOStats` | System-wide I/O statistics |
| `IOPressure` | I/O pressure metrics from PSI |

### Process Metrics

| Type | Description |
|------|-------------|
| `ProcessMetrics` | Aggregated process metrics with state (uses `process.State`) |

Note: Process lifecycle state is defined in `domain/process.State` (not duplicated here).

## Port Interfaces

| Interface | Purpose |
|-----------|---------|
| `CPUCollector` | Collect CPU metrics (system, process, load, pressure) |
| `MemoryCollector` | Collect memory metrics (system, process, pressure) |
| `DiskCollector` | Collect disk metrics (partitions, usage, I/O) |
| `NetworkCollector` | Collect network metrics (interfaces, stats) |
| `IOCollector` | Collect I/O metrics (stats, pressure) |
| `SystemCollector` | Aggregates all collectors |

## Dependencies

- Depends on: `domain/process` (for `process.State` in `ProcessMetrics`)
- Used by: `application/metrics`, `infrastructure/metrics/linux`

## Usage Pattern

```go
// Application layer uses port interface
type MetricsMonitor struct {
    collector metrics.SystemCollector
}

// Access specific collectors
cpu := monitor.collector.CPU()
mem := monitor.collector.Memory()
disk := monitor.collector.Disk()
net := monitor.collector.Network()
io := monitor.collector.IO()

// Collect metrics
sysCPU, err := cpu.CollectSystem(ctx)
load, err := cpu.CollectLoadAverage(ctx)
diskUsage, err := disk.CollectUsage(ctx, "/")
netStats, err := net.CollectStats(ctx, "eth0")
```

## PSI (Pressure Stall Information)

Linux 4.20+ provides PSI metrics via `/proc/pressure/{cpu,memory,io}`.
These indicate resource contention levels:

- **some**: Percentage of time some tasks are stalled
- **full**: Percentage of time all tasks are stalled

The `IsUnderPressure()` methods use 10% on the 10-second average as threshold.
