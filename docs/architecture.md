# Architecture

superviz.io is built using **Hexagonal Architecture** (also known as Ports & Adapters), ensuring clean separation of concerns and testability.

## Hexagonal Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                          │
│                           EXTERNAL WORLD                                 │
│                                                                          │
│    ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│    │   CLI    │  │   YAML   │  │    OS    │  │   File   │              │
│    │  (main)  │  │  Files   │  │ Signals  │  │  System  │              │
│    └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘              │
│         │             │             │             │                      │
│    ═════╪═════════════╪═════════════╪═════════════╪══════════════════   │
│         │             │             │             │                      │
│         │      ┌──────┴─────────────┴─────────────┴──────┐              │
│         │      │           INFRASTRUCTURE                 │              │
│         │      │                                          │              │
│         │      │  ┌─────────────┐  ┌─────────────┐       │              │
│         │      │  │ config/yaml │  │   kernel    │       │              │
│         │      │  │   Loader    │  │  Signals    │       │              │
│         │      │  └─────────────┘  │  Reaper     │       │              │
│         │      │                    │  Credentials│       │              │
│         │      │  ┌─────────────┐  └─────────────┘       │              │
│         │      │  │   health    │                         │              │
│         │      │  │  HTTP/TCP   │  ┌─────────────┐       │              │
│         │      │  │  Command    │  │   logging   │       │              │
│         │      │  └─────────────┘  │   Writers   │       │              │
│         │      │                    │   Rotation  │       │              │
│         │      │  ┌─────────────┐  └─────────────┘       │              │
│         │      │  │   process   │                         │              │
│         │      │  │  Executor   │                         │              │
│         │      │  └─────────────┘                         │              │
│         │      └──────────────────────────────────────────┘              │
│         │                          │                                     │
│         │                          │ implements                          │
│         │                          ▼                                     │
│         │      ┌──────────────────────────────────────────┐              │
│         │      │              DOMAIN                       │              │
│         │      │                                           │              │
│         │      │  ┌─────────┐ ┌─────────┐ ┌─────────┐    │              │
│         │      │  │ service │ │ process │ │ health  │    │              │
│         │      │  │ Config  │ │  Spec   │ │ Status  │    │              │
│         │      │  │         │ │  State  │ │ Result  │    │              │
│         │      │  │         │ │ Executor│ │ Checker │    │              │
│         │      │  │         │ │  (port) │ │  (port) │    │              │
│         │      │  └─────────┘ └─────────┘ └─────────┘    │              │
│         │      │                                           │              │
│         │      │  ┌─────────────────────────────────┐     │              │
│         │      │  │             shared              │     │              │
│         │      │  │   Duration    Size    Errors    │     │              │
│         │      │  └─────────────────────────────────┘     │              │
│         │      └──────────────────────────────────────────┘              │
│         │                          ▲                                     │
│         │                          │ uses                                │
│         │                          │                                     │
│         │      ┌──────────────────────────────────────────┐              │
│         └─────▶│            APPLICATION                    │              │
│                │                                           │              │
│                │  ┌───────────────────────────────────┐   │              │
│                │  │           Supervisor              │   │              │
│                │  │   • Orchestrates all services     │   │              │
│                │  │   • Handles signals               │   │              │
│                │  │   • Manages lifecycle             │   │              │
│                │  └───────────────────────────────────┘   │              │
│                │                    │                      │              │
│                │         ┌──────────┴──────────┐          │              │
│                │         ▼                     ▼          │              │
│                │  ┌─────────────┐      ┌─────────────┐   │              │
│                │  │ProcessManager│     │HealthMonitor│   │              │
│                │  │             │      │             │   │              │
│                │  │ • Start/Stop│      │ • Schedule  │   │              │
│                │  │ • Restart   │      │ • Execute   │   │              │
│                │  │ • Monitor   │      │ • Report    │   │              │
│                │  └─────────────┘      └─────────────┘   │              │
│                │                                           │              │
│                │  ┌─────────────────────────────────┐     │              │
│                │  │        config (ports)           │     │              │
│                │  │   Loader    Reloader            │     │              │
│                │  └─────────────────────────────────┘     │              │
│                └──────────────────────────────────────────┘              │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### Application Layer (`internal/application/`)

**Purpose**: Orchestration and use case implementation.

```mermaid
graph LR
    subgraph Application
        SUP[Supervisor]
        PM[ProcessManager]
        HM[HealthMonitor]
        CFG[config.Loader<br/>port]
    end

    SUP -->|manages| PM
    SUP -->|manages| HM
    SUP -->|uses| CFG
```

| Package | Role | Key Types |
|---------|------|-----------|
| `supervisor` | Service orchestration, signal handling | `Supervisor`, `ServiceStats` |
| `process` | Process lifecycle, restart logic | `ProcessManager` |
| `health` | Health check coordination | `HealthMonitor` |
| `config` | Configuration loading port | `Loader`, `Reloader` |

### Domain Layer (`internal/domain/`)

**Purpose**: Core business logic, entities, and port interfaces.

```mermaid
graph TB
    subgraph Domain
        subgraph service
            CFG[Config]
            SVC[ServiceConfig]
            HC[HealthCheckConfig]
        end

        subgraph process
            SPEC[Spec]
            STATE[State]
            EVENT[Event]
            EXEC[Executor<br/>port]
        end

        subgraph health
            STATUS[Status]
            RESULT[Result]
            CHECKER[Checker<br/>port]
        end

        subgraph shared
            DUR[Duration]
            SIZE[Size]
        end
    end

    CFG --> SVC
    SVC --> HC
    SVC --> SPEC
```

| Package | Role | Key Types |
|---------|------|-----------|
| `service` | Configuration entities | `Config`, `ServiceConfig`, `RestartConfig` |
| `process` | Process entities and ports | `Spec`, `State`, `Event`, `Executor` |
| `health` | Health entities | `Status`, `Result`, `Event`, `Checker` |
| `shared` | Value objects | `Duration`, `Size` |

### Infrastructure Layer (`internal/infrastructure/`)

**Purpose**: External system adapters implementing domain ports.

```mermaid
graph TB
    subgraph Infrastructure
        subgraph config/yaml
            YAML[Loader]
        end

        subgraph health
            HTTP[HTTPChecker]
            TCP[TCPChecker]
            CMD[CommandChecker]
        end

        subgraph process
            EXEC[UnixExecutor]
        end

        subgraph kernel
            SIG[SignalManager]
            REAP[Reaper]
            CRED[Credentials]
        end

        subgraph logging
            WRITER[Writer]
            CAPTURE[Capture]
            ROTATE[Rotation]
        end
    end
```

| Package | Role | Key Types |
|---------|------|-----------|
| `config/yaml` | YAML parsing | `Loader` |
| `health` | Health check implementations | `HTTPChecker`, `TCPChecker`, `CommandChecker` |
| `process` | Process execution | `UnixExecutor` |
| `kernel` | OS abstractions | `SignalManager`, `Reaper`, `Credentials` |
| `logging` | Log file management | `Writer`, `Capture`, `MultiWriter` |

## Dependency Rules

```mermaid
graph TB
    CMD[cmd/daemon] --> APP[Application]
    CMD --> INFRA[Infrastructure]

    APP --> DOMAIN[Domain]
    INFRA --> DOMAIN

    style DOMAIN fill:#f9f,stroke:#333
    style APP fill:#bbf,stroke:#333
    style INFRA fill:#bfb,stroke:#333
    style CMD fill:#fbb,stroke:#333
```

| Rule | Description |
|------|-------------|
| Domain is pure | No imports from application or infrastructure |
| Application uses domain | Imports domain entities and ports |
| Infrastructure implements domain | Adapters implement domain port interfaces |
| cmd is composition root | Wires infrastructure into application |

## Package Dependencies Diagram

```
┌────────────────────────────────────────────────────────────────┐
│                         cmd/daemon                              │
│                      (composition root)                         │
└────────────────────────────┬───────────────────────────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
┌─────────────────┐ ┌───────────────┐ ┌─────────────────┐
│   application   │ │    domain     │ │ infrastructure  │
│                 │ │               │ │                 │
│ • supervisor    │ │ • service     │ │ • config/yaml   │
│ • process       │ │ • process     │ │ • health        │
│ • health        │ │ • health      │ │ • process       │
│ • config        │ │ • shared      │ │ • kernel        │
│                 │ │               │ │ • logging       │
└────────┬────────┘ └───────────────┘ └────────┬────────┘
         │                  ▲                   │
         │                  │                   │
         └──────────────────┴───────────────────┘
                     depends on
```

## Port & Adapter Pattern

### Ports (Interfaces in Domain)

```go
// domain/process/port.go
type Executor interface {
    Start(spec Spec) (ProcessInfo, error)
    Stop(pid int, timeout time.Duration) error
    Signal(pid int, sig os.Signal) error
}

// domain/health/checker.go
type Checker interface {
    Check(ctx context.Context) Result
}
```

### Adapters (Implementations in Infrastructure)

```go
// infrastructure/process/executor.go
type UnixExecutor struct {
    // implements domain.Executor
}

// infrastructure/health/http.go
type HTTPChecker struct {
    // implements domain.Checker
}
```

## File Structure

```
src/internal/
├── application/                    # USE CASES
│   ├── config/
│   │   └── port.go                # Loader, Reloader interfaces
│   ├── health/
│   │   ├── monitor.go             # HealthMonitor
│   │   └── port.go                # CheckerFactory interface
│   ├── process/
│   │   ├── manager.go             # ProcessManager
│   │   └── signals.go             # Signal handling
│   └── supervisor/
│       ├── supervisor.go          # Main Supervisor
│       └── service_info.go        # ServiceInfo type
│
├── domain/                         # CORE BUSINESS LOGIC
│   ├── health/
│   │   ├── status.go              # Status enum
│   │   ├── result.go              # Result struct
│   │   └── event.go               # HealthEvent
│   ├── process/
│   │   ├── spec.go                # ProcessSpec
│   │   ├── state.go               # ProcessState enum
│   │   ├── event.go               # ProcessEvent
│   │   ├── port.go                # Executor interface
│   │   └── restart_policy.go      # RestartPolicy
│   ├── service/
│   │   ├── config.go              # Root Config
│   │   ├── serviceconfig.go       # ServiceConfig
│   │   ├── healthcheck.go         # HealthCheckConfig
│   │   ├── loggingconfig.go       # LoggingConfig
│   │   └── restart.go             # RestartConfig
│   └── shared/
│       ├── duration.go            # Duration value object
│       └── size.go                # Size value object
│
└── infrastructure/                 # ADAPTERS
    ├── config/yaml/
    │   ├── loader.go              # YAML Loader
    │   └── types.go               # DTO types
    ├── health/
    │   ├── http.go                # HTTPChecker
    │   ├── tcp.go                 # TCPChecker
    │   ├── command.go             # CommandChecker
    │   └── factory.go             # CheckerFactory
    ├── kernel/
    │   ├── adapters/
    │   │   ├── signals_unix.go    # Signal forwarding
    │   │   ├── reaper_unix.go     # Zombie reaping
    │   │   └── credentials_unix.go # User/Group
    │   └── ports/
    │       ├── signals.go         # SignalManager interface
    │       ├── reaper.go          # Reaper interface
    │       └── credentials.go     # Credentials interface
    ├── logging/
    │   ├── writer.go              # File writer with rotation
    │   ├── capture.go             # Stdout/stderr capture
    │   ├── multiwriter.go         # Multiple destinations
    │   └── linewriter.go          # Line-buffered writer
    └── process/
        └── executor.go            # UnixExecutor
```

## Benefits of This Architecture

| Benefit | How |
|---------|-----|
| **Testability** | Domain can be tested without infrastructure |
| **Flexibility** | Swap implementations (e.g., different config formats) |
| **Maintainability** | Clear boundaries between concerns |
| **Platform Independence** | OS specifics isolated in infrastructure/kernel |
