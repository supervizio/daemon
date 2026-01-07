package supervisor

import (
	"context"
	"fmt"
	"sync"

	"github.com/kodflow/daemon/internal/config"
	"github.com/kodflow/daemon/internal/kernel"
	"github.com/kodflow/daemon/internal/kernel/ports"
	"github.com/kodflow/daemon/internal/process"
)

// State represents the supervisor state.
type State int

const (
	StateStopped State = iota
	StateStarting
	StateRunning
	StateStopping
)

// Supervisor manages multiple services and their lifecycle.
type Supervisor struct {
	mu       sync.RWMutex
	config   *config.Config
	managers map[string]*process.Manager
	reaper   ports.ZombieReaper
	state    State
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// New creates a new supervisor from configuration.
func New(cfg *config.Config) (*Supervisor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	s := &Supervisor{
		config:   cfg,
		managers: make(map[string]*process.Manager),
		state:    StateStopped,
	}

	// Create process managers for each service
	for i := range cfg.Services {
		svc := &cfg.Services[i]
		s.managers[svc.Name] = process.NewManager(svc)
	}

	// Create reaper if running as PID 1
	if kernel.Default.Reaper.IsPID1() {
		s.reaper = kernel.Default.Reaper
	}

	return s, nil
}

// Start starts all managed services.
func (s *Supervisor) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.state != StateStopped {
		s.mu.Unlock()
		return fmt.Errorf("supervisor already running")
	}
	s.state = StateStarting
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.mu.Unlock()

	// Start zombie reaper if PID 1
	if s.reaper != nil {
		s.reaper.Start()
	}

	// Start all services
	var startErr error
	for name, mgr := range s.managers {
		if err := mgr.Start(); err != nil {
			startErr = fmt.Errorf("failed to start service %s: %w", name, err)
			break
		}
	}

	if startErr != nil {
		// Stop already started services
		s.stopAll()
		s.mu.Lock()
		s.state = StateStopped
		s.mu.Unlock()
		return startErr
	}

	// Start event monitoring
	for name, mgr := range s.managers {
		s.wg.Add(1)
		go s.monitorService(name, mgr)
	}

	s.mu.Lock()
	s.state = StateRunning
	s.mu.Unlock()

	return nil
}

// Stop gracefully stops all managed services.
func (s *Supervisor) Stop() error {
	s.mu.Lock()
	if s.state != StateRunning {
		s.mu.Unlock()
		return nil
	}
	s.state = StateStopping
	s.cancel()
	s.mu.Unlock()

	// Stop all services
	s.stopAll()

	// Wait for all monitors to finish
	s.wg.Wait()

	// Stop reaper
	if s.reaper != nil {
		s.reaper.Stop()
	}

	s.mu.Lock()
	s.state = StateStopped
	s.mu.Unlock()

	return nil
}

// stopAll stops all managed services.
func (s *Supervisor) stopAll() {
	var wg sync.WaitGroup
	for _, mgr := range s.managers {
		wg.Add(1)
		go func(m *process.Manager) {
			defer wg.Done()
			m.Stop()
		}(mgr)
	}
	wg.Wait()
}

// Reload reloads the configuration and restarts changed services.
func (s *Supervisor) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != StateRunning {
		return fmt.Errorf("supervisor not running")
	}

	// Reload configuration
	newCfg, err := config.Load(s.config.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	// Update services
	for i := range newCfg.Services {
		svc := &newCfg.Services[i]
		if mgr, exists := s.managers[svc.Name]; exists {
			// Restart service with new config
			mgr.Stop()
			s.managers[svc.Name] = process.NewManager(svc)
			s.managers[svc.Name].Start()
		} else {
			// New service
			s.managers[svc.Name] = process.NewManager(svc)
			s.managers[svc.Name].Start()
			s.wg.Add(1)
			go s.monitorService(svc.Name, s.managers[svc.Name])
		}
	}

	// Remove deleted services
	newServices := make(map[string]bool)
	for _, svc := range newCfg.Services {
		newServices[svc.Name] = true
	}
	for name, mgr := range s.managers {
		if !newServices[name] {
			mgr.Stop()
			delete(s.managers, name)
		}
	}

	s.config = newCfg
	return nil
}

// monitorService monitors a service for events.
func (s *Supervisor) monitorService(name string, mgr *process.Manager) {
	defer s.wg.Done()

	events := mgr.Events()
	for {
		select {
		case <-s.ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			s.handleEvent(name, event)
		}
	}
}

// handleEvent handles a process event.
func (s *Supervisor) handleEvent(name string, event process.Event) {
	switch event.Type {
	case process.EventStarted:
		// Service started successfully
	case process.EventStopped:
		// Service stopped
	case process.EventFailed:
		// Service failed - manager handles restart logic
	case process.EventRestarting:
		// Service is restarting
	}
}

// State returns the current supervisor state.
func (s *Supervisor) State() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Services returns information about all managed services.
func (s *Supervisor) Services() map[string]ServiceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := make(map[string]ServiceInfo)
	for name, mgr := range s.managers {
		info[name] = ServiceInfo{
			Name:   name,
			State:  mgr.State(),
			PID:    mgr.PID(),
			Uptime: mgr.Uptime(),
		}
	}
	return info
}

// ServiceInfo contains information about a service.
type ServiceInfo struct {
	Name   string
	State  process.State
	PID    int
	Uptime int64 // seconds
}

// Service returns a specific service manager.
func (s *Supervisor) Service(name string) (*process.Manager, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	mgr, ok := s.managers[name]
	return mgr, ok
}

// StartService starts a specific service.
func (s *Supervisor) StartService(name string) error {
	s.mu.RLock()
	mgr, ok := s.managers[name]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("service %s not found", name)
	}
	return mgr.Start()
}

// StopService stops a specific service.
func (s *Supervisor) StopService(name string) error {
	s.mu.RLock()
	mgr, ok := s.managers[name]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("service %s not found", name)
	}
	return mgr.Stop()
}

// RestartService restarts a specific service.
func (s *Supervisor) RestartService(name string) error {
	s.mu.RLock()
	mgr, ok := s.managers[name]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("service %s not found", name)
	}

	if err := mgr.Stop(); err != nil {
		return err
	}
	return mgr.Start()
}
