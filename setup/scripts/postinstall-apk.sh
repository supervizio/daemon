#!/bin/sh
# Post-installation script for Alpine APK packages
set -e

# Create supervizio user if it doesn't exist
if ! getent passwd supervizio >/dev/null 2>&1; then
    adduser -S -D -H -s /sbin/nologin -G nogroup supervizio
fi

# Ensure directories exist with correct permissions
mkdir -p /etc/supervizio /var/log/supervizio
chown supervizio:nogroup /var/log/supervizio

# Create default config from example if not exists
if [ ! -f /etc/supervizio/config.yaml ] && [ -f /etc/supervizio/config.example.yaml ]; then
    cp /etc/supervizio/config.example.yaml /etc/supervizio/config.yaml
    chmod 644 /etc/supervizio/config.yaml
fi

echo "supervizio installed successfully"
echo "Configure: /etc/supervizio/config.yaml"
echo "Start: rc-service supervizio start"
echo "Enable: rc-update add supervizio"
