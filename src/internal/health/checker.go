// Package health provides health checking for services.
package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kodflow/daemon/internal/config"
)

// Status represents the health status of a service.
type Status int

const (
	StatusUnknown Status = iota
	StatusHealthy
	StatusUnhealthy
	StatusDegraded
)

func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusUnhealthy:
		return "unhealthy"
	case StatusDegraded:
		return "degraded"
	default:
		return "unknown"
	}
}

// Result represents the result of a health check.
type Result struct {
	Status    Status
	Message   string
	Duration  time.Duration
	Timestamp time.Time
	Error     error
}

// Checker is the interface for health checkers.
type Checker interface {
	// Check performs a health check and returns the result.
	Check(ctx context.Context) Result
	// Name returns the name of this checker.
	Name() string
	// Type returns the type of this checker.
	Type() string
}

// Monitor monitors the health of a service using multiple checkers.
type Monitor struct {
	mu          sync.RWMutex
	checkers    []Checker
	config      []config.HealthCheckConfig
	status      Status
	results     map[string]Result
	consecutive map[string]int
	events      chan<- Event
	stopCh      chan struct{}
	running     bool
}

// Event represents a health check event.
type Event struct {
	Checker   string
	Status    Status
	Result    Result
	Timestamp time.Time
}

// NewMonitor creates a new health monitor from configuration.
func NewMonitor(configs []config.HealthCheckConfig, events chan<- Event) (*Monitor, error) {
	m := &Monitor{
		config:      configs,
		checkers:    make([]Checker, 0, len(configs)),
		results:     make(map[string]Result),
		consecutive: make(map[string]int),
		events:      events,
		stopCh:      make(chan struct{}),
		status:      StatusUnknown,
	}

	for i, cfg := range configs {
		checker, err := NewChecker(&cfg)
		if err != nil {
			return nil, fmt.Errorf("creating checker %d: %w", i, err)
		}
		m.checkers = append(m.checkers, checker)
	}

	return m, nil
}

// NewChecker creates a checker from configuration.
func NewChecker(cfg *config.HealthCheckConfig) (Checker, error) {
	switch cfg.Type {
	case config.HealthCheckHTTP:
		return NewHTTPChecker(cfg), nil
	case config.HealthCheckTCP:
		return NewTCPChecker(cfg), nil
	case config.HealthCheckCommand:
		return NewCommandChecker(cfg), nil
	default:
		return nil, fmt.Errorf("unknown health check type: %s", cfg.Type)
	}
}

// Start starts the health monitor.
func (m *Monitor) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.stopCh = make(chan struct{})
	m.mu.Unlock()

	for i, checker := range m.checkers {
		go m.runChecker(ctx, checker, &m.config[i])
	}
}

// Stop stops the health monitor.
func (m *Monitor) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	close(m.stopCh)
	m.mu.Unlock()
}

// runChecker runs a single checker in a loop.
func (m *Monitor) runChecker(ctx context.Context, checker Checker, cfg *config.HealthCheckConfig) {
	ticker := time.NewTicker(cfg.Interval.Duration())
	defer ticker.Stop()

	// Initial check
	m.performCheck(ctx, checker, cfg)

	for {
		select {
		case <-m.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.performCheck(ctx, checker, cfg)
		}
	}
}

// performCheck performs a single health check.
func (m *Monitor) performCheck(ctx context.Context, checker Checker, cfg *config.HealthCheckConfig) {
	checkCtx, cancel := context.WithTimeout(ctx, cfg.Timeout.Duration())
	defer cancel()

	result := checker.Check(checkCtx)

	m.mu.Lock()
	defer m.mu.Unlock()

	name := checker.Name()
	prevResult, hasPrev := m.results[name]
	m.results[name] = result

	// Track consecutive failures/successes
	if result.Status == StatusHealthy {
		m.consecutive[name] = 0
	} else {
		m.consecutive[name]++
	}

	// Update overall status
	m.updateStatus()

	// Emit event if status changed
	if !hasPrev || prevResult.Status != result.Status {
		if m.events != nil {
			event := Event{
				Checker:   name,
				Status:    result.Status,
				Result:    result,
				Timestamp: time.Now(),
			}
			select {
			case m.events <- event:
			default:
				// Drop if channel full
			}
		}
	}
}

// updateStatus updates the overall status based on all checkers.
func (m *Monitor) updateStatus() {
	if len(m.results) == 0 {
		m.status = StatusUnknown
		return
	}

	unhealthyCount := 0
	for _, result := range m.results {
		if result.Status == StatusUnhealthy {
			unhealthyCount++
		}
	}

	switch {
	case unhealthyCount == len(m.results):
		m.status = StatusUnhealthy
	case unhealthyCount > 0:
		m.status = StatusDegraded
	default:
		m.status = StatusHealthy
	}
}

// Status returns the current health status.
func (m *Monitor) Status() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// Results returns all check results.
func (m *Monitor) Results() map[string]Result {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make(map[string]Result, len(m.results))
	for k, v := range m.results {
		results[k] = v
	}
	return results
}

// IsHealthy returns true if all checks are healthy.
func (m *Monitor) IsHealthy() bool {
	return m.Status() == StatusHealthy
}
