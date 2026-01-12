# Metrics Domain Package

Domain types and port interfaces for system and process metrics collection.

## Structure

```
metrics/
├── cpu.go              # SystemCPU, ProcessCPU value objects
├── memory.go           # SystemMemory, ProcessMemory value objects
└── port.go             # CPUCollector, MemoryCollector interfaces
```

## Types

### CPU Metrics

| Type | Description |
|------|-------------|
| `SystemCPU` | System-wide CPU metrics from /proc/stat |
| `ProcessCPU` | Per-process CPU metrics from /proc/[pid]/stat |

### Memory Metrics

| Type | Description |
|------|-------------|
| `SystemMemory` | System-wide memory metrics from /proc/meminfo |
| `ProcessMemory` | Per-process memory metrics from /proc/[pid]/status |

## Port Interfaces

| Interface | Methods |
|-----------|---------|
| `CPUCollector` | `CollectSystem`, `CollectProcess`, `CollectAllProcesses` |
| `MemoryCollector` | `CollectSystem`, `CollectProcess`, `CollectAllProcesses` |

## Dependencies

- Depends on: nothing (pure domain)
- Used by: `application/metrics`, `infrastructure/proc`

## Usage Pattern

```go
// Application layer uses port interface
type MetricsMonitor struct {
    cpu    metrics.CPUCollector
    memory metrics.MemoryCollector
}

// Infrastructure provides implementation
cpuCollector := proc.NewCPUCollector()
memCollector := proc.NewMemoryCollector()
```
