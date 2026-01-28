# TUI - Terminal User Interface

Terminal user interface for superviz.io providing raw (MOTD) and interactive modes.

## Structure

```
tui/
├── tui.go             # Main entry, mode selection
├── raw.go             # Raw mode (static MOTD)
├── interactive.go     # Interactive mode (Bubble Tea)
├── adapter.go         # Supervisor/metrics adapters
├── ansi/              # ANSI codes and themes
├── collector/         # Data collectors (procfs, sysfs)
├── layout/            # Responsive layout
├── model/             # Data types (Snapshot, ServiceSnapshot)
├── screen/            # Screen renderers (header, services, system)
├── terminal/          # Terminal size detection (Linux/Darwin/BSD)
└── widget/            # UI components (box, bar, status, table)
```

## Modes

| Mode | Description | Flag |
|------|-------------|------|
| Raw | Static startup banner + log stream | Default |
| Interactive | Real-time TUI with 1Hz refresh | `--tui` |

**Raw mode default**: No `--raw` flag exists.

## Raw Mode Display

Static startup information only (no dynamic data that would become stale):
- Version, hostname, OS/arch, runtime mode, config path
- System metrics "at start" (CPU/RAM/Swap/Disk)
- Cgroup limits, detected sandboxes
- Service names with ports (plain text, no colors, from config only)

**NOT shown** (dynamic): Uptime, service states/PIDs, network rates.

## Key Types

| Type | Description |
|------|-------------|
| `TUI` | Main TUI struct |
| `Snapshot` | Complete state for display |
| `ServiceSnapshot` | Per-service state |
| `Collector` | Data collection interface |

## Data Flow

```
TUI.Run() → Collectors.CollectAll() → ServiceProvider.Services() → Renderer.Render()
```

## Constraints

- **No exec.Command**: All data from procfs/sysfs/syscalls
- **Pure Go**: No CGO dependencies
- **Cross-platform**: Linux priority, BSD/macOS best effort
- **Graceful degradation**: Missing data shows "-" or "unknown"

## Layout Breakpoints

| Width | Raw Mode | Interactive |
|-------|----------|-------------|
| <80 | Compact (header + services) | 1 column |
| 80-159 | Normal (stacked sections) | 1-2 columns |
| ≥160 | Wide (side-by-side) | 2-3 columns |

## Dependencies

- Depends on: `domain/process`, `domain/health`, `domain/metrics`
- Used by: `bootstrap` (CLI entry point)

## Related Directories

| Directory | See |
|-----------|-----|
| collector | `collector/CLAUDE.md` |
| component | `component/CLAUDE.md` |
| layout | `layout/CLAUDE.md` |
| model | `model/CLAUDE.md` |
| screen | `screen/CLAUDE.md` |
| widget | `widget/CLAUDE.md` |
