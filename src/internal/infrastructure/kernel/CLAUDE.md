# Kernel - OS Abstraction Layer

OS abstraction layer using hexagonal architecture (ports & adapters).

## Structure

```
kernel/
├── kernel.go        # Default kernel instance
├── errors.go        # Kernel-level errors
├── ports/           # Interface definitions
│   ├── signals.go       # SignalHandler interface
│   ├── process.go       # ProcessConfigurator interface
│   ├── reaper.go        # ZombieReaper interface
│   ├── credentials.go   # CredentialsResolver interface
│   ├── user.go          # User lookup interface
│   ├── group.go         # Group lookup interface
│   └── errors.go        # Port errors
└── adapters/        # Platform-specific implementations
    ├── signals_linux.go     # Linux (prctl subreaper)
    ├── signals_unix.go      # Generic Unix
    ├── signals_bsd.go       # BSD variants
    ├── reaper_unix.go       # Zombie reaping (wait4)
    ├── process_unix.go      # Process groups
    └── credentials_unix.go  # User/group resolution
```

## Hexagonal Architecture

```
           ┌─────────────────────────┐
           │      Application        │
           │  (supervisor, process)  │
           └───────────┬─────────────┘
                       │ uses
           ┌───────────▼─────────────┐
           │        PORTS            │
           │  (abstract interfaces)  │
           └───────────┬─────────────┘
                       │ implements
           ┌───────────▼─────────────┐
           │       ADAPTERS          │
           │ (linux, bsd, darwin)    │
           └─────────────────────────┘
```

## Ports (Interfaces)

### SignalHandler

```go
type SignalHandler interface {
    Handle(ctx context.Context, signals <-chan os.Signal)
    Forward(pid int, sig os.Signal) error
    SetSubreaper() error  // Linux only
}
```

### ZombieReaper

```go
type ZombieReaper interface {
    Start()
    Stop()
    ReapOnce() int
    IsPID1() bool
}
```

### CredentialsResolver

```go
type CredentialsResolver interface {
    ResolveCredentials(user, group string) (uid, gid uint32, err error)
    ApplyCredentials(cmd *exec.Cmd, uid, gid uint32) error
}
```

## Adapters by Platform

| File | Platform | Specifics |
|------|----------|-----------|
| `signals_linux.go` | Linux | prctl(PR_SET_CHILD_SUBREAPER) |
| `signals_bsd.go` | FreeBSD, OpenBSD | wait4 variants |
| `signals_unix.go` | All Unix | Generic signal handling |
| `reaper_unix.go` | All Unix | WNOHANG wait loop |

## Usage

```go
import "github.com/kodflow/daemon/internal/kernel"

// Use default kernel instance
kernel.Default.Signals.Handle(ctx, ch)
kernel.Default.Credentials.ResolveCredentials(user, group)
kernel.Default.Reaper.Start()
```

## Related Directories

| Directory | Relation |
|-----------|----------|
| `../../application/process/` | Uses for exec and signals |
| `../../application/supervisor/` | Uses reaper if PID 1 |
| `../process/` | Uses for executor |
