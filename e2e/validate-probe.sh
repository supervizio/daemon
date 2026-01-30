#!/bin/sh
# =============================================================================
# Validate supervizio --probe JSON output
# =============================================================================
# This script validates that the --probe command returns valid JSON with all
# expected fields for the current platform.
#
# Usage: validate-probe.sh
#
# Exit codes:
#   0 - All expected fields present
#   1 - Validation failed (missing fields or invalid JSON)
#
# Note: Uses /bin/sh for maximum portability (BSD, busybox).
# Security: SAST-SAFE - No user input, all paths are hardcoded constants.
# =============================================================================
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

# Detect platform
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$PLATFORM" in
    dragonfly) PLATFORM="dragonfly" ;;
    *) ;;
esac

echo "=== Probe Metrics Validation ==="
echo "Platform: $PLATFORM"

# Collect probe output
echo "Collecting metrics..."
JSON=$(/usr/local/bin/supervizio --probe 2>&1) || {
    printf "${RED}[FAIL]${NC} --probe command failed\n"
    echo "$JSON"
    exit 1
}

# Validate JSON is parseable
echo "$JSON" | jq empty 2>/dev/null || {
    printf "${RED}[FAIL]${NC} Invalid JSON output\n"
    echo "$JSON"
    exit 1
}

printf "${GREEN}[OK]${NC} Valid JSON\n"

# Function to check if a field exists (not null)
check_field() {
    FIELD="$1"
    VALUE=$(echo "$JSON" | jq -e "$FIELD" 2>/dev/null)
    if [ $? -ne 0 ]; then
        printf "${RED}[FAIL]${NC} Missing field: %s\n" "$FIELD"
        return 1
    fi
    return 0
}

# Function to check if a field exists and is not null
check_field_not_null() {
    FIELD="$1"
    VALUE=$(echo "$JSON" | jq -e "$FIELD != null" 2>/dev/null)
    if [ "$VALUE" != "true" ]; then
        printf "${RED}[FAIL]${NC} Field is null: %s\n" "$FIELD"
        return 1
    fi
    return 0
}

FAILED=0

echo ""
echo "=== Checking required fields (all platforms) ==="

# Metadata fields
for field in \
    ".timestamp" \
    ".platform" \
    ".collected_at_ns"
do
    check_field "$field" || FAILED=1
done

# CPU fields
for field in \
    ".cpu.usage_percent" \
    ".cpu.cores"
do
    check_field_not_null "$field" || FAILED=1
done

# Memory fields
for field in \
    ".memory.total_bytes" \
    ".memory.available_bytes" \
    ".memory.used_bytes" \
    ".memory.used_percent"
do
    check_field_not_null "$field" || FAILED=1
done

# Load fields
for field in \
    ".load.load_1min" \
    ".load.load_5min" \
    ".load.load_15min"
do
    check_field_not_null "$field" || FAILED=1
done

# Disk fields (check array exists and has items)
check_field ".disk" || FAILED=1

# Network fields (check arrays exist)
check_field ".network.interfaces" || FAILED=1
check_field ".network.stats" || FAILED=1

# I/O fields
for field in \
    ".io.read_bytes" \
    ".io.write_bytes"
do
    check_field_not_null "$field" || FAILED=1
done

# Process fields
check_field_not_null ".process.current_pid" || FAILED=1

# Container/Runtime fields
check_field_not_null ".container.is_containerized" || FAILED=1
check_field_not_null ".runtime.is_containerized" || FAILED=1

# Linux-specific fields
if [ "$PLATFORM" = "linux" ]; then
    echo ""
    echo "=== Checking Linux-specific fields ==="

    # Thermal (supported flag must exist, zones may be empty)
    check_field_not_null ".thermal.supported" || FAILED=1

    # Context switches
    check_field_not_null ".context_switches.system_total" || FAILED=1

    # Connections (tcp_stats must exist)
    check_field ".connections.tcp_stats" || FAILED=1

    # Quota (supported flag must exist)
    check_field_not_null ".quota.supported" || FAILED=1
fi

echo ""
if [ $FAILED -eq 1 ]; then
    printf "${RED}[FAIL]${NC} Probe validation failed\n"
    echo ""
    echo "=== Full JSON output ==="
    echo "$JSON" | jq .
    exit 1
fi

printf "${GREEN}[PASS]${NC} All expected fields present for platform=%s\n" "$PLATFORM"
exit 0
