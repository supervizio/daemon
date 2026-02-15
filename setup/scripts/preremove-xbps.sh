#!/bin/sh
# Pre-removal script for Void Linux XBPS packages
set -e

# Stop runit service if running
if command -v sv >/dev/null 2>&1; then
    sv stop supervizio 2>/dev/null || true
    rm -f /var/service/supervizio
fi
