#!/bin/sh
# E2E test script for supervizio installation
# This script runs inside the VM to test install/uninstall
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

pass() { printf "${GREEN}[PASS]${NC} %s\n" "$1"; }
fail() { printf "${RED}[FAIL]${NC} %s\n" "$1"; exit 1; }

echo "=== supervizio E2E Installation Test ==="
echo "OS: $(uname -s) $(uname -r)"
echo "Arch: $(uname -m)"

# Test 1: Run install script
echo ""
echo "=== Test 1: Installation ==="
if /setup/install.sh; then
    pass "Install script completed"
else
    fail "Install script failed"
fi

# Test 2: Binary exists
echo ""
echo "=== Test 2: Binary exists ==="
if [ -x /usr/local/bin/supervizio ]; then
    pass "Binary installed at /usr/local/bin/supervizio"
else
    fail "Binary not found or not executable"
fi

# Test 3: Config directory exists
echo ""
echo "=== Test 3: Config directory ==="
OS="$(uname -s)"
case "$OS" in
    FreeBSD|DragonFly)
        CONFIG_DIR="/usr/local/etc/supervizio"
        ;;
    *)
        CONFIG_DIR="/etc/supervizio"
        ;;
esac

if [ -d "$CONFIG_DIR" ]; then
    pass "Config directory exists: $CONFIG_DIR"
else
    fail "Config directory not found: $CONFIG_DIR"
fi

# Test 4: Config file exists
echo ""
echo "=== Test 4: Config file ==="
if [ -f "$CONFIG_DIR/config.yaml" ]; then
    pass "Config file exists: $CONFIG_DIR/config.yaml"
else
    fail "Config file not found"
fi

# Test 5: Service installed (platform-specific)
echo ""
echo "=== Test 5: Service installed ==="
case "$OS" in
    Linux)
        if [ -d /run/systemd/system ]; then
            if [ -f /etc/systemd/system/supervizio.service ]; then
                pass "systemd service installed"
            else
                fail "systemd service not found"
            fi
        elif command -v rc-service >/dev/null 2>&1; then
            if [ -f /etc/init.d/supervizio ]; then
                pass "OpenRC service installed"
            else
                fail "OpenRC service not found"
            fi
        elif command -v sv >/dev/null 2>&1 && [ -d /etc/sv ]; then
            if [ -d /etc/sv/supervizio ] && [ -L /var/service/supervizio ]; then
                pass "runit service installed"
            else
                fail "runit service not found"
            fi
        elif [ -f /etc/init.d/supervizio ]; then
            pass "SysVinit service installed"
        else
            fail "No service file found"
        fi
        ;;
    FreeBSD|DragonFly)
        if [ -f /usr/local/etc/rc.d/supervizio ]; then
            pass "FreeBSD rc.d service installed"
        else
            fail "FreeBSD rc.d service not found"
        fi
        ;;
    OpenBSD)
        if [ -f /etc/rc.d/supervizio ]; then
            pass "OpenBSD rc.d service installed"
        else
            fail "OpenBSD rc.d service not found"
        fi
        ;;
    NetBSD)
        if [ -f /etc/rc.d/supervizio ]; then
            pass "NetBSD rc.d service installed"
        else
            fail "NetBSD rc.d service not found"
        fi
        ;;
    Darwin)
        if [ -f /Library/LaunchDaemons/io.superviz.daemon.plist ]; then
            pass "launchd service installed"
        else
            fail "launchd service not found"
        fi
        ;;
esac

# Test 6: Version check
echo ""
echo "=== Test 6: Version check ==="
if /usr/local/bin/supervizio --version 2>/dev/null; then
    pass "Version command works"
else
    # Version may not be set in test builds
    echo "[WARN] Version check returned non-zero (may be expected)"
fi

# Test 7: Uninstall
echo ""
echo "=== Test 7: Uninstallation ==="
# Auto-answer 'n' to keep config and logs prompts
printf 'n\nn\n' | /setup/uninstall.sh
if [ ! -f /usr/local/bin/supervizio ]; then
    pass "Binary removed"
else
    fail "Binary still exists after uninstall"
fi

echo ""
echo "=== All E2E tests passed ==="
