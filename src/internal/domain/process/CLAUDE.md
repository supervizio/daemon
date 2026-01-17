# Domain Process Package

Domain entities and value objects for process lifecycle management.

This package defines the core abstractions for managing supervised processes,
including process specifications, state tracking, execution ports, restart
policies, and lifecycle events.

## Files

| File | Purpose |
|------|---------|
| `spec.go` | `Spec` - process execution specification |
| `spec_params.go` | `SpecParams` - parameters for creating Spec |
| `state.go` | `State` enum - process lifecycle states |
| `status.go` | Process status utilities |
| `executor.go` | `Executor` port interface for OS process execution |
| `exit_result.go` | `ExitResult` - process exit information |
| `restart_policy.go` | `RestartTracker` - restart attempt tracking with backoff |
| `event.go` | `Event`, `EventType` - process lifecycle events |
| `errors.go` | Domain errors for process operations |

## Key Types

### Spec (Value Object)

```go
type Spec struct {
    Command string
    Args    []string
    Dir     string
    Env     map[string]string
    User    string
    Group   string
    Stdout  io.Writer
    Stderr  io.Writer
}
```

### State (Enum)

```go
const (
    StateStopped  State = iota  // Not running, stopped normally
    StateStarting               // Process starting up
    StateRunning                // Currently executing
    StateStopping               // Graceful shutdown in progress
    StateFailed                 // Terminated with error
)
```

**State Machine:**
```
STOPPED ──→ STARTING ──→ RUNNING ──→ STOPPING ──→ STOPPED
                │            │
                └────────────┴──→ FAILED
```

### Executor (Port Interface)

```go
type Executor interface {
    Start(ctx context.Context, spec Spec) (pid int, wait <-chan ExitResult, err error)
    Stop(pid int, timeout time.Duration) error
    Signal(pid int, sig os.Signal) error
}
```

### ExitResult (Value Object)

```go
type ExitResult struct {
    Code  int    // Exit code (0 = success)
    Error error  // Any error during execution
}
```

### RestartTracker (Entity)

```go
type RestartTracker struct {
    config      *config.RestartConfig
    attempts    int
    lastAttempt time.Time
    window      time.Duration  // Default: 5 minutes
}
```

Features:
- Tracks restart attempts
- Implements exponential backoff: `delay * 2^attempts`
- Resets counter after stability window (5 min of stable running)
- Caps backoff at 30 attempts to prevent overflow

### EventType (Enum)

```go
const (
    EventStarted    EventType = iota  // Process started
    EventStopped                       // Process stopped normally
    EventFailed                        // Process failed
    EventRestarting                    // Process restarting
    EventHealthy                       // Process became healthy
    EventUnhealthy                     // Process became unhealthy
)
```

### Event (Value Object)

```go
type Event struct {
    Type      EventType
    Process   string
    PID       int
    ExitCode  int
    Timestamp time.Time
    Error     error
}
```

## Factory Functions

- `NewSpec(params)` - Create process specification
- `NewEvent(eventType, processName, pid, exitCode, err)` - Create event
- `NewRestartTracker(cfg)` - Create restart tracker

## Builder Methods

### Spec
- `WithOutput(stdout, stderr)` - Set output writers

### State Methods
- `String()` - Human-readable state name
- `IsTerminal()` - Check if stopped or failed
- `IsActive()` - Check if starting or running
- `IsRunning()` - Check if running
- `IsStopping()` - Check if stopping
- `IsStarting()` - Check if starting
- `IsFailed()` - Check if failed
- `IsStopped()` - Check if stopped

### RestartTracker Methods
- `ShouldRestart(exitCode)` - Determine if restart needed
- `RecordAttempt()` - Record restart attempt
- `Reset()` - Reset attempt counter
- `MaybeReset(uptime)` - Reset if stable
- `Attempts()` - Get current attempt count
- `NextDelay()` - Calculate next delay with backoff
- `IsExhausted()` - Check if retries exhausted
- `SetWindow(window)` - Set stability window

## Domain Errors

```go
var (
    ErrAlreadyRunning     // Process already running
    ErrNotRunning         // Process not running
    ErrMaxRetriesExceeded // Max restart retries exceeded
    ErrInvalidTransition  // Invalid state transition
    ErrProcessFailed      // Process exited with non-zero code
)
```

## Restart Policy Integration

The RestartTracker works with `config.RestartPolicy`:
- `RestartAlways` - Restart up to MaxRetries regardless of exit code
- `RestartOnFailure` - Restart only on non-zero exit, up to MaxRetries
- `RestartNever` - Never restart
- `RestartUnless` - Always restart unless explicitly stopped

## Dependencies

- Depends on: `domain/config` (for RestartConfig in RestartTracker)
- Used by: `application/process`, `infrastructure/process/executor`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/config` | RestartConfig used by RestartTracker |
| `domain/metrics` | ProcessMetrics references State |
| `domain/health` | AggregatedHealth uses State |
| `infrastructure/process/executor` | Implements Executor port |
| `application/process` | Uses process domain types |
