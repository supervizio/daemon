# TUI - Terminal User Interface

Terminal user interface for superviz.io providing both raw (MOTD) and interactive modes.

## Structure

```
tui/
├── tui.go             # Main entry point, TUI struct, mode selection
├── raw.go             # Raw mode (static MOTD) renderer
├── interactive.go     # Interactive mode (Bubble Tea) renderer
├── adapter.go         # Adapters for supervisor/metrics integration
├── ansi/              # ANSI escape codes and theme
│   ├── codes.go       # Escape sequences
│   └── theme.go       # Color theme
├── collector/         # Data collectors (procfs, sysfs - no exec.Command)
│   ├── collector.go   # Collector interface and aggregator
│   ├── context.go     # Hostname, OS, kernel, runtime mode
│   ├── limits.go      # Cgroup limits interface
│   ├── limits_linux.go # Linux cgroup v1/v2 implementation
│   ├── network.go     # Network interface stats
│   ├── sandbox.go     # Container runtime detection
│   └── system.go      # CPU, memory, load average
├── layout/            # Responsive layout system
│   └── responsive.go  # Breakpoint calculations
├── model/             # Data types for display
│   └── snapshot.go    # Snapshot, RuntimeContext, ServiceSnapshot, etc.
├── screen/            # Screen renderers
│   ├── header.go      # superviz.io branding
│   ├── services.go    # Service table
│   ├── system.go      # CPU/RAM/Swap bars
│   ├── network.go     # Network interface table
│   ├── logs.go        # Log summary
│   └── context.go     # Context and sandboxes
├── terminal/          # Terminal utilities
│   ├── size.go        # Terminal size detection interface
│   ├── size_linux.go  # Linux ioctl implementation
│   ├── size_darwin.go # macOS implementation
│   └── size_bsd.go    # BSD implementation
└── widget/            # Reusable UI components
    ├── box.go         # Bordered boxes
    ├── bar.go         # Progress bars
    ├── status.go      # Status indicators
    ├── text.go        # Text formatting
    └── table.go       # Data tables
```

## Modes

| Mode | Description | Usage |
|------|-------------|-------|
| Raw | Static startup banner + log stream | Default (no flag needed) |
| Interactive | Real-time TUI with 1Hz refresh | `--tui` flag |

### Raw Mode Banner

The raw mode shows a **static startup snapshot** followed by a **live log stream**.
Only startup-time information is displayed (no dynamic metrics that would become stale):

```
╭──────────────────────────────────────────────────────────────────────────────╮
│  superviz.io v0.2.0                                                          │
│  Started: 2026-01-25T13:43:30Z                                               │
│  Host: hostname │ linux arm64 │ container (docker)                           │
│  Config: /etc/supervizio/config.yaml                                         │
╰──────────────────────────────────────────────────────────────────────────────╯
┌─ System (at start) ──────────────────────────────────────────────────────────┐
│  CPU   [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]   2%  Load: 2.33 1.80             │
│  RAM   [████████████░░░░░░░░░░░░░░░░░░░░░]  38%  4.4 GB / 11.7 GB            │
│  Swap  [██░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]   7%  137 MB / 2.0 GB             │
│  Disk  [█████████░░░░░░░░░░░░░░░░░░░░░░░░]  30%  75 GB / 251 GB              │
│  Limits: CPUSet: 0-7                                                         │
└──────────────────────────────────────────────────────────────────────────────┘
┌─ Sandboxes ──────────────────────────────────────────────────────────────────┐
│  ● Docker       /var/run/docker.sock                                         │
└──────────────────────────────────────────────────────────────────────────────┘
┌─ Services (3 configured) ────────────────────────────────────────────────────┐
│  sleeper              ticker              crasher                            │
└──────────────────────────────────────────────────────────────────────────────┘

2026-01-25T13:43:30Z [INFO] Supervisor started version=v0.2.0
2026-01-25T13:43:30Z [INFO] sleeper Service started pid=3220
...
```

**What's shown (static at startup):**
- Version, start timestamp, hostname, OS/arch, runtime mode
- Config file path
- System metrics "at start" (CPU/RAM/Swap/Disk)
- Cgroup limits
- Detected sandboxes (Docker, Podman, etc.)
- Service names with ports from config (plain text, no colors, no status verification)

**What's NOT shown (dynamic, would become stale):**
- Uptime (changes every second)
- Service states, PIDs, restarts
- Network RX/s, TX/s (requires delta)
- Logs summary (contradicts live stream)

## Key Types

| Type | Package | Description |
|------|---------|-------------|
| `TUI` | tui | Main TUI struct |
| `Snapshot` | model | Complete state for display |
| `ServiceSnapshot` | model | Per-service state |
| `SystemMetrics` | model | CPU/RAM/Swap metrics |
| `RuntimeContext` | model | Environment info |
| `Collector` | collector | Data collection interface |
| `Box` | widget | Bordered box component |
| `ProgressBar` | widget | Progress bar component |

## CLI Flags

| Flag | Description |
|------|-------------|
| `--tui` | Enable interactive TUI mode (real-time refresh) |

**Note**: Raw mode is the default. No `--raw` flag exists.

## Data Flow

```
TUI.Run()
    │
    ├── Collectors.CollectAll()     # Gather system data
    │   ├── ContextCollector        # Hostname, OS, runtime, config path
    │   ├── LimitsCollector         # Cgroup limits
    │   ├── SystemCollector         # CPU, RAM, Swap, Disk
    │   ├── NetworkCollector        # Interface stats (interactive only)
    │   └── SandboxCollector        # Container detection
    │
    ├── ServiceProvider.Services()  # Get service names
    │
    └── RawRenderer.Render()        # Output startup banner
        ├── renderHeader()          # Version, timestamp, host, config
        ├── SystemRenderer.RenderForRaw()  # System "at start"
        ├── renderSandboxes()       # Detected runtimes
        └── ServicesRenderer.RenderNamesOnly()  # Names in columns
```

## Constraints

- **No exec.Command**: All data from kernel interfaces (procfs, sysfs, syscalls)
- **Pure Go**: No CGO dependencies
- **Cross-platform**: Linux priority, BSD/macOS best effort
- **1Hz max refresh**: Avoid CPU overhead
- **Graceful degradation**: Missing data shows "-" or "unknown"

## Responsive Layout

### Raw Mode Layout

| Width | Layout | Behavior |
|-------|--------|----------|
| <80 | Compact | Header + Services names only |
| 80-159 | Normal | Header + System + Sandboxes (stacked) + Services |
| ≥160 | Wide | Header + System & Sandboxes (side-by-side) + Services |

### Interactive Mode Layout

| Width | Layout | Columns |
|-------|--------|---------|
| <80 | Compact | 1 |
| 80-120 | Normal | 1-2 |
| 120-160 | Wide | 2 |
| >160 | UltraWide | 2-3 |

### Terminal Size Detection

Size detection priority (standard Unix behavior):
1. `COLUMNS`/`LINES` environment variables (if explicitly set)
2. ioctl TIOCGWINSZ on stdout/stdin/stderr
3. Default fallback: 80x24

## Dependencies

- Depends on: `domain/process`, `domain/health`, `domain/metrics`
- Used by: `bootstrap` (CLI entry point)

## Related Packages

| Package | Role |
|---------|------|
| `application/supervisor` | Service data source |
| `application/metrics` | Process metrics source |
| `infrastructure/resources/metrics` | System metrics |
