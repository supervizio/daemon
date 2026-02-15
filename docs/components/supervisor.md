# Supervisor

The `Supervisor` is the top-level orchestrator in the application layer. It coordinates all service lifecycle managers, health monitors, metrics trackers, and external monitors.

**Package**: `internal/application/supervisor`

---

## Responsibilities

- Start and stop all configured services
- Coordinate lifecycle managers (one per service)
- Manage health monitoring across all services
- Track metrics for all processes
- Handle graceful shutdown with ordered cleanup

---

## Data Flow

```mermaid
graph TB
    SUP["Supervisor"]
    SUP --> LM1["Manager (service-1)"]
    SUP --> LM2["Manager (service-2)"]
    SUP --> LMN["Manager (service-N)"]
    SUP --> HM["ProbeMonitor"]
    SUP --> MT["Metrics Tracker"]
    SUP --> EM["External Monitor"]

    LM1 --> EX["Executor Port"]
    LM2 --> EX
    HM --> PR["Prober Port"]
    MT --> CO["Collector Port"]
    EM --> DI["Discoverer Port"]

    style SUP fill:#df41fb1a,stroke:#df41fb,color:#d4d8e0
    style LM1 fill:#df41fb1a,stroke:#df41fb,color:#d4d8e0
    style LM2 fill:#df41fb1a,stroke:#df41fb,color:#d4d8e0
    style LMN fill:#df41fb1a,stroke:#df41fb,color:#d4d8e0
    style HM fill:#df41fb1a,stroke:#df41fb,color:#d4d8e0
    style MT fill:#df41fb1a,stroke:#df41fb,color:#d4d8e0
    style EM fill:#df41fb1a,stroke:#df41fb,color:#d4d8e0
    style EX fill:#fbdf411a,stroke:#fbdf41,color:#d4d8e0
    style PR fill:#fbdf411a,stroke:#fbdf41,color:#d4d8e0
    style CO fill:#fbdf411a,stroke:#fbdf41,color:#d4d8e0
    style DI fill:#fbdf411a,stroke:#fbdf41,color:#d4d8e0
```

---

## Lifecycle

1. **Construction**: Wire DI creates Supervisor with all dependencies injected
2. **Start**: Iterates service configs, creates a `lifecycle.Manager` per service
3. **Running**: All managers run concurrently; health/metrics/monitoring loops active
4. **Shutdown**: Receives signal → stops all managers in reverse order → cleanup

---

## Key Interfaces

The Supervisor implements provider interfaces consumed by the gRPC server:

```go
type MetricsProvider interface {
    GetProcessMetrics(serviceName string) (metrics.ProcessMetrics, error)
    GetAllProcessMetrics() []metrics.ProcessMetrics
}

type StateProvider interface {
    GetState() state.DaemonState
}
```
