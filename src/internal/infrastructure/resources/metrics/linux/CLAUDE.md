# Linux Metrics Infrastructure Package

Linux /proc filesystem adapters for CPU and memory metrics collection.

## Structure

```
linux/
├── cpu.go                    # CPUCollector adapter
├── cpu_external_test.go      # CPU collector tests
├── memory.go                 # MemoryCollector adapter
├── memory_external_test.go   # Memory collector tests
├── collector.go              # ProcessCollector (combined)
└── collector_external_test.go # Collector tests
```

## Types

| Type | Implements | Description |
|------|------------|-------------|
| `CPUCollector` | `metrics.CPUCollector` (partial) | Reads CPU metrics from /proc/stat and /proc/[pid]/stat |
| `MemoryCollector` | `metrics.MemoryCollector` (partial) | Reads memory metrics from /proc/meminfo and /proc/[pid]/status |
| `ProcessCollector` | Combined interface | Wraps CPU and memory collectors |

## Build Constraints

All files require Linux:
```go
//go:build linux
```

## /proc Files Parsed

| File | Content |
|------|---------|
| `/proc/stat` | System-wide CPU time (jiffies) |
| `/proc/meminfo` | System memory info (kB) |
| `/proc/[pid]/stat` | Per-process CPU time |
| `/proc/[pid]/status` | Per-process memory info |

## CPU Metrics Mapping

### System CPU (/proc/stat)
```
cpu  user nice system idle iowait irq softirq steal guest guest_nice
```

| Field | Source |
|-------|--------|
| User | Column 1 |
| Nice | Column 2 |
| System | Column 3 |
| Idle | Column 4 |
| IOWait | Column 5 |
| IRQ | Column 6 |
| SoftIRQ | Column 7 |
| Steal | Column 8 |
| Guest | Column 9 |
| GuestNice | Column 10 |

### Process CPU (/proc/[pid]/stat)

| Field | Position | Description |
|-------|----------|-------------|
| PID | 1 | Process ID |
| Name | 2 | Command name (in parentheses) |
| User | 14 | User mode time (jiffies) |
| System | 15 | Kernel mode time (jiffies) |
| ChildrenUser | 16 | Waited-for children user time |
| ChildrenSystem | 17 | Waited-for children system time |
| StartTime | 22 | Process start time |

## Memory Metrics Mapping

### System Memory (/proc/meminfo)

| Field | Source |
|-------|--------|
| Total | MemTotal |
| Free | MemFree |
| Available | MemAvailable |
| Buffers | Buffers |
| Cached | Cached |
| SwapTotal | SwapTotal |
| SwapFree | SwapFree |
| Shared | Shmem |

Derived values:
- `SwapUsed = SwapTotal - SwapFree`
- `Used = Total - Available`
- `UsagePercent = Used / Total * 100`

### Process Memory (/proc/[pid]/status)

| Field | Source |
|-------|--------|
| Name | Name |
| RSS | VmRSS |
| VMS | VmSize |
| Swap | VmSwap |
| Data | VmData |
| Stack | VmStk |
| Shared | RssShmem + RssFile |

## Usage

```go
// System metrics
cpuCollector := linux.NewCPUCollector()
memCollector := linux.NewMemoryCollector()

cpu, err := cpuCollector.CollectSystem(ctx)
mem, err := memCollector.CollectSystem(ctx)

// Process metrics
procCPU, err := cpuCollector.CollectProcess(ctx, pid)
procMem, err := memCollector.CollectProcess(ctx, pid)

// All processes
allCPU, err := cpuCollector.CollectAllProcesses(ctx)
allMem, err := memCollector.CollectAllProcesses(ctx)

// Combined collector
collector := linux.NewProcessCollector()
cpuMetrics, _ := collector.CollectCPU(ctx, pid)
memMetrics, _ := collector.CollectMemory(ctx, pid)
```

## Testing

Uses mock /proc filesystem:
```go
collector := linux.NewCPUCollectorWithPath(t.TempDir())
```

## Dependencies

- Depends on: `domain/metrics`
- Used by: `application/metrics`

## Unit Conversion

- `/proc/meminfo` values are in kB, converted to bytes (*1024)
- `/proc/stat` values are in jiffies (typically 100Hz = 10ms)

## Future Extensions

To implement the full `metrics.SystemCollector` interface, add:
- Disk collectors (`/proc/diskstats`, `/sys/block/`)
- Network collectors (`/proc/net/dev`, `/sys/class/net/`)
- PSI collectors (`/proc/pressure/{cpu,memory,io}`)
- Load average collector (`/proc/loadavg`)
