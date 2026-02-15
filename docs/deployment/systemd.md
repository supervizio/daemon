# systemd Integration

superviz.io can run as a systemd service for managing application processes on Linux systems.

---

## Unit File

```ini
[Unit]
Description=superviz.io Process Supervisor
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/supervizio --config /etc/supervizio/config.yaml
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
RestartSec=5s

# Security hardening
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/log/supervizio /var/lib/supervizio

# Resource limits
LimitNOFILE=65535
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
```

---

## Installation

```bash
# Copy binary
sudo cp supervizio /usr/local/bin/

# Create config directory
sudo mkdir -p /etc/supervizio
sudo cp config.yaml /etc/supervizio/

# Create log directory
sudo mkdir -p /var/log/supervizio

# Install unit file
sudo cp supervizio.service /etc/systemd/system/
sudo systemctl daemon-reload

# Enable and start
sudo systemctl enable supervizio
sudo systemctl start supervizio
```

---

## Management

```bash
# Status
sudo systemctl status supervizio

# Logs
sudo journalctl -u supervizio -f

# Reload configuration (sends SIGHUP)
sudo systemctl reload supervizio

# Restart
sudo systemctl restart supervizio
```

---

## systemd Discovery

When running on a systemd-based system, superviz.io can discover and monitor other systemd services:

```yaml
monitoring:
  discovery:
    systemd:
      enabled: true
      patterns:
        - "nginx.service"
        - "postgresql*.service"
```

This allows superviz.io to monitor the health of services it doesn't directly manage, providing a unified view through the gRPC API and TUI.
