package process

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kodflow/daemon/internal/config"
)

// Event represents a process lifecycle event.
type Event struct {
	Type      EventType
	Process   string
	PID       int
	ExitCode  int
	Timestamp time.Time
	Error     error
}

// EventType represents the type of lifecycle event.
type EventType int

const (
	EventStarted EventType = iota
	EventStopped
	EventFailed
	EventRestarting
	EventHealthy
	EventUnhealthy
)

func (e EventType) String() string {
	switch e {
	case EventStarted:
		return "started"
	case EventStopped:
		return "stopped"
	case EventFailed:
		return "failed"
	case EventRestarting:
		return "restarting"
	case EventHealthy:
		return "healthy"
	case EventUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// Manager manages the lifecycle of a single process with restart policies.
type Manager struct {
	mu      sync.RWMutex
	process *Process
	config  *config.ServiceConfig
	events  chan Event
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
}

// NewManager creates a new process lifecycle manager.
func NewManager(cfg *config.ServiceConfig) *Manager {
	return &Manager{
		process: New(cfg),
		config:  cfg,
		events:  make(chan Event, 16),
	}
}

// Process returns the underlying process.
func (m *Manager) Process() *Process {
	return m.process
}

// Events returns the event channel for monitoring.
func (m *Manager) Events() <-chan Event {
	return m.events
}

// State returns the current process state.
func (m *Manager) State() State {
	return m.process.State()
}

// PID returns the current process PID.
func (m *Manager) PID() int {
	return m.process.PID()
}

// Uptime returns the process uptime in seconds.
func (m *Manager) Uptime() int64 {
	return int64(m.process.Uptime().Seconds())
}

// Start starts the managed process with automatic restart handling.
func (m *Manager) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("manager already running")
	}
	m.running = true
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.mu.Unlock()

	go m.run()
	return nil
}

// run is the main loop that manages the process lifecycle.
func (m *Manager) run() {
	defer func() {
		m.mu.Lock()
		m.running = false
		m.mu.Unlock()
	}()

	if m.config.Oneshot {
		m.runOnce()
		return
	}

	m.runWithRestart()
}

// runOnce runs the process once without restart.
func (m *Manager) runOnce() {
	if err := m.process.Start(m.ctx); err != nil {
		m.sendEvent(EventFailed, err)
		return
	}

	m.sendEvent(EventStarted, nil)

	<-m.process.Wait()

	if m.process.ExitCode() != 0 {
		m.sendEvent(EventFailed, fmt.Errorf("exit code: %d", m.process.ExitCode()))
		return
	}

	m.sendEvent(EventStopped, nil)
}

// runWithRestart runs the process with automatic restart based on policy.
func (m *Manager) runWithRestart() {
	for {
		select {
		case <-m.ctx.Done():
			return
		default:
		}

		if err := m.process.Start(m.ctx); err != nil {
			m.sendEvent(EventFailed, err)
			if !m.shouldRestart(err) {
				return
			}
			if !m.waitAndRestart() {
				return
			}
			continue
		}

		m.sendEvent(EventStarted, nil)

		select {
		case <-m.ctx.Done():
			m.process.Stop(30 * time.Second)
			return
		case <-m.process.Wait():
			exitCode := m.process.ExitCode()

			if exitCode == 0 {
				m.sendEvent(EventStopped, nil)
			} else {
				m.sendEvent(EventFailed, fmt.Errorf("exit code: %d", exitCode))
			}

			if !m.shouldRestartExitCode(exitCode) {
				return
			}

			if !m.waitAndRestart() {
				return
			}
		}
	}
}

// shouldRestart determines if the process should be restarted after an error.
func (m *Manager) shouldRestart(err error) bool {
	switch m.config.Restart.Policy {
	case config.RestartAlways:
		return m.process.Restarts() < m.config.Restart.MaxRetries
	case config.RestartOnFailure:
		return m.process.Restarts() < m.config.Restart.MaxRetries
	case config.RestartNever:
		return false
	case config.RestartUnless:
		return true // Only manual stop prevents restart
	default:
		return false
	}
}

// shouldRestartExitCode determines if the process should be restarted based on exit code.
func (m *Manager) shouldRestartExitCode(exitCode int) bool {
	switch m.config.Restart.Policy {
	case config.RestartAlways:
		return m.process.Restarts() < m.config.Restart.MaxRetries
	case config.RestartOnFailure:
		if exitCode == 0 {
			return false
		}
		return m.process.Restarts() < m.config.Restart.MaxRetries
	case config.RestartNever:
		return false
	case config.RestartUnless:
		return true
	default:
		return false
	}
}

// waitAndRestart waits the configured delay and prepares for restart.
func (m *Manager) waitAndRestart() bool {
	m.process.IncrementRestarts()
	m.sendEvent(EventRestarting, nil)

	delay := m.calculateDelay()

	select {
	case <-m.ctx.Done():
		return false
	case <-time.After(delay):
		return true
	}
}

// calculateDelay calculates the restart delay with exponential backoff.
func (m *Manager) calculateDelay() time.Duration {
	baseDelay := m.config.Restart.Delay.Duration()
	maxDelay := m.config.Restart.DelayMax.Duration()

	if maxDelay == 0 {
		maxDelay = baseDelay * 10
	}

	// Exponential backoff: delay * 2^restarts
	delay := baseDelay * time.Duration(1<<uint(m.process.Restarts()))

	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// Stop stops the managed process.
func (m *Manager) Stop() error {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
	}
	return m.process.Stop(30 * time.Second)
}

// Reload reloads the process configuration.
func (m *Manager) Reload() error {
	return m.process.Reload()
}

// sendEvent sends a lifecycle event.
func (m *Manager) sendEvent(eventType EventType, err error) {
	event := Event{
		Type:      eventType,
		Process:   m.config.Name,
		PID:       m.process.PID(),
		ExitCode:  m.process.ExitCode(),
		Timestamp: time.Now(),
		Error:     err,
	}

	select {
	case m.events <- event:
	default:
		// Drop event if channel is full
	}
}

// Status represents the status of a managed process.
type Status struct {
	Name     string
	State    State
	PID      int
	Uptime   time.Duration
	Restarts int
	ExitCode int
}

// Status returns the current status of the process.
func (m *Manager) Status() Status {
	return Status{
		Name:     m.config.Name,
		State:    m.process.State(),
		PID:      m.process.PID(),
		Uptime:   m.process.Uptime(),
		Restarts: m.process.Restarts(),
		ExitCode: m.process.ExitCode(),
	}
}
