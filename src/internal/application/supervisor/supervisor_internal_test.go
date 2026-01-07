// Package supervisor provides internal tests for supervisor.go.
// It tests the internal implementation of the Supervisor type.
package supervisor

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/domain/service"
)

// mockLoader implements appconfig.Loader for testing.
// It provides a mock implementation that returns predefined configurations.
type mockLoader struct {
	// cfg is the configuration to return.
	cfg *service.Config
	// err is the error to return.
	err error
}

// Load returns the mock configuration.
//
// Params:
//   - path: the configuration path (unused).
//
// Returns:
//   - *service.Config: the mock configuration.
//   - error: the mock error.
func (ml *mockLoader) Load(_ string) (*service.Config, error) {
	// Return the configured mock values.
	return ml.cfg, ml.err
}

// mockExecutor implements domain.Executor for testing.
// It provides a mock implementation for process execution.
type mockExecutor struct {
	// startErr is the error to return from Start.
	startErr error
	// stopErr is the error to return from Stop.
	stopErr error
	// signalErr is the error to return from Signal.
	signalErr error
}

// Start starts a mock process.
//
// Params:
//   - ctx: the context for cancellation.
//   - spec: the process specification.
//
// Returns:
//   - int: the mock process ID.
//   - <-chan domain.ExitResult: channel for exit result.
//   - error: the mock start error.
func (ex *mockExecutor) Start(_ context.Context, _ domain.Spec) (pid int, wait <-chan domain.ExitResult, err error) {
	// Check if start error is configured.
	if ex.startErr != nil {
		// Return error when start fails.
		return 0, nil, ex.startErr
	}
	exitCh := make(chan domain.ExitResult, 1)
	// Return mock PID and exit channel.
	return 1234, exitCh, nil
}

// Stop stops a mock process.
//
// Params:
//   - pid: the process ID to stop.
//   - timeout: the stop timeout.
//
// Returns:
//   - error: the mock stop error.
func (ex *mockExecutor) Stop(_ int, _ time.Duration) error {
	// Return the configured stop error.
	return ex.stopErr
}

// Signal sends a signal to a mock process.
//
// Params:
//   - pid: the process ID.
//   - sig: the signal to send.
//
// Returns:
//   - error: the mock signal error.
func (ex *mockExecutor) Signal(_ int, _ os.Signal) error {
	// Return the configured signal error.
	return ex.signalErr
}

// mockEventser implements Eventser for testing monitorService.
// It provides a mock implementation that returns a channel.
type mockEventser struct {
	// events is the channel to return.
	events chan domain.Event
}

// Events returns the mock events channel.
//
// Returns:
//   - <-chan domain.Event: the mock event channel.
func (ev *mockEventser) Events() <-chan domain.Event {
	// Return the mock events channel.
	return ev.events
}

// createTestConfig creates a valid configuration for internal testing.
//
// Returns:
//   - *service.Config: a valid test configuration.
func createTestConfig() *service.Config {
	// Return a valid test configuration.
	return &service.Config{
		ConfigPath: "/test/config.yaml",
		Services: []service.ServiceConfig{
			{
				Name:    "test-service",
				Command: "/bin/echo",
				Args:    []string{"hello"},
			},
		},
	}
}

// createTestSupervisor creates a supervisor for internal testing.
//
// Params:
//   - t: the testing context.
//   - cfg: the service configuration.
//
// Returns:
//   - *Supervisor: the created supervisor.
func createTestSupervisor(t *testing.T, cfg *service.Config) *Supervisor {
	t.Helper()

	loader := &mockLoader{cfg: cfg}
	executor := &mockExecutor{}

	sup, err := NewSupervisor(cfg, loader, executor, nil)
	require.NoError(t, err)

	return sup
}

// Test_Supervisor_stopAll tests the stopAll method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_stopAll(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// setup is a description of the setup action.
		setup string
	}{
		{
			name:  "stop_all_after_start",
			setup: "start_supervisor",
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			sup := createTestSupervisor(t, cfg)

			// Setup the supervisor state.
			if testCase.setup == "start_supervisor" {
				ctx := context.Background()
				err := sup.Start(ctx)
				require.NoError(t, err)
			}

			// stopAll should stop all services without panic.
			sup.stopAll()

			// Cleanup.
			_ = sup.Stop()
		})
	}
}

// Test_Supervisor_updateServices tests the updateServices method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_updateServices(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// initialConfig is the initial configuration.
		initialConfig *service.Config
		// newConfig is the new configuration to apply.
		newConfig *service.Config
		// expectedServices is the list of expected service names after update.
		expectedServices []string
	}{
		{
			name:          "add_new_service",
			initialConfig: createTestConfig(),
			newConfig: &service.Config{
				ConfigPath: "/test/config.yaml",
				Services: []service.ServiceConfig{
					{
						Name:    "test-service",
						Command: "/bin/echo",
						Args:    []string{"updated"},
					},
					{
						Name:    "new-service",
						Command: "/bin/echo",
						Args:    []string{"new"},
					},
				},
			},
			expectedServices: []string{"test-service", "new-service"},
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			sup := createTestSupervisor(t, testCase.initialConfig)

			// Start the supervisor.
			ctx := context.Background()
			err := sup.Start(ctx)
			require.NoError(t, err)
			defer func() { _ = sup.Stop() }()

			// Update services.
			sup.updateServices(testCase.newConfig)

			// Verify expected services exist.
			for _, serviceName := range testCase.expectedServices {
				_, found := sup.Service(serviceName)
				assert.True(t, found, "expected service %s to exist", serviceName)
			}
		})
	}
}

// Test_Supervisor_removeDeletedServices tests the removeDeletedServices method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_removeDeletedServices(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// initialConfig is the initial configuration.
		initialConfig *service.Config
		// newConfig is the new configuration to apply.
		newConfig *service.Config
		// keptServices is the list of services that should still exist.
		keptServices []string
		// removedServices is the list of services that should be removed.
		removedServices []string
	}{
		{
			name: "remove_one_service",
			initialConfig: &service.Config{
				ConfigPath: "/test/config.yaml",
				Services: []service.ServiceConfig{
					{Name: "service-1", Command: "/bin/echo", Args: []string{"one"}},
					{Name: "service-2", Command: "/bin/echo", Args: []string{"two"}},
				},
			},
			newConfig: &service.Config{
				ConfigPath: "/test/config.yaml",
				Services: []service.ServiceConfig{
					{Name: "service-1", Command: "/bin/echo", Args: []string{"one"}},
				},
			},
			keptServices:    []string{"service-1"},
			removedServices: []string{"service-2"},
		},
		{
			name: "remove_multiple_services",
			initialConfig: &service.Config{
				ConfigPath: "/test/config.yaml",
				Services: []service.ServiceConfig{
					{Name: "service-a", Command: "/bin/echo", Args: []string{"a"}},
					{Name: "service-b", Command: "/bin/echo", Args: []string{"b"}},
					{Name: "service-c", Command: "/bin/echo", Args: []string{"c"}},
				},
			},
			newConfig: &service.Config{
				ConfigPath: "/test/config.yaml",
				Services: []service.ServiceConfig{
					{Name: "service-a", Command: "/bin/echo", Args: []string{"a"}},
				},
			},
			keptServices:    []string{"service-a"},
			removedServices: []string{"service-b", "service-c"},
		},
		{
			name: "no_services_removed_all_kept",
			initialConfig: &service.Config{
				ConfigPath: "/test/config.yaml",
				Services: []service.ServiceConfig{
					{Name: "service-x", Command: "/bin/echo", Args: []string{"x"}},
					{Name: "service-y", Command: "/bin/echo", Args: []string{"y"}},
				},
			},
			newConfig: &service.Config{
				ConfigPath: "/test/config.yaml",
				Services: []service.ServiceConfig{
					{Name: "service-x", Command: "/bin/echo", Args: []string{"x"}},
					{Name: "service-y", Command: "/bin/echo", Args: []string{"y"}},
				},
			},
			keptServices:    []string{"service-x", "service-y"},
			removedServices: nil,
		},
		{
			name: "remove_all_services_empty_new_config",
			initialConfig: &service.Config{
				ConfigPath: "/test/config.yaml",
				Services: []service.ServiceConfig{
					{Name: "service-only", Command: "/bin/echo", Args: []string{"only"}},
				},
			},
			newConfig: &service.Config{
				ConfigPath: "/test/config.yaml",
				Services:   nil,
			},
			keptServices:    nil,
			removedServices: []string{"service-only"},
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			sup := createTestSupervisor(t, testCase.initialConfig)

			// Start the supervisor.
			ctx := context.Background()
			err := sup.Start(ctx)
			require.NoError(t, err)
			defer func() { _ = sup.Stop() }()

			// Remove deleted services.
			sup.removeDeletedServices(testCase.newConfig)

			// Verify kept services still exist.
			for _, serviceName := range testCase.keptServices {
				_, found := sup.Service(serviceName)
				assert.True(t, found, "expected service %s to exist", serviceName)
			}

			// Verify removed services no longer exist.
			for _, serviceName := range testCase.removedServices {
				_, found := sup.Service(serviceName)
				assert.False(t, found, "expected service %s to be removed", serviceName)
			}
		})
	}
}

// Test_Supervisor_monitorService tests the monitorService method.
// This test creates a goroutine for monitoring that is managed by the supervisor's
// WaitGroup and cancelled via context. The goroutine terminates when the context
// is cancelled after event processing.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_monitorService(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// event is the event to send.
		event domain.Event
	}{
		{
			name:  "monitor_receives_started_event",
			event: domain.Event{Type: domain.EventStarted},
		},
		{
			name:  "monitor_receives_stopped_event",
			event: domain.Event{Type: domain.EventStopped},
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			sup := createTestSupervisor(t, cfg)

			// Create context with cancel.
			ctx, cancel := context.WithCancel(context.Background())
			sup.ctx = ctx
			sup.cancel = cancel

			// Create mock eventser with channel.
			eventsCh := make(chan domain.Event, 1)
			mockMgr := &mockEventser{events: eventsCh}

			// Add to wait group before starting monitor.
			sup.wg.Add(1)

			// Start monitor in goroutine.
			// Goroutine lifecycle:
			// - Managed by supervisor's WaitGroup
			// - Terminates when context is cancelled
			go sup.monitorService("test", mockMgr)

			// Send an event.
			eventsCh <- testCase.event

			// Give some time for event processing.
			time.Sleep(10 * time.Millisecond)

			// Cancel context to stop monitoring.
			cancel()

			// Wait for goroutine to finish.
			sup.wg.Wait()
		})
	}
}

// Test_Supervisor_handleEvent tests the handleEvent method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_handleEvent(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// serviceName is the service name.
		serviceName string
		// event is the process event.
		event domain.Event
	}{
		{
			name:        "handle_started_event",
			serviceName: "test-service",
			event:       domain.Event{Type: domain.EventStarted},
		},
		{
			name:        "handle_stopped_event",
			serviceName: "test-service",
			event:       domain.Event{Type: domain.EventStopped},
		},
		{
			name:        "handle_failed_event",
			serviceName: "test-service",
			event:       domain.Event{Type: domain.EventFailed},
		},
	}

	cfg := createTestConfig()
	sup := createTestSupervisor(t, cfg)

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			// handleEvent should not panic with any event type.
			sup.handleEvent(testCase.serviceName, &testCase.event)
		})
	}
}
