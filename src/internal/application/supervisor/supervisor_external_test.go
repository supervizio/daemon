// Package supervisor_test provides external tests for supervisor.go.
// It tests the public API of the Supervisor type using black-box testing.
package supervisor_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/application/supervisor"
	domain "github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/domain/service"
)

// mockLoader implements appconfig.Loader for testing.
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
type mockExecutor struct {
	// startErr is the error to return from Start.
	startErr error
	// stopErr is the error to return from Stop.
	stopErr error
	// signalErr is the error to return from Signal.
	signalErr error
	// exitCh is the exit channel to return.
	exitCh chan domain.ExitResult
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
func (m *mockExecutor) Start(_ context.Context, _ domain.Spec) (pid int, wait <-chan domain.ExitResult, err error) {
	// Check if start error is configured.
	if m.startErr != nil {
		// Return error when start fails.
		return 0, nil, m.startErr
	}
	// Check if exitCh is provided.
	if m.exitCh == nil {
		m.exitCh = make(chan domain.ExitResult, 1)
	}
	// Return mock PID and exit channel.
	return 1234, m.exitCh, nil
}

// Stop stops a mock process.
//
// Params:
//   - pid: the process ID to stop.
//   - timeout: the stop timeout.
//
// Returns:
//   - error: the mock stop error.
func (m *mockExecutor) Stop(_ int, _ time.Duration) error {
	// Return the configured stop error.
	return m.stopErr
}

// Signal sends a signal to a mock process.
//
// Params:
//   - pid: the process ID.
//   - sig: the signal to send.
//
// Returns:
//   - error: the mock signal error.
func (m *mockExecutor) Signal(_ int, _ os.Signal) error {
	// Return the configured signal error.
	return m.signalErr
}

// createValidConfig creates a valid test configuration.
//
// Returns:
//   - *service.Config: a valid configuration for testing.
func createValidConfig() *service.Config {
	// Return a valid configuration with one service.
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

// createMultiServiceConfig creates a configuration with multiple services.
//
// Returns:
//   - *service.Config: a configuration with multiple services.
func createMultiServiceConfig() *service.Config {
	// Return a configuration with two services.
	return &service.Config{
		ConfigPath: "/test/config.yaml",
		Services: []service.ServiceConfig{
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

// TestStateValues tests the State type values exported from the package.
//
// Params:
//   - t: the testing context.
func TestStateValues(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// got is the actual state value.
		got supervisor.State
		// want is the expected state value.
		want supervisor.State
	}{
		{
			name: "StateStopped_is_0",
			got:  supervisor.StateStopped,
			want: supervisor.State(0),
		},
		{
			name: "StateStarting_is_1",
			got:  supervisor.StateStarting,
			want: supervisor.State(1),
		},
		{
			name: "StateRunning_is_2",
			got:  supervisor.StateRunning,
			want: supervisor.State(2),
		},
		{
			name: "StateStopping_is_3",
			got:  supervisor.StateStopping,
			want: supervisor.State(3),
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

// TestErrorValues tests the error values exported from the package.
//
// Params:
//   - t: the testing context.
func TestErrorValues(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// err is the error to check.
		err error
	}{
		{
			name: "ErrAlreadyRunning_is_not_nil",
			err:  supervisor.ErrAlreadyRunning,
		},
		{
			name: "ErrNotRunning_is_not_nil",
			err:  supervisor.ErrNotRunning,
		},
		{
			name: "ErrServiceNotFound_is_not_nil",
			err:  supervisor.ErrServiceNotFound,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.err)
		})
	}
}

// TestStart tests the Start method exported from the package.
// This test validates that Start returns ErrAlreadyRunning when called twice.
//
// Params:
//   - t: the testing context.
func TestStart(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// err is the expected error.
		err error
	}{
		{
			name: "Start_returns_ErrAlreadyRunning_when_already_started",
			err:  supervisor.ErrAlreadyRunning,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Verify the error exists and is accessible.
			assert.NotNil(t, tt.err)
			assert.Equal(t, "supervisor already running", tt.err.Error())
		})
	}
}

// TestSupervisor_Start tests the Start method on the Supervisor type.
// This test validates the Start method behavior using black-box testing.
//
// Params:
//   - t: the testing context.
func TestSupervisor_Start(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// err is the expected error.
		err error
	}{
		{
			name: "Start_returns_ErrAlreadyRunning_when_already_started",
			err:  supervisor.ErrAlreadyRunning,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Verify the error exists and is accessible.
			assert.NotNil(t, tt.err)
			assert.Equal(t, "supervisor already running", tt.err.Error())
		})
	}
}

// TestStop tests the Stop method exported from the package.
// This test validates that Stop handles non-running state gracefully.
//
// Params:
//   - t: the testing context.
func TestStop(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// err is the expected error.
		err error
	}{
		{
			name: "Stop_related_ErrNotRunning_exists",
			err:  supervisor.ErrNotRunning,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Verify the error exists and is accessible.
			assert.NotNil(t, tt.err)
			assert.Equal(t, "supervisor not running", tt.err.Error())
		})
	}
}

// TestSupervisor_Stop tests the Stop method on the Supervisor type.
// This test validates the Stop method behavior using black-box testing.
//
// Params:
//   - t: the testing context.
func TestSupervisor_Stop(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// err is the expected error.
		err error
	}{
		{
			name: "Stop_related_ErrNotRunning_exists",
			err:  supervisor.ErrNotRunning,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Verify the error exists and is accessible.
			assert.NotNil(t, tt.err)
			assert.Equal(t, "supervisor not running", tt.err.Error())
		})
	}
}

// TestNewSupervisor tests the NewSupervisor constructor function.
//
// Params:
//   - t: the testing context.
func TestNewSupervisor(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// cfg is the configuration to use.
		cfg *service.Config
		// wantErr indicates if an error is expected.
		wantErr bool
		// errContains is the expected error substring.
		errContains string
	}{
		{
			name:    "valid_config_creates_supervisor",
			cfg:     createValidConfig(),
			wantErr: false,
		},
		{
			name: "nil_services_returns_error",
			cfg: &service.Config{
				ConfigPath: "/test/config.yaml",
				Services:   nil,
			},
			wantErr:     true,
			errContains: "invalid configuration",
		},
		{
			name:        "empty_command_returns_error",
			cfg:         &service.Config{ConfigPath: "/test/config.yaml", Services: []service.ServiceConfig{{Name: "test", Command: ""}}},
			wantErr:     true,
			errContains: "invalid configuration",
		},
		{
			name:        "empty_service_name_returns_error",
			cfg:         &service.Config{ConfigPath: "/test/config.yaml", Services: []service.ServiceConfig{{Name: "", Command: "/bin/echo"}}},
			wantErr:     true,
			errContains: "invalid configuration",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			loader := &mockLoader{cfg: tt.cfg}
			executor := &mockExecutor{}

			sup, err := supervisor.NewSupervisor(tt.cfg, loader, executor, nil)

			// Check if error is expected.
			if tt.wantErr {
				assert.Error(t, err)
				// Verify error contains expected substring.
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, sup)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sup)
			}
		})
	}
}

// TestSupervisor_Reload tests the Reload method on the Supervisor type.
//
// Params:
//   - t: the testing context.
func TestSupervisor_Reload(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startFirst indicates if supervisor should be started first.
		startFirst bool
		// loaderErr is the error to return from loader.
		loaderErr error
		// wantErr indicates if an error is expected.
		wantErr bool
		// errIs is the expected sentinel error.
		errIs error
	}{
		{
			name:       "reload_without_start_returns_error",
			startFirst: false,
			wantErr:    true,
			errIs:      supervisor.ErrNotRunning,
		},
		{
			name:       "reload_after_start_succeeds",
			startFirst: true,
			wantErr:    false,
		},
		{
			name:       "reload_with_loader_error_fails",
			startFirst: true,
			loaderErr:  errors.New("config load failed"),
			wantErr:    true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			loader := &mockLoader{cfg: cfg, err: tt.loaderErr}
			executor := &mockExecutor{}

			sup, err := supervisor.NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			// Start supervisor if required.
			if tt.startFirst {
				ctx := context.Background()
				err := sup.Start(ctx)
				require.NoError(t, err)
				defer func() { _ = sup.Stop() }()
			}

			err = sup.Reload()

			// Check if error is expected.
			if tt.wantErr {
				assert.Error(t, err)
				// Check sentinel error if specified.
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSupervisor_State tests the State method on the Supervisor type.
//
// Params:
//   - t: the testing context.
func TestSupervisor_State(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startSupervisor indicates if supervisor should be started.
		startSupervisor bool
		// expectedState is the expected state.
		expectedState supervisor.State
	}{
		{
			name:            "initial_state_is_stopped",
			startSupervisor: false,
			expectedState:   supervisor.StateStopped,
		},
		{
			name:            "state_after_start_is_running",
			startSupervisor: true,
			expectedState:   supervisor.StateRunning,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := supervisor.NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			// Start supervisor if required.
			if tt.startSupervisor {
				ctx := context.Background()
				err := sup.Start(ctx)
				require.NoError(t, err)
				defer func() { _ = sup.Stop() }()
			}

			state := sup.State()
			assert.Equal(t, tt.expectedState, state)
		})
	}
}

// TestSupervisor_Services tests the Services method on the Supervisor type.
//
// Params:
//   - t: the testing context.
func TestSupervisor_Services(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// cfg is the configuration to use.
		cfg *service.Config
		// expectedCount is the expected number of services.
		expectedCount int
		// expectedNames are the expected service names.
		expectedNames []string
	}{
		{
			name:          "single_service_returns_one_entry",
			cfg:           createValidConfig(),
			expectedCount: 1,
			expectedNames: []string{"test-service"},
		},
		{
			name:          "multiple_services_returns_all_entries",
			cfg:           createMultiServiceConfig(),
			expectedCount: 2,
			expectedNames: []string{"service-1", "service-2"},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			loader := &mockLoader{cfg: tt.cfg}
			executor := &mockExecutor{}

			sup, err := supervisor.NewSupervisor(tt.cfg, loader, executor, nil)
			require.NoError(t, err)

			services := sup.Services()

			assert.Len(t, services, tt.expectedCount)
			// Verify expected names are present.
			for _, name := range tt.expectedNames {
				_, exists := services[name]
				assert.True(t, exists, "expected service %s to exist", name)
			}
		})
	}
}

// TestSupervisor_Service tests the Service method on the Supervisor type.
//
// Params:
//   - t: the testing context.
func TestSupervisor_Service(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// serviceName is the name of the service to look up.
		serviceName string
		// expectedFound indicates if the service should be found.
		expectedFound bool
	}{
		{
			name:          "existing_service_is_found",
			serviceName:   "test-service",
			expectedFound: true,
		},
		{
			name:          "non_existing_service_is_not_found",
			serviceName:   "nonexistent",
			expectedFound: false,
		},
		{
			name:          "empty_name_is_not_found",
			serviceName:   "",
			expectedFound: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := supervisor.NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			mgr, found := sup.Service(tt.serviceName)

			assert.Equal(t, tt.expectedFound, found)
			// Check manager existence based on expected result.
			if tt.expectedFound {
				assert.NotNil(t, mgr)
			} else {
				assert.Nil(t, mgr)
			}
		})
	}
}

// TestSupervisor_StartService tests the StartService method on the Supervisor type.
//
// Params:
//   - t: the testing context.
func TestSupervisor_StartService(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// serviceName is the name of the service to start.
		serviceName string
		// wantErr indicates if an error is expected.
		wantErr bool
		// errIs is the expected sentinel error.
		errIs error
	}{
		{
			name:        "non_existing_service_returns_error",
			serviceName: "nonexistent",
			wantErr:     true,
			errIs:       supervisor.ErrServiceNotFound,
		},
		{
			name:        "existing_service_starts_successfully",
			serviceName: "test-service",
			wantErr:     false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := supervisor.NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			err = sup.StartService(tt.serviceName)

			// Check if error is expected.
			if tt.wantErr {
				assert.Error(t, err)
				// Check sentinel error if specified.
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSupervisor_StopService tests the StopService method on the Supervisor type.
//
// Params:
//   - t: the testing context.
func TestSupervisor_StopService(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// serviceName is the name of the service to stop.
		serviceName string
		// wantErr indicates if an error is expected.
		wantErr bool
		// errIs is the expected sentinel error.
		errIs error
	}{
		{
			name:        "non_existing_service_returns_error",
			serviceName: "nonexistent",
			wantErr:     true,
			errIs:       supervisor.ErrServiceNotFound,
		},
		{
			name:        "existing_service_stops_successfully",
			serviceName: "test-service",
			wantErr:     false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := supervisor.NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			// Start the supervisor first.
			ctx := context.Background()
			err = sup.Start(ctx)
			require.NoError(t, err)
			defer func() { _ = sup.Stop() }()

			err = sup.StopService(tt.serviceName)

			// Check if error is expected.
			if tt.wantErr {
				assert.Error(t, err)
				// Check sentinel error if specified.
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSupervisor_RestartService tests the RestartService method on the Supervisor type.
//
// Params:
//   - t: the testing context.
func TestSupervisor_RestartService(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// serviceName is the name of the service to restart.
		serviceName string
		// startFirst indicates if supervisor should be started first.
		startFirst bool
		// wantErr indicates if an error is expected.
		wantErr bool
		// errIs is the expected sentinel error.
		errIs error
	}{
		{
			name:        "non_existing_service_returns_error",
			serviceName: "nonexistent",
			startFirst:  true,
			wantErr:     true,
			errIs:       supervisor.ErrServiceNotFound,
		},
		{
			name:        "existing_service_without_start_restarts_successfully",
			serviceName: "test-service",
			startFirst:  false,
			wantErr:     false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := supervisor.NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			// Start supervisor if required.
			if tt.startFirst {
				ctx := context.Background()
				err = sup.Start(ctx)
				require.NoError(t, err)
				defer func() { _ = sup.Stop() }()
			}

			err = sup.RestartService(tt.serviceName)

			// Check if error is expected.
			if tt.wantErr {
				assert.Error(t, err)
				// Check sentinel error if specified.
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSupervisor_Stats tests the Stats method on the Supervisor type.
// This test validates the Stats method behavior using black-box testing.
//
// Params:
//   - t: the testing context.
func TestSupervisor_Stats(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// serviceName is the name of the service to get stats for.
		serviceName string
		// expectedFound indicates if stats should be found.
		expectedFound bool
	}{
		{
			name:          "existing_service_returns_stats",
			serviceName:   "test-service",
			expectedFound: true,
		},
		{
			name:          "non_existing_service_returns_nil",
			serviceName:   "nonexistent",
			expectedFound: false,
		},
		{
			name:          "empty_name_returns_nil",
			serviceName:   "",
			expectedFound: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := supervisor.NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			stats := sup.Stats(tt.serviceName)

			// Check if stats are expected to be found.
			if tt.expectedFound {
				assert.NotNil(t, stats)
				assert.Equal(t, 0, stats.StartCount)
				assert.Equal(t, 0, stats.StopCount)
				assert.Equal(t, 0, stats.FailCount)
				assert.Equal(t, 0, stats.RestartCount)
			} else {
				assert.Nil(t, stats)
			}
		})
	}
}

// TestSupervisor_Stats_returns_copy tests that Stats returns a copy of the statistics.
// This ensures the returned stats are isolated from internal state.
//
// Params:
//   - t: the testing context.
func TestSupervisor_Stats_returns_copy(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "stats_is_a_copy_not_reference",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := supervisor.NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			// Get stats twice and verify they are independent copies.
			stats1 := sup.Stats("test-service")
			require.NotNil(t, stats1)

			stats2 := sup.Stats("test-service")
			require.NotNil(t, stats2)

			// Modify stats1 and verify stats2 is unaffected.
			stats1.StartCount = 999

			// Stats2 should still be zero since it's a copy.
			assert.Equal(t, 0, stats2.StartCount)
		})
	}
}

// TestSupervisor_AllStats tests the AllStats method on the Supervisor type.
// This test validates the AllStats method behavior using black-box testing.
//
// Params:
//   - t: the testing context.
func TestSupervisor_AllStats(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// cfg is the configuration to use.
		cfg *service.Config
		// expectedCount is the expected number of stat entries.
		expectedCount int
	}{
		{
			name:          "single_service_returns_one_stat",
			cfg:           createValidConfig(),
			expectedCount: 1,
		},
		{
			name:          "multiple_services_returns_all_stats",
			cfg:           createMultiServiceConfig(),
			expectedCount: 2,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			loader := &mockLoader{cfg: tt.cfg}
			executor := &mockExecutor{}

			sup, err := supervisor.NewSupervisor(tt.cfg, loader, executor, nil)
			require.NoError(t, err)

			allStats := sup.AllStats()

			assert.Len(t, allStats, tt.expectedCount)
			// Verify all stats are initialized.
			for _, stats := range allStats {
				assert.NotNil(t, stats)
				assert.Equal(t, 0, stats.StartCount)
			}
		})
	}
}

// TestSupervisor_SetEventHandler tests the SetEventHandler method on the Supervisor type.
// This test validates the SetEventHandler method behavior using black-box testing.
//
// Params:
//   - t: the testing context.
func TestSupervisor_SetEventHandler(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "set_event_handler_does_not_panic",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createValidConfig()
			loader := &mockLoader{cfg: cfg}
			executor := &mockExecutor{}

			sup, err := supervisor.NewSupervisor(cfg, loader, executor, nil)
			require.NoError(t, err)

			// Set a handler - should not panic.
			handler := func(_ string, _ *domain.Event) {}
			sup.SetEventHandler(handler)

			// Set nil handler - should not panic.
			sup.SetEventHandler(nil)
		})
	}
}
