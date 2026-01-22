# Domain Storage Package

Domain port interface for metrics persistence.

## Files

| File | Purpose |
|------|---------|
| `metrics_store.go` | `MetricsStore` port interface, `StoreConfig` |

## Segregated Interfaces (ISP)

### MetricsWriter
- `WriteSystemCPU(ctx, m)`, `WriteSystemMemory(ctx, m)`, `WriteProcessMetrics(ctx, m)`

### MetricsReader
- `GetSystemCPU(ctx, since, until)`, `GetSystemMemory(ctx, since, until)`
- `GetProcessMetrics(ctx, serviceName, since, until)`
- `GetLatestSystemCPU(ctx)`, `GetLatestSystemMemory(ctx)`, `GetLatestProcessMetrics(ctx, svc)`

### MetricsMaintainer
- `Prune(ctx, olderThan)` - Returns count deleted
- `Close()` - Release resources

### MetricsStore (Composed)
```go
type MetricsStore interface {
    MetricsWriter
    MetricsReader
    MetricsMaintainer
}
```

## StoreConfig

| Setting | Default |
|---------|---------|
| Path | `/var/lib/supervizio/metrics.db` |
| Retention | 24 hours |
| PruneInterval | 1 hour |

## Dependencies

- Depends on: `domain/metrics` (metric types)
- Used by: `application/metrics`, `infrastructure/persistence/storage/boltdb`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/metrics` | Types stored by this interface |
| `infrastructure/persistence/storage/boltdb` | Implements MetricsStore |
