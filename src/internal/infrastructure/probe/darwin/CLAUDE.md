# Darwin Probe Infrastructure Package

System metrics collection for macOS (stub implementation).

## Status: Not Implemented

This package contains stub implementations that return `ErrNotImplemented`.
It serves as a placeholder for future macOS-specific metrics collection.

## Build Constraint

```go
//go:build darwin
```

## Future Implementation Notes

### CPU Metrics
- Use `sysctl hw.ncpu` for CPU count
- Use `host_processor_info` Mach API for CPU time
- Use `processor_info` for per-core stats
- Use `sysctl vm.loadavg` for load average

### Memory Metrics
- Use `sysctl hw.memsize` for total memory
- Use `host_statistics64` Mach API for detailed stats
- Use `vm_statistics64` for page-level info
- Process-level via `proc_pidinfo`

### Disk Metrics
- Use `statfs` for mount point usage
- Use `getmntinfo` for partition enumeration
- Use IOKit for disk I/O statistics

### Network Metrics
- Use `getifaddrs` for interface enumeration
- Use `sysctl net.link.generic.ifdata` for statistics
- Consider `nettop` parsing as fallback

### Note on PSI
macOS doesn't have Linux PSI (Pressure Stall Information).
Pressure methods will always return `ErrNotImplemented`.
macOS has its own memory pressure notification system but
it's event-based rather than metrics-based.
