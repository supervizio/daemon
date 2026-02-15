<!-- updated: 2026-02-15T21:30:00Z -->
# Control - Process Groups

Process group management for signal forwarding.

## Context

When sending SIGTERM to a process, we also want to reach its children. Process groups enable this.

## Interface

```go
type ProcessControl interface {
    SetProcessGroup(cmd *exec.Cmd)
    GetProcessGroup(pid int) (int, error)
}
```

## Files

| File | Role |
|------|------|
| `control.go` | `ProcessControl` interface |
| `process_unix.go` | Implementation via `syscall.Setpgid` |

## Implementation

```go
func (c *Control) SetProcessGroup(cmd *exec.Cmd) {
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Setpgid: true,  // New process group
    }
}

func (c *Control) GetProcessGroup(pid int) (int, error) {
    pgid, err := syscall.Getpgid(pid)
    if err != nil {
        return 0, process.WrapError("getpgid", err)
    }
    return pgid, nil
}
```

## Usage in Executor

```go
func (e *Executor) buildCommand(ctx, spec) (*exec.Cmd, error) {
    cmd := TrustedCommand(ctx, parts[0], args...)
    e.process.SetProcessGroup(cmd)  // ← Here
    return cmd, nil
}
```

## Constructor

```go
New() *Control
```
