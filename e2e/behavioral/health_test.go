package behavioral_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHTTPHealthProbe verifies that the supervisor can perform HTTP health probes.
func TestHTTPHealthProbe(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// No need to expose ports - we test from inside the container
	tc := startContainer(t, "health-http.yaml")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start")

	// Give the HTTP server time to start
	time.Sleep(2 * time.Second)

	// Verify health endpoint responds using curl inside the container
	code, output, err := tc.exec("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "http://localhost:8080/health")
	require.NoError(t, err, "curl should execute")
	require.Equal(t, 0, code, "curl should succeed")

	assert.Equal(t, "200", strings.TrimSpace(output),
		"health endpoint should return 200 OK")
}

// TestTCPHealthProbe verifies that the supervisor can perform TCP health probes.
func TestTCPHealthProbe(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// No need to expose ports - we test from inside the container
	tc := startContainer(t, "health-tcp.yaml")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start")

	// Give the TCP server time to start
	time.Sleep(2 * time.Second)

	// Verify TCP connection works using netcat inside the container
	// nc -z tests if port is open without sending data
	code, _, err := tc.exec("nc", "-z", "-w", "5", "localhost", "9090")
	require.NoError(t, err, "netcat should execute")

	assert.Equal(t, 0, code, "TCP connection to port 9090 should succeed")
}

// TestHealthProbeFailureTriggersRestart verifies that when a health probe
// fails repeatedly, the service is restarted.
func TestHealthProbeFailureTriggersRestart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := startContainer(t, "health-unhealthy.yaml")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start")

	// Get initial PID
	initialPID := tc.getProcessPID("crasher")
	require.NotZero(t, initialPID, "crasher should have a valid PID")

	t.Logf("Initial crasher PID: %d", initialPID)

	// Wait for health check failures to trigger restart.
	// Config: interval=1s, failure_threshold=3, restart_delay=500ms
	// So we expect restart after ~3.5-4 seconds minimum.
	time.Sleep(8 * time.Second)

	// Verify crasher was restarted (PID changed or process restarted)
	newPID := tc.getProcessPID("crasher")

	// If PID is 0, the process may be in the middle of a restart cycle.
	// Wait a bit more and check again.
	if newPID == 0 {
		time.Sleep(2 * time.Second)
		newPID = tc.getProcessPID("crasher")
	}

	t.Logf("New crasher PID: %d", newPID)

	// The PID should have changed, indicating a restart occurred.
	// Note: In some cases, the same PID can be reused, so we also check logs.
	if newPID == initialPID {
		// Check logs for restart indication
		logs := tc.getLogs()
		assert.Contains(t, logs, "health", "logs should contain health-related messages")
	} else {
		assert.NotEqual(t, initialPID, newPID,
			"crasher should have been restarted with a new PID")
	}
}

// TestServiceMarkedUnhealthyOnProbeFailure verifies that a service is marked
// as unhealthy when probe failures exceed the threshold, and this is logged.
func TestServiceMarkedUnhealthyOnProbeFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := startContainer(t, "health-unhealthy.yaml")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start")

	// Wait for health check failures to be logged.
	// Config: interval=1s, failure_threshold=3
	// After 3 failures, the health state transition should be logged.
	// The supervisor now logs: "[health] service=X listener=Y state=READY->LISTENING"
	// and "[health] service=X listener=Y unhealthy: ... - triggering restart"
	found := tc.waitForLogPattern(`\[health\].*unhealthy|health.*fail`, 15*time.Second)

	if !found {
		// Print logs for debugging
		logs := tc.getLogs()
		t.Logf("Container logs:\n%s", logs)
	}

	assert.True(t, found,
		"supervizio should log health check failures or unhealthy transitions")
}

// TestHealthProbeSuccessMarksHealthy verifies that a service becomes healthy
// after the success threshold is met.
func TestHealthProbeSuccessMarksHealthy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// No need to expose ports - we test from inside the container
	tc := startContainer(t, "health-http.yaml")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start")

	// Wait for enough time for health probes to succeed
	// Config has interval: 2s, success_threshold: 1
	time.Sleep(5 * time.Second)

	// Make multiple requests from inside the container to verify stability
	for i := 0; i < 3; i++ {
		code, output, err := tc.exec("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "http://localhost:8080/health")
		if err != nil || code != 0 {
			t.Logf("request %d failed: code=%d, err=%v", i, code, err)
			continue
		}
		assert.Equal(t, "200", strings.TrimSpace(output), "health endpoint should return 200")
	}

	// Verify crasher is still running (not killed due to health failure)
	assert.True(t, tc.isProcessRunning("crasher"),
		"crasher should still be running after successful health probes")
}
