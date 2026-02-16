<!-- updated: 2026-02-15T21:30:00Z -->
# probe-metrics - Abstract Metrics Traits

Abstract metrics collection traits that platform-specific code implements.

## Structure

```
src/
└── lib.rs      # Collector traits and metric types
```

## Role

Define the collector interfaces (traits) for system and process metrics.
Platform-specific implementations live in `probe-platform`.

## Key Traits

- CPU, memory, disk, network, I/O collectors
- System-level and process-level metrics

## Related

| Crate | Relation |
|-------|----------|
| `probe-platform` | Implements these traits per OS |
| `probe-ffi` | Exposes collected metrics via FFI |
