# Domain Logging Package

Daemon event logging domain types and port interfaces.

## Files

| File | Purpose |
|------|---------|
| `level.go` | Level enum (Debug, Info, Warn, Error) |
| `event.go` | LogEvent entity |
| `writer.go` | Writer port interface |
| `logger.go` | Logger port interface |

## Key Types

### Level (Enum)

```go
const (
    LevelDebug Level = iota
    LevelInfo
    LevelWarn
    LevelError
)

func (l Level) String() string
func ParseLevel(s string) (Level, error)
```

### LogEvent (Entity)

```go
type LogEvent struct {
    Timestamp time.Time
    Level     Level
    Service   string          // Empty for daemon-level events
    EventType string          // "started", "stopped", "failed", etc.
    Message   string
    Metadata  map[string]any  // PID, ExitCode, Error, etc.
}

func NewLogEvent(level Level, service, eventType, message string) LogEvent
func (e LogEvent) WithMeta(key string, value any) LogEvent
func (e LogEvent) WithMetadata(meta map[string]any) LogEvent
```

### Writer (Port Interface)

```go
type Writer interface {
    Write(event LogEvent) error
    Close() error
}
```

### Logger (Port Interface)

```go
type Logger interface {
    Log(event LogEvent)
    Debug(service, eventType, message string, meta map[string]any)
    Info(service, eventType, message string, meta map[string]any)
    Warn(service, eventType, message string, meta map[string]any)
    Error(service, eventType, message string, meta map[string]any)
    Close() error
}
```

## Dependencies

- Depends on: nothing (pure domain)
- Used by: `infrastructure/observability/logging/daemon`

## Related Packages

| Package | Relation |
|---------|----------|
| `infrastructure/observability/logging/daemon` | Implements Writer and Logger |
| `application/supervisor` | Uses Logger for event logging |
