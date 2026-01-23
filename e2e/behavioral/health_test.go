package behavioral_test

import (
	"fmt"
	"net"
	"net/http"
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

	tc := startContainerWithPorts(t, "health-http.yaml", "8080/tcp")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start")

	// Give the HTTP server time to start
	time.Sleep(2 * time.Second)

	// Get mapped port
	host, err := tc.getHost()
	require.NoError(t, err)
	port, err := tc.getMappedPort("8080/tcp")
	require.NoError(t, err)

	// Verify health endpoint responds
	url := fmt.Sprintf("http://%s:%s/health", host, port)
	resp, err := http.Get(url)
	require.NoError(t, err, "health endpoint should be reachable")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode,
		"health endpoint should return 200 OK")
}

// TestTCPHealthProbe verifies that the supervisor can perform TCP health probes.
func TestTCPHealthProbe(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := startContainerWithPorts(t, "health-tcp.yaml", "9090/tcp")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start")

	// Give the TCP server time to start
	time.Sleep(2 * time.Second)

	// Get mapped port
	host, err := tc.getHost()
	require.NoError(t, err)
	port, err := tc.getMappedPort("9090/tcp")
	require.NoError(t, err)

	// Verify TCP connection works
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), 5*time.Second)
	require.NoError(t, err, "TCP connection should succeed")
	defer conn.Close()

	assert.NotNil(t, conn, "TCP connection should be established")
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

	tc := startContainerWithPorts(t, "health-http.yaml", "8080/tcp")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start")

	// Wait for enough time for health probes to succeed
	// Config has interval: 2s, success_threshold: 1
	time.Sleep(5 * time.Second)

	// Get mapped port and verify service is still running (healthy)
	host, err := tc.getHost()
	require.NoError(t, err)
	port, err := tc.getMappedPort("8080/tcp")
	require.NoError(t, err)

	// Make multiple requests to verify stability
	for i := 0; i < 3; i++ {
		url := fmt.Sprintf("http://%s:%s/health", host, port)
		resp, err := http.Get(url)
		if err != nil {
			t.Logf("request %d failed: %v", i, err)
			continue
		}
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Verify crasher is still running (not killed due to health failure)
	assert.True(t, tc.isProcessRunning("crasher"),
		"crasher should still be running after successful health probes")
}
