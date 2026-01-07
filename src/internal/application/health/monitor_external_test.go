// Package health_test provides external tests for the health monitoring application service.
package health_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	health "github.com/kodflow/daemon/internal/application/health"
	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/domain/shared"
)

// monitorMockChecker is a test implementation of the Checker interface for monitor tests.
type monitorMockChecker struct {
	name   string
	typ    string
	result domain.Result
}

// Check performs a health check and returns the result.
//
// Params:
//   - _ctx: context for cancellation (unused in mock).
//
// Returns:
//   - domain.Result: the configured mock result.
func (mc *monitorMockChecker) Check(_ctx context.Context) domain.Result {
	// Return the pre-configured result.
	return mc.result
}

// Name returns the name of this checker.
//
// Returns:
//   - string: the checker name.
func (mc *monitorMockChecker) Name() string {
	// Return the configured name.
	return mc.name
}

// Type returns the type of this checker.
//
// Returns:
//   - string: the checker type.
func (mc *monitorMockChecker) Type() string {
	// Return the configured type.
	return mc.typ
}

// monitorMockFactory is a test implementation of the Creator interface for monitor tests.
type monitorMockFactory struct {
	checkers map[string]*monitorMockChecker
	err      error
}

// Create creates a checker from the given configuration.
//
// Params:
//   - cfg: checker configuration.
//
// Returns:
//   - health.Checker: the created checker.
//   - error: any error during creation.
func (mf *monitorMockFactory) Create(cfg health.CheckerConfig) (health.Checker, error) {
	// Return error if configured.
	if mf.err != nil {
		// Return the configured error.
		return nil, mf.err
	}
	// Look up checker by name.
	if checker, ok := mf.checkers[cfg.Name]; ok {
		// Return the matching checker.
		return checker, nil
	}
	// Create default healthy checker.
	return &monitorMockChecker{
		name:   cfg.Name,
		typ:    cfg.Type,
		result: domain.NewHealthyResult("ok", time.Millisecond),
	}, nil
}

// TestNewMonitor tests the NewMonitor constructor.
func TestNewMonitor(t *testing.T) {
	t.Parallel()
	// Define test cases for NewMonitor.
	tests := []struct {
		name           string
		configs        []service.HealthCheckConfig
		factory        *monitorMockFactory
		expectError    bool
		expectNil      bool
		expectedStatus domain.Status
	}{
		{
			name:    "empty config",
			configs: nil,
			factory: &monitorMockFactory{
				checkers: make(map[string]*monitorMockChecker, 0),
			},
			expectError:    false,
			expectNil:      false,
			expectedStatus: domain.StatusUnknown,
		},
		{
			name: "single checker",
			configs: []service.HealthCheckConfig{
				{
					Name:     "test-check",
					Type:     service.HealthCheckHTTP,
					Interval: shared.Duration(time.Second),
					Timeout:  shared.Duration(time.Second),
				},
			},
			factory: &monitorMockFactory{
				checkers: make(map[string]*monitorMockChecker, 0),
			},
			expectError:    false,
			expectNil:      false,
			expectedStatus: domain.StatusUnknown,
		},
		{
			name: "factory error",
			configs: []service.HealthCheckConfig{
				{
					Name:     "test-check",
					Type:     service.HealthCheckHTTP,
					Interval: shared.Duration(time.Second),
					Timeout:  shared.Duration(time.Second),
				},
			},
			factory: &monitorMockFactory{
				checkers: make(map[string]*monitorMockChecker, 0),
				err:      assert.AnError,
			},
			expectError:    true,
			expectNil:      true,
			expectedStatus: domain.StatusUnknown,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Create monitor with config.
			monitor, err := health.NewMonitor(testCase.configs, testCase.factory, nil)

			// Verify error expectation.
			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify nil expectation.
			if testCase.expectNil {
				assert.Nil(t, monitor)
			} else {
				assert.NotNil(t, monitor)
				// Verify initial status.
				assert.Equal(t, testCase.expectedStatus, monitor.Status())
			}
		})
	}
}

// TestMonitor_Start tests the Start method.
func TestMonitor_Start(t *testing.T) {
	t.Parallel()
	// Define test cases for Start.
	tests := []struct {
		name        string
		configs     []service.HealthCheckConfig
		doubleStart bool
		waitTime    time.Duration
	}{
		{
			name: "start and stop",
			configs: []service.HealthCheckConfig{
				{
					Name:     "test-check",
					Type:     service.HealthCheckHTTP,
					Interval: shared.Duration(100 * time.Millisecond),
					Timeout:  shared.Duration(50 * time.Millisecond),
				},
			},
			doubleStart: false,
			waitTime:    50 * time.Millisecond,
		},
		{
			name:        "double start",
			configs:     nil,
			doubleStart: true,
			waitTime:    0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Create factory with no checkers.
			factory := &monitorMockFactory{
				checkers: make(map[string]*monitorMockChecker, 0),
			}
			// Create monitor with config.
			monitor, err := health.NewMonitor(testCase.configs, factory, nil)
			// Verify no error occurred.
			require.NoError(t, err)

			// Get context from test.
			ctx := t.Context()

			// Start the monitor.
			monitor.Start(ctx)

			// Handle double start case.
			if testCase.doubleStart {
				monitor.Start(ctx)
			}

			// Wait if needed.
			if testCase.waitTime > 0 {
				time.Sleep(testCase.waitTime)
			}

			// Stop the monitor.
			monitor.Stop()
		})
	}
}

// TestMonitor_Stop tests the Stop method.
func TestMonitor_Stop(t *testing.T) {
	t.Parallel()
	// Define test cases for Stop.
	tests := []struct {
		name       string
		startFirst bool
		doubleStop bool
	}{
		{
			name:       "double stop",
			startFirst: true,
			doubleStop: true,
		},
		{
			name:       "stop without start",
			startFirst: false,
			doubleStop: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Create factory with no checkers.
			factory := &monitorMockFactory{
				checkers: make(map[string]*monitorMockChecker, 0),
			}
			// Create monitor with empty config.
			monitor, err := health.NewMonitor(nil, factory, nil)
			// Verify no error occurred.
			require.NoError(t, err)

			// Get context from test.
			ctx := t.Context()

			// Start the monitor if needed.
			if testCase.startFirst {
				monitor.Start(ctx)
			}

			// Stop the monitor.
			monitor.Stop()

			// Handle double stop case.
			if testCase.doubleStop {
				monitor.Stop()
			}
		})
	}
}

// TestMonitor_Status tests the Status method.
func TestMonitor_Status(t *testing.T) {
	t.Parallel()
	// Define test cases for Status.
	tests := []struct {
		name           string
		expectedStatus domain.Status
	}{
		{
			name:           "initial status is unknown",
			expectedStatus: domain.StatusUnknown,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Create factory with no checkers.
			factory := &monitorMockFactory{
				checkers: make(map[string]*monitorMockChecker, 0),
			}
			// Create monitor with empty config.
			monitor, err := health.NewMonitor(nil, factory, nil)
			// Verify no error occurred.
			require.NoError(t, err)
			// Verify status matches expectation.
			assert.Equal(t, testCase.expectedStatus, monitor.Status())
		})
	}
}

// TestMonitor_Results tests the Results method.
func TestMonitor_Results(t *testing.T) {
	t.Parallel()
	// Define test cases for Results.
	tests := []struct {
		name        string
		expectEmpty bool
	}{
		{
			name:        "empty results initially",
			expectEmpty: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Create factory with no checkers.
			factory := &monitorMockFactory{
				checkers: make(map[string]*monitorMockChecker, 0),
			}
			// Create monitor with empty config.
			monitor, err := health.NewMonitor(nil, factory, nil)
			// Verify no error occurred.
			require.NoError(t, err)
			// Get results.
			results := monitor.Results()
			// Verify results match expectation.
			if testCase.expectEmpty {
				assert.Empty(t, results)
			} else {
				assert.NotEmpty(t, results)
			}
		})
	}
}

// TestMonitor_IsHealthy tests the IsHealthy method.
func TestMonitor_IsHealthy(t *testing.T) {
	t.Parallel()
	// Define test cases for IsHealthy.
	tests := []struct {
		name          string
		expectHealthy bool
	}{
		{
			name:          "not healthy initially",
			expectHealthy: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Create factory with no checkers.
			factory := &monitorMockFactory{
				checkers: make(map[string]*monitorMockChecker, 0),
			}
			// Create monitor with empty config.
			monitor, err := health.NewMonitor(nil, factory, nil)
			// Verify no error occurred.
			require.NoError(t, err)
			// Verify health status matches expectation.
			assert.Equal(t, testCase.expectHealthy, monitor.IsHealthy())
		})
	}
}

// TestMonitorWithEvents tests the monitor with event channel.
func TestMonitorWithEvents(t *testing.T) {
	t.Parallel()
	// Define test cases for events.
	tests := []struct {
		name           string
		checkerName    string
		checkerType    string
		expectedStatus domain.Status
	}{
		{
			name:           "events are sent on status change",
			checkerName:    "test-check",
			checkerType:    "http",
			expectedStatus: domain.StatusHealthy,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Create event channel.
			events := make(chan domain.Event, 10)
			// Create mock checker.
			checker := &monitorMockChecker{
				name:   testCase.checkerName,
				typ:    testCase.checkerType,
				result: domain.NewHealthyResult("ok", time.Millisecond),
			}
			// Create factory with the checker.
			factory := &monitorMockFactory{
				checkers: map[string]*monitorMockChecker{
					testCase.checkerName: checker,
				},
			}
			// Create health check config.
			configs := []service.HealthCheckConfig{
				{
					Name:     testCase.checkerName,
					Type:     service.HealthCheckHTTP,
					Interval: shared.Duration(50 * time.Millisecond),
					Timeout:  shared.Duration(25 * time.Millisecond),
				},
			}
			// Create monitor with config and events.
			monitor, err := health.NewMonitor(configs, factory, events)
			// Verify no error occurred.
			require.NoError(t, err)

			// Get context from test.
			ctx := t.Context()

			// Start the monitor.
			monitor.Start(ctx)

			// Wait for event with timeout.
			select {
			// Receive event.
			case event := <-events:
				// Verify event checker name.
				assert.Equal(t, testCase.checkerName, event.Checker)
				// Verify event status.
				assert.Equal(t, testCase.expectedStatus, event.Status)
			// Timeout after waiting.
			case <-time.After(200 * time.Millisecond):
				// Fail on timeout.
				t.Fatal("timeout waiting for event")
			}

			// Stop the monitor.
			monitor.Stop()
		})
	}
}
