// Package health provides internal tests for private functions.
package health

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/domain/shared"
)

// testHealthCheckConfig creates a test health check configuration.
//
// Params:
//   - interval: the interval between checks.
//   - timeout: the timeout for each check.
//
// Returns:
//   - service.HealthCheckConfig: the test configuration.
func testHealthCheckConfig(interval, timeout time.Duration) service.HealthCheckConfig {
	// Return a new health check config with the given values.
	return service.HealthCheckConfig{
		Name:     "test",
		Type:     service.HealthCheckHTTP,
		Interval: shared.Duration(interval),
		Timeout:  shared.Duration(timeout),
	}
}

// Test_Monitor_updateStatus tests the updateStatus private method.
func Test_Monitor_updateStatus(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name           string
		results        map[string]domain.Result
		expectedStatus domain.Status
	}{
		{
			name:           "empty results returns unknown",
			results:        make(map[string]domain.Result, 0),
			expectedStatus: domain.StatusUnknown,
		},
		{
			name: "all healthy returns healthy",
			results: map[string]domain.Result{
				"check1": {Status: domain.StatusHealthy},
				"check2": {Status: domain.StatusHealthy},
			},
			expectedStatus: domain.StatusHealthy,
		},
		{
			name: "all unhealthy returns unhealthy",
			results: map[string]domain.Result{
				"check1": {Status: domain.StatusUnhealthy},
				"check2": {Status: domain.StatusUnhealthy},
			},
			expectedStatus: domain.StatusUnhealthy,
		},
		{
			name: "mixed results returns degraded",
			results: map[string]domain.Result{
				"check1": {Status: domain.StatusHealthy},
				"check2": {Status: domain.StatusUnhealthy},
			},
			expectedStatus: domain.StatusDegraded,
		},
		{
			name: "single healthy returns healthy",
			results: map[string]domain.Result{
				"check1": {Status: domain.StatusHealthy},
			},
			expectedStatus: domain.StatusHealthy,
		},
		{
			name: "single unhealthy returns unhealthy",
			results: map[string]domain.Result{
				"check1": {Status: domain.StatusUnhealthy},
			},
			expectedStatus: domain.StatusUnhealthy,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create monitor with the test results.
			monitor := &Monitor{
				results: testCase.results,
			}

			// Call updateStatus.
			monitor.updateStatus()

			// Verify status matches expected.
			assert.Equal(t, testCase.expectedStatus, monitor.status)
		})
	}
}

// internalMockChecker is a test implementation for internal tests.
type internalMockChecker struct {
	name   string
	typ    string
	result domain.Result
}

// Check performs a health check.
//
// Params:
//   - _ctx: context for cancellation (unused but required by interface).
//
// Returns:
//   - domain.Result: the mock result.
func (ic *internalMockChecker) Check(_ctx context.Context) domain.Result {
	// Return the configured result.
	return ic.result
}

// Name returns the checker name.
//
// Returns:
//   - string: the name.
func (ic *internalMockChecker) Name() string {
	// Return the configured name.
	return ic.name
}

// Type returns the checker type.
//
// Returns:
//   - string: the type.
func (ic *internalMockChecker) Type() string {
	// Return the configured type.
	return ic.typ
}

// Test_Monitor_performCheck tests the performCheck private method.
func Test_Monitor_performCheck(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name                 string
		checkerName          string
		checkerType          string
		checkerResult        domain.Result
		initialResults       map[string]domain.Result
		initialConsecutive   map[string]int
		initialStatus        domain.Status
		hasEventChannel      bool
		expectEvent          bool
		expectedConsecutive  int
		expectedResultStatus domain.Status
	}{
		{
			name:                 "healthy check updates results",
			checkerName:          "test",
			checkerType:          "http",
			checkerResult:        domain.NewHealthyResult("ok", time.Millisecond),
			initialResults:       make(map[string]domain.Result, 1),
			initialConsecutive:   make(map[string]int, 1),
			initialStatus:        domain.StatusUnknown,
			hasEventChannel:      false,
			expectEvent:          false,
			expectedConsecutive:  0,
			expectedResultStatus: domain.StatusHealthy,
		},
		{
			name:                 "unhealthy check increments consecutive",
			checkerName:          "test",
			checkerType:          "http",
			checkerResult:        domain.NewUnhealthyResult("failed", time.Millisecond, nil),
			initialResults:       make(map[string]domain.Result, 1),
			initialConsecutive:   make(map[string]int, 1),
			initialStatus:        domain.StatusUnknown,
			hasEventChannel:      false,
			expectEvent:          false,
			expectedConsecutive:  1,
			expectedResultStatus: domain.StatusUnhealthy,
		},
		{
			name:                 "sends event on status change",
			checkerName:          "test",
			checkerType:          "http",
			checkerResult:        domain.NewHealthyResult("ok", time.Millisecond),
			initialResults:       make(map[string]domain.Result, 1),
			initialConsecutive:   make(map[string]int, 1),
			initialStatus:        domain.StatusUnknown,
			hasEventChannel:      true,
			expectEvent:          true,
			expectedResultStatus: domain.StatusHealthy,
		},
		{
			name:          "no event on same status",
			checkerName:   "test",
			checkerType:   "http",
			checkerResult: domain.NewHealthyResult("ok", time.Millisecond),
			initialResults: map[string]domain.Result{
				"test": {Status: domain.StatusHealthy},
			},
			initialConsecutive:   make(map[string]int, 1),
			initialStatus:        domain.StatusHealthy,
			hasEventChannel:      true,
			expectEvent:          false,
			expectedResultStatus: domain.StatusHealthy,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create mock checker.
			mockChecker := &internalMockChecker{
				name:   testCase.checkerName,
				typ:    testCase.checkerType,
				result: testCase.checkerResult,
			}

			// Create monitor with the test configuration.
			monitor := &Monitor{
				results:     testCase.initialResults,
				consecutive: testCase.initialConsecutive,
				status:      testCase.initialStatus,
			}

			// Create event channel if needed.
			var events chan domain.Event
			if testCase.hasEventChannel {
				events = make(chan domain.Event, 1)
				monitor.events = events
			}

			// Create test configuration.
			healthCheckConfig := testHealthCheckConfig(time.Second, time.Second)

			// Perform the check.
			monitor.performCheck(context.Background(), mockChecker, &healthCheckConfig)

			// Verify result was stored.
			result, resultExists := monitor.results[testCase.checkerName]
			assert.True(t, resultExists)
			assert.Equal(t, testCase.expectedResultStatus, result.Status)

			// Verify consecutive count if applicable.
			if testCase.expectedConsecutive > 0 {
				assert.Equal(t, testCase.expectedConsecutive, monitor.consecutive[testCase.checkerName])
			}

			// Verify event behavior if channel exists.
			if testCase.hasEventChannel {
				select {
				case event := <-events:
					if !testCase.expectEvent {
						t.Fatal("expected no event to be sent")
					}
					assert.Equal(t, testCase.checkerName, event.Checker)
				default:
					if testCase.expectEvent {
						t.Fatal("expected event to be sent")
					}
				}
			}
		})
	}
}

// Test_Monitor_performCheck_consecutiveIncrement tests consecutive counter increment behavior.
func Test_Monitor_performCheck_consecutiveIncrement(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name                     string
		checkerName              string
		checksCount              int
		expectedFinalConsecutive int
	}{
		{
			name:                     "single unhealthy check increments once",
			checkerName:              "test",
			checksCount:              1,
			expectedFinalConsecutive: 1,
		},
		{
			name:                     "two unhealthy checks increment twice",
			checkerName:              "test",
			checksCount:              2,
			expectedFinalConsecutive: 2,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create mock checker with unhealthy result.
			mockChecker := &internalMockChecker{
				name:   testCase.checkerName,
				typ:    "http",
				result: domain.NewUnhealthyResult("failed", time.Millisecond, nil),
			}

			// Create monitor with the checker.
			monitor := &Monitor{
				results:     make(map[string]domain.Result, 1),
				consecutive: make(map[string]int, 1),
				status:      domain.StatusUnknown,
			}

			// Create test configuration.
			healthCheckConfig := testHealthCheckConfig(time.Second, time.Second)

			// Perform the checks.
			for range testCase.checksCount {
				monitor.performCheck(context.Background(), mockChecker, &healthCheckConfig)
			}

			// Verify consecutive count matches expected.
			assert.Equal(t, testCase.expectedFinalConsecutive, monitor.consecutive[testCase.checkerName])
		})
	}
}

// Test_Monitor_runChecker tests the runChecker private method.
//
// Goroutine Lifecycle:
//   - Test goroutines are started to run the checker method under test.
//   - Each goroutine terminates when: (a) stopChannel is closed, or (b) context is cancelled.
//   - Completion is signaled via a done channel with a 100ms timeout guard.
//   - Resources are cleaned up by deferred cancel() and test framework.
func Test_Monitor_runChecker(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name             string
		useContextCancel bool
		useStopChannel   bool
		verifyResult     bool
		errorMessage     string
	}{
		{
			name:             "checker runs until stop signal",
			useContextCancel: false,
			useStopChannel:   true,
			verifyResult:     true,
			errorMessage:     "runChecker did not terminate after stop signal",
		},
		{
			name:             "checker terminates on context cancellation",
			useContextCancel: true,
			useStopChannel:   false,
			verifyResult:     false,
			errorMessage:     "runChecker did not terminate after context cancellation",
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create mock checker.
			mockChecker := &internalMockChecker{
				name:   "test",
				typ:    "http",
				result: domain.NewHealthyResult("ok", time.Millisecond),
			}

			// Create monitor with stop channel.
			stopChannel := make(chan struct{})
			monitor := &Monitor{
				results:     make(map[string]domain.Result, 1),
				consecutive: make(map[string]int, 1),
				status:      domain.StatusUnknown,
				stopCh:      stopChannel,
			}

			// Create test configuration with short interval.
			healthCheckConfig := service.HealthCheckConfig{
				Name:     "test",
				Type:     service.HealthCheckHTTP,
				Interval: shared.Duration(10 * time.Millisecond),
				Timeout:  shared.Duration(5 * time.Millisecond),
			}

			// Create cancellable context.
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Start the checker in a goroutine.
			done := make(chan struct{})
			go func() {
				monitor.runChecker(ctx, mockChecker, &healthCheckConfig)
				close(done)
			}()

			// Wait for a few checks to occur.
			time.Sleep(30 * time.Millisecond)

			// Signal termination based on test case.
			if testCase.useStopChannel {
				close(stopChannel)
			}
			if testCase.useContextCancel {
				cancel()
			}

			// Wait for goroutine to finish with timeout.
			select {
			case <-done:
				// Success - goroutine terminated.
			case <-time.After(100 * time.Millisecond):
				t.Fatal(testCase.errorMessage)
			}

			// Verify check was performed if required.
			if testCase.verifyResult {
				_, resultExists := monitor.results["test"]
				assert.True(t, resultExists)
			}
		})
	}
}

// Test_Monitor_concurrency tests thread safety of the monitor.
//
// Goroutine Lifecycle:
//   - Multiple goroutines are spawned to simulate concurrent access to monitor.
//   - Each goroutine performs iterationCount reads and then signals completion.
//   - All goroutines are waited for via the done channel before test completes.
//   - No resources require cleanup as goroutines only perform read operations.
func Test_Monitor_concurrency(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name           string
		goroutineCount int
		iterationCount int
		initialStatus  domain.Status
	}{
		{
			name:           "concurrent status reads",
			goroutineCount: 10,
			iterationCount: 100,
			initialStatus:  domain.StatusHealthy,
		},
		{
			name:           "high concurrency stress test",
			goroutineCount: 20,
			iterationCount: 50,
			initialStatus:  domain.StatusUnknown,
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create monitor.
			monitor := &Monitor{
				results:     make(map[string]domain.Result, 1),
				consecutive: make(map[string]int, 1),
				status:      testCase.initialStatus,
			}

			// Run concurrent reads.
			done := make(chan struct{}, testCase.goroutineCount)

			// Start multiple goroutines.
			for range testCase.goroutineCount {
				// Launch goroutine.
				go func() {
					// Read status multiple times.
					for range testCase.iterationCount {
						// Read status.
						_ = monitor.Status()
						// Read results.
						_ = monitor.Results()
					}
					// Signal done.
					done <- struct{}{}
				}()
			}

			// Wait for all goroutines.
			for range testCase.goroutineCount {
				// Wait for done signal.
				<-done
			}
		})
	}
}
