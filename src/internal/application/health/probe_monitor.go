// Package health provides the application service for health monitoring.
package health

import (
	"context"
	"sync"
	"time"

	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/listener"
	"github.com/kodflow/daemon/internal/domain/probe"
	"github.com/kodflow/daemon/internal/domain/process"
)

// ProbeMonitor monitors health using the Prober interface.
// It supports multiple listeners with different probe types and aggregates health status.
type ProbeMonitor struct {
	// mu protects concurrent access to monitor state.
	mu sync.RWMutex
	// listeners contains all monitored listeners.
	listeners []*ListenerProbe
	// health stores the aggregated health state.
	health *domain.AggregatedHealth
	// processState tracks the current process state.
	processState process.State
	// customStatus stores a custom status string.
	customStatus string
	// events channel for sending health events.
	events chan<- domain.Event
	// stopCh signals goroutines to stop.
	stopCh chan struct{}
	// running indicates if the monitor is active.
	running bool
	// factory creates probers.
	factory Creator
	// defaultTimeout is the default probe timeout.
	defaultTimeout time.Duration
	// defaultInterval is the default probe interval.
	defaultInterval time.Duration
}

// NewProbeMonitor creates a new probe-based health monitor.
//
// Params:
//   - config: the monitor configuration.
//
// Returns:
//   - *ProbeMonitor: the new monitor instance.
func NewProbeMonitor(config ProbeMonitorConfig) *ProbeMonitor {
	defaultTimeout := config.DefaultTimeout
	// Use probe default timeout when not configured.
	if defaultTimeout == 0 {
		defaultTimeout = probe.DefaultTimeout
	}

	defaultInterval := config.DefaultInterval
	// Use probe default interval when not configured.
	if defaultInterval == 0 {
		defaultInterval = probe.DefaultInterval
	}

	// Return configured monitor with initialized state.
	return &ProbeMonitor{
		listeners:       nil,
		health:          domain.NewAggregatedHealth(process.StateStopped),
		processState:    process.StateStopped,
		events:          config.Events,
		stopCh:          make(chan struct{}),
		factory:         config.Factory,
		defaultTimeout:  defaultTimeout,
		defaultInterval: defaultInterval,
	}
}

// AddListener adds a listener to monitor.
//
// Params:
//   - l: the listener to add.
//
// Returns:
//   - error: if prober creation fails.
func (m *ProbeMonitor) AddListener(l *listener.Listener) error {
	// Lock for thread-safe update.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if listener has probe configuration.
	if !l.HasProbe() {
		// Add listener without prober.
		m.listeners = append(m.listeners, &ListenerProbe{
			Listener: l,
		})
		// Return nil since listener was added successfully without probe.
		return nil
	}

	// Check if factory is configured before attempting to create prober.
	if m.factory == nil {
		// Return error when factory is missing but probe is configured.
		return ErrProberFactoryMissing
	}

	// Use effective timeout, falling back to default if not set.
	timeout := l.ProbeConfig.Timeout
	// Check if timeout is not configured.
	if timeout == 0 {
		// Use default timeout from monitor configuration.
		timeout = m.defaultTimeout
	}

	// Create prober for this listener with effective timeout.
	prober, err := m.factory.Create(l.ProbeType, timeout)
	// Return error if prober creation fails.
	if err != nil {
		// Propagate factory error to caller.
		return err
	}

	// Add listener with prober.
	m.listeners = append(m.listeners, &ListenerProbe{
		Listener: l,
		Prober:   prober,
		Config:   *l.ProbeConfig,
		Target:   l.ProbeTarget,
	})

	// Return nil to indicate successful listener addition.
	return nil
}

// SetProcessState updates the process state.
//
// Params:
//   - state: the new process state.
func (m *ProbeMonitor) SetProcessState(state process.State) {
	// Lock for thread-safe update.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update process state in monitor.
	m.processState = state

	// Update process state in aggregated health.
	m.health.ProcessState = state
}

// SetCustomStatus sets a custom status string.
//
// Params:
//   - status: the custom status string.
func (m *ProbeMonitor) SetCustomStatus(status string) {
	// Lock for thread-safe update.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update custom status in monitor.
	m.customStatus = status

	// Update custom status in aggregated health.
	m.health.SetCustomStatus(status)
}

// Start starts the probe monitor.
// This method spawns goroutines for each listener with a prober configured.
// Each goroutine runs a probe loop that terminates when the context is cancelled
// or when Stop() is called (which closes stopCh). Resources are cleaned up via
// deferred ticker.Stop() in each runProber goroutine.
//
// Params:
//   - ctx: context for cancellation.
func (m *ProbeMonitor) Start(ctx context.Context) {
	// Lock to check and update running state.
	m.mu.Lock()

	// Check if already running to prevent duplicate goroutines.
	if m.running {
		m.mu.Unlock()
		// Return early to avoid starting duplicate probers.
		return
	}

	// Mark as running and create new stop channel.
	m.running = true
	stopCh := make(chan struct{})
	m.stopCh = stopCh

	// Snapshot listeners to avoid races with AddListener().
	listeners := append([]*ListenerProbe(nil), m.listeners...)
	m.mu.Unlock()

	// Start a goroutine for each listener with a prober.
	// Pass stopCh as parameter to avoid race conditions on restart.
	for _, lp := range listeners {
		// Only start probers for listeners that have one configured.
		if lp.Prober != nil {
			go m.runProber(ctx, stopCh, lp)
		}
	}
}

// Stop stops the probe monitor.
func (m *ProbeMonitor) Stop() {
	// Lock to check and update running state.
	m.mu.Lock()

	// Check if not running to avoid closing already-closed channel.
	if !m.running {
		m.mu.Unlock()
		// Return early since monitor is not running.
		return
	}

	// Mark as not running and signal all goroutines to stop.
	m.running = false
	close(m.stopCh)
	m.mu.Unlock()
}

// runProber runs a single prober in a loop.
//
// Params:
//   - ctx: context for cancellation.
//   - stopCh: channel to signal stop (passed as param to avoid race on restart).
//   - lp: the listener probe to run.
func (m *ProbeMonitor) runProber(ctx context.Context, stopCh <-chan struct{}, lp *ListenerProbe) {
	interval := lp.Config.Interval
	// Use default interval when not specified in config.
	if interval == 0 {
		interval = m.defaultInterval
	}

	// Create ticker for periodic probes.
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Perform initial probe immediately.
	m.performProbe(ctx, lp)

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
			// Perform periodic probe.
			m.performProbe(ctx, lp)
		}
	}
}

// performProbe performs a single probe.
//
// Params:
//   - ctx: parent context.
//   - lp: the listener probe to use.
func (m *ProbeMonitor) performProbe(ctx context.Context, lp *ListenerProbe) {
	timeout := lp.Config.Timeout
	// Use default timeout when not specified in config.
	if timeout == 0 {
		timeout = m.defaultTimeout
	}

	// Create timeout context for this probe.
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	target := lp.Target
	// Use listener address when target address is not specified.
	if target.Address == "" {
		target.Address = lp.Listener.ProbeAddress()
	}

	// Execute the probe.
	result := lp.Prober.Probe(probeCtx, target)

	// Update state with probe result.
	m.updateProbeResult(lp, result)
}

// updateProbeResult updates the listener status based on probe result.
//
// Params:
//   - lp: the listener probe that was executed.
//   - result: the probe result.
func (m *ProbeMonitor) updateProbeResult(lp *ListenerProbe, result probe.Result) {
	// Lock for thread-safe state updates.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find or create listener status in health.
	ls := m.findOrCreateListenerStatus(lp)

	// Get previous state for event comparison.
	prevState := lp.Listener.State

	// Get normalized thresholds (zero values become 1).
	successThreshold, failureThreshold := m.normalizeThresholds(lp.Config)

	// Update state based on probe result.
	m.updateListenerState(lp, ls, result, successThreshold, failureThreshold)

	// Store the probe result and update latency.
	m.storeProbeResult(ls, result)

	// Send event if state changed.
	m.sendEventIfChanged(lp, ls, prevState, result)
}

// normalizeThresholds returns thresholds with zero values normalized to 1.
//
// Params:
//   - config: the probe configuration.
//
// Returns:
//   - int: normalized success threshold (minimum 1).
//   - int: normalized failure threshold (minimum 1).
func (m *ProbeMonitor) normalizeThresholds(config probe.Config) (successThreshold, failureThreshold int) {
	// Default to 1 for zero or negative success threshold.
	successThreshold = config.SuccessThreshold
	// Check if success threshold needs normalization.
	if successThreshold <= 0 {
		// Use minimum of 1 for unset threshold.
		successThreshold = 1
	}

	// Default to 1 for zero or negative failure threshold.
	failureThreshold = config.FailureThreshold
	// Check if failure threshold needs normalization.
	if failureThreshold <= 0 {
		// Use minimum of 1 for unset threshold.
		failureThreshold = 1
	}

	// Return both normalized values.
	return successThreshold, failureThreshold
}

// updateListenerState updates listener state based on probe result and thresholds.
//
// Params:
//   - lp: the listener probe.
//   - ls: the listener status to update.
//   - result: the probe result.
//   - successThreshold: number of successes needed.
//   - failureThreshold: number of failures needed.
func (m *ProbeMonitor) updateListenerState(lp *ListenerProbe, ls *domain.ListenerStatus, result probe.Result, successThreshold, failureThreshold int) {
	// Check if probe succeeded to update counters accordingly.
	if result.Success {
		// Probe succeeded: reset failures, increment successes.
		ls.ConsecutiveFailures = 0
		ls.ConsecutiveSuccesses++

		// Check if success threshold met to transition to Ready.
		if ls.ConsecutiveSuccesses >= successThreshold {
			// Check return value to prevent state divergence.
			if lp.Listener.MarkReady() {
				ls.State = listener.Ready
			} else {
				// Sync state if transition failed.
				ls.State = lp.Listener.State
			}
		}
	} else {
		// Probe failed: reset successes, increment failures.
		ls.ConsecutiveSuccesses = 0
		ls.ConsecutiveFailures++

		// Check if failure threshold met to transition to Listening.
		if ls.ConsecutiveFailures >= failureThreshold {
			// Check return value to prevent state divergence.
			if lp.Listener.MarkListening() {
				ls.State = listener.Listening
			} else {
				// Sync state if transition failed.
				ls.State = lp.Listener.State
			}
		}
	}
}

// storeProbeResult stores the probe result and updates latency.
//
// Params:
//   - ls: the listener status to update.
//   - result: the probe result to store.
func (m *ProbeMonitor) storeProbeResult(ls *domain.ListenerStatus, result probe.Result) {
	// Store last result with all details.
	ls.LastProbeResult = &domain.Result{
		Status:    m.resultToStatus(result),
		Message:   result.Output,
		Duration:  result.Latency,
		Timestamp: time.Now(),
		Error:     result.Error,
	}

	// Update latency in aggregated health.
	m.health.SetLatency(result.Latency)
}

// findOrCreateListenerStatus finds or creates a listener status entry.
//
// Params:
//   - lp: the listener probe.
//
// Returns:
//   - *domain.ListenerStatus: pointer to the listener status.
func (m *ProbeMonitor) findOrCreateListenerStatus(lp *ListenerProbe) *domain.ListenerStatus {
	// Search for existing listener status.
	for i := range m.health.Listeners {
		// Return existing status if found by name.
		if m.health.Listeners[i].Name == lp.Listener.Name {
			// Return pointer to existing listener status.
			return &m.health.Listeners[i]
		}
	}

	// Create new listener status if not found.
	m.health.Listeners = append(m.health.Listeners, domain.ListenerStatus{
		Name:  lp.Listener.Name,
		State: lp.Listener.State,
	})
	// Return pointer to newly created listener status.
	return &m.health.Listeners[len(m.health.Listeners)-1]
}

// sendEventIfChanged sends a health event if state changed.
//
// Params:
//   - lp: the listener probe.
//   - ls: the listener status.
//   - prevState: the previous listener state.
//   - result: the probe result.
func (m *ProbeMonitor) sendEventIfChanged(lp *ListenerProbe, ls *domain.ListenerStatus, prevState listener.State, result probe.Result) {
	// Check if state changed and events channel is available.
	if prevState != lp.Listener.State && m.events != nil {
		event := domain.NewEvent(lp.Listener.Name, m.resultToStatus(result), *ls.LastProbeResult)

		// Non-blocking send to avoid deadlocks.
		select {
		case m.events <- event:
			// Event sent successfully.
		default:
			// Channel full, skip event.
		}
	}
}

// resultToStatus converts a probe result to a health status.
//
// Params:
//   - result: the probe result.
//
// Returns:
//   - domain.Status: the corresponding health status.
func (m *ProbeMonitor) resultToStatus(result probe.Result) domain.Status {
	// Map success to healthy, failure to unhealthy.
	if result.Success {
		// Return healthy status for successful probe.
		return domain.StatusHealthy
	}
	// Return unhealthy when probe did not succeed.
	return domain.StatusUnhealthy
}

// Status returns the current aggregated health status.
//
// Returns:
//   - domain.Status: the current status.
func (m *ProbeMonitor) Status() domain.Status {
	// Lock for thread-safe read.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return computed status.
	return m.health.Status()
}

// Health returns the full aggregated health.
//
// Returns:
//   - *domain.AggregatedHealth: copy of the aggregated health.
func (m *ProbeMonitor) Health() *domain.AggregatedHealth {
	// Lock for thread-safe read.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a deep copy to avoid data races.
	health := *m.health
	health.Listeners = make([]domain.ListenerStatus, 0, len(m.health.Listeners))
	// Deep copy each listener status including nested pointers.
	for i := range m.health.Listeners {
		// Copy listener status struct.
		lsCopy := m.health.Listeners[i]
		// Deep copy LastProbeResult pointer to avoid concurrent access.
		if lsCopy.LastProbeResult != nil {
			resultCopy := *lsCopy.LastProbeResult
			lsCopy.LastProbeResult = &resultCopy
		}
		health.Listeners = append(health.Listeners, lsCopy)
	}
	// Return pointer to health copy.
	return &health
}

// IsHealthy returns true if all checks are healthy.
//
// Returns:
//   - bool: true if status is healthy.
func (m *ProbeMonitor) IsHealthy() bool {
	// Compare current status against healthy threshold.
	return m.Status() == domain.StatusHealthy
}

// Latency returns the latest probe latency.
//
// Returns:
//   - time.Duration: the latest latency measurement.
func (m *ProbeMonitor) Latency() time.Duration {
	// Lock for thread-safe read.
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return stored latency.
	return m.health.Latency
}
