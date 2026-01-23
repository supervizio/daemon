// Package health provides the application service for health monitoring.
package health

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/listener"
	"github.com/kodflow/daemon/internal/domain/process"
)

// subjectStatus defines the interface for subject status operations.
// This internal interface enables interface-based programming for the updateListenerState method.
type subjectStatus interface {
	// ApplyProbeEvaluation applies a previously computed evaluation.
	ApplyProbeEvaluation(eval domain.ProbeEvaluation)
	// EvaluateProbeResult evaluates a probe result without mutating state.
	EvaluateProbeResult(success bool, successThreshold, failureThreshold int) domain.ProbeEvaluation
	// ResetCounters resets consecutive success and failure counts.
	ResetCounters()
	// SetState updates the subject state.
	SetState(state domain.SubjectState)
}

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
	// wg tracks prober goroutines for clean shutdown.
	wg sync.WaitGroup
	// factory creates probers.
	factory Creator
	// defaultTimeout is the default probe timeout.
	defaultTimeout time.Duration
	// defaultInterval is the default probe interval.
	defaultInterval time.Duration
	// onStateChange is called on health state transitions.
	onStateChange HealthStateLogger
	// onUnhealthy is called when a service becomes unhealthy.
	onUnhealthy UnhealthyCallback
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
		defaultTimeout = domain.DefaultTimeout
	}

	defaultInterval := config.DefaultInterval
	// Use probe default interval when not configured.
	if defaultInterval == 0 {
		defaultInterval = domain.DefaultInterval
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
		onStateChange:   config.OnStateChange,
		onUnhealthy:     config.OnUnhealthy,
	}
}

// listenerStateToSubjectState converts listener.State to domain.SubjectState.
//
// Params:
//   - state: the listener state to convert.
//
// Returns:
//   - domain.SubjectState: the equivalent subject state.
func listenerStateToSubjectState(state listener.State) domain.SubjectState {
	// Map listener state to corresponding subject state.
	switch state {
	// Listener is ready (passed healthcheck).
	case listener.StateReady:
		// Return subject ready state.
		return domain.SubjectReady
	// Listener is listening but not yet ready.
	case listener.StateListening:
		// Return subject listening state.
		return domain.SubjectListening
	// Listener is closed or stopped.
	case listener.StateClosed:
		// Return subject closed state.
		return domain.SubjectClosed
	// Unrecognized or invalid listener state.
	default:
		// Return unknown subject state for safety.
		return domain.SubjectUnknown
	}
}

// AddListener adds a listener to monitor without probe binding.
//
// Params:
//   - l: the listener to add.
//
// Returns:
//   - error: nil on success.
func (m *ProbeMonitor) AddListener(l *listener.Listener) error {
	// Lock for thread-safe update.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add listener without binding (no probing).
	m.listeners = append(m.listeners, NewListenerProbe(l))
	// Return success for listener without healthcheck.
	return nil
}

// AddListenerWithBinding adds a listener with probe binding to monitor.
//
// Params:
//   - l: the listener to add.
//   - binding: the probe binding configuration.
//
// Returns:
//   - error: if prober creation fails.
func (m *ProbeMonitor) AddListenerWithBinding(l *listener.Listener, binding *ProbeBinding) error {
	// Lock for thread-safe update.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create listener probe with binding.
	lp := NewListenerProbeWithBinding(l, binding)

	// Create prober if binding is provided.
	if binding != nil {
		prober, err := m.createProberFromBinding(binding)
		// Check for prober creation failure.
		if err != nil {
			// Return prober creation error.
			return err
		}
		lp.Prober = prober
	}

	// Add listener probe.
	m.listeners = append(m.listeners, lp)
	// Return success.
	return nil
}

// createProberFromBinding creates a prober from a probe binding.
//
// Params:
//   - binding: the probe binding configuration.
//
// Returns:
//   - domain.Prober: the created prober.
//   - error: if factory is missing or creation fails.
func (m *ProbeMonitor) createProberFromBinding(binding *ProbeBinding) (domain.Prober, error) {
	// Check if factory is configured.
	if m.factory == nil {
		// Return error when factory is missing.
		return nil, ErrProberFactoryMissing
	}

	// Validate probe type is not empty.
	if binding.Type == "" {
		// Return error when probe type is missing.
		return nil, ErrEmptyProbeType
	}

	// Use binding timeout, falling back to default if not set.
	timeout := binding.Config.Timeout
	// Apply default timeout when not configured.
	if timeout == 0 {
		timeout = m.defaultTimeout
	}

	// Create prober with effective timeout.
	prober, err := m.factory.Create(string(binding.Type), timeout)
	// Check for factory creation failure.
	if err != nil {
		// Wrap factory error with listener context.
		return nil, fmt.Errorf("create prober for binding %q: %w", binding.ListenerName, err)
	}
	// Return successfully created prober.
	return prober, nil
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
	listeners := slices.Clone(m.listeners)
	m.mu.Unlock()

	// Start a goroutine for each listener with a prober.
	// Pass stopCh as parameter to avoid race conditions on restart.
	for _, lp := range listeners {
		// Only start probers for listeners that have one configured.
		if lp.Prober != nil {
			m.wg.Add(1)
			go func(lp *ListenerProbe) {
				defer m.wg.Done()
				m.runProber(ctx, stopCh, lp)
			}(lp)
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

	// Wait for all prober goroutines to terminate.
	m.wg.Wait()
}

// runProber runs a single prober in a loop.
//
// Params:
//   - ctx: context for cancellation.
//   - stopCh: channel to signal stop (passed as param to avoid race on restart).
//   - lp: the listener probe to run.
func (m *ProbeMonitor) runProber(ctx context.Context, stopCh <-chan struct{}, lp *ListenerProbe) {
	config := lp.ProbeConfig()
	interval := config.Interval
	// Use default interval when not specified in config.
	if interval == 0 {
		interval = m.defaultInterval
	}

	// Create ticker for periodic probes.
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Perform initial probe immediately (unless already stopped/cancelled).
	select {
	case <-stopCh:
		// Stop signal received before initial healthcheck.
		return
	case <-ctx.Done():
		// Context cancelled before initial healthcheck.
		return
	default:
		// Perform initial healthcheck.
		m.performProbe(ctx, lp)
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
			// Perform periodic healthcheck.
			m.performProbe(ctx, lp)
		}
	}
}

// performProbe performs a single healthcheck.
//
// Params:
//   - ctx: parent context.
//   - lp: the listener probe to use.
func (m *ProbeMonitor) performProbe(ctx context.Context, lp *ListenerProbe) {
	// Guard against nil prober to prevent panic.
	if lp.Prober == nil {
		// Skip probe execution when prober is not configured.
		return
	}

	config := lp.ProbeConfig()
	timeout := config.Timeout
	// Use default timeout when not specified in config.
	if timeout == 0 {
		timeout = m.defaultTimeout
	}

	// Create timeout context for this healthcheck.
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Get target from listener probe.
	target := lp.ProbeTarget()

	// Execute the healthcheck.
	result := lp.Prober.Probe(probeCtx, target)

	// Update state with probe result.
	m.updateProbeResult(lp, result)
}

// updateProbeResult updates the listener status based on probe result.
//
// Params:
//   - lp: the listener probe that was executed.
//   - result: the probe result.
func (m *ProbeMonitor) updateProbeResult(lp *ListenerProbe, result domain.CheckResult) {
	// Lock for thread-safe state updates.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find or create listener status in health.
	ls := m.findOrCreateSubjectStatus(lp)

	// Get previous state for event comparison.
	prevState := lp.Listener.State

	// Store previous failure count for threshold checking.
	prevFailures := ls.ConsecutiveFailures

	// Get probe config for thresholds.
	config := lp.ProbeConfig()

	// Get normalized thresholds (zero values become 1).
	successThreshold, failureThreshold := m.normalizeThresholds(config)

	// Update state based on probe result.
	m.updateListenerState(lp, ls, result, successThreshold, failureThreshold)

	// Store the probe result and update latency.
	m.storeProbeResult(ls, result)

	// Send event if state changed.
	m.sendEventIfChanged(lp, ls, prevState, result)

	// Check if failure threshold was just reached (triggers restart).
	m.checkFailureThresholdReached(lp, ls, prevFailures, result)
}

// normalizeThresholds returns thresholds with zero values normalized to 1.
//
// Params:
//   - config: the probe configuration.
//
// Returns:
//   - int: normalized success threshold (minimum 1).
//   - int: normalized failure threshold (minimum 1).
func (m *ProbeMonitor) normalizeThresholds(config domain.CheckConfig) (successThreshold, failureThreshold int) {
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
// Uses the domain's pure EvaluateProbeResult to compute state changes,
// then applies only if the Listener accepts the transition.
//
// Params:
//   - lp: the listener probe.
//   - ls: the subject status to update.
//   - result: the probe result.
//   - successThreshold: number of successes needed.
//   - failureThreshold: number of failures needed.
func (m *ProbeMonitor) updateListenerState(lp *ListenerProbe, ls subjectStatus, result domain.CheckResult, successThreshold, failureThreshold int) {
	// 1. Pure evaluation - no side effects, computes what should happen.
	eval := ls.EvaluateProbeResult(result.Success, successThreshold, failureThreshold)

	// 2. No transition needed - safe to apply counters directly.
	if !eval.ShouldTransition {
		ls.ApplyProbeEvaluation(eval)

		// Return early when no state transition is needed.
		return
	}

	// 3. Check if already in target state (no transition needed).
	currentSubjectState := listenerStateToSubjectState(lp.Listener.State)
	// Skip state transition when already in target state.
	if currentSubjectState == eval.TargetState {
		// Already in target state - apply counter updates but skip state change.
		ls.ApplyProbeEvaluation(eval)

		// Return early when already in target state.
		return
	}

	// 4. Check if Listener accepts the transition.
	var accepted bool

	//exhaustive:ignore
	switch eval.TargetState {
	// Handle transition to Ready state.
	case domain.SubjectReady:
		accepted = lp.Listener.MarkReady()
	// Handle transition to Listening state.
	case domain.SubjectListening:
		accepted = lp.Listener.MarkListening()
	// Handle invalid transition targets for listeners.
	case domain.SubjectUnknown, domain.SubjectClosed, domain.SubjectRunning, domain.SubjectStopped, domain.SubjectFailed:
		// These states are not valid transition targets for listeners.
		accepted = false
	}

	// 5. Apply evaluation only if Listener accepted the transition.
	if accepted {
		ls.ApplyProbeEvaluation(eval)
	} else {
		// Listener refused for unexpected reason - sync with Listener's actual state.
		ls.SetState(listenerStateToSubjectState(lp.Listener.State))
		// Reset counters to avoid drift between domain and listener state.
		ls.ResetCounters()
	}
}

// storeProbeResult stores the probe result and updates latency.
//
// Params:
//   - ls: the subject status to update.
//   - result: the probe result to store.
func (m *ProbeMonitor) storeProbeResult(ls *domain.SubjectStatus, result domain.CheckResult) {
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

// findOrCreateSubjectStatus finds or creates a subject status entry.
//
// Params:
//   - lp: the listener probe.
//
// Returns:
//   - *domain.SubjectStatus: pointer to the subject status.
func (m *ProbeMonitor) findOrCreateSubjectStatus(lp *ListenerProbe) *domain.SubjectStatus {
	// Search for existing subject status.
	for i := range m.health.Subjects {
		// Return existing status if found by name.
		if m.health.Subjects[i].Name == lp.Listener.Name {
			// Return pointer to existing subject status.
			return &m.health.Subjects[i]
		}
	}

	// Create new subject status if not found.
	m.health.Subjects = append(m.health.Subjects, domain.SubjectStatus{
		Name:  lp.Listener.Name,
		State: listenerStateToSubjectState(lp.Listener.State),
	})
	// Return pointer to newly created subject status.
	return &m.health.Subjects[len(m.health.Subjects)-1]
}

// extractFailureReason extracts a human-readable reason from a probe result.
//
// Params:
//   - result: the probe result.
//
// Returns:
//   - string: the failure reason extracted from error or output.
func extractFailureReason(result domain.CheckResult) string {
	// Check if error message is available.
	if result.Error != nil {
		// Return the error message string.
		return result.Error.Error()
	}
	// Check if output is available.
	if result.Output != "" {
		// Return the output as reason.
		return result.Output
	}
	// Return default failure message.
	return "health probe failed"
}

// sendEventIfChanged sends a health event if state changed.
//
// Params:
//   - lp: the listener probe.
//   - ls: the subject status.
//   - prevState: the previous listener state.
//   - result: the probe result.
func (m *ProbeMonitor) sendEventIfChanged(lp *ListenerProbe, ls *domain.SubjectStatus, prevState listener.State, result domain.CheckResult) {
	// Convert previous state to subject state for comparison.
	prevSubjectState := listenerStateToSubjectState(prevState)

	// Check if state has changed.
	if prevSubjectState == ls.State {
		// Return early when state unchanged.
		return
	}

	// Notify state change callback if configured.
	m.notifyStateChange(lp.Listener.Name, prevSubjectState, ls.State, result)

	// Handle unhealthy transition (ready -> listening).
	m.handleUnhealthyTransition(lp.Listener.Name, prevSubjectState, ls.State, result)

	// Send event to channel.
	m.sendEvent(lp.Listener.Name, ls, result)
}

// notifyStateChange calls the state change callback if configured.
//
// Params:
//   - name: the listener name.
//   - prevState: the previous subject state.
//   - newState: the new subject state.
//   - result: the probe result.
func (m *ProbeMonitor) notifyStateChange(name string, prevState, newState domain.SubjectState, result domain.CheckResult) {
	// Check if callback is configured.
	if m.onStateChange == nil {
		// Return early when no callback configured.
		return
	}
	// Call the state change callback.
	m.onStateChange(name, prevState, newState, result)
}

// handleUnhealthyTransition triggers unhealthy callback on ready->listening transition.
//
// Params:
//   - name: the listener name.
//   - prevState: the previous subject state.
//   - newState: the new subject state.
//   - result: the probe result.
func (m *ProbeMonitor) handleUnhealthyTransition(name string, prevState, newState domain.SubjectState, result domain.CheckResult) {
	// Check if this is a ready -> listening transition.
	if newState != domain.SubjectListening || prevState != domain.SubjectReady {
		// Return early for non-unhealthy transitions.
		return
	}
	// Check if callback is configured.
	if m.onUnhealthy == nil {
		// Return early when no callback configured.
		return
	}
	// Extract failure reason and call callback.
	reason := extractFailureReason(result)
	m.onUnhealthy(name, reason)
}

// sendEvent sends a health event to the events channel.
//
// Params:
//   - name: the listener name.
//   - ls: the subject status.
//   - result: the probe result.
func (m *ProbeMonitor) sendEvent(name string, ls *domain.SubjectStatus, result domain.CheckResult) {
	// Check if events channel is configured.
	if m.events == nil {
		// Return early when no event channel.
		return
	}
	// Check if probe result is stored.
	if ls.LastProbeResult == nil {
		// Return early when no probe result available.
		return
	}

	// Create and send event.
	event := domain.NewEvent(name, m.resultToStatus(result), *ls.LastProbeResult)

	// Non-blocking send to avoid deadlocks.
	select {
	case m.events <- event:
		// Event sent successfully.
	default:
		// Channel full, skip event.
	}
}

// checkFailureThresholdReached checks if failure threshold was just reached.
// This implements the Kubernetes liveness probe pattern: when consecutive failures
// reach the failure threshold, the service should be restarted regardless of
// whether there was a state transition.
//
// Params:
//   - lp: the listener probe.
//   - ls: the subject status.
//   - prevFailures: the failure count before the probe.
//   - result: the probe result.
func (m *ProbeMonitor) checkFailureThresholdReached(lp *ListenerProbe, ls *domain.SubjectStatus, prevFailures int, result domain.CheckResult) {
	// Check if probe was successful.
	if result.Success {
		// Return early on successful probes.
		return
	}

	// Get the failure threshold from config.
	failureThreshold := m.getFailureThreshold(lp)

	// Check if threshold was crossed.
	if prevFailures >= failureThreshold || ls.ConsecutiveFailures < failureThreshold {
		// Return early when threshold not crossed.
		return
	}

	// Check if callback is configured.
	if m.onUnhealthy == nil {
		// Return early when no callback configured.
		return
	}

	// Trigger unhealthy callback with extracted reason.
	reason := extractFailureReason(result)
	m.onUnhealthy(lp.Listener.Name, reason)

	// Reset failure counter after triggering restart.
	// This gives the restarted process a fresh chance (Kubernetes pattern).
	ls.ResetCounters()
}

// getFailureThreshold returns the failure threshold from config with a default of 1.
//
// Params:
//   - lp: the listener probe.
//
// Returns:
//   - int: the failure threshold (minimum 1).
func (m *ProbeMonitor) getFailureThreshold(lp *ListenerProbe) int {
	config := lp.ProbeConfig()
	// Check if threshold is valid.
	if config.FailureThreshold <= 0 {
		// Return minimum threshold of 1.
		return 1
	}
	// Return configured threshold.
	return config.FailureThreshold
}

// resultToStatus converts a probe result to a health status.
//
// Params:
//   - result: the probe result.
//
// Returns:
//   - domain.Status: the corresponding health status.
func (m *ProbeMonitor) resultToStatus(result domain.CheckResult) domain.Status {
	// Map success to healthy, failure to unhealthy.
	if result.Success {
		// Return healthy status for successful healthcheck.
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
	health.Subjects = make([]domain.SubjectStatus, 0, len(m.health.Subjects))
	// Deep copy each subject status including nested pointers.
	for i := range m.health.Subjects {
		// Copy subject status struct.
		ssCopy := m.health.Subjects[i]
		// Deep copy LastProbeResult pointer to avoid concurrent access.
		if ssCopy.LastProbeResult != nil {
			resultCopy := *ssCopy.LastProbeResult
			ssCopy.LastProbeResult = &resultCopy
		}
		health.Subjects = append(health.Subjects, ssCopy)
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
