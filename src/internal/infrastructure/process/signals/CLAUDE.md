<!-- updated: 2026-02-15T21:30:00Z -->
# Signals - Unix Signal Management

Notification, forwarding, and system signal handling.

## Interface

```go
type SignalManager interface {
    Notify(signals ...os.Signal) <-chan os.Signal
    Stop(ch chan<- os.Signal)
    Forward(pid int, sig os.Signal) error
    ForwardToGroup(pgid int, sig syscall.Signal) error
    IsTermSignal(sig os.Signal) bool
    IsReloadSignal(sig os.Signal) bool
    SignalByName(name string) (os.Signal, bool)
    SetSubreaper() error      // Linux only
    ClearSubreaper() error
    IsSubreaper() (bool, error)
}
```

## Files

| File | Role |
|------|------|
| `manager.go` | `SignalManager` interface |
| `signals_unix.go` | Base implementation (SIGTERM, SIGHUP, etc.) |
| `signals_linux.go` | Linux extensions (SIGRTMIN, subreaper via prctl) |
| `signals_darwin.go` | macOS (subreaper not supported) |
| `signals_bsd.go` | BSD (subreaper not supported) |

## Subreaper

Linux allows becoming a "subreaper": orphans are reassigned to us instead of init.

```go
// Linux
func (m *Manager) SetSubreaper() error {
    return prctl(PR_SET_CHILD_SUBREAPER, 1)
}

// macOS/BSD
func (m *Manager) SetSubreaper() error {
    return ErrSignalNotSupported
}
```

## Constructor

```go
New() *Manager
```
