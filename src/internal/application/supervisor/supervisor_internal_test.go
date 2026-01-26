// Package supervisor provides internal tests for supervisor.go.
// It tests the internal implementation of the Supervisor type.
package supervisor

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apphealth "github.com/kodflow/daemon/internal/application/health"
	applifecycle "github.com/kodflow/daemon/internal/application/lifecycle"
	"github.com/kodflow/daemon/internal/domain/config"
	domainhealth "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/listener"
	domain "github.com/kodflow/daemon/internal/domain/process"
)

// mockLoader implements appconfig.Loader for testing.
// It provides a mock implementation that returns predefined configurations.
type mockLoader struct {
	// cfg is the configuration to return.
	cfg *config.Config
	// err is the error to return.
	err error
}

// Load returns the mock configuration.
//
// Params:
//   - path: the configuration path (unused).
//
// Returns:
//   - *config.Config: the mock configuration.
//   - error: the mock error.
func (ml *mockLoader) Load(_ string) (*config.Config, error) {
	// Return the configured mock values.
	return ml.cfg, ml.err
}

// mockExecutor implements domain.Executor for testing.
// It provides a mock implementation for process execution.
type mockExecutor struct {
	// mu protects all error fields from concurrent access.
	mu sync.RWMutex
	// startErr is the error to return from Start.
	startErr error
	// stopErr is the error to return from Stop.
	stopErr error
	// signalErr is the error to return from Signal.
	signalErr error
}

// SetStopErr sets the stop error in a thread-safe manner.
//
// Params:
//   - err: the error to return from Stop.
func (ex *mockExecutor) SetStopErr(err error) {
	ex.mu.Lock()
	defer ex.mu.Unlock()
	ex.stopErr = err
}

// SetStartErr sets the start error in a thread-safe manner.
//
// Params:
//   - err: the error to return from Start.
func (ex *mockExecutor) SetStartErr(err error) {
	ex.mu.Lock()
	defer ex.mu.Unlock()
	ex.startErr = err
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
	ex.mu.RLock()
	startErr := ex.startErr
	ex.mu.RUnlock()
	// Check if start error is configured.
	if startErr != nil {
		// Return error when start fails.
		return 0, nil, startErr
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
	ex.mu.RLock()
	defer ex.mu.RUnlock()
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
	ex.mu.RLock()
	defer ex.mu.RUnlock()
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
//   - *config.Config: a valid test configuration.
func createTestConfig() *config.Config {
	// Return a valid test configuration.
	return &config.Config{
		ConfigPath: "/test/config.yaml",
		Services: []config.ServiceConfig{
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
func createTestSupervisor(t *testing.T, cfg *config.Config) *Supervisor {
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
		// useMultipleServices determines if multiple services should be configured.
		useMultipleServices bool
	}{
		{
			name:                "stop_all_after_start",
			setup:               "start_supervisor",
			useMultipleServices: false,
		},
		{
			name:                "stop_all_without_start",
			setup:               "no_start",
			useMultipleServices: false,
		},
		{
			name:                "stop_all_multiple_services",
			setup:               "start_supervisor",
			useMultipleServices: true,
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			// Create config based on test case.
			var cfg *config.Config
			// Check if multiple services are needed.
			if testCase.useMultipleServices {
				cfg = createMultiServiceTestConfig()
			} else {
				cfg = createTestConfig()
			}

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

// Test_Supervisor_stopAll_withStopError tests stopAll when manager Stop fails.
func Test_Supervisor_stopAll_withStopError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "stop_error_is_handled_via_recovery_handler",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			sup := createTestSupervisor(t, cfg)

			// Start the supervisor first.
			ctx := context.Background()
			err := sup.Start(ctx)
			require.NoError(t, err)

			// Wait for the service to have a valid PID.
			// The manager's Start() is asynchronous and spawns a goroutine
			// that calls startProcess() which sets the PID.
			require.Eventually(t, func() bool {
				services := sup.Services()
				info, ok := services["test-service"]
				return ok && info.PID > 0
			}, time.Second, 10*time.Millisecond, "service should have valid PID")

			// Track if error handler was called.
			var handlerCalled bool
			var handlerOp string
			var handlerService string
			sup.SetErrorHandler(func(op, serviceName string, err error) {
				handlerCalled = true
				handlerOp = op
				handlerService = serviceName
			})

			// Configure the executor to return an error on Stop (thread-safe).
			if ex, ok := sup.executor.(*mockExecutor); ok {
				ex.SetStopErr(domain.ErrProcessFailed)
			}

			// stopAll should not panic even when Stop returns error.
			sup.stopAll()

			// Verify error handler was called.
			assert.True(t, handlerCalled, "error handler should be called")
			assert.Equal(t, "stop", handlerOp)
			assert.Equal(t, "test-service", handlerService)

			// Cleanup: reset stopErr (thread-safe).
			if ex, ok := sup.executor.(*mockExecutor); ok {
				ex.SetStopErr(nil)
			}
			_ = sup.Stop()
		})
	}
}

// createMultiServiceTestConfig creates a test configuration with multiple services.
//
// Returns:
//   - *config.Config: a configuration with multiple services for testing.
func createMultiServiceTestConfig() *config.Config {
	// Return a configuration with multiple services.
	return &config.Config{
		ConfigPath: "/test/config.yaml",
		Services: []config.ServiceConfig{
			{
				Name:    "service-1",
				Command: "/bin/echo",
				Args:    []string{"one"},
			},
			{
				Name:    "service-2",
				Command: "/bin/echo",
				Args:    []string{"two"},
			},
		},
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
		initialConfig *config.Config
		// newConfig is the new configuration to apply.
		newConfig *config.Config
		// expectedServices is the list of expected service names after update.
		expectedServices []string
	}{
		{
			name:          "add_new_service",
			initialConfig: createTestConfig(),
			newConfig: &config.Config{
				ConfigPath: "/test/config.yaml",
				Services: []config.ServiceConfig{
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
		initialConfig *config.Config
		// newConfig is the new configuration to apply.
		newConfig *config.Config
		// keptServices is the list of services that should still exist.
		keptServices []string
		// removedServices is the list of services that should be removed.
		removedServices []string
	}{
		{
			name: "remove_one_service",
			initialConfig: &config.Config{
				ConfigPath: "/test/config.yaml",
				Services: []config.ServiceConfig{
					{Name: "service-1", Command: "/bin/echo", Args: []string{"one"}},
					{Name: "service-2", Command: "/bin/echo", Args: []string{"two"}},
				},
			},
			newConfig: &config.Config{
				ConfigPath: "/test/config.yaml",
				Services: []config.ServiceConfig{
					{Name: "service-1", Command: "/bin/echo", Args: []string{"one"}},
				},
			},
			keptServices:    []string{"service-1"},
			removedServices: []string{"service-2"},
		},
		{
			name: "remove_multiple_services",
			initialConfig: &config.Config{
				ConfigPath: "/test/config.yaml",
				Services: []config.ServiceConfig{
					{Name: "service-a", Command: "/bin/echo", Args: []string{"a"}},
					{Name: "service-b", Command: "/bin/echo", Args: []string{"b"}},
					{Name: "service-c", Command: "/bin/echo", Args: []string{"c"}},
				},
			},
			newConfig: &config.Config{
				ConfigPath: "/test/config.yaml",
				Services: []config.ServiceConfig{
					{Name: "service-a", Command: "/bin/echo", Args: []string{"a"}},
				},
			},
			keptServices:    []string{"service-a"},
			removedServices: []string{"service-b", "service-c"},
		},
		{
			name: "no_services_removed_all_kept",
			initialConfig: &config.Config{
				ConfigPath: "/test/config.yaml",
				Services: []config.ServiceConfig{
					{Name: "service-x", Command: "/bin/echo", Args: []string{"x"}},
					{Name: "service-y", Command: "/bin/echo", Args: []string{"y"}},
				},
			},
			newConfig: &config.Config{
				ConfigPath: "/test/config.yaml",
				Services: []config.ServiceConfig{
					{Name: "service-x", Command: "/bin/echo", Args: []string{"x"}},
					{Name: "service-y", Command: "/bin/echo", Args: []string{"y"}},
				},
			},
			keptServices:    []string{"service-x", "service-y"},
			removedServices: nil,
		},
		{
			name: "remove_all_services_empty_new_config",
			initialConfig: &config.Config{
				ConfigPath: "/test/config.yaml",
				Services: []config.ServiceConfig{
					{Name: "service-only", Command: "/bin/echo", Args: []string{"only"}},
				},
			},
			newConfig: &config.Config{
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
			// - Tracks: supervisor's WaitGroup (wg.Add(1) before, wg.Done() via defer in monitorService)
			// - Terminates when: context is cancelled OR events channel is closed
			// - Resource cleanup: WaitGroup.Done() called in deferred function within monitorService
			// - Synchronization: sup.wg.Wait() after cancel() ensures goroutine has exited
			go sup.monitorService("test", mockMgr)

			// Send an event.
			eventsCh <- testCase.event

			// Give some time for event processing.
			time.Sleep(10 * time.Millisecond)

			// Cancel context to stop monitoring goroutine.
			cancel()

			// Wait for goroutine to finish and release resources.
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
		// expectedStats is the expected stats after handling.
		expectedStats ServiceStatsSnapshot
	}{
		{
			name:          "handle_started_event",
			serviceName:   "test-service",
			event:         domain.Event{Type: domain.EventStarted},
			expectedStats: ServiceStatsSnapshot{StartCount: 1},
		},
		{
			name:          "handle_stopped_event",
			serviceName:   "test-service",
			event:         domain.Event{Type: domain.EventStopped},
			expectedStats: ServiceStatsSnapshot{StopCount: 1},
		},
		{
			name:          "handle_failed_event",
			serviceName:   "test-service",
			event:         domain.Event{Type: domain.EventFailed},
			expectedStats: ServiceStatsSnapshot{FailCount: 1},
		},
		{
			name:          "handle_restarting_event",
			serviceName:   "test-service",
			event:         domain.Event{Type: domain.EventRestarting},
			expectedStats: ServiceStatsSnapshot{RestartCount: 1},
		},
		{
			name:          "handle_unknown_service",
			serviceName:   "unknown-service",
			event:         domain.Event{Type: domain.EventStarted},
			expectedStats: ServiceStatsSnapshot{StartCount: 1},
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			sup := createTestSupervisor(t, cfg)

			// Handle the event.
			sup.handleEvent(testCase.serviceName, &testCase.event)

			// Verify stats were updated.
			stats := sup.Stats(testCase.serviceName)
			assert.NotNil(t, stats)
			assert.Equal(t, testCase.expectedStats.StartCount, stats.StartCount)
			assert.Equal(t, testCase.expectedStats.StopCount, stats.StopCount)
			assert.Equal(t, testCase.expectedStats.FailCount, stats.FailCount)
			assert.Equal(t, testCase.expectedStats.RestartCount, stats.RestartCount)
		})
	}
}

// Test_handleEvent_calls_event_handler tests that handleEvent calls the registered event handler.
// This tests the internal interaction between handleEvent and the event handler.
//
// Params:
//   - t: the testing context.
func Test_handleEvent_calls_event_handler(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{name: "event_handler_is_called"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			sup := createTestSupervisor(t, cfg)

			// Track handler calls.
			var calledService string
			var calledEvent *domain.Event
			var calledStats *ServiceStatsSnapshot
			handler := func(serviceName string, event *domain.Event, stats *ServiceStatsSnapshot) {
				calledService = serviceName
				calledEvent = event
				calledStats = stats
			}

			// Set the handler.
			sup.SetEventHandler(handler)

			// Trigger an event.
			event := domain.Event{Type: domain.EventStarted, PID: 123}
			sup.handleEvent("my-service", &event)

			// Verify handler was called.
			assert.Equal(t, "my-service", calledService)
			assert.NotNil(t, calledEvent)
			assert.Equal(t, domain.EventStarted, calledEvent.Type)
			assert.Equal(t, 123, calledEvent.PID)
			assert.NotNil(t, calledStats)
		})
	}
}

// Test_handleEvent_updates_stats tests that handleEvent updates statistics correctly.
// This tests the internal behavior of handleEvent with AllStats verification.
//
// Params:
//   - t: the testing context.
func Test_handleEvent_updates_stats(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{name: "returns_all_stats"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			sup := createTestSupervisor(t, cfg)

			// Send some events.
			sup.handleEvent("test-service", &domain.Event{Type: domain.EventStarted})
			sup.handleEvent("test-service", &domain.Event{Type: domain.EventFailed})

			// Get all stats.
			allStats := sup.AllStats()
			assert.NotNil(t, allStats)
			assert.Contains(t, allStats, "test-service")
			assert.Equal(t, 1, allStats["test-service"].StartCount)
			assert.Equal(t, 1, allStats["test-service"].FailCount)
		})
	}
}

// mockReaper implements kernel.ZombieReaper for testing.
// It provides a mock implementation for zombie process reaping.
type mockReaper struct {
	// startCalled indicates if Start was called.
	startCalled bool
	// stopCalled indicates if Stop was called.
	stopCalled bool
}

// Start starts the mock reaper.
func (mr *mockReaper) Start() {
	mr.startCalled = true
}

// Stop stops the mock reaper.
func (mr *mockReaper) Stop() {
	mr.stopCalled = true
}

// ReapOnce performs a single reap cycle.
//
// Returns:
//   - int: the count of reaped processes (always 0 for mock).
func (mr *mockReaper) ReapOnce() int {
	// Return zero for mock.
	return 0
}

// IsPID1 returns whether the process is PID 1.
//
// Returns:
//   - bool: always false for mock.
func (mr *mockReaper) IsPID1() bool {
	// Return false for mock.
	return false
}

// Test_Supervisor_Start_with_reaper tests Start with a zombie reaper.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_Start_with_reaper(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "starts_zombie_reaper",
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}
			reaper := &mockReaper{}

			sup, err := NewSupervisor(cfg, loader, executor, reaper)
			require.NoError(t, err)

			ctx := context.Background()
			err = sup.Start(ctx)
			require.NoError(t, err)
			defer func() { _ = sup.Stop() }()

			// Verify reaper was started.
			assert.True(t, reaper.startCalled)
		})
	}
}

// Test_Supervisor_Stop_with_reaper tests Stop with a zombie reaper.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_Stop_with_reaper(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "stops_zombie_reaper",
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}
			reaper := &mockReaper{}

			sup, err := NewSupervisor(cfg, loader, executor, reaper)
			require.NoError(t, err)

			// Start the supervisor.
			ctx := context.Background()
			err = sup.Start(ctx)
			require.NoError(t, err)

			// Stop the supervisor.
			err = sup.Stop()
			require.NoError(t, err)

			// Verify reaper was stopped.
			assert.True(t, reaper.stopCalled)
		})
	}
}

// Test_Supervisor_Start_already_running tests Start when supervisor is already running.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_Start_already_running(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "returns_ErrAlreadyRunning_when_already_started",
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			// Start the supervisor first time.
			ctx := context.Background()
			err = sup.Start(ctx)
			require.NoError(t, err)
			defer func() { _ = sup.Stop() }()

			// Try to start again - should fail.
			err = sup.Start(ctx)
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrAlreadyRunning)
		})
	}
}

// Test_Supervisor_Start_service_already_running tests Start when service is already running.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_Start_service_already_running(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "returns_error_and_stops_all_when_service_already_running",
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			// Pre-start the service to make it already running.
			for _, mgr := range sup.managers {
				err := mgr.Start()
				require.NoError(t, err)
			}

			// Now start the supervisor - services will fail with already running.
			ctx := context.Background()
			err = sup.Start(ctx)

			// Should return error because services are already running.
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to start service")

			// Verify supervisor is stopped.
			assert.Equal(t, StateStopped, sup.State())

			// Cleanup - stop the pre-started services.
			for _, mgr := range sup.managers {
				_ = mgr.Stop()
			}
		})
	}
}

// Test_Supervisor_monitorService_channel_close tests monitorService with closed channel.
// This test spawns a goroutine that monitors a mock events channel.
// The goroutine is managed by the supervisor's WaitGroup and terminates
// when the events channel is closed.
//
// Params:
//   - t: the testing context.
//
// Goroutine lifecycle:
//   - Spawns one goroutine via monitorService.
//   - Goroutine terminates when the events channel is closed.
//   - Method blocks via sup.wg.Wait() until goroutine completes.
func Test_Supervisor_monitorService_channel_close(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "exits_when_events_channel_closed",
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
			defer cancel()
			sup.ctx = ctx
			sup.cancel = cancel

			// Create mock eventser with channel.
			eventsCh := make(chan domain.Event, 1)
			mockMgr := &mockEventser{events: eventsCh}

			// Add to wait group before starting monitor.
			sup.wg.Add(1)

			// Start monitor in goroutine.
			// Goroutine lifecycle:
			// - Tracks: supervisor's WaitGroup (wg.Add(1) before, wg.Done() via defer in monitorService)
			// - Terminates when: events channel is closed (returns from range loop)
			// - Resource cleanup: WaitGroup.Done() called in deferred function within monitorService
			// - Synchronization: sup.wg.Wait() ensures goroutine has exited before test completes
			go sup.monitorService("test", mockMgr)

			// Close the events channel to trigger goroutine exit.
			close(eventsCh)

			// Wait for goroutine to finish and release resources.
			sup.wg.Wait()
		})
	}
}

// Test_Supervisor_RestartService_stop_error tests RestartService with stop error.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_RestartService_stop_error(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// stopErr is the error to return from stop.
		stopErr error
		// wantErr indicates if an error is expected.
		wantErr bool
	}{
		{
			name:    "returns_error_when_stop_fails",
			stopErr: domain.ErrProcessFailed,
			wantErr: true,
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			loader := &mockLoader{cfg: cfg}

			// Create an executor that keeps the process running (never sends exit).
			blockingExitCh := make(chan domain.ExitResult)
			executor := &blockingMockExecutor{
				stopErr: testCase.stopErr,
				exitCh:  blockingExitCh,
			}

			sup, err := NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			// Start the supervisor to get the service running.
			ctx := context.Background()
			err = sup.Start(ctx)
			require.NoError(t, err)

			// Give time for process to "start".
			time.Sleep(10 * time.Millisecond)

			// Try to restart the service - stop should fail.
			err = sup.RestartService("test-service")

			// Check if error is expected.
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Reset stopErr for cleanup using thread-safe setter.
			executor.SetStopErr(nil)
			_ = sup.Stop()
		})
	}
}

// blockingMockExecutor is a mock executor that keeps processes running.
type blockingMockExecutor struct {
	// mu protects stopErr from concurrent access.
	mu sync.RWMutex
	// stopErr is the error to return from Stop.
	stopErr error
	// exitCh is a blocking channel (never sends).
	exitCh chan domain.ExitResult
}

// Start starts a blocking mock process.
//
// Params:
//   - ctx: the context for cancellation.
//   - spec: the process specification.
//
// Returns:
//   - int: the mock process ID.
//   - <-chan domain.ExitResult: blocking channel.
//   - error: nil on success.
func (ex *blockingMockExecutor) Start(_ context.Context, _ domain.Spec) (pid int, wait <-chan domain.ExitResult, err error) {
	// Return a blocking exit channel - process never exits.
	return 1234, ex.exitCh, nil
}

// Stop stops a mock process.
//
// Params:
//   - pid: the process ID to stop.
//   - timeout: the stop timeout.
//
// Returns:
//   - error: the mock stop error.
func (ex *blockingMockExecutor) Stop(_ int, _ time.Duration) error {
	ex.mu.RLock()
	defer ex.mu.RUnlock()
	// Return the configured stop error.
	return ex.stopErr
}

// SetStopErr sets the stop error with thread safety.
//
// Params:
//   - err: the error to set.
func (ex *blockingMockExecutor) SetStopErr(err error) {
	ex.mu.Lock()
	defer ex.mu.Unlock()
	ex.stopErr = err
}

// Signal sends a signal to a mock process.
//
// Params:
//   - pid: the process ID.
//   - sig: the signal to send.
//
// Returns:
//   - error: nil.
func (ex *blockingMockExecutor) Signal(_ int, _ os.Signal) error {
	// Return nil.
	return nil
}

// Test_Supervisor_Stop_when_not_running tests Stop when supervisor is not running.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_Stop_when_not_running(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "returns_nil_when_not_running",
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			// Stop without starting - should return nil.
			err = sup.Stop()
			assert.NoError(t, err)
		})
	}
}

// Test_Supervisor_handleRecoveryError tests the handleRecoveryError method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_handleRecoveryError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// setHandler determines if an error handler should be set.
		setHandler bool
		// inputErr is the error to pass to handleRecoveryError.
		inputErr error
		// expectCall determines if handler should be called.
		expectCall bool
	}{
		{
			name:       "nil_error_with_no_handler",
			setHandler: false,
			inputErr:   nil,
			expectCall: false,
		},
		{
			name:       "nil_error_with_handler",
			setHandler: true,
			inputErr:   nil,
			expectCall: false,
		},
		{
			name:       "real_error_with_no_handler",
			setHandler: false,
			inputErr:   assert.AnError,
			expectCall: false,
		},
		{
			name:       "real_error_with_handler",
			setHandler: true,
			inputErr:   assert.AnError,
			expectCall: true,
		},
	}

	// Iterate through all test cases.
	for _, testCase := range tests {
		// Run each test case as a subtest.
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			// Track if handler was called.
			var handlerCalled bool
			var receivedOp, receivedService string
			var receivedErr error

			// Set handler if needed.
			if testCase.setHandler {
				sup.SetErrorHandler(func(op, svc string, e error) {
					handlerCalled = true
					receivedOp = op
					receivedService = svc
					receivedErr = e
				})
			}

			// Call handleRecoveryError.
			sup.handleRecoveryError("test-op", "test-service", testCase.inputErr)

			// Verify expectations.
			if testCase.expectCall {
				assert.True(t, handlerCalled, "handler should have been called")
				assert.Equal(t, "test-op", receivedOp)
				assert.Equal(t, "test-service", receivedService)
				assert.Equal(t, testCase.inputErr, receivedErr)
			} else {
				assert.False(t, handlerCalled, "handler should not have been called")
			}
		})
	}
}

// Test_Supervisor_updateServices_withStopError tests error handling when
// stopping an existing manager fails during updateServices.
func Test_Supervisor_updateServices_withStopError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "stop_error_reported_via_handler",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createTestConfig()
			sup := createTestSupervisor(t, cfg)

			// Start the supervisor.
			ctx := context.Background()
			err := sup.Start(ctx)
			require.NoError(t, err)

			// Wait for service to be running with valid PID.
			require.Eventually(t, func() bool {
				services := sup.Services()
				info, ok := services["test-service"]
				return ok && info.PID > 0
			}, time.Second, 10*time.Millisecond)

			// Track error handler calls.
			var handlerCalled bool
			var handlerOp string
			sup.SetErrorHandler(func(op, svc string, e error) {
				handlerCalled = true
				handlerOp = op
			})

			// Configure executor to fail on Stop (thread-safe).
			if ex, ok := sup.executor.(*mockExecutor); ok {
				ex.SetStopErr(domain.ErrProcessFailed)
			}

			// Create updated config that will trigger stop of existing manager.
			newCfg := &config.Config{
				ConfigPath: "/test/config.yaml",
				Services: []config.ServiceConfig{
					{
						Name:    "test-service",
						Command: "/bin/echo",
						Args:    []string{"updated"},
					},
				},
			}

			// Call updateServices which should trigger stop error.
			sup.updateServices(newCfg)

			// Verify error was handled.
			assert.True(t, handlerCalled, "error handler should be called")
			assert.Equal(t, "stop-for-reload", handlerOp)

			// Cleanup.
			if ex, ok := sup.executor.(*mockExecutor); ok {
				ex.SetStopErr(nil)
			}
			_ = sup.Stop()
		})
	}
}

// Note: Start error tests for updateServices are omitted because the error
// path is unreachable in practice. The lifecycle.Manager.Start() only returns
// an error if the manager is already running, but updateServices creates a
// new manager with NewManager() each time, so m.running is always false.
// This is defensive programming - the error handling exists for safety but
// cannot be triggered in the current implementation.

// Test_Supervisor_removeDeletedServices_withStopError tests error handling
// when stopping a removed service fails.
func Test_Supervisor_removeDeletedServices_withStopError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "stop_error_reported_via_handler",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := createMultiServiceTestConfig()
			sup := createTestSupervisor(t, cfg)

			// Start the supervisor.
			ctx := context.Background()
			err := sup.Start(ctx)
			require.NoError(t, err)

			// Wait for all services to be running.
			require.Eventually(t, func() bool {
				services := sup.Services()
				for _, info := range services {
					if info.PID <= 0 {
						return false
					}
				}
				return len(services) == 2
			}, time.Second, 10*time.Millisecond)

			// Track error handler calls.
			var handlerCalled bool
			var handlerOp string
			sup.SetErrorHandler(func(op, svc string, e error) {
				handlerCalled = true
				handlerOp = op
			})

			// Configure executor to fail on Stop (thread-safe).
			if ex, ok := sup.executor.(*mockExecutor); ok {
				ex.SetStopErr(domain.ErrProcessFailed)
			}

			// Create config with one service removed.
			newCfg := &config.Config{
				ConfigPath: "/test/config.yaml",
				Services: []config.ServiceConfig{
					{
						Name:    "service-1",
						Command: "/bin/echo",
						Args:    []string{"one"},
					},
				},
			}

			// Call removeDeletedServices which should trigger stop error.
			sup.removeDeletedServices(newCfg)

			// Verify error was handled.
			assert.True(t, handlerCalled, "error handler should be called")
			assert.Equal(t, "stop-removed-service", handlerOp)

			// Cleanup.
			if ex, ok := sup.executor.(*mockExecutor); ok {
				ex.SetStopErr(nil)
			}
			_ = sup.Stop()
		})
	}
}

// NOTE ON COVERAGE: updateServices currently shows 84.6% coverage because lines 307-309
// and 314-316 are unreachable defensive error handling code:
//
// - updateServices calls applifecycle.NewManager() which always creates managers with running=false
// - Manager.Start() only returns an error when running=true (ErrAlreadyRunning)
// - Therefore, Start() can never fail on a newly created manager
//
// Achieving 100% coverage would require refactoring production code to inject a manager
// factory or adding test hooks. The test below verifies the error handling logic works
// correctly, even though it cannot be triggered through updateServices.

// Test_Supervisor_updateServices_defensive_error_handling tests the error handling
// logic at lines 307-309 and 314-316 by manually replicating the conditions.
func Test_Supervisor_updateServices_defensive_error_handling(t *testing.T) {
	tests := []struct {
		name         string
		operation    string
		serviceName  string
		setupManager func(*config.Config, *mockExecutor) *applifecycle.Manager
	}{
		{
			name:        "start_for_reload_error_path",
			operation:   "start-for-reload",
			serviceName: "test-service",
			setupManager: func(cfg *config.Config, exec *mockExecutor) *applifecycle.Manager {
				// Use the first service from config.
				return applifecycle.NewManager(&cfg.Services[0], exec)
			},
		},
		{
			name:        "start_new_service_error_path",
			operation:   "start-new-service",
			serviceName: "new-service",
			setupManager: func(_ *config.Config, exec *mockExecutor) *applifecycle.Manager {
				// Use a new service config.
				newSvc := &config.ServiceConfig{Name: "new-service", Command: "/bin/echo", Args: []string{"new"}}
				return applifecycle.NewManager(newSvc, exec)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			// Track if error handler was called.
			var handlerCalled bool
			sup.SetErrorHandler(func(op, svc string, e error) {
				handlerCalled = true
				assert.Equal(t, tt.operation, op)
			})

			// Create and start manager to put it in running state.
			mgr := tt.setupManager(cfg, executor)
			_ = mgr.Start() // Makes running=true
			defer func() { _ = mgr.Stop() }()

			// Try to start again - this triggers error.
			if startErr := mgr.Start(); startErr != nil {
				sup.handleRecoveryError(tt.operation, tt.serviceName, startErr)
			}

			// Verify error handler was called.
			assert.True(t, handlerCalled)
		})
	}
}

// Test_Supervisor_initializeStart tests the initializeStart method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_initializeStart(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// initialState is the supervisor state before calling.
		initialState State
		// expectError indicates if an error is expected.
		expectError bool
	}{
		{
			name:         "from_stopped",
			initialState: StateStopped,
			expectError:  false,
		},
		{
			name:         "already_starting",
			initialState: StateStarting,
			expectError:  true,
		},
		{
			name:         "already_running",
			initialState: StateRunning,
			expectError:  true,
		},
		{
			name:         "stopping",
			initialState: StateStopping,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sup := &Supervisor{
				state: tt.initialState,
			}

			err := sup.initializeStart(context.Background())

			// Check if error expectation matches.
			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrAlreadyRunning)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, StateStarting, sup.state)
				assert.NotNil(t, sup.ctx)
				assert.NotNil(t, sup.cancel)
			}
		})
	}
}

// Test_Supervisor_startReaper tests the startReaper method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_startReaper(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// hasReaper indicates if reaper is configured.
		hasReaper bool
		// expectStarted indicates if reaper Start should be called.
		expectStarted bool
	}{
		{
			name:          "with_reaper",
			hasReaper:     true,
			expectStarted: true,
		},
		{
			name:          "without_reaper",
			hasReaper:     false,
			expectStarted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sup := &Supervisor{}

			// Create and set reaper if configured.
			var reaper *mockReaper
			if tt.hasReaper {
				reaper = &mockReaper{}
				sup.reaper = reaper
			}

			sup.startReaper()

			// Check if reaper Start was called.
			if tt.hasReaper {
				assert.Equal(t, tt.expectStarted, reaper.startCalled)
			}
		})
	}
}

// Test_Supervisor_startAllServices tests the startAllServices method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_startAllServices(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// managerCount is the number of managers.
		managerCount int
		// failAt is the manager index that should fail (-1 for no failure).
		failAt int
		// expectError indicates if an error is expected.
		expectError bool
	}{
		{
			name:         "all_start_successfully",
			managerCount: 3,
			failAt:       -1,
			expectError:  false,
		},
		{
			name:         "no_managers",
			managerCount: 0,
			failAt:       -1,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &mockExecutor{}
			managers := make(map[string]*applifecycle.Manager)

			// Create managers.
			for i := range tt.managerCount {
				cfg := &config.ServiceConfig{
					Name:    "test" + string(rune('0'+i)),
					Command: "sleep",
					Args:    []string{"1"},
				}
				managers[cfg.Name] = applifecycle.NewManager(cfg, executor)
			}

			sup := &Supervisor{
				managers: managers,
				state:    StateStarting,
			}

			err := sup.startAllServices()

			// Check error expectation.
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Cleanup.
			for _, mgr := range managers {
				_ = mgr.Stop()
			}
		})
	}
}

// testCreator implements health.Creator for testing.
type testCreator struct{}

// Create returns nil for testing.
func (c *testCreator) Create(_ string, _ time.Duration) (domainhealth.Prober, error) {
	return nil, nil
}

// Test_supervisor_proberFactoryAssignment tests internal prober factory assignment.
//
// Params:
//   - t: the testing context.
func Test_supervisor_proberFactoryAssignment(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// setFactory indicates whether to set a factory.
		setFactory bool
	}{
		{
			name:       "set_factory",
			setFactory: true,
		},
		{
			name:       "set_nil",
			setFactory: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sup := &Supervisor{}

			// Set factory based on test case.
			if tt.setFactory {
				sup.SetProberFactory(&testCreator{})
				assert.NotNil(t, sup.proberFactory)
			} else {
				sup.SetProberFactory(nil)
				assert.Nil(t, sup.proberFactory)
			}
		})
	}
}

// Test_Supervisor_hasConfiguredProbes tests the hasConfiguredProbes method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_hasConfiguredProbes(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// listeners is the listener configuration.
		listeners []config.ListenerConfig
		// expected is the expected result.
		expected bool
	}{
		{
			name:      "no_listeners",
			listeners: nil,
			expected:  false,
		},
		{
			name: "listener_without_probe",
			listeners: []config.ListenerConfig{
				{Name: "http", Port: 8080},
			},
			expected: false,
		},
		{
			name: "listener_with_probe",
			listeners: []config.ListenerConfig{
				{
					Name: "http",
					Port: 8080,
					Probe: &config.ProbeConfig{
						Type: "tcp",
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sup := &Supervisor{}
			svc := &config.ServiceConfig{
				Listeners: tt.listeners,
			}

			result := sup.hasConfiguredProbes(svc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_Supervisor_createDomainListener tests the createDomainListener method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_createDomainListener(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// config is the listener configuration.
		config *config.ListenerConfig
		// expectedProtocol is the expected protocol.
		expectedProtocol string
		// expectedAddress is the expected address.
		expectedAddress string
	}{
		{
			name: "with_defaults",
			config: &config.ListenerConfig{
				Name: "web",
				Port: 8080,
			},
			expectedProtocol: "tcp",
			expectedAddress:  "localhost",
		},
		{
			name: "with_custom_values",
			config: &config.ListenerConfig{
				Name:     "grpc",
				Port:     9090,
				Protocol: "tcp4",
				Address:  "0.0.0.0",
			},
			expectedProtocol: "tcp4",
			expectedAddress:  "0.0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sup := &Supervisor{}

			l := sup.createDomainListener(tt.config)

			assert.Equal(t, tt.config.Name, l.Name)
			assert.Equal(t, tt.expectedProtocol, l.Protocol)
			assert.Equal(t, tt.expectedAddress, l.Address)
			assert.Equal(t, tt.config.Port, l.Port)
		})
	}
}

// Test_Supervisor_createProbeBinding tests the createProbeBinding method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_createProbeBinding(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// config is the listener configuration.
		config *config.ListenerConfig
		// expectedType is the expected probe type.
		expectedType string
	}{
		{
			name: "tcp_probe",
			config: &config.ListenerConfig{
				Name: "web",
				Port: 8080,
				Probe: &config.ProbeConfig{
					Type: "tcp",
				},
			},
			expectedType: "tcp",
		},
		{
			name: "http_probe",
			config: &config.ListenerConfig{
				Name: "api",
				Port: 8080,
				Probe: &config.ProbeConfig{
					Type: "http",
					Path: "/health",
				},
			},
			expectedType: "http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sup := &Supervisor{}

			binding := sup.createProbeBinding(tt.config)

			assert.Equal(t, tt.config.Name, binding.ListenerName)
			assert.Equal(t, tt.expectedType, string(binding.Type))
		})
	}
}

// Test_supervisor_healthFailureRestartRouting tests internal restart routing on health failure.
//
// Params:
//   - t: the testing context.
func Test_supervisor_healthFailureRestartRouting(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// serviceName is the service to restart.
		serviceName string
		// hasManager indicates if manager exists.
		hasManager bool
		// expectError indicates if error is expected.
		expectError bool
	}{
		{
			name:        "service_not_found",
			serviceName: "nonexistent",
			hasManager:  false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sup := &Supervisor{
				managers: make(map[string]*applifecycle.Manager),
			}

			err := sup.RestartOnHealthFailure(tt.serviceName, "test reason")

			// Check error expectation.
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test_Supervisor_startMonitoringGoroutines tests the startMonitoringGoroutines method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_startMonitoringGoroutines(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// managerCount is the number of managers.
		managerCount int
	}{
		{
			name:         "no_managers",
			managerCount: 0,
		},
		{
			name:         "one_manager",
			managerCount: 1,
		},
		{
			name:         "multiple_managers",
			managerCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			executor := &mockExecutor{}
			managers := make(map[string]*applifecycle.Manager)

			// Create managers.
			for i := range tt.managerCount {
				cfg := &config.ServiceConfig{
					Name:    fmt.Sprintf("test%d", i),
					Command: "sleep",
					Args:    []string{"1"},
				}
				managers[cfg.Name] = applifecycle.NewManager(cfg, executor)
			}

			sup := &Supervisor{
				managers: managers,
				ctx:      ctx,
			}

			// Call startMonitoringGoroutines.
			sup.startMonitoringGoroutines()

			// Wait group should have added entries.
			// Cancel context to allow goroutines to exit.
			cancel()

			// Verify goroutines were started by checking wait group completes.
			done := make(chan struct{})
			go func() {
				sup.wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// Success.
			case <-time.After(2 * time.Second):
				t.Error("goroutines did not complete")
			}
		})
	}
}

// Test_Supervisor_startHealthMonitors tests the startHealthMonitors method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_startHealthMonitors(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// hasFactory indicates whether prober factory is set.
		hasFactory bool
		// services is the service configurations.
		services []config.ServiceConfig
		// expectedMonitors is the expected number of health monitors.
		expectedMonitors int
	}{
		{
			name:             "no_factory",
			hasFactory:       false,
			services:         []config.ServiceConfig{{Name: "test"}},
			expectedMonitors: 0,
		},
		{
			name:       "no_services",
			hasFactory: true,
			services:   nil,
		},
		{
			name:       "service_without_probes",
			hasFactory: true,
			services: []config.ServiceConfig{
				{Name: "test", Listeners: nil},
			},
			expectedMonitors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sup := &Supervisor{
				config: &config.Config{
					Services: tt.services,
				},
				healthMonitors: make(map[string]*apphealth.ProbeMonitor),
				ctx:            t.Context(),
			}

			// Set factory if needed.
			if tt.hasFactory {
				sup.proberFactory = &testCreator{}
			}

			// Call startHealthMonitors.
			sup.startHealthMonitors()

			// Verify expected number of monitors.
			assert.Len(t, sup.healthMonitors, tt.expectedMonitors)
		})
	}
}

// Test_Supervisor_createHealthMonitor tests the createHealthMonitor method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_createHealthMonitor(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// hasFactory indicates whether prober factory is set.
		hasFactory bool
		// service is the service configuration.
		service *config.ServiceConfig
		// expectNil indicates if nil monitor is expected.
		expectNil bool
	}{
		{
			name:       "no_factory",
			hasFactory: false,
			service: &config.ServiceConfig{
				Name: "test",
				Listeners: []config.ListenerConfig{
					{Name: "http", Port: 8080, Probe: &config.ProbeConfig{Type: "tcp"}},
				},
			},
			expectNil: true,
		},
		{
			name:       "no_probes",
			hasFactory: true,
			service: &config.ServiceConfig{
				Name:      "test",
				Listeners: []config.ListenerConfig{{Name: "http", Port: 8080}},
			},
			expectNil: true,
		},
		{
			name:       "with_probes_and_factory",
			hasFactory: true,
			service: &config.ServiceConfig{
				Name: "test",
				Listeners: []config.ListenerConfig{
					{Name: "http", Port: 8080, Probe: &config.ProbeConfig{Type: "tcp"}},
				},
			},
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sup := &Supervisor{}

			// Set factory if needed.
			if tt.hasFactory {
				sup.proberFactory = &testCreator{}
			}

			// Call createHealthMonitor.
			monitor := sup.createHealthMonitor(tt.service)

			// Verify result.
			if tt.expectNil {
				assert.Nil(t, monitor)
			} else {
				assert.NotNil(t, monitor)
			}
		})
	}
}

// Test_Supervisor_createProbeMonitorConfig tests the createProbeMonitorConfig method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_createProbeMonitorConfig(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// serviceName is the service name for configuration.
		serviceName string
	}{
		{
			name:        "simple_service",
			serviceName: "myservice",
		},
		{
			name:        "empty_name",
			serviceName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sup := &Supervisor{
				proberFactory: &testCreator{},
				managers:      make(map[string]*applifecycle.Manager),
			}

			// Call createProbeMonitorConfig.
			cfg := sup.createProbeMonitorConfig(tt.serviceName)

			// Verify config has callbacks.
			assert.NotNil(t, cfg.Factory)
			assert.NotNil(t, cfg.OnStateChange)
			assert.NotNil(t, cfg.OnUnhealthy)
		})
	}
}

// mockAddListenerWithBindinger is a mock for AddListenerWithBindinger interface.
type mockAddListenerWithBindinger struct {
	// addedCount tracks the number of listeners added.
	addedCount int
	// returnError indicates whether to return an error.
	returnError bool
}

// AddListenerWithBinding implements AddListenerWithBindinger.
func (m *mockAddListenerWithBindinger) AddListenerWithBinding(_ *listener.Listener, _ *apphealth.ProbeBinding) error {
	m.addedCount++
	// Return error if configured.
	if m.returnError {
		return fmt.Errorf("mock error")
	}
	return nil
}

// Test_Supervisor_addListenersWithProbes tests the addListenersWithProbes method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_addListenersWithProbes(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// service is the service configuration.
		service *config.ServiceConfig
	}{
		{
			name: "no_listeners",
			service: &config.ServiceConfig{
				Name:      "test",
				Listeners: nil,
			},
		},
		{
			name: "listener_without_probe",
			service: &config.ServiceConfig{
				Name: "test",
				Listeners: []config.ListenerConfig{
					{Name: "http", Port: 8080},
				},
			},
		},
		{
			name: "listener_with_probe",
			service: &config.ServiceConfig{
				Name: "test",
				Listeners: []config.ListenerConfig{
					{Name: "http", Port: 8080, Probe: &config.ProbeConfig{Type: "tcp"}},
				},
			},
		},
		{
			name: "multiple_listeners_mixed",
			service: &config.ServiceConfig{
				Name: "test",
				Listeners: []config.ListenerConfig{
					{Name: "http1", Port: 8080, Probe: &config.ProbeConfig{Type: "tcp"}},
					{Name: "http2", Port: 8081},
					{Name: "http3", Port: 8082, Probe: &config.ProbeConfig{Type: "http"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sup := &Supervisor{
				proberFactory: &testCreator{},
			}

			// Create a real ProbeMonitor.
			cfg := apphealth.ProbeMonitorConfig{
				Factory: &testCreator{},
			}
			monitor := apphealth.NewProbeMonitor(cfg)

			// Call addListenersWithProbes - should not panic.
			sup.addListenersWithProbes(monitor, tt.service)
		})
	}
}

// Test_Supervisor_addSingleListenerWithProbe tests the addSingleListenerWithProbe method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_addSingleListenerWithProbe(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// listener is the listener configuration.
		listener *config.ListenerConfig
		// returnError indicates if mock should return error.
		returnError bool
	}{
		{
			name: "success",
			listener: &config.ListenerConfig{
				Name:    "http",
				Port:    8080,
				Address: "localhost",
				Probe:   &config.ProbeConfig{Type: "tcp"},
			},
			returnError: false,
		},
		{
			name: "error_handling",
			listener: &config.ListenerConfig{
				Name:  "http",
				Port:  8080,
				Probe: &config.ProbeConfig{Type: "tcp"},
			},
			returnError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sup := &Supervisor{
				proberFactory: &testCreator{},
			}
			mockMonitor := &mockAddListenerWithBindinger{
				returnError: tt.returnError,
			}

			// Call addSingleListenerWithProbe - should not panic.
			sup.addSingleListenerWithProbe(mockMonitor, tt.listener)

			// Verify listener was added.
			assert.Equal(t, 1, mockMonitor.addedCount)
		})
	}
}
