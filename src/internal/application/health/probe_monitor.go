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

// ProberFactory creates probers based on type.
// This is the port that infrastructure adapters implement.
type ProberFactory interface {
	// Create creates a prober of the specified type.
	//
	// Params:
	//   - proberType: the type of prober to create.
	//   - timeout: the timeout for the prober.
	//
	// Returns:
	//   - probe.Prober: the created prober.
	//   - error: if creation fails.
	Create(proberType string, timeout time.Duration) (probe.Prober, error)
}

// ListenerProbe represents a listener with its probe configuration.
type ListenerProbe struct {
	// Listener is the listener to probe.
	Listener *listener.Listener
	// Prober is the prober to use.
	Prober probe.Prober
	// Config is the probe configuration.
	Config probe.Config
	// Target is the probe target.
	Target probe.Target
}

// ProbeMonitor monitors health using the new Prober interface.
// It supports multiple listeners with different probe types.
type ProbeMonitor struct {
	mu              sync.RWMutex
	listeners       []*ListenerProbe
	health          *domain.AggregatedHealth
	processState    process.State
	customStatus    string
	events          chan<- domain.Event
	stopCh          chan struct{}
	running         bool
	factory         ProberFactory
	defaultTimeout  time.Duration
	defaultInterval time.Duration
}

// ProbeMonitorConfig contains configuration for ProbeMonitor.
type ProbeMonitorConfig struct {
	// Factory creates probers.
	Factory ProberFactory
	// Events channel for health events.
	Events chan<- domain.Event
	// DefaultTimeout for probes.
	DefaultTimeout time.Duration
	// DefaultInterval between probes.
	DefaultInterval time.Duration
}

// NewProbeMonitor creates a new probe-based health monitor.
//
// Params:
//   - config: the monitor configuration.
//
// Returns:
//   - *ProbeMonitor: the new monitor instance.
func NewProbeMonitor(config ProbeMonitorConfig) *ProbeMonitor {
	// Set defaults if not specified.
	defaultTimeout := config.DefaultTimeout
	if defaultTimeout == 0 {
		defaultTimeout = probe.DefaultTimeout
	}
	defaultInterval := config.DefaultInterval
	if defaultInterval == 0 {
		defaultInterval = probe.DefaultInterval
	}

	// Return configured monitor.
	return &ProbeMonitor{
		listeners:       make([]*ListenerProbe, 0),
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

	// Skip if no probe configured.
	if !l.HasProbe() {
		// Add listener without prober.
		m.listeners = append(m.listeners, &ListenerProbe{
			Listener: l,
		})
		return nil
	}

	// Create prober for this listener.
	prober, err := m.factory.Create(l.ProbeType, l.ProbeConfig.Timeout)
	if err != nil {
		return err
	}

	// Add listener with prober.
	m.listeners = append(m.listeners, &ListenerProbe{
		Listener: l,
		Prober:   prober,
		Config:   *l.ProbeConfig,
		Target:   l.ProbeTarget,
	})

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
	// Update process state.
	m.processState = state
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
	// Update custom status.
	m.customStatus = status
	m.health.SetCustomStatus(status)
}

// Start starts the probe monitor.
//
// Params:
//   - ctx: context for cancellation.
func (m *ProbeMonitor) Start(ctx context.Context) {
	// Lock to check and update running state.
	m.mu.Lock()
	// Check if already running.
	if m.running {
		m.mu.Unlock()
		return
	}
	// Mark as running.
	m.running = true
	// Create new stop channel.
	m.stopCh = make(chan struct{})
	m.mu.Unlock()

	// Start a goroutine for each listener with a prober.
	for _, lp := range m.listeners {
		if lp.Prober != nil {
			// Run prober in its own goroutine.
			go m.runProber(ctx, lp)
		}
	}
}

// Stop stops the probe monitor.
func (m *ProbeMonitor) Stop() {
	// Lock to check and update running state.
	m.mu.Lock()
	// Check if not running.
	if !m.running {
		m.mu.Unlock()
		return
	}
	// Mark as not running.
	m.running = false
	// Signal all goroutines to stop.
	close(m.stopCh)
	m.mu.Unlock()
}

// runProber runs a single prober in a loop.
//
// Params:
//   - ctx: context for cancellation.
//   - lp: the listener probe to run.
func (m *ProbeMonitor) runProber(ctx context.Context, lp *ListenerProbe) {
	// Get interval.
	interval := lp.Config.Interval
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
		// Handle stop signal.
		case <-m.stopCh:
			return
		// Handle context cancellation.
		case <-ctx.Done():
			return
		// Handle ticker tick.
		case <-ticker.C:
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
	// Get timeout.
	timeout := lp.Config.Timeout
	if timeout == 0 {
		timeout = m.defaultTimeout
	}

	// Create timeout context for this probe.
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build target with listener address.
	target := lp.Target
	if target.Address == "" {
		target.Address = lp.Listener.GetProbeAddress()
	}

	// Execute the probe.
	result := lp.Prober.Probe(probeCtx, target)

	// Lock for thread-safe state updates.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find listener status in health.
	var ls *domain.ListenerStatus
	for i := range m.health.Listeners {
		if m.health.Listeners[i].Name == lp.Listener.Name {
			ls = &m.health.Listeners[i]
			break
		}
	}

	// Create listener status if not found.
	if ls == nil {
		m.health.Listeners = append(m.health.Listeners, domain.ListenerStatus{
			Name:  lp.Listener.Name,
			State: lp.Listener.State,
		})
		ls = &m.health.Listeners[len(m.health.Listeners)-1]
	}

	// Get previous state for event comparison.
	prevState := lp.Listener.State

	// Update based on result.
	if result.Success {
		// Reset failure count, increment success count.
		ls.ConsecutiveFailures = 0
		ls.ConsecutiveSuccesses++
		// Check if threshold met.
		if ls.ConsecutiveSuccesses >= lp.Config.SuccessThreshold {
			// Transition to Ready.
			lp.Listener.MarkReady()
			ls.State = listener.Ready
		}
	} else {
		// Reset success count, increment failure count.
		ls.ConsecutiveSuccesses = 0
		ls.ConsecutiveFailures++
		// Check if threshold met.
		if ls.ConsecutiveFailures >= lp.Config.FailureThreshold {
			// Transition to Listening (not ready).
			ls.State = listener.Listening
		}
	}

	// Store last result.
	ls.LastProbeResult = &domain.Result{
		Status:    m.resultToStatus(result),
		Message:   result.Output,
		Duration:  result.Latency,
		Timestamp: time.Now(),
		Error:     result.Error,
	}

	// Update latency.
	m.health.SetLatency(result.Latency)

	// Send event if state changed.
	if prevState != lp.Listener.State && m.events != nil {
		event := domain.NewEvent(lp.Listener.Name, m.resultToStatus(result), *ls.LastProbeResult)
		select {
		case m.events <- event:
		default:
		}
	}
}

// resultToStatus converts a probe result to a health status.
func (m *ProbeMonitor) resultToStatus(result probe.Result) domain.Status {
	if result.Success {
		return domain.StatusHealthy
	}
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

	// Return a copy.
	health := *m.health
	health.Listeners = make([]domain.ListenerStatus, len(m.health.Listeners))
	copy(health.Listeners, m.health.Listeners)
	return &health
}

// IsHealthy returns true if all checks are healthy.
//
// Returns:
//   - bool: true if status is healthy.
func (m *ProbeMonitor) IsHealthy() bool {
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
	return m.health.Latency
}
