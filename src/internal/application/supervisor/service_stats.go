// Package supervisor provides the application service for orchestrating multiple services.
// It manages the lifecycle of services including start, stop, restart, and reload operations.
package supervisor

// ServiceStats holds statistics for a single service.
// It tracks the number of starts, stops, failures, and restarts that have occurred
// during the lifetime of a service. These statistics are useful for monitoring
// service health and identifying problematic services that may be restarting frequently.
//
// Fields:
//   - StartCount: Number of times the service has been started.
//   - StopCount: Number of times the service has stopped normally.
//   - FailCount: Number of times the service has failed (non-zero exit or crash).
//   - RestartCount: Number of times the service has been automatically restarted.
type ServiceStats struct {
	// StartCount is the number of times the service has started.
	StartCount int
	// StopCount is the number of times the service has stopped normally.
	StopCount int
	// FailCount is the number of times the service has failed.
	FailCount int
	// RestartCount is the number of times the service has restarted.
	RestartCount int
}

// NewServiceStats creates a new ServiceStats instance with zero values.
//
// Returns:
//   - *ServiceStats: a new ServiceStats instance initialized to zero.
func NewServiceStats() *ServiceStats {
	// Return a new ServiceStats with default zero values.
	return &ServiceStats{}
}
