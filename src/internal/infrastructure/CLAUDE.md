# Infrastructure Layer

Adapters implementing domain ports for external systems.

## Structure

```
infrastructure/
├── config/
│   └── yaml/     # YAML configuration loader
├── health/       # Health check adapters
└── process/      # Process executor adapter
```

## Packages

| Package | Role |
|---------|------|
| `config/yaml` | Loads and parses YAML config files |
| `health` | HTTP, TCP, Command health checkers |
| `process` | Unix process execution |

## Dependencies

- Depends on: `domain`, `kernel`
- Implements: domain port interfaces
- Never imported by: `domain`

## Key Types

### config/yaml
- `Loader` - YAML file loader
- `ConfigDTO` - YAML data transfer objects
- `Duration` - YAML duration parsing

### health
- `HTTPChecker` - HTTP endpoint checks
- `TCPChecker` - TCP port connectivity
- `CommandChecker` - Command execution checks
- `Factory` - Creates checkers by type

### process
- `UnixExecutor` - Implements domain.Executor
