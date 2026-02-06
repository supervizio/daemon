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

# All platforms - using actual JSON structure
check_field_numeric ".cpu.usage_percent"
check_field_not_null ".cpu.cores"
check_field_numeric ".cpu.frequency_mhz"

# Note: The probe outputs a single usage_percent rather than user/system/idle breakdown
# This is by design for cross-platform compatibility

# =============================================================================
# SECTION 3: Memory Metrics
# =============================================================================
printf "\n${BLUE}=== Section 3: Memory Metrics ===${NC}\n"

# All platforms
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

# =============================================================================
# SECTION 4: PSI (Pressure Stall Information) - Linux 4.20+ only
# =============================================================================
printf "\n${BLUE}=== Section 4: PSI Metrics ===${NC}\n"

case "$PLATFORM" in
    linux)
        # PSI is embedded in cpu/memory/io pressure sub-objects
        # Check if pressure data exists (optional - kernel may not support it)
        if echo "$JSON" | jq -e '.cpu.pressure != null' >/dev/null 2>&1; then
            check_field_numeric ".cpu.pressure.some_avg10"
            printf "${GREEN}[OK]${NC} CPU pressure metrics present\n"
            TOTAL=$((TOTAL + 1))
            PASSED=$((PASSED + 1))
        else
            printf "${YELLOW}[INFO]${NC} CPU pressure not available (may be kernel limitation)\n"
        fi
        if echo "$JSON" | jq -e '.memory.pressure != null' >/dev/null 2>&1; then
            check_field_numeric ".memory.pressure.some_avg10"
            printf "${GREEN}[OK]${NC} Memory pressure metrics present\n"
            TOTAL=$((TOTAL + 1))
            PASSED=$((PASSED + 1))
        else
            printf "${YELLOW}[INFO]${NC} Memory pressure not available (may be kernel limitation)\n"
        fi
        if echo "$JSON" | jq -e '.io.pressure != null' >/dev/null 2>&1; then
            check_field_numeric ".io.pressure.some_avg10"
            printf "${GREEN}[OK]${NC} I/O pressure metrics present\n"
            TOTAL=$((TOTAL + 1))
            PASSED=$((PASSED + 1))
        else
            printf "${YELLOW}[INFO]${NC} I/O pressure not available (may be kernel limitation)\n"
        fi
        ;;
    *)
        skip_field ".*.pressure" "Linux kernel 4.20+ feature"
        ;;
esac

# =============================================================================
# SECTION 5: Load Average
# =============================================================================
printf "\n${BLUE}=== Section 5: Load Average ===${NC}\n"

check_field_numeric ".load.load_1min"
check_field_numeric ".load.load_5min"
check_field_numeric ".load.load_15min"

# =============================================================================
# SECTION 6: Process Metrics
# =============================================================================
printf "\n${BLUE}=== Section 6: Process Metrics ===${NC}\n"

check_field_not_null ".process.current_pid"
# Process metrics are in top_processes array
if echo "$JSON" | jq -e '.process.top_processes | length > 0' >/dev/null 2>&1; then
    check_field_numeric ".process.top_processes[0].memory_rss_bytes"
    check_field_numeric ".process.top_processes[0].memory_vms_bytes"
else
    printf "${YELLOW}[WARN]${NC} No process info in top_processes (may be expected)\n"
fi

# =============================================================================
# SECTION 7: Disk Metrics
# =============================================================================
printf "\n${BLUE}=== Section 7: Disk Metrics ===${NC}\n"

# Disk usage is under .disk.usage array
check_field ".disk"
check_array_not_empty ".disk.usage"

# Verify first disk has expected fields
check_field_not_null ".disk.usage[0].path"
check_field_numeric ".disk.usage[0].total_bytes"
check_field_numeric ".disk.usage[0].used_bytes"
check_field_numeric ".disk.usage[0].free_bytes"

# =============================================================================
# SECTION 8: Disk I/O Stats
# =============================================================================
printf "\n${BLUE}=== Section 8: Disk I/O Stats ===${NC}\n"

case "$PLATFORM" in
    linux|freebsd|openbsd|netbsd|darwin)
        # Disk I/O is under .disk.io array
        if echo "$JSON" | jq -e '.disk.io != null and (.disk.io | type) == "array"' >/dev/null 2>&1; then
            printf "${GREEN}[OK]${NC} .disk.io array present\n"
            TOTAL=$((TOTAL + 1))
            PASSED=$((PASSED + 1))

            # Check if we have at least one device with I/O stats
            if echo "$JSON" | jq -e '.disk.io | length > 0' >/dev/null 2>&1; then
                check_field_not_null ".disk.io[0].device"
                check_field_numeric ".disk.io[0].reads_completed"
                check_field_numeric ".disk.io[0].writes_completed"
            else
                printf "${YELLOW}[WARN]${NC} disk.io array is empty (may be expected on some VMs)\n"
            fi
        else
            printf "${YELLOW}[WARN]${NC} .disk.io not present or not array\n"
        fi
        ;;
esac

# =============================================================================
# SECTION 9: Network Interfaces
# =============================================================================
printf "\n${BLUE}=== Section 9: Network Interfaces ===${NC}\n"

check_field ".network.interfaces"
check_array_not_empty ".network.interfaces"

# Check first interface has expected fields
check_field_not_null ".network.interfaces[0].name"

# =============================================================================
# SECTION 10: Network Stats
# =============================================================================
printf "\n${BLUE}=== Section 10: Network Stats ===${NC}\n"

check_field ".network.stats"
check_array_not_empty ".network.stats"

# Check first stat has expected fields (using actual field names)
check_field_not_null ".network.stats[0].interface"
check_field_numeric ".network.stats[0].bytes_recv"
check_field_numeric ".network.stats[0].bytes_sent"
check_field_numeric ".network.stats[0].packets_recv"
check_field_numeric ".network.stats[0].packets_sent"

# =============================================================================
# SECTION 11: Network Connections
# =============================================================================
printf "\n${BLUE}=== Section 11: Network Connections ===${NC}\n"

case "$PLATFORM" in
    linux|freebsd|openbsd|netbsd|darwin)
        # TCP stats should exist
        check_field ".connections.tcp_stats"
        check_field_numeric ".connections.tcp_stats.established"
        check_field_numeric ".connections.tcp_stats.listen"
        check_field_numeric ".connections.tcp_stats.time_wait"

        # Connections arrays should exist (may be empty if no connections)
        check_field ".connections.tcp_connections"
        check_field ".connections.udp_sockets"
        ;;
esac

# =============================================================================
# SECTION 12: Aggregated I/O Stats
# =============================================================================
printf "\n${BLUE}=== Section 12: Aggregated I/O ===${NC}\n"

check_field_numeric ".io.read_bytes"
check_field_numeric ".io.write_bytes"
check_field_numeric ".io.read_ops"
check_field_numeric ".io.write_ops"

# =============================================================================
# SECTION 13: Context Switches
# =============================================================================
printf "\n${BLUE}=== Section 13: Context Switches ===${NC}\n"

check_field_numeric ".context_switches.system_total"
# Self context switches are nested under .self
if echo "$JSON" | jq -e '.context_switches.self != null' >/dev/null 2>&1; then
    check_field_numeric ".context_switches.self.voluntary"
    check_field_numeric ".context_switches.self.involuntary"
else
    printf "${YELLOW}[WARN]${NC} context_switches.self not present\n"
fi

# =============================================================================
# SECTION 14: Thermal Monitoring
# =============================================================================
printf "\n${BLUE}=== Section 14: Thermal Monitoring ===${NC}\n"

case "$PLATFORM" in
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

check_field_not_null ".container.is_containerized"
check_field_not_null ".runtime.is_containerized"

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
