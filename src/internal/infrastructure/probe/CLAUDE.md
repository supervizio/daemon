# Probe Infrastructure Package

Platform-specific adapters for system metrics collection.

## Structure

```
probe/
├── factory.go          # Platform detection and SystemCollector factory
├── linux/              # Linux /proc filesystem adapters
│   ├── cpu.go          # CPU metrics from /proc/stat
│   ├── memory.go       # Memory metrics from /proc/meminfo
│   └── collector.go    # Combined process collector
├── scratch/            # Minimal fallback for unsupported platforms
│   └── probe.go        # Returns ErrNotSupported for all operations
├── bsd/                # TODO: BSD sysctl adapters
├── darwin/             # TODO: macOS adapters
└── factory_test.go     # Factory tests
```

## Platform Detection

The factory automatically detects the best available implementation:

| Platform | Detection | Implementation |
|----------|-----------|----------------|
| Linux | `/proc/stat` exists | `linux.*Collector` |
| FreeBSD | `runtime.GOOS == "freebsd"` | `bsd.*Collector` (TODO) |
| OpenBSD | `runtime.GOOS == "openbsd"` | `bsd.*Collector` (TODO) |
| NetBSD | `runtime.GOOS == "netbsd"` | `bsd.*Collector` (TODO) |
| macOS | `runtime.GOOS == "darwin"` | `darwin.*Collector` (TODO) |
| Other | fallback | `scratch.*Collector` |

## Usage

```go
import "github.com/kodflow/daemon/internal/infrastructure/probe"

// Create platform-appropriate collector
collector := probe.NewSystemCollector()

// Log detected platform
log.Info("using metrics backend", "platform", probe.DetectedPlatform())

// Use collectors
cpuMetrics, err := collector.CPU().CollectSystem(ctx)
memMetrics, err := collector.Memory().CollectSystem(ctx)
diskUsage, err := collector.Disk().CollectUsage(ctx, "/")
netStats, err := collector.Network().CollectStats(ctx, "eth0")
```

## Error Handling

All platforms may return errors. The scratch platform always returns
`scratch.ErrNotSupported`. Other platforms return platform-specific errors.

```go
import (
    "errors"
    "github.com/kodflow/daemon/internal/infrastructure/probe/scratch"
)

cpu, err := collector.CPU().CollectSystem(ctx)
if errors.Is(err, scratch.ErrNotSupported) {
    // Platform doesn't support CPU metrics
    return
}
if err != nil {
    // Other error (permission, I/O, etc.)
    return fmt.Errorf("collect CPU: %w", err)
}
```

## Dependencies

- Depends on: `domain/probe`
- Used by: `application/metrics`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/probe` | Port interfaces |
| `application/metrics` | Orchestration layer |
| `infrastructure/probe/linux` | Linux implementation |
| `infrastructure/probe/scratch` | Fallback implementation |
