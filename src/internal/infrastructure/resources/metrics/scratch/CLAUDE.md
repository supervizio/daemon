# Scratch Metrics Infrastructure Package

Minimal metrics collector for environments without system metrics access.

## Purpose

The scratch package provides a `metrics.SystemCollector` implementation that
returns `ErrNotSupported` for all operations. It's designed for:

- Scratch containers (no /proc filesystem)
- Windows environments (different metrics API)
- Unknown platforms
- Testing fallback behavior

## Structure

```
scratch/
└── probe.go    # ScratchProbe and all collectors
```

## Types

| Type | Implements |
|------|------------|
| `ScratchProbe` | `metrics.SystemCollector` |
| `CPUCollector` | `metrics.CPUCollector` |
| `MemoryCollector` | `metrics.MemoryCollector` |
| `DiskCollector` | `metrics.DiskCollector` |
| `NetworkCollector` | `metrics.NetworkCollector` |
| `IOCollector` | `metrics.IOCollector` |

## Behavior

All collector methods return `ErrNotSupported` with a zero-value result.
This allows the application layer to detect unsupported platforms and
handle them gracefully.

## Usage

```go
// Direct usage (rare)
probe := scratch.NewScratchProbe()
cpu, err := probe.CPU().CollectSystem(ctx)
if errors.Is(err, scratch.ErrNotSupported) {
    // Handle unsupported platform
}

// Via factory (recommended)
collector := metrics.NewSystemCollector() // Returns scratch on unknown platforms
```

## Error Handling

The application layer should check for `ErrNotSupported`:

```go
import "errors"

cpu, err := collector.CPU().CollectSystem(ctx)
if errors.Is(err, scratch.ErrNotSupported) {
    log.Warn("CPU metrics not available on this platform")
    return
}
```
