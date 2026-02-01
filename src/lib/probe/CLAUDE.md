# Probe - Rust System Metrics Library

Cross-platform Rust library providing system metrics and resource quota management.

## Structure

```
src/lib/probe/
├── Cargo.toml          # Workspace configuration
├── include/probe.h     # C header for FFI
└── crates/
    ├── probe-cache/    # Metrics caching with TTL policies
    ├── probe-ffi/      # C ABI exports (staticlib)
    ├── probe-metrics/  # Metrics collection (CPU, memory, disk, network, I/O)
    ├── probe-platform/ # Platform detection and abstractions
    ├── probe-quota/    # Resource quotas (cgroups, launchd, jail)
    └── probe-runtime/  # Container runtime detection (Docker, K8s, etc.)
```

## Crates

| Crate | Purpose |
|-------|---------|
| `probe-cache` | Metrics caching with TTL policies |
| `probe-ffi` | FFI layer exposing C ABI for Go CGO bindings |
| `probe-metrics` | System and process metrics collection |
| `probe-platform` | Platform detection, OS-specific implementations |
| `probe-quota` | Resource quota management (cgroups v1/v2, launchd, jail) |
| `probe-runtime` | Container/orchestrator runtime detection |

## Build

```bash
cargo build --release           # Build all crates
make build-probe                # Build and copy to dist/lib/
```

## Output

Static library: `dist/lib/{platform}/libprobe.a`
Header: `include/probe.h`

## Dependencies

| Crate | Usage |
|-------|-------|
| `procfs` | Linux /proc filesystem parsing |
| `mach2` | macOS Mach kernel APIs |
| `nix` | Unix system calls |
| `libc` | C FFI types |

## Related

| Location | Description |
|----------|-------------|
| `src/internal/infrastructure/probe/` | Go CGO bindings |
| `dist/lib/` | Compiled libraries |
