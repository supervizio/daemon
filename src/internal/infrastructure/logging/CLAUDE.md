# Logging - Log Management

Log management utilities for process output capture and formatting.

## Structure

```
logging/
├── capture.go        # Output capture coordination
├── linewriter.go     # Line-buffered writing
├── multiwriter.go    # Multiple destination writer
├── timestamp.go      # Timestamp prefixing
├── writer.go         # Base log writer
├── fileopener.go     # File opening utilities
└── nopcloser.go      # No-op closer wrapper
```

## Components

| Type | Role |
|------|------|
| `Capture` | Coordinates stdout/stderr capture |
| `LineWriter` | Ensures complete lines before flush |
| `MultiWriter` | Writes to multiple destinations |
| `TimestampWriter` | Adds timestamp prefixes |
| `Writer` | Base writer implementation |
| `FileOpener` | Opens log files (rotation-ready) |
| `NopCloser` | Wraps writers without close |

## Usage

```go
import "github.com/kodflow/daemon/internal/logging"

// Create timestamped line writer
tw := logging.NewTimestampWriter(file, "2006-01-02 15:04:05")
lw := logging.NewLineWriter(tw)

// Attach to process
cmd.Stdout = lw
cmd.Stderr = lw
```

## Configuration

```yaml
logging:
  base_dir: /var/log/supervizio
  defaults:
    timestamp_format: iso8601
    rotation:
      max_size: 100MB
      max_age: 7d
      max_files: 10
      compress: true
```

## Rotation Triggers

| Trigger | Action |
|---------|--------|
| `max_size` | Rotate when file > size |
| `max_age` | Delete files > age |
| `max_files` | Keep N files max |
| `compress` | Gzip old files |

## Dependencies

- Depends on: nothing (utility layer)
- Used by: `application/process`

## Related Directories

| Directory | Relation |
|-----------|----------|
| `../application/process/` | Attaches writers to processes |
| `../domain/service/` | Receives LoggingConfig |
