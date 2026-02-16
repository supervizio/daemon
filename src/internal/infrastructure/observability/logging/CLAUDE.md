<!-- updated: 2026-02-15T21:30:00Z -->
# Logging - Log Management

Capture and formatting of supervised process outputs.

## Role

Capture process stdout/stderr, add timestamps, write to files with rotation.

## Components

| Type | File | Role |
|------|------|------|
| `Capture` | `capture.go` | Coordinates stdout/stderr |
| `LineWriter` | `linewriter.go` | Line-by-line buffering |
| `MultiWriter` | `multiwriter.go` | Writes to multiple destinations |
| `TimestampWriter` | `timestamp.go` | Adds timestamp prefix |
| `Writer` | `writer.go` | Base file writer |
| `FileOpener` | `fileopener.go` | Opens files (rotation-ready) |
| `NopCloser` | `nopcloser.go` | No-op Close() wrapper |

## Usage

```go
// Create a writer with timestamp
tw := logging.NewTimestampWriter(file, "2006-01-02 15:04:05")
lw := logging.NewLineWriter(tw)

// Attach to process
cmd.Stdout = lw
cmd.Stderr = lw
```

## Writer Chain

```
Process stdout/stderr
         │
         ▼
    LineWriter       ← Buffer until \n
         │
         ▼
  TimestampWriter    ← Adds [2024-01-15 10:30:45]
         │
         ▼
    MultiWriter      ← File + Console
         │
         ▼
      Writer         ← File with rotation
```

## YAML Configuration

```yaml
logging:
  base_dir: /var/log/supervizio
  defaults:
    timestamp_format: iso8601
    rotation:
      max_size: 100MB
      max_files: 10
```

## Constructors

```go
NewCapture(stdout, stderr io.Writer) *Capture
NewLineWriter(w io.Writer) *LineWriter
NewTimestampWriter(w io.Writer, format string) *TimestampWriter
NewMultiWriter(writers ...io.Writer) *MultiWriter
```
