# Examples - Configuration Samples

Example configurations for superviz.io.

## Structure

```
examples/
└── config.yaml    # Complete configuration
```

## config.yaml

Complete example showing all options:

- Global logging configuration
- Multiple services with different policies
- HTTP, TCP, and command health checks
- Per-service log rotation
- Environment variables
- User/group per service

## Usage

```bash
# Copy as base
cp examples/config.yaml /etc/supervizio/config.yaml

# Edit as needed
vim /etc/supervizio/config.yaml

# Start daemon
supervizio --config /etc/supervizio/config.yaml
```

## Main Sections

| Section | Description |
|---------|-------------|
| `version` | Config format version |
| `logging` | Global log config |
| `services` | List of services to manage |

## Related Directories

| Directory | Relation |
|-----------|----------|
| `../src/internal/config/` | Parses this format |
| `../` | README.md documents usage |
