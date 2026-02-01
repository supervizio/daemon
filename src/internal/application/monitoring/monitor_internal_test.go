package monitoring

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProber is a mock prober for internal testing.
type mockProber struct {
	probeType  string
	result     health.CheckResult
	probeCount atomic.Int64
}

// Probe returns the configured test result and increments probe count.
func (p *mockProber) Probe(_ context.Context, _ health.Target) health.CheckResult {
	p.probeCount.Add(1)
	return p.result
}

// Type returns the prober type.
func (p *mockProber) Type() string {
	return p.probeType
}

// mockCreator is a mock factory for internal testing.
type mockCreator struct {
	probers map[string]*mockProber
	err     error
}

// Create returns a mock prober for the given type.
func (f *mockCreator) Create(proberType string, _ time.Duration) (health.Prober, error) {
	if f.err != nil {
		return nil, f.err
	}
	if p, ok := f.probers[proberType]; ok {
		return p, nil
	}
	return &mockProber{
		probeType: proberType,
		result:    health.CheckResult{Success: true},
	}, nil
}

// mockDiscoverer is a mock discoverer for testing.
type mockDiscoverer struct {
	targets    []target.ExternalTarget
	err        error
	targetType target.Type
}

// Discover returns pre-configured targets.
func (d *mockDiscoverer) Discover(_ context.Context) ([]target.ExternalTarget, error) {
	return d.targets, d.err
}

// Type returns the discoverer type.
func (d *mockDiscoverer) Type() target.Type {
	if d.targetType != "" {
		return d.targetType
	}
	return target.TypeDocker
}

// mockWatcher is a mock watcher for testing.
type mockWatcher struct {
	events     chan target.Event
	err        error
	targetType target.Type
}

// Watch returns the pre-configured events channel.
func (w *mockWatcher) Watch(_ context.Context) (<-chan target.Event, error) {
	if w.err != nil {
		return nil, w.err
	}
	return w.events, nil
}

// Type returns the watcher type.
func (w *mockWatcher) Type() target.Type {
	if w.targetType != "" {
		return w.targetType
	}
	return target.TypeDocker
}

// TestExternalMonitor_internal tests internal state of ExternalMonitor.
func TestExternalMonitor_internal(t *testing.T) {
	type testCase struct {
		name       string
		setupFunc  func() *ExternalMonitor
		verifyFunc func(*testing.T, *ExternalMonitor)
	}

	tests := []testCase{
		{
			name: "internal state is initialized correctly",
			setupFunc: func() *ExternalMonitor {
				config := NewConfig()
				return NewExternalMonitor(config)
			},
			verifyFunc: func(t *testing.T, monitor *ExternalMonitor) {
				assert.NotNil(t, monitor.registry)
				assert.NotNil(t, monitor.probers)
				assert.NotNil(t, monitor.stopCh)
				assert.False(t, monitor.running)
				assert.Equal(t, 0, len(monitor.probers))
			},
		},
		{
			name: "config is stored correctly",
			setupFunc: func() *ExternalMonitor {
				config := NewConfig().WithFactory(&mockCreator{})
				return NewExternalMonitor(config)
			},
			verifyFunc: func(t *testing.T, monitor *ExternalMonitor) {
				assert.NotNil(t, monitor.config.Factory)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			monitor := tc.setupFunc()
			tc.verifyFunc(t, monitor)
		})
	}
}

// TestExternalMonitor_createProber tests the createProber method.
func TestExternalMonitor_createProber(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		target      *target.ExternalTarget
		expectError bool
		expectedErr error
	}{
		{
			name: "creates_prober_successfully",
			config: NewConfig().WithFactory(&mockCreator{
				probers: map[string]*mockProber{
					"tcp": {probeType: "tcp", result: health.CheckResult{Success: true}},
				},
			}),
			target:      target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			expectError: false,
		},
		{
			name:        "error_when_factory_missing",
			config:      NewConfig(),
			target:      target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			expectError: true,
			expectedErr: ErrProberFactoryMissing,
		},
		{
			name:   "error_when_probe_type_empty",
			config: NewConfig().WithFactory(&mockCreator{}),
			target: &target.ExternalTarget{
				ID:          "test-1",
				ProbeType:   "",
				ProbeTarget: health.Target{Address: "localhost:8080"},
			},
			expectError: true,
			expectedErr: ErrEmptyProbeType,
		},
		{
			name: "error_when_factory_fails",
			config: NewConfig().WithFactory(&mockCreator{
				err: errors.New("factory error"),
			}),
			target:      target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor := NewExternalMonitor(tt.config)

			prober, err := monitor.createProber(tt.target)

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

// TestExternalMonitor_startProber tests the startProber method.
func TestExternalMonitor_startProber(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "spawns_prober_goroutine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &mockCreator{}
			config := NewConfig().WithFactory(factory)
			monitor := NewExternalMonitor(config)

			// Set up monitor as running.
			monitor.running = true
			monitor.stopCh = make(chan struct{})

			// Create target and prober.
			tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
			prober := &mockProber{
				probeType: "tcp",
				result:    health.CheckResult{Success: true},
			}

			// GOROUTINE-LIFECYCLE: startProber spawns a prober goroutine via wg.Go.
			// The goroutine runs until stopCh is closed or context is cancelled.
			// Clean exit verified by wg.Wait after closing stopCh.
			monitor.startProber(tgt, prober)

			// Wait briefly for goroutine to start.
			time.Sleep(20 * time.Millisecond)

			// Stop monitor.
			close(monitor.stopCh)
			monitor.wg.Wait()

			// Verify prober was called.
			assert.GreaterOrEqual(t, prober.probeCount.Load(), int64(1))
		})
	}
}

// TestExternalMonitor_runProber tests the runProber method.
// GOROUTINE-LIFECYCLE: Spawns a test goroutine to verify runProber behavior.
// Test goroutine exits after stopCh is closed, verified via done channel.
func TestExternalMonitor_runProber(t *testing.T) {
	tests := []struct {
		name        string
		interval    time.Duration
		cancelAfter time.Duration
		minProbes   int
	}{
		{
			name:        "performs_initial_probe",
			interval:    100 * time.Millisecond,
			cancelAfter: 50 * time.Millisecond,
			minProbes:   1,
		},
		{
			name:        "performs_periodic_probes",
			interval:    20 * time.Millisecond,
			cancelAfter: 60 * time.Millisecond,
			minProbes:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prober := &mockProber{
				probeType: "tcp",
				result:    health.CheckResult{Success: true},
			}

			config := NewConfig().WithFactory(&mockCreator{})
			config.Defaults.Interval = tt.interval
			monitor := NewExternalMonitor(config)

			// Add target to registry.
			tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
			tgt.Interval = tt.interval
			err := monitor.registry.Add(tgt)
			require.NoError(t, err)

			// Create context with cancel.
			ctx := t.Context()

			// Create stop channel.
			stopCh := make(chan struct{})

			done := make(chan struct{})
			go func() {
				monitor.runProber(ctx, stopCh, tgt, prober)
				close(done)
			}()

			// Wait and cancel.
			time.Sleep(tt.cancelAfter)
			close(stopCh)

			// Wait for goroutine to finish.
			select {
			case <-done:
				// Success.
			case <-time.After(200 * time.Millisecond):
				t.Fatal("runProber did not exit")
			}

			// Verify minimum probes.
			assert.GreaterOrEqual(t, prober.probeCount.Load(), int64(tt.minProbes))
		})
	}
}

// TestExternalMonitor_runProber_stopBeforeInitialProbe tests early exit on stop signal.
func TestExternalMonitor_runProber_stopBeforeInitialProbe(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "exits_when_stop_signal_received_immediately",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prober := &mockProber{
				probeType: "tcp",
				result:    health.CheckResult{Success: true},
			}

			config := NewConfig().WithFactory(&mockCreator{})
			monitor := NewExternalMonitor(config)

			tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")

			// Create already-closed stop channel.
			stopCh := make(chan struct{})
			close(stopCh)

			// Run prober - should exit immediately.
			ctx := context.Background()
			monitor.runProber(ctx, stopCh, tgt, prober)

			// Verify prober was never called.
			assert.Equal(t, int64(0), prober.probeCount.Load())
		})
	}
}

// TestExternalMonitor_runProber_ctxCancelBeforeInitialProbe tests early exit on context cancel.
func TestExternalMonitor_runProber_ctxCancelBeforeInitialProbe(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "exits_when_ctx_cancelled_immediately",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prober := &mockProber{
				probeType: "tcp",
				result:    health.CheckResult{Success: true},
			}

			config := NewConfig().WithFactory(&mockCreator{})
			monitor := NewExternalMonitor(config)

			tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")

			// Create already-cancelled context.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			stopCh := make(chan struct{})

			// Run prober - should exit immediately.
			monitor.runProber(ctx, stopCh, tgt, prober)

			// Verify prober was never called.
			assert.Equal(t, int64(0), prober.probeCount.Load())
		})
	}
}

// TestExternalMonitor_runProber_targetRemoved tests exit when target is removed.
// GOROUTINE-LIFECYCLE: Spawns a test goroutine to verify prober auto-cleanup.
// Test goroutine exits when target is removed from registry, verified via done channel.
func TestExternalMonitor_runProber_targetRemoved(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "exits_when_target_removed_from_registry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prober := &mockProber{
				probeType: "tcp",
				result:    health.CheckResult{Success: true},
			}

			config := NewConfig().WithFactory(&mockCreator{})
			config.Defaults.Interval = 20 * time.Millisecond
			monitor := NewExternalMonitor(config)

			// Add target to registry.
			tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
			tgt.Interval = 20 * time.Millisecond
			err := monitor.registry.Add(tgt)
			require.NoError(t, err)

			ctx := context.Background()
			stopCh := make(chan struct{})

			done := make(chan struct{})
			go func() {
				monitor.runProber(ctx, stopCh, tgt, prober)
				close(done)
			}()

			// Wait for initial probe.
			time.Sleep(30 * time.Millisecond)

			// Remove target from registry.
			err = monitor.registry.Remove(tgt.ID)
			require.NoError(t, err)

			// Wait for goroutine to exit.
			select {
			case <-done:
				// Success - prober exited after target removed.
			case <-time.After(100 * time.Millisecond):
				t.Fatal("runProber did not exit after target removed")
			}

			// Clean up.
			close(stopCh)
		})
	}
}

// TestExternalMonitor_performProbe tests the performProbe method.
func TestExternalMonitor_performProbe(t *testing.T) {
	tests := []struct {
		name    string
		success bool
	}{
		{
			name:    "successful_probe",
			success: true,
		},
		{
			name:    "failed_probe",
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prober := &mockProber{
				probeType: "tcp",
				result:    health.CheckResult{Success: tt.success},
			}

			config := NewConfig().WithFactory(&mockCreator{})
			monitor := NewExternalMonitor(config)

			// Add target to registry.
			tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
			err := monitor.registry.Add(tgt)
			require.NoError(t, err)

			// Perform probe.
			ctx := context.Background()
			monitor.performProbe(ctx, tgt, prober)

			// Verify probe was called.
			assert.Equal(t, int64(1), prober.probeCount.Load())

			// Verify status was updated.
			status := monitor.registry.GetStatus(tgt.ID)
			require.NotNil(t, status)
		})
	}
}

// TestExternalMonitor_updateProbeResult tests the updateProbeResult method.
func TestExternalMonitor_updateProbeResult(t *testing.T) {
	tests := []struct {
		name             string
		result           health.CheckResult
		successThreshold int
		failureThreshold int
	}{
		{
			name:             "updates_status_on_success",
			result:           health.CheckResult{Success: true},
			successThreshold: 1,
			failureThreshold: 3,
		},
		{
			name:             "updates_status_on_failure",
			result:           health.CheckResult{Success: false},
			successThreshold: 1,
			failureThreshold: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig()
			config.Defaults.SuccessThreshold = tt.successThreshold
			config.Defaults.FailureThreshold = tt.failureThreshold
			monitor := NewExternalMonitor(config)

			// Add target to registry.
			tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
			err := monitor.registry.Add(tgt)
			require.NoError(t, err)

			// Update probe result.
			monitor.updateProbeResult(tgt, tt.result)

			// Verify status was updated.
			status := monitor.registry.GetStatus(tgt.ID)
			require.NotNil(t, status)
		})
	}
}

// TestExternalMonitor_notifyStateChange tests the notifyStateChange method.
func TestExternalMonitor_notifyStateChange(t *testing.T) {
	tests := []struct {
		name            string
		hasOnChange     bool
		hasOnUnhealthy  bool
		hasOnHealthy    bool
		prevState       target.State
		newState        target.State
		expectChange    bool
		expectUnhealthy bool
		expectHealthy   bool
	}{
		{
			name:         "calls_on_health_change",
			hasOnChange:  true,
			prevState:    target.StateUnknown,
			newState:     target.StateHealthy,
			expectChange: true,
		},
		{
			name:            "calls_on_unhealthy",
			hasOnUnhealthy:  true,
			prevState:       target.StateHealthy,
			newState:        target.StateUnhealthy,
			expectUnhealthy: true,
		},
		{
			name:          "calls_on_healthy_from_unhealthy",
			hasOnHealthy:  true,
			prevState:     target.StateUnhealthy,
			newState:      target.StateHealthy,
			expectHealthy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var changeCallCount atomic.Int64
			var unhealthyCallCount atomic.Int64
			var healthyCallCount atomic.Int64

			config := NewConfig()
			if tt.hasOnChange {
				config.OnHealthChange = func(string, string, string) {
					changeCallCount.Add(1)
				}
			}
			if tt.hasOnUnhealthy {
				config.OnUnhealthy = func(string, string) {
					unhealthyCallCount.Add(1)
				}
			}
			if tt.hasOnHealthy {
				config.OnHealthy = func(string) {
					healthyCallCount.Add(1)
				}
			}
			monitor := NewExternalMonitor(config)

			tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
			result := health.CheckResult{Success: tt.newState == target.StateHealthy}

			// Call method.
			monitor.notifyStateChange(tgt, tt.prevState, tt.newState, result)

			// Verify callbacks.
			if tt.expectChange {
				assert.Equal(t, int64(1), changeCallCount.Load())
			} else {
				assert.Equal(t, int64(0), changeCallCount.Load())
			}
			if tt.expectUnhealthy {
				assert.Equal(t, int64(1), unhealthyCallCount.Load())
			} else {
				assert.Equal(t, int64(0), unhealthyCallCount.Load())
			}
			if tt.expectHealthy {
				assert.Equal(t, int64(1), healthyCallCount.Load())
			} else {
				assert.Equal(t, int64(0), healthyCallCount.Load())
			}
		})
	}
}

// TestExternalMonitor_sendEvent tests the sendEvent method.
func TestExternalMonitor_sendEvent(t *testing.T) {
	tests := []struct {
		name        string
		hasChannel  bool
		expectEvent bool
	}{
		{
			name:        "sends_event_when_channel_exists",
			hasChannel:  true,
			expectEvent: true,
		},
		{
			name:        "skips_event_when_no_channel",
			hasChannel:  false,
			expectEvent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var eventCh chan target.Event
			if tt.hasChannel {
				eventCh = make(chan target.Event, 1)
			}

			config := NewConfig().WithEvents(eventCh)
			monitor := NewExternalMonitor(config)

			tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
			event := target.NewAddedEvent(tgt)

			// Send event.
			monitor.sendEvent(event)

			// Verify.
			if tt.expectEvent {
				select {
				case <-eventCh:
					// Event received as expected.
				case <-time.After(100 * time.Millisecond):
					t.Fatal("expected event not received")
				}
			} else if eventCh != nil {
				select {
				case <-eventCh:
					t.Fatal("unexpected event received")
				default:
					// No event as expected.
				}
			}
		})
	}
}

// TestExternalMonitor_sendEvent_fullChannel tests non-blocking behavior with full channel.
// GOROUTINE-LIFECYCLE: Spawns a test goroutine to verify sendEvent non-blocking behavior.
// Test goroutine completes immediately as sendEvent uses non-blocking send (select/default).
func TestExternalMonitor_sendEvent_fullChannel(t *testing.T) {
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
			eventCh := make(chan target.Event, 1)

			// Pre-fill the channel.
			tgt := target.NewRemoteTarget("prefill", "localhost:8080", "tcp")
			eventCh <- target.NewAddedEvent(tgt)

			config := NewConfig().WithEvents(eventCh)
			monitor := NewExternalMonitor(config)

			tgt2 := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
			event := target.NewAddedEvent(tgt2)

			done := make(chan struct{})
			go func() {
				monitor.sendEvent(event)
				close(done)
			}()

			// Wait with timeout.
			select {
			case <-done:
				// Success - method completed without blocking.
			case <-time.After(100 * time.Millisecond):
				t.Fatal("sendEvent blocked on full channel")
			}
		})
	}
}

// TestExternalMonitor_handleWatcherEvent tests the handleWatcherEvent method.
func TestExternalMonitor_handleWatcherEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType target.EventType
	}{
		{
			name:      "handles_added_event",
			eventType: target.EventAdded,
		},
		{
			name:      "handles_removed_event",
			eventType: target.EventRemoved,
		},
		{
			name:      "handles_updated_event",
			eventType: target.EventUpdated,
		},
		{
			name:      "handles_health_changed_event",
			eventType: target.EventHealthChanged,
		},
		{
			name:      "handles_unknown_event_type",
			eventType: target.EventType("unknown"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventCh := make(chan target.Event, 1)
			config := NewConfig().WithFactory(&mockCreator{}).WithEvents(eventCh)
			monitor := NewExternalMonitor(config)

			tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")

			// Pre-add target for removed event test.
			if tt.eventType == target.EventRemoved {
				err := monitor.AddTarget(tgt)
				require.NoError(t, err)
			}

			event := target.Event{
				Type:   tt.eventType,
				Target: *tgt,
			}

			// Handle event - should not panic.
			assert.NotPanics(t, func() {
				monitor.handleWatcherEvent(event)
			})
		})
	}
}

// TestExternalMonitor_discover tests the discover method.
func TestExternalMonitor_discover(t *testing.T) {
	tests := []struct {
		name              string
		discovererTargets []target.ExternalTarget
		discovererErr     error
	}{
		{
			name: "discovers_and_adds_new_targets",
			discovererTargets: []target.ExternalTarget{
				*target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
				*target.NewRemoteTarget("test-2", "localhost:8081", "tcp"),
			},
		},
		{
			name:          "handles_discoverer_error",
			discovererErr: errors.New("discovery error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discoverer := &mockDiscoverer{
				targets: tt.discovererTargets,
				err:     tt.discovererErr,
			}

			config := NewConfig().WithFactory(&mockCreator{}).WithDiscoverers(discoverer)
			monitor := NewExternalMonitor(config)

			ctx := context.Background()

			// Discover - should not panic.
			assert.NotPanics(t, func() {
				monitor.discover(ctx)
			})
		})
	}
}

// TestExternalMonitor_runWatcher tests the runWatcher method.
// GOROUTINE-LIFECYCLE: Spawns a test goroutine to verify runWatcher event processing.
// Test goroutine exits when stopCh is closed or ctx is cancelled, verified via done channel.
func TestExternalMonitor_runWatcher(t *testing.T) {
	tests := []struct {
		name       string
		watcherErr error
	}{
		{
			name: "processes_watcher_events",
		},
		{
			name:       "handles_watcher_error",
			watcherErr: errors.New("watcher error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventsCh := make(chan target.Event, 1)
			watcher := &mockWatcher{
				events: eventsCh,
				err:    tt.watcherErr,
			}

			config := NewConfig().WithFactory(&mockCreator{})
			monitor := NewExternalMonitor(config)

			ctx := t.Context()
			stopCh := make(chan struct{})

			done := make(chan struct{})
			go func() {
				monitor.runWatcher(ctx, stopCh, watcher)
				close(done)
			}()

			// If no error, send an event and verify handling.
			if tt.watcherErr == nil {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				eventsCh <- target.NewAddedEvent(tgt)

				// Wait briefly for event processing.
				time.Sleep(20 * time.Millisecond)
			}

			// Stop watcher.
			close(stopCh)

			// Wait for goroutine to finish.
			select {
			case <-done:
				// Success.
			case <-time.After(200 * time.Millisecond):
				t.Fatal("runWatcher did not exit")
			}
		})
	}
}

// TestExternalMonitor_runDiscovery tests the runDiscovery method.
// GOROUTINE-LIFECYCLE: Spawns a test goroutine to verify runDiscovery periodic execution.
// Test goroutine exits when stopCh is closed or ctx is cancelled, verified via done channel.
func TestExternalMonitor_runDiscovery(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "performs_periodic_discovery",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discoverer := &mockDiscoverer{
				targets: []target.ExternalTarget{
					*target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
				},
			}

			config := NewConfig().WithFactory(&mockCreator{}).WithDiscoverers(discoverer)
			config.Discovery.Interval = 50 * time.Millisecond
			monitor := NewExternalMonitor(config)

			ctx := t.Context()
			stopCh := make(chan struct{})

			done := make(chan struct{})
			go func() {
				monitor.runDiscovery(ctx, stopCh)
				close(done)
			}()

			// Wait for initial discovery and one interval.
			time.Sleep(100 * time.Millisecond)

			// Stop discovery.
			close(stopCh)

			// Wait for goroutine to finish.
			select {
			case <-done:
				// Success.
			case <-time.After(200 * time.Millisecond):
				t.Fatal("runDiscovery did not exit")
			}
		})
	}
}

// TestExternalMonitor_startExistingProbers tests the startExistingProbers method.
func TestExternalMonitor_startExistingProbers(t *testing.T) {
	tests := []struct {
		name             string
		setupTargets     []*target.ExternalTarget
		expectProberCall bool
	}{
		{
			name: "starts_probers_for_existing_targets",
			setupTargets: []*target.ExternalTarget{
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			},
			expectProberCall: true,
		},
		{
			name:             "handles_empty_target_list",
			setupTargets:     []*target.ExternalTarget{},
			expectProberCall: false,
		},
		{
			name: "starts_multiple_probers",
			setupTargets: []*target.ExternalTarget{
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
				target.NewRemoteTarget("test-2", "localhost:9090", "tcp"),
			},
			expectProberCall: true,
		},
		{
			name: "skips_targets_without_probers",
			setupTargets: []*target.ExternalTarget{
				func() *target.ExternalTarget {
					tgt := target.NewRemoteTarget("test-no-probe", "localhost:8080", "")
					tgt.ProbeType = ""
					return tgt
				}(),
			},
			expectProberCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prober := &mockProber{
				probeType: "tcp",
				result:    health.CheckResult{Success: true},
			}

			config := NewConfig().WithFactory(&mockCreator{})
			monitor := NewExternalMonitor(config)

			// Add targets and probers.
			for _, tgt := range tt.setupTargets {
				err := monitor.registry.Add(tgt)
				require.NoError(t, err)
				if tgt.ProbeType != "" {
					monitor.probers[tgt.ID] = prober
				}
			}

			// Set monitor as running.
			monitor.running = true
			monitor.stopCh = make(chan struct{})

			// GOROUTINE-LIFECYCLE: startExistingProbers spawns prober goroutines for existing targets.
			// Each prober runs until stopCh is closed. Clean shutdown verified by wg.Wait.
			monitor.startExistingProbers(tt.setupTargets)

			// Wait briefly for goroutine to start.
			time.Sleep(20 * time.Millisecond)

			// Stop monitor.
			close(monitor.stopCh)
			monitor.wg.Wait()

			// Verify prober was called if expected.
			if tt.expectProberCall {
				assert.GreaterOrEqual(t, prober.probeCount.Load(), int64(1))
			} else {
				assert.Equal(t, int64(0), prober.probeCount.Load())
			}
		})
	}
}

// TestExternalMonitor_startDiscoveryWorkers tests the startDiscoveryWorkers method.
func TestExternalMonitor_startDiscoveryWorkers(t *testing.T) {
	tests := []struct {
		name            string
		enableDiscovery bool
		hasWatchers     bool
	}{
		{
			name:            "starts_discovery_when_enabled",
			enableDiscovery: true,
		},
		{
			name:        "starts_watchers",
			hasWatchers: true,
		},
		{
			name:            "starts_both_discovery_and_watchers",
			enableDiscovery: true,
			hasWatchers:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig().WithFactory(&mockCreator{})

			if tt.enableDiscovery {
				discoverer := &mockDiscoverer{}
				config = config.WithDiscoverers(discoverer)
			}

			if tt.hasWatchers {
				watcher := &mockWatcher{events: make(chan target.Event)}
				config = config.WithWatchers(watcher)
			}

			monitor := NewExternalMonitor(config)

			ctx := t.Context()

			stopCh := make(chan struct{})

			// GOROUTINE-LIFECYCLE: startDiscoveryWorkers spawns discovery and watcher goroutines.
			// All goroutines exit when stopCh is closed, verified by wg.Wait.
			monitor.startDiscoveryWorkers(ctx, stopCh)

			// Wait briefly for workers to start.
			time.Sleep(50 * time.Millisecond)

			// Stop workers.
			close(stopCh)
			monitor.wg.Wait()
		})
	}
}
