#!/bin/sh
# Post-installation script for Arch Linux / Artix packages
set -e

# Create supervizio user if it doesn't exist
if ! getent passwd supervizio >/dev/null 2>&1; then
    useradd --system --no-create-home --shell /usr/bin/nologin supervizio
fi

# Ensure directories exist with correct permissions
mkdir -p /etc/supervizio /var/log/supervizio
chown supervizio:supervizio /var/log/supervizio

# Create default config from example if not exists
if [ ! -f /etc/supervizio/config.yaml ] && [ -f /etc/supervizio/config.example.yaml ]; then
    cp /etc/supervizio/config.example.yaml /etc/supervizio/config.yaml
    chmod 644 /etc/supervizio/config.yaml
fi

# Detect init system and configure
if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
    # systemd (Arch, Manjaro)
    systemctl daemon-reload || true
    echo "supervizio installed successfully"
    echo "Configure: /etc/supervizio/config.yaml"
    echo "Start: systemctl start supervizio"
elif command -v dinitctl >/dev/null 2>&1; then
    # dinit (Artix)
    dinitctl enable supervizio 2>/dev/null || true
    echo "supervizio installed successfully"
    echo "Configure: /etc/supervizio/config.yaml"
    echo "Start: dinitctl start supervizio"
elif command -v s6-svc >/dev/null 2>&1 || [ -d /etc/s6 ]; then
    # s6 (Artix)
    echo "supervizio installed successfully"
    echo "Configure: /etc/supervizio/config.yaml"
    echo "Enable: s6-rc-bundle-update add default supervizio"
elif command -v sv >/dev/null 2>&1 && [ -d /etc/sv ]; then
    # runit (Artix)
    ln -sf /etc/sv/supervizio /var/service/ 2>/dev/null || true
    echo "supervizio installed successfully"
    echo "Configure: /etc/supervizio/config.yaml"
    echo "Start: sv start supervizio"
elif command -v rc-service >/dev/null 2>&1; then
    # OpenRC (Artix)
    rc-update add supervizio default 2>/dev/null || true
    echo "supervizio installed successfully"
    echo "Configure: /etc/supervizio/config.yaml"
    echo "Start: rc-service supervizio start"
else
    echo "supervizio installed successfully"
    echo "Configure: /etc/supervizio/config.yaml"
    echo "Note: No supported init system detected. Start manually."
fi
