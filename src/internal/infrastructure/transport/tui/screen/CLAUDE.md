# Screen - Screen Renderers

Full-screen section renderers for TUI.

## Structure

```
screen/
├── header.go        # Header section (title, hostname, version)
├── services.go      # Services section
├── system.go        # System metrics (CPU, RAM, disk)
├── network.go       # Network statistics
├── logs.go          # Logs section
├── context.go       # Context/environment info
└── service_entry.go # Single service row rendering
```

## Key Types

| Type | Description |
|------|-------------|
| `HeaderRenderer` | Renders header with branding |
| `ServicesRenderer` | Renders service table |
| `SystemRenderer` | Renders system metrics bars |
| `NetworkRenderer` | Renders network I/O stats |
| `LogsRenderer` | Renders log panel |

## Render Flow

```
Screen.Render(snapshot, layout)
    → Build sections based on layout
    → Combine into final output string
```

## Layout Adaptation

Each renderer adapts to available dimensions:
- Compact: Minimal info
- Normal: Full sections stacked
- Wide: Side-by-side panels

## Dependencies

- `widget/` for UI primitives
- `model/Snapshot` for data
- `layout/` for dimensions
