# Config - Configuration Port

Application port interface for configuration loading.

## Role

Define the port interface that infrastructure adapters implement for loading and reloading service configuration. This is a pure interface package with no implementation.

## Structure

```
config/
└── loader.go    # Loader and Reloader interfaces
```

## Key Types

| Type | Description |
|------|-------------|
| `Loader` | Port interface for loading configuration from a path |
| `Reloader` | Port interface for reloading configuration at runtime |

## Port Interfaces

```go
// Loader loads configuration from external sources.
// Infrastructure adapters implement this interface.
type Loader interface {
    // Load loads configuration from the given path.
    Load(path string) (*config.Config, error)
}

// Reloader can reload configuration at runtime.
type Reloader interface {
    // Reload reloads configuration from its original source.
    Reload() (*config.Config, error)
}
```

## Dependencies

- Depends on: `domain/config`
- Used by: `application/supervisor`
- Implemented by: `infrastructure/persistence/config/yaml`

## Related Packages

| Package | Role |
|---------|------|
| `domain/config` | Config entity returned by Loader |
| `infrastructure/persistence/config/yaml` | YAML implementation of Loader |
| `supervisor` | Uses Loader for configuration management |
