<!-- updated: 2026-02-15T21:30:00Z -->
# BoltDB - Key-Value Storage

BoltDB adapter for embedded storage.

## Context

BoltDB is an embedded key-value database, ideal for a daemon:
- No external server
- ACID transactions
- Single file

## Implemented Interface

```go
type MetricsStore interface {
    MetricsWriter      // WriteSystemCPU, WriteSystemMemory, WriteProcessMetrics
    MetricsReader      // GetSystemCPU, GetSystemMemory, GetProcessMetrics
    MetricsMaintainer  // Prune, Close
}
```

## Structure

| File | Role |
|------|------|
| `store.go` | `Store` wrapping `*bolt.DB` |

## Constructor

```go
New(path string) (*Store, error)
NewWithOptions(path string, opts Options) (*Store, error)
```

## Usage

```go
store, err := boltdb.New("/var/lib/supervizio/metrics.db")
defer store.Close()

err = store.WriteSystemCPU(ctx, cpuMetrics)
cpu, err := store.GetLatestSystemCPU(ctx)
```

## Buckets

BoltDB organizes data in "buckets" (equivalent to tables):

```go
// Metrics storage with time-series keys
store.WriteSystemCPU(ctx, metrics)       // bucket=system_cpu
store.WriteProcessMetrics(ctx, metrics)  // bucket=process_metrics
```
