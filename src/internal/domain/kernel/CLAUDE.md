# Domain Kernel - OS Abstraction Ports

Domain interfaces for operating system abstraction.

## Structure

```
kernel/
└── port.go    # ZombieReaper interface
```

## Interfaces

### ZombieReaper

```go
type ZombieReaper interface {
    Start()
    Stop()
    ReapOnce() int
    IsPID1() bool
}
```

Handles zombie process reaping for PID1 scenarios in containers.

## Dependencies

- Depends on: nothing (pure domain)
- Implemented by: `infrastructure/kernel/adapters`
- Used by: `application/supervisor`

## Why Domain Kernel?

The application layer needs to interact with OS-level abstractions
without depending on infrastructure directly. This package provides
the port interfaces that infrastructure adapters implement.

## Related Directories

| Directory | Relation |
|-----------|----------|
| `../../infrastructure/kernel/` | Implements these interfaces |
| `../../application/supervisor/` | Uses ZombieReaper port |
