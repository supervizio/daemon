# Transport - Network Communication

Servers and clients for external communication.

## Role

Expose APIs to control the daemon (start/stop services, metrics, status).

## Navigation

| Protocol | Package |
|----------|---------|
| gRPC | `grpc/` |
| TUI | `tui/` |

## Structure

```
transport/
├── grpc/              # gRPC API
│   └── server.go      # gRPC server
└── tui/               # Terminal User Interface
    ├── tui.go         # Main TUI entry
    ├── raw.go         # Static MOTD mode
    ├── interactive.go # Real-time TUI
    └── ...            # See tui/CLAUDE.md
```

## Services Exposed

| Service | Description |
|---------|-------------|
| `DaemonService` | Start, Stop, Restart, Status of services |
| `MetricsService` | Process metrics (CPU, RAM) |
| Health | gRPC health/v1 standard |

## TUI Modes

| Mode | Flag | Description |
|------|------|-------------|
| Raw | `--status` | Static MOTD snapshot, exits immediately |
| Interactive | `--tui` | Real-time TUI with 1Hz refresh (future) |

## Related Packages

| Package | See |
|---------|-----|
| gRPC | `grpc/CLAUDE.md` |
| TUI | `tui/CLAUDE.md` |
