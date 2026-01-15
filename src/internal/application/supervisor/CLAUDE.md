# Supervisor - Service Orchestration

Application service for orchestrating multiple services and their lifecycle.

## Role

Manage the lifecycle of multiple services including start, stop, restart, and reload operations. Coordinates process managers, handles events, and tracks per-service statistics.

## Structure

```
supervisor/
├── supervisor.go                  # Supervisor main orchestrator
├── supervisor_external_test.go    # Black-box tests
├── supervisor_internal_test.go    # White-box tests
├── service_info.go                # ServiceInfo type
├── service_stats.go               # ServiceStats type
└── service_stats_external_test.go # Stats black-box tests
```

## Key Types

| Type | Description |
|------|-------------|
| `Supervisor` | Main service orchestrator managing multiple services |
| `ServiceInfo` | Runtime information about a managed service |
| `ServiceStats` | Statistics for a single service (starts, stops, failures, restarts) |
| `State` | Supervisor operational state |
| `EventHandler` | Callback function for process events |
| `Eventser` | Interface for monitoring service events |

## Supervisor Methods

| Method | Description |
|--------|-------------|
| `NewSupervisor(cfg, loader, executor, reaper)` | Create a new supervisor |
| `Start(ctx)` | Start all managed services |
| `Stop()` | Gracefully stop all managed services |
| `Reload()` | Reload configuration and restart changed services |
| `State()` | Return current supervisor state |
| `Services()` | Return information about all managed services |
| `Service(name)` | Return a specific service manager |
| `StartService(name)` | Start a specific service |
| `StopService(name)` | Stop a specific service |
| `RestartService(name)` | Restart a specific service |
| `SetEventHandler(handler)` | Set callback for process events |
| `Stats(name)` | Return statistics for a specific service |
| `AllStats()` | Return statistics for all services |

## Supervisor States

| State | Description |
|-------|-------------|
| `StateStopped` | Supervisor is not running |
| `StateStarting` | Supervisor is starting services |
| `StateRunning` | Supervisor is running |
| `StateStopping` | Supervisor is stopping services |

## Errors

| Error | Description |
|-------|-------------|
| `ErrAlreadyRunning` | Supervisor is already running |
| `ErrNotRunning` | Supervisor is not running |
| `ErrServiceNotFound` | Service not found |

## ServiceInfo Fields

| Field | Description |
|-------|-------------|
| `Name` | Service name |
| `State` | Current process state |
| `PID` | Process ID |
| `Uptime` | Uptime in seconds |

## ServiceStats Fields

| Field | Description |
|-------|-------------|
| `StartCount` | Number of times started |
| `StopCount` | Number of normal stops |
| `FailCount` | Number of failures |
| `RestartCount` | Number of automatic restarts |

## Dependencies

- Depends on: `application/config`, `application/lifecycle`, `domain/config`, `domain/lifecycle`, `domain/process`
- Used by: `cmd/daemon`, `bootstrap`

## Related Packages

| Package | Role |
|---------|------|
| `application/lifecycle` | Manager for individual process lifecycle |
| `application/config` | Loader interface for configuration |
| `domain/config` | Config entity |
| `domain/lifecycle` | Reaper interface for zombie processes |
| `domain/process` | Process entities, states, events, Executor port |
