// Package metrics provides application services for process metrics tracking.
package metrics

import (
	"context"
	"sync"
	"time"
	"unsafe"

	domainmetrics "github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
)

// Default configuration values.
const (
	defaultCollectionInterval time.Duration = 5 * time.Second
	defaultSubscriberBuffer   int           = 64
	collectionTimeoutDivisor  int           = 2
	defaultProcessMapCap      int           = 16
	defaultSubscriberMapCap   int           = 4
	maxSubscribers            int           = 64
)

// Tracker implements ProcessTracker using infrastructure collectors.
//
// It periodically collects CPU and memory metrics for tracked processes,
// maintains process state, and publishes updates to subscribers.
// The collection loop runs in a background goroutine started by Start().
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
//
// Params:
//   - d: collection interval (must be > 0, ignored if <= 0)
//
// Returns:
//   - TrackerOption: option that sets the interval
func WithCollectionInterval(d time.Duration) TrackerOption {
	// Return option that sets interval if valid.
	return func(t *Tracker) {
		// Only set interval if positive.
		if d > 0 {
			t.interval = d
		}
	}
}

// NewTracker creates a new process metrics tracker.
//
// Params:
//   - collector: infrastructure adapter for collecting process metrics
//   - opts: optional configuration functions
//
// Returns:
//   - *Tracker: configured tracker instance
func NewTracker(collector Collector, opts ...TrackerOption) *Tracker {
	t := &Tracker{
		collector:   collector,
		processes:   make(map[string]*trackedProcess, defaultProcessMapCap),
		interval:    defaultCollectionInterval,
		subscribers: make(map[chan domainmetrics.ProcessMetrics]struct{}, defaultSubscriberMapCap),
	}

	// Apply all options.
	for _, opt := range opts {
		opt(t)
	}

	// Return configured tracker.
	return t
}

// Start begins the metrics collection loop in a background goroutine.
// The goroutine exits when ctx is cancelled or Stop() is called.
// Safe to call multiple times (idempotent).
//
// Params:
//   - ctx: parent context for lifecycle management; cancelled to stop collection
//
// Returns:
//   - error: always nil (reserved for future use)
func (t *Tracker) Start(ctx context.Context) error {
	t.mu.Lock()
	// Check if already running.
	if t.running {
		t.mu.Unlock()
		// Already running, no-op.
		return nil
	}
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.running = true
	t.mu.Unlock()

	go t.collectLoop()
	// Success.
	return nil
}

// Stop stops the metrics collection loop.
func (t *Tracker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if running.
	if !t.running {
		// Not running, no-op.
		return
	}

	// Cancel context to stop goroutine.
	if t.cancel != nil {
		t.cancel()
	}
	t.running = false
}

// Track starts tracking metrics for a service with the given PID.
//
// Params:
//   - serviceName: unique service identifier
//   - pid: process ID to track
//
// Returns:
//   - error: always nil (reserved for future use)
func (t *Tracker) Track(serviceName string, pid int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	existing, exists := t.processes[serviceName]
	// Check if service already tracked.
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

	// Success.
	return nil
}

// buildMetrics creates a ProcessMetrics from tracked process state.
// Must be called with t.mu held.
//
// Params:
//   - proc: tracked process state
//   - now: current timestamp
//
// Returns:
//   - ProcessMetrics: snapshot of process metrics
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

	// Calculate uptime if process is running.
	if !proc.startTime.IsZero() && proc.pid > 0 {
		m.Uptime = now.Sub(proc.startTime)
	}

	// Return built metrics.
	return m
}

// Untrack stops tracking metrics for a service.
//
// Params:
//   - serviceName: service to stop tracking
func (t *Tracker) Untrack(serviceName string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.processes, serviceName)
}

// Get returns the current metrics for a service.
//
// Params:
//   - serviceName: service to query
//
// Returns:
//   - ProcessMetrics: current metrics snapshot
//   - bool: true if service found, false otherwise
func (t *Tracker) Get(serviceName string) (domainmetrics.ProcessMetrics, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	proc, exists := t.processes[serviceName]
	// Check if service exists.
	if !exists {
		// Service not found.
		return domainmetrics.ProcessMetrics{}, false
	}

	// Return cached metrics.
	return proc.lastMetrics, true
}

// All returns metrics for all tracked services.
//
// Returns:
//   - []ProcessMetrics: metrics for all tracked processes
func (t *Tracker) All() []domainmetrics.ProcessMetrics {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]domainmetrics.ProcessMetrics, 0, len(t.processes))
	// Collect all lastMetrics.
	for _, proc := range t.processes {
		result = append(result, proc.lastMetrics)
	}
	// Return all metrics.
	return result
}

// Subscribe returns a channel that receives metrics updates.
// Returns nil if max subscribers limit is reached to prevent resource exhaustion.
//
// Returns:
//   - <-chan ProcessMetrics: receive-only channel for updates, or nil if limit reached.
func (t *Tracker) Subscribe() <-chan domainmetrics.ProcessMetrics {
	t.subsMu.Lock()
	// Enforce max subscribers limit to prevent resource exhaustion.
	if len(t.subscribers) >= maxSubscribers {
		t.subsMu.Unlock()
		return nil
	}
	ch := make(chan domainmetrics.ProcessMetrics, defaultSubscriberBuffer)
	t.subscribers[ch] = struct{}{}
	t.subsMu.Unlock()

	return ch
}

// Unsubscribe removes a subscription channel.
//
// Params:
//   - ch: channel to unsubscribe
func (t *Tracker) Unsubscribe(ch <-chan domainmetrics.ProcessMetrics) {
	// Get pointer value for channel identity comparison.
	// Uses unsafe.Pointer instead of reflect.ValueOf().Pointer() for efficiency.
	// Both <-chan and chan have the same underlying pointer representation.
	recvPtr := *(*uintptr)(unsafe.Pointer(&ch))

	t.subsMu.Lock()
	var bidirCh chan domainmetrics.ProcessMetrics
	var found bool

	// Find the bidirectional channel with matching pointer.
	for c := range t.subscribers {
		// Check if this channel's pointer matches the receive channel.
		if *(*uintptr)(unsafe.Pointer(&c)) == recvPtr {
			bidirCh = c
			found = true
			break
		}
	}

	// Remove subscriber if found.
	if found {
		delete(t.subscribers, bidirCh)
	}
	t.subsMu.Unlock()

	// Close channel outside lock to avoid blocking.
	if found {
		close(bidirCh)
	}
}

// UpdateState updates the state of a tracked process.
//
// Params:
//   - serviceName: service to update
//   - state: new process state
//   - lastError: error message (empty if no error)
func (t *Tracker) UpdateState(serviceName string, state process.State, lastError string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	proc, exists := t.processes[serviceName]
	// Check if service exists.
	if !exists {
		// Service not found, no-op.
		return
	}

	proc.state = state
	// Update error if provided.
	if lastError != "" {
		proc.lastError = lastError
	}
	// Clear PID if process failed or stopped.
	if state == process.StateFailed || state == process.StateStopped {
		proc.pid = 0
	}
	// Update lastMetrics with new state
	proc.lastMetrics = t.buildMetrics(proc, time.Now())
}

// UpdateHealth updates the health status of a tracked process.
//
// Params:
//   - serviceName: service to update
//   - healthy: new health status
func (t *Tracker) UpdateHealth(serviceName string, healthy bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	proc, exists := t.processes[serviceName]
	// Check if service exists.
	if !exists {
		// Service not found, no-op.
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

	// Main collection loop.
	for {
		select {
		case <-t.ctx.Done():
			// Context cancelled, exit.
			return
		case <-ticker.C:
			t.collectAll()
		}
	}
}

// collectAll collects metrics for all tracked processes.
func (t *Tracker) collectAll() {
	t.mu.Lock()
	// Collect process snapshots to avoid holding lock during collection.
	// Preallocate slice to avoid allocation in slices.Collect.
	processes := make([]*trackedProcess, 0, len(t.processes))
	for _, p := range t.processes {
		processes = append(processes, p)
	}
	t.mu.Unlock()

	// Collect metrics for each process.
	for _, proc := range processes {
		t.collectProcess(proc)
	}
}

// collectProcess collects metrics for a single process.
//
// Params:
//   - proc: process to collect metrics for
func (t *Tracker) collectProcess(proc *trackedProcess) {
	// Check if process has valid PID.
	if proc.pid <= 0 {
		t.updateProcessMetrics(proc, domainmetrics.ProcessCPU{}, domainmetrics.ProcessMemory{})
		// No PID, skip collection.
		return
	}

	ctx, cancel := context.WithTimeout(t.ctx, t.interval/time.Duration(collectionTimeoutDivisor))
	defer cancel()

	cpu, cpuErr := t.collector.CollectCPU(ctx, proc.pid)
	mem, memErr := t.collector.CollectMemory(ctx, proc.pid)

	// If both fail, process may have exited
	// Check if both collections failed.
	if cpuErr != nil && memErr != nil {
		t.UpdateState(proc.serviceName, process.StateFailed, "process not found")
		// Process likely dead.
		return
	}

	// Calculate CPU percentage using delta between snapshots.
	now := time.Now()
	if cpuErr == nil && !proc.prevCPUTime.IsZero() {
		cpu.UsagePercent = t.calculateCPUPercent(proc.prevCPU, cpu, proc.prevCPUTime, now)
	}

	// Store current CPU snapshot for next calculation.
	if cpuErr == nil {
		proc.prevCPU = cpu
		proc.prevCPUTime = now
	}

	t.updateProcessMetrics(proc, cpu, mem)
}

// calculateCPUPercent calculates CPU usage percentage from two snapshots.
// The formula compares the change in CPU jiffies over time.
//
// Params:
//   - prev: previous CPU snapshot
//   - curr: current CPU snapshot
//   - prevTime: time of previous snapshot
//   - currTime: time of current snapshot
//
// Returns:
//   - float64: CPU usage percentage (0-100 per core, can exceed 100 for multi-core)
func (t *Tracker) calculateCPUPercent(prev, curr domainmetrics.ProcessCPU, prevTime, currTime time.Time) float64 {
	// Calculate elapsed time in seconds.
	elapsed := currTime.Sub(prevTime).Seconds()
	if elapsed <= 0 {
		return 0
	}

	// Calculate total jiffies used (user + system) for both snapshots.
	prevTotal := prev.User + prev.System
	currTotal := curr.User + curr.System

	// Avoid underflow if counters wrapped or process restarted.
	if currTotal < prevTotal {
		return 0
	}

	// Calculate jiffies delta.
	delta := currTotal - prevTotal

	// Convert jiffies to seconds (assuming 100 Hz = 100 jiffies per second).
	// This is the standard USER_HZ on Linux.
	const jiffiesPerSecond float64 = 100.0
	cpuSeconds := float64(delta) / jiffiesPerSecond

	// Calculate percentage relative to elapsed wall time.
	// Result can exceed 100% for multi-threaded processes using multiple cores.
	percent := (cpuSeconds / elapsed) * 100.0

	// Clamp negative values to zero.
	if percent < 0 {
		return 0
	}

	return percent
}

// updateProcessMetrics updates and publishes metrics for a process.
//
// Params:
//   - proc: process to update
//   - cpu: collected CPU metrics
//   - mem: collected memory metrics
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

	// Calculate uptime if process is running.
	if !proc.startTime.IsZero() && proc.pid > 0 {
		m.Uptime = now.Sub(proc.startTime)
	}

	proc.lastMetrics = m
	t.mu.Unlock()

	t.publish(&m)
}

// publish sends metrics to all subscribers.
//
// Params:
//   - m: metrics to publish
func (t *Tracker) publish(m *domainmetrics.ProcessMetrics) {
	t.subsMu.RLock()
	defer t.subsMu.RUnlock()

	// Send to all subscribers.
	for ch := range t.subscribers {
		select {
		case ch <- *m:
		default:
			// Drop if subscriber is slow
		}
	}
}
