#!/bin/sh
# Build native NetBSD package
# Run on NetBSD VM via SSH from CI runner
set -e

VERSION="${1:-0.0.0}"
BINARY="${2:-/tmp/supervizio}"
STAGING="/tmp/pkg-staging"

echo "=== Building NetBSD package v${VERSION} ==="

# Clean staging
rm -rf "$STAGING"

# Create staging directory structure
mkdir -p "$STAGING/usr/local/bin"
mkdir -p "$STAGING/etc/rc.d"
mkdir -p "$STAGING/etc/supervizio"
mkdir -p "$STAGING/var/log/supervizio"

# Install files
cp "$BINARY" "$STAGING/usr/local/bin/supervizio"
chmod 755 "$STAGING/usr/local/bin/supervizio"

# Copy rc.d script if available
if [ -f /tmp/setup/init/netbsd/supervizio ]; then
    cp /tmp/setup/init/netbsd/supervizio "$STAGING/etc/rc.d/supervizio"
    chmod 755 "$STAGING/etc/rc.d/supervizio"
fi

# Copy example config if available
if [ -f /tmp/setup/examples/config.yaml ] || [ -f /tmp/examples/config.yaml ]; then
    SRC="/tmp/setup/examples/config.yaml"
    [ -f /tmp/examples/config.yaml ] && SRC="/tmp/examples/config.yaml"
    cp "$SRC" "$STAGING/etc/supervizio/config.example.yaml"
    chmod 644 "$STAGING/etc/supervizio/config.example.yaml"
fi

# Build package using tar-based approach (most portable across NetBSD versions)
cd /tmp
tar czf "supervizio-${VERSION}.tgz" -C "$STAGING" .

echo "=== Package built ==="
ls -la /tmp/supervizio-*.tgz 2>/dev/null || echo "Package file not found"
