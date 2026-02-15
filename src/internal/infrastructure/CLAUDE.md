<!-- updated: 2026-02-15T21:30:00Z -->
# Infrastructure Layer

Technical adapters for the hexagonal architecture.

## Role

Implement domain interfaces (ports) with concrete technologies: OS, databases, network, files.

**Rule**: Domain never depends on infrastructure.

## Navigation

| Need | Package |
|------|---------|
| Execute processes, manage signals, credentials | `process/` |
| System metrics & quotas (CPU, RAM, cgroups) | `probe/` |
| External target discovery (Docker, systemd, K8s) | `discovery/` |
| Store data, load config | `persistence/` |
| Logs, health checks, events | `observability/` |
| gRPC API, TUI | `transport/` |

## Structure

```
infrastructure/
├── discovery/         # Target discovery (Docker, systemd, K8s, Nomad, etc.)
├── probe/             # Cross-platform metrics & quotas (Rust FFI)
├── process/           # OS process management (control, credentials, executor, reaper, signals)
├── persistence/       # Storage (config/yaml, storage/boltdb)
├── observability/     # Monitoring (healthcheck, logging, events)
└── transport/         # Communication (grpc, tui)
```

## Probe Package

The `probe/` package is the unified adapter for all system resources:
- **Metrics**: CPU, memory, disk, network, I/O (cross-platform)
- **Quotas**: cgroups (Linux), launchd (macOS), jail (BSD)

See `probe/CLAUDE.md` for details.

## Global Conventions

### Files

| Pattern | Usage |
|---------|-------|
| `{concept}.go` | Interface + public types |
| `{concept}_{platform}.go` | Platform-specific code |
| `errors.go` | Shared sentinel errors |

### Constructors

| Pattern | When |
|---------|------|
| `New()` | Standard creation |
| `NewWithDeps(...)` | Wire DI injection |
| `NewWithOptions(...)` | Tests with mocks |

### Build Tags

```go
//go:build cgo          // Requires CGO (probe package)
//go:build linux        // Linux only
//go:build darwin       // macOS only
//go:build unix         // Linux + macOS + BSD
```

### Tests

| Suffix | Type |
|--------|------|
| `_external_test.go` | Black-box (package_test) |
| `_internal_test.go` | White-box (same package) |
