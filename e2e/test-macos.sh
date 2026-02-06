#!/bin/sh
# =============================================================================
# macOS E2E test script for supervizio
# =============================================================================
# This script runs on macOS (GitHub Actions or local) to test:
# - Installation via setup/install.sh
# - Binary functionality
# - Probe metrics validation
#
# Usage: test-macos.sh [--skip-install]
#   --skip-install: Skip install.sh (use when installed via Homebrew)
#
# Note: Uses /bin/sh for maximum portability.
# Security: SAST-SAFE - No user input, all paths are hardcoded constants.
# =============================================================================
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

SKIP_INSTALL=false
for arg in "$@"; do
    case "$arg" in
        --skip-install) SKIP_INSTALL=true ;;
    esac
done

pass() { printf "${GREEN}[PASS]${NC} %s\n" "$1"; }
fail() { printf "${RED}[FAIL]${NC} %s\n" "$1"; exit 1; }
warn() { printf "${YELLOW}[WARN]${NC} %s\n" "$1"; }

echo "=== supervizio macOS E2E Test ==="
echo "OS: $(uname -s) $(uname -r)"
echo "Arch: $(uname -m)"
echo "macOS Version: $(sw_vers -productVersion 2>/dev/null || echo 'unknown')"
echo "Skip install: $SKIP_INSTALL"

# Verify we're on macOS
if [ "$(uname -s)" != "Darwin" ]; then
    fail "This script is for macOS only"
fi

# Test 1: Installation
echo ""
echo "=== Test 1: Installation ==="
if [ "$SKIP_INSTALL" = "true" ]; then
    pass "Skipped (installed via Homebrew)"
else
    SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
    if [ -x "$SCRIPT_DIR/setup/install.sh" ]; then
        if sudo "$SCRIPT_DIR/setup/install.sh"; then
            pass "Install script completed"
        else
            fail "Install script failed"
        fi
    else
        fail "Install script not found at $SCRIPT_DIR/setup/install.sh"
    fi
fi

# Test 2: Binary exists
echo ""
echo "=== Test 2: Binary exists ==="
if [ -x /usr/local/bin/supervizio ]; then
    pass "Binary installed at /usr/local/bin/supervizio"
else
    fail "Binary not found or not executable"
fi

# Test 3: Config directory
echo ""
echo "=== Test 3: Config directory ==="
if [ -d /etc/supervizio ]; then
    pass "Config directory exists: /etc/supervizio"
else
    fail "Config directory not found: /etc/supervizio"
fi

# Test 4: Config file
echo ""
echo "=== Test 4: Config file ==="
if [ -f /etc/supervizio/config.yaml ]; then
    pass "Config file exists: /etc/supervizio/config.yaml"
else
    fail "Config file not found"
fi

# Test 5: launchd service
echo ""
echo "=== Test 5: launchd service ==="
if [ -f /Library/LaunchDaemons/io.superviz.daemon.plist ]; then
    pass "launchd plist installed"
else
    fail "launchd plist not found at /Library/LaunchDaemons/io.superviz.daemon.plist"
fi

# Test 6: Version check
echo ""
echo "=== Test 6: Version check ==="
if /usr/local/bin/supervizio --version 2>/dev/null; then
    pass "Version command works"
else
    warn "Version check returned non-zero (may be expected in test builds)"
fi

# Test 7: Probe metrics validation
echo ""
echo "=== Test 7: Probe metrics validation ==="

# Check if jq is installed
if ! command -v jq >/dev/null 2>&1; then
    warn "jq not installed, skipping detailed probe validation"
    # Simple check: just verify --probe returns JSON
    if /usr/local/bin/supervizio --probe 2>&1 | head -c 1 | grep -q '{'; then
        pass "Probe returns JSON-like output"
    else
        fail "Probe does not return JSON"
    fi
else
    # Full validation
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
    if [ -x "$SCRIPT_DIR/validate-probe.sh" ]; then
        if "$SCRIPT_DIR/validate-probe.sh"; then
            pass "Probe metrics validation passed"
        else
            fail "Probe metrics validation failed"
        fi
    else
        warn "validate-probe.sh not found, running inline validation"

        # Inline macOS validation
        JSON=$(/usr/local/bin/supervizio --probe 2>&1)

        # Check JSON is valid
        if ! echo "$JSON" | jq empty 2>/dev/null; then
            fail "Invalid JSON from --probe"
        fi

        # Check critical macOS fields
        CHECKS_PASSED=0
        CHECKS_FAILED=0

        check_field() {
            if echo "$JSON" | jq -e "$1" >/dev/null 2>&1; then
                CHECKS_PASSED=$((CHECKS_PASSED + 1))
            else
                printf "${RED}[FAIL]${NC} Missing: %s\n" "$1"
                CHECKS_FAILED=$((CHECKS_FAILED + 1))
            fi
        }

        # Core fields that must exist on macOS
        check_field ".platform"
        check_field ".cpu.user_percent"
        check_field ".cpu.system_percent"
        check_field ".cpu.cores"
        check_field ".memory.total_bytes"
        check_field ".memory.available_bytes"
        check_field ".load.load_1min"
        check_field ".disk"
        check_field ".network.interfaces"
        check_field ".network.stats"
        check_field ".io.read_bytes"
        check_field ".thermal.supported"
        check_field ".connections.tcp_stats"
        check_field ".container.is_containerized"

        if [ "$CHECKS_FAILED" -gt 0 ]; then
            fail "Probe validation failed: $CHECKS_FAILED field(s) missing"
        fi

        pass "Probe metrics valid ($CHECKS_PASSED fields checked)"
    fi
fi

# Test 8: Uninstall (optional)
echo ""
echo "=== Test 8: Uninstallation ==="
if [ "$SKIP_INSTALL" = "true" ]; then
    pass "Skipped (uninstall via Homebrew)"
else
    SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
    if [ -x "$SCRIPT_DIR/setup/uninstall.sh" ]; then
        # Auto-answer 'n' to keep config and logs prompts
        printf 'n\nn\n' | sudo "$SCRIPT_DIR/setup/uninstall.sh"
        if [ ! -f /usr/local/bin/supervizio ]; then
            pass "Binary removed"
        else
            fail "Binary still exists after uninstall"
        fi
    else
        warn "Uninstall script not found, skipping"
    fi
fi

echo ""
echo "=== All macOS E2E tests passed ==="
