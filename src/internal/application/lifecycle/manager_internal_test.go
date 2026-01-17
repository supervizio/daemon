// Package lifecycle provides internal tests for manager.go.
// It tests internal implementation details using white-box testing.
package lifecycle

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/config"
	domain "github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/domain/shared"
)

// testExecutor implements domain.Executor for internal testing.
//
// testExecutor provides a controllable executor implementation for
// testing internal Manager behavior and state transitions.
type testExecutor struct {
	startFunc  func(ctx context.Context, spec domain.Spec) (int, <-chan domain.ExitResult, error)
	stopFunc   func(pid int, timeout time.Duration) error
	signalFunc func(pid int, sig os.Signal) error
}

// Start starts a test process.
//
// Params:
//   - ctx: the context for cancellation.
//   - spec: the process specification.
//
// Returns:
//   - int: the test process ID.
//   - <-chan domain.ExitResult: channel for exit results.
//   - error: nil on success, error on failure.
func (e *testExecutor) Start(ctx context.Context, spec domain.Spec) (pid int, wait <-chan domain.ExitResult, err error) {
	// Check if custom start function is defined.
	if e.startFunc != nil {
		// Delegate to custom start function.
		return e.startFunc(ctx, spec)
	}
	ch := make(chan domain.ExitResult, 1)
	// Return default test values.
	return 1234, ch, nil
}

// Stop stops a test process.
//
// Params:
//   - pid: the process ID to stop.
//   - timeout: the timeout for stopping.
//
// Returns:
//   - error: nil on success, error on failure.
func (e *testExecutor) Stop(pid int, timeout time.Duration) error {
	// Check if custom stop function is defined.
	if e.stopFunc != nil {
		// Delegate to custom stop function.
		return e.stopFunc(pid, timeout)
	}
	// Return nil for default behavior.
	return nil
}

// Signal sends a signal to a test process.
//
// Params:
//   - pid: the process ID to signal.
//   - sig: the signal to send.
//
// Returns:
//   - error: nil on success, error on failure.
func (e *testExecutor) Signal(pid int, sig os.Signal) error {
	// Check if custom signal function is defined.
	if e.signalFunc != nil {
		// Delegate to custom signal function.
		return e.signalFunc(pid, sig)
	}
	// Return nil for default behavior.
	return nil
}

// createInternalTestConfig creates a test service configuration.
//
// Params:
//   - name: the service name.
//   - command: the command to run.
//
// Returns:
//   - *config.ServiceConfig: the test configuration.
//
//nolint:unparam // Test helper designed for flexibility even if currently used with same values
func createInternalTestConfig(name, command string) *config.ServiceConfig {
	// Return a new service config with defaults.
	return &config.ServiceConfig{
		Name:    name,
		Command: command,
		Restart: config.RestartConfig{
			Policy:     config.RestartOnFailure,
			MaxRetries: 3,
			Delay:      shared.Seconds(1),
		},
	}
}

// Test_Manager_isContextCancelled tests the isContextCancelled method.
//
// Params:
//   - t: the testing context.
func Test_Manager_isContextCancelled(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// cancelContext indicates whether to cancel the context.
		cancelContext bool
		// expected is the expected result.
		expected bool
	}{
		{
			name:          "returns_false_when_context_active",
			cancelContext: false,
			expected:      false,
		},
		{
			name:          "returns_true_when_context_cancelled",
			cancelContext: true,
			expected:      true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			executor := &testExecutor{}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())

			// Cancel context if requested.
			if tt.cancelContext {
				mgr.cancel()
			}

			result := mgr.isContextCancelled()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_Manager_sendEvent tests the sendEvent method.
//
// Params:
//   - t: the testing context.
func Test_Manager_sendEvent(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// eventType is the event type to send.
		eventType domain.EventType
		// hasError indicates whether to include an error.
		hasError bool
	}{
		{
			name:      "sends_started_event",
			eventType: domain.EventStarted,
			hasError:  false,
		},
		{
			name:      "sends_failed_event_with_error",
			eventType: domain.EventFailed,
			hasError:  true,
		},
		{
			name:      "sends_stopped_event",
			eventType: domain.EventStopped,
			hasError:  false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			executor := &testExecutor{}

			mgr := NewManager(cfg, executor)

			var err error
			// Set error if requested.
			if tt.hasError {
				err = domain.ErrProcessFailed
			}

			mgr.sendEvent(tt.eventType, err)

			// Read event from channel.
			select {
			case event := <-mgr.events:
				assert.Equal(t, tt.eventType, event.Type)
				assert.Equal(t, "test-service", event.Process)
				// Check error presence.
				if tt.hasError {
					assert.NotNil(t, event.Error)
				} else {
					// Assert no error when not expected.
					assert.Nil(t, event.Error)
				}
			// Timeout after short duration.
			case <-time.After(100 * time.Millisecond):
				t.Fatal("expected event not received")
			}
		})
	}
}

// Test_Manager_startProcess tests the startProcess method.
//
// Params:
//   - t: the testing context.
func Test_Manager_startProcess(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startError is the error to return from executor start.
		startError error
		// expectError indicates whether an error is expected.
		expectError bool
		// expectedState is the expected state after start.
		expectedState domain.State
	}{
		{
			name:          "starts_process_successfully",
			startError:    nil,
			expectError:   false,
			expectedState: domain.StateRunning,
		},
		{
			name:          "returns_error_on_start_failure",
			startError:    shared.ErrEmptyCommand,
			expectError:   true,
			expectedState: domain.StateFailed,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			executor := &testExecutor{
				startFunc: func(ctx context.Context, spec domain.Spec) (int, <-chan domain.ExitResult, error) {
					ch := make(chan domain.ExitResult, 1)
					// Check if start should fail.
					if tt.startError != nil {
						// Return error on failure.
						return 0, nil, tt.startError
					}
					// Return success.
					return 1234, ch, nil
				},
			}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())

			err := mgr.startProcess()

			// Check if error is expected.
			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Assert no error when not expected.
				require.NoError(t, err)
				assert.Equal(t, 1234, mgr.pid)
			}
			assert.Equal(t, tt.expectedState, mgr.state)
		})
	}
}

// Test_Manager_handleProcessExit tests the handleProcessExit method.
//
// Params:
//   - t: the testing context.
func Test_Manager_handleProcessExit(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// exitCode is the exit code to test.
		exitCode int
		// expectedState is the expected state after exit.
		expectedState domain.State
	}{
		{
			name:          "handles_successful_exit",
			exitCode:      0,
			expectedState: domain.StateStopped,
		},
		{
			name:          "handles_failed_exit",
			exitCode:      1,
			expectedState: domain.StateFailed,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			cfg.Restart.Policy = config.RestartNever
			executor := &testExecutor{}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())
			mgr.pid = 1234

			result := domain.ExitResult{Code: tt.exitCode}
			_ = mgr.handleProcessExit(result)

			assert.Equal(t, tt.expectedState, mgr.state)
			assert.Equal(t, tt.exitCode, mgr.exitCode)
			assert.Equal(t, 0, mgr.pid)
		})
	}
}

// Test_Manager_waitAndRestart tests the waitAndRestart method.
//
// This test spawns a goroutine in the cancel_during_wait case to simulate
// context cancellation during the restart delay. The goroutine terminates
// after calling mgr.cancel() and has no resources to release.
//
// Params:
//   - t: the testing context.
func Test_Manager_waitAndRestart(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// cancelDuringWait indicates whether to cancel during wait.
		cancelDuringWait bool
		// initialRestarts sets the initial restart count.
		initialRestarts int
		// expected is the expected result.
		expected bool
		// expectedRestarts is the expected restart count after the call.
		expectedRestarts int
	}{
		{
			name:             "proceeds_with_restart_after_delay",
			cancelDuringWait: false,
			initialRestarts:  0,
			expected:         true,
			expectedRestarts: 1,
		},
		{
			name:             "returns_false_when_context_cancelled_during_wait",
			cancelDuringWait: true,
			initialRestarts:  0,
			expected:         false,
			expectedRestarts: 1, // restarts is incremented before waiting
		},
		{
			name:             "increments_restart_count_on_success",
			cancelDuringWait: false,
			initialRestarts:  5,
			expected:         true,
			expectedRestarts: 6,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			cfg.Restart.Delay = shared.FromTimeDuration(10 * time.Millisecond)
			executor := &testExecutor{}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())
			mgr.restarts = tt.initialRestarts

			// Cancel context during wait if requested.
			// The goroutine terminates after calling cancel().
			if tt.cancelDuringWait {
				go func() {
					time.Sleep(5 * time.Millisecond)
					mgr.cancel()
				}()
			}

			result := mgr.waitAndRestart()
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedRestarts, mgr.restarts)
		})
	}
}

// Test_constants tests the package constants.
//
// Params:
//   - t: the testing context.
func Test_constants(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// constantName is the name of the constant.
		constantName string
		// expectedValue is the expected value.
		expectedValue any
	}{
		{
			name:          "eventBufferSize_is_16",
			constantName:  "eventBufferSize",
			expectedValue: 16,
		},
		{
			name:          "defaultStopTimeout_is_30_seconds",
			constantName:  "defaultStopTimeout",
			expectedValue: 30 * time.Second,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Switch on constant name to verify value.
			switch tt.constantName {
			// Verify eventBufferSize.
			case "eventBufferSize":
				assert.Equal(t, tt.expectedValue, eventBufferSize)
			// Verify defaultStopTimeout.
			case "defaultStopTimeout":
				assert.Equal(t, tt.expectedValue, defaultStopTimeout)
			}
		})
	}
}

// Test_Manager_run tests the run method.
//
// Params:
//   - t: the testing context.
func Test_Manager_run(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// oneshot indicates if the service is oneshot.
		oneshot bool
		// expectRunning indicates expected running state after run.
		expectRunning bool
	}{
		{
			name:          "sets_running_false_after_oneshot_completes",
			oneshot:       true,
			expectRunning: false,
		},
		{
			name:          "sets_running_false_after_restart_loop_exits",
			oneshot:       false,
			expectRunning: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			cfg.Oneshot = tt.oneshot
			cfg.Restart.Policy = config.RestartNever
			exitCh := make(chan domain.ExitResult, 1)
			exitCh <- domain.ExitResult{Code: 0}

			executor := &testExecutor{
				startFunc: func(ctx context.Context, spec domain.Spec) (int, <-chan domain.ExitResult, error) {
					// Return the prepared exit channel.
					return 1234, exitCh, nil
				},
			}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())
			mgr.running = true

			mgr.run()

			assert.Equal(t, tt.expectRunning, mgr.running)
		})
	}
}

// Test_Manager_runOnce tests the runOnce method.
//
// Params:
//   - t: the testing context.
func Test_Manager_runOnce(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startError is the error to return from start.
		startError error
		// exitCode is the exit code of the process.
		exitCode int
		// expectedEventType is the expected final event type.
		expectedEventType domain.EventType
	}{
		{
			name:              "sends_stopped_on_success",
			startError:        nil,
			exitCode:          0,
			expectedEventType: domain.EventStopped,
		},
		{
			name:              "sends_failed_on_start_error",
			startError:        shared.ErrEmptyCommand,
			exitCode:          0,
			expectedEventType: domain.EventFailed,
		},
		{
			name:              "sends_failed_on_nonzero_exit",
			startError:        nil,
			exitCode:          1,
			expectedEventType: domain.EventFailed,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			cfg.Oneshot = true
			exitCh := make(chan domain.ExitResult, 1)
			exitCh <- domain.ExitResult{Code: tt.exitCode}

			executor := &testExecutor{
				startFunc: func(ctx context.Context, spec domain.Spec) (int, <-chan domain.ExitResult, error) {
					// Check if start should fail.
					if tt.startError != nil {
						// Return error on failure.
						return 0, nil, tt.startError
					}
					// Return success with exit channel.
					return 1234, exitCh, nil
				},
			}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())

			mgr.runOnce()

			// Drain events to find the final event.
			var lastEvent domain.Event
			eventFound := false
			// Loop to read all available events.
		drainLoop:
			for {
				select {
				case event := <-mgr.events:
					lastEvent = event
					eventFound = true
				// Default case for non-blocking read.
				default:
					// Exit loop when no more events.
					break drainLoop
				}
			}
			require.True(t, eventFound, "expected at least one event")
			assert.Equal(t, tt.expectedEventType, lastEvent.Type)
		})
	}
}

// Test_Manager_runWithRestart tests the runWithRestart method.
//
// Params:
//   - t: the testing context.
func Test_Manager_runWithRestart(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// cancelBeforeStart indicates whether to cancel before start.
		cancelBeforeStart bool
		// startError is the error to return from start.
		startError error
		// exitCode is the exit code of the process.
		exitCode int
		// expectedRestarts is the expected number of restarts.
		expectedRestarts int
	}{
		{
			name:              "exits_on_cancelled_context",
			cancelBeforeStart: true,
			startError:        nil,
			exitCode:          0,
			expectedRestarts:  0,
		},
		{
			name:              "exits_on_successful_exit_with_never_policy",
			cancelBeforeStart: false,
			startError:        nil,
			exitCode:          0,
			expectedRestarts:  0,
		},
		{
			name:              "exits_when_start_fails_with_never_policy",
			cancelBeforeStart: false,
			startError:        shared.ErrEmptyCommand,
			exitCode:          0,
			expectedRestarts:  0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			cfg.Restart.Policy = config.RestartNever
			exitCh := make(chan domain.ExitResult, 1)
			exitCh <- domain.ExitResult{Code: tt.exitCode}

			executor := &testExecutor{
				startFunc: func(ctx context.Context, spec domain.Spec) (int, <-chan domain.ExitResult, error) {
					// Check if start should fail.
					if tt.startError != nil {
						// Return error on failure.
						return 0, nil, tt.startError
					}
					// Return success with exit channel.
					return 1234, exitCh, nil
				},
			}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())

			// Cancel before start if requested.
			if tt.cancelBeforeStart {
				mgr.cancel()
			}

			mgr.runWithRestart()

			assert.Equal(t, tt.expectedRestarts, mgr.restarts)
		})
	}
}

// Test_Manager_tryStartProcess tests the tryStartProcess method.
//
// Params:
//   - t: the testing context.
func Test_Manager_tryStartProcess(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startError is the error to return from start.
		startError error
		// restartPolicy is the restart policy to use.
		restartPolicy config.RestartPolicy
		// expected is the expected return value.
		expected bool
	}{
		{
			name:          "returns_true_on_success",
			startError:    nil,
			restartPolicy: config.RestartNever,
			expected:      true,
		},
		{
			name:          "returns_false_on_error_with_never_policy",
			startError:    shared.ErrEmptyCommand,
			restartPolicy: config.RestartNever,
			expected:      false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			cfg.Restart.Policy = tt.restartPolicy
			exitCh := make(chan domain.ExitResult, 1)

			executor := &testExecutor{
				startFunc: func(ctx context.Context, spec domain.Spec) (int, <-chan domain.ExitResult, error) {
					// Check if start should fail.
					if tt.startError != nil {
						// Return error on failure.
						return 0, nil, tt.startError
					}
					// Return success with exit channel.
					return 1234, exitCh, nil
				},
			}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())

			result := mgr.tryStartProcess()

			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_Manager_waitForProcessOrShutdown tests the waitForProcessOrShutdown method.
//
// Params:
//   - t: the testing context.
func Test_Manager_waitForProcessOrShutdown(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// cancelContext indicates whether to cancel context.
		cancelContext bool
		// sendExit indicates whether to send exit result.
		sendExit bool
		// exitCode is the exit code to send.
		exitCode int
		// expected is the expected return value.
		expected bool
	}{
		{
			name:          "returns_true_on_context_cancelled",
			cancelContext: true,
			sendExit:      false,
			exitCode:      0,
			expected:      true,
		},
		{
			name:          "returns_true_on_process_exit_with_never_policy",
			cancelContext: false,
			sendExit:      true,
			exitCode:      0,
			expected:      true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			cfg.Restart.Policy = config.RestartNever
			exitCh := make(chan domain.ExitResult, 1)

			executor := &testExecutor{
				stopFunc: func(pid int, timeout time.Duration) error {
					// Return nil for successful stop.
					return nil
				},
			}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())
			mgr.waitCh = exitCh
			mgr.pid = 1234

			// Cancel context or send exit based on test case.
			if tt.cancelContext {
				mgr.cancel()
			}
			// Send exit result if requested.
			if tt.sendExit {
				exitCh <- domain.ExitResult{Code: tt.exitCode}
			}

			result := mgr.waitForProcessOrShutdown()

			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_Manager_Uptime_when_running tests Uptime when process is running.
//
// Params:
//   - t: the testing context.
func Test_Manager_Uptime_when_running(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// state is the process state.
		state domain.State
		// expectNonZero indicates if uptime should be non-zero.
		expectNonZero bool
	}{
		{
			name:          "returns_nonzero_when_running",
			state:         domain.StateRunning,
			expectNonZero: true,
		},
		{
			name:          "returns_zero_when_stopped",
			state:         domain.StateStopped,
			expectNonZero: false,
		},
		{
			name:          "returns_zero_when_failed",
			state:         domain.StateFailed,
			expectNonZero: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			executor := &testExecutor{}

			mgr := NewManager(cfg, executor)
			mgr.state = tt.state
			mgr.startTime = time.Now().Add(-1 * time.Second)

			uptime := mgr.Uptime()

			// Check uptime based on expected result.
			if tt.expectNonZero {
				assert.Greater(t, uptime, int64(0))
			} else {
				assert.Equal(t, int64(0), uptime)
			}
		})
	}
}

// Test_Manager_tryStartProcess_with_restart tests tryStartProcess with restart policy.
// This test may spawn a goroutine to cancel the context during restart wait.
// The goroutine terminates after calling cancel(), typically within 5ms.
//
// Params:
//   - t: the testing context.
func Test_Manager_tryStartProcess_with_restart(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startError is the error to return from start.
		startError error
		// restartPolicy is the restart policy to use.
		restartPolicy config.RestartPolicy
		// maxRetries is the maximum retries allowed.
		maxRetries int
		// cancelDuringWait indicates whether to cancel during wait.
		cancelDuringWait bool
		// expected is the expected return value.
		expected bool
	}{
		{
			name:          "returns_true_on_error_with_always_policy_and_restart",
			startError:    shared.ErrEmptyCommand,
			restartPolicy: config.RestartAlways,
			maxRetries:    3,
			expected:      true,
		},
		{
			name:             "returns_false_on_error_when_cancelled_during_wait",
			startError:       shared.ErrEmptyCommand,
			restartPolicy:    config.RestartAlways,
			maxRetries:       3,
			cancelDuringWait: true,
			expected:         false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			cfg.Restart.Policy = tt.restartPolicy
			cfg.Restart.MaxRetries = tt.maxRetries
			cfg.Restart.Delay = shared.FromTimeDuration(10 * time.Millisecond)

			executor := &testExecutor{
				startFunc: func(ctx context.Context, spec domain.Spec) (int, <-chan domain.ExitResult, error) {
					// Return error on failure.
					return 0, nil, tt.startError
				},
			}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())

			// Cancel during wait if requested.
			// The goroutine terminates after calling cancel().
			if tt.cancelDuringWait {
				go func() {
					time.Sleep(5 * time.Millisecond)
					mgr.cancel()
				}()
			}

			result := mgr.tryStartProcess()

			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_Manager_handleProcessExit_with_restart tests handleProcessExit with restart.
// This test may spawn a goroutine to cancel the context during restart wait.
// The goroutine terminates after calling cancel(), typically within 5ms.
//
// Params:
//   - t: the testing context.
func Test_Manager_handleProcessExit_with_restart(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// exitCode is the exit code to test.
		exitCode int
		// restartPolicy is the restart policy.
		restartPolicy config.RestartPolicy
		// cancelDuringWait indicates whether to cancel during wait.
		cancelDuringWait bool
		// expected is the expected return value.
		expected bool
	}{
		{
			name:          "continues_restart_loop_on_failure_with_always_policy",
			exitCode:      1,
			restartPolicy: config.RestartAlways,
			expected:      false,
		},
		{
			name:          "continues_restart_loop_on_success_with_always_policy",
			exitCode:      0,
			restartPolicy: config.RestartAlways,
			expected:      false,
		},
		{
			name:             "stops_when_cancelled_during_restart_wait",
			exitCode:         1,
			restartPolicy:    config.RestartAlways,
			cancelDuringWait: true,
			expected:         true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			cfg.Restart.Policy = tt.restartPolicy
			cfg.Restart.MaxRetries = 3
			cfg.Restart.Delay = shared.FromTimeDuration(10 * time.Millisecond)
			executor := &testExecutor{}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())
			mgr.pid = 1234

			// Cancel during wait if requested.
			// The goroutine terminates after calling cancel().
			if tt.cancelDuringWait {
				go func() {
					time.Sleep(5 * time.Millisecond)
					mgr.cancel()
				}()
			}

			result := domain.ExitResult{Code: tt.exitCode}
			shouldStop := mgr.handleProcessExit(result)

			assert.Equal(t, tt.expected, shouldStop)
		})
	}
}

// Test_Manager_Stop_with_running_process tests Stop with a running process.
//
// Params:
//   - t: the testing context.
func Test_Manager_Stop_with_running_process(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// pid is the process ID.
		pid int
		// stopErr is the error to return from stop.
		stopErr error
		// expectError indicates whether an error is expected.
		expectError bool
	}{
		{
			name:        "stops_running_process_successfully",
			pid:         1234,
			stopErr:     nil,
			expectError: false,
		},
		{
			name:        "returns_error_when_stop_fails",
			pid:         1234,
			stopErr:     domain.ErrProcessFailed,
			expectError: true,
		},
		{
			name:        "succeeds_when_pid_is_zero",
			pid:         0,
			stopErr:     nil,
			expectError: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			executor := &testExecutor{
				stopFunc: func(pid int, timeout time.Duration) error {
					// Return the configured stop error.
					return tt.stopErr
				},
			}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())
			mgr.running = true
			mgr.pid = tt.pid

			err := mgr.Stop()

			// Check if error is expected.
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test_Manager_Reload_when_running tests Reload when process is running.
//
// Params:
//   - t: the testing context.
func Test_Manager_Reload_when_running(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// pid is the process ID.
		pid int
		// signalErr is the error to return from signal.
		signalErr error
		// expectError indicates whether an error is expected.
		expectError bool
	}{
		{
			name:        "reloads_running_process_successfully",
			pid:         1234,
			signalErr:   nil,
			expectError: false,
		},
		{
			name:        "returns_error_when_signal_fails",
			pid:         1234,
			signalErr:   domain.ErrProcessFailed,
			expectError: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			executor := &testExecutor{
				signalFunc: func(pid int, sig os.Signal) error {
					// Return the configured signal error.
					return tt.signalErr
				},
			}

			mgr := NewManager(cfg, executor)
			mgr.pid = tt.pid

			err := mgr.Reload()

			// Check if error is expected.
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test_Manager_runWithRestart_with_process_exit tests runWithRestart with process exit.
// This test may spawn a goroutine to cancel the context after start.
// The goroutine terminates after calling cancel(), typically within 50ms.
//
// Params:
//   - t: the testing context.
func Test_Manager_runWithRestart_with_process_exit(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// exitCode is the exit code of the process.
		exitCode int
		// restartPolicy is the restart policy.
		restartPolicy config.RestartPolicy
		// maxRetries is the maximum retries.
		maxRetries int
		// cancelAfterStart indicates whether to cancel after start.
		cancelAfterStart bool
	}{
		{
			name:             "exits_on_context_cancel_after_process_exit",
			exitCode:         1,
			restartPolicy:    config.RestartAlways,
			maxRetries:       3,
			cancelAfterStart: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			cfg.Restart.Policy = tt.restartPolicy
			cfg.Restart.MaxRetries = tt.maxRetries
			cfg.Restart.Delay = shared.FromTimeDuration(10 * time.Millisecond)

			callCount := 0
			executor := &testExecutor{
				startFunc: func(ctx context.Context, spec domain.Spec) (int, <-chan domain.ExitResult, error) {
					callCount++
					exitCh := make(chan domain.ExitResult, 1)
					exitCh <- domain.ExitResult{Code: tt.exitCode}
					// Return success with exit channel.
					return 1234, exitCh, nil
				},
			}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())

			// Cancel after short delay if requested.
			// The goroutine terminates after calling cancel().
			if tt.cancelAfterStart {
				go func() {
					time.Sleep(50 * time.Millisecond)
					mgr.cancel()
				}()
			}

			mgr.runWithRestart()

			// Verify at least one start attempt was made.
			assert.GreaterOrEqual(t, callCount, 1)
		})
	}
}

// Test_Manager_runWithRestart_shutdown_during_wait tests runWithRestart shutdown.
// This test covers the case where waitForProcessOrShutdown returns true (shutdown).
// This test spawns a goroutine to cancel the context during process wait.
// The goroutine terminates after calling cancel(), typically within 20ms.
//
// Params:
//   - t: the testing context.
func Test_Manager_runWithRestart_shutdown_during_wait(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "exits_when_shutdown_during_wait_for_process",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			cfg.Restart.Policy = config.RestartNever

			// Create a blocking exit channel that never sends.
			blockingExitCh := make(chan domain.ExitResult)

			startCalled := false
			executor := &testExecutor{
				startFunc: func(ctx context.Context, spec domain.Spec) (int, <-chan domain.ExitResult, error) {
					startCalled = true
					// Return blocking exit channel - process never exits.
					return 1234, blockingExitCh, nil
				},
				stopFunc: func(pid int, timeout time.Duration) error {
					// Return nil for successful stop.
					return nil
				},
			}

			mgr := NewManager(cfg, executor)
			mgr.ctx, mgr.cancel = context.WithCancel(context.Background())

			// Cancel context shortly after start to trigger shutdown.
			// The goroutine terminates after calling cancel().
			go func() {
				time.Sleep(20 * time.Millisecond)
				mgr.cancel()
			}()

			mgr.runWithRestart()

			// Verify start was called.
			assert.True(t, startCalled)
		})
	}
}

// Test_Manager_sendEvent_channel_full tests sendEvent when channel is full.
//
// Params:
//   - t: the testing context.
func Test_Manager_sendEvent_channel_full(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// fillChannel indicates whether to fill the channel first.
		fillChannel bool
	}{
		{
			name:        "drops_event_when_channel_full",
			fillChannel: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createInternalTestConfig("test-service", "/bin/echo")
			executor := &testExecutor{}

			mgr := NewManager(cfg, executor)

			// Fill the channel if requested.
			if tt.fillChannel {
				// Fill the channel to capacity.
				for range eventBufferSize {
					mgr.events <- domain.Event{}
				}
			}

			// This should not block even when channel is full.
			mgr.sendEvent(domain.EventStarted, nil)

			// Verify channel size hasn't changed.
			assert.Len(t, mgr.events, eventBufferSize)
		})
	}
}
