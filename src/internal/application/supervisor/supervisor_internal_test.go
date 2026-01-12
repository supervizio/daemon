// Package supervisor provides internal tests for supervisor.go.
// It tests the internal implementation of the Supervisor type.
package supervisor

import (
	"context"
	"os"
	"sync"
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
		expectedStats ServiceStats
	}{
		{
			name:          "handle_started_event",
			serviceName:   "test-service",
			event:         domain.Event{Type: domain.EventStarted},
			expectedStats: ServiceStats{StartCount: 1},
		},
		{
			name:          "handle_stopped_event",
			serviceName:   "test-service",
			event:         domain.Event{Type: domain.EventStopped},
			expectedStats: ServiceStats{StopCount: 1},
		},
		{
			name:          "handle_failed_event",
			serviceName:   "test-service",
			event:         domain.Event{Type: domain.EventFailed},
			expectedStats: ServiceStats{FailCount: 1},
		},
		{
			name:          "handle_restarting_event",
			serviceName:   "test-service",
			event:         domain.Event{Type: domain.EventRestarting},
			expectedStats: ServiceStats{RestartCount: 1},
		},
		{
			name:          "handle_unknown_service",
			serviceName:   "unknown-service",
			event:         domain.Event{Type: domain.EventStarted},
			expectedStats: ServiceStats{StartCount: 1},
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
			handler := func(serviceName string, event *domain.Event) {
				calledService = serviceName
				calledEvent = event
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

// mockReaper implements ports.ZombieReaper for testing.
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
