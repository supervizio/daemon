# Domain Listener Package

Domain entities for network listeners.

## Files

| File | Purpose |
|------|---------|
| `state.go` | `State` enum - listener states |
| `listener.go` | `Listener` entity - network listener |

## Key Types

### State (Enum)
```go
const (
    Closed    State = iota  // Port not open
    Listening               // Port open, accepting connections
    Ready                   // Health checks passed
)
```

**State Machine:**
```
CLOSED ──→ LISTENING ──→ READY
   ↑          │            │
   └──────────┴────────────┘ (probe fails)
```

### Listener (Entity)
```go
type Listener struct {
    Name        string
    Protocol    string         // "tcp", "udp"
    Address     string
    Port        int
    State       State
    ProbeConfig *probe.Config
    ProbeType   string
    ProbeTarget probe.Target
}
```

## Factory Functions

- `New(name, protocol, address, port)` - Create listener
- `NewTCP(name, address, port)` - Create TCP listener
- `NewUDP(name, address, port)` - Create UDP listener

## Methods

- `WithProbe(type, config, target)` - Add probe configuration
- `SetState(state)` - Transition state (validates transitions)
- `MarkListening()` - Transition to Listening
- `MarkReady()` - Transition to Ready
- `MarkClosed()` - Transition to Closed
- `HasProbe()` - Check if probe configured
- `GetProbeAddress()` - Get address for probing
