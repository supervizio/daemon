# Widget - UI Components

Reusable TUI building blocks for rendering.

## Structure

```
widget/
├── box.go           # Box drawing with borders and styles
├── bar.go           # Progress bars (CPU, memory, disk)
├── table.go         # Dynamic tables with columns
├── text.go          # Text rendering with truncation
├── status.go        # Status indicators (running, stopped)
├── spark_line.go    # Mini sparkline charts
├── progress_bar.go  # Percentage-based progress bars
├── sanitize.go      # ANSI escape sequence stripping (security)
├── truncate_state.go# Truncation with state preservation
├── column.go        # Column definition for tables
└── *_style.go       # Style configurations
```

## Key Types

| Type | Description |
|------|-------------|
| `Box` | Bordered container with title |
| `Table` | Dynamic-width column table |
| `ProgressBar` | Percentage bar with gradient |
| `SparkLine` | History graph in single line |
| `StripANSI()` | Security: removes ANSI escape codes |

## Security

`sanitize.go` provides `StripANSI()` to prevent terminal escape injection attacks.
Service names and log messages are sanitized before display.

## Testing

- `*_external_test.go`: Public API tests
- `*_internal_test.go`: Internal implementation tests
