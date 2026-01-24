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

# Enable service based on init system
if [ -d /run/systemd/system ]; then
    # systemd (Debian, Ubuntu, etc.)
    systemctl daemon-reload || true
    echo "supervizio installed successfully"
    echo "Configure: /etc/supervizio/config.yaml"
    echo "Start: systemctl start supervizio"
elif [ -d /etc/init.d ] && [ -f /etc/init.d/supervizio ]; then
    # SysVinit (Devuan, etc.)
    update-rc.d supervizio defaults 2>/dev/null || true
    echo "supervizio installed successfully"
    echo "Configure: /etc/supervizio/config.yaml"
    echo "Start: /etc/init.d/supervizio start"
else
    echo "supervizio installed successfully"
    echo "Configure: /etc/supervizio/config.yaml"
    echo "Note: No supported init system detected. Start manually."
fi
