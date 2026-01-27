// Package logging provides domain types for daemon event logging.
package logging

import "time"

// defaultMetadataCapacity is the initial capacity for metadata maps.
// Preallocated for typical 2-4 metadata entries to reduce allocations.
const defaultMetadataCapacity int = 4

// LogEvent represents a daemon event to be logged.
//
// This entity captures all information about a daemon or service event,
// including timestamp, severity, service context, and arbitrary metadata.
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
	// Create event with preallocated metadata map.
	return LogEvent{
		Timestamp: time.Now(),
		Level:     level,
		Service:   service,
		EventType: eventType,
		Message:   message,
		Metadata:  make(map[string]any, defaultMetadataCapacity),
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
	// Copy existing metadata.
	for k, v := range e.Metadata {
		newMeta[k] = v
	}
	newMeta[key] = value

	// Return new event with updated metadata.
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
	// Return unchanged if no metadata to add.
	if meta == nil {
		// No changes needed.
		return e
	}

	// Create a copy of metadata to avoid mutating the original.
	newMeta := make(map[string]any, len(e.Metadata)+len(meta))
	// Copy existing metadata.
	for k, v := range e.Metadata {
		newMeta[k] = v
	}
	// Merge new metadata.
	for k, v := range meta {
		newMeta[k] = v
	}

	// Return new event with merged metadata.
	return LogEvent{
		Timestamp: e.Timestamp,
		Level:     e.Level,
		Service:   e.Service,
		EventType: e.EventType,
		Message:   e.Message,
		Metadata:  newMeta,
	}
}
