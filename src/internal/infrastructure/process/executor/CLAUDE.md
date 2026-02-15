<!-- updated: 2026-02-15T21:30:00Z -->
# Executor - Process Execution

Implements `domain.Executor`: start, stop, signal processes.

## Interface

```go
type Executor interface {
    Start(ctx, spec) (pid int, wait <-chan ExitResult, err error)
    Stop(pid, timeout) error
    Signal(pid, sig) error
}
```

## Files

| File | Role |
|------|------|
| `executor.go` | Start/Stop/Signal implementation |
| `command.go` | `TrustedCommand()` - secure exec wrapper |
| `os_process_wrapper.go` | os.Process abstraction for testing |

## Constructors

```go
New()                      // Standalone
NewWithDeps(creds, ctrl)   // Wire DI
NewWithOptions(...)        // Tests with mocks
```

## Security

All commands go through `TrustedCommand()`:

```go
func TrustedCommand(ctx, name, args...) *exec.Cmd
```

**Trust model**: Commands come from admin YAML config, never from user input.

## Dependencies

- `credentials.CredentialManager`: user/group resolution
- `control.ProcessControl`: process group configuration
