# Scratch Probe Infrastructure Package

Minimal metrics collector for environments without system metrics access.

## Purpose

The scratch package provides a `probe.SystemCollector` implementation that
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
| `ScratchProbe` | `probe.SystemCollector` |
| `CPUCollector` | `probe.CPUCollector` |
| `MemoryCollector` | `probe.MemoryCollector` |
| `DiskCollector` | `probe.DiskCollector` |
| `NetworkCollector` | `probe.NetworkCollector` |
| `IOCollector` | `probe.IOCollector` |

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
collector := probe.NewSystemCollector() // Returns scratch on unknown platforms
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
