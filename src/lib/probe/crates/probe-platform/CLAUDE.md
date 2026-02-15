<!-- updated: 2026-02-15T21:30:00Z -->
# probe-platform - Platform Implementations

Platform-specific metrics collection for Linux, macOS, BSD.

## Structure

```
src/
├── lib.rs          # Platform detection and module selection
├── linux/          # Linux: procfs, sysfs, cgroups
├── darwin/         # macOS: Mach APIs, sysctl
├── bsd/            # FreeBSD/OpenBSD/NetBSD: sysctl, kvm
└── stub/           # Fallback stubs for unsupported platforms
benches/            # Performance benchmarks
```

## Platform Support

| Platform | Source | Notes |
|----------|--------|-------|
| Linux | `/proc`, `/sys` | Full metrics support |
| macOS | Mach APIs, sysctl | Partial (no PSI) |
| FreeBSD | sysctl, procstat | Partial (no PSI) |
| OpenBSD | sysctl | Partial (limited thermal) |
| NetBSD | sysctl | Partial (limited thermal) |

## Related

| Crate | Relation |
|-------|----------|
| `probe-metrics` | Trait definitions implemented here |
| `probe-cache` | Caching layer for collected metrics |
