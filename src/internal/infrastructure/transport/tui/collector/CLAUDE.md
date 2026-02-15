<!-- updated: 2026-02-15T21:30:00Z -->
# Collector - Data Collectors

System data collectors for TUI snapshots using procfs/sysfs.

## Structure

```
collector/
├── collector.go         # Main Collector, CollectAll()
├── context.go           # Context metadata (hostname, version)
├── context_linux.go     # Linux kernel version, primary IP
├── context_darwin.go    # macOS-specific context
├── context_bsd.go       # BSD-specific context
├── sandbox.go           # Container runtime detection
├── sandbox_check.go     # Sandbox check configuration
├── system_linux.go      # CPU, RAM, swap, disk from procfs
├── system_other.go      # Stub for non-Linux platforms
├── limits.go            # Cgroup limits base interface
├── limits_linux.go      # Cgroup limits (v1/v2)
├── limits_other.go      # Stub limits
├── network.go           # Network collection base interface
├── network_linux.go     # Network interface stats from /proc/net
├── network_other.go     # Stub network stats
├── cpu_sample_linux.go  # CPU usage sampling
├── dns_config_result.go # DNS configuration result type
├── net_stats.go         # Network statistics type
└── runtime_mode_result.go # Runtime mode detection
```

## Key Functions

| Function | Description |
|----------|-------------|
| `Collector.CollectAll()` | Gather complete Snapshot |
| `getKernelVersion()` | Platform-specific kernel info |
| `getPrimaryIP()` | Primary non-loopback IP |
| `SandboxCollector.Gather()` | Detect Docker/Podman/K8s |

## Data Sources

| Data | Source |
|------|--------|
| CPU/RAM | `/proc/meminfo`, `/proc/stat` |
| Disk | `syscall.Statfs` |
| Network | `/proc/net/dev` |
| Cgroups | `/sys/fs/cgroup` |
| Sandboxes | Socket file existence |

## Constraints

- No `exec.Command`: All data from procfs/sysfs
- Platform-specific files with build tags
- Graceful degradation on missing data
