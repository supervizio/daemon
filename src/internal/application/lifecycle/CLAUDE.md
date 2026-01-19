# Lifecycle - Process Lifecycle Management

Application service for managing individual process lifecycles with restart policies.

## Role

Coordinate process execution, monitor exit status, and apply restart policies including exponential backoff. This package was renamed from `process/` to better reflect its responsibility.

## Structure

```
lifecycle/
├── manager.go                  # ProcessManager with restart handling
├── manager_external_test.go    # Black-box tests
├── manager_internal_test.go    # White-box tests
└── signals.go                  # Signal constants (SIGHUP)
```

## Key Types

| Type | Description |
|------|-------------|
| `Manager` | Manages lifecycle of a single process with restart policies |

## Manager Methods

| Method | Description |
|--------|-------------|
| `NewManager(cfg, executor)` | Create a new process lifecycle manager |
| `Start()` | Start the managed process with automatic restart handling |
| `Stop()` | Stop the managed process |
| `Reload()` | Send SIGHUP signal for configuration reload |
| `State()` | Return current process state |
| `PID()` | Return current process PID |
| `Uptime()` | Return process uptime in seconds |
| `Events()` | Return event channel for monitoring |
| `Status()` | Return complete process status |

## Process States

The manager tracks process through these states (from `domain/process`):
- `StateStopped` - Process not running
- `StateStarting` - Process starting up
- `StateRunning` - Process is running
- `StateFailed` - Process exited with error

## Restart Handling

- Uses `domain/process.RestartTracker` for backoff calculations
- Supports oneshot services (run once, no restart)
- Emits events: `EventStarted`, `EventStopped`, `EventFailed`, `EventRestarting`

## Dependencies

- Depends on: `domain/config`, `domain/process`
- Used by: `application/supervisor`

## Related Packages

| Package | Role |
|---------|------|
| `domain/process` | Process entities, states, events, Executor port |
| `domain/config` | ServiceConfig for process configuration |
| `supervisor` | Orchestrates multiple Manager instances |
