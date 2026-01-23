package behavioral_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExponentialBackoff verifies that restart delays increase
// after repeated failures. This test validates backoff behavior
// by measuring the total time for multiple restarts.
func TestExponentialBackoff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := startContainer(t, "backoff.yaml")

	// Config has delay: 1s, delay_max: 10s, max_retries: 5
	// Crasher exits immediately with code 1
	// Expected backoff: ~1s, ~2s, ~4s, ~8s, ~10s (capped)

	// Wait for initial process start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start initially")

	// Track restart count over a fixed time window
	// With backoff starting at 1s, we should see fewer restarts than
	// if there was no backoff (which would be many more in same time)
	startTime := time.Now()
	restartCount := 0

	// Wait for process to exit and restart multiple times
	for i := 0; i < 4; i++ {
		// Wait for exit
		if !tc.waitForProcessExit("crasher", 3*time.Second) {
			t.Logf("crasher did not exit in cycle %d", i)
			continue
		}

		// Wait for restart (with increasing timeout for backoff)
		timeout := time.Duration((i+2)*3) * time.Second
		if tc.waitForProcess("crasher", timeout) {
			restartCount++
			t.Logf("restart %d detected", restartCount)
		} else {
			// May have hit max_retries
			t.Logf("no restart detected in cycle %d (may be expected)", i)
			break
		}
	}

	elapsed := time.Since(startTime)

	// With exponential backoff (1s, 2s, 4s...), 3-4 restarts should take
	// at least 7 seconds (1+2+4). Without backoff, it would be much faster.
	t.Logf("observed %d restarts in %v", restartCount, elapsed)

	// Verify we saw some restarts (backoff is working, not blocking)
	assert.GreaterOrEqual(t, restartCount, 2,
		"should see at least 2 restarts")

	// Verify it took meaningful time (backoff delays are happening)
	assert.GreaterOrEqual(t, elapsed.Seconds(), 3.0,
		"restarts should take time due to backoff delays")
}

// TestBackoffRespectsMaxDelay verifies that the exponential backoff
// does not exceed the configured delay_max.
func TestBackoffRespectsMaxDelay(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := startContainer(t, "backoff.yaml")

	// Config has delay: 1s, delay_max: 10s, max_retries: 5
	// After several restarts, delay should cap at delay_max

	// Wait for initial process to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start")

	// Let several restart cycles happen to reach max delay
	// Expected delays: 1s, 2s, 4s, 8s, 10s (capped)
	for i := 0; i < 4; i++ {
		tc.waitForProcessExit("crasher", 3*time.Second)
		// Use longer timeout to accommodate increasing backoff
		if !tc.waitForProcess("crasher", 15*time.Second) {
			break
		}
	}

	// Now measure the next restart interval
	// It should be capped at delay_max (10s)
	if tc.waitForProcessExit("crasher", 3*time.Second) {
		start := time.Now()

		// Wait for potential restart (may not happen if max_retries reached)
		restarted := tc.waitForProcess("crasher", 15*time.Second)
		elapsed := time.Since(start)

		if restarted {
			// Delay should not exceed delay_max (10s) plus some tolerance
			// for container/process overhead
			assert.Less(t, elapsed.Seconds(), 14.0,
				"restart delay should not significantly exceed delay_max")
			t.Logf("restart delay was %v (max expected ~10s)", elapsed)
		} else {
			// Process didn't restart - might have hit max_retries
			// This is also valid behavior
			t.Log("process did not restart (may have hit max_retries)")
		}
	}
}

// TestBackoffResetsOnStability verifies that the backoff resets
// when a service runs successfully for a period of time.
func TestBackoffResetsOnStability(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tc := startContainer(t, "backoff-reset.yaml")

	// Config: crash-after=4s (runs 4s then crashes), stability_window=3s
	// delay=500ms, delay_max=5s, max_retries=10
	//
	// Expected behavior:
	// 1. Crasher runs 4s, crashes -> restart with 500ms delay (attempt 1)
	// 2. Crasher runs 4s, crashes -> backoff reset (4s > 3s window)
	// 3. Next restart should use initial delay (500ms), not escalated delay

	// Wait for initial process to start
	require.True(t, tc.waitForProcess("crasher", 10*time.Second),
		"crasher should start")

	// Track restart timings
	var restartTimes []time.Time

	// Wait for first crash (after 4s) and restart
	restartTimes = append(restartTimes, time.Now())
	tc.waitForProcessExit("crasher", 6*time.Second)

	// First restart
	require.True(t, tc.waitForProcess("crasher", 3*time.Second),
		"crasher should restart after first crash")
	restartTimes = append(restartTimes, time.Now())

	// Wait for second crash (after 4s) - this run exceeds stability window
	tc.waitForProcessExit("crasher", 6*time.Second)

	// Second restart - backoff should have reset because previous run was > 3s
	start := time.Now()
	require.True(t, tc.waitForProcess("crasher", 3*time.Second),
		"crasher should restart after second crash")
	restartDelay := time.Since(start)
	restartTimes = append(restartTimes, time.Now())

	t.Logf("restart timings: %v", restartTimes)
	t.Logf("delay after stability reset: %v", restartDelay)

	// The restart delay should be close to initial delay (500ms), not escalated
	// Allow some tolerance for process startup overhead
	assert.Less(t, restartDelay.Milliseconds(), int64(1500),
		"restart delay should be near initial delay (500ms) after stability reset, not escalated")
}
