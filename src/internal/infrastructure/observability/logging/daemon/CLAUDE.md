# Daemon Logger Package

Infrastructure adapters for daemon event logging.

## Files

| File | Purpose |
|------|---------|
| `logger.go` | MultiLogger - aggregates multiple writers |
| `formatter.go` | TextFormatter for human-readable output |
| `writer_console.go` | ConsoleWriter - stdout/stderr split by level |
| `writer_file.go` | FileWriter - file output with rotation |
| `writer_json.go` | JSONWriter - structured JSON output |
| `level_filter.go` | LevelFilter - filters events by level |
| `factory.go` | BuildLogger - creates logger from config |

## Default Behavior

When no configuration is provided:
- Console only (no file logging)
- Base level: INFO (DEBUG filtered out)
- Output split by level:
  - DEBUG, INFO → stdout
  - WARN, ERROR → stderr
- Colors enabled if stdout is a TTY

## Writers

### ConsoleWriter (Default)

```go
// Splits output by level
cw := daemon.NewConsoleWriter()
// DEBUG/INFO → stdout
// WARN/ERROR → stderr
```

### FileWriter

```go
fw, err := daemon.NewFileWriter("/var/log/daemon.log", config.RotationConfig{
    MaxSize:  "50MB",
    MaxFiles: 5,
})
```

### JSONWriter

```go
jw, err := daemon.NewJSONWriter("/var/log/daemon.json", config.RotationConfig{})
// Output: {"ts":"...","level":"info","service":"nginx","event":"started","pid":1234}
```

### LevelFilter

```go
// Wrap any writer with level filtering
filtered := daemon.WithLevelFilter(writer, logging.LevelInfo)
```

## Factory

```go
// Build from config
logger, err := daemon.BuildLogger(cfg.Logging.Daemon, cfg.Logging.BaseDir)

// Or use default
logger := daemon.DefaultLogger()
```

## Usage

```go
logger.Info("nginx", "started", "Service started", map[string]any{"pid": 1234})
logger.Warn("nginx", "failed", "Service failed", map[string]any{"exit_code": 1})
```

## Dependencies

- Implements: `domain/logging.Logger`, `domain/logging.Writer`
- Uses: `domain/config.DaemonLogging`, `domain/config.RotationConfig`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/logging` | Port interfaces |
| `domain/config` | Configuration types |
| `application/supervisor` | Uses logger for events |
