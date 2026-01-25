// Package process provides domain entities and value objects for process lifecycle management.
package process

import "time"

// EventType represents the type of lifecycle event.
//
// EventType is used to categorize process lifecycle transitions such as
// starting, stopping, failing, restarting, and health status changes.
type EventType int

// Event type constants.
const (
	// EventStarted indicates the process has started.
	EventStarted EventType = iota
	// EventStopped indicates the process has stopped normally (exit code 0).
	EventStopped
	// EventFailed indicates the process has failed (non-zero exit code).
	EventFailed
	// EventRestarting indicates the process is restarting.
	EventRestarting
	// EventHealthy indicates the process became healthy.
	EventHealthy
	// EventUnhealthy indicates the process became unhealthy.
	EventUnhealthy
	// EventExhausted indicates max restart attempts have been reached.
	EventExhausted
)

// String returns the string representation of the event type.
//
// Returns:
//   - string: event type name
func (e EventType) String() string {
	// Switch on event type to return appropriate string representation.
	switch e {
	// Handle started event type.
	case EventStarted:
		// Return "started" for EventStarted.
		return "started"
	// Handle stopped event type.
	case EventStopped:
		// Return "stopped" for EventStopped.
		return "stopped"
	// Handle failed event type.
	case EventFailed:
		// Return "failed" for EventFailed.
		return "failed"
	// Handle restarting event type.
	case EventRestarting:
		// Return "restarting" for EventRestarting.
		return "restarting"
	// Handle healthy event type.
	case EventHealthy:
		// Return "healthy" for EventHealthy.
		return "healthy"
	// Handle unhealthy event type.
	case EventUnhealthy:
		// Return "unhealthy" for EventUnhealthy.
		return "unhealthy"
	// Handle exhausted event type.
	case EventExhausted:
		// Return "exhausted" for EventExhausted.
		return "exhausted"
	// Handle unknown or unrecognized event types.
	default:
		// Return "unknown" for any unrecognized event type.
		return "unknown"
	}
}

// Event represents a process lifecycle event.
//
// Event encapsulates all information about a lifecycle transition including
// the event type, process identification, and any associated error details.
type Event struct {
	// Type is the event type (started, stopped, failed, etc.).
	Type EventType
	// Process is the name of the process.
	Process string
	// PID is the process ID.
	PID int
	// ExitCode is the exit code if process exited.
	ExitCode int
	// Timestamp is when the event occurred.
	Timestamp time.Time
	// Error contains any error associated with the event.
	Error error
}

// NewEvent creates a new process event.
//
// Params:
//   - eventType: the type of lifecycle event
//   - processName: the name of the process
//   - pid: the process ID
//   - exitCode: the exit code if process exited
//   - err: any error associated with the event
//
// Returns:
//   - Event: the newly created process event
func NewEvent(eventType EventType, processName string, pid, exitCode int, err error) Event {
	// Return a new Event struct with all fields populated.
	return Event{
		Type:      eventType,
		Process:   processName,
		PID:       pid,
		ExitCode:  exitCode,
		Timestamp: time.Now(),
		Error:     err,
	}
}
