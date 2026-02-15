# Supervisor - Service Orchestration

Application service for orchestrating multiple services and their lifecycle.

## Structure

```
supervisor/
├── supervisor.go                     # Main orchestrator
├── supervisor_external_test.go       # Black-box tests
├── supervisor_internal_test.go       # White-box tests
├── supervisor_benchmark_test.go      # Performance benchmarks
├── service_info.go                   # ServiceInfo type
├── service_stats.go                  # ServiceStats type
├── service_stats_external_test.go    # Stats tests
├── service_stats_snapshot.go         # Stats snapshot for TUI
├── service_snapshot_for_tui.go       # Service snapshot for TUI display
├── listener_snapshot_for_tui.go      # Listener snapshot for TUI display
├── ports_linux.go                    # Linux-specific port detection
├── ports_linux_internal_test.go      # Port detection tests
└── ports_other.go                    # Non-Linux port stub
```

## Key Types

| Type | Description |
|------|-------------|
| `Supervisor` | Main orchestrator managing multiple services |
| `ServiceInfo` | Runtime info (Name, State, PID, Uptime) |
| `ServiceStats` | Stats (StartCount, StopCount, FailCount, RestartCount) |
| `State` | Supervisor state enum |
| `EventHandler` | Callback for process events |

## Supervisor Methods

| Method | Description |
|--------|-------------|
| `NewSupervisor(cfg, loader, executor, reaper)` | Create supervisor |
| `Start(ctx)` / `Stop()` | Start/stop all services |
| `Reload()` | Reload config, restart changed services |
| `State()` / `Services()` | Get state and service info |
| `Service(name)` | Get specific service manager |
| `StartService` / `StopService` / `RestartService` | Per-service control |
| `SetEventHandler(handler)` | Set event callback |
| `Stats(name)` / `AllStats()` | Get statistics |

## States

`StateStopped` → `StateStarting` → `StateRunning` → `StateStopping` → `StateStopped`

## Errors

| Error | Description |
|-------|-------------|
| `ErrAlreadyRunning` | Supervisor already running |
| `ErrNotRunning` | Supervisor not running |
| `ErrServiceNotFound` | Service not found |

## Error Handling

Non-fatal errors via optional `SetErrorHandler(handler)`:
- Errors during shutdown (stop services)
- Errors during reload (stop/start services)
- Silently discarded if no handler set

## Dependencies

- Depends on: `application/config`, `application/lifecycle`, `domain/config`, `domain/lifecycle`, `domain/process`
- Used by: `cmd/daemon`, `bootstrap`

## Related Packages

| Package | Role |
|---------|------|
| `application/lifecycle` | Per-process lifecycle manager |
| `application/config` | Config loader interface |
| `domain/lifecycle` | Reaper interface |
| `domain/process` | Process entities, Executor port |
