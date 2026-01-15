# Domain Storage Package

Domain port interface for metrics persistence.

This package defines the `MetricsStore` interface that infrastructure adapters
implement to provide persistent storage for system and process metrics, enabling
time-series queries and historical analysis.

## Files

| File | Purpose |
|------|---------|
| `metrics_store.go` | `MetricsStore` port interface, `StoreConfig` |

## Key Types

### MetricsStore (Port Interface)

```go
type MetricsStore interface {
    // Write operations
    WriteSystemCPU(ctx context.Context, m *metrics.SystemCPU) error
    WriteSystemMemory(ctx context.Context, m *metrics.SystemMemory) error
    WriteProcessMetrics(ctx context.Context, m *metrics.ProcessMetrics) error

    // Time-range queries
    GetSystemCPU(ctx context.Context, since, until time.Time) ([]metrics.SystemCPU, error)
    GetSystemMemory(ctx context.Context, since, until time.Time) ([]metrics.SystemMemory, error)
    GetProcessMetrics(ctx context.Context, serviceName string, since, until time.Time) ([]metrics.ProcessMetrics, error)

    // Latest value queries
    GetLatestSystemCPU(ctx context.Context) (metrics.SystemCPU, error)
    GetLatestSystemMemory(ctx context.Context) (metrics.SystemMemory, error)
    GetLatestProcessMetrics(ctx context.Context, serviceName string) (metrics.ProcessMetrics, error)

    // Maintenance
    Prune(ctx context.Context, olderThan time.Duration) (int, error)
    Close() error
}
```

### StoreConfig (Value Object)

```go
type StoreConfig struct {
    Path          string         // Database file path
    Retention     time.Duration  // How long to keep metrics
    PruneInterval time.Duration  // Auto-prune frequency
}
```

## Operations

### Write Operations
- `WriteSystemCPU` - Persist CPU metrics snapshot
- `WriteSystemMemory` - Persist memory metrics snapshot
- `WriteProcessMetrics` - Persist per-process metrics

### Time-Range Queries
- `GetSystemCPU(since, until)` - Query CPU history
- `GetSystemMemory(since, until)` - Query memory history
- `GetProcessMetrics(service, since, until)` - Query process history

### Latest Value Queries
- `GetLatestSystemCPU()` - Most recent CPU metrics
- `GetLatestSystemMemory()` - Most recent memory metrics
- `GetLatestProcessMetrics(service)` - Most recent process metrics

### Maintenance
- `Prune(olderThan)` - Remove old metrics, returns count deleted
- `Close()` - Release resources

## Default Configuration

```go
func DefaultStoreConfig() StoreConfig {
    return StoreConfig{
        Path:          "/var/lib/supervizio/metrics.db",
        Retention:     24 * time.Hour,
        PruneInterval: time.Hour,
    }
}
```

| Setting | Default |
|---------|---------|
| Path | `/var/lib/supervizio/metrics.db` |
| Retention | 24 hours |
| PruneInterval | 1 hour |

## Usage Pattern

```go
// Application layer uses the port interface
type MetricsRecorder struct {
    store storage.MetricsStore
}

// Record metrics periodically
func (r *MetricsRecorder) RecordCPU(ctx context.Context, cpu *metrics.SystemCPU) error {
    return r.store.WriteSystemCPU(ctx, cpu)
}

// Query historical data
func (r *MetricsRecorder) GetCPUHistory(ctx context.Context, duration time.Duration) ([]metrics.SystemCPU, error) {
    now := time.Now()
    return r.store.GetSystemCPU(ctx, now.Add(-duration), now)
}
```

## Dependencies

- Depends on: `domain/metrics` (for metric types)
- Used by: `application/metrics`, `infrastructure/persistence/storage`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/metrics` | Metric types stored by this interface |
| `infrastructure/persistence/storage/boltdb` | Implements MetricsStore with BoltDB |
| `application/metrics` | Uses MetricsStore for recording |
