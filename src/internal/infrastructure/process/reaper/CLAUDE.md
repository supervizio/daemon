<!-- updated: 2026-02-15T21:30:00Z -->
# Reaper - Zombie Process Cleanup

Zombie process cleanup loop when running as PID1.

## Context

When the daemon runs as PID1 (in a container), orphan processes are reassigned to it. Without a reaper, they become zombies.

## Interface

```go
type ZombieReaper interface {
    Start()           // Start background loop
    Stop()            // Stop the loop
    ReapOnce() int    // One manual cycle (for tests)
    IsPID1() bool     // True if PID == 1
}
```

## Files

| File | Role |
|------|------|
| `zombie_reaper.go` | `ZombieReaper` interface |
| `reaper_unix.go` | Implementation with `waitpid(-1, WNOHANG)` |

## How It Works

```go
func (r *Reaper) Start() {
    go func() {
        for {
            select {
            case <-r.stopCh:
                return
            case <-ticker.C:
                r.ReapOnce()
            }
        }
    }()
}

func (r *Reaper) ReapOnce() int {
    count := 0
    for {
        pid, _ := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
        if pid <= 0 {
            break
        }
        count++
    }
    return count
}
```

## Constructor

```go
New() *Reaper
```

## Conditional Activation

In `bootstrap/providers.go`:

```go
func ProvideReaper(r *reaper.Reaper) reaper.ZombieReaper {
    if r.IsPID1() {
        return r
    }
    return nil  // No reaper if not PID1
}
```
