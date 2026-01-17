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
