# Ports - Interface Contracts

Interfaces defining OS contracts.

## Structure

```
ports/
├── signals.go       # SignalManager interface
├── process.go       # ProcessControl interface
├── reaper.go        # ZombieReaper interface
├── credentials.go   # CredentialManager interface
└── errors.go        # Port error types
```

## Principle

Ports define **what** the application expects from the OS,
without specifying **how** it's implemented.

## Interfaces

### SignalManager (signals.go)

```go
type SignalManager interface {
    Notify(c chan<- os.Signal, sig ...os.Signal)
    Stop(c chan<- os.Signal)
    Forward(pid int, sig os.Signal) error
    IsTermSignal(sig os.Signal) bool
    SetSubreaper() error
}
```

### ProcessControl (process.go)

```go
type ProcessControl interface {
    SetProcessGroup(cmd *exec.Cmd) error
}
```

### ZombieReaper (reaper.go)

```go
type ZombieReaper interface {
    Start()
    Stop()
    ReapOnce() (pid int, err error)
    IsPID1() bool
}
```

### CredentialManager (credentials.go)

```go
type CredentialManager interface {
    ResolveCredentials(user, group string) (*Credentials, error)
    ApplyCredentials(creds *Credentials) error
}
```

## Related Directories

| Directory | Relation |
|-----------|----------|
| `../adapters/` | Implements these interfaces |
| `../../process/` | Consumes these interfaces |
