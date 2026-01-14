# BSD Probe Infrastructure Package

System metrics collection for BSD variants (stub implementation).

## Status: Not Implemented

This package contains stub implementations that return `ErrNotImplemented`.
It serves as a placeholder for future BSD-specific metrics collection.

## Supported Systems

- FreeBSD
- OpenBSD
- NetBSD
- DragonFly BSD

## Build Constraint

```go
//go:build freebsd || openbsd || netbsd || dragonfly
```

## Future Implementation Notes

### CPU Metrics
- Use `sysctl hw.ncpu` for CPU count
- Use `sysctl kern.cp_time` for CPU time breakdown
- Process-level metrics via `kinfo_proc`

### Memory Metrics
- Use `sysctl hw.physmem` for total memory
- Use `sysctl vm.stats` for detailed memory stats
- Process-level via `kinfo_proc` and `kvm_getprocs`

### Disk Metrics
- Use `geom` library for disk statistics
- Use `statfs` for mount point usage
- Device I/O via `devstat`

### Network Metrics
- Use `getifaddrs` for interface enumeration
- Use `sysctl net.link` for interface statistics
- Consider `netstat` parsing as fallback

### Note on PSI
BSD systems don't have Linux PSI (Pressure Stall Information).
Pressure methods will always return `ErrNotImplemented`.
