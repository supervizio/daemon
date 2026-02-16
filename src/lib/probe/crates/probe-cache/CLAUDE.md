<!-- updated: 2026-02-15T21:30:00Z -->
# probe-cache - Metrics Caching

Metrics caching with TTL policies for the probe library.

## Structure

```
src/
├── lib.rs      # Cache type and public API
├── policy.rs   # CachePolicy enum (TTL strategies)
└── ttl.rs      # TTL-based cache implementation
```

## Role

Reduce system call overhead by caching metrics values with configurable TTL.

## Related

| Crate | Relation |
|-------|----------|
| `probe-metrics` | Cached metric types |
| `probe-platform` | Uses cache for collected metrics |
