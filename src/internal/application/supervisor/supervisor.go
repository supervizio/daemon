// Package supervisor provides the application service for orchestrating multiple services.
// It manages the lifecycle of services including start, stop, restart, and reload operations.
package supervisor

import (
	"context"
	"fmt"
	"log"
	"sync"

	appconfig "github.com/kodflow/daemon/internal/application/config"
	apphealth "github.com/kodflow/daemon/internal/application/health"
	applifecycle "github.com/kodflow/daemon/internal/application/lifecycle"
	domainconfig "github.com/kodflow/daemon/internal/domain/config"
	domainhealth "github.com/kodflow/daemon/internal/domain/health"
	domainlifecycle "github.com/kodflow/daemon/internal/domain/lifecycle"
	"github.com/kodflow/daemon/internal/domain/listener"
	domain "github.com/kodflow/daemon/internal/domain/process"
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

// ErrorHandler is a callback function for handling non-fatal errors.
// These errors occur in recovery/cleanup paths where the supervisor
// continues operating despite the error. Examples include:
//   - Errors stopping services during shutdown (best-effort cleanup)
//   - Errors restarting services during configuration reload
//
// If no handler is set, these errors are silently discarded.
// Setting a handler allows logging or monitoring of these conditions.
type ErrorHandler func(operation string, serviceName string, err error)

// AddListenerWithBindinger defines the minimal interface for adding listeners to health monitors.
// This satisfies KTN-API-MINIF by accepting only the methods actually used.
// Named following ktn-linter requirement: single-method interfaces use methodName+er pattern.
type AddListenerWithBindinger interface {
	AddListenerWithBinding(l *listener.Listener, binding *apphealth.ProbeBinding) error
}

// Supervisor manages multiple services and their lifecycle.
// It coordinates starting, stopping, and monitoring of all configured services.
type Supervisor struct {
	// mu is the mutex for thread-safe access.
	mu sync.RWMutex
	// config is the service configuration.
	config *domainconfig.Config
	// loader is the configuration loader.
	loader appconfig.Loader
	// executor is the process execution.
	executor domain.Executor
	// managers is the map of service managers.
	managers map[string]*applifecycle.Manager
	// healthMonitors is the map of health monitors per service.
	healthMonitors map[string]*apphealth.ProbeMonitor
	// proberFactory creates health probers.
	proberFactory apphealth.Creator
	// reaper is the zombie process reaper (domain port).
	reaper domainlifecycle.Reaper
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
	// errorHandler is the optional callback for non-fatal errors in recovery paths.
	errorHandler ErrorHandler
	// stats holds per-service statistics.
	stats map[string]*ServiceStats
}

// NewSupervisor creates a new supervisor from configuration.
//
// Params:
//   - cfg: the service configuration.
//   - loader: the configuration loader for reloading.
//   - executor: the process execution.
//   - reaper: the zombie process reaper.
//
// Returns:
//   - *Supervisor: the new supervisor instance.
//   - error: an error if configuration is invalid.
func NewSupervisor(cfg *domainconfig.Config, loader appconfig.Loader, executor domain.Executor, reaper domainlifecycle.Reaper) (*Supervisor, error) {
	// Validate the configuration before creating the supervisor.
	if err := cfg.Validate(); err != nil {
		// Return nil supervisor and validation error.
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	s := &Supervisor{
		config:         cfg,
		loader:         loader,
		executor:       executor,
		managers:       make(map[string]*applifecycle.Manager, len(cfg.Services)),
		healthMonitors: make(map[string]*apphealth.ProbeMonitor, len(cfg.Services)),
		reaper:         reaper,
		state:          StateStopped,
		stats:          make(map[string]*ServiceStats, len(cfg.Services)),
	}

	// Create a manager for each configured service.
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		s.managers[svc.Name] = applifecycle.NewManager(svc, executor)
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
	// Initialize supervisor state and context.
	if err := s.initializeStart(ctx); err != nil {
		// Return initialization error.
		return err
	}

	// Start zombie reaper if configured.
	s.startReaper()

	// Start all managed services.
	if err := s.startAllServices(); err != nil {
		// Return service start error.
		return err
	}

	// Start monitoring goroutines for each service.
	s.startMonitoringGoroutines()

	// Create and start health monitors.
	s.startHealthMonitors()

	// Mark supervisor as running.
	s.mu.Lock()
	s.state = StateRunning
	s.mu.Unlock()

	// Return nil on successful start.
	return nil
}

// initializeStart sets up the supervisor state for starting.
//
// Params:
//   - ctx: the parent context.
//
// Returns:
//   - error: ErrAlreadyRunning if supervisor is already running.
func (s *Supervisor) initializeStart(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already running.
	if s.state != StateStopped {
		// Return error when supervisor is already running.
		return ErrAlreadyRunning
	}

	// Set state and create cancellable context.
	s.state = StateStarting
	s.ctx, s.cancel = context.WithCancel(ctx)
	// Return nil on successful initialization.
	return nil
}

// startReaper starts the zombie reaper if configured.
func (s *Supervisor) startReaper() {
	// Skip if reaper is not configured.
	if s.reaper == nil {
		// Return early when reaper is nil.
		return
	}
	s.reaper.Start()
}

// startAllServices starts all managed services.
//
// Returns:
//   - error: first error encountered, or nil on success.
func (s *Supervisor) startAllServices() error {
	// Iterate through all managed services.
	for name, mgr := range s.managers {
		// Check if service fails to start.
		if err := mgr.Start(); err != nil {
			// Handle startup failure by stopping all services.
			s.stopAll()
			s.mu.Lock()
			s.state = StateStopped
			s.mu.Unlock()
			// Return wrapped start error.
			return fmt.Errorf("failed to start service %s: %w", name, err)
		}
	}
	// Return nil when all services started successfully.
	return nil
}

// startMonitoringGoroutines spawns monitoring goroutines for each service.
func (s *Supervisor) startMonitoringGoroutines() {
	// Start monitoring goroutine for each service.
	for name, mgr := range s.managers {
		s.wg.Add(1)
		go s.monitorService(name, mgr)
	}
}

// startHealthMonitors creates and starts health monitors for services with probes.
func (s *Supervisor) startHealthMonitors() {
	// Create and start health monitor for each service.
	for i := range s.config.Services {
		svc := &s.config.Services[i]
		monitor := s.createHealthMonitor(svc)
		// Skip services without health monitors.
		if monitor == nil {
			continue
		}
		// Store and start the monitor.
		s.mu.Lock()
		s.healthMonitors[svc.Name] = monitor
		s.mu.Unlock()
		monitor.Start(s.ctx)
	}
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

	// Stop all health monitors first.
	s.mu.RLock()
	// Iterate through health monitors and stop each one.
	for _, monitor := range s.healthMonitors {
		monitor.Stop()
	}
	s.mu.RUnlock()

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
// Errors during stop are reported via handleRecoveryError (best-effort cleanup).
//
// Goroutine lifecycle:
//   - Spawns one goroutine per service for concurrent stop.
//   - All goroutines complete when their respective services stop.
//   - Method blocks until all goroutines complete via WaitGroup.
func (s *Supervisor) stopAll() {
	var wg sync.WaitGroup
	// Iterate through all managers.
	for name, mgr := range s.managers {
		serviceName := name
		m := mgr
		// Stop each manager in a goroutine using Go 1.25's wg.Go().
		wg.Go(func() {
			// Handle stop errors via recovery handler (best-effort cleanup).
			if err := m.Stop(); err != nil {
				s.handleRecoveryError("stop", serviceName, err)
			}
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
// Errors during stop/start are reported via handleRecoveryError (best-effort reload).
//
// Params:
//   - newCfg: the new service configuration.
//
// Goroutine lifecycle:
//   - May spawn new goroutines for monitoring newly added services.
//   - Goroutines run until Stop is called or context is cancelled.
//   - Use Stop() to terminate all monitoring goroutines.
func (s *Supervisor) updateServices(newCfg *domainconfig.Config) {
	// Iterate through all services in the new configuration.
	for i := range newCfg.Services {
		svc := &newCfg.Services[i]
		// Check if the service already exists.
		if mgr, exists := s.managers[svc.Name]; exists {
			// Stop existing manager (best-effort).
			if err := mgr.Stop(); err != nil {
				s.handleRecoveryError("stop-for-reload", svc.Name, err)
			}
			s.managers[svc.Name] = applifecycle.NewManager(svc, s.executor)
			// Start new manager (best-effort).
			if err := s.managers[svc.Name].Start(); err != nil {
				s.handleRecoveryError("start-for-reload", svc.Name, err)
			}
		} else {
			// Create and start a new manager for the new service.
			s.managers[svc.Name] = applifecycle.NewManager(svc, s.executor)
			// Start new manager (best-effort).
			if err := s.managers[svc.Name].Start(); err != nil {
				s.handleRecoveryError("start-new-service", svc.Name, err)
			}
			s.wg.Add(1)
			go s.monitorService(svc.Name, s.managers[svc.Name])
		}
	}
}

// removeDeletedServices removes managers for services no longer in configuration.
// Errors during stop are reported via handleRecoveryError (best-effort cleanup).
//
// Params:
//   - newCfg: the new service configuration.
func (s *Supervisor) removeDeletedServices(newCfg *domainconfig.Config) {
	newServices := make(map[string]bool, len(newCfg.Services))
	// Build a map of services in the new configuration.
	for i := range newCfg.Services {
		newServices[newCfg.Services[i].Name] = true
	}
	// Remove services that are no longer in the configuration.
	for name, mgr := range s.managers {
		// Check if the service should be removed.
		if !newServices[name] {
			// Stop removed service (best-effort).
			if err := mgr.Stop(); err != nil {
				s.handleRecoveryError("stop-removed-service", name, err)
			}
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

// SetErrorHandler sets the callback for non-fatal errors in recovery paths.
// These errors occur during best-effort operations like shutdown cleanup
// or configuration reload where the supervisor continues despite errors.
//
// Params:
//   - handler: the callback function to invoke on non-fatal errors.
func (s *Supervisor) SetErrorHandler(handler ErrorHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorHandler = handler
}

// SetProberFactory sets the health prober factory.
// When set, the supervisor will create health monitors for services
// with listeners that have probe configurations. Health probe failures
// will trigger service restarts following the Kubernetes liveness probe pattern.
//
// Params:
//   - factory: the prober factory for creating health probers.
func (s *Supervisor) SetProberFactory(factory apphealth.Creator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proberFactory = factory
}

// createHealthMonitor creates a health monitor for a service if it has probes configured.
//
// Params:
//   - svc: the service configuration.
//
// Returns:
//   - *apphealth.ProbeMonitor: the health monitor if probes are configured, nil otherwise.
func (s *Supervisor) createHealthMonitor(svc *domainconfig.ServiceConfig) *apphealth.ProbeMonitor {
	// Check prerequisites for health monitoring.
	if !s.hasConfiguredProbes(svc) || s.proberFactory == nil {
		// Return nil when no probes configured or no factory available.
		return nil
	}

	// Create and configure the health monitor.
	monitor := apphealth.NewProbeMonitor(s.createProbeMonitorConfig(svc.Name))

	// Add all listeners with probe bindings.
	s.addListenersWithProbes(monitor, svc)

	// Return the configured monitor.
	return monitor
}

// hasConfiguredProbes checks if any listener in the service has a probe configured.
//
// Params:
//   - svc: the service configuration.
//
// Returns:
//   - bool: true if at least one listener has a probe configured.
func (s *Supervisor) hasConfiguredProbes(svc *domainconfig.ServiceConfig) bool {
	// Iterate through listeners looking for configured probes.
	for i := range svc.Listeners {
		// Check if listener has a valid probe configuration.
		if svc.Listeners[i].Probe != nil && svc.Listeners[i].Probe.Type != "" {
			// Return true when probe found.
			return true
		}
	}
	// Return false when no probes configured.
	return false
}

// createProbeMonitorConfig creates the configuration for a health monitor.
//
// Params:
//   - serviceName: the name of the service being monitored.
//
// Returns:
//   - apphealth.ProbeMonitorConfig: the monitor configuration with callbacks.
func (s *Supervisor) createProbeMonitorConfig(serviceName string) apphealth.ProbeMonitorConfig {
	// Build and return the monitor configuration.
	return apphealth.ProbeMonitorConfig{
		Factory: s.proberFactory,
		OnStateChange: func(listenerName string, prevState, newState domainhealth.SubjectState, result domainhealth.CheckResult) {
			// Log health state transition.
			log.Printf("[health] service=%s listener=%s state=%s->%s success=%v latency=%s",
				serviceName, listenerName, prevState, newState, result.Success, result.Latency)
		},
		OnUnhealthy: func(listenerName, reason string) {
			// Log unhealthy transition and trigger restart.
			log.Printf("[health] service=%s listener=%s unhealthy: %s - triggering restart",
				serviceName, listenerName, reason)
			// Trigger restart on health failure.
			if err := s.RestartOnHealthFailure(serviceName, reason); err != nil {
				s.handleRecoveryError("health-restart", serviceName, err)
			}
		},
	}
}

// addListenersWithProbes adds all listeners with probe configurations to the monitor.
//
// Params:
//   - monitor: the health monitor to add listeners to.
//   - svc: the service configuration containing listeners.
func (s *Supervisor) addListenersWithProbes(monitor *apphealth.ProbeMonitor, svc *domainconfig.ServiceConfig) {
	// Iterate through all listeners in the service.
	for i := range svc.Listeners {
		lc := &svc.Listeners[i]
		// Skip listeners without valid probe configuration.
		if lc.Probe == nil || lc.Probe.Type == "" {
			continue
		}
		// Add the listener with its probe binding.
		s.addSingleListenerWithProbe(monitor, lc, svc.Name)
	}
}

// addSingleListenerWithProbe adds a single listener with its probe binding to the monitor.
//
// Params:
//   - monitor: the health monitor (minimal interface).
//   - lc: the listener configuration.
//   - serviceName: the name of the service for logging.
func (s *Supervisor) addSingleListenerWithProbe(monitor AddListenerWithBindinger, lc *domainconfig.ListenerConfig, serviceName string) {
	// Create domain listener with resolved defaults.
	domainListener := s.createDomainListener(lc)

	// Create probe binding from configuration.
	binding := s.createProbeBinding(lc)

	// Add listener with binding to monitor.
	if err := monitor.AddListenerWithBinding(domainListener, binding); err != nil {
		log.Printf("[health] failed to add listener %s for service %s: %v",
			lc.Name, serviceName, err)
	}
}

// createDomainListener creates a domain listener from configuration with defaults.
//
// Params:
//   - lc: the listener configuration.
//
// Returns:
//   - *listener.Listener: the domain listener ready for health monitoring.
func (s *Supervisor) createDomainListener(lc *domainconfig.ListenerConfig) *listener.Listener {
	// Resolve protocol with default.
	protocol := lc.Protocol
	// Apply default protocol if not specified.
	if protocol == "" {
		protocol = "tcp"
	}
	// Resolve address with default.
	address := lc.Address
	// Apply default address if not specified.
	if address == "" {
		address = "localhost"
	}
	// Create and mark listener as listening.
	domainListener := listener.NewListener(lc.Name, protocol, address, lc.Port)
	domainListener.MarkListening()
	// Return the configured listener.
	return domainListener
}

// createProbeBinding creates a probe binding from listener configuration.
//
// Params:
//   - lc: the listener configuration with probe settings.
//
// Returns:
//   - *apphealth.ProbeBinding: the probe binding for health monitoring.
func (s *Supervisor) createProbeBinding(lc *domainconfig.ListenerConfig) *apphealth.ProbeBinding {
	// Resolve address with default for target.
	address := lc.Address
	// Apply default address if not specified.
	if address == "" {
		address = "localhost"
	}
	// Build and return the probe binding.
	return &apphealth.ProbeBinding{
		ListenerName: lc.Name,
		Type:         apphealth.ProbeType(lc.Probe.Type),
		Target: apphealth.ProbeTarget{
			Address: fmt.Sprintf("%s:%d", address, lc.Port),
			Path:    lc.Probe.Path,
			Service: lc.Probe.Service,
		},
		Config: apphealth.ProbeConfig{
			Timeout:          lc.Probe.Timeout.Duration(),
			Interval:         lc.Probe.Interval.Duration(),
			SuccessThreshold: lc.Probe.SuccessThreshold,
			FailureThreshold: lc.Probe.FailureThreshold,
		},
	}
}

// handleRecoveryError reports a non-fatal error to the error handler if set.
// This method is called from recovery/cleanup paths where errors don't stop
// the overall operation.
//
// Params:
//   - operation: description of the operation that failed (e.g., "stop", "start").
//   - serviceName: the name of the affected service.
//   - err: the error that occurred.
func (s *Supervisor) handleRecoveryError(operation, serviceName string, err error) {
	// Skip if no error.
	if err == nil {
		// Return early when there is no error to handle.
		return
	}

	s.mu.RLock()
	handler := s.errorHandler
	s.mu.RUnlock()

	// Call handler if registered.
	if handler != nil {
		handler(operation, serviceName, err)
	}
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

// RestartOnHealthFailure triggers a restart for a service due to health probe failure.
// This implements the Kubernetes liveness probe pattern: when health probes
// fail consecutively beyond the failure threshold, the service is restarted.
//
// Params:
//   - serviceName: the name of the service to restart.
//   - reason: description of why the health check failed.
//
// Returns:
//   - error: ErrServiceNotFound if service doesn't exist, or error from manager.
func (s *Supervisor) RestartOnHealthFailure(serviceName, reason string) error {
	s.mu.RLock()
	mgr, ok := s.managers[serviceName]
	s.mu.RUnlock()

	// Check if the service exists.
	if !ok {
		// Return error for missing service.
		return fmt.Errorf("%w: %s", ErrServiceNotFound, serviceName)
	}

	// Delegate to the manager's health failure restart method.
	return mgr.RestartOnHealthFailure(reason)
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

// ServiceSnapshotForTUI contains service info for TUI display.
// This struct uses basic types to avoid import cycles with TUI packages.
type ServiceSnapshotForTUI struct {
	Name         string
	StateInt     int
	StateName    string
	PID          int
	UptimeSecs   int64
	CPUPercent   float64
	MemoryRSS    uint64
	RestartCount int
}

// ServiceSnapshotsForTUI returns service data formatted for TUI display.
//
// Returns:
//   - []ServiceSnapshotForTUI: a slice of service snapshots.
func (s *Supervisor) ServiceSnapshotsForTUI() []ServiceSnapshotForTUI {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]ServiceSnapshotForTUI, 0, len(s.managers))
	for _, mgr := range s.managers {
		state := mgr.State()
		result = append(result, ServiceSnapshotForTUI{
			Name:       mgr.Name(),
			StateInt:   int(state),
			StateName:  state.String(),
			PID:        mgr.PID(),
			UptimeSecs: mgr.Uptime(),
		})
	}
	return result
}

// Service returns a specific service manager.
//
// Params:
//   - name: the service name.
//
// Returns:
//   - *applifecycle.Manager: the manager if found.
//   - bool: true if the service was found.
func (s *Supervisor) Service(name string) (*applifecycle.Manager, bool) {
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
