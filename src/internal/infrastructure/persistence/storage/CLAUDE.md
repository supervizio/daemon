<!-- updated: 2026-02-15T21:30:00Z -->
# Storage - Storage Adapters

Persistent storage implementations.

## Role

Provide adapters for persistent data storage and retrieval.

## Navigation

| Backend | Package |
|---------|---------|
| BoltDB (embedded) | `boltdb/` |

## Structure

```
storage/
└── boltdb/           # Embedded database
    └── store.go      # Store implementing domain/storage.MetricsStore
```

## Implemented Interface

```go
// domain/storage/metrics_store.go
type MetricsStore interface {
    MetricsWriter
    MetricsReader
    MetricsMaintainer
}
```
