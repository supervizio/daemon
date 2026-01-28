// Package supervisor provides internal tests for supervisor.go.
// It tests internal implementation details using white-box testing.
package supervisor

import (
	"context"
	"time"
	"testing"

	"github.com/stretchr/testify/assert"
	apphealth "github.com/kodflow/daemon/internal/application/health"
	applifecycle "github.com/kodflow/daemon/internal/application/lifecycle"
	appmetrics "github.com/kodflow/daemon/internal/application/metrics"
	domainconfig "github.com/kodflow/daemon/internal/domain/config"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/health"
	domain "github.com/kodflow/daemon/internal/domain/process"
)

// Test_Supervisor_getOrCreateStats tests the getOrCreateStats method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_getOrCreateStats(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// serviceName is the service name.
		serviceName string
		// existingStats indicates if stats already exist.
		existingStats bool
	}{
		{
			name:          "creates_new_stats_when_not_found",
			serviceName:   "new-service",
			existingStats: false,
		},
		{
			name:          "returns_existing_stats_when_found",
			serviceName:   "existing-service",
			existingStats: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				stats: make(map[string]*ServiceStats),
			}

			// Pre-create stats if requested.
			if tt.existingStats {
				s.stats[tt.serviceName] = NewServiceStats()
				s.stats[tt.serviceName].IncrementStart()
			}

			stats := s.getOrCreateStats(tt.serviceName)

			// Verify stats was returned.
			require.NotNil(t, stats)

			// Verify correct stats instance.
			if tt.existingStats {
				assert.Equal(t, 1, stats.StartCount())
			} else {
				assert.Equal(t, 0, stats.StartCount())
			}
		})
	}
}

// Test_Supervisor_updateStatsForEvent tests the updateStatsForEvent method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_updateStatsForEvent(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// eventType is the event type.
		eventType domain.EventType
		// expectedStartCount is the expected start count after update.
		expectedStartCount int
		// expectedStopCount is the expected stop count after update.
		expectedStopCount int
		// expectedFailCount is the expected fail count after update.
		expectedFailCount int
		// expectedRestartCount is the expected restart count after update.
		expectedRestartCount int
	}{
		{
			name:                 "started_event_increments_start",
			eventType:            domain.EventStarted,
			expectedStartCount:   1,
			expectedStopCount:    0,
			expectedFailCount:    0,
			expectedRestartCount: 0,
		},
		{
			name:                 "stopped_event_increments_stop",
			eventType:            domain.EventStopped,
			expectedStartCount:   0,
			expectedStopCount:    1,
			expectedFailCount:    0,
			expectedRestartCount: 0,
		},
		{
			name:                 "failed_event_increments_fail",
			eventType:            domain.EventFailed,
			expectedStartCount:   0,
			expectedStopCount:    0,
			expectedFailCount:    1,
			expectedRestartCount: 0,
		},
		{
			name:                 "restarting_event_increments_restart",
			eventType:            domain.EventRestarting,
			expectedStartCount:   0,
			expectedStopCount:    0,
			expectedFailCount:    0,
			expectedRestartCount: 1,
		},
		{
			name:                 "exhausted_event_increments_fail",
			eventType:            domain.EventExhausted,
			expectedStartCount:   0,
			expectedStopCount:    0,
			expectedFailCount:    1,
			expectedRestartCount: 0,
		},
		{
			name:                 "healthy_event_no_increment",
			eventType:            domain.EventHealthy,
			expectedStartCount:   0,
			expectedStopCount:    0,
			expectedFailCount:    0,
			expectedRestartCount: 0,
		},
		{
			name:                 "unhealthy_event_no_increment",
			eventType:            domain.EventUnhealthy,
			expectedStartCount:   0,
			expectedStopCount:    0,
			expectedFailCount:    0,
			expectedRestartCount: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				stats: make(map[string]*ServiceStats),
			}

			stats := NewServiceStats()
			event := &domain.Event{Type: tt.eventType}

			s.updateStatsForEvent(stats, event)

			// Verify counts.
			assert.Equal(t, tt.expectedStartCount, stats.StartCount())
			assert.Equal(t, tt.expectedStopCount, stats.StopCount())
			assert.Equal(t, tt.expectedFailCount, stats.FailCount())
			assert.Equal(t, tt.expectedRestartCount, stats.RestartCount())
		})
	}
}

// Test_Supervisor_updateHealthMonitor tests the updateHealthMonitor method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_updateHealthMonitor(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// eventType is the event type.
		eventType domain.EventType
		// hasMonitor indicates if health monitor exists.
		hasMonitor bool
		// expectStateUpdate indicates if state update is expected.
		expectStateUpdate bool
	}{
		{
			name:              "started_event_updates_state_when_monitor_exists",
			eventType:         domain.EventStarted,
			hasMonitor:        true,
			expectStateUpdate: true,
		},
		{
			name:              "stopped_event_updates_state_when_monitor_exists",
			eventType:         domain.EventStopped,
			hasMonitor:        true,
			expectStateUpdate: true,
		},
		{
			name:              "no_update_when_monitor_missing",
			eventType:         domain.EventStarted,
			hasMonitor:        false,
			expectStateUpdate: false,
		},
		{
			name:              "healthy_event_no_state_update",
			eventType:         domain.EventHealthy,
			hasMonitor:        true,
			expectStateUpdate: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				healthMonitors: make(map[string]*apphealth.ProbeMonitor),
			}

			if tt.hasMonitor {
				// Create a health monitor.
				monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{})
				s.healthMonitors["test-service"] = monitor
			}

			event := &domain.Event{Type: tt.eventType}

			// Should not panic.
			s.updateHealthMonitor("test-service", event)
		})
	}
}

// Test_Supervisor_updateMetricsTracker tests the updateMetricsTracker method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_updateMetricsTracker(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// eventType is the event type.
		eventType domain.EventType
		// pid is the process ID.
		pid int
		// hasTracker indicates if metrics tracker exists.
		hasTracker bool
	}{
		{
			name:       "started_event_tracks_when_tracker_exists",
			eventType:  domain.EventStarted,
			pid:        1234,
			hasTracker: true,
		},
		{
			name:       "stopped_event_untracks_when_tracker_exists",
			eventType:  domain.EventStopped,
			pid:        1234,
			hasTracker: true,
		},
		{
			name:       "no_action_when_tracker_missing",
			eventType:  domain.EventStarted,
			pid:        1234,
			hasTracker: false,
		},
		{
			name:       "started_with_invalid_pid_no_track",
			eventType:  domain.EventStarted,
			pid:        0,
			hasTracker: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{}

			if tt.hasTracker {
				// Create a mock metrics tracker.
				tracker := appmetrics.NewTracker(nil)
				_ = tracker.Start(context.Background())
				defer tracker.Stop()
				s.metricsTracker = tracker
			}

			event := &domain.Event{
				Type: tt.eventType,
				PID:  tt.pid,
			}

			// Should not panic.
			s.updateMetricsTracker("test-service", event)
		})
	}
}

// Test_Supervisor_getStatsSnapshot tests the getStatsSnapshot method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_getStatsSnapshot(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// stats is the input stats (nil or valid).
		stats *ServiceStats
		// expectNil indicates if result should be nil.
		expectNil bool
	}{
		{
			name:      "returns_nil_for_nil_stats",
			stats:     nil,
			expectNil: true,
		},
		{
			name:      "returns_snapshot_for_valid_stats",
			stats:     NewServiceStats(),
			expectNil: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{}

			result := s.getStatsSnapshot(tt.stats)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

// Test_Supervisor_callEventHandler tests the callEventHandler method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_callEventHandler(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// hasHandler indicates if event handler is set.
		hasHandler bool
		// expectCall indicates if handler should be called.
		expectCall bool
	}{
		{
			name:       "calls_handler_when_set",
			hasHandler: true,
			expectCall: true,
		},
		{
			name:       "no_call_when_handler_not_set",
			hasHandler: false,
			expectCall: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			called := false
			s := &Supervisor{}

			if tt.hasHandler {
				s.eventHandler = func(_ string, _ *domain.Event, _ *ServiceStatsSnapshot) {
					called = true
				}
			}

			event := &domain.Event{Type: domain.EventStarted}
			snap := &ServiceStatsSnapshot{}

			s.callEventHandler("test-service", event, snap)

			assert.Equal(t, tt.expectCall, called)
		})
	}
}

// Test_Supervisor_enrichSnapshotWithConfig tests the enrichSnapshotWithConfig method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_enrichSnapshotWithConfig(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// serviceName is the service name.
		serviceName string
		// hasConfig indicates if config exists.
		hasConfig bool
	}{
		{
			name:        "enriches_when_config_exists",
			serviceName: "test-service",
			hasConfig:   true,
		},
		{
			name:        "no_enrich_when_config_missing",
			serviceName: "missing-service",
			hasConfig:   false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				config: &domainconfig.Config{},
			}

			if tt.hasConfig {
				s.config.Services = []domainconfig.ServiceConfig{
					{Name: tt.serviceName, Command: "/bin/echo"},
				}
			}

			snap := &ServiceSnapshotForTUI{Name: tt.serviceName}

			s.enrichSnapshotWithConfig(snap, tt.serviceName)

			// Verify method completes without panic.
			assert.NotNil(t, snap)
		})
	}
}

// Test_Supervisor_enrichSnapshotWithMetrics tests the enrichSnapshotWithMetrics method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_enrichSnapshotWithMetrics(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// hasTracker indicates if metrics tracker exists.
		hasTracker bool
	}{
		{
			name:       "enriches_when_tracker_exists",
			hasTracker: true,
		},
		{
			name:       "no_enrich_when_tracker_missing",
			hasTracker: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{}

			if tt.hasTracker {
				tracker := appmetrics.NewTracker(nil)
				s.metricsTracker = tracker
			}

			snap := &ServiceSnapshotForTUI{Name: "test-service"}

			s.enrichSnapshotWithMetrics(snap, "test-service")

			// Verify method completes without panic.
			assert.NotNil(t, snap)
		})
	}
}

// Test_Supervisor_buildListenerSnapshots tests the buildListenerSnapshots method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_buildListenerSnapshots(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// hasListeners indicates if service has listeners.
		hasListeners bool
	}{
		{
			name:         "builds_empty_when_no_listeners",
			hasListeners: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				healthMonitors: make(map[string]*apphealth.ProbeMonitor),
			}

			svc := &domainconfig.ServiceConfig{Name: "test"}
			ports := []int{}

			result := s.buildListenerSnapshots(svc, ports)

			// Verify method returns a valid slice.
			assert.NotNil(t, result)
			assert.Empty(t, result)
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
		// initialState is the supervisor's initial state.
		initialState State
		// expectError indicates if an error is expected.
		expectError bool
	}{
		{
			name:         "initializes_from_stopped_state",
			initialState: StateStopped,
			expectError:  false,
		},
		{
			name:         "returns_error_when_already_running",
			initialState: StateRunning,
			expectError:  true,
		},
		{
			name:         "returns_error_when_starting",
			initialState: StateStarting,
			expectError:  true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				state: tt.initialState,
			}
			ctx := context.Background()

			err := s.initializeStart(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, ErrAlreadyRunning, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, StateStarting, s.state)
				assert.NotNil(t, s.ctx)
				assert.NotNil(t, s.cancel)
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
	}{
		{
			name:      "starts_reaper_when_configured",
			hasReaper: true,
		},
		{
			name:      "skips_when_reaper_is_nil",
			hasReaper: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{}

			if tt.hasReaper {
				s.reaper = &mockReaper{}
			}

			// Should not panic.
			s.startReaper()
		})
	}
}

// mockReaper is a mock implementation of Reaper for testing.
type mockReaper struct {
	started bool
	stopped bool
}

func (m *mockReaper) Start() {
	m.started = true
}

func (m *mockReaper) IsPID1() bool {
	return false
}


func (m *mockReaper) ReapOnce() int {
	return 0
}
func (m *mockReaper) Stop() {
	m.stopped = true
}

// Test_Supervisor_startAllServices tests the startAllServices method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_startAllServices(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// numServices is the number of services.
		numServices int
		// expectError indicates if an error is expected.
		expectError bool
	}{
		{
			name:        "starts_zero_services",
			numServices: 0,
			expectError: false,
		},
		{
			name:        "starts_single_service",
			numServices: 1,
			expectError: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				managers: make(map[string]*applifecycle.Manager),
				state:    StateStarting,
			}

			err := s.startAllServices()

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
		// numServices is the number of services.
		numServices int
	}{
		{
			name:        "starts_zero_goroutines",
			numServices: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				managers: make(map[string]*applifecycle.Manager),
			}

			// Should not panic.
			s.startMonitoringGoroutines()
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
		// services is the list of services to configure.
		services []domainconfig.ServiceConfig
		// hasProberFactory indicates if prober factory is set.
		hasProberFactory bool
		// expectedMonitors is the expected number of health monitors.
		expectedMonitors int
	}{
		{
			name:             "empty_services_list",
			services:         []domainconfig.ServiceConfig{},
			hasProberFactory: false,
			expectedMonitors: 0,
		},
		{
			name: "services_without_probes_returns_nil_monitor",
			services: []domainconfig.ServiceConfig{
				{Name: "svc1", Command: "/bin/true", Listeners: nil},
				{Name: "svc2", Command: "/bin/true", Listeners: []domainconfig.ListenerConfig{{Name: "http", Port: 8080}}},
			},
			hasProberFactory: false,
			expectedMonitors: 0,
		},
		{
			name: "services_without_factory_returns_nil_monitor",
			services: []domainconfig.ServiceConfig{
				{
					Name:    "svc-with-probe",
					Command: "/bin/true",
					Listeners: []domainconfig.ListenerConfig{
						{
							Name: "http",
							Port: 8080,
							Probe: &domainconfig.ProbeConfig{
								Type: "http",
								Path: "/health",
							},
						},
					},
				},
			},
			hasProberFactory: false,
			expectedMonitors: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &Supervisor{
				config: &domainconfig.Config{
					Services: tt.services,
				},
				healthMonitors: make(map[string]*apphealth.ProbeMonitor),
				ctx:            context.Background(),
			}

			if tt.hasProberFactory {
				s.proberFactory = &mockProberFactory{}
			}

			// Should not panic.
			s.startHealthMonitors()

			// Verify expected number of monitors created.
			assert.Equal(t, tt.expectedMonitors, len(s.healthMonitors))
		})
	}
}


// Test_Supervisor_stopAll tests the stopAll method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_stopAll(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// numManagers is the number of managers.
		numManagers int
		// setupManagers sets up managers for the test.
		setupManagers func(s *Supervisor)
		// expectErrorHandlerCalled indicates if error handler should be called.
		expectErrorHandlerCalled bool
	}{
		{
			name:        "stops_zero_managers",
			numManagers: 0,
			setupManagers: func(s *Supervisor) {
				// No managers to set up.
			},
			expectErrorHandlerCalled: false,
		},
		{
			name:        "stops_single_manager_successfully",
			numManagers: 1,
			setupManagers: func(s *Supervisor) {
				// Create a real manager with a simple config (will be stopped).
				svcCfg := &domainconfig.ServiceConfig{
					Name:    "test-svc",
					Command: "/bin/true",
				}
				s.managers["test-svc"] = applifecycle.NewManager(svcCfg, nil)
			},
			expectErrorHandlerCalled: false,
		},
		{
			name:        "stops_multiple_managers_concurrently",
			numManagers: 3,
			setupManagers: func(s *Supervisor) {
				// Create multiple managers.
				for i := range 3 {
					name := "svc-" + string(rune('a'+i))
					svcCfg := &domainconfig.ServiceConfig{
						Name:    name,
						Command: "/bin/true",
					}
					s.managers[name] = applifecycle.NewManager(svcCfg, nil)
				}
			},
			expectErrorHandlerCalled: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errorHandlerCalled := false
			s := &Supervisor{
				managers: make(map[string]*applifecycle.Manager),
				errorHandler: func(_, _ string, _ error) {
					errorHandlerCalled = true
				},
			}

			tt.setupManagers(s)

			// Should not panic.
			s.stopAll()

			// Verify error handler behavior.
			assert.Equal(t, tt.expectErrorHandlerCalled, errorHandlerCalled)
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
	}{
		{
			name: "updates_services_from_new_config",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				managers: make(map[string]*applifecycle.Manager),
			}
			newCfg := &domainconfig.Config{
				Services: []domainconfig.ServiceConfig{},
			}

			// Should not panic.
			s.updateServices(newCfg)
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
		// existingManagers is the list of existing manager names.
		existingManagers []string
		// newConfigServices is the list of services in the new config.
		newConfigServices []string
		// expectedRemainingManagers is the expected remaining manager names.
		expectedRemainingManagers []string
		// expectErrorHandler indicates if error handler may be called.
		expectErrorHandler bool
	}{
		{
			name:                      "empty_managers_empty_config",
			existingManagers:          []string{},
			newConfigServices:         []string{},
			expectedRemainingManagers: []string{},
			expectErrorHandler:        false,
		},
		{
			name:                      "empty_managers_with_new_services",
			existingManagers:          []string{},
			newConfigServices:         []string{"new-svc-1", "new-svc-2"},
			expectedRemainingManagers: []string{},
			expectErrorHandler:        false,
		},
		{
			name:                      "all_services_remain_in_config",
			existingManagers:          []string{"svc-1", "svc-2"},
			newConfigServices:         []string{"svc-1", "svc-2"},
			expectedRemainingManagers: []string{"svc-1", "svc-2"},
			expectErrorHandler:        false,
		},
		{
			name:                      "one_service_removed_from_config",
			existingManagers:          []string{"svc-1", "svc-2"},
			newConfigServices:         []string{"svc-1"},
			expectedRemainingManagers: []string{"svc-1"},
			expectErrorHandler:        false,
		},
		{
			name:                      "all_services_removed_from_config",
			existingManagers:          []string{"svc-1", "svc-2", "svc-3"},
			newConfigServices:         []string{},
			expectedRemainingManagers: []string{},
			expectErrorHandler:        false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &Supervisor{
				managers: make(map[string]*applifecycle.Manager),
			}

			// Set up existing managers.
			for _, name := range tt.existingManagers {
				svcCfg := &domainconfig.ServiceConfig{
					Name:    name,
					Command: "/bin/true",
				}
				s.managers[name] = applifecycle.NewManager(svcCfg, nil)
			}

			// Build new config.
			newCfg := &domainconfig.Config{
				Services: make([]domainconfig.ServiceConfig, len(tt.newConfigServices)),
			}
			for i, name := range tt.newConfigServices {
				newCfg.Services[i] = domainconfig.ServiceConfig{
					Name:    name,
					Command: "/bin/true",
				}
			}

			// Should not panic.
			s.removeDeletedServices(newCfg)

			// Verify remaining managers.
			assert.Equal(t, len(tt.expectedRemainingManagers), len(s.managers))
			for _, name := range tt.expectedRemainingManagers {
				_, exists := s.managers[name]
				assert.True(t, exists, "expected manager %s to remain", name)
			}
		})
	}
}

// mockEventser is a mock implementation of Eventser for testing.
type mockEventser struct {
	events chan domain.Event
}

func (m *mockEventser) Events() <-chan domain.Event {
	return m.events
}

// Test_Supervisor_monitorService tests the monitorService method.
//
// Goroutine Lifecycle:
//   - Launched: s.monitorService() in "cancel_context" cases
//   - Terminated: via context cancellation or channel close
//   - Waited: s.wg.Wait() ensures completion
//
// Params:
//   - t: the testing context.
func Test_Supervisor_monitorService(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// setupBehavior describes how to set up the test behavior.
		setupBehavior string
		// eventsToSend is the number of events to send before exit.
		eventsToSend int
		// expectedHandledEvents is the number of events expected to be handled.
		expectedHandledEvents int
	}{
		{
			name:                  "exits_when_events_channel_closed",
			setupBehavior:         "close_channel",
			eventsToSend:          0,
			expectedHandledEvents: 0,
		},
		{
			name:                  "exits_when_context_cancelled",
			setupBehavior:         "cancel_context",
			eventsToSend:          0,
			expectedHandledEvents: 0,
		},
		{
			name:                  "handles_single_event_then_channel_closed",
			setupBehavior:         "close_channel",
			eventsToSend:          1,
			expectedHandledEvents: 1,
		},
		{
			name:                  "handles_multiple_events_then_channel_closed",
			setupBehavior:         "close_channel",
			eventsToSend:          3,
			expectedHandledEvents: 3,
		},
		{
			name:                  "handles_event_before_context_cancelled",
			setupBehavior:         "event_then_cancel",
			eventsToSend:          1,
			expectedHandledEvents: 1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			handledEvents := 0
			eventProcessed := make(chan struct{})
			s := &Supervisor{
				ctx:   ctx,
				stats: make(map[string]*ServiceStats),
				eventHandler: func(_ string, _ *domain.Event, _ *ServiceStatsSnapshot) {
					handledEvents++
					// Signal that an event was processed for sync.
					select {
					case eventProcessed <- struct{}{}:
					default:
					}
				},
			}

			mock := &mockEventser{
				events: make(chan domain.Event, tt.eventsToSend+1),
			}

			// Handle different setup behaviors.
			switch tt.setupBehavior {
			case "close_channel":
				// Send events first, then close.
				for range tt.eventsToSend {
					mock.events <- domain.Event{Type: domain.EventStarted}
				}
				close(mock.events)

				// Add to wait group before calling (monitorService calls wg.Done()).
				s.wg.Add(1)

				// Run monitorService synchronously.
				s.monitorService("test-service", mock)

			case "cancel_context":
				// Cancel immediately, no events.
				cancel()

				// Add to wait group before calling (monitorService calls wg.Done()).
				s.wg.Add(1)

				// Run monitorService synchronously.
				s.monitorService("test-service", mock)

			case "event_then_cancel":
				// Add to wait group before calling (monitorService calls wg.Done()).
				s.wg.Add(1)

				// Start monitor in goroutine.
				go s.monitorService("test-service", mock)

				// Send events and wait for processing.
				for range tt.eventsToSend {
					mock.events <- domain.Event{Type: domain.EventStarted}
					// Wait for event to be processed.
					<-eventProcessed
				}

				// Now cancel context to exit.
				cancel()
				s.wg.Wait()
			}

			// Verify expected number of events handled.
			assert.Equal(t, tt.expectedHandledEvents, handledEvents)
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
		// eventType is the event type.
		eventType domain.EventType
	}{
		{
			name:      "handles_started_event",
			eventType: domain.EventStarted,
		},
		{
			name:      "handles_stopped_event",
			eventType: domain.EventStopped,
		},
		{
			name:      "handles_failed_event",
			eventType: domain.EventFailed,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				stats:          make(map[string]*ServiceStats),
				healthMonitors: make(map[string]*apphealth.ProbeMonitor),
			}

			event := &domain.Event{
				Type: tt.eventType,
				PID:  1234,
			}

			// Should not panic.
			s.handleEvent("test-service", event)
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
		// hasProbes indicates if service has probes configured.
		hasProbes bool
		// hasFactory indicates if prober factory is set.
		hasFactory bool
		// expectNil indicates if nil is expected.
		expectNil bool
	}{
		{
			name:       "returns_nil_when_no_probes",
			hasProbes:  false,
			hasFactory: true,
			expectNil:  true,
		},
		{
			name:       "returns_nil_when_no_factory",
			hasProbes:  true,
			hasFactory: false,
			expectNil:  true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{}

			if tt.hasFactory {
				s.proberFactory = &mockProberFactory{}
			}

			svc := &domainconfig.ServiceConfig{
				Name: "test-service",
			}

			if tt.hasProbes {
				svc.Listeners = []domainconfig.ListenerConfig{
					{
						Probe: &domainconfig.ProbeConfig{
							Type: "http",
						},
					},
				}
			}

			monitor := s.createHealthMonitor(svc)

			if tt.expectNil {
				assert.Nil(t, monitor)
			}
		})
	}
}

// mockProberFactory is a mock implementation of Creator for testing.
type mockProberFactory struct{}

func (m *mockProberFactory) Create(proberType string, timeout time.Duration) (health.Prober, error) {
	return nil, nil
}

// Test_Supervisor_hasConfiguredProbes tests the hasConfiguredProbes method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_hasConfiguredProbes(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// svc is the service configuration.
		svc *domainconfig.ServiceConfig
		// expected is the expected result.
		expected bool
	}{
		{
			name: "returns_false_for_no_probes",
			svc: &domainconfig.ServiceConfig{
				Listeners: []domainconfig.ListenerConfig{},
			},
			expected: false,
		},
		{
			name: "returns_true_when_probe_configured",
			svc: &domainconfig.ServiceConfig{
				Listeners: []domainconfig.ListenerConfig{
					{
						Probe: &domainconfig.ProbeConfig{
							Type: "http",
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "returns_false_when_probe_nil",
			svc: &domainconfig.ServiceConfig{
				Listeners: []domainconfig.ListenerConfig{
					{
						Probe: nil,
					},
				},
			},
			expected: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{}

			result := s.hasConfiguredProbes(tt.svc)

			assert.Equal(t, tt.expected, result)
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
		// serviceName is the service name.
		serviceName string
	}{
		{
			name:        "creates_config_for_service",
			serviceName: "test-service",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				proberFactory: &mockProberFactory{},
				stats:         make(map[string]*ServiceStats),
			}

			cfg := s.createProbeMonitorConfig(tt.serviceName)

			assert.NotNil(t, cfg.Factory)
			assert.NotNil(t, cfg.OnStateChange)
			assert.NotNil(t, cfg.OnUnhealthy)
			assert.NotNil(t, cfg.OnHealthy)
		})
	}
}

// Test_Supervisor_addListenersWithProbes tests the addListenersWithProbes method.
//
// Params:
//   - t: the testing context.
func Test_Supervisor_addListenersWithProbes(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// numListeners is the number of listeners with probes.
		numListeners int
	}{
		{
			name:         "adds_zero_listeners",
			numListeners: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{}
			monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{})

			svc := &domainconfig.ServiceConfig{
				Listeners: []domainconfig.ListenerConfig{},
			}

			// Should not panic.
			s.addListenersWithProbes(monitor, svc)
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
	}{
		{
			name: "adds_single_listener",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{}
			monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{})

			lc := &domainconfig.ListenerConfig{
				Name:     "test-listener",
				Protocol: "tcp",
				Port:     8080,
				Probe: &domainconfig.ProbeConfig{
					Type: "tcp",
				},
			}

			// Should not panic.
			s.addSingleListenerWithProbe(monitor, lc)
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
		// lc is the listener configuration.
		lc *domainconfig.ListenerConfig
		// expectedProtocol is the expected protocol.
		expectedProtocol string
		// expectedAddress is the expected address.
		expectedAddress string
	}{
		{
			name: "creates_listener_with_defaults",
			lc: &domainconfig.ListenerConfig{
				Name: "test-listener",
				Port: 8080,
			},
			expectedProtocol: "tcp",
			expectedAddress:  "localhost",
		},
		{
			name: "creates_listener_with_custom_values",
			lc: &domainconfig.ListenerConfig{
				Name:     "custom-listener",
				Protocol: "udp",
				Address:  "0.0.0.0",
				Port:     9090,
			},
			expectedProtocol: "udp",
			expectedAddress:  "0.0.0.0",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{}

			listener := s.createDomainListener(tt.lc)

			assert.NotNil(t, listener)
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
		// lc is the listener configuration.
		lc *domainconfig.ListenerConfig
	}{
		{
			name: "creates_binding_with_defaults",
			lc: &domainconfig.ListenerConfig{
				Name: "test-listener",
				Port: 8080,
				Probe: &domainconfig.ProbeConfig{
					Type: "http",
					Path: "/health",
				},
			},
		},
		{
			name: "creates_binding_with_custom_address",
			lc: &domainconfig.ListenerConfig{
				Name:    "custom-listener",
				Address: "127.0.0.1",
				Port:    9090,
				Probe: &domainconfig.ProbeConfig{
					Type:    "grpc",
					Service: "health.v1.Health",
				},
			},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{}

			binding := s.createProbeBinding(tt.lc)

			assert.NotNil(t, binding)
			assert.Equal(t, tt.lc.Name, binding.ListenerName)
			assert.Equal(t, apphealth.ProbeType(tt.lc.Probe.Type), binding.Type)
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
		// err is the error to handle.
		err error
		// hasHandler indicates if error handler is set.
		hasHandler bool
		// expectCall indicates if handler should be called.
		expectCall bool
	}{
		{
			name:       "calls_handler_when_set",
			err:        assert.AnError,
			hasHandler: true,
			expectCall: true,
		},
		{
			name:       "does_nothing_when_no_handler",
			err:        assert.AnError,
			hasHandler: false,
			expectCall: false,
		},
		{
			name:       "does_nothing_when_no_error",
			err:        nil,
			hasHandler: true,
			expectCall: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			called := false
			s := &Supervisor{}

			if tt.hasHandler {
				s.errorHandler = func(operation, serviceName string, err error) {
					called = true
					assert.Equal(t, "test-op", operation)
					assert.Equal(t, "test-service", serviceName)
					assert.Equal(t, tt.err, err)
				}
			}

			s.handleRecoveryError("test-op", "test-service", tt.err)

			assert.Equal(t, tt.expectCall, called)
		})
	}
}
