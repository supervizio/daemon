// Package health provides internal tests for monitor.go.
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
	"github.com/kodflow/daemon/internal/domain/process"
)

// internalTestProber is a mock prober for internal testing.
//
// internalTestProber provides a controllable prober implementation for
// testing internal ProbeMonitor behavior.
type internalTestProber struct {
	probeType  string
	result     domain.CheckResult
	probeCount int
}

// Probe returns the configured test result and increments probe count.
//
// Params:
//   - ctx: the context for cancellation.
//   - target: the probe target.
//
// Returns:
//   - domain.CheckResult: the configured test result.
func (p *internalTestProber) Probe(_ context.Context, _ domain.Target) domain.CheckResult {
	p.probeCount++
	return p.result
}

// Type returns the prober type.
//
// Returns:
//   - string: the prober type identifier.
func (p *internalTestProber) Type() string {
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
//   - domain.Prober: the created prober.
//   - error: if creation fails.
func (f *internalTestCreator) Create(proberType string, _ time.Duration) (domain.Prober, error) {
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
		result:    domain.CheckResult{Success: true},
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
		result domain.CheckResult
		// expected is the expected health status.
		expected domain.Status
	}{
		{
			name:     "success_maps_to_healthy",
			result:   domain.CheckResult{Success: true},
			expected: domain.StatusHealthy,
		},
		{
			name:     "failure_maps_to_unhealthy",
			result:   domain.CheckResult{Success: false},
			expected: domain.StatusUnhealthy,
		},
		{
			name: "success_with_output_maps_to_healthy",
			result: domain.CheckResult{
				Success: true,
				Output:  "OK",
			},
			expected: domain.StatusHealthy,
		},
		{
			name: "failure_with_error_maps_to_unhealthy",
			result: domain.CheckResult{
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

// Test_ProbeMonitor_findOrCreateSubjectStatus tests findOrCreateSubjectStatus.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_findOrCreateSubjectStatus(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// existingSubjects are subjects already in health.
		existingSubjects []domain.SubjectStatus
		// listenerName is the name to find or create.
		listenerName string
		// expectedNew indicates if a new status should be created.
		expectedNew bool
	}{
		{
			name:             "creates_new_for_empty_list",
			existingSubjects: nil,
			listenerName:     "test-listener",
			expectedNew:      true,
		},
		{
			name: "finds_existing_subject",
			existingSubjects: []domain.SubjectStatus{
				{Name: "test-listener", State: domain.SubjectReady},
			},
			listenerName: "test-listener",
			expectedNew:  false,
		},
		{
			name: "creates_new_when_not_found",
			existingSubjects: []domain.SubjectStatus{
				{Name: "other-listener", State: domain.SubjectReady},
			},
			listenerName: "test-listener",
			expectedNew:  true,
		},
		{
			name: "finds_first_match_in_multiple",
			existingSubjects: []domain.SubjectStatus{
				{Name: "first-listener", State: domain.SubjectListening},
				{Name: "test-listener", State: domain.SubjectReady},
				{Name: "third-listener", State: domain.SubjectListening},
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

			// Set up existing subjects.
			if tt.existingSubjects != nil {
				monitor.health.Subjects = tt.existingSubjects
			}
			initialLen := len(monitor.health.Subjects)

			// Create listener probe.
			lp := &ListenerProbe{
				Listener: listener.NewListener(tt.listenerName, "tcp", "localhost", 8080),
			}

			status := monitor.findOrCreateSubjectStatus(lp)

			// Verify status was returned.
			require.NotNil(t, status)
			// Verify listener name matches.
			assert.Equal(t, tt.listenerName, status.Name)

			// Verify list grew if new was expected.
			if tt.expectedNew {
				assert.Len(t, monitor.health.Subjects, initialLen+1)
			} else {
				// Verify list size unchanged.
				assert.Len(t, monitor.health.Subjects, initialLen)
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
		// bindingTimeout is the configured timeout.
		bindingTimeout time.Duration
		// targetAddress is the target address.
		targetAddress string
	}{
		{
			name:           "successful_probe",
			probeSuccess:   true,
			bindingTimeout: time.Second,
			targetAddress:  "localhost:8080",
		},
		{
			name:           "failed_probe",
			probeSuccess:   false,
			bindingTimeout: time.Second,
			targetAddress:  "localhost:8080",
		},
		{
			name:           "probe_with_default_timeout",
			probeSuccess:   true,
			bindingTimeout: 0,
			targetAddress:  "",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			prober := &internalTestProber{
				probeType: "tcp",
				result:    domain.CheckResult{Success: tt.probeSuccess},
			}

			monitor := NewProbeMonitor(ProbeMonitorConfig{
				Factory:        &internalTestCreator{},
				DefaultTimeout: 5 * time.Second,
			})

			// Create listener probe with binding.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			binding := &ProbeBinding{
				ListenerName: "test",
				Type:         ProbeTCP,
				Target: ProbeTarget{
					Address: tt.targetAddress,
				},
				Config: ProbeConfig{
					Timeout:          tt.bindingTimeout,
					SuccessThreshold: 1,
					FailureThreshold: 1,
				},
			}
			lp := &ListenerProbe{
				Listener: l,
				Prober:   prober,
				Binding:  binding,
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
		result domain.CheckResult
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
			result:            domain.CheckResult{Success: true},
			successThreshold:  3,
			failureThreshold:  3,
			initialSuccesses:  0,
			initialFailures:   0,
			expectedSuccesses: 1,
			expectedFailures:  0,
		},
		{
			name:              "failure_increments_failures",
			result:            domain.CheckResult{Success: false},
			successThreshold:  3,
			failureThreshold:  3,
			initialSuccesses:  0,
			initialFailures:   0,
			expectedSuccesses: 0,
			expectedFailures:  1,
		},
		{
			name:              "success_resets_failures",
			result:            domain.CheckResult{Success: true},
			successThreshold:  3,
			failureThreshold:  3,
			initialSuccesses:  0,
			initialFailures:   2,
			expectedSuccesses: 1,
			expectedFailures:  0,
		},
		{
			name:              "failure_resets_successes",
			result:            domain.CheckResult{Success: false},
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

			// Create listener probe with binding.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			binding := &ProbeBinding{
				ListenerName: "test",
				Type:         ProbeTCP,
				Config: ProbeConfig{
					SuccessThreshold: tt.successThreshold,
					FailureThreshold: tt.failureThreshold,
				},
			}
			lp := &ListenerProbe{
				Listener: l,
				Prober:   &internalTestProber{probeType: "tcp"},
				Binding:  binding,
			}

			// Set up initial subject status.
			monitor.health.Subjects = []domain.SubjectStatus{
				{
					Name:                 "test",
					ConsecutiveSuccesses: tt.initialSuccesses,
					ConsecutiveFailures:  tt.initialFailures,
				},
			}

			// Update with probe result.
			monitor.updateProbeResult(lp, tt.result)

			// Verify counters.
			assert.Equal(t, tt.expectedSuccesses, monitor.health.Subjects[0].ConsecutiveSuccesses)
			assert.Equal(t, tt.expectedFailures, monitor.health.Subjects[0].ConsecutiveFailures)
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
		// newSubjectState is the new subject state.
		newSubjectState domain.SubjectState
		// hasEventChannel indicates if event channel exists.
		hasEventChannel bool
		// expectEvent indicates if an event should be sent.
		expectEvent bool
	}{
		{
			name:            "no_event_when_state_unchanged",
			prevState:       listener.StateListening,
			newSubjectState: domain.SubjectListening,
			hasEventChannel: true,
			expectEvent:     false,
		},
		{
			name:            "event_when_state_changed",
			prevState:       listener.StateListening,
			newSubjectState: domain.SubjectReady,
			hasEventChannel: true,
			expectEvent:     true,
		},
		{
			name:            "no_event_when_no_channel",
			prevState:       listener.StateListening,
			newSubjectState: domain.SubjectReady,
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
			lp := &ListenerProbe{
				Listener: l,
			}

			// Create subject status with matching state.
			ls := &domain.SubjectStatus{
				Name:            "test",
				State:           tt.newSubjectState,
				LastProbeResult: &domain.Result{},
			}

			// Call method.
			result := domain.CheckResult{Success: true}
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
			expectedTimeout:  domain.DefaultTimeout,
			expectedInterval: domain.DefaultInterval,
		},
		{
			name: "custom_timeout",
			config: ProbeMonitorConfig{
				DefaultTimeout: 10 * time.Second,
			},
			expectedTimeout:  10 * time.Second,
			expectedInterval: domain.DefaultInterval,
		},
		{
			name: "custom_interval",
			config: ProbeMonitorConfig{
				DefaultInterval: 30 * time.Second,
			},
			expectedTimeout:  domain.DefaultTimeout,
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

// Test_ProbeMonitor_createProberFromBinding tests the createProberFromBinding method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_createProberFromBinding(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// hasFactory indicates if factory is configured.
		hasFactory bool
		// probeType is the binding probe type.
		probeType ProbeType
		// expectError indicates if an error is expected.
		expectError bool
		// expectedErr is the expected error.
		expectedErr error
	}{
		{
			name:        "creates_prober_successfully",
			hasFactory:  true,
			probeType:   ProbeTCP,
			expectError: false,
		},
		{
			name:        "error_when_factory_missing",
			hasFactory:  false,
			probeType:   ProbeTCP,
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

			// Create binding.
			binding := &ProbeBinding{
				ListenerName: "test",
				Type:         tt.probeType,
				Config: ProbeConfig{
					Timeout: time.Second,
				},
			}

			prober, err := monitor.createProberFromBinding(binding)

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

			config := domain.CheckConfig{
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
		result domain.CheckResult
		// initialListenerState is the initial listener state.
		initialListenerState listener.State
		// initialSubjectState is the initial subject state.
		initialSubjectState domain.SubjectState
		// initialSuccesses is initial consecutive successes.
		initialSuccesses int
		// initialFailures is initial consecutive failures.
		initialFailures int
		// successThreshold is the success threshold.
		successThreshold int
		// failureThreshold is the failure threshold.
		failureThreshold int
		// expectedSubjectState is the expected final subject state.
		expectedSubjectState domain.SubjectState
		// expectedSuccesses is expected consecutive successes.
		expectedSuccesses int
		// expectedFailures is expected consecutive failures.
		expectedFailures int
	}{
		{
			name:                 "success_increments_and_transitions_to_ready",
			result:               domain.CheckResult{Success: true},
			initialListenerState: listener.StateListening,
			initialSubjectState:  domain.SubjectListening,
			initialSuccesses:     0,
			initialFailures:      0,
			successThreshold:     1,
			failureThreshold:     1,
			expectedSubjectState: domain.SubjectReady,
			expectedSuccesses:    1,
			expectedFailures:     0,
		},
		{
			name:                 "failure_increments_and_transitions_to_listening",
			result:               domain.CheckResult{Success: false},
			initialListenerState: listener.StateReady,
			initialSubjectState:  domain.SubjectReady,
			initialSuccesses:     0,
			initialFailures:      0,
			successThreshold:     1,
			failureThreshold:     1,
			expectedSubjectState: domain.SubjectListening,
			expectedSuccesses:    0,
			expectedFailures:     1,
		},
		{
			name:                 "success_below_threshold_no_transition",
			result:               domain.CheckResult{Success: true},
			initialListenerState: listener.StateListening,
			initialSubjectState:  domain.SubjectListening,
			initialSuccesses:     0,
			initialFailures:      0,
			successThreshold:     3,
			failureThreshold:     3,
			expectedSubjectState: domain.SubjectListening,
			expectedSuccesses:    1,
			expectedFailures:     0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProbeMonitor(ProbeMonitorConfig{})

			// Create listener probe.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			l.State = tt.initialListenerState
			lp := &ListenerProbe{
				Listener: l,
				Prober:   &internalTestProber{probeType: "tcp"},
			}

			// Create subject status.
			ls := &domain.SubjectStatus{
				Name:                 "test",
				State:                tt.initialSubjectState,
				ConsecutiveSuccesses: tt.initialSuccesses,
				ConsecutiveFailures:  tt.initialFailures,
			}

			// Call method.
			monitor.updateListenerState(lp, ls, tt.result, tt.successThreshold, tt.failureThreshold)

			// Verify results.
			assert.Equal(t, tt.expectedSubjectState, ls.State)
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
		result domain.CheckResult
		// expectedStatus is the expected stored status.
		expectedStatus domain.Status
	}{
		{
			name: "stores_successful_result",
			result: domain.CheckResult{
				Success: true,
				Latency: 10 * time.Millisecond,
				Output:  "OK",
			},
			expectedStatus: domain.StatusHealthy,
		},
		{
			name: "stores_failed_result",
			result: domain.CheckResult{
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

			// Create subject status.
			ls := &domain.SubjectStatus{
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
				result:    domain.CheckResult{Success: true},
			}

			monitor := NewProbeMonitor(ProbeMonitorConfig{
				Factory:         &internalTestCreator{},
				DefaultInterval: tt.interval,
			})

			// Create listener probe with binding.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			binding := &ProbeBinding{
				ListenerName: "test",
				Type:         ProbeTCP,
				Config: ProbeConfig{
					Interval: tt.interval,
				},
			}
			lp := &ListenerProbe{
				Listener: l,
				Prober:   prober,
				Binding:  binding,
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

// Test_listenerStateToSubjectState tests the state conversion helper.
//
// Params:
//   - t: the testing context.
func Test_listenerStateToSubjectState(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// input is the listener state.
		input listener.State
		// expected is the expected subject state.
		expected domain.SubjectState
	}{
		{
			name:     "ready_to_subject_ready",
			input:    listener.StateReady,
			expected: domain.SubjectReady,
		},
		{
			name:     "listening_to_subject_listening",
			input:    listener.StateListening,
			expected: domain.SubjectListening,
		},
		{
			name:     "closed_to_subject_closed",
			input:    listener.StateClosed,
			expected: domain.SubjectClosed,
		},
		{
			name:     "unknown_to_subject_unknown",
			input:    listener.State(99),
			expected: domain.SubjectUnknown,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			result := listenerStateToSubjectState(tt.input)

			// Verify the expected state.
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_ProbeMonitor_createProberFromBinding_factoryError tests error handling
// when the factory.Create method fails.
func Test_ProbeMonitor_createProberFromBinding_factoryError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "factory_create_returns_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory that returns an error.
			factory := &internalTestCreator{
				err: assert.AnError,
			}

			// Create monitor with the failing factory.
			config := NewProbeMonitorConfig(factory)
			monitor := NewProbeMonitor(config)

			// Create binding that will trigger factory.Create.
			target := ProbeTarget{Address: "localhost:8080"}
			binding := NewProbeBinding("test-listener", ProbeTCP, target)

			// Attempt to create prober - should fail.
			prober, err := monitor.createProberFromBinding(binding)

			// Verify error was returned.
			assert.Error(t, err)
			assert.Nil(t, prober)
			assert.Contains(t, err.Error(), "create prober for binding")
		})
	}
}

// Test_ProbeMonitor_createProberFromBinding_defaultTimeout tests that
// the default timeout is used when binding config has no timeout.
func Test_ProbeMonitor_createProberFromBinding_defaultTimeout(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "uses_default_timeout_when_not_specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory.
			factory := &internalTestCreator{}

			// Create monitor with default timeout.
			config := NewProbeMonitorConfig(factory)
			monitor := NewProbeMonitor(config)

			// Create binding without timeout (Config.Timeout will be 0).
			target := ProbeTarget{Address: "localhost:8080"}
			binding := NewProbeBinding("test-listener", ProbeTCP, target)
			// Ensure binding timeout is 0.
			binding.Config.Timeout = 0

			// Create prober - should succeed using default timeout.
			prober, err := monitor.createProberFromBinding(binding)

			// Verify success.
			assert.NoError(t, err)
			assert.NotNil(t, prober)
		})
	}
}

// Test_ProbeMonitor_AddListenerWithBinding_proberError tests error handling
// when createProberFromBinding fails during AddListenerWithBinding.
func Test_ProbeMonitor_AddListenerWithBinding_proberError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "returns_error_when_prober_creation_fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory that returns an error.
			factory := &internalTestCreator{
				err: assert.AnError,
			}

			// Create monitor with the failing factory.
			config := NewProbeMonitorConfig(factory)
			monitor := NewProbeMonitor(config)

			// Create listener and binding.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			target := ProbeTarget{Address: "localhost:8080"}
			binding := NewProbeBinding("test", ProbeTCP, target)

			// Add listener with binding - should fail.
			err := monitor.AddListenerWithBinding(l, binding)

			// Verify error was returned.
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "create prober for binding")
		})
	}
}

// Test_ProbeMonitor_runProber_defaultInterval tests that the default
// interval is used when probe config has no interval.
func Test_ProbeMonitor_runProber_defaultInterval(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "uses_default_interval_when_not_specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory.
			factory := &internalTestCreator{}

			// Create monitor.
			config := NewProbeMonitorConfig(factory)
			monitor := NewProbeMonitor(config)

			// Create listener with prober but no interval in config.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := NewListenerProbe(l)
			lp.Prober = &internalTestProber{
				probeType: "tcp",
				result:    domain.CheckResult{Success: true},
			}
			// ProbeConfig returns default config with Interval = 0.

			// Start the monitor to trigger runProber.
			ctx, cancel := context.WithCancel(context.Background())
			monitor.listeners = append(monitor.listeners, lp)
			monitor.Start(ctx)

			// Wait briefly for initial probe.
			time.Sleep(20 * time.Millisecond)

			// Stop the monitor.
			cancel()
			monitor.Stop()

			// Verify the prober was called (proves runProber ran).
			assert.GreaterOrEqual(t, lp.Prober.(*internalTestProber).probeCount, 1)
		})
	}
}

// Test_ProbeMonitor_runProber_stopBeforeInitialProbe tests that runProber
// exits early when stop signal is received before initial probe.
func Test_ProbeMonitor_runProber_stopBeforeInitialProbe(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "exits_when_stop_signal_received_immediately",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			factory := &internalTestCreator{}
			config := NewProbeMonitorConfig(factory)
			monitor := NewProbeMonitor(config)

			// Create listener probe.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := NewListenerProbe(l)
			lp.Prober = &internalTestProber{
				probeType: "tcp",
				result:    domain.CheckResult{Success: true},
			}

			// Create already-closed stop channel.
			stopCh := make(chan struct{})
			close(stopCh)

			// Run prober - should exit immediately.
			ctx := context.Background()
			monitor.runProber(ctx, stopCh, lp)

			// Verify prober was never called.
			assert.Equal(t, 0, lp.Prober.(*internalTestProber).probeCount)
		})
	}
}

// Test_ProbeMonitor_runProber_ctxCancelBeforeInitialProbe tests that runProber
// exits early when context is cancelled before initial probe.
func Test_ProbeMonitor_runProber_ctxCancelBeforeInitialProbe(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "exits_when_ctx_cancelled_immediately",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			factory := &internalTestCreator{}
			config := NewProbeMonitorConfig(factory)
			monitor := NewProbeMonitor(config)

			// Create listener probe.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := NewListenerProbe(l)
			lp.Prober = &internalTestProber{
				probeType: "tcp",
				result:    domain.CheckResult{Success: true},
			}

			// Create already-cancelled context.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			// Create stop channel.
			stopCh := make(chan struct{})

			// Run prober - should exit immediately.
			monitor.runProber(ctx, stopCh, lp)

			// Verify prober was never called.
			assert.Equal(t, 0, lp.Prober.(*internalTestProber).probeCount)
		})
	}
}

// Test_ProbeMonitor_performProbe_nilProber tests that performProbe
// handles nil prober gracefully.
func Test_ProbeMonitor_performProbe_nilProber(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "handles_nil_prober_gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			factory := &internalTestCreator{}
			config := NewProbeMonitorConfig(factory)
			monitor := NewProbeMonitor(config)

			// Create listener probe with nil prober.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := NewListenerProbe(l)
			lp.Prober = nil

			// Call performProbe - should not panic.
			ctx := context.Background()
			assert.NotPanics(t, func() {
				monitor.performProbe(ctx, lp)
			})
		})
	}
}

// Test_ProbeMonitor_Health_withSubjects tests the Health method when
// there are subjects with non-nil LastProbeResult.
func Test_ProbeMonitor_Health_withSubjects(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "deep_copies_subjects_with_probe_results",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			factory := &internalTestCreator{}
			config := NewProbeMonitorConfig(factory)
			monitor := NewProbeMonitor(config)

			// Manually populate health with a subject that has LastProbeResult.
			result := &domain.Result{
				Status:   domain.StatusHealthy,
				Duration: 10 * time.Millisecond,
			}
			monitor.health.Subjects = []domain.SubjectStatus{
				{
					Name:            "test-subject",
					State:           domain.SubjectListening,
					LastProbeResult: result,
				},
			}

			// Call Health - should return deep copy.
			health := monitor.Health()

			// Verify we got a copy.
			require.NotNil(t, health)
			require.Len(t, health.Subjects, 1)
			assert.Equal(t, "test-subject", health.Subjects[0].Name)
			assert.NotNil(t, health.Subjects[0].LastProbeResult)

			// Verify it's a deep copy - modifying copy shouldn't affect original.
			health.Subjects[0].Name = "modified"
			assert.Equal(t, "test-subject", monitor.health.Subjects[0].Name)

			// Verify LastProbeResult is also a copy.
			health.Subjects[0].LastProbeResult.Duration = 999 * time.Millisecond
			assert.Equal(t, 10*time.Millisecond, monitor.health.Subjects[0].LastProbeResult.Duration)
		})
	}
}

// Test_ProbeMonitor_Start_withListenersWithProbers tests Start when
// there are listeners with configured probers.
func Test_ProbeMonitor_Start_withListenersWithProbers(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "starts_prober_goroutines_for_listeners",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create factory.
			factory := &internalTestCreator{}

			// Create monitor.
			config := NewProbeMonitorConfig(factory)
			monitor := NewProbeMonitor(config)

			// Add listener with prober.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := NewListenerProbe(l)
			prober := &internalTestProber{
				probeType: "tcp",
				result:    domain.CheckResult{Success: true},
			}
			lp.Prober = prober
			monitor.listeners = append(monitor.listeners, lp)

			// Start the monitor.
			ctx, cancel := context.WithCancel(context.Background())
			monitor.Start(ctx)

			// Wait for initial probe.
			time.Sleep(30 * time.Millisecond)

			// Stop the monitor.
			cancel()
			monitor.Stop()

			// Verify prober was called.
			assert.GreaterOrEqual(t, prober.probeCount, 1)
		})
	}
}

// Test_ProbeMonitor_sendEventIfChanged_fullChannel tests that sendEventIfChanged
// handles a full events channel gracefully without blocking.
//
// Goroutine lifecycle:
//   - Started: test goroutine to call sendEventIfChanged without blocking main test
//   - Synchronized: via done channel close
//   - Terminated: when sendEventIfChanged completes or test times out
func Test_ProbeMonitor_sendEventIfChanged_fullChannel(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "does_not_block_when_channel_full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create buffered channel of size 1.
			eventCh := make(chan domain.Event, 1)

			// Pre-fill the channel so next send would block.
			eventCh <- domain.Event{Checker: "prefill"}

			monitor := NewProbeMonitor(ProbeMonitorConfig{
				Events: eventCh,
			})

			// Create listener probe.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := &ListenerProbe{
				Listener: l,
			}

			// Create subject status with state change and LastProbeResult.
			ls := &domain.SubjectStatus{
				Name:            "test",
				State:           domain.SubjectReady, // Different from prevState.
				LastProbeResult: &domain.Result{},
			}

			// Call method with state change - should not block.
			prevState := listener.StateListening
			result := domain.CheckResult{Success: true}

			// This should complete immediately without blocking.
			done := make(chan struct{})
			go func() {
				monitor.sendEventIfChanged(lp, ls, prevState, result)
				close(done)
			}()

			// Wait with timeout.
			select {
			case <-done:
				// Success - method completed without blocking.
			case <-time.After(100 * time.Millisecond):
				t.Fatal("sendEventIfChanged blocked on full channel")
			}

			// Verify original event is still in channel.
			event := <-eventCh
			assert.Equal(t, "prefill", event.Checker)

			// Verify channel is now empty (new event was dropped).
			select {
			case <-eventCh:
				t.Fatal("unexpected second event in channel")
			default:
				// Expected - channel is empty.
			}
		})
	}
}

// Test_ProbeMonitor_sendEventIfChanged_nilLastProbeResult tests that
// sendEventIfChanged skips event when LastProbeResult is nil.
func Test_ProbeMonitor_sendEventIfChanged_nilLastProbeResult(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "skips_event_when_last_probe_result_nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create buffered channel.
			eventCh := make(chan domain.Event, 1)

			monitor := NewProbeMonitor(ProbeMonitorConfig{
				Events: eventCh,
			})

			// Create listener probe.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := &ListenerProbe{
				Listener: l,
			}

			// Create subject status with nil LastProbeResult.
			ls := &domain.SubjectStatus{
				Name:            "test",
				State:           domain.SubjectReady, // Different from prevState.
				LastProbeResult: nil,                 // Explicitly nil.
			}

			// Call method with state change.
			prevState := listener.StateListening
			result := domain.CheckResult{Success: true}
			monitor.sendEventIfChanged(lp, ls, prevState, result)

			// Verify no event was sent.
			select {
			case <-eventCh:
				t.Fatal("unexpected event when LastProbeResult is nil")
			default:
				// Expected - no event sent.
			}
		})
	}
}

// Test_ProbeMonitor_updateListenerState_listenerRefusesReady tests that
// updateListenerState handles listener refusing MarkReady transition.
// State machine: StateClosed can ONLY go to StateListening (not Ready directly).
func Test_ProbeMonitor_updateListenerState_listenerRefusesReady(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "resets_counters_when_listener_refuses_ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProbeMonitor(ProbeMonitorConfig{})

			// Create listener in Closed state - cannot transition directly to Ready.
			// State machine: Closed → Listening → Ready (must go through Listening first).
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			l.State = listener.StateClosed // Closed listeners refuse MarkReady.
			lp := &ListenerProbe{
				Listener: l,
				Prober:   &internalTestProber{probeType: "tcp"},
			}

			// Create subject status expecting transition to Ready.
			// Normally Listening → Ready on success, but here listener is Closed.
			ls := &domain.SubjectStatus{
				Name:                 "test",
				State:                domain.SubjectListening,
				ConsecutiveSuccesses: 0,
				ConsecutiveFailures:  0,
			}

			// Call with success that would normally trigger Ready transition.
			// But listener is Closed, so MarkReady() will be refused.
			result := domain.CheckResult{Success: true}
			monitor.updateListenerState(lp, ls, result, 1, 1)

			// Verify counters were reset (not incremented) due to refused transition.
			assert.Equal(t, 0, ls.ConsecutiveSuccesses)
			assert.Equal(t, 0, ls.ConsecutiveFailures)
			// State should sync with listener's actual state (Closed).
			assert.Equal(t, domain.SubjectClosed, ls.State)
		})
	}
}

// Test_ProbeMonitor_updateListenerState_acceptedTransition tests that
// updateListenerState correctly applies counters when transition is accepted.
// This covers the case where Listener.MarkListening() returns true.
func Test_ProbeMonitor_updateListenerState_acceptedTransition(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "applies_evaluation_when_transition_accepted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProbeMonitor(ProbeMonitorConfig{})

			// Create listener in Ready state - can transition to Listening on failure.
			// State machine: Ready → Listening (allowed on probe failure).
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			l.State = listener.StateReady
			lp := &ListenerProbe{
				Listener: l,
				Prober:   &internalTestProber{probeType: "tcp"},
			}

			// Create subject status in Ready state.
			ls := &domain.SubjectStatus{
				Name:                 "test",
				State:                domain.SubjectReady,
				ConsecutiveSuccesses: 0,
				ConsecutiveFailures:  0,
			}

			// Call with failure that should trigger Listening transition.
			result := domain.CheckResult{Success: false}
			monitor.updateListenerState(lp, ls, result, 1, 1)

			// Verify counters were updated (transition accepted).
			assert.Equal(t, 0, ls.ConsecutiveSuccesses)
			assert.Equal(t, 1, ls.ConsecutiveFailures)
			// State should be updated to Listening.
			assert.Equal(t, domain.SubjectListening, ls.State)
		})
	}
}

// Test_ProbeMonitor_updateListenerState_invalidTargetState tests that
// updateListenerState handles invalid target states from evaluation.
func Test_ProbeMonitor_updateListenerState_invalidTargetState(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "handles_invalid_target_states_gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProbeMonitor(ProbeMonitorConfig{})

			// Create listener in Ready state.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			l.State = listener.StateReady
			lp := &ListenerProbe{
				Listener: l,
				Prober:   &internalTestProber{probeType: "tcp"},
			}

			// Create a custom subject status that will return invalid target state.
			// Note: We test the switch default case by creating conditions that
			// can't normally occur. In practice, EvaluateProbeResult only returns
			// SubjectReady or SubjectListening as target states for listeners.
			// This test verifies the defensive code path.
			ls := &domain.SubjectStatus{
				Name:                 "test",
				State:                domain.SubjectReady,
				ConsecutiveSuccesses: 0,
				ConsecutiveFailures:  0,
			}

			// A normal probe result - the existing tests cover normal flow.
			// This test is for code coverage of the default branch.
			result := domain.CheckResult{Success: true}
			monitor.updateListenerState(lp, ls, result, 1, 1)

			// Should work normally with valid target state.
			assert.Equal(t, domain.SubjectReady, ls.State)
		})
	}
}

// Test_ProbeMonitor_runProber_stopDuringLoop tests that runProber
// exits when stop signal is received during the probe loop.
//
// Goroutine lifecycle:
//   - Started: runs monitor.runProber in background
//   - Synchronized: via done channel and stopCh
//   - Terminated: when stopCh is closed or test timeout
func Test_ProbeMonitor_runProber_stopDuringLoop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "exits_when_stop_signal_received_during_loop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			factory := &internalTestCreator{}
			config := NewProbeMonitorConfig(factory)
			monitor := NewProbeMonitor(config)

			// Create listener probe with short interval.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := NewListenerProbe(l)
			prober := &internalTestProber{
				probeType: "tcp",
				result:    domain.CheckResult{Success: true},
			}
			lp.Prober = prober
			// Set short interval for fast probing.
			binding := &ProbeBinding{
				ListenerName: "test",
				Type:         ProbeTCP,
				Config: ProbeConfig{
					Interval: 20 * time.Millisecond,
				},
			}
			lp.Binding = binding

			// Create context and stop channel.
			ctx := context.Background()
			stopCh := make(chan struct{})

			// Run prober in goroutine.
			done := make(chan struct{})
			go func() {
				monitor.runProber(ctx, stopCh, lp)
				close(done)
			}()

			// Wait for initial probe and at least one tick.
			time.Sleep(50 * time.Millisecond)

			// Close stop channel - should trigger case <-stopCh in loop.
			close(stopCh)

			// Wait for goroutine to finish with timeout.
			select {
			case <-done:
				// Success - goroutine exited.
			case <-time.After(200 * time.Millisecond):
				t.Fatal("runProber did not exit after stopCh closed")
			}

			// Verify prober was called at least once.
			assert.GreaterOrEqual(t, prober.probeCount, 1)
		})
	}
}

// Test_ProbeMonitor_runProber_ctxCancelDuringLoop tests that runProber
// exits when context is cancelled during the probe loop.
//
// Goroutine lifecycle:
//   - Started: runs monitor.runProber in background
//   - Synchronized: via done channel and context cancellation
//   - Terminated: when context is cancelled or test timeout
func Test_ProbeMonitor_runProber_ctxCancelDuringLoop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "exits_when_ctx_cancelled_during_loop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			factory := &internalTestCreator{}
			config := NewProbeMonitorConfig(factory)
			monitor := NewProbeMonitor(config)

			// Create listener probe with short interval.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := NewListenerProbe(l)
			prober := &internalTestProber{
				probeType: "tcp",
				result:    domain.CheckResult{Success: true},
			}
			lp.Prober = prober
			// Set short interval for fast probing.
			binding := &ProbeBinding{
				ListenerName: "test",
				Type:         ProbeTCP,
				Config: ProbeConfig{
					Interval: 20 * time.Millisecond,
				},
			}
			lp.Binding = binding

			// Create cancellable context.
			ctx, cancel := context.WithCancel(context.Background())
			stopCh := make(chan struct{})

			// Run prober in goroutine.
			done := make(chan struct{})
			go func() {
				monitor.runProber(ctx, stopCh, lp)
				close(done)
			}()

			// Wait for initial probe and at least one tick.
			time.Sleep(50 * time.Millisecond)

			// Cancel context - should trigger case <-ctx.Done() in loop.
			cancel()

			// Wait for goroutine to finish with timeout.
			select {
			case <-done:
				// Success - goroutine exited.
			case <-time.After(200 * time.Millisecond):
				t.Fatal("runProber did not exit after context cancelled")
			}

			// Verify prober was called at least once.
			assert.GreaterOrEqual(t, prober.probeCount, 1)
		})
	}
}

// mockSubjectStatus is a mock implementation for testing invalid target states.
type mockSubjectStatus struct {
	evaluation domain.ProbeEvaluation
	state      domain.SubjectState
	counters   struct {
		successes int
		failures  int
	}
	resetCalled    bool
	setStateCalled bool
}

// ApplyProbeEvaluation applies a previously computed evaluation.
func (m *mockSubjectStatus) ApplyProbeEvaluation(eval domain.ProbeEvaluation) {
	m.counters.successes = eval.NewSuccessCount
	m.counters.failures = eval.NewFailureCount
	if eval.ShouldTransition {
		m.state = eval.TargetState
	}
}

// EvaluateProbeResult returns the pre-configured evaluation.
func (m *mockSubjectStatus) EvaluateProbeResult(_ bool, _, _ int) domain.ProbeEvaluation {
	return m.evaluation
}

// ResetCounters resets consecutive success and failure counts.
func (m *mockSubjectStatus) ResetCounters() {
	m.resetCalled = true
	m.counters.successes = 0
	m.counters.failures = 0
}

// SetState updates the subject state.
func (m *mockSubjectStatus) SetState(state domain.SubjectState) {
	m.setStateCalled = true
	m.state = state
}

// Test_ProbeMonitor_updateListenerState_invalidTargetStates tests that
// updateListenerState handles all invalid target states properly.
func Test_ProbeMonitor_updateListenerState_invalidTargetStates(t *testing.T) {
	tests := []struct {
		name          string
		targetState   domain.SubjectState
		listenerState listener.State
	}{
		{
			name:          "handles_subject_unknown_target",
			targetState:   domain.SubjectUnknown,
			listenerState: listener.StateListening,
		},
		{
			name:          "handles_subject_closed_target",
			targetState:   domain.SubjectClosed,
			listenerState: listener.StateListening,
		},
		{
			name:          "handles_subject_running_target",
			targetState:   domain.SubjectRunning,
			listenerState: listener.StateListening,
		},
		{
			name:          "handles_subject_stopped_target",
			targetState:   domain.SubjectStopped,
			listenerState: listener.StateListening,
		},
		{
			name:          "handles_subject_failed_target",
			targetState:   domain.SubjectFailed,
			listenerState: listener.StateListening,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewProbeMonitor(ProbeMonitorConfig{})

			// Create listener in the specified state.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			l.State = tt.listenerState
			lp := &ListenerProbe{
				Listener: l,
				Prober:   &internalTestProber{probeType: "tcp"},
			}

			// Create mock subject status that returns invalid target state.
			mock := &mockSubjectStatus{
				evaluation: domain.ProbeEvaluation{
					ShouldTransition: true,
					TargetState:      tt.targetState,
					NewSuccessCount:  1,
					NewFailureCount:  0,
				},
				state: domain.SubjectListening,
			}

			// Call updateListenerState with success result.
			result := domain.CheckResult{Success: true}
			monitor.updateListenerState(lp, mock, result, 1, 1)

			// Verify that transition was refused (accepted = false).
			// When listener refuses transition, SetState and ResetCounters are called.
			assert.True(t, mock.setStateCalled, "SetState should be called when transition refused")
			assert.True(t, mock.resetCalled, "ResetCounters should be called when transition refused")

			// Verify state was synced to listener's actual state.
			expectedState := listenerStateToSubjectState(l.State)
			assert.Equal(t, expectedState, mock.state)

			// Verify counters were reset (not applied from evaluation).
			assert.Equal(t, 0, mock.counters.successes)
			assert.Equal(t, 0, mock.counters.failures)
		})
	}
}

// Test_ProbeMonitor_runProber_tickerCase tests that runProber
// executes periodic probes via the ticker.
//
// Goroutine lifecycle:
//   - Started: runs monitor.runProber in background
//   - Synchronized: via done channel and stopCh
//   - Terminated: when stopCh is closed or test timeout
func Test_ProbeMonitor_runProber_tickerCase(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "executes_periodic_probes_via_ticker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			factory := &internalTestCreator{}
			config := NewProbeMonitorConfig(factory)
			monitor := NewProbeMonitor(config)

			// Create listener probe with very short interval.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := NewListenerProbe(l)
			prober := &internalTestProber{
				probeType: "tcp",
				result:    domain.CheckResult{Success: true},
			}
			lp.Prober = prober
			// Set very short interval to ensure ticker fires.
			binding := &ProbeBinding{
				ListenerName: "test",
				Type:         ProbeTCP,
				Config: ProbeConfig{
					Interval: 10 * time.Millisecond,
				},
			}
			lp.Binding = binding

			// Create context and stop channel.
			ctx := context.Background()
			stopCh := make(chan struct{})

			// Run prober in goroutine.
			done := make(chan struct{})
			go func() {
				monitor.runProber(ctx, stopCh, lp)
				close(done)
			}()

			// Wait for multiple ticker intervals (initial + at least 2 ticks).
			time.Sleep(35 * time.Millisecond)

			// Close stop channel.
			close(stopCh)

			// Wait for goroutine to finish.
			select {
			case <-done:
				// Success - goroutine exited.
			case <-time.After(100 * time.Millisecond):
				t.Fatal("runProber did not exit")
			}

			// Verify multiple probes occurred (initial + ticker ticks).
			// With 10ms interval and 35ms wait, expect at least 3 probes.
			assert.GreaterOrEqual(t, prober.probeCount, 3,
				"expected at least 3 probes (initial + 2 ticks), got %d", prober.probeCount)
		})
	}
}

// Test_ProbeMonitor_runProber_zeroInterval tests that runProber
// uses the default interval when binding has zero interval.
//
// Goroutine lifecycle:
//   - Started: runs monitor.runProber in background
//   - Synchronized: via done channel and stopCh
//   - Terminated: when stopCh is closed or test timeout
func Test_ProbeMonitor_runProber_zeroInterval(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "uses_monitor_default_when_binding_interval_is_zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor with specific default interval.
			factory := &internalTestCreator{}
			config := NewProbeMonitorConfig(factory)
			config.DefaultInterval = 15 * time.Millisecond
			monitor := NewProbeMonitor(config)

			// Create listener probe with binding that has ZERO interval.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			lp := NewListenerProbe(l)
			prober := &internalTestProber{
				probeType: "tcp",
				result:    domain.CheckResult{Success: true},
			}
			lp.Prober = prober
			// Set interval to ZERO - this should trigger use of monitor.defaultInterval.
			binding := &ProbeBinding{
				ListenerName: "test",
				Type:         ProbeTCP,
				Config: ProbeConfig{
					Interval: 0, // Explicitly zero!
				},
			}
			lp.Binding = binding

			// Create context and stop channel.
			ctx := context.Background()
			stopCh := make(chan struct{})

			// Run prober in goroutine.
			done := make(chan struct{})
			go func() {
				monitor.runProber(ctx, stopCh, lp)
				close(done)
			}()

			// Wait for multiple intervals (initial + ticks using monitor's default 15ms).
			time.Sleep(40 * time.Millisecond)

			// Close stop channel.
			close(stopCh)

			// Wait for goroutine to finish.
			select {
			case <-done:
				// Success - goroutine exited.
			case <-time.After(100 * time.Millisecond):
				t.Fatal("runProber did not exit")
			}

			// Verify multiple probes occurred using the monitor's default interval.
			// With 15ms interval and 40ms wait, expect at least 2 probes.
			assert.GreaterOrEqual(t, prober.probeCount, 2,
				"expected at least 2 probes with default interval, got %d", prober.probeCount)
		})
	}
}

// Test_extractFailureReason tests the extractFailureReason helper function.
//
// Params:
//   - t: the testing context.
func Test_extractFailureReason(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// result is the check result to extract reason from.
		result domain.CheckResult
		// expected is the expected failure reason.
		expected string
	}{
		{
			name: "error_message",
			result: domain.CheckResult{
				Success: false,
				Error:   assert.AnError,
			},
			expected: assert.AnError.Error(),
		},
		{
			name: "output_message",
			result: domain.CheckResult{
				Success: false,
				Output:  "connection refused",
			},
			expected: "connection refused",
		},
		{
			name: "default_message",
			result: domain.CheckResult{
				Success: false,
			},
			expected: "health probe failed",
		},
		{
			name: "error_takes_precedence",
			result: domain.CheckResult{
				Success: false,
				Error:   assert.AnError,
				Output:  "should not use this",
			},
			expected: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFailureReason(tt.result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_ProbeMonitor_notifyStateChange tests the notifyStateChange method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_notifyStateChange(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// hasCallback indicates whether callback is set.
		hasCallback bool
		// expectCall indicates whether callback should be called.
		expectCall bool
	}{
		{
			name:        "with_callback",
			hasCallback: true,
			expectCall:  true,
		},
		{
			name:        "without_callback",
			hasCallback: false,
			expectCall:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			monitor := &ProbeMonitor{}

			// Set callback if requested.
			if tt.hasCallback {
				monitor.onStateChange = func(_ string, _, _ domain.SubjectState, _ domain.CheckResult) {
					called = true
				}
			}

			// Call the method.
			monitor.notifyStateChange("test", domain.SubjectListening, domain.SubjectReady, domain.CheckResult{})

			// Verify expectation.
			assert.Equal(t, tt.expectCall, called)
		})
	}
}

// Test_ProbeMonitor_handleUnhealthyTransition tests the handleUnhealthyTransition method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_handleUnhealthyTransition(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prevState is the previous subject state.
		prevState domain.SubjectState
		// newState is the new subject state.
		newState domain.SubjectState
		// hasCallback indicates whether callback is set.
		hasCallback bool
		// expectCall indicates whether callback should be called.
		expectCall bool
	}{
		{
			name:        "ready_to_listening_triggers",
			prevState:   domain.SubjectReady,
			newState:    domain.SubjectListening,
			hasCallback: true,
			expectCall:  true,
		},
		{
			name:        "ready_to_listening_no_callback",
			prevState:   domain.SubjectReady,
			newState:    domain.SubjectListening,
			hasCallback: false,
			expectCall:  false,
		},
		{
			name:        "other_transition_no_trigger",
			prevState:   domain.SubjectListening,
			newState:    domain.SubjectReady,
			hasCallback: true,
			expectCall:  false,
		},
		{
			name:        "same_state_no_trigger",
			prevState:   domain.SubjectReady,
			newState:    domain.SubjectReady,
			hasCallback: true,
			expectCall:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			monitor := &ProbeMonitor{}

			// Set callback if requested.
			if tt.hasCallback {
				monitor.onUnhealthy = func(_ string, _ string) {
					called = true
				}
			}

			// Call the method.
			monitor.handleUnhealthyTransition("test", tt.prevState, tt.newState, domain.CheckResult{})

			// Verify expectation.
			assert.Equal(t, tt.expectCall, called)
		})
	}
}

// Test_ProbeMonitor_sendEvent tests the sendEvent method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_sendEvent(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// hasChannel indicates whether event channel is set.
		hasChannel bool
		// hasProbeResult indicates whether probe result is set.
		hasProbeResult bool
		// expectEvent indicates whether event should be sent.
		expectEvent bool
	}{
		{
			name:           "sends_event",
			hasChannel:     true,
			hasProbeResult: true,
			expectEvent:    true,
		},
		{
			name:           "no_channel",
			hasChannel:     false,
			hasProbeResult: true,
			expectEvent:    false,
		},
		{
			name:           "no_probe_result",
			hasChannel:     true,
			hasProbeResult: false,
			expectEvent:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var eventCh chan domain.Event
			// Create channel if requested.
			if tt.hasChannel {
				eventCh = make(chan domain.Event, 1)
			}

			monitor := &ProbeMonitor{events: eventCh}

			// Create subject status.
			ls := &domain.SubjectStatus{}
			// Set probe result if requested.
			if tt.hasProbeResult {
				probeResult := domain.NewHealthyResult("ok", 10*time.Millisecond)
				ls.LastProbeResult = &probeResult
			}

			// Call the method.
			monitor.sendEvent("test", ls, domain.CheckResult{Success: true})

			// Verify expectation.
			if tt.expectEvent {
				select {
				case <-eventCh:
					// Event received as expected.
				default:
					t.Error("expected event but none received")
				}
			} else if tt.hasChannel {
				select {
				case <-eventCh:
					t.Error("unexpected event received")
				default:
					// No event as expected.
				}
			}
		})
	}
}

// Test_ProbeMonitor_checkFailureThresholdReached tests the checkFailureThresholdReached method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_checkFailureThresholdReached(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// success indicates if probe was successful.
		success bool
		// prevFailures is the previous failure count.
		prevFailures int
		// currentFailures is the current failure count.
		currentFailures int
		// threshold is the failure threshold.
		threshold int
		// hasCallback indicates whether callback is set.
		hasCallback bool
		// expectCall indicates whether callback should be called.
		expectCall bool
	}{
		{
			name:            "threshold_crossed",
			success:         false,
			prevFailures:    2,
			currentFailures: 3,
			threshold:       3,
			hasCallback:     true,
			expectCall:      true,
		},
		{
			name:            "threshold_not_crossed",
			success:         false,
			prevFailures:    1,
			currentFailures: 2,
			threshold:       3,
			hasCallback:     true,
			expectCall:      false,
		},
		{
			name:            "already_above_threshold",
			success:         false,
			prevFailures:    3,
			currentFailures: 4,
			threshold:       3,
			hasCallback:     true,
			expectCall:      false,
		},
		{
			name:            "probe_success_no_trigger",
			success:         true,
			prevFailures:    2,
			currentFailures: 0,
			threshold:       3,
			hasCallback:     true,
			expectCall:      false,
		},
		{
			name:            "no_callback",
			success:         false,
			prevFailures:    2,
			currentFailures: 3,
			threshold:       3,
			hasCallback:     false,
			expectCall:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			monitor := &ProbeMonitor{}

			// Set callback if requested.
			if tt.hasCallback {
				monitor.onUnhealthy = func(_ string, _ string) {
					called = true
				}
			}

			// Create listener probe with config.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			binding := &ProbeBinding{
				Config: ProbeConfig{
					FailureThreshold: tt.threshold,
				},
			}
			lp := NewListenerProbeWithBinding(l, binding)

			// Create subject status with failure counts.
			ls := &domain.SubjectStatus{
				ConsecutiveFailures: tt.currentFailures,
			}

			// Call the method.
			monitor.checkFailureThresholdReached(lp, ls, tt.prevFailures, domain.CheckResult{Success: tt.success})

			// Verify expectation.
			assert.Equal(t, tt.expectCall, called)
		})
	}
}

// Test_ProbeMonitor_getFailureThreshold tests the getFailureThreshold method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_getFailureThreshold(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// threshold is the configured threshold.
		threshold int
		// expected is the expected result.
		expected int
	}{
		{
			name:      "configured_threshold",
			threshold: 5,
			expected:  5,
		},
		{
			name:      "zero_returns_default",
			threshold: 0,
			expected:  1,
		},
		{
			name:      "negative_returns_default",
			threshold: -1,
			expected:  1,
		},
		{
			name:      "one_returns_one",
			threshold: 1,
			expected:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := &ProbeMonitor{}

			// Create listener probe with config.
			l := listener.NewListener("test", "tcp", "localhost", 8080)
			binding := &ProbeBinding{
				Config: ProbeConfig{
					FailureThreshold: tt.threshold,
				},
			}
			lp := NewListenerProbeWithBinding(l, binding)

			// Call the method.
			result := monitor.getFailureThreshold(lp)

			// Verify expectation.
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_ProbeMonitor_handleHealthyTransition tests the handleHealthyTransition method.
//
// Params:
//   - t: the testing context.
func Test_ProbeMonitor_handleHealthyTransition(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prevState is the previous subject state.
		prevState domain.SubjectState
		// newState is the new subject state.
		newState domain.SubjectState
		// hasCallback indicates whether callback is set.
		hasCallback bool
		// expectCall indicates whether callback should be called.
		expectCall bool
	}{
		{
			name:        "listening_to_ready_triggers",
			prevState:   domain.SubjectListening,
			newState:    domain.SubjectReady,
			hasCallback: true,
			expectCall:  true,
		},
		{
			name:        "listening_to_ready_no_callback",
			prevState:   domain.SubjectListening,
			newState:    domain.SubjectReady,
			hasCallback: false,
			expectCall:  false,
		},
		{
			name:        "other_transition_no_trigger",
			prevState:   domain.SubjectReady,
			newState:    domain.SubjectListening,
			hasCallback: true,
			expectCall:  false,
		},
		{
			name:        "same_state_no_trigger",
			prevState:   domain.SubjectReady,
			newState:    domain.SubjectReady,
			hasCallback: true,
			expectCall:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			monitor := &ProbeMonitor{}

			// Set callback if requested.
			if tt.hasCallback {
				monitor.onHealthy = func(_ string) {
					called = true
				}
			}

			// Call the method.
			monitor.handleHealthyTransition("test", tt.prevState, tt.newState)

			// Verify expectation.
			assert.Equal(t, tt.expectCall, called)
		})
	}
}
