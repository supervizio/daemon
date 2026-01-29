// Package supervisor provides the application service for orchestrating multiple services.
// It manages the lifecycle of services including start, stop, restart, and reload operations.
package supervisor

import "sync/atomic"

// ServiceStats holds statistics for a single service using atomic counters.
// It tracks the number of starts, stops, failures, and restarts that have occurred
// during the lifetime of a service. These statistics are useful for monitoring
// service health and identifying problematic services that may be restarting frequently.
//
// All counters use atomic operations for lock-free thread safety.
//
// Fields:
//   - startCount: Number of times the service has been started.
//   - stopCount: Number of times the service has stopped normally.
//   - failCount: Number of times the service has failed (non-zero exit or crash).
//   - restartCount: Number of times the service has been automatically restarted.
type ServiceStats struct {
	startCount   atomic.Int64
	stopCount    atomic.Int64
	failCount    atomic.Int64
	restartCount atomic.Int64
}

// NewServiceStats creates a new ServiceStats instance with zero values.
//
// Returns:
//   - *ServiceStats: a new ServiceStats instance initialized to zero.
func NewServiceStats() *ServiceStats {
	// construct empty stats
	return &ServiceStats{}
}

// IncrementStart atomically increments the start counter.
func (s *ServiceStats) IncrementStart() {
	// increment start count
	s.startCount.Add(1)
}

// IncrementStop atomically increments the stop counter.
func (s *ServiceStats) IncrementStop() {
	// increment stop count
	s.stopCount.Add(1)
}

// IncrementFail atomically increments the fail counter.
func (s *ServiceStats) IncrementFail() {
	// increment fail count
	s.failCount.Add(1)
}

// IncrementRestart atomically increments the restart counter.
func (s *ServiceStats) IncrementRestart() {
	// increment restart count
	s.restartCount.Add(1)
}

// StartCount returns the current start count.
//
// Returns:
//   - int: the number of times the service has been started.
func (s *ServiceStats) StartCount() int {
	// load and return start count
	return int(s.startCount.Load())
}

// StopCount returns the current stop count.
//
// Returns:
//   - int: the number of times the service has stopped normally.
func (s *ServiceStats) StopCount() int {
	// load and return stop count
	return int(s.stopCount.Load())
}

// FailCount returns the current fail count.
//
// Returns:
//   - int: the number of times the service has failed.
func (s *ServiceStats) FailCount() int {
	// load and return fail count
	return int(s.failCount.Load())
}

// RestartCount returns the current restart count.
//
// Returns:
//   - int: the number of times the service has been automatically restarted.
func (s *ServiceStats) RestartCount() int {
	// load and return restart count
	return int(s.restartCount.Load())
}

// Snapshot returns a copy of all counters for safe reading.
// This is useful when you need all values at a consistent point in time.
//
// Returns:
//   - ServiceStatsSnapshot: a copy of all counter values.
func (s *ServiceStats) Snapshot() ServiceStatsSnapshot {
	// construct snapshot with current values
	return ServiceStatsSnapshot{
		StartCount:   int(s.startCount.Load()),
		StopCount:    int(s.stopCount.Load()),
		FailCount:    int(s.failCount.Load()),
		RestartCount: int(s.restartCount.Load()),
	}
}

// SnapshotPtr returns a pointer to a copy of all counters.
// Use this instead of &Snapshot() to avoid escape analysis issues
// where taking address of return value causes heap allocation.
//
// Returns:
//   - *ServiceStatsSnapshot: a pointer to a copy of all counter values.
func (s *ServiceStats) SnapshotPtr() *ServiceStatsSnapshot {
	// construct snapshot pointer with current values
	return &ServiceStatsSnapshot{
		StartCount:   int(s.startCount.Load()),
		StopCount:    int(s.stopCount.Load()),
		FailCount:    int(s.failCount.Load()),
		RestartCount: int(s.restartCount.Load()),
	}
}
