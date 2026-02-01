# Metrics Domain Package

Domain types and port interfaces for system and process metrics collection.

## Structure

| File | Purpose |
|------|---------|
| `cpu.go` | SystemCPU, ProcessCPU, CPUPressure, LoadAverage |
| `memory.go` | SystemMemory, ProcessMemory, MemoryPressure |
| `disk.go` | Partition, DiskUsage, DiskIOStats |
| `network.go` | NetInterface, NetStats, Bandwidth |
| `io.go` | IOStats, IOPressure |
| `process.go` | ProcessMetrics (uses process.State) |
| `collector.go` | Collector interfaces |

## Value Objects

| Type | Description |
|------|-------------|
| `SystemCPU` | System-wide CPU (user, system, idle, iowait) |
| `ProcessCPU` | Per-process CPU (utime, stime, start time) |
| `SystemMemory` | System memory (total, available, cached, swap) |
| `ProcessMemory` | Per-process memory (RSS, VMS, swap, shared) |
| `DiskUsage` | Disk space (total, used, free, inodes) |
| `NetStats` | Interface stats (bytes, packets, errors) |
| `ProcessMetrics` | Aggregated process metrics with state |

## Port Interfaces

| Interface | Purpose |
|-----------|---------|
| `CPUCollector` | Collect CPU metrics |
| `MemoryCollector` | Collect memory metrics |
| `DiskCollector` | Collect disk metrics |
| `NetworkCollector` | Collect network metrics |
| `IOCollector` | Collect I/O metrics |
| `SystemCollector` | Aggregates all collectors |

## PSI (Pressure Stall Information)

Linux 4.20+ provides PSI via `/proc/pressure/{cpu,memory,io}`.
- **some**: % time some tasks stalled
- **full**: % time all tasks stalled
- `IsUnderPressure()`: 10% threshold on 10-second average

## Dependencies

- Depends on: `domain/process` (for State)
- Used by: `application/metrics`, `infrastructure/probe`
