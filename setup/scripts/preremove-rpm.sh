#!/bin/sh
# Pre-removal script for RPM packages
set -e

# Stop service if running
if command -v systemctl >/dev/null 2>&1; then
    systemctl stop supervizio 2>/dev/null || true
    systemctl disable supervizio 2>/dev/null || true
fi
