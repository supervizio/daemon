// Package supervisor provides the application service for orchestrating multiple services.
// It manages the lifecycle of services including start, stop, restart, and reload operations.
package supervisor

import (
	"context"
	"fmt"
	"sync"

	appconfig "github.com/kodflow/daemon/internal/application/config"
	appprocess "github.com/kodflow/daemon/internal/application/process"
	"github.com/kodflow/daemon/internal/domain/kernel"
	domain "github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/domain/service"
)

// State represents the supervisor state.
// It defines the current operational status of the supervisor.
type State int

// Supervisor state constants.
const (
	StateStopped State = iota
	StateStarting
	StateRunning
	StateStopping
)

// Errors for supervisor operations.
var (
	// ErrAlreadyRunning is returned when the supervisor is already running.
	ErrAlreadyRunning error = fmt.Errorf("supervisor already running")
	// ErrNotRunning is returned when the supervisor is not running.
	ErrNotRunning error = fmt.Errorf("supervisor not running")
	// ErrServiceNotFound is returned when a service is not found.
	ErrServiceNotFound error = fmt.Errorf("service not found")
)

// EventHandler is a callback function for process events.
// It is called when a service emits a lifecycle event.
type EventHandler func(serviceName string, event *domain.Event)

// Supervisor manages multiple services and their lifecycle.
// It coordinates starting, stopping, and monitoring of all configured services.
type Supervisor struct {
	// mu is the mutex for thread-safe access.
	mu sync.RWMutex
	// config is the service configuration.
	config *service.Config
	// loader is the configuration loader.
	loader appconfig.Loader
	// executor is the process executor.
	executor domain.Executor
	// managers is the map of service managers.
	managers map[string]*appprocess.Manager
	// reaper is the zombie process reaper.
	reaper kernel.ZombieReaper
	// state is the current supervisor state.
	state State
	// ctx is the context for cancellation.
	ctx context.Context
	// cancel is the cancel function.
	cancel context.CancelFunc
	// wg is the wait group for goroutines.
	wg sync.WaitGroup
	// eventHandler is the optional callback for events.
	eventHandler EventHandler
	// stats holds per-service statistics.
	stats map[string]*ServiceStats
}

// NewSupervisor creates a new supervisor from configuration.
//
// Params:
//   - cfg: the service configuration.
//   - loader: the configuration loader for reloading.
//   - executor: the process executor.
//   - reaper: the zombie process reaper.
//
// Returns:
//   - *Supervisor: the new supervisor instance.
//   - error: an error if configuration is invalid.
func NewSupervisor(cfg *service.Config, loader appconfig.Loader, executor domain.Executor, reaper kernel.ZombieReaper) (*Supervisor, error) {
	// Validate the configuration before creating the supervisor.
	if err := cfg.Validate(); err != nil {
		// Return nil supervisor and validation error.
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	s := &Supervisor{
		config:   cfg,
		loader:   loader,
		executor: executor,
		managers: make(map[string]*appprocess.Manager, len(cfg.Services)),
		reaper:   reaper,
		state:    StateStopped,
		stats:    make(map[string]*ServiceStats, len(cfg.Services)),
	}

	// Create a manager for each configured service.
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		s.managers[svc.Name] = appprocess.NewManager(svc, executor)
		s.stats[svc.Name] = NewServiceStats()
	}

	// Return the configured supervisor.
	return s, nil
}

// Start starts all managed services.
//
// Params:
//   - ctx: the context for cancellation.
//
// Returns:
//   - error: an error if any service fails to start.
//
// Goroutine lifecycle:
//   - Spawns one goroutine per service for monitoring.
//   - Goroutines run until Stop is called or context is cancelled.
//   - Use Stop() to terminate all monitoring goroutines.
func (s *Supervisor) Start(ctx context.Context) error {
	s.mu.Lock()
	// Check if the supervisor is already running.
	if s.state != StateStopped {
		s.mu.Unlock()
		// Return error for already running state.
		return ErrAlreadyRunning
	}
	s.state = StateStarting
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.mu.Unlock()

	// Start the zombie reaper if available.
	if s.reaper != nil {
		s.reaper.Start()
	}

	var startErr error
	// Start each managed service.
	for name, mgr := range s.managers {
		// Attempt to start the service.
		if err := mgr.Start(); err != nil {
			startErr = fmt.Errorf("failed to start service %s: %w", name, err)
			break
		}
	}

	// Handle startup failure by stopping all services.
	if startErr != nil {
		s.stopAll()
		s.mu.Lock()
		s.state = StateStopped
		s.mu.Unlock()
		// Return the startup error.
		return startErr
	}

	// Start monitoring goroutines for each service.
	for name, mgr := range s.managers {
		s.wg.Add(1)
		go s.monitorService(name, mgr)
	}

	s.mu.Lock()
	s.state = StateRunning
	s.mu.Unlock()

	// Return nil on successful start.
	return nil
}

// Stop gracefully stops all managed services.
//
// Returns:
//   - error: always nil, provided for interface compatibility.
func (s *Supervisor) Stop() error {
	s.mu.Lock()
	// Check if the supervisor is running.
	if s.state != StateRunning {
		s.mu.Unlock()
		// Return nil when not running.
		return nil
	}
	s.state = StateStopping
	s.cancel()
	s.mu.Unlock()

	s.stopAll()
	s.wg.Wait()

	// Stop the zombie reaper if available.
	if s.reaper != nil {
		s.reaper.Stop()
	}

	s.mu.Lock()
	s.state = StateStopped
	s.mu.Unlock()

	// Return nil on successful stop.
	return nil
}

// stopAll stops all managed services concurrently.
//
// Goroutine lifecycle:
//   - Spawns one goroutine per service for concurrent stop.
//   - All goroutines complete when their respective services stop.
//   - Method blocks until all goroutines complete via WaitGroup.
func (s *Supervisor) stopAll() {
	var wg sync.WaitGroup
	// Iterate through all managers.
	for _, mgr := range s.managers {
		m := mgr
		// Stop each manager in a goroutine using Go 1.25's wg.Go().
		wg.Go(func() {
			_ = m.Stop()
		})
	}
	wg.Wait()
}

// Reload reloads the configuration and restarts changed services.
//
// Returns:
//   - error: an error if the reload fails.
func (s *Supervisor) Reload() error {
	// Check state and get config path without holding lock during I/O.
	s.mu.RLock()
	state := s.state
	configPath := s.config.ConfigPath
	s.mu.RUnlock()

	// Check if the supervisor is running.
	if state != StateRunning {
		// Return error when not running.
		return ErrNotRunning
	}

	// Load configuration without holding lock (I/O operation).
	newCfg, err := s.loader.Load(configPath)
	// Handle configuration load error.
	if err != nil {
		// Return wrapped error on load failure.
		return fmt.Errorf("failed to reload config: %w", err)
	}

	// Acquire write lock for state updates.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-check state after acquiring lock (may have changed).
	if s.state != StateRunning {
		// Return error when no longer running.
		return ErrNotRunning
	}

	s.updateServices(newCfg)
	s.removeDeletedServices(newCfg)

	s.config = newCfg
	// Return nil on successful reload.
	return nil
}

// updateServices updates or adds managers for services in the new configuration.
//
// Params:
//   - newCfg: the new service configuration.
//
// Goroutine lifecycle:
//   - May spawn new goroutines for monitoring newly added services.
//   - Goroutines run until Stop is called or context is cancelled.
//   - Use Stop() to terminate all monitoring goroutines.
func (s *Supervisor) updateServices(newCfg *service.Config) {
	// Iterate through all services in the new configuration.
	for i := range newCfg.Services {
		svc := &newCfg.Services[i]
		// Check if the service already exists.
		if mgr, exists := s.managers[svc.Name]; exists {
			_ = mgr.Stop()
			s.managers[svc.Name] = appprocess.NewManager(svc, s.executor)
			_ = s.managers[svc.Name].Start()
		} else {
			// Create and start a new manager for the new service.
			s.managers[svc.Name] = appprocess.NewManager(svc, s.executor)
			_ = s.managers[svc.Name].Start()
			s.wg.Add(1)
			go s.monitorService(svc.Name, s.managers[svc.Name])
		}
	}
}

// removeDeletedServices removes managers for services no longer in configuration.
//
// Params:
//   - newCfg: the new service configuration.
func (s *Supervisor) removeDeletedServices(newCfg *service.Config) {
	newServices := make(map[string]bool, len(newCfg.Services))
	// Build a map of services in the new configuration.
	for i := range newCfg.Services {
		newServices[newCfg.Services[i].Name] = true
	}
	// Remove services that are no longer in the configuration.
	for name, mgr := range s.managers {
		// Check if the service should be removed.
		if !newServices[name] {
			_ = mgr.Stop()
			delete(s.managers, name)
		}
	}
}

// monitorService monitors a service for events.
//
// Params:
//   - name: the service name.
//   - mgr: the process manager interface.
func (s *Supervisor) monitorService(name string, mgr Eventser) {
	defer s.wg.Done()

	events := mgr.Events()
	// Loop until context is cancelled or events channel is closed.
	for {
		// Select between context cancellation and events.
		select {
		case <-s.ctx.Done():
			// Return when context is cancelled.
			return
		case event, ok := <-events:
			// Check if the events channel is closed.
			if !ok {
				// Return when channel is closed.
				return
			}
			// Handle the event: update stats and call handler.
			s.handleEvent(name, &event)
		}
	}
}

// handleEvent processes a service event.
// It updates statistics and calls the optional event handler.
//
// Params:
//   - name: the service name.
//   - event: the process event.
func (s *Supervisor) handleEvent(name string, event *domain.Event) {
	s.mu.Lock()
	// Get or create stats for this service.
	stats, ok := s.stats[name]
	// Create new stats entry if service not tracked yet.
	if !ok {
		stats = NewServiceStats()
		s.stats[name] = stats
	}

	// Update statistics based on event type.
	switch event.Type {
	// Started: service successfully started.
	case domain.EventStarted:
		stats.StartCount++
	// Stopped: service stopped normally.
	case domain.EventStopped:
		stats.StopCount++
	// Failed: service exited with error.
	case domain.EventFailed:
		stats.FailCount++
	// Restarting: service is restarting after exit.
	case domain.EventRestarting:
		stats.RestartCount++
	// Health events: tracked by health monitor, not stats.
	case domain.EventHealthy, domain.EventUnhealthy:
		// Health events are tracked by the health monitor, not stats.
	}
	s.mu.Unlock()

	// Call user event handler if registered.
	if s.eventHandler != nil {
		s.eventHandler(name, event)
	}
}

// SetEventHandler sets the callback for process events.
//
// Params:
//   - handler: the callback function to invoke on events.
func (s *Supervisor) SetEventHandler(handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.eventHandler = handler
}

// Stats returns statistics for a specific service.
//
// Params:
//   - name: the service name.
//
// Returns:
//   - *ServiceStats: the service statistics, or nil if not found.
func (s *Supervisor) Stats(name string) *ServiceStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a copy to avoid race conditions.
	if stats, ok := s.stats[name]; ok {
		statsCopy := *stats
		// Return copy of stats for the requested service.
		return &statsCopy
	}
	// Service not found, return nil.
	return nil
}

// AllStats returns statistics for all services.
//
// Returns:
//   - map[string]*ServiceStats: a copy of all service statistics.
func (s *Supervisor) AllStats() map[string]*ServiceStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a copy of all stats.
	result := make(map[string]*ServiceStats, len(s.stats))
	// Copy each service's stats to avoid race conditions.
	for name, stats := range s.stats {
		statsCopy := *stats
		result[name] = &statsCopy
	}
	// Return the copied stats map.
	return result
}

// Eventser defines the interface for monitoring services.
// This interface is used to abstract the manager for testing.
type Eventser interface {
	// Events returns the event channel for monitoring.
	Events() <-chan domain.Event
}

// State returns the current supervisor state.
//
// Returns:
//   - State: the current state.
func (s *Supervisor) State() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return the current state.
	return s.state
}

// Services returns information about all managed services.
//
// Returns:
//   - map[string]ServiceInfo: a map of service names to their information.
func (s *Supervisor) Services() map[string]ServiceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := make(map[string]ServiceInfo, len(s.managers))
	// Collect information from each manager.
	for name, mgr := range s.managers {
		info[name] = ServiceInfo{
			Name:   name,
			State:  mgr.State(),
			PID:    mgr.PID(),
			Uptime: mgr.Uptime(),
		}
	}
	// Return the service information map.
	return info
}

// Service returns a specific service manager.
//
// Params:
//   - name: the service name.
//
// Returns:
//   - *appprocess.Manager: the manager if found.
//   - bool: true if the service was found.
func (s *Supervisor) Service(name string) (*appprocess.Manager, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	mgr, ok := s.managers[name]
	// Return the manager and found status.
	return mgr, ok
}

// StartService starts a specific service.
//
// Params:
//   - name: the service name.
//
// Returns:
//   - error: an error if the service is not found or fails to start.
func (s *Supervisor) StartService(name string) error {
	s.mu.RLock()
	mgr, ok := s.managers[name]
	s.mu.RUnlock()

	// Check if the service exists.
	if !ok {
		// Return error for missing service.
		return fmt.Errorf("%w: %s", ErrServiceNotFound, name)
	}
	// Return the result of starting the service.
	return mgr.Start()
}

// StopService stops a specific service.
//
// Params:
//   - name: the service name.
//
// Returns:
//   - error: an error if the service is not found or fails to stop.
func (s *Supervisor) StopService(name string) error {
	s.mu.RLock()
	mgr, ok := s.managers[name]
	s.mu.RUnlock()

	// Check if the service exists.
	if !ok {
		// Return error for missing service.
		return fmt.Errorf("%w: %s", ErrServiceNotFound, name)
	}
	// Return the result of stopping the service.
	return mgr.Stop()
}

// RestartService restarts a specific service.
//
// Params:
//   - name: the service name.
//
// Returns:
//   - error: an error if the service is not found or fails to restart.
func (s *Supervisor) RestartService(name string) error {
	s.mu.RLock()
	mgr, ok := s.managers[name]
	s.mu.RUnlock()

	// Check if the service exists.
	if !ok {
		// Return error for missing service.
		return fmt.Errorf("%w: %s", ErrServiceNotFound, name)
	}

	// Stop the service first.
	if err := mgr.Stop(); err != nil {
		// Return stop error if failed.
		return err
	}
	// Return the result of starting the service.
	return mgr.Start()
}
