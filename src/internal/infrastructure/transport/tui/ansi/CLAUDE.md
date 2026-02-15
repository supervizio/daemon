<!-- updated: 2026-02-15T21:30:00Z -->
# ANSI - Terminal Escape Sequences

ANSI escape codes for terminal styling and status icons.

## Files

| File | Purpose |
|------|---------|
| `codes.go` | ANSI escape sequences (colors, cursor, text attributes) |
| `status_icon.go` | Unicode/ASCII status icons for service states |
| `theme.go` | Color scheme definitions for TUI |

## Key Types

| Type | Description |
|------|-------------|
| `StatusIcon` | Icon set for running/stopped/failed/etc states |
| `Theme` | Color configuration (header, status, process colors) |

## Usage

```go
// Status icons
icon := ansi.StatusIcon{}.Running()  // "●" or "*"

// Theme colors
theme := ansi.DefaultTheme()
fmt.Print(theme.Header.Apply("Title"))
```

## Related

| Package | Relation |
|---------|----------|
| `tui/widget` | Uses Theme for styling |
| `tui/component` | Uses StatusIcon for process states |
