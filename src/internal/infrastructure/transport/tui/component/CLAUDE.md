# Component - High-Level TUI Components

Composite TUI panels combining widgets with data.

## Structure

```
component/
├── services.go        # Services panel (table of services)
├── service_columns.go # Column definitions for service table
└── logs.go            # Log viewer panel
```

## Key Types

| Type | Description |
|------|-------------|
| `ServicesPanel` | Service list with status, PID, ports |
| `LogsPanel` | Scrollable log viewer with filtering |

## ServicesPanel

Displays service table with columns:
- Name (sanitized), Status, PID, CPU%, Mem%, Uptime, Ports

Uses `widget.StripANSI()` for security.

## LogsPanel

Scrollable log viewer with:
- Service filtering
- Color-coded levels (error, warn, info)
- Message sanitization against escape injection

## Dependencies

- `widget/` for rendering primitives
- `model/` for ServiceSnapshot data
