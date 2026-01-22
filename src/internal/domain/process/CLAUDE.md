# Domain Process Package

Domain entities and value objects for process lifecycle management.

## Files

| File | Purpose |
|------|---------|
| `spec.go` | `Spec` - process execution specification |
| `spec_params.go` | `SpecParams` - parameters for creating Spec |
| `state.go` | `State` enum - process lifecycle states |
| `executor.go` | `Executor` port interface |
| `exit_result.go` | `ExitResult` - exit information |
| `restart_policy.go` | `RestartTracker` - restart with backoff |
| `event.go` | `Event`, `EventType` - lifecycle events |
| `errors.go` | Domain errors |

## Key Types

### Spec (Value Object)
- `Command`, `Args`, `Dir`, `Env`, `User`, `Group`, `Stdout`, `Stderr`
- Factory: `NewSpec(params)`
- Builder: `WithOutput(stdout, stderr)`

### State (Enum)
- `StateStopped` → `StateStarting` → `StateRunning` → `StateStopping` → `StateStopped`
- `StateFailed` (from Starting or Running)
- Methods: `IsTerminal()`, `IsActive()`, `IsRunning()`, `IsFailed()`

### Executor (Port Interface)
```go
type Executor interface {
    Start(ctx, spec) (pid int, wait <-chan ExitResult, err error)
    Stop(pid int, timeout time.Duration) error
    Signal(pid int, sig os.Signal) error
}
```

### ExitResult
- `Code int` - Exit code (0 = success)
- `Error error` - Any execution error

### RestartTracker
- Tracks restart attempts with exponential backoff
- Resets after stability window (5 min stable)
- Methods: `ShouldRestart(exitCode)`, `RecordAttempt()`, `NextDelay()`, `IsExhausted()`

### EventType
- `EventStarted`, `EventStopped`, `EventFailed`, `EventRestarting`
- `EventHealthy`, `EventUnhealthy`

## Domain Errors

```go
var (
    ErrAlreadyRunning     // Process already running
    ErrNotRunning         // Process not running
    ErrMaxRetriesExceeded // Max restart retries exceeded
    ErrInvalidTransition  // Invalid state transition
    ErrProcessFailed      // Non-zero exit code
)
```

## Dependencies

- Depends on: `domain/config` (RestartConfig)
- Used by: `application/lifecycle`, `infrastructure/process/executor`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/config` | RestartConfig for RestartTracker |
| `domain/metrics` | ProcessMetrics references State |
| `infrastructure/process/executor` | Implements Executor port |
