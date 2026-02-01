# Metrics - Process Metrics Tracking

Application service for tracking process-level metrics (CPU, memory) for supervised services.

## Role

Aggregate CPU and memory metrics per supervised process, track process state changes, and publish metric updates to subscribers. Uses infrastructure collectors for actual metric collection.

## Structure

```
metrics/
├── tracker.go                  # Tracker implementation
├── tracker_external_test.go    # Black-box tests
└── collector.go                # Collector and ProcessTracker interfaces
```

## Key Types

| Type | Description |
|------|-------------|
| `Tracker` | Tracks metrics for all supervised processes |
| `ProcessTracker` | Interface for process metrics tracking |
| `Collector` | Port interface for collecting process metrics |
| `TrackerOption` | Functional option for configuring Tracker |

## Tracker Methods

| Method | Description |
|--------|-------------|
| `NewTracker(collector, opts...)` | Create a new process metrics tracker |
| `Start(ctx)` | Begin the metrics collection loop |
| `Stop()` | Stop the metrics collection loop |
| `Track(ctx, serviceName, pid)` | Start tracking metrics for a service |
| `Untrack(serviceName)` | Stop tracking metrics for a service |
| `Get(serviceName)` | Get current metrics for a service |
| `GetAll()` | Get metrics for all tracked services |
| `Subscribe()` | Return channel receiving metrics updates |
| `Unsubscribe(ch)` | Remove a subscription channel |
| `UpdateState(serviceName, state, lastError)` | Update process state |
| `UpdateHealth(serviceName, healthy)` | Update health status |

## Port Interfaces

```go
// Collector abstracts the collection of process metrics.
// Implemented by infrastructure adapters (e.g., /proc readers).
type Collector interface {
    CollectCPU(ctx context.Context, pid int) (ProcessCPU, error)
    CollectMemory(ctx context.Context, pid int) (ProcessMemory, error)
}

// ProcessTracker defines the interface for tracking process-level metrics.
type ProcessTracker interface {
    Track(ctx context.Context, serviceName string, pid int) error
    Untrack(serviceName string)
    Get(serviceName string) (ProcessMetrics, bool)
    GetAll() []ProcessMetrics
    Subscribe() <-chan ProcessMetrics
    Unsubscribe(ch <-chan ProcessMetrics)
}
```

## Options

| Option | Description |
|--------|-------------|
| `WithCollectionInterval(d)` | Set the metrics collection interval (default: 5s) |

## Dependencies

- Depends on: `domain/metrics`, `domain/process`
- Used by: `cmd/daemon`, gRPC streaming
- Implemented by: `infrastructure/probe`

## Related Packages

| Package | Role |
|---------|------|
| `domain/metrics` | ProcessMetrics, ProcessCPU, ProcessMemory types |
| `domain/process` | Process State enum |
| `infrastructure/probe` | Cross-platform Collector implementation (Rust FFI) |
