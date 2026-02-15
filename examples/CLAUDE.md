<!-- updated: 2026-02-12T17:05:00Z -->
# Examples - Configuration Templates

YAML configuration examples for supervizio.

## Files

| File | Purpose |
|------|---------|
| `config.yaml` | Basic production configuration |
| `config-dev.yaml` | Development configuration with debug logging |
| `full-monitoring.yaml` | Complete monitoring setup (Docker, systemd, K8s, health checks) |

## Usage

```bash
# Copy and customize
cp examples/config.yaml /etc/supervizio/config.yaml

# Run with custom config
supervizio --config /etc/supervizio/config.yaml
```

## Related

| Directory | See |
|-----------|-----|
| `docs/configuration/` | Configuration reference |
| `setup/` | Installation scripts |
