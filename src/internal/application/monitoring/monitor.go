// Package monitoring provides the application service for external target monitoring.
package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
)

// defaultProberMapCapacity is the initial capacity for the probers map.
const defaultProberMapCapacity int = 16

// ExternalMonitor monitors external targets (services, containers, hosts).
// Unlike the health.ProbeMonitor which monitors managed service listeners,
// ExternalMonitor observes external resources that supervizio does not manage.
type ExternalMonitor struct {
	// mu protects concurrent access to monitor state.
	mu sync.RWMutex

	// config contains the monitor configuration.
	config Config

	// registry stores targets and their status.
	registry *Registry

	// probers maps target ID to its prober instance.
	probers map[string]health.Prober

	// stopCh signals goroutines to stop.
	stopCh chan struct{}

	// running indicates if the monitor is active.
	running bool

	// wg tracks prober goroutines for clean shutdown.
	wg sync.WaitGroup
}

// NewExternalMonitor creates a new external target monitor.
//
// Params:
//   - config: the monitor configuration.
//
// Returns:
//   - *ExternalMonitor: the new monitor instance.
func NewExternalMonitor(config Config) *ExternalMonitor {
	// construct monitor with configuration
	return &ExternalMonitor{
		config:   config,
		registry: NewRegistry(),
		probers:  make(map[string]health.Prober, defaultProberMapCapacity),
		stopCh:   make(chan struct{}),
	}
}

// AddTarget adds a target to monitor.
// If the monitor is running, starts probing immediately.
//
// Params:
//   - t: the target to add.
//
// Returns:
//   - error: if target already exists or prober creation fails.
func (m *ExternalMonitor) AddTarget(t *target.ExternalTarget) error {
	// Lock for thread-safe update.
	m.mu.Lock()
	defer m.mu.Unlock()

	// check if target can be added to registry
	if err := m.registry.Add(t); err != nil {
		// return error if target already exists
		return err
	}

	// check if target has a probe configured
	if t.HasProbe() {
		prober, err := m.createProber(t)
		// check for prober creation error
		if err != nil {
			// Remove target if prober creation fails.
			_ = m.registry.Remove(t.ID)

			// return prober creation error
			return err
		}
		m.probers[t.ID] = prober

		// Start probing if monitor is running.
		if m.running {
			m.startProber(t, prober)
		}
	}

	// return success
	return nil
}

// RemoveTarget removes a target from monitoring.
// The prober goroutine will stop on next probe cycle.
//
// Params:
//   - id: the target ID to remove.
//
// Returns:
//   - error: if target not found.
func (m *ExternalMonitor) RemoveTarget(id string) error {
	// Lock for thread-safe update.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove prober for this target.
	delete(m.probers, id)

	// return registry removal result
	return m.registry.Remove(id)
}

// AddTargets adds multiple targets at once.
//
// Params:
//   - targets: the targets to add.
//
// Returns:
//   - error: if any target fails to add.
func (m *ExternalMonitor) AddTargets(targets []*target.ExternalTarget) error {
	// iterate over all targets to add
	for _, t := range targets {
		// check if target addition fails
		if err := m.AddTarget(t); err != nil {
			// return error with target context
			return fmt.Errorf("add target %s: %w", t.ID, err)
		}
	}

	// return success
	return nil
}

// Start starts the external monitor.
// This method spawns goroutines for each target with a prober configured.
//
// Params:
//   - ctx: context for cancellation.
func (m *ExternalMonitor) Start(ctx context.Context) {
	// Lock to check and update running state.
	m.mu.Lock()

	// check if already running
	if m.running {
		m.mu.Unlock()

		// Return early to avoid starting duplicate probers.
		return
	}

	// Mark as running and create new stop channel.
	m.running = true
	stopCh := make(chan struct{})
	m.stopCh = stopCh

	// Snapshot targets to start probing.
	targets := m.registry.All()
	m.mu.Unlock()

	// Start probers for existing targets.
	m.startExistingProbers(targets)

	// Start discovery and watchers.
	m.startDiscoveryWorkers(ctx, stopCh)
}

// startExistingProbers starts prober goroutines for all existing targets.
//
// Params:
//   - targets: the targets to start probing.
func (m *ExternalMonitor) startExistingProbers(targets []*target.ExternalTarget) {
	// iterate over all targets
	for _, t := range targets {
		m.mu.RLock()
		prober, exists := m.probers[t.ID]
		m.mu.RUnlock()

		// check if prober exists for target
		if exists && prober != nil {
			m.startProber(t, prober)
		}
	}
}

// startDiscoveryWorkers starts discovery and watcher goroutines.
//
// Params:
//   - ctx: context for cancellation.
//   - stopCh: channel to signal stop.
func (m *ExternalMonitor) startDiscoveryWorkers(ctx context.Context, stopCh chan struct{}) {
	// Start discovery if enabled.
	if m.config.Discovery.Enabled {
		// spawn discovery worker goroutine
		m.wg.Go(func() {
			m.runDiscovery(ctx, stopCh)
		})
	}

	// Start watchers for real-time updates.
	for _, watcher := range m.config.Discovery.Watchers {
		w := watcher

		// spawn watcher goroutine
		m.wg.Go(func() {
			m.runWatcher(ctx, stopCh, w)
		})
	}
}

// Stop stops the external monitor.
func (m *ExternalMonitor) Stop() {
	// Lock to check and update running state.
	m.mu.Lock()

	// check if not running
	if !m.running {
		m.mu.Unlock()

		// Return early since monitor is not running.
		return
	}

	// Mark as not running and signal all goroutines to stop.
	m.running = false
	close(m.stopCh)
	m.mu.Unlock()

	// Wait for all prober goroutines to terminate.
	m.wg.Wait()
}

// createProber creates a prober for a target.
//
// Params:
//   - t: the target needing a prober.
//
// Returns:
//   - health.Prober: the created prober.
//   - error: if factory is missing or creation fails.
func (m *ExternalMonitor) createProber(t *target.ExternalTarget) (health.Prober, error) {
	// check if factory is missing
	if m.config.Factory == nil {
		// return error when factory not set
		return nil, ErrProberFactoryMissing
	}

	// check if probe type is empty
	if t.ProbeType == "" {
		// return error for empty probe type
		return nil, ErrEmptyProbeType
	}

	timeout := m.config.GetTimeout(t.Timeout)

	prober, err := m.config.Factory.Create(t.ProbeType, timeout)
	// check for factory creation failure
	if err != nil {
		// return error with target context
		return nil, fmt.Errorf("create prober for target %q: %w", t.ID, err)
	}

	// return created prober
	return prober, nil
}

// startProber starts a prober goroutine for a target.
//
// Params:
//   - t: the target to probe.
//   - prober: the prober to use.
func (m *ExternalMonitor) startProber(t *target.ExternalTarget, prober health.Prober) {
	m.mu.RLock()
	stopCh := m.stopCh
	m.mu.RUnlock()

	tgt := t
	prb := prober

	// spawn prober goroutine
	m.wg.Go(func() {
		m.runProber(context.Background(), stopCh, tgt, prb)
	})
}

// runProber runs a single prober in a loop.
//
// Params:
//   - ctx: context for cancellation.
//   - stopCh: channel to signal stop.
//   - t: the target to probe.
//   - prober: the prober to use.
func (m *ExternalMonitor) runProber(ctx context.Context, stopCh <-chan struct{}, t *target.ExternalTarget, prober health.Prober) {
	interval := m.config.GetInterval(t.Interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Perform initial probe immediately.
	select {
	case <-stopCh:
		// Stop signal received before initial probe.
		return
	case <-ctx.Done():
		// Context cancelled before initial probe.
		return
	default:
		// Perform initial probe.
		m.performProbe(ctx, t, prober)
	}

	// Loop until stopped.
	for {
		select {
		case <-stopCh:
			// Stop signal received.
			return
		case <-ctx.Done():
			// Context cancelled.
			return
		case <-ticker.C:
			// Check if target still exists before probing.
			if m.registry.Get(t.ID) == nil {
				// Target was removed, exit loop.
				return
			}
			// Perform periodic probe.
			m.performProbe(ctx, t, prober)
		}
	}
}

// performProbe performs a single health probe.
//
// Params:
//   - ctx: parent context.
//   - t: the target to probe.
//   - prober: the prober to use.
func (m *ExternalMonitor) performProbe(ctx context.Context, t *target.ExternalTarget, prober health.Prober) {
	timeout := m.config.GetTimeout(t.Timeout)

	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute the probe.
	result := prober.Probe(probeCtx, t.ProbeTarget)

	// Update status with probe result.
	m.updateProbeResult(t, result)
}

// updateProbeResult updates the target status based on probe result.
//
// Params:
//   - t: the target that was probed.
//   - result: the probe result.
func (m *ExternalMonitor) updateProbeResult(t *target.ExternalTarget, result health.CheckResult) {
	successThreshold := m.config.GetSuccessThreshold(t.SuccessThreshold)
	failureThreshold := m.config.GetFailureThreshold(t.FailureThreshold)

	var previousState target.State

	err := m.registry.UpdateStatus(t.ID, func(status *target.Status) {
		previousState = status.State
		status.UpdateFromProbe(result, successThreshold, failureThreshold)
	})
	// check if status update failed
	if err != nil {
		// Target was removed, skip notification.
		return
	}

	status := m.registry.GetStatus(t.ID)
	// check if status exists
	if status == nil {
		// Target was removed, skip notification.
		return
	}

	// check if state changed
	if status.State != previousState {
		// Notify callbacks about state change.
		m.notifyStateChange(t, previousState, status.State, result)
	}
}

// notifyStateChange handles health state changes.
//
// Params:
//   - t: the target.
//   - previousState: the previous health state.
//   - newState: the new health state.
//   - result: the probe result.
func (m *ExternalMonitor) notifyStateChange(t *target.ExternalTarget, previousState, newState target.State, result health.CheckResult) {
	// Call health change callback if configured.
	if m.config.OnHealthChange != nil {
		m.config.OnHealthChange(t.ID, string(previousState), string(newState))
	}

	// Call unhealthy callback if target became unhealthy.
	if newState == target.StateUnhealthy && m.config.OnUnhealthy != nil {
		reason := result.Output
		// check if error is present
		if result.Error != nil {
			reason = result.Error.Error()
		}
		m.config.OnUnhealthy(t.ID, reason)
	}

	// Call healthy callback if target recovered from unhealthy.
	if newState == target.StateHealthy && previousState == target.StateUnhealthy && m.config.OnHealthy != nil {
		m.config.OnHealthy(t.ID)
	}

	// Send event to channel.
	m.sendEvent(target.NewHealthChangedEvent(t, previousState, newState))
}

// sendEvent sends an event to the events channel.
//
// Params:
//   - event: the event to send.
func (m *ExternalMonitor) sendEvent(event target.Event) {
	// check if event channel is configured
	if m.config.Events == nil {
		// Return early when no event channel.
		return
	}

	// Non-blocking send to avoid deadlocks.
	select {
	case m.config.Events <- event:
		// Event sent successfully.
	default:
		// Channel full, skip event.
	}
}

// runDiscovery runs periodic discovery.
//
// Params:
//   - ctx: context for cancellation.
//   - stopCh: channel to signal stop.
func (m *ExternalMonitor) runDiscovery(ctx context.Context, stopCh <-chan struct{}) {
	// Perform initial discovery.
	m.discover(ctx)

	interval := m.config.Discovery.Interval
	// check if interval needs default
	if interval <= 0 {
		interval = DefaultDiscoveryInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Loop until stopped.
	for {
		select {
		case <-stopCh:
			// Stop signal received.
			return
		case <-ctx.Done():
			// Context cancelled.
			return
		case <-ticker.C:
			// Perform periodic discovery.
			m.discover(ctx)
		}
	}
}

// discover runs all discoverers and updates targets.
//
// Params:
//   - ctx: context for cancellation.
func (m *ExternalMonitor) discover(ctx context.Context) {
	// iterate over all discoverers
	for _, discoverer := range m.config.Discovery.Discoverers {
		targets, err := discoverer.Discover(ctx)
		// check if discovery failed
		if err != nil {
			// Skip this discoverer on error.
			continue
		}

		// process each discovered target
		for i := range targets {
			t := &targets[i]
			existing := m.registry.Get(t.ID)

			// check if target is new
			if existing == nil {
				// New target discovered.
				err := m.AddTarget(t)
				// check if target was added successfully
				if err == nil {
					m.sendEvent(target.NewAddedEvent(t))
				}
			} else {
				// Update existing target.
				m.registry.AddOrUpdate(t)
			}
		}
	}
}

// runWatcher runs a watcher for real-time updates.
//
// Params:
//   - ctx: context for cancellation.
//   - stopCh: channel to signal stop.
//   - watcher: the watcher to run.
func (m *ExternalMonitor) runWatcher(ctx context.Context, stopCh <-chan struct{}, watcher target.Watcher) {
	events, err := watcher.Watch(ctx)
	// check if watch failed
	if err != nil {
		// Return early on watch error.
		return
	}

	// Loop until stopped.
	for {
		select {
		case <-stopCh:
			// Stop signal received.
			return
		case <-ctx.Done():
			// Context cancelled.
			return
		case event, ok := <-events:
			// check if channel is closed
			if !ok {
				// Channel closed, exit loop.
				return
			}
			// Handle the watcher event.
			m.handleWatcherEvent(event)
		}
	}
}

// handleWatcherEvent handles an event from a watcher.
//
// Params:
//   - event: the event to handle.
func (m *ExternalMonitor) handleWatcherEvent(event target.Event) {
	// dispatch based on event type
	switch event.Type {
	// Handle target added event.
	case target.EventAdded:
		t := &event.Target
		// check if target was added successfully
		if err := m.AddTarget(t); err == nil {
			m.sendEvent(event)
		}

	// Handle target removed event.
	case target.EventRemoved:
		// check if target was removed successfully
		if err := m.RemoveTarget(event.Target.ID); err == nil {
			m.sendEvent(event)
		}

	// Handle target updated event.
	case target.EventUpdated:
		t := &event.Target
		m.registry.AddOrUpdate(t)
		m.sendEvent(event)

	// EventHealthChanged is handled internally, not from watchers.
	case target.EventHealthChanged:
		// Health changes are generated internally, forward event as-is.
		m.sendEvent(event)

	// Default case for unknown event types (external enum safety).
	default:
		// Unknown event type, ignore.
	}
}

// Registry returns the target registry for external access.
//
// Returns:
//   - *Registry: the target registry.
func (m *ExternalMonitor) Registry() *Registry {
	// return registry reference
	return m.registry
}

// Health returns the health summary for all targets.
//
// Returns:
//   - HealthSummary: the health summary.
func (m *ExternalMonitor) Health() HealthSummary {
	// return computed health summary
	return m.registry.HealthSummary()
}

// IsRunning returns whether the monitor is running.
//
// Returns:
//   - bool: true if the monitor is running.
func (m *ExternalMonitor) IsRunning() bool {
	// Lock for thread-safe read.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// return running state
	return m.running
}

// TargetCount returns the number of monitored targets.
//
// Returns:
//   - int: the number of targets.
func (m *ExternalMonitor) TargetCount() int {
	// return registry count
	return m.registry.Count()
}

// GetStatus returns the status for a specific target.
//
// Params:
//   - id: the target ID.
//
// Returns:
//   - *target.Status: the status or nil if not found.
func (m *ExternalMonitor) GetStatus(id string) *target.Status {
	// return status from registry
	return m.registry.GetStatus(id)
}

// AllStatuses returns all target statuses.
//
// Returns:
//   - []*target.Status: slice of all statuses.
func (m *ExternalMonitor) AllStatuses() []*target.Status {
	// return all statuses from registry
	return m.registry.AllStatuses()
}
