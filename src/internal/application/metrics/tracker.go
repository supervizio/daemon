// Package metrics provides application services for process metrics tracking.
package metrics

import (
	"context"
	"sync"
	"time"

	domainmetrics "github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
)

// Default configuration values.
const (
	defaultCollectionInterval = 5 * time.Second
	defaultSubscriberBuffer   = 64
)

// trackedProcess holds the state for a single tracked process.
type trackedProcess struct {
	serviceName  string
	pid          int
	state        process.State
	healthy      bool
	startTime    time.Time
	restartCount int
	lastError    string
	lastMetrics  domainmetrics.ProcessMetrics
}

// Tracker implements ProcessTracker using infrastructure collectors.
type Tracker struct {
	mu          sync.RWMutex
	collector   Collector
	processes   map[string]*trackedProcess
	interval    time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
	running     bool
	subsMu      sync.RWMutex
	subscribers map[chan domainmetrics.ProcessMetrics]struct{}
}

// TrackerOption configures a Tracker.
type TrackerOption func(*Tracker)

// WithCollectionInterval sets the metrics collection interval.
func WithCollectionInterval(d time.Duration) TrackerOption {
	return func(t *Tracker) {
		if d > 0 {
			t.interval = d
		}
	}
}

// NewTracker creates a new process metrics tracker.
func NewTracker(collector Collector, opts ...TrackerOption) *Tracker {
	t := &Tracker{
		collector:   collector,
		processes:   make(map[string]*trackedProcess),
		interval:    defaultCollectionInterval,
		subscribers: make(map[chan domainmetrics.ProcessMetrics]struct{}),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Start begins the metrics collection loop.
func (t *Tracker) Start(ctx context.Context) error {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return nil
	}
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.running = true
	t.mu.Unlock()

	go t.collectLoop()
	return nil
}

// Stop stops the metrics collection loop.
func (t *Tracker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return
	}

	if t.cancel != nil {
		t.cancel()
	}
	t.running = false
}

// Track starts tracking metrics for a service with the given PID.
func (t *Tracker) Track(_ context.Context, serviceName string, pid int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	existing, exists := t.processes[serviceName]
	if exists {
		// Same service, new PID = restart
		existing.pid = pid
		existing.state = process.StateRunning
		existing.startTime = now
		existing.restartCount++
		existing.lastError = ""
		// Update lastMetrics with new state
		existing.lastMetrics = t.buildMetrics(existing, now)
	} else {
		// New service
		proc := &trackedProcess{
			serviceName:  serviceName,
			pid:          pid,
			state:        process.StateRunning,
			healthy:      true,
			startTime:    now,
			restartCount: 0,
		}
		// Initialize lastMetrics
		proc.lastMetrics = t.buildMetrics(proc, now)
		t.processes[serviceName] = proc
	}

	return nil
}

// buildMetrics creates a ProcessMetrics from tracked process state.
// Must be called with t.mu held.
func (t *Tracker) buildMetrics(proc *trackedProcess, now time.Time) domainmetrics.ProcessMetrics {
	m := domainmetrics.ProcessMetrics{
		ServiceName:  proc.serviceName,
		PID:          proc.pid,
		State:        proc.state,
		Healthy:      proc.healthy,
		CPU:          proc.lastMetrics.CPU,
		Memory:       proc.lastMetrics.Memory,
		StartTime:    proc.startTime,
		RestartCount: proc.restartCount,
		LastError:    proc.lastError,
		Timestamp:    now,
	}

	if !proc.startTime.IsZero() && proc.pid > 0 {
		m.Uptime = now.Sub(proc.startTime)
	}

	return m
}

// Untrack stops tracking metrics for a service.
func (t *Tracker) Untrack(serviceName string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.processes, serviceName)
}

// Get returns the current metrics for a service.
func (t *Tracker) Get(serviceName string) (domainmetrics.ProcessMetrics, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	proc, exists := t.processes[serviceName]
	if !exists {
		return domainmetrics.ProcessMetrics{}, false
	}

	return proc.lastMetrics, true
}

// GetAll returns metrics for all tracked services.
func (t *Tracker) GetAll() []domainmetrics.ProcessMetrics {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]domainmetrics.ProcessMetrics, 0, len(t.processes))
	for _, proc := range t.processes {
		result = append(result, proc.lastMetrics)
	}
	return result
}

// Subscribe returns a channel that receives metrics updates.
func (t *Tracker) Subscribe() <-chan domainmetrics.ProcessMetrics {
	ch := make(chan domainmetrics.ProcessMetrics, defaultSubscriberBuffer)

	t.subsMu.Lock()
	t.subscribers[ch] = struct{}{}
	t.subsMu.Unlock()

	return ch
}

// Unsubscribe removes a subscription channel.
func (t *Tracker) Unsubscribe(ch <-chan domainmetrics.ProcessMetrics) {
	// Type assert to get the sendable channel
	sendCh, ok := interface{}(ch).(chan domainmetrics.ProcessMetrics)
	if !ok {
		return
	}

	t.subsMu.Lock()
	delete(t.subscribers, sendCh)
	t.subsMu.Unlock()

	close(sendCh)
}

// UpdateState updates the state of a tracked process.
func (t *Tracker) UpdateState(serviceName string, state process.State, lastError string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	proc, exists := t.processes[serviceName]
	if !exists {
		return
	}

	proc.state = state
	if lastError != "" {
		proc.lastError = lastError
	}
	if state == process.StateFailed || state == process.StateStopped {
		proc.pid = 0
	}
	// Update lastMetrics with new state
	proc.lastMetrics = t.buildMetrics(proc, time.Now())
}

// UpdateHealth updates the health status of a tracked process.
func (t *Tracker) UpdateHealth(serviceName string, healthy bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	proc, exists := t.processes[serviceName]
	if !exists {
		return
	}

	proc.healthy = healthy
	// Update lastMetrics with new health
	proc.lastMetrics = t.buildMetrics(proc, time.Now())
}

// collectLoop periodically collects metrics for all tracked processes.
func (t *Tracker) collectLoop() {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			t.collectAll()
		}
	}
}

// collectAll collects metrics for all tracked processes.
func (t *Tracker) collectAll() {
	t.mu.Lock()
	processes := make([]*trackedProcess, 0, len(t.processes))
	for _, proc := range t.processes {
		processes = append(processes, proc)
	}
	t.mu.Unlock()

	for _, proc := range processes {
		t.collectProcess(proc)
	}
}

// collectProcess collects metrics for a single process.
func (t *Tracker) collectProcess(proc *trackedProcess) {
	if proc.pid <= 0 {
		t.updateProcessMetrics(proc, domainmetrics.ProcessCPU{}, domainmetrics.ProcessMemory{})
		return
	}

	ctx, cancel := context.WithTimeout(t.ctx, t.interval/2)
	defer cancel()

	cpu, cpuErr := t.collector.CollectCPU(ctx, proc.pid)
	mem, memErr := t.collector.CollectMemory(ctx, proc.pid)

	// If both fail, process may have exited
	if cpuErr != nil && memErr != nil {
		t.UpdateState(proc.serviceName, process.StateFailed, "process not found")
		return
	}

	t.updateProcessMetrics(proc, cpu, mem)
}

// updateProcessMetrics updates and publishes metrics for a process.
func (t *Tracker) updateProcessMetrics(proc *trackedProcess, cpu domainmetrics.ProcessCPU, mem domainmetrics.ProcessMemory) {
	t.mu.Lock()
	now := time.Now()

	m := domainmetrics.ProcessMetrics{
		ServiceName:  proc.serviceName,
		PID:          proc.pid,
		State:        proc.state,
		Healthy:      proc.healthy,
		CPU:          cpu,
		Memory:       mem,
		StartTime:    proc.startTime,
		RestartCount: proc.restartCount,
		LastError:    proc.lastError,
		Timestamp:    now,
	}

	if !proc.startTime.IsZero() && proc.pid > 0 {
		m.Uptime = now.Sub(proc.startTime)
	}

	proc.lastMetrics = m
	t.mu.Unlock()

	t.publish(&m)
}

// publish sends metrics to all subscribers.
func (t *Tracker) publish(m *domainmetrics.ProcessMetrics) {
	t.subsMu.RLock()
	defer t.subsMu.RUnlock()

	for ch := range t.subscribers {
		select {
		case ch <- *m:
		default:
			// Drop if subscriber is slow
		}
	}
}
