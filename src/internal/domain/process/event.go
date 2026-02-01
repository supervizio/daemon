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
	// match event type to string representation
	switch e {
	// started event type
	case EventStarted:
		// return started string
		return "started"
	// stopped event type
	case EventStopped:
		// return stopped string
		return "stopped"
	// failed event type
	case EventFailed:
		// return failed string
		return "failed"
	// restarting event type
	case EventRestarting:
		// return restarting string
		return "restarting"
	// healthy event type
	case EventHealthy:
		// return healthy string
		return "healthy"
	// unhealthy event type
	case EventUnhealthy:
		// return unhealthy string
		return "unhealthy"
	// exhausted event type
	case EventExhausted:
		// return exhausted string
		return "exhausted"
	// unknown event type
	default:
		// return unknown string
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
	// construct event with all fields populated
	return Event{
		Type:      eventType,
		Process:   processName,
		PID:       pid,
		ExitCode:  exitCode,
		Timestamp: time.Now(),
		Error:     err,
	}
}
