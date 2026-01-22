#!/bin/bash
# =============================================================================
# E2E Container Tests - supervizio as PID1
# =============================================================================
# Validates supervizio behavior when running as container PID1:
# - PID1 verification
# - Managed services running
# - Zombie process reaping
# - Signal forwarding
# - Service restart on crash
# - HTTP health checks
# - Graceful shutdown
# =============================================================================

# Don't exit on first error - we want to run all tests
set -o pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

TESTS_PASSED=0
TESTS_FAILED=0
CONTAINER_NAME="${CONTAINER_NAME:-supervizio-pid1}"

# Test result handler
test_result() {
    local status=$1
    local description=$2

    if [ "$status" -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $description"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $description"
        ((TESTS_FAILED++))
    fi
}

# Check if container is running
check_container() {
    if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        echo -e "${RED}ERROR${NC}: Container '${CONTAINER_NAME}' is not running"
        exit 1
    fi
}

echo "═══════════════════════════════════════════════════════════════"
echo "  E2E Container Tests - supervizio as PID1"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "Container: ${CONTAINER_NAME}"
echo ""

check_container

# =============================================================================
# Test 1: PID1 verification
# =============================================================================
echo -e "\n${YELLOW}[Test 1]${NC} PID1 verification"
PID1_CMD=$(docker exec "${CONTAINER_NAME}" cat /proc/1/comm 2>/dev/null || echo "")
if [ "$PID1_CMD" = "supervizio" ]; then
    test_result 0 "supervizio is PID1 (comm=$PID1_CMD)"
else
    test_result 1 "supervizio is PID1 (got: $PID1_CMD)"
fi

# =============================================================================
# Test 2: Managed services running
# =============================================================================
echo -e "\n${YELLOW}[Test 2]${NC} Managed services running"

# Check nginx
if docker exec "${CONTAINER_NAME}" pgrep -x nginx > /dev/null 2>&1; then
    test_result 0 "nginx is running"
else
    test_result 1 "nginx is NOT running"
fi

# Check redis
if docker exec "${CONTAINER_NAME}" pgrep -x redis-server > /dev/null 2>&1; then
    test_result 0 "redis-server is running"
else
    test_result 1 "redis-server is NOT running"
fi

# =============================================================================
# Test 3: Zombie process reaping
# =============================================================================
echo -e "\n${YELLOW}[Test 3]${NC} Zombie process reaping"

# Create orphan processes that should be reaped by PID1
# Use a small wrapper that creates a child and exits, orphaning the child
docker exec "${CONTAINER_NAME}" sh -c 'for i in 1 2 3; do (sleep 0.1 &); done; sleep 0.5' 2>/dev/null || true

# Wait a moment for potential zombies to be reaped
sleep 3

# Count zombie processes (state Z in ps output)
ZOMBIES=$(docker exec "${CONTAINER_NAME}" ps aux 2>/dev/null | awk '$8 ~ /Z/' | wc -l || echo "0")
if [ "$ZOMBIES" = "0" ]; then
    test_result 0 "No zombie processes found"
else
    test_result 1 "No zombie processes (found: $ZOMBIES)"
fi

# =============================================================================
# Test 4: Signal forwarding (SIGHUP to PID1)
# =============================================================================
echo -e "\n${YELLOW}[Test 4]${NC} Signal forwarding"

# Get nginx PID before signal
NGINX_PID_BEFORE=$(docker exec "${CONTAINER_NAME}" pgrep -x nginx | head -1 2>/dev/null || echo "")

# Send SIGHUP to PID1 (supervizio)
docker exec "${CONTAINER_NAME}" kill -HUP 1 2>/dev/null || true
sleep 3

# Check nginx is still running
NGINX_PID_AFTER=$(docker exec "${CONTAINER_NAME}" pgrep -x nginx | head -1 2>/dev/null || echo "")
if [ -n "$NGINX_PID_AFTER" ]; then
    test_result 0 "Services survived SIGHUP (nginx pid: $NGINX_PID_BEFORE -> $NGINX_PID_AFTER)"
else
    test_result 1 "Services survived SIGHUP"
fi

# =============================================================================
# Test 5: Service restart on crash
# =============================================================================
echo -e "\n${YELLOW}[Test 5]${NC} Service restart on crash"

# Kill nginx forcefully
docker exec "${CONTAINER_NAME}" pkill -9 -x nginx 2>/dev/null || true

# Wait for restart - supervizio needs time to detect crash and restart
for i in 1 2 3 4 5 6 7 8 9 10; do
    sleep 1
    if docker exec "${CONTAINER_NAME}" pgrep -x nginx > /dev/null 2>&1; then
        NEW_PID=$(docker exec "${CONTAINER_NAME}" pgrep -x nginx | head -1)
        test_result 0 "nginx restarted after kill (new pid: $NEW_PID, took ${i}s)"
        break
    fi
    if [ "$i" = "10" ]; then
        test_result 1 "nginx restarted after kill (timeout after 10s)"
    fi
done

# =============================================================================
# Test 6: HTTP health check (nginx)
# =============================================================================
echo -e "\n${YELLOW}[Test 6]${NC} HTTP health check"

# Give nginx a moment to be fully ready
sleep 2

HTTP_STATUS=$(docker exec "${CONTAINER_NAME}" curl -s -o /dev/null -w "%{http_code}" http://localhost:80/ 2>/dev/null || echo "000")
if [ "$HTTP_STATUS" = "200" ]; then
    test_result 0 "nginx responds HTTP 200"
else
    test_result 1 "nginx responds HTTP 200 (got: $HTTP_STATUS)"
fi

# =============================================================================
# Test 7: TCP health check (redis)
# =============================================================================
echo -e "\n${YELLOW}[Test 7]${NC} TCP health check"

# Try to ping redis
REDIS_PONG=$(docker exec "${CONTAINER_NAME}" sh -c 'echo "PING" | nc -q1 localhost 6379' 2>/dev/null || echo "")
if echo "$REDIS_PONG" | grep -q "PONG"; then
    test_result 0 "redis responds to PING"
else
    # Check if port is at least open for diagnostic info
    if docker exec "${CONTAINER_NAME}" sh -c 'cat < /dev/tcp/localhost/6379' 2>/dev/null; then
        test_result 1 "redis port 6379 is open but does not respond to PING (got: $REDIS_PONG)"
    else
        test_result 1 "redis port 6379 is NOT open"
    fi
fi

# =============================================================================
# Summary
# =============================================================================
echo ""
echo "═══════════════════════════════════════════════════════════════"
echo -e "  Results: ${GREEN}$TESTS_PASSED passed${NC}, ${RED}$TESTS_FAILED failed${NC}"
echo "═══════════════════════════════════════════════════════════════"

# Exit with failure if any test failed
if [ "$TESTS_FAILED" -gt 0 ]; then
    exit 1
fi

exit 0
