#!/bin/sh
# Post-installation script for RPM packages (Rocky/RHEL/Fedora)
set -e

# Create supervizio user if it doesn't exist
if ! getent passwd supervizio >/dev/null 2>&1; then
    useradd --system --no-create-home --shell /sbin/nologin supervizio
fi

# Ensure directories exist with correct permissions
mkdir -p /etc/supervizio /var/log/supervizio
chown supervizio:supervizio /var/log/supervizio

# Reload systemd
if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload || true
fi

echo "supervizio installed successfully"
echo "Configure: /etc/supervizio/config.yaml"
echo "Start: systemctl start supervizio"
