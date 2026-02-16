<!-- updated: 2026-02-15T21:30:00Z -->
# Persistence - Data Storage

Adapters for persistent storage and configuration loading.

## Role

Abstract data access: configuration files, key-value database.

## Navigation

| Need | Package |
|------|---------|
| Store key-value data | `storage/boltdb/` |
| Load YAML configuration | `config/yaml/` |

## Structure

```
persistence/
├── storage/           # Storage adapters
│   └── boltdb/        # BoltDB embedded database
│       └── store.go   # Implements domain/storage.MetricsStore
│
└── config/            # Configuration loading
    └── yaml/          # YAML parser
        ├── loader.go  # Main loader
        ├── types.go   # YAML → domain mapping types
        └── metrics_dto.go # Metrics configuration DTO
```

## Separation of Concerns

- **storage/**: Runtime persistence (state, metrics, cache)
- **config/**: Static configuration at startup
