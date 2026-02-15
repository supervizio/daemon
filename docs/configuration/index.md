# Configuration

superviz.io uses YAML configuration files to define supervised services, monitoring targets, and daemon behavior.

---

## Configuration File

Default location: `/etc/supervizio/config.yaml`

Override with: `--config <path>`

```yaml
version: "1"

logging:
  base_dir: /var/log/supervizio
  defaults:
    timestamp_format: iso8601
    rotation:
      max_size: "10MB"
      max_files: 5

services:
  - name: my-app
    command: /usr/local/bin/my-app
    restart:
      policy: on-failure
      max_retries: 5
      delay: 5s

monitoring:
  defaults:
    interval: 30s
    timeout: 5s
```

---

## Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | `string` | Yes | Configuration format version (`"1"`) |
| `logging` | `object` | No | [Logging configuration](#logging) |
| `services` | `list` | No | [Service definitions](services.md) |
| `monitoring` | `object` | No | [Monitoring configuration](monitoring.md) |

---

## Logging

```yaml
logging:
  level: info           # debug, info, warn, error
  format: json          # json, text
  output: stdout        # stdout, stderr, file
  base_dir: /var/log/supervizio
  defaults:
    timestamp_format: iso8601  # iso8601, unix, rfc3339
    rotation:
      max_size: "10MB"
      max_files: 5
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `level` | `string` | `info` | Log level: debug, info, warn, error |
| `format` | `string` | `json` | Output format: json, text |
| `output` | `string` | `stdout` | Output target: stdout, stderr, file |
| `base_dir` | `string` | `/var/log/supervizio` | Base directory for log files |
| `defaults.timestamp_format` | `string` | `iso8601` | Timestamp format |
| `defaults.rotation.max_size` | `string` | `10MB` | Max log file size before rotation |
| `defaults.rotation.max_files` | `int` | `5` | Max number of rotated files |

---

## Configuration Reload

The daemon supports live configuration reload via `SIGHUP`:

```bash
kill -HUP $(pidof supervizio)
```

This triggers the `Reloader` port interface, which re-reads the YAML file and applies changes to service definitions and monitoring configuration without restarting the daemon.
