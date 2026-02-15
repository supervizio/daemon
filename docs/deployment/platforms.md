# Platform Support

superviz.io is built with cross-platform support using build tags and the Rust FFI probe library.

---

## Support Matrix

| Feature | Linux amd64 | Linux arm64 | macOS amd64 | macOS arm64 | FreeBSD | OpenBSD | NetBSD |
|---------|------------|------------|-------------|-------------|---------|---------|--------|
| Process supervision | Full | Full | Full | Full | Full | Full | Full |
| PID1 mode | Full | Full | N/A | N/A | Full | Full | Full |
| CPU metrics | Full | Full | Partial | Partial | Partial | Partial | Partial |
| Memory metrics | Full | Full | Partial | Partial | Partial | Partial | Partial |
| Disk metrics | Full | Full | Partial | Partial | Partial | Partial | Partial |
| Network metrics | Full | Full | Partial | Partial | Partial | Partial | Partial |
| Resource quotas | cgroups v1/v2 | cgroups v1/v2 | getrlimit | getrlimit | rctl | getrlimit | getrlimit |
| Docker discovery | Yes | Yes | Yes | Yes | Yes | - | - |
| Podman discovery | Yes | Yes | Yes | Yes | - | - | - |
| systemd discovery | Yes | Yes | - | - | - | - | - |
| OpenRC discovery | Yes | Yes | - | - | - | - | - |
| BSD rc discovery | - | - | - | - | Yes | Yes | Yes |
| K8s discovery | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| Nomad discovery | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| Port scan | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| TCP probes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| HTTP probes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| gRPC probes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| ICMP probes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| TUI interactive | Yes | Yes | Yes | Yes | Yes | Yes | Yes |

---

## Build Tags

Platform-specific code is selected at compile time:

```go
//go:build cgo          // Requires CGO (probe package)
//go:build linux        // Linux only
//go:build darwin       // macOS only
//go:build unix         // Linux + macOS + BSD
//go:build bsd          // FreeBSD + OpenBSD + NetBSD
```

---

## Build Requirements

| Platform | Toolchain | Notes |
|----------|-----------|-------|
| Linux | Go 1.25.6, Rust stable, Zig 0.13.0 | Full support, CGO required for probe |
| macOS | Go 1.25.6, Rust stable | Partial metrics via sysctl |
| FreeBSD | Go 1.25.6, Rust stable | Partial metrics via sysctl, rctl quotas |
| Cross-compile | Zig as CC | Zig provides C cross-compiler for CGO |

---

## Graceful Degradation

On platforms where certain features are unavailable:

- **Missing metrics**: Returns zero values, TUI shows "-"
- **Missing discovery**: Discoverer not created, no error
- **Missing quotas**: Quota detection returns "unlimited"
- **Missing capabilities**: ICMP falls back to ping command
