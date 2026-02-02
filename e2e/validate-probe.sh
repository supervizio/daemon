#!/bin/sh
# shellcheck disable=SC2059
# SC2059: ANSI color codes (RED, GREEN, etc.) are safe constants, not user input.
# =============================================================================
# Validate supervizio --probe JSON output
# =============================================================================
# Comprehensive E2E validation ensuring probe returns all expected fields
# for each supported platform based on the comparison matrix.
#
# Expected coverage:
#   Linux:        100% (reference implementation)
#   FreeBSD:       98% (missing: PSI, iowait, steal, buffers)
#   macOS:         95% (missing: PSI, iowait, steal, buffers)
#   OpenBSD:       92% (missing: PSI, iowait, steal, buffers, temp_max/crit)
#   NetBSD:        90% (missing: PSI, iowait, steal, buffers, temp_max/crit)
#   DragonFlyBSD:   0% (stub only)
#
# Usage: validate-probe.sh
#
# Exit codes:
#   0 - All expected fields present for platform
#   1 - Validation failed (missing fields or invalid JSON)
#
# Note: Uses /bin/sh for maximum portability (BSD, busybox).
# Security: SAST-SAFE - No user input, all paths are hardcoded constants.
# =============================================================================
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Detect platform
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$PLATFORM" in
    dragonfly) PLATFORM="dragonfly" ;;
    *) ;;
esac

echo "=============================================="
echo " Probe Metrics E2E Validation"
echo "=============================================="
echo "Platform: $PLATFORM"
echo ""

# Collect probe output
echo "Collecting metrics via --probe..."
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

printf "${GREEN}[OK]${NC} Valid JSON structure\n"
echo ""

# Counters
TOTAL=0
PASSED=0
FAILED=0
SKIPPED=0

# Function to check if a field exists (path exists in JSON)
check_field() {
    FIELD="$1"
    TOTAL=$((TOTAL + 1))
    VALUE=$(echo "$JSON" | jq -e "$FIELD" 2>/dev/null)
    if [ $? -ne 0 ]; then
        printf "${RED}[FAIL]${NC} Missing: %s\n" "$FIELD"
        FAILED=$((FAILED + 1))
        return 1
    fi
    PASSED=$((PASSED + 1))
    return 0
}

# Function to check if a field exists and is not null
check_field_not_null() {
    FIELD="$1"
    TOTAL=$((TOTAL + 1))
    VALUE=$(echo "$JSON" | jq -e "$FIELD != null" 2>/dev/null)
    if [ "$VALUE" != "true" ]; then
        printf "${RED}[FAIL]${NC} Null value: %s\n" "$FIELD"
        FAILED=$((FAILED + 1))
        return 1
    fi
    PASSED=$((PASSED + 1))
    return 0
}

# Function to check if a field is a number >= 0
check_field_numeric() {
    FIELD="$1"
    TOTAL=$((TOTAL + 1))
    VALUE=$(echo "$JSON" | jq -e "($FIELD != null) and (($FIELD | type) == \"number\")" 2>/dev/null)
    if [ "$VALUE" != "true" ]; then
        printf "${RED}[FAIL]${NC} Invalid numeric: %s\n" "$FIELD"
        FAILED=$((FAILED + 1))
        return 1
    fi
    PASSED=$((PASSED + 1))
    return 0
}

# Function to check if array exists and is not empty
check_array_not_empty() {
    FIELD="$1"
    TOTAL=$((TOTAL + 1))
    VALUE=$(echo "$JSON" | jq -e "($FIELD | type) == \"array\" and ($FIELD | length) > 0" 2>/dev/null)
    if [ "$VALUE" != "true" ]; then
        printf "${RED}[FAIL]${NC} Empty or missing array: %s\n" "$FIELD"
        FAILED=$((FAILED + 1))
        return 1
    fi
    PASSED=$((PASSED + 1))
    return 0
}

# Function to skip a check with reason
skip_field() {
    FIELD="$1"
    REASON="$2"
    printf "${YELLOW}[SKIP]${NC} %s (%s)\n" "$FIELD" "$REASON"
    SKIPPED=$((SKIPPED + 1))
}

# =============================================================================
# SECTION 1: Metadata (all platforms)
# =============================================================================
printf "${BLUE}=== Section 1: Metadata ===${NC}\n"

check_field_not_null ".timestamp"
check_field_not_null ".platform"
check_field_not_null ".collected_at_ns"

# =============================================================================
# SECTION 2: CPU Metrics
# =============================================================================
printf "\n${BLUE}=== Section 2: CPU Metrics ===${NC}\n"

# All platforms
check_field_numeric ".cpu.user_percent"
check_field_numeric ".cpu.system_percent"
check_field_numeric ".cpu.idle_percent"
check_field_not_null ".cpu.cores"
check_field_numeric ".cpu.frequency_mhz"

# Linux-only fields
case "$PLATFORM" in
    linux)
        check_field_numeric ".cpu.iowait_percent"
        check_field_numeric ".cpu.steal_percent"
        ;;
    freebsd|openbsd|netbsd|darwin)
        skip_field ".cpu.iowait_percent" "Linux kernel scheduler concept"
        skip_field ".cpu.steal_percent" "Linux virtualization only"
        ;;
    dragonfly)
        skip_field ".cpu.*" "DragonFlyBSD not supported"
        ;;
esac

# =============================================================================
# SECTION 3: Memory Metrics
# =============================================================================
printf "\n${BLUE}=== Section 3: Memory Metrics ===${NC}\n"

# All platforms except DragonFlyBSD
case "$PLATFORM" in
    dragonfly)
        skip_field ".memory.*" "DragonFlyBSD not supported"
        ;;
    *)
        check_field_numeric ".memory.total_bytes"
        check_field_numeric ".memory.available_bytes"
        check_field_numeric ".memory.used_bytes"
        check_field_numeric ".memory.cached_bytes"
        check_field_numeric ".memory.swap_total_bytes"
        check_field_numeric ".memory.swap_used_bytes"
        check_field_numeric ".memory.used_percent"

        # buffers_bytes: Linux only
        case "$PLATFORM" in
            linux)
                check_field_numeric ".memory.buffers_bytes"
                ;;
            *)
                skip_field ".memory.buffers_bytes" "Linux page cache architecture only"
                ;;
        esac
        ;;
esac

# =============================================================================
# SECTION 4: PSI (Pressure Stall Information) - Linux 4.20+ only
# =============================================================================
printf "\n${BLUE}=== Section 4: PSI Metrics ===${NC}\n"

case "$PLATFORM" in
    linux)
        # PSI fields (may not be supported on older kernels)
        check_field_not_null ".cpu_pressure.supported" || true
        check_field_not_null ".memory_pressure.supported" || true
        check_field_not_null ".io_pressure.supported" || true
        ;;
    *)
        skip_field ".cpu_pressure" "Linux kernel 4.20+ feature"
        skip_field ".memory_pressure" "Linux kernel 4.20+ feature"
        skip_field ".io_pressure" "Linux kernel 4.20+ feature"
        ;;
esac

# =============================================================================
# SECTION 5: Load Average
# =============================================================================
printf "\n${BLUE}=== Section 5: Load Average ===${NC}\n"

case "$PLATFORM" in
    dragonfly)
        skip_field ".load.*" "DragonFlyBSD not supported"
        ;;
    *)
        check_field_numeric ".load.load_1min"
        check_field_numeric ".load.load_5min"
        check_field_numeric ".load.load_15min"
        ;;
esac

# =============================================================================
# SECTION 6: Process Metrics
# =============================================================================
printf "\n${BLUE}=== Section 6: Process Metrics ===${NC}\n"

case "$PLATFORM" in
    dragonfly)
        skip_field ".process.*" "DragonFlyBSD not supported"
        ;;
    *)
        check_field_not_null ".process.current_pid"
        check_field_numeric ".process.memory_rss_bytes"
        check_field_numeric ".process.memory_vms_bytes"
        check_field_not_null ".process.state"
        ;;
esac

# =============================================================================
# SECTION 7: Disk Metrics
# =============================================================================
printf "\n${BLUE}=== Section 7: Disk Metrics ===${NC}\n"

case "$PLATFORM" in
    dragonfly)
        skip_field ".disk.*" "DragonFlyBSD not supported"
        ;;
    *)
        # Disk partitions and usage
        check_field ".disk"
        check_array_not_empty ".disk"

        # Verify first disk has expected fields
        check_field_not_null ".disk[0].mount_point"
        check_field_numeric ".disk[0].total_bytes"
        check_field_numeric ".disk[0].used_bytes"
        check_field_numeric ".disk[0].available_bytes"
        ;;
esac

# =============================================================================
# SECTION 8: Disk I/O Stats
# =============================================================================
printf "\n${BLUE}=== Section 8: Disk I/O Stats ===${NC}\n"

case "$PLATFORM" in
    dragonfly)
        skip_field ".disk_io.*" "DragonFlyBSD not supported"
        ;;
    linux|freebsd|openbsd|netbsd|darwin)
        # Disk I/O should return stats
        if echo "$JSON" | jq -e '.disk_io != null and (.disk_io | type) == "array"' >/dev/null 2>&1; then
            printf "${GREEN}[OK]${NC} .disk_io array present\n"
            TOTAL=$((TOTAL + 1))
            PASSED=$((PASSED + 1))

            # Check if we have at least one device with I/O stats
            if echo "$JSON" | jq -e '.disk_io | length > 0' >/dev/null 2>&1; then
                check_field_not_null ".disk_io[0].device"
                check_field_numeric ".disk_io[0].reads_completed"
                check_field_numeric ".disk_io[0].writes_completed"
            else
                printf "${YELLOW}[WARN]${NC} disk_io array is empty (may be expected on some VMs)\n"
            fi
        else
            printf "${YELLOW}[WARN]${NC} .disk_io not present or not array\n"
        fi
        ;;
esac

# =============================================================================
# SECTION 9: Network Interfaces
# =============================================================================
printf "\n${BLUE}=== Section 9: Network Interfaces ===${NC}\n"

case "$PLATFORM" in
    dragonfly)
        skip_field ".network.*" "DragonFlyBSD not supported"
        ;;
    *)
        check_field ".network.interfaces"
        check_array_not_empty ".network.interfaces"

        # Check first interface has expected fields
        check_field_not_null ".network.interfaces[0].name"
        ;;
esac

# =============================================================================
# SECTION 10: Network Stats
# =============================================================================
printf "\n${BLUE}=== Section 10: Network Stats ===${NC}\n"

case "$PLATFORM" in
    dragonfly)
        skip_field ".network.stats.*" "DragonFlyBSD not supported"
        ;;
    *)
        check_field ".network.stats"
        check_array_not_empty ".network.stats"

        # Check first stat has expected fields
        check_field_not_null ".network.stats[0].interface"
        check_field_numeric ".network.stats[0].rx_bytes"
        check_field_numeric ".network.stats[0].tx_bytes"
        check_field_numeric ".network.stats[0].rx_packets"
        check_field_numeric ".network.stats[0].tx_packets"
        ;;
esac

# =============================================================================
# SECTION 11: Network Connections
# =============================================================================
printf "\n${BLUE}=== Section 11: Network Connections ===${NC}\n"

case "$PLATFORM" in
    dragonfly)
        skip_field ".connections.*" "DragonFlyBSD not supported"
        ;;
    linux|freebsd|openbsd|netbsd|darwin)
        # TCP stats should exist
        check_field ".connections.tcp_stats"
        check_field_numeric ".connections.tcp_stats.established"
        check_field_numeric ".connections.tcp_stats.listen"
        check_field_numeric ".connections.tcp_stats.time_wait"

        # Connections array should exist (may be empty if no connections)
        check_field ".connections.tcp"
        check_field ".connections.udp"
        ;;
esac

# =============================================================================
# SECTION 12: Aggregated I/O Stats
# =============================================================================
printf "\n${BLUE}=== Section 12: Aggregated I/O ===${NC}\n"

case "$PLATFORM" in
    dragonfly)
        skip_field ".io.*" "DragonFlyBSD not supported"
        ;;
    *)
        check_field_numeric ".io.read_bytes"
        check_field_numeric ".io.write_bytes"
        check_field_numeric ".io.read_ops"
        check_field_numeric ".io.write_ops"
        ;;
esac

# =============================================================================
# SECTION 13: Context Switches
# =============================================================================
printf "\n${BLUE}=== Section 13: Context Switches ===${NC}\n"

case "$PLATFORM" in
    dragonfly)
        skip_field ".context_switches.*" "DragonFlyBSD not supported"
        ;;
    *)
        check_field_numeric ".context_switches.system_total"
        check_field_numeric ".context_switches.voluntary"
        # involuntary may be 0 on macOS (not distinguished by Mach)
        check_field_numeric ".context_switches.involuntary"
        ;;
esac

# =============================================================================
# SECTION 14: Thermal Monitoring
# =============================================================================
printf "\n${BLUE}=== Section 14: Thermal Monitoring ===${NC}\n"

case "$PLATFORM" in
    dragonfly)
        skip_field ".thermal.*" "DragonFlyBSD not supported"
        ;;
    linux|freebsd|darwin)
        check_field_not_null ".thermal.supported"

        # If supported, check zones
        THERMAL_SUPPORTED=$(echo "$JSON" | jq -r '.thermal.supported // false')
        if [ "$THERMAL_SUPPORTED" = "true" ]; then
            check_field ".thermal.zones"
            # Check if we have at least one zone
            if echo "$JSON" | jq -e '.thermal.zones | length > 0' >/dev/null 2>&1; then
                check_field_not_null ".thermal.zones[0].name"
                check_field_numeric ".thermal.zones[0].temp_celsius"
            else
                printf "${YELLOW}[WARN]${NC} thermal.zones is empty (may be expected on VMs)\n"
            fi
        else
            printf "${YELLOW}[INFO]${NC} Thermal not supported on this system\n"
        fi
        ;;
    openbsd|netbsd)
        check_field_not_null ".thermal.supported"

        THERMAL_SUPPORTED=$(echo "$JSON" | jq -r '.thermal.supported // false')
        if [ "$THERMAL_SUPPORTED" = "true" ]; then
            check_field ".thermal.zones"
            # temp_max and temp_crit not available on OpenBSD/NetBSD
            skip_field ".thermal.zones[].temp_max" "OpenBSD/NetBSD hw.sensors limitation"
            skip_field ".thermal.zones[].temp_crit" "OpenBSD/NetBSD hw.sensors limitation"
        fi
        ;;
esac

# =============================================================================
# SECTION 15: Container/Runtime Detection
# =============================================================================
printf "\n${BLUE}=== Section 15: Container/Runtime Detection ===${NC}\n"

case "$PLATFORM" in
    dragonfly)
        skip_field ".container.*" "DragonFlyBSD not supported"
        skip_field ".runtime.*" "DragonFlyBSD not supported"
        ;;
    *)
        check_field_not_null ".container.is_containerized"
        check_field_not_null ".runtime.is_containerized"
        ;;
esac

# =============================================================================
# SECTION 16: Quota Information
# =============================================================================
printf "\n${BLUE}=== Section 16: Quota Information ===${NC}\n"

case "$PLATFORM" in
    linux)
        check_field_not_null ".quota.supported"
        ;;
    freebsd)
        # FreeBSD may have jail quotas
        check_field_not_null ".quota.supported" || skip_field ".quota" "Optional on FreeBSD"
        ;;
    *)
        skip_field ".quota.*" "Primarily Linux cgroups feature"
        ;;
esac

# =============================================================================
# RESULTS SUMMARY
# =============================================================================
echo ""
echo "=============================================="
printf " ${BLUE}RESULTS SUMMARY${NC}\n"
echo "=============================================="
echo ""
printf "Platform:  %s\n" "$PLATFORM"
printf "Total:     %d checks\n" "$TOTAL"
printf "Passed:    ${GREEN}%d${NC}\n" "$PASSED"
printf "Failed:    ${RED}%d${NC}\n" "$FAILED"
printf "Skipped:   ${YELLOW}%d${NC} (platform limitations)\n" "$SKIPPED"
echo ""

# Calculate coverage
if [ "$TOTAL" -gt 0 ]; then
    COVERAGE=$((PASSED * 100 / TOTAL))
else
    COVERAGE=0
fi

printf "Coverage:  %d%%\n" "$COVERAGE"
echo ""

# Expected coverage per platform
case "$PLATFORM" in
    linux)
        EXPECTED=95  # Allow some margin for kernel variations
        ;;
    freebsd)
        EXPECTED=90
        ;;
    darwin)
        EXPECTED=85
        ;;
    openbsd)
        EXPECTED=80
        ;;
    netbsd)
        EXPECTED=80
        ;;
    dragonfly)
        EXPECTED=0  # Stub only
        ;;
    *)
        EXPECTED=50
        ;;
esac

if [ "$FAILED" -gt 0 ]; then
    printf "${RED}[FAIL]${NC} %d field(s) missing or invalid\n" "$FAILED"
    echo ""
    echo "=== Full JSON output ==="
    echo "$JSON" | jq .
    exit 1
fi

if [ "$COVERAGE" -lt "$EXPECTED" ]; then
    printf "${YELLOW}[WARN]${NC} Coverage %d%% below expected %d%% for %s\n" "$COVERAGE" "$EXPECTED" "$PLATFORM"
fi

printf "${GREEN}[PASS]${NC} All expected fields present for platform=%s\n" "$PLATFORM"
exit 0
