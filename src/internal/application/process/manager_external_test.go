// Package process_test provides external tests for manager.go.
// It tests the public API of the Manager type using black-box testing.
package process_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/application/process"
	domain "github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/domain/shared"
)

// mockExecutor implements domain.Executor for testing.
//
// mockExecutor provides a controllable executor implementation that allows
// tests to simulate various process lifecycle scenarios.
type mockExecutor struct {
	startFunc  func(ctx context.Context, spec domain.Spec) (int, <-chan domain.ExitResult, error)
	stopFunc   func(pid int, timeout time.Duration) error
	signalFunc func(pid int, sig os.Signal) error
}

// Start starts a mock process.
//
// Params:
//   - ctx: the context for cancellation.
//   - spec: the process specification.
//
// Returns:
//   - int: the mock process ID.
//   - <-chan domain.ExitResult: channel for exit results.
//   - error: nil on success, error on failure.
func (m *mockExecutor) Start(ctx context.Context, spec domain.Spec) (pid int, wait <-chan domain.ExitResult, err error) {
	// Check if custom start function is defined.
	if m.startFunc != nil {
		// Delegate to custom start function.
		return m.startFunc(ctx, spec)
	}
	ch := make(chan domain.ExitResult, 1)
	// Return default mock values.
	return 1234, ch, nil
}

// Stop stops a mock process.
//
// Params:
//   - pid: the process ID to stop.
//   - timeout: the timeout for stopping.
//
// Returns:
//   - error: nil on success, error on failure.
func (m *mockExecutor) Stop(pid int, timeout time.Duration) error {
	// Check if custom stop function is defined.
	if m.stopFunc != nil {
		// Delegate to custom stop function.
		return m.stopFunc(pid, timeout)
	}
	// Return nil for default behavior.
	return nil
}

// Signal sends a signal to a mock process.
//
// Params:
//   - pid: the process ID to signal.
//   - sig: the signal to send.
//
// Returns:
//   - error: nil on success, error on failure.
func (m *mockExecutor) Signal(pid int, sig os.Signal) error {
	// Check if custom signal function is defined.
	if m.signalFunc != nil {
		// Delegate to custom signal function.
		return m.signalFunc(pid, sig)
	}
	// Return nil for default behavior.
	return nil
}

// createTestConfig creates a test service configuration.
//
// Params:
//   - name: the service name.
//   - command: the command to run.
//
// Returns:
//   - *service.ServiceConfig: the test configuration.
func createTestConfig(name, command string) *service.ServiceConfig {
	// Return a new service config with defaults.
	return &service.ServiceConfig{
		Name:    name,
		Command: command,
		Restart: service.RestartConfig{
			Policy:     service.RestartOnFailure,
			MaxRetries: 3,
			Delay:      shared.Seconds(1),
		},
	}
}

// TestNewManager tests the NewManager constructor.
//
// Params:
//   - t: the testing context.
func TestNewManager(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// serviceName is the name for the test service.
		serviceName string
		// command is the command for the test service.
		command string
	}{
		{
			name:        "creates_manager_with_defaults",
			serviceName: "test-service",
			command:     "/bin/echo",
		},
		{
			name:        "creates_manager_with_custom_name",
			serviceName: "custom-service",
			command:     "/bin/sleep",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig(tt.serviceName, tt.command)
			executor := &mockExecutor{}

			mgr := process.NewManager(cfg, executor)

			require.NotNil(t, mgr)
			assert.Equal(t, domain.StateStopped, mgr.State())
			assert.Equal(t, 0, mgr.PID())
			assert.NotNil(t, mgr.Events())
		})
	}
}

// TestManager_State tests the State method.
//
// Params:
//   - t: the testing context.
func TestManager_State(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// expectedState is the expected initial state.
		expectedState domain.State
	}{
		{
			name:          "initial_state_is_stopped",
			expectedState: domain.StateStopped,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig("test-service", "/bin/echo")
			executor := &mockExecutor{}

			mgr := process.NewManager(cfg, executor)

			assert.Equal(t, tt.expectedState, mgr.State())
		})
	}
}

// TestManager_PID tests the PID method.
//
// Params:
//   - t: the testing context.
func TestManager_PID(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// expectedPID is the expected initial PID.
		expectedPID int
	}{
		{
			name:        "initial_pid_is_zero",
			expectedPID: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig("test-service", "/bin/echo")
			executor := &mockExecutor{}

			mgr := process.NewManager(cfg, executor)

			assert.Equal(t, tt.expectedPID, mgr.PID())
		})
	}
}

// TestManager_Uptime tests the Uptime method.
//
// Params:
//   - t: the testing context.
func TestManager_Uptime(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// expectedUptime is the expected initial uptime.
		expectedUptime int64
	}{
		{
			name:           "initial_uptime_is_zero",
			expectedUptime: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig("test-service", "/bin/echo")
			executor := &mockExecutor{}

			mgr := process.NewManager(cfg, executor)

			assert.Equal(t, tt.expectedUptime, mgr.Uptime())
		})
	}
}

// TestManager_Events tests the Events method.
//
// Params:
//   - t: the testing context.
func TestManager_Events(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "returns_non_nil_channel",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig("test-service", "/bin/echo")
			executor := &mockExecutor{}

			mgr := process.NewManager(cfg, executor)

			events := mgr.Events()
			assert.NotNil(t, events)
		})
	}
}

// TestManager_Start tests the Start method.
//
// Params:
//   - t: the testing context.
func TestManager_Start(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startTwice indicates whether to start twice.
		startTwice bool
		// expectError indicates whether an error is expected on second start.
		expectError bool
	}{
		{
			name:        "starts_successfully",
			startTwice:  false,
			expectError: false,
		},
		{
			name:        "returns_error_when_already_running",
			startTwice:  true,
			expectError: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig("test-service", "/bin/echo")
			executor := &mockExecutor{}

			mgr := process.NewManager(cfg, executor)

			err := mgr.Start()
			require.NoError(t, err)
			defer func() { _ = mgr.Stop() }()

			// Attempt second start if requested.
			if tt.startTwice {
				err = mgr.Start()
				// Check if error is expected.
				if tt.expectError {
					assert.Error(t, err)
				} else {
					// Assert no error when not expected.
					assert.NoError(t, err)
				}
			}
		})
	}
}

// TestManager_Stop tests the Stop method.
//
// Params:
//   - t: the testing context.
func TestManager_Stop(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startFirst indicates whether to start before stopping.
		startFirst bool
	}{
		{
			name:       "stops_running_manager",
			startFirst: true,
		},
		{
			name:       "stops_without_error_when_not_running",
			startFirst: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig("test-service", "/bin/echo")
			executor := &mockExecutor{}

			mgr := process.NewManager(cfg, executor)

			// Start if requested.
			if tt.startFirst {
				err := mgr.Start()
				require.NoError(t, err)
			}

			err := mgr.Stop()
			assert.NoError(t, err)
		})
	}
}

// TestManager_Reload tests the Reload method.
//
// Params:
//   - t: the testing context.
func TestManager_Reload(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// expectError indicates whether an error is expected.
		expectError bool
	}{
		{
			name:        "returns_error_when_not_running",
			expectError: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig("test-service", "/bin/echo")
			executor := &mockExecutor{}

			mgr := process.NewManager(cfg, executor)

			err := mgr.Reload()

			// Check if error is expected.
			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrNotRunning)
			} else {
				// Assert no error when not expected.
				assert.NoError(t, err)
			}
		})
	}
}

// TestManager_Status tests the Status method.
//
// Params:
//   - t: the testing context.
func TestManager_Status(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// serviceName is the expected service name in status.
		serviceName string
	}{
		{
			name:        "returns_status_with_service_name",
			serviceName: "test-service",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig(tt.serviceName, "/bin/echo")
			executor := &mockExecutor{}

			mgr := process.NewManager(cfg, executor)

			status := mgr.Status()

			assert.Equal(t, tt.serviceName, status.Name)
			assert.Equal(t, domain.StateStopped, status.State)
			assert.Equal(t, 0, status.PID)
			assert.Equal(t, 0, status.Restarts)
		})
	}
}
