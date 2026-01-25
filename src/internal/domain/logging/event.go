package logging

import "time"

// LogEvent represents a daemon event to be logged.
type LogEvent struct {
	// Timestamp is when the event occurred.
	Timestamp time.Time
	// Level is the severity level.
	Level Level
	// Service is the service name (empty for daemon-level events).
	Service string
	// EventType is the event type (e.g., "started", "stopped", "failed").
	EventType string
	// Message is a human-readable description.
	Message string
	// Metadata contains additional event data (PID, ExitCode, Error, etc.).
	Metadata map[string]any
}

// NewLogEvent creates a new LogEvent with the current timestamp.
//
// Params:
//   - level: the severity level.
//   - service: the service name (empty for daemon-level).
//   - eventType: the event type.
//   - message: the event message.
//
// Returns:
//   - LogEvent: the created event.
func NewLogEvent(level Level, service, eventType, message string) LogEvent {
	return LogEvent{
		Timestamp: time.Now(),
		Level:     level,
		Service:   service,
		EventType: eventType,
		Message:   message,
		Metadata:  make(map[string]any),
	}
}

// WithMeta returns a copy of the event with the specified metadata key-value pair added.
//
// Params:
//   - key: the metadata key.
//   - value: the metadata value.
//
// Returns:
//   - LogEvent: the event with the added metadata.
func (e LogEvent) WithMeta(key string, value any) LogEvent {
	// Create a copy of metadata to avoid mutating the original.
	newMeta := make(map[string]any, len(e.Metadata)+1)
	for k, v := range e.Metadata {
		newMeta[k] = v
	}
	newMeta[key] = value

	return LogEvent{
		Timestamp: e.Timestamp,
		Level:     e.Level,
		Service:   e.Service,
		EventType: e.EventType,
		Message:   e.Message,
		Metadata:  newMeta,
	}
}

// WithMetadata returns a copy of the event with all specified metadata added.
//
// Params:
//   - meta: the metadata map to add.
//
// Returns:
//   - LogEvent: the event with the added metadata.
func (e LogEvent) WithMetadata(meta map[string]any) LogEvent {
	if meta == nil {
		return e
	}

	// Create a copy of metadata to avoid mutating the original.
	newMeta := make(map[string]any, len(e.Metadata)+len(meta))
	for k, v := range e.Metadata {
		newMeta[k] = v
	}
	for k, v := range meta {
		newMeta[k] = v
	}

	return LogEvent{
		Timestamp: e.Timestamp,
		Level:     e.Level,
		Service:   e.Service,
		EventType: e.EventType,
		Message:   e.Message,
		Metadata:  newMeta,
	}
}
