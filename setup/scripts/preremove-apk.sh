#!/bin/sh
# Pre-removal script for Alpine APK packages
set -e

# Stop service if running
if command -v rc-service >/dev/null 2>&1; then
    rc-service supervizio stop 2>/dev/null || true
    rc-update del supervizio 2>/dev/null || true
fi
