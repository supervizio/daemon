# Lib - Non-Go Libraries

External libraries used by the daemon, written in other languages.

## Structure

```
lib/
└── probe/    # Rust system metrics library
```

## Packages

| Directory | Language | Purpose |
|-----------|----------|---------|
| `probe/` | Rust | Cross-platform system metrics & quotas |

## Build

```bash
make build-probe    # Build Rust libraries
```

## Related

| Location | Description |
|----------|-------------|
| `lib/probe/` | Rust source code |
| `internal/infrastructure/probe/` | Go CGO bindings |
