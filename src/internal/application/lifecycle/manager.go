// Package lifecycle provides the application service for managing process lifecycle.
package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kodflow/daemon/internal/domain/config"
	domain "github.com/kodflow/daemon/internal/domain/process"
)

// Manager configuration constants.
const (
	// eventBufferSize defines the channel buffer size for lifecycle events.
	eventBufferSize int = 16
	// defaultStopTimeout defines the default timeout for stopping processes.
	defaultStopTimeout time.Duration = 30 * time.Second
)

// Manager manages the lifecycle of a single process with restart policies.
//
// Manager coordinates process execution, monitors exit status, and applies
// restart policies including exponential backoff. It emits lifecycle events
// for monitoring and integrates with the domain executor abstraction.
type Manager struct {
	mu       sync.RWMutex
	config   *config.ServiceConfig
	executor domain.Executor
	tracker  *domain.RestartTracker
	events   chan domain.Event
	ctx      context.Context
	cancel   context.CancelFunc
	running  bool

	// Current process state
	pid       int
	state     domain.State
	exitCode  int
	startTime time.Time
	restarts  int
	waitCh    <-chan domain.ExitResult
}

// NewManager creates a new process lifecycle manager.
//
// Params:
//   - cfg: the service configuration defining command, restart policy, etc.
//   - executor: the domain executor for process operations.
//
// Returns:
//   - *Manager: a new manager instance ready to start the process.
func NewManager(cfg *config.ServiceConfig, executor domain.Executor) *Manager {
	// Return a new Manager with initialized fields.
	return &Manager{
		config:   cfg,
		executor: executor,
		tracker:  domain.NewRestartTracker(&cfg.Restart),
		events:   make(chan domain.Event, eventBufferSize),
		state:    domain.StateStopped,
	}
}

// Events returns the event channel for monitoring.
//
// Returns:
//   - <-chan domain.Event: read-only channel for lifecycle events.
func (m *Manager) Events() <-chan domain.Event {
	// Return the events channel for external monitoring.
	return m.events
}

// Name returns the service name.
//
// Returns:
//   - string: the service name from configuration.
//
// TODO(test): Add test coverage for Name method.
func (m *Manager) Name() string {
	// Return the service name from config.
	return m.config.Name
}

// State returns the current process state.
//
// Returns:
//   - domain.State: the current state of the managed process.
func (m *Manager) State() domain.State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return the current state under read lock.
	return m.state
}

// PID returns the current process PID.
//
// Returns:
//   - int: the process ID or 0 if not running.
func (m *Manager) PID() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return the current PID under read lock.
	return m.pid
}

// Uptime returns the process uptime in seconds.
//
// Returns:
//   - int64: the uptime in seconds or 0 if not running.
func (m *Manager) Uptime() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Check if process is not running.
	if m.state != domain.StateRunning {
		// Return zero uptime when not running.
		return 0
	}
	// Calculate and return uptime in seconds.
	return int64(time.Since(m.startTime).Seconds())
}

// Start starts the managed process with automatic restart handling.
//
// Returns:
//   - error: ErrAlreadyRunning if manager is already running, nil otherwise.
//
// Goroutine lifecycle:
//   - Spawns a goroutine that runs until Stop is called or context is cancelled.
//   - The goroutine handles process lifecycle including restarts based on policy.
//   - Use Stop() to terminate the goroutine and cleanup resources.
func (m *Manager) Start() error {
	m.mu.Lock()
	// Check if manager is already running.
	if m.running {
		m.mu.Unlock()
		// Return error wrapped with already running context.
		return fmt.Errorf("manager already running: %w", domain.ErrAlreadyRunning)
	}
	m.running = true
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.mu.Unlock()

	go m.run()
	// Return nil on successful start.
	return nil
}

// run is the main loop that manages the process lifecycle.
func (m *Manager) run() {
	defer func() {
		m.mu.Lock()
		m.running = false
		m.mu.Unlock()
	}()

	// Check if service is configured as oneshot.
	if m.config.Oneshot {
		m.runOnce()
		// Return after oneshot execution completes.
		return
	}

	m.runWithRestart()
}

// runOnce runs the process once without restart.
func (m *Manager) runOnce() {
	// Attempt to start the process.
	if err := m.startProcess(); err != nil {
		m.sendEvent(domain.EventFailed, err)
		// Return early on start failure.
		return
	}

	m.sendEvent(domain.EventStarted, nil)
	result := <-m.waitCh

	// Check if process exited with non-zero code.
	if result.Code != 0 {
		m.sendEvent(domain.EventFailed, fmt.Errorf("exit code %d: %w", result.Code, domain.ErrProcessFailed))
		// Return after reporting failure.
		return
	}

	m.sendEvent(domain.EventStopped, nil)
}

// runWithRestart runs the process with automatic restart based on policy.
func (m *Manager) runWithRestart() {
	// Loop continuously for restart handling.
	for {
		// Check if context was cancelled.
		if m.isContextCancelled() {
			// Return to exit restart loop.
			return
		}

		// Attempt to start the process.
		if !m.tryStartProcess() {
			// Return when start fails and no restart.
			return
		}

		// Wait for process exit or shutdown.
		if m.waitForProcessOrShutdown() {
			// Return when shutdown requested.
			return
		}
	}
}

// isContextCancelled checks if the context has been cancelled.
//
// Returns:
//   - bool: true if context is cancelled, false otherwise.
func (m *Manager) isContextCancelled() bool {
	select {
	// Check if context done channel is closed.
	case <-m.ctx.Done():
		// Return true when context is cancelled.
		return true
	// Default case for non-blocking check.
	default:
		// Return false when context is still active.
		return false
	}
}

// tryStartProcess attempts to start the process.
//
// Returns:
//   - bool: true if process started or restart scheduled, false to stop.
func (m *Manager) tryStartProcess() bool {
	// Attempt to start the underlying process.
	if err := m.startProcess(); err != nil {
		m.sendEvent(domain.EventFailed, err)
		// Check if restart policy allows retry.
		if !m.tracker.ShouldRestart(-1) {
			// Check if restarts were exhausted.
			if m.tracker.IsExhausted() {
				m.sendEvent(domain.EventExhausted, fmt.Errorf("max restarts (%d) exceeded: %w", m.tracker.Attempts(), domain.ErrMaxRetriesExceeded))
			}
			// Return false when no more restarts allowed.
			return false
		}
		// Wait and schedule restart.
		return m.waitAndRestart()
	}

	m.sendEvent(domain.EventStarted, nil)
	// Return true on successful start.
	return true
}

// startProcess starts the underlying process.
//
// Returns:
//   - error: nil on success, error on failure.
func (m *Manager) startProcess() error {
	m.mu.Lock()
	m.state = domain.StateStarting
	m.mu.Unlock()

	spec := domain.NewSpec(domain.SpecParams{
		Command: m.config.Command,
		Args:    m.config.Args,
		Dir:     m.config.WorkingDirectory,
		Env:     m.config.Environment,
		User:    m.config.User,
		Group:   m.config.Group,
	})

	pid, wait, err := m.executor.Start(m.ctx, spec)
	// Check if start failed.
	if err != nil {
		m.mu.Lock()
		m.state = domain.StateFailed
		m.mu.Unlock()
		// Return the start error.
		return err
	}

	m.mu.Lock()
	m.pid = pid
	m.waitCh = wait
	m.startTime = time.Now()
	m.state = domain.StateRunning
	m.mu.Unlock()

	// Return nil on successful process start.
	return nil
}

// waitForProcessOrShutdown waits for process exit or shutdown signal.
// Stop errors during shutdown are intentionally discarded (best-effort cleanup).
// The process will be terminated when the parent exits regardless.
// A Failed event has already been sent if the process crashed.
//
// Returns:
//   - bool: true if shutdown requested, false if process exited.
func (m *Manager) waitForProcessOrShutdown() bool {
	select {
	// Handle context cancellation (shutdown).
	case <-m.ctx.Done():
		m.mu.Lock()
		pid := m.pid
		m.mu.Unlock()
		// Stop process if running (best-effort, errors discarded during shutdown).
		if pid > 0 {
			_ = m.executor.Stop(pid, defaultStopTimeout)
		}
		// Return true to indicate shutdown.
		return true
	// Handle process exit result.
	case result := <-m.waitCh:
		// Process the exit result.
		return m.handleProcessExit(result)
	}
}

// handleProcessExit handles process exit and determines if restart should occur.
//
// Params:
//   - result: the exit result containing exit code and error.
//
// Returns:
//   - bool: true if no restart should occur, false to continue restart loop.
func (m *Manager) handleProcessExit(result domain.ExitResult) bool {
	m.updateStateAfterExit(result)
	m.sendExitEvent(result)

	// Reset backoff counter if process ran stably for the configured window.
	uptime := m.calculateUptime()
	m.tracker.MaybeReset(uptime)

	// Check if restart policy allows restart.
	if !m.tracker.ShouldRestart(result.Code) {
		m.handleExhaustedRestarts(result)
		// Return true to stop restart loop.
		return true
	}

	// Attempt to wait and restart.
	if !m.waitAndRestart() {
		// Return true when restart cancelled.
		return true
	}
	// Return false to continue restart loop.
	return false
}

// updateStateAfterExit updates the manager state after process exit.
//
// Params:
//   - result: the exit result containing exit code.
func (m *Manager) updateStateAfterExit(result domain.ExitResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.exitCode = result.Code
	m.pid = 0

	// Check if process exited successfully.
	if result.Code == 0 {
		m.state = domain.StateStopped
	} else {
		// Set failed state for non-zero exit.
		m.state = domain.StateFailed
	}
}

// sendExitEvent sends the appropriate event based on exit code.
//
// Params:
//   - result: the exit result containing exit code.
func (m *Manager) sendExitEvent(result domain.ExitResult) {
	// Check exit code for event type.
	if result.Code == 0 {
		m.sendEvent(domain.EventStopped, nil)
	} else {
		// Send failed event with exit code error.
		m.sendEvent(domain.EventFailed, fmt.Errorf("exit code %d: %w", result.Code, domain.ErrProcessFailed))
	}
}

// calculateUptime returns the process uptime before exit.
//
// Returns:
//   - time.Duration: the process uptime.
func (m *Manager) calculateUptime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Calculate uptime based on start time.
	return time.Since(m.startTime)
}

// handleExhaustedRestarts checks if restarts are exhausted and emits event if needed.
//
// Params:
//   - result: the exit result containing exit code.
func (m *Manager) handleExhaustedRestarts(result domain.ExitResult) {
	// Check if restarts were exhausted.
	// For RestartAlways: exhausted if attempts >= max (regardless of exit code).
	// For RestartOnFailure: exhausted only if exit code != 0 and attempts >= max.
	if !m.tracker.IsExhausted() {
		// Return early when not exhausted.
		return
	}

	// Determine if exhausted event should be emitted based on restart policy.
	if m.shouldEmitExhaustedEvent(result.Code) {
		m.sendEvent(domain.EventExhausted, fmt.Errorf("max restarts (%d) exceeded: %w", m.tracker.Attempts(), domain.ErrMaxRetriesExceeded))
	}
}

// shouldEmitExhaustedEvent determines if exhausted event should be emitted.
//
// Params:
//   - exitCode: the process exit code.
//
// Returns:
//   - bool: true if exhausted event should be emitted, false otherwise.
func (m *Manager) shouldEmitExhaustedEvent(exitCode int) bool {
	policy := m.config.Restart.Policy

	// Determine exhausted event emission based on restart policy.
	switch policy {
	// Always emit exhausted event for RestartAlways policy.
	case config.RestartAlways:
		// Always emit when exhausted, even for clean exits (e.g., killed by health check).
		return true
	// Only emit exhausted event if process failed for RestartOnFailure policy.
	case config.RestartOnFailure:
		// Only emit if the process actually failed.
		return exitCode != 0
	// Never emit exhausted event for RestartNever policy.
	case config.RestartNever:
		// Never restarts, so exhausted doesn't apply.
		return false
	// Never emit exhausted event for RestartUnless policy.
	case config.RestartUnless:
		// Always restarts (no max), so exhausted doesn't apply.
		return false
	}

	// Return false for unknown policies.
	return false
}

// waitAndRestart waits the configured delay and prepares for restart.
//
// Returns:
//   - bool: true if restart should proceed, false if cancelled.
func (m *Manager) waitAndRestart() bool {
	m.tracker.RecordAttempt()
	m.mu.Lock()
	m.restarts++
	m.mu.Unlock()

	m.sendEvent(domain.EventRestarting, nil)

	delay := m.tracker.NextDelay()

	// Use NewTimer instead of time.After to allow proper cleanup.
	// time.After creates a timer that won't be GC'd until it fires.
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	// Handle context cancellation during delay.
	case <-m.ctx.Done():
		// Return false to cancel restart.
		return false
	// Wait for delay duration.
	case <-timer.C:
		// Return true to proceed with restart.
		return true
	}
}

// Stop stops the managed process.
//
// Returns:
//   - error: nil on success, error from executor on failure.
func (m *Manager) Stop() error {
	m.mu.Lock()
	// Check if manager is not running.
	if !m.running {
		m.mu.Unlock()
		// Return nil when already stopped.
		return nil
	}
	pid := m.pid
	m.mu.Unlock()

	// Cancel the context if set.
	if m.cancel != nil {
		m.cancel()
	}

	// Stop the process if PID is valid.
	if pid > 0 {
		// Return the result of executor stop.
		return m.executor.Stop(pid, defaultStopTimeout)
	}
	// Return nil when no process to stop.
	return nil
}

// Reload reloads the process (sends SIGHUP).
//
// Returns:
//   - error: ErrNotRunning if no process, error from signal otherwise.
func (m *Manager) Reload() error {
	m.mu.RLock()
	pid := m.pid
	m.mu.RUnlock()

	// Check if process is not running.
	if pid == 0 {
		// Return error when not running.
		return domain.ErrNotRunning
	}

	// Send SIGHUP signal to process.
	return m.executor.Signal(pid, signalHUP)
}

// sendEvent sends a lifecycle event.
//
// Params:
//   - eventType: the type of lifecycle event.
//   - err: optional error associated with the event.
func (m *Manager) sendEvent(eventType domain.EventType, err error) {
	m.mu.RLock()
	event := domain.NewEvent(eventType, m.config.Name, m.pid, m.exitCode, err)
	m.mu.RUnlock()

	select {
	// Attempt to send event to channel.
	case m.events <- event:
	// Drop event if channel is full.
	default:
	}
}

// Status returns the current status of the process.
//
// Returns:
//   - domain.Status: the complete status information.
func (m *Manager) Status() domain.Status {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return the current status snapshot.
	return domain.Status{
		Name:     m.config.Name,
		State:    m.state,
		PID:      m.pid,
		Uptime:   time.Since(m.startTime),
		Restarts: m.restarts,
		ExitCode: m.exitCode,
	}
}

// RestartOnHealthFailure triggers a process restart due to health probe failure.
// This implements the Kubernetes liveness probe pattern: when health probes
// fail consecutively beyond the failure threshold, the process is killed
// and will be restarted by the normal restart policy.
//
// Params:
//   - reason: description of why the health check failed.
//
// Returns:
//   - error: ErrNotRunning if no process, error from executor on stop failure.
//
// TODO(test): Add test coverage for RestartOnHealthFailure method.
func (m *Manager) RestartOnHealthFailure(reason string) error {
	m.mu.Lock()
	pid := m.pid
	running := m.running
	m.mu.Unlock()

	// Check if manager is not running.
	if !running {
		// Return error when manager is not active.
		return domain.ErrNotRunning
	}

	// Check if process is not running.
	if pid == 0 {
		// Return error when no process to restart.
		return domain.ErrNotRunning
	}

	// Send unhealthy event before stopping process.
	m.sendEvent(domain.EventUnhealthy, fmt.Errorf("%s: %w", reason, domain.ErrHealthProbeFailed))

	// Stop the process; restart loop will handle restart based on policy.
	return m.executor.Stop(pid, defaultStopTimeout)
}
