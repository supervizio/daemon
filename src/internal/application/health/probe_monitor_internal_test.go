// Package health provides internal tests for probe_monitor.go.
// It tests internal implementation details using white-box testing.
package health

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/listener"
	"github.com/kodflow/daemon/internal/domain/probe"
	"github.com/kodflow/daemon/internal/domain/process"
)

// internalTestProber is a mock prober for internal testing.
//
// internalTestProber provides a controllable prober implementation for
// testing internal ProbeMonitor behavior.
type internalTestProber struct {
	probeType  string
	result     probe.Result
	probeCount int
}

// Probe returns the configured test result and increments probe count.
//
// Params:
//   - ctx: the context for cancellation.
//   - target: the probe target.
//
// Returns:
//   - probe.Result: the configured test result.
func (p *internalTestProber) Probe(_ context.Context, _ probe.Target) probe.Result {
	// Increment probe count for verification.
	p.probeCount++
	// Return the pre-configured result for testing.
	return p.result
}

// Type returns the prober type.
//
// Returns:
//   - string: the prober type identifier.
func (p *internalTestProber) Type() string {
	// Return the configured prober type.
	return p.probeType
}

// internalTestCreator is a mock factory for internal testing.
//
// internalTestCreator provides a controllable factory implementation
// for testing internal ProbeMonitor behavior.
type internalTestCreator struct {
	probers map[string]*internalTestProber
	err     error
}

// Create returns a mock prober for the given type.
//
// Params:
//   - proberType: the type of prober to create.
//   - timeout: the timeout for the prober.
//
// Returns:
//   - probe.Prober: the created prober.
//   - error: if creation fails.
func (f *internalTestCreator) Create(proberType string, _ time.Duration) (probe.Prober, error) {
	// Return error if factory is configured to fail.
	if f.err != nil {
		return nil, f.err
	}

	// Return existing prober if available.
	if p, ok := f.probers[proberType]; ok {
		return p, nil
	}

	// Return default successful prober.
	return &internalTestProber{
		probeType: proberType,
		result:    probe.Result{Success: true},
	}, nil
}

// Test_ProbeMonitor_resultToStatus tests the resultToStatus method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_resultToStatus(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// result is the probe result to convert.
		result probe.Result
		// expected is the expected health status.
		expected domain.Status
	}{
		{
			name:     "success_maps_to_healthy",
			result:   probe.Result{Success: true},
			expected: domain.StatusHealthy,
		},
		{
			name:     "failure_maps_to_unhealthy",
			result:   probe.Result{Success: false},
			expected: domain.StatusUnhealthy,
		},
		{
			name: "success_with_output_maps_to_healthy",
			result: probe.Result{
				Success: true,
				Output:  "OK",
			},
			expected: domain.StatusHealthy,
		},
		{
			name: "failure_with_error_maps_to_unhealthy",
			result: probe.Result{
				Success: false,
				Output:  "connection refused",
			},
			expected: domain.StatusUnhealthy,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProbeMonitor(ProbeMonitorConfig{})

			status := monitor.resultToStatus(tt.result)

			// Verify the expected status.
			assert.Equal(t, tt.expected, status)
		})
	}
}

// Test_ProbeMonitor_findOrCreateListenerStatus tests findOrCreateListenerStatus.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_findOrCreateListenerStatus(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// existingListeners are listeners already in health.
		existingListeners []domain.ListenerStatus
		// listenerName is the name to find or create.
		listenerName string
		// expectedNew indicates if a new status should be created.
		expectedNew bool
	}{
		{
			name:              "creates_new_for_empty_list",
			existingListeners: nil,
			listenerName:      "test-listener",
			expectedNew:       true,
		},
		{
			name: "finds_existing_listener",
			existingListeners: []domain.ListenerStatus{
				{Name: "test-listener", State: listener.Ready},
			},
			listenerName: "test-listener",
			expectedNew:  false,
		},
		{
			name: "creates_new_when_not_found",
			existingListeners: []domain.ListenerStatus{
				{Name: "other-listener", State: listener.Ready},
			},
			listenerName: "test-listener",
			expectedNew:  true,
		},
		{
			name: "finds_first_match_in_multiple",
			existingListeners: []domain.ListenerStatus{
				{Name: "first-listener", State: listener.Listening},
				{Name: "test-listener", State: listener.Ready},
				{Name: "third-listener", State: listener.Listening},
			},
			listenerName: "test-listener",
			expectedNew:  false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProbeMonitor(ProbeMonitorConfig{})

			// Set up existing listeners.
			if tt.existingListeners != nil {
				monitor.health.Listeners = tt.existingListeners
			}
			initialLen := len(monitor.health.Listeners)

			// Create listener probe.
			lp := &ListenerProbe{
				Listener: listener.NewListener(tt.listenerName, "tcp", "localhost", 8080),
			}

			status := monitor.findOrCreateListenerStatus(lp)

			// Verify status was returned.
			require.NotNil(t, status)
			// Verify listener name matches.
			assert.Equal(t, tt.listenerName, status.Name)

			// Verify list grew if new was expected.
			if tt.expectedNew {
				assert.Len(t, monitor.health.Listeners, initialLen+1)
			} else {
				// Verify list size unchanged.
				assert.Len(t, monitor.health.Listeners, initialLen)
			}
		})
	}
}

// Test_ProbeMonitor_performProbe tests the performProbe method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_performProbe(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// probeSuccess indicates if probe should succeed.
		probeSuccess bool
		// configTimeout is the configured timeout.
		configTimeout time.Duration
		// targetAddress is the target address.
		targetAddress string
	}{
		{
			name:          "successful_probe",
			probeSuccess:  true,
			configTimeout: time.Second,
			targetAddress: "localhost:8080",
		},
		{
			name:          "failed_probe",
			probeSuccess:  false,
			configTimeout: time.Second,
			targetAddress: "localhost:8080",
		},
		{
			name:          "probe_with_default_timeout",
			probeSuccess:  true,
			configTimeout: 0,
			targetAddress: "",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			prober := &internalTestProber{
				probeType: "tcp",
				result:    probe.Result{Success: tt.probeSuccess},
			}

			monitor := NewProbeMonitor(ProbeMonitorConfig{
				Factory:        &internalTestCreator{},
				DefaultTimeout: 5 * time.Second,
			})

			// Create listener probe.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := &ListenerProbe{
				Listener: l,
				Prober:   prober,
				Config: probe.Config{
					Timeout:          tt.configTimeout,
					SuccessThreshold: 1,
					FailureThreshold: 1,
				},
				Target: probe.Target{
					Address: tt.targetAddress,
				},
			}

			// Add to monitor's listeners.
			monitor.listeners = append(monitor.listeners, lp)

			// Perform probe.
			ctx := context.Background()
			monitor.performProbe(ctx, lp)

			// Verify probe was called.
			assert.Equal(t, 1, prober.probeCount)
		})
	}
}

// Test_ProbeMonitor_updateProbeResult tests the updateProbeResult method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_updateProbeResult(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// result is the probe result.
		result probe.Result
		// successThreshold is the success threshold.
		successThreshold int
		// failureThreshold is the failure threshold.
		failureThreshold int
		// initialSuccesses is the initial consecutive successes.
		initialSuccesses int
		// initialFailures is the initial consecutive failures.
		initialFailures int
		// expectedSuccesses is the expected consecutive successes.
		expectedSuccesses int
		// expectedFailures is the expected consecutive failures.
		expectedFailures int
	}{
		{
			name:              "success_increments_successes",
			result:            probe.Result{Success: true},
			successThreshold:  3,
			failureThreshold:  3,
			initialSuccesses:  0,
			initialFailures:   0,
			expectedSuccesses: 1,
			expectedFailures:  0,
		},
		{
			name:              "failure_increments_failures",
			result:            probe.Result{Success: false},
			successThreshold:  3,
			failureThreshold:  3,
			initialSuccesses:  0,
			initialFailures:   0,
			expectedSuccesses: 0,
			expectedFailures:  1,
		},
		{
			name:              "success_resets_failures",
			result:            probe.Result{Success: true},
			successThreshold:  3,
			failureThreshold:  3,
			initialSuccesses:  0,
			initialFailures:   2,
			expectedSuccesses: 1,
			expectedFailures:  0,
		},
		{
			name:              "failure_resets_successes",
			result:            probe.Result{Success: false},
			successThreshold:  3,
			failureThreshold:  3,
			initialSuccesses:  2,
			initialFailures:   0,
			expectedSuccesses: 0,
			expectedFailures:  1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProbeMonitor(ProbeMonitorConfig{})

			// Create listener probe.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := &ListenerProbe{
				Listener: l,
				Prober:   &internalTestProber{probeType: "tcp"},
				Config: probe.Config{
					SuccessThreshold: tt.successThreshold,
					FailureThreshold: tt.failureThreshold,
				},
			}

			// Set up initial listener status.
			monitor.health.Listeners = []domain.ListenerStatus{
				{
					Name:                 "test",
					ConsecutiveSuccesses: tt.initialSuccesses,
					ConsecutiveFailures:  tt.initialFailures,
				},
			}

			// Update with probe result.
			monitor.updateProbeResult(lp, tt.result)

			// Verify counters.
			assert.Equal(t, tt.expectedSuccesses, monitor.health.Listeners[0].ConsecutiveSuccesses)
			assert.Equal(t, tt.expectedFailures, monitor.health.Listeners[0].ConsecutiveFailures)
		})
	}
}

// Test_ProbeMonitor_sendEventIfChanged tests the sendEventIfChanged method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_sendEventIfChanged(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prevState is the previous listener state.
		prevState listener.State
		// newState is the new listener state.
		newState listener.State
		// hasEventChannel indicates if event channel exists.
		hasEventChannel bool
		// expectEvent indicates if an event should be sent.
		expectEvent bool
	}{
		{
			name:            "no_event_when_state_unchanged",
			prevState:       listener.Listening,
			newState:        listener.Listening,
			hasEventChannel: true,
			expectEvent:     false,
		},
		{
			name:            "event_when_state_changed",
			prevState:       listener.Listening,
			newState:        listener.Ready,
			hasEventChannel: true,
			expectEvent:     true,
		},
		{
			name:            "no_event_when_no_channel",
			prevState:       listener.Listening,
			newState:        listener.Ready,
			hasEventChannel: false,
			expectEvent:     false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			var eventCh chan domain.Event
			// Create event channel if requested.
			if tt.hasEventChannel {
				eventCh = make(chan domain.Event, 1)
			}

			monitor := NewProbeMonitor(ProbeMonitorConfig{
				Events: eventCh,
			})

			// Create listener probe.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			l.State = tt.newState
			lp := &ListenerProbe{
				Listener: l,
			}

			// Create listener status with matching state.
			// In production, updateListenerState syncs ls.State with lp.Listener.State.
			ls := &domain.ListenerStatus{
				Name:            "test",
				State:           tt.newState,
				LastProbeResult: &domain.Result{},
			}

			// Call method.
			result := probe.Result{Success: true}
			monitor.sendEventIfChanged(lp, ls, tt.prevState, result)

			// Verify event was sent or not.
			if tt.expectEvent {
				select {
				case event := <-eventCh:
					assert.Equal(t, "test", event.Checker)
				// Timeout after short duration.
				case <-time.After(100 * time.Millisecond):
					t.Fatal("expected event not received")
				}
			} else if eventCh != nil {
				// Verify no event if channel exists.
				select {
				case <-eventCh:
					t.Fatal("unexpected event received")
				// Default case for non-blocking read.
				default:
					// No event as expected.
				}
			}
		})
	}
}

// Test_ProbeMonitor_struct tests the ProbeMonitor struct fields.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_struct(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// config is the monitor configuration.
		config ProbeMonitorConfig
		// expectedTimeout is the expected default timeout.
		expectedTimeout time.Duration
		// expectedInterval is the expected default interval.
		expectedInterval time.Duration
	}{
		{
			name:             "default_config",
			config:           ProbeMonitorConfig{},
			expectedTimeout:  probe.DefaultTimeout,
			expectedInterval: probe.DefaultInterval,
		},
		{
			name: "custom_timeout",
			config: ProbeMonitorConfig{
				DefaultTimeout: 10 * time.Second,
			},
			expectedTimeout:  10 * time.Second,
			expectedInterval: probe.DefaultInterval,
		},
		{
			name: "custom_interval",
			config: ProbeMonitorConfig{
				DefaultInterval: 30 * time.Second,
			},
			expectedTimeout:  probe.DefaultTimeout,
			expectedInterval: 30 * time.Second,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProbeMonitor(tt.config)

			// Verify default timeout.
			assert.Equal(t, tt.expectedTimeout, monitor.defaultTimeout)
			// Verify default interval.
			assert.Equal(t, tt.expectedInterval, monitor.defaultInterval)
			// Verify initial process state.
			assert.Equal(t, process.StateStopped, monitor.processState)
			// Verify not running.
			assert.False(t, monitor.running)
		})
	}
}

// Test_ProbeMonitor_createProber tests the createProber method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_createProber(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// hasFactory indicates if factory is configured.
		hasFactory bool
		// probeType is the listener probe type.
		probeType string
		// expectError indicates if an error is expected.
		expectError bool
		// expectedErr is the expected error.
		expectedErr error
	}{
		{
			name:        "creates_prober_successfully",
			hasFactory:  true,
			probeType:   "tcp",
			expectError: false,
		},
		{
			name:        "error_when_factory_missing",
			hasFactory:  false,
			probeType:   "tcp",
			expectError: true,
			expectedErr: ErrProberFactoryMissing,
		},
		{
			name:        "error_when_probe_type_empty",
			hasFactory:  true,
			probeType:   "",
			expectError: true,
			expectedErr: ErrEmptyProbeType,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			config := ProbeMonitorConfig{
				DefaultTimeout: 5 * time.Second,
			}
			// Configure factory if requested.
			if tt.hasFactory {
				config.Factory = &internalTestCreator{}
			}

			monitor := NewProbeMonitor(config)

			// Create listener with probe configuration.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			l.ProbeType = tt.probeType
			l.ProbeConfig = &probe.Config{Timeout: time.Second}

			prober, err := monitor.createProber(l)

			// Verify error expectation.
			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, prober)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, prober)
			}
		})
	}
}

// Test_ProbeMonitor_normalizeThresholds tests the normalizeThresholds method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_normalizeThresholds(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// successThreshold is the input success threshold.
		successThreshold int
		// failureThreshold is the input failure threshold.
		failureThreshold int
		// expectedSuccess is the expected normalized success threshold.
		expectedSuccess int
		// expectedFailure is the expected normalized failure threshold.
		expectedFailure int
	}{
		{
			name:             "positive_thresholds_unchanged",
			successThreshold: 3,
			failureThreshold: 5,
			expectedSuccess:  3,
			expectedFailure:  5,
		},
		{
			name:             "zero_thresholds_normalized_to_one",
			successThreshold: 0,
			failureThreshold: 0,
			expectedSuccess:  1,
			expectedFailure:  1,
		},
		{
			name:             "negative_thresholds_normalized_to_one",
			successThreshold: -1,
			failureThreshold: -2,
			expectedSuccess:  1,
			expectedFailure:  1,
		},
		{
			name:             "mixed_thresholds",
			successThreshold: 2,
			failureThreshold: 0,
			expectedSuccess:  2,
			expectedFailure:  1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProbeMonitor(ProbeMonitorConfig{})

			config := probe.Config{
				SuccessThreshold: tt.successThreshold,
				FailureThreshold: tt.failureThreshold,
			}

			success, failure := monitor.normalizeThresholds(config)

			// Verify normalized thresholds.
			assert.Equal(t, tt.expectedSuccess, success)
			assert.Equal(t, tt.expectedFailure, failure)
		})
	}
}

// Test_ProbeMonitor_updateListenerState tests the updateListenerState method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_updateListenerState(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// result is the probe result.
		result probe.Result
		// initialState is the initial listener state.
		initialState listener.State
		// initialSuccesses is initial consecutive successes.
		initialSuccesses int
		// initialFailures is initial consecutive failures.
		initialFailures int
		// successThreshold is the success threshold.
		successThreshold int
		// failureThreshold is the failure threshold.
		failureThreshold int
		// expectedState is the expected final state.
		expectedState listener.State
		// expectedSuccesses is expected consecutive successes.
		expectedSuccesses int
		// expectedFailures is expected consecutive failures.
		expectedFailures int
	}{
		{
			name:              "success_increments_and_transitions_to_ready",
			result:            probe.Result{Success: true},
			initialState:      listener.Listening,
			initialSuccesses:  0,
			initialFailures:   0,
			successThreshold:  1,
			failureThreshold:  1,
			expectedState:     listener.Ready,
			expectedSuccesses: 1,
			expectedFailures:  0,
		},
		{
			name:              "failure_increments_and_transitions_to_listening",
			result:            probe.Result{Success: false},
			initialState:      listener.Ready,
			initialSuccesses:  0,
			initialFailures:   0,
			successThreshold:  1,
			failureThreshold:  1,
			expectedState:     listener.Listening,
			expectedSuccesses: 0,
			expectedFailures:  1,
		},
		{
			name:              "success_below_threshold_no_transition",
			result:            probe.Result{Success: true},
			initialState:      listener.Listening,
			initialSuccesses:  0,
			initialFailures:   0,
			successThreshold:  3,
			failureThreshold:  3,
			expectedState:     listener.Listening,
			expectedSuccesses: 1,
			expectedFailures:  0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProbeMonitor(ProbeMonitorConfig{})

			// Create listener probe.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			l.State = tt.initialState
			lp := &ListenerProbe{
				Listener: l,
				Prober:   &internalTestProber{probeType: "tcp"},
			}

			// Create listener status.
			ls := &domain.ListenerStatus{
				Name:                 "test",
				State:                tt.initialState,
				ConsecutiveSuccesses: tt.initialSuccesses,
				ConsecutiveFailures:  tt.initialFailures,
			}

			// Call method.
			monitor.updateListenerState(lp, ls, tt.result, tt.successThreshold, tt.failureThreshold)

			// Verify results.
			assert.Equal(t, tt.expectedState, ls.State)
			assert.Equal(t, tt.expectedSuccesses, ls.ConsecutiveSuccesses)
			assert.Equal(t, tt.expectedFailures, ls.ConsecutiveFailures)
		})
	}
}

// Test_ProbeMonitor_storeProbeResult tests the storeProbeResult method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_storeProbeResult(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// result is the probe result to store.
		result probe.Result
		// expectedStatus is the expected stored status.
		expectedStatus domain.Status
	}{
		{
			name: "stores_successful_result",
			result: probe.Result{
				Success: true,
				Latency: 10 * time.Millisecond,
				Output:  "OK",
			},
			expectedStatus: domain.StatusHealthy,
		},
		{
			name: "stores_failed_result",
			result: probe.Result{
				Success: false,
				Latency: 50 * time.Millisecond,
				Output:  "connection refused",
			},
			expectedStatus: domain.StatusUnhealthy,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProbeMonitor(ProbeMonitorConfig{})

			// Create listener status.
			ls := &domain.ListenerStatus{
				Name: "test",
			}

			// Call method.
			monitor.storeProbeResult(ls, tt.result)

			// Verify result was stored.
			require.NotNil(t, ls.LastProbeResult)
			assert.Equal(t, tt.expectedStatus, ls.LastProbeResult.Status)
			assert.Equal(t, tt.result.Output, ls.LastProbeResult.Message)
			assert.Equal(t, tt.result.Latency, ls.LastProbeResult.Duration)
			// Verify latency was updated in health.
			assert.Equal(t, tt.result.Latency, monitor.health.Latency)
		})
	}
}

// Test_ProbeMonitor_runProber tests the runProber method.
// This test spawns a goroutine that runs the prober in a loop.
// The goroutine terminates when the context is cancelled.
// Resources are cleaned up via deferred ticker.Stop() in runProber.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_runProber(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// interval is the probe interval.
		interval time.Duration
		// cancelAfter is how long to wait before canceling.
		cancelAfter time.Duration
		// minProbes is the minimum expected probes.
		minProbes int
	}{
		{
			name:        "performs_initial_probe",
			interval:    100 * time.Millisecond,
			cancelAfter: 50 * time.Millisecond,
			minProbes:   1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			prober := &internalTestProber{
				probeType: "tcp",
				result:    probe.Result{Success: true},
			}

			monitor := NewProbeMonitor(ProbeMonitorConfig{
				Factory:         &internalTestCreator{},
				DefaultInterval: tt.interval,
			})

			// Create listener probe.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := &ListenerProbe{
				Listener: l,
				Prober:   prober,
				Config: probe.Config{
					Interval: tt.interval,
				},
			}

			// Create context with cancel.
			ctx, cancel := context.WithCancel(context.Background())

			// Start monitor to initialize stopCh.
			monitor.running = true
			stopCh := make(chan struct{})
			monitor.stopCh = stopCh

			// Run prober in goroutine.
			done := make(chan struct{})
			go func() {
				monitor.runProber(ctx, stopCh, lp)
				close(done)
			}()

			// Wait and cancel.
			time.Sleep(tt.cancelAfter)
			cancel()

			// Wait for goroutine to finish.
			<-done

			// Verify minimum probes.
			assert.GreaterOrEqual(t, prober.probeCount, tt.minProbes)
		})
	}
}
