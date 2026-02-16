<!-- updated: 2026-02-15T21:30:00Z -->
# Process - OS Process Management

Everything related to Unix process execution and control.

## Role

Abstract OS operations: start/stop processes, send signals, manage credentials, clean up zombies.

## Navigation

| Need | Package |
|------|---------|
| Start/stop a process | `executor/` |
| Send signals (SIGTERM, SIGHUP) | `signals/` |
| Reap zombie processes (PID1) | `reaper/` |
| Resolve user/group to UID/GID | `credentials/` |
| Manage process groups | `control/` |

## Structure

```
process/
├── errors.go       # Shared errors across sub-packages
├── executor/       # Start(), Stop(), Signal()
├── signals/        # Notification, forwarding, subreaper
├── reaper/         # waitpid() loop for PID1
├── credentials/    # LookupUser(), ApplyCredentials()
└── control/        # SetProcessGroup(), GetProcessGroup()
```

## Shared Errors (errors.go)

```go
var (
    ErrProcessNotFound  = errors.New("process not found")
    ErrPermissionDenied = errors.New("permission denied")
    ErrNotSupported     = errors.New("operation not supported")
)

func WrapError(op string, err error) error  // Adds context
```

## Main Flow

```
Supervisor
    │
    ▼
executor.Start(spec)
    ├── credentials.ResolveCredentials(user, group)
    ├── credentials.ApplyCredentials(cmd, uid, gid)
    ├── control.SetProcessGroup(cmd)
    └── cmd.Start()

executor.Stop(pid, timeout)
    └── signals.Forward(pid, SIGTERM)
        └── if timeout → Kill()

reaper.Start()  ← Runs in background if PID1
    └── waitpid(-1) in loop
```
