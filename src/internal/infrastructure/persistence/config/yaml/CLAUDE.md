<!-- updated: 2026-02-15T21:30:00Z -->
# YAML Config - Configuration Parser

Loading and parsing YAML configuration files.

## Role

Read a YAML file and convert it to domain `*config.Config`.

## Structure

| File | Role |
|------|------|
| `loader.go` | `Loader` with `Load(path)` |
| `types.go` | Intermediate YAML types |
| `metrics_dto.go` | DTO for metrics configuration mapping |

## Flow

```
config.yaml
    │
    ▼
yaml.Unmarshal() → types.go (YAMLConfig)
    │
    ▼
Mapping → domain/config.Config
```

## Intermediate Types

```go
// types.go
type YAMLConfig struct {
    Version  string        `yaml:"version"`
    Services []YAMLService `yaml:"services"`
}

type YAMLService struct {
    Name    string `yaml:"name"`
    Command string `yaml:"command"`
    // ...
}
```

## Constructor

```go
NewLoader() *Loader
```

## Usage

```go
loader := yaml.NewLoader()
cfg, err := loader.Load("/etc/daemon/config.yaml")
```

## Validation

The `Loader` validates:
- YAML syntax
- Required fields
- Acceptable values

Errors returned with context (line, field).
