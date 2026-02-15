# Probe - Cross-Platform System Resource Collector

CGO bindings to the Rust probe library for unified cross-platform system metrics and resource quota management.

## Role

Unified adapter for system resources via CGO/FFI to Rust library:
- **Metrics**: CPU, memory, disk, network, I/O (cross-platform)
- **Quotas**: cgroups (Linux), launchd (macOS), jail (BSD)

**Architecture**: Go → CGO/FFI → libprobe.a (Rust) → OS APIs

## Files

| File | Purpose |
|------|---------|
| `bindings.go` | CGO configuration and FFI helpers |
| `errors.go` | Error codes and sentinel errors |
| `collector.go` | SystemCollector implementation |
| `cpu.go` | CPUCollector + LoadAverage |
| `memory.go` | MemoryCollector |
| `disk.go` | DiskCollector (partitions, usage, I/O) |
| `network.go` | NetworkCollector (interfaces, stats) |
| `io.go` | IOCollector (stats, pressure) |
| `quota.go` | Resource quota management (cgroups/launchd/jail) |

## Usage

```go
probe.Init()                    // Initialize once at startup
defer probe.Shutdown()
c := probe.NewCollector()       // Create collector
cpu, _ := c.Cpu().CollectSystem(ctx)
mem, _ := c.Memory().CollectSystem(ctx)
```

## Build

```bash
make build-probe    # Build Rust library
make build-daemon   # Build Go with probe linked
make build-hybrid   # Build both
```

## Supported Platforms

| Platform | Metrics | Quota Detection |
|----------|---------|-----------------|
| Linux (amd64/arm64, glibc+musl) | ✅ Full | ✅ cgroups v1/v2 |
| Linux (arm/386/riscv64, glibc+musl) | ✅ Full | ✅ cgroups v1/v2 |
| macOS (amd64/arm64) | ✅ Partial | ✅ getrlimit |
| FreeBSD (amd64) | ✅ Partial | ✅ rctl |
| OpenBSD (amd64) | ✅ Partial | ✅ getrlimit |
| NetBSD (amd64) | ✅ Partial | ✅ getrlimit |

## Related

| Location | Description |
|----------|-------------|
| `src/lib/probe/` | Rust source code |
| `dist/lib/` | Compiled libraries |
| `domain/metrics/` | Port interfaces |
