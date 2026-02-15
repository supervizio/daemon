<!-- updated: 2026-02-15T21:30:00Z -->
# Config - Configuration Loading

Adapters for loading configuration at startup.

## Role

Parse configuration files and map them to domain types.

## Navigation

| Format | Package |
|--------|---------|
| YAML | `yaml/` |

## Structure

```
config/
└── yaml/              # YAML parser
    ├── loader.go      # Main loader
    ├── types.go       # Intermediate mapping types
    └── metrics_dto.go # Metrics configuration DTO
```

## Implemented Interface

```go
// application/config/loader.go
type Loader interface {
    Load(path string) (*config.Config, error)
}
```
