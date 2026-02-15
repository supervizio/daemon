#!/bin/sh
# Pre-removal script for Arch Linux / Artix packages
set -e

# Stop and disable service based on init system
if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
    systemctl stop supervizio 2>/dev/null || true
    systemctl disable supervizio 2>/dev/null || true
elif command -v dinitctl >/dev/null 2>&1; then
    dinitctl stop supervizio 2>/dev/null || true
    dinitctl disable supervizio 2>/dev/null || true
fi
