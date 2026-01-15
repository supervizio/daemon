# Signals - Gestion des Signaux Unix

Notification, forwarding et gestion des signaux système.

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

## Fichiers

| Fichier | Rôle |
|---------|------|
| `manager.go` | Interface `SignalManager` |
| `signals_unix.go` | Implémentation de base (SIGTERM, SIGHUP, etc.) |
| `signals_linux.go` | Extensions Linux (SIGRTMIN, subreaper via prctl) |
| `signals_darwin.go` | macOS (subreaper non supporté) |
| `signals_bsd.go` | BSD (subreaper non supporté) |

## Subreaper

Linux permet de devenir "subreaper" : les orphelins sont réassignés à nous plutôt qu'à init.

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

## Constructeur

```go
New() *Manager
```
