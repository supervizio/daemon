#!/bin/sh
# Build native FreeBSD .pkg package
# Run on FreeBSD VM via SSH from CI runner
set -e

VERSION="${1:-0.0.0}"
BINARY="${2:-/tmp/supervizio}"
STAGING="/tmp/pkg-staging"

echo "=== Building FreeBSD package v${VERSION} ==="

# Clean staging
rm -rf "$STAGING"

# Create staging directory structure
mkdir -p "$STAGING/usr/local/bin"
mkdir -p "$STAGING/usr/local/etc/rc.d"
mkdir -p "$STAGING/usr/local/etc/supervizio"
mkdir -p "$STAGING/var/log/supervizio"

# Install files
cp "$BINARY" "$STAGING/usr/local/bin/supervizio"
chmod 755 "$STAGING/usr/local/bin/supervizio"

# Copy rc.d script if available
if [ -f /tmp/setup/init/freebsd/supervizio ]; then
    cp /tmp/setup/init/freebsd/supervizio "$STAGING/usr/local/etc/rc.d/supervizio"
    chmod 755 "$STAGING/usr/local/etc/rc.d/supervizio"
fi

# Copy example config if available
if [ -f /tmp/setup/examples/config.yaml ] || [ -f /tmp/examples/config.yaml ]; then
    SRC="/tmp/setup/examples/config.yaml"
    [ -f /tmp/examples/config.yaml ] && SRC="/tmp/examples/config.yaml"
    cp "$SRC" "$STAGING/usr/local/etc/supervizio/config.example.yaml"
    chmod 644 "$STAGING/usr/local/etc/supervizio/config.example.yaml"
fi

# Create +MANIFEST
cat > /tmp/+MANIFEST << EOF
name: supervizio
version: ${VERSION}
origin: sysutils/supervizio
comment: PID1-capable process supervisor
maintainer: noreply@superviz.io
www: https://github.com/supervizio/daemon
prefix: /usr/local
desc: PID1-capable process supervisor for containers and Unix systems
EOF

# Generate plist from staging (paths relative to prefix)
cd "$STAGING/usr/local"
find . -type f | sed 's|^\./||' | sort > /tmp/pkg-plist

echo "=== Plist ==="
cat /tmp/pkg-plist

# Build package (plist required for -M mode to include files)
cd /tmp
pkg create -M /tmp/+MANIFEST -p /tmp/pkg-plist -r "$STAGING" -o /tmp/

echo "=== Package built ==="
ls -la /tmp/supervizio-*.pkg 2>/dev/null || ls -la /tmp/supervizio-*.txz 2>/dev/null || echo "Package file not found"
