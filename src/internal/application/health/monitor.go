// Package health provides the application service for health monitoring.
package health

import (
	"context"
	"maps"
	"sync"
	"time"

	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/service"
)

// Monitor monitors the health of a service using multiple checkers.
// It runs periodic health checks and aggregates results to determine
// the overall health status of the monitored service.
type Monitor struct {
	mu          sync.RWMutex
	checkers    []Checker
	config      []service.HealthCheckConfig
	status      domain.Status
	results     map[string]domain.Result
	consecutive map[string]int
	events      chan<- domain.Event
	stopCh      chan struct{}
	running     bool
}

// NewMonitor creates a new health monitor from configuration.
//
// Params:
//   - configs: slice of health check configurations to use.
//   - creator: creator for creating health checkers.
//   - events: channel to send health events to (can be nil).
//
// Returns:
//   - *Monitor: the new monitor instance.
//   - error: if checker creation fails.
func NewMonitor(configs []service.HealthCheckConfig, creator Creator, events chan<- domain.Event) (*Monitor, error) {
	// Initialize the monitor with empty collections.
	m := &Monitor{
		config:      configs,
		checkers:    make([]Checker, 0, len(configs)),
		results:     make(map[string]domain.Result, len(configs)),
		consecutive: make(map[string]int, len(configs)),
		events:      events,
		stopCh:      make(chan struct{}),
		status:      domain.StatusUnknown,
	}

	// Create a checker for each configuration.
	for i := range configs {
		// Build checker configuration from service config.
		cfg := CheckerConfig{
			Name:       configs[i].Name,
			Type:       configs[i].Type.String(),
			Endpoint:   configs[i].Endpoint,
			Method:     configs[i].Method,
			StatusCode: configs[i].StatusCode,
			Host:       configs[i].Host,
			Port:       configs[i].Port,
			Command:    configs[i].Command,
		}
		// Create checker using the creator.
		checker, err := creator.Create(cfg)
		// Return error if checker creation fails.
		if err != nil {
			// Return nil monitor and the error.
			return nil, err
		}
		// Append the new checker to the list.
		m.checkers = append(m.checkers, checker)
	}

	// Return the configured monitor.
	return m, nil
}

// Start starts the health monitor.
// This method spawns one goroutine per health checker to run periodic health checks.
// Each goroutine terminates when the context is cancelled or Stop() is called.
// To stop the monitor gracefully, call Stop() or cancel the provided context.
//
// Params:
//   - ctx: context for cancellation.
func (m *Monitor) Start(ctx context.Context) {
	// Lock to check and update running state.
	m.mu.Lock()
	// Check if already running.
	if m.running {
		// Unlock and return early.
		m.mu.Unlock()
		// Exit without starting again.
		return
	}
	// Mark as running.
	m.running = true
	// Create new stop channel.
	m.stopCh = make(chan struct{})
	// Unlock before starting goroutines.
	m.mu.Unlock()

	// Start a goroutine for each checker.
	for i, checker := range m.checkers {
		// Run checker in its own goroutine.
		go m.runChecker(ctx, checker, &m.config[i])
	}
}

// Stop stops the health monitor.
func (m *Monitor) Stop() {
	// Lock to check and update running state.
	m.mu.Lock()
	// Check if not running.
	if !m.running {
		// Unlock and return early.
		m.mu.Unlock()
		// Exit without stopping.
		return
	}
	// Mark as not running.
	m.running = false
	// Signal all goroutines to stop.
	close(m.stopCh)
	// Unlock after state update.
	m.mu.Unlock()
}

// runChecker runs a single checker in a loop.
//
// Params:
//   - ctx: context for cancellation.
//   - checker: the health checker to run.
//   - cfg: configuration for this checker.
func (m *Monitor) runChecker(ctx context.Context, checker Checker, cfg *service.HealthCheckConfig) {
	// Create ticker for periodic checks.
	ticker := time.NewTicker(cfg.Interval.Duration())
	// Ensure ticker is stopped on exit.
	defer ticker.Stop()

	// Perform initial check immediately.
	m.performCheck(ctx, checker, cfg)

	// Loop until stopped.
	for {
		// Wait for next event.
		select {
		// Handle stop signal.
		case <-m.stopCh:
			// Exit the loop.
			return
		// Handle context cancellation.
		case <-ctx.Done():
			// Exit the loop.
			return
		// Handle ticker tick.
		case <-ticker.C:
			// Perform the health check.
			m.performCheck(ctx, checker, cfg)
		}
	}
}

// performCheck performs a single health check.
//
// Params:
//   - ctx: parent context.
//   - checker: the health checker to use.
//   - cfg: configuration for this checker.
func (m *Monitor) performCheck(ctx context.Context, checker Checker, cfg *service.HealthCheckConfig) {
	// Create timeout context for this check.
	checkCtx, cancel := context.WithTimeout(ctx, cfg.Timeout.Duration())
	// Ensure context is cancelled after check.
	defer cancel()

	// Execute the health check.
	result := checker.Check(checkCtx)

	// Lock for thread-safe state updates.
	m.mu.Lock()
	// Ensure unlock on exit.
	defer m.mu.Unlock()

	// Get the checker name.
	name := checker.Name()
	// Get previous result if any.
	prevResult, hasPrev := m.results[name]
	// Store new result.
	m.results[name] = result

	// Update consecutive failure count.
	if result.Status == domain.StatusHealthy {
		// Reset count on healthy result.
		m.consecutive[name] = 0
	} else {
		// Increment count on unhealthy result.
		m.consecutive[name]++
	}

	// Update overall status.
	m.updateStatus()

	// Send event if status changed.
	if !hasPrev || prevResult.Status != result.Status {
		// Check if event channel is available.
		if m.events != nil {
			// Create the health event.
			event := domain.NewEvent(name, result.Status, result)
			// Try to send event without blocking.
			select {
			// Send event to channel.
			case m.events <- event:
			// Skip if channel is full.
			default:
			}
		}
	}
}

// updateStatus updates the overall status based on all checkers.
func (m *Monitor) updateStatus() {
	// Check if no results yet.
	if len(m.results) == 0 {
		// Set status to unknown.
		m.status = domain.StatusUnknown
		// Exit early.
		return
	}

	// Count unhealthy checkers.
	unhealthyCount := 0
	// Iterate through all results.
	for _, result := range m.results {
		// Check if result is unhealthy.
		if result.Status == domain.StatusUnhealthy {
			// Increment unhealthy count.
			unhealthyCount++
		}
	}

	// Determine overall status based on unhealthy count.
	switch {
	// All checkers are unhealthy.
	case unhealthyCount == len(m.results):
		// Set status to unhealthy.
		m.status = domain.StatusUnhealthy
	// Some checkers are unhealthy.
	case unhealthyCount > 0:
		// Set status to degraded.
		m.status = domain.StatusDegraded
	// All checkers are healthy.
	default:
		// Set status to healthy.
		m.status = domain.StatusHealthy
	}
}

// Status returns the current health status.
//
// Returns:
//   - domain.Status: the current aggregated health status.
func (m *Monitor) Status() domain.Status {
	// Lock for thread-safe read.
	m.mu.RLock()
	// Unlock after read.
	defer m.mu.RUnlock()
	// Return current status.
	return m.status
}

// Results returns all check results.
//
// Returns:
//   - map[string]domain.Result: copy of all checker results by name.
func (m *Monitor) Results() map[string]domain.Result {
	// Lock for thread-safe read.
	m.mu.RLock()
	// Unlock after read.
	defer m.mu.RUnlock()
	// Return cloned results map.
	return maps.Clone(m.results)
}

// IsHealthy returns true if all checks are healthy.
//
// Returns:
//   - bool: true if status is healthy.
func (m *Monitor) IsHealthy() bool {
	// Check if status equals healthy.
	return m.Status() == domain.StatusHealthy
}
