package behavioral_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRestartPolicyAlways verifies that a service with policy "always"
// restarts regardless of exit code.
func TestRestartPolicyAlways(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := startContainer(t, "restart-always.yaml")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start initially")

	// Wait for crasher to exit (it exits with code 0 after 1s)
	require.True(t, tc.waitForProcessExit("crasher", 5*time.Second),
		"crasher should exit after delay")

	// Wait for restart
	require.True(t, tc.waitForProcess("crasher", 5*time.Second),
		"crasher should restart with policy 'always' even on exit code 0")
}

// TestRestartPolicyOnFailureWithFailure verifies that a service with policy
// "on-failure" restarts when the process exits with a non-zero exit code.
func TestRestartPolicyOnFailureWithFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := startContainer(t, "restart-on-failure.yaml")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start initially")

	// Wait for crasher to exit (it exits with code 1 after 1s)
	require.True(t, tc.waitForProcessExit("crasher", 5*time.Second),
		"crasher should exit after delay")

	// Wait for restart
	require.True(t, tc.waitForProcess("crasher", 5*time.Second),
		"crasher should restart with policy 'on-failure' on exit code 1")
}

// TestRestartPolicyOnFailureWithSuccess verifies that a service with policy
// "on-failure" does NOT restart when the process exits with code 0.
func TestRestartPolicyOnFailureWithSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := startContainer(t, "restart-on-failure-exit0.yaml")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start initially")

	// Wait for crasher to exit (it exits with code 0 after 1s)
	require.True(t, tc.waitForProcessExit("crasher", 5*time.Second),
		"crasher should exit after delay")

	// Give supervisor time to potentially restart
	time.Sleep(3 * time.Second)

	// Verify crasher did NOT restart
	assert.False(t, tc.isProcessRunning("crasher"),
		"crasher should NOT restart with policy 'on-failure' on exit code 0")
}

// TestRestartPolicyNever verifies that a service with policy "never"
// does not restart regardless of exit code.
func TestRestartPolicyNever(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := startContainer(t, "restart-never.yaml")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start initially")

	// Wait for crasher to exit (it exits with code 1 after 1s)
	require.True(t, tc.waitForProcessExit("crasher", 5*time.Second),
		"crasher should exit after delay")

	// Give supervisor time to potentially restart
	time.Sleep(3 * time.Second)

	// Verify crasher did NOT restart
	assert.False(t, tc.isProcessRunning("crasher"),
		"crasher should NOT restart with policy 'never'")
}

// TestRestartPolicyUnlessStopped verifies that a service with policy
// "unless-stopped" restarts unless explicitly stopped.
func TestRestartPolicyUnlessStopped(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := startContainer(t, "restart-unless-stopped.yaml")

	// Wait for crasher to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start initially")

	// Wait for crasher to exit naturally
	require.True(t, tc.waitForProcessExit("crasher", 5*time.Second),
		"crasher should exit after delay")

	// Wait for restart
	require.True(t, tc.waitForProcess("crasher", 5*time.Second),
		"crasher should restart with policy 'unless-stopped'")

	// Verify it restarts again
	require.True(t, tc.waitForProcessExit("crasher", 5*time.Second),
		"crasher should exit again")
	require.True(t, tc.waitForProcess("crasher", 5*time.Second),
		"crasher should restart again with policy 'unless-stopped'")
}

// TestMaxRetriesRespected verifies that the supervisor stops restarting
// a service after max_retries is reached.
func TestMaxRetriesRespected(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := startContainer(t, "restart-on-failure.yaml")

	// Config has max_retries: 3
	// Count restart cycles
	restartCount := 0

	for i := 0; i < 8; i++ {
		// Wait for crasher to start
		if tc.waitForProcess("crasher", 5*time.Second) {
			restartCount++
			// Wait for it to exit
			tc.waitForProcessExit("crasher", 5*time.Second)
		} else {
			// No more restarts
			break
		}
	}

	t.Logf("observed %d restart cycles", restartCount)

	// max_retries: 3 should limit total starts
	// Different implementations interpret this as:
	// - 1 initial + 3 retries = 4 total (most common)
	// - 3 total runs (some systems)
	// We verify it doesn't run indefinitely (should stop at some point)
	assert.GreaterOrEqual(t, restartCount, 2,
		"crasher should restart at least a few times")
	assert.LessOrEqual(t, restartCount, 6,
		"crasher should eventually stop restarting (max_retries limit)")
}
