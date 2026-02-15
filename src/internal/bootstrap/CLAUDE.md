# Bootstrap - Dependency Injection with Wire

Wire-based dependency injection for the superviz.io daemon.

## Structure

```
bootstrap/
├── app.go                          # App struct, Run(), signal handling
├── app_external_test.go            # Black-box tests for App
├── app_internal_test.go            # White-box tests for App
├── providers.go                    # Custom Wire providers
├── providers_external_test.go      # Providers tests
├── providers_internal_test.go      # Providers white-box tests
├── service_provider.go             # Service provider abstraction
├── service_provider_external_test.go
├── service_provider_internal_test.go
├── tui_mode_config.go              # TUI mode configuration
├── wire.go                         # Wire injector (build tag: wireinject)
└── wire_gen.go                     # Generated code (DO NOT EDIT)
```

## Key Types

### App
```go
type App struct {
    Supervisor *appsupervisor.Supervisor
    Cleanup    func()
}
```

### Providers

| Provider | Purpose |
|----------|---------|
| `ProvideReaper` | Returns ZombieReaper only if PID 1 |
| `LoadConfig` | Loads config from path via Loader |
| `NewApp` | Creates final App struct |

## Interface Bindings

| Interface | Implementation |
|-----------|----------------|
| `appconfig.Loader` | `*infraconfig.Loader` |
| `domain.Executor` | `*infraprocess.UnixExecutor` |
| `domainkernel.ZombieReaper` | `*reaper.UnixZombieReaper` (or nil) |

## Commands

```bash
wire ./internal/bootstrap/           # Generate wire_gen.go
go build ./cmd/daemon                # Build (uses wire_gen.go)
```

## Wire Build Tags

| File | Build Tag | When Used |
|------|-----------|-----------|
| `wire.go` | `//go:build wireinject` | Only by Wire tool |
| `wire_gen.go` | `//go:build !wireinject` | Normal builds |

## Usage

```go
// cmd/daemon/main.go
func main() {
    os.Exit(bootstrap.Run())
}
```

## Dependencies

- `github.com/google/wire` - Compile-time DI

## Related Packages

| Package | Relation |
|---------|----------|
| `application/supervisor` | Supervisor dependency |
| `infrastructure/process/executor` | Process executor |
| `infrastructure/persistence/config/yaml` | Config loader |
