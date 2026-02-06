// Package supervisor provides the application service for orchestrating multiple services.
// It manages the lifecycle of services including start, stop, restart, and reload operations.
package supervisor

import (
	"context"
	"fmt"
	"sort"
	"sync"

	appconfig "github.com/kodflow/daemon/internal/application/config"
	apphealth "github.com/kodflow/daemon/internal/application/health"
	applifecycle "github.com/kodflow/daemon/internal/application/lifecycle"
	appmetrics "github.com/kodflow/daemon/internal/application/metrics"
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
// The stats parameter contains an atomic snapshot of service statistics.
type EventHandler func(serviceName string, event *domain.Event, stats *ServiceStatsSnapshot)

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
	// metricsTracker tracks process CPU and memory metrics.
	metricsTracker appmetrics.ProcessTracker
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
		// return error if configuration is invalid
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

	// create managers and stats for each service
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		s.managers[svc.Name] = applifecycle.NewManager(svc, executor)
		s.stats[svc.Name] = NewServiceStats()
	}

	// return initialized supervisor
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
		// initialize supervisor state and context
		return err
	}

	// Start zombie reaper if configured.
	s.startReaper()

	// Start all managed services.
	if err := s.startAllServices(); err != nil {
		// start all managed services
		return err
	}

	// Start monitoring goroutines for each service.
	s.startMonitoringGoroutines()

	s.startHealthMonitors()

	// Mark supervisor as running.
	s.mu.Lock()
	s.state = StateRunning
	s.mu.Unlock()

	// mark supervisor as running
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

	// check if already running
	if s.state != StateStopped {
		// Return error when supervisor is already running.
		return ErrAlreadyRunning
	}

	s.state = StateStarting
	s.ctx, s.cancel = context.WithCancel(ctx)
	// return success after initialization
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
		err := mgr.Start()
		// Skip successfully started services.
		if err == nil {
			continue
		}
		// Handle startup failure by stopping all services.
		s.stopAll()
		s.mu.Lock()
		s.state = StateStopped
		s.mu.Unlock()
		// Return wrapped start error.
		return fmt.Errorf("failed to start service %s: %w", name, err)
	}
	// return success after starting all services
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
	// start monitoring goroutine for each service
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
	// return nil when not running
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

	// return success after graceful stop
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
	s.mu.RLock()
	state := s.state
	configPath := s.config.ConfigPath
	s.mu.RUnlock()

	// return error when not running
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
	// return success after reload
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
	// iterate to find removed services
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
	stats := s.getOrCreateStats(name)

	s.updateStatsForEvent(stats, event)
	s.updateHealthMonitor(name, event)
	s.updateMetricsTracker(name, event)

	statsSnap := s.getStatsSnapshot(stats)
	s.mu.Unlock()

	s.callEventHandler(name, event, statsSnap)
}

// getOrCreateStats gets or creates stats for a service.
//
// Params:
//   - name: the service name.
//
// Returns:
//   - *ServiceStats: the service statistics.
func (s *Supervisor) getOrCreateStats(name string) *ServiceStats {
	stats, ok := s.stats[name]
	// create new stats if not found
	if !ok {
		stats = NewServiceStats()
		s.stats[name] = stats
	}
	// return existing or new stats
	return stats
}

// updateStatsForEvent updates statistics based on event type.
//
// Params:
//   - stats: the service statistics to update.
//   - event: the process event.
func (s *Supervisor) updateStatsForEvent(stats *ServiceStats, event *domain.Event) {
	// Increment counters based on event type.
	switch event.Type {
	// Process started.
	case domain.EventStarted:
		stats.IncrementStart()
	// Process stopped cleanly.
	case domain.EventStopped:
		stats.IncrementStop()
	// Process failed.
	case domain.EventFailed:
		stats.IncrementFail()
	// Process restarting.
	case domain.EventRestarting:
		stats.IncrementRestart()
	// Restart attempts exhausted.
	case domain.EventExhausted:
		stats.IncrementFail()
	// Health events are tracked separately.
	case domain.EventHealthy, domain.EventUnhealthy:
		// Health events are tracked by the health monitor, not stats.
	default:
		// Unknown event type, ignore.
	}
}

// updateHealthMonitor updates health monitor process state if monitor exists.
//
// Params:
//   - name: the service name.
//   - event: the process event.
func (s *Supervisor) updateHealthMonitor(name string, event *domain.Event) {
	monitor, ok := s.healthMonitors[name]
	// Skip if no health monitor configured.
	if !ok {
		// Not found.
		return
	}

	// determine which counter to increment
	switch event.Type {
	// process started successfully
	case domain.EventStarted:
		monitor.SetProcessState(domain.StateRunning)
	// process stopped or failed
	case domain.EventStopped, domain.EventFailed, domain.EventExhausted:
		monitor.SetProcessState(domain.StateStopped)
	// No state change for these events.
	case domain.EventRestarting, domain.EventHealthy, domain.EventUnhealthy:
		// No state change needed.
	default:
		// Unknown event type, ignore.
	}
}

// updateMetricsTracker updates metrics tracker if configured and event is relevant.
//
// Params:
//   - name: the service name.
//   - event: the process event.
func (s *Supervisor) updateMetricsTracker(name string, event *domain.Event) {
	// Skip if metrics tracking disabled.
	if s.metricsTracker == nil {
		// Metrics tracking disabled.
		return
	}

	// Start or stop metrics tracking based on event.
	switch event.Type {
	// Start tracking process metrics.
	case domain.EventStarted:
		// Validate PID before tracking.
		if event.PID > 0 {
			_ = s.metricsTracker.Track(name, event.PID)
		}
	// Stop tracking metrics.
	case domain.EventStopped, domain.EventFailed, domain.EventExhausted:
		s.metricsTracker.Untrack(name)
	// No metrics action needed.
	case domain.EventRestarting, domain.EventHealthy, domain.EventUnhealthy:
		// No action needed.
	default:
		// Unknown event type, ignore.
	}
}

// getStatsSnapshot returns a snapshot pointer for the service stats.
//
// Params:
//   - stats: the service statistics.
//
// Returns:
//   - *ServiceStatsSnapshot: atomic snapshot pointer or nil.
func (s *Supervisor) getStatsSnapshot(stats *ServiceStats) *ServiceStatsSnapshot {
	// handle nil stats
	if stats == nil {
		// Return nil.
		return nil
	}
	// return snapshot pointer
	return stats.SnapshotPtr()
}

// callEventHandler calls user event handler if registered.
//
// Params:
//   - name: the service name.
//   - event: the process event.
//   - statsSnap: the statistics snapshot.
func (s *Supervisor) callEventHandler(name string, event *domain.Event, statsSnap *ServiceStatsSnapshot) {
	// call handler if registered
	if s.eventHandler != nil {
		s.eventHandler(name, event, statsSnap)
	}
}

// SetEventHandler sets the callback for process events.
//
// Params:
//   - handler: the callback function to invoke on events.
func (s *Supervisor) SetEventHandler(handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// store event handler
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
	// store error handler
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
	// store prober factory
	defer s.mu.Unlock()
	s.proberFactory = factory
}

// SetMetricsTracker sets the process metrics tracker.
// When set, the supervisor will track CPU and memory usage per service.
//
// Params:
//   - tracker: the metrics tracker to use.
func (s *Supervisor) SetMetricsTracker(tracker appmetrics.ProcessTracker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// store metrics tracker
	s.metricsTracker = tracker
}

// createHealthMonitor creates a health monitor for a service if it has probes configured.
//
// Params:
//   - svc: the service configuration.
//
// Returns:
//   - *apphealth.ProbeMonitor: the health monitor if probes are configured, nil otherwise.
func (s *Supervisor) createHealthMonitor(svc *domainconfig.ServiceConfig) *apphealth.ProbeMonitor {
	// return nil if no probes configured or factory unavailable
	if !s.hasConfiguredProbes(svc) || s.proberFactory == nil {
		// Return nil when no probes configured or no factory available.
		return nil
	}

	monitor := apphealth.NewProbeMonitor(s.createProbeMonitorConfig(svc.Name))

	// Add all listeners with probe bindings.
	s.addListenersWithProbes(monitor, svc)

	// return configured monitor
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
	// no probes configured
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
	// return monitor configuration
	return apphealth.ProbeMonitorConfig{
		Factory: s.proberFactory,
		OnStateChange: func(_ string, _, _ domainhealth.SubjectState, _ domainhealth.CheckResult) {
			// Health state transitions are tracked internally.
			// Events are emitted via OnHealthy/OnUnhealthy callbacks.
		},
		OnUnhealthy: func(_, reason string) {
			// Trigger restart on health failure (event emitted by restart logic).
			// Attempt to restart the service on health failure.
			if err := s.RestartOnHealthFailure(serviceName, reason); err != nil {
				s.handleRecoveryError("health-restart", serviceName, err)
			}
		},
		OnHealthy: func(_ string) {
			// Emit healthy event when service becomes healthy.
			// Call event handler if registered.
			if s.eventHandler != nil {
				s.mu.RLock()
				stats, ok := s.stats[serviceName]
				var statsSnap *ServiceStatsSnapshot
				// Get stats snapshot if available.
				if ok && stats != nil {
					snap := stats.Snapshot()
					statsSnap = &snap
				}
				s.mu.RUnlock()
				event := &domain.Event{
					Type: domain.EventHealthy,
				}
				s.eventHandler(serviceName, event, statsSnap)
			}
		},
	}
	// restart service on consecutive failures
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
		s.addSingleListenerWithProbe(monitor, lc)
	}
}

// addSingleListenerWithProbe adds a single listener with its probe binding to the monitor.
//
// Params:
//   - monitor: the health monitor (minimal interface).
//   - lc: the listener configuration.
func (s *Supervisor) addSingleListenerWithProbe(monitor AddListenerWithBindinger, lc *domainconfig.ListenerConfig) {
	domainListener := s.createDomainListener(lc)

	binding := s.createProbeBinding(lc)

	// Add listener with binding to monitor.
	// Errors are silently ignored - health monitoring is best-effort.
	_ = monitor.AddListenerWithBinding(domainListener, binding)
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
	// use default protocol if not specified
	if protocol == "" {
		protocol = "tcp"
	}
	// Resolve address with default.
	address := lc.Address
	// use default address if not specified
	if address == "" {
		address = "localhost"
	}
	domainListener := listener.NewListener(lc.Name, protocol, address, lc.Port)
	domainListener.MarkListening()
	// return configured domain listener
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
	// use default address if not specified
	if address == "" {
		address = "localhost"
	}
	// return probe binding configuration
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
	// skip if no error to handle
	if err == nil {
		// Return early when there is no error to handle.
		return
	}

	s.mu.RLock()
	handler := s.errorHandler
	s.mu.RUnlock()

	// invoke error handler if set
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
//   - *ServiceStatsSnapshot: the service statistics snapshot, or nil if not found.
func (s *Supervisor) Stats(name string) *ServiceStatsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// return snapshot if stats exist
	if stats, ok := s.stats[name]; ok {
		snap := stats.Snapshot()
		// Return pointer to snapshot.
		return &snap
	}
	// Service not found, return nil.
	return nil
}

// AllStats returns statistics for all services.
//
// Returns:
//   - map[string]*ServiceStatsSnapshot: atomic snapshots of all service statistics.
func (s *Supervisor) AllStats() map[string]*ServiceStatsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Use SnapshotPtr to avoid escape analysis issue from &Snapshot().
	result := make(map[string]*ServiceStatsSnapshot, len(s.stats))
	// Iterate through stats and collect snapshots.
	for name, stats := range s.stats {
		result[name] = stats.SnapshotPtr()
	}
	// return all stats snapshots
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

	// validate service exists
	if !ok {
		// Return error for missing service.
		return fmt.Errorf("%w: %s", ErrServiceNotFound, serviceName)
	}

	// delegate to manager
	return mgr.RestartOnHealthFailure(reason)
}

// State returns the current supervisor state.
//
// Returns:
//   - State: the current state.
func (s *Supervisor) State() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// return current supervisor state
	return s.state
}

// Services returns information about all managed services.
//
// Returns:
//   - map[string]ServiceInfo: a map of service names to their information.
func (s *Supervisor) Services() map[string]ServiceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// collect information from each manager
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
	// return collected service information
	return info
}

// ServiceSnapshotsForTUI returns service data formatted for TUI display.
// Services are returned in alphabetical order for stable display.
//
// Returns:
//   - []ServiceSnapshotForTUI: a slice of service snapshots sorted by name.
func (s *Supervisor) ServiceSnapshotsForTUI() []ServiceSnapshotForTUI {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect all services.
	// collect all services
	result := make([]ServiceSnapshotForTUI, 0, len(s.managers))
	// Iterate through managers and build snapshots.
	for name, mgr := range s.managers {
		state := mgr.State()
		snap := ServiceSnapshotForTUI{
			Name:       mgr.Name(),
			StateInt:   int(state),
			StateName:  state.String(),
			PID:        mgr.PID(),
			UptimeSecs: mgr.Uptime(),
		}

		// Get listening ports for running processes.
		if snap.PID > 0 {
			snap.Ports = getListeningPorts(snap.PID)
		}

		// Enrich with config-based data (health checks, listeners).
		s.enrichSnapshotWithConfig(&snap, name)

		// Enrich with runtime metrics (health status, restart count, CPU/memory).
		s.enrichSnapshotWithMetrics(&snap, name)

		result = append(result, snap)
	}

	// Sort alphabetically by name for stable display.
	sort.Slice(result, func(i, j int) bool {
		// Compare service names for alphabetical ordering.
		return result[i].Name < result[j].Name
	})

	// return sorted service snapshots
	return result
}

// enrichSnapshotWithConfig enriches snapshot with config-based data.
//
// Params:
//   - snap: target snapshot to enrich.
//   - name: service name to lookup in config.
func (s *Supervisor) enrichSnapshotWithConfig(snap *ServiceSnapshotForTUI, name string) {
	// Find service configuration.
	for i := range s.config.Services {
		// Match service by name.
		if s.config.Services[i].Name == name {
			snap.HasHealthChecks = s.hasConfiguredProbes(&s.config.Services[i])
			snap.Listeners = s.buildListenerSnapshots(&s.config.Services[i], snap.Ports)
			// Found and enriched - exit early.
			return
		}
	}
}

// enrichSnapshotWithMetrics enriches snapshot with runtime metrics.
//
// Params:
//   - snap: target snapshot to enrich.
//   - name: service name to lookup metrics for.
func (s *Supervisor) enrichSnapshotWithMetrics(snap *ServiceSnapshotForTUI, name string) {
	// add health status if monitor exists
	if monitor, ok := s.healthMonitors[name]; ok {
		snap.HealthInt = int(monitor.Status())
	}

	// add restart count if stats exist
	if stats, ok := s.stats[name]; ok {
		snap.RestartCount = stats.RestartCount()
	}

	// add cpu and memory metrics if available
	if s.metricsTracker != nil {
		// Retrieve metrics if available for this service.
		if metrics, ok := s.metricsTracker.Get(name); ok {
			snap.CPUPercent = metrics.CPU.UsagePercent
			snap.MemoryRSS = metrics.Memory.RSS
		}
	}
}

// buildListenerSnapshots creates listener snapshots with status from config.
// Status logic:
//   - 0 (OK/Green): port listening and state matches config
//   - 1 (Warning/Yellow): mismatch (exposed but not reachable, or vice versa)
//   - 2 (Error/Red): expected port but nothing listening
//
// Params:
//   - svc: the service configuration with listener definitions.
//   - listeningPorts: the list of actually listening ports from the process.
//
// Returns:
//   - []ListenerSnapshotForTUI: listener snapshots with status indicators.
func (s *Supervisor) buildListenerSnapshots(svc *domainconfig.ServiceConfig, listeningPorts []int) []ListenerSnapshotForTUI {
	listening := make(map[int]bool, len(listeningPorts))
	// build listening ports map
	for _, p := range listeningPorts {
		listening[p] = true
	}

	result := make([]ListenerSnapshotForTUI, 0, len(svc.Listeners))
	// create listener snapshot for each configured listener
	for _, lc := range svc.Listeners {
		ls := ListenerSnapshotForTUI{
			Name:      lc.Name,
			Port:      lc.Port,
			Protocol:  lc.Protocol,
			Exposed:   lc.Exposed,
			Listening: listening[lc.Port],
		}

		// Determine status based on listening state.
		// StatusInt: 0=OK, 2=Error
		if ls.Listening {
			// Listening → OK (green), whether exposed or internal.
			ls.StatusInt = ListenerStatusOK
		} else {
			// Expected port but nothing listening → Error (red).
			ls.StatusInt = ListenerStatusError
		}

		result = append(result, ls)
	}

	// return listener snapshots
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
	// return manager and existence flag
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

	// validate service exists
	if !ok {
		// Return error for missing service.
		return fmt.Errorf("%w: %s", ErrServiceNotFound, name)
	}
	// start the service
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

	// validate service exists
	if !ok {
		// Return error for missing service.
		return fmt.Errorf("%w: %s", ErrServiceNotFound, name)
	}
	// stop the service
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

	// validate service exists
	if !ok {
		// Return error for missing service.
		return fmt.Errorf("%w: %s", ErrServiceNotFound, name)
	}

	// Stop the service first.
	// stop the service first
	if err := mgr.Stop(); err != nil {
		// Return stop error if failed.
		return err
	}
	// start the service after stop
	return mgr.Start()
}
