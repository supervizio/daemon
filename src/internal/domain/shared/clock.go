// Package shared provides common value objects and interfaces for the domain layer.
package shared

import "time"

// Nower is an interface for obtaining the current time.
// It allows time-dependent code to be tested with deterministic values.
// Implementations can return fixed times for testing or real system time.
type Nower interface {
	// Now returns the current time.
	Now() time.Time
}

// RealClock implements Nower using the system time.
// It is a stateless implementation that delegates to time.Now().
type RealClock struct{}

// NewRealClock creates a new RealClock instance.
//
// Returns:
//   - *RealClock: a new clock that returns system time.
func NewRealClock() *RealClock {
	// construct real clock instance
	return &RealClock{}
}

// Now returns the current system time.
//
// Returns:
//   - time.Time: the current time from the system clock.
func (RealClock) Now() time.Time {
	// delegate to system time
	return time.Now()
}

// DefaultClock is the default clock instance using system time.
var DefaultClock Nower = &RealClock{}
