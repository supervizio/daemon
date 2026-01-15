# Reaper - Domain Layer

Port interface for zombie process cleanup in PID 1 mode.

## Purpose

When a process runs as PID 1 (init) in a container, it becomes responsible for reaping orphaned child processes. Without proper reaping, these processes become zombies and consume system resources.

This package defines the **domain port** (interface) that the application layer uses. The actual implementation is in `infrastructure/process/reaper/`.

## Key Types

| Type | Role |
|------|------|
| `Reaper` | Port interface for process reaping operations |

## Interface

```go
type Reaper interface {
    Start()           // Begin background reaping loop
    Stop()            // Stop the reaping loop
    ReapOnce() int    // Manual single reap cycle (testing)
    IsPID1() bool     // Check if running as PID 1
}
```

## Dependencies

| Direction | Package |
|-----------|---------|
| Depends on | None (pure interface) |
| Used by | `application/supervisor` |
| Implemented by | `infrastructure/process/reaper` |

## Hexagonal Architecture

```
┌─────────────────────────────────────────┐
│ Application Layer (supervisor)          │
│                                         │
│   uses Reaper interface ────────────────┤
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│ Domain Layer (this package)             │
│                                         │
│   defines Reaper interface              │
└─────────────────────────────────────────┘
                    ▲
                    │ implements
┌─────────────────────────────────────────┐
│ Infrastructure Layer                    │
│                                         │
│   infrastructure/process/reaper/        │
│   - Reaper struct                       │
│   - waitpid() implementation            │
└─────────────────────────────────────────┘
```

## Related

| Package | Relation |
|---------|----------|
| `infrastructure/process/reaper/` | Implementation |
| `application/supervisor/` | Consumer |
| `bootstrap/` | Wire bindings |
