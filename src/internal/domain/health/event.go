// Package health provides domain entities and value objects for health checking.
package health

import "time"

// Event represents a health check event.
// It is sent when a checker's status changes.
type Event struct {
	// Checker holds the name of the checker that generated this event.
	Checker string
	// Status holds the health status reported by the checker.
	Status Status
	// Result contains the full result details from the check.
	Result Result
	// Timestamp records when this event was generated.
	Timestamp time.Time
}

// NewEvent creates a new health event.
// It captures the current timestamp automatically.
//
// Params:
//   - checker: name of the health checker that generated this event
//   - status: current health status from the checker
//   - result: full result details from the health check
//
// Returns:
//   - Event: a new event with the current timestamp
func NewEvent(checker string, status Status, result Result) Event {
	// Build and return the event with current timestamp.
	return Event{
		Checker:   checker,
		Status:    status,
		Result:    result,
		Timestamp: time.Now(),
	}
}
