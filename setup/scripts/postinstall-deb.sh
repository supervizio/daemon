#!/bin/sh
# Post-installation script for Debian/Devuan packages
set -e

# Create supervizio user if it doesn't exist
if ! getent passwd supervizio >/dev/null 2>&1; then
    useradd --system --no-create-home --shell /usr/sbin/nologin supervizio
fi

# Ensure directories exist with correct permissions
mkdir -p /etc/supervizio /var/log/supervizio
chown supervizio:supervizio /var/log/supervizio

# Create default config from example if not exists
if [ ! -f /etc/supervizio/config.yaml ] && [ -f /etc/supervizio/config.example.yaml ]; then
    cp /etc/supervizio/config.example.yaml /etc/supervizio/config.yaml
    chmod 644 /etc/supervizio/config.yaml
fi

# Reload systemd if available
if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload || true
fi

echo "supervizio installed successfully"
echo "Configure: /etc/supervizio/config.yaml"
echo "Start: systemctl start supervizio"
