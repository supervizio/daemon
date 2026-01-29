// Package health provides domain entities and value objects for health checking.
package health

import "time"

// Result represents the result of a health check.
// It contains the status, message, duration, timestamp, and any error.
type Result struct {
	// Status holds the health status outcome of the check.
	Status Status
	// Message contains a human-readable description of the check result.
	Message string
	// Duration records how long the health check took to complete.
	Duration time.Duration
	// Timestamp records when the health check was performed.
	Timestamp time.Time
	// Error holds any error that occurred during the health check.
	Error error
}

// NewHealthyResult creates a healthy result.
// It captures the current timestamp automatically.
//
// Params:
//   - message: human-readable description of the healthy state
//   - duration: how long the health check took to complete
//
// Returns:
//   - Result: a new result with StatusHealthy and current timestamp
func NewHealthyResult(message string, duration time.Duration) Result {
	// create healthy result with current timestamp
	return NewHealthyResultAt(message, duration, time.Now())
}

// NewHealthyResultAt creates a healthy result with a specific timestamp.
// This is useful for testing or reconstructing results from stored data.
//
// Params:
//   - message: human-readable description of the healthy state
//   - duration: how long the health check took to complete
//   - timestamp: the timestamp to use for this result
//
// Returns:
//   - Result: a new result with StatusHealthy and specified timestamp
func NewHealthyResultAt(message string, duration time.Duration, timestamp time.Time) Result {
	// return healthy result with all fields
	return Result{
		Status:    StatusHealthy,
		Message:   message,
		Duration:  duration,
		Timestamp: timestamp,
	}
}

// NewUnhealthyResult creates an unhealthy result.
// It captures the current timestamp and the error that caused the failure.
//
// Params:
//   - message: human-readable description of the unhealthy state
//   - duration: how long the health check took to complete
//   - err: the error that caused the health check to fail
//
// Returns:
//   - Result: a new result with StatusUnhealthy and current timestamp
func NewUnhealthyResult(message string, duration time.Duration, err error) Result {
	// create unhealthy result with current timestamp
	return NewUnhealthyResultAt(message, duration, err, time.Now())
}

// NewUnhealthyResultAt creates an unhealthy result with a specific timestamp.
// This is useful for testing or reconstructing results from stored data.
//
// Params:
//   - message: human-readable description of the unhealthy state
//   - duration: how long the health check took to complete
//   - err: the error that caused the health check to fail
//   - timestamp: the timestamp to use for this result
//
// Returns:
//   - Result: a new result with StatusUnhealthy and specified timestamp
func NewUnhealthyResultAt(message string, duration time.Duration, err error, timestamp time.Time) Result {
	// return unhealthy result with all fields
	return Result{
		Status:    StatusUnhealthy,
		Message:   message,
		Duration:  duration,
		Timestamp: timestamp,
		Error:     err,
	}
}
