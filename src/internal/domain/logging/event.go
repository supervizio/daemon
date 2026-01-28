// Package logging provides domain types for daemon event logging.
package logging

import (
	"maps"
	"time"
)

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
	// Uses any type to support diverse runtime metadata: primitives, errors, custom types.
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
		Metadata:  make(map[string]any, defaultMetadataCapacity),
	}
}

// WithMeta returns a copy of the event with the specified metadata key-value pair added.
//
// Params:
//   - key: the metadata key.
//   - value: the metadata value (any type for runtime flexibility: int, string, error, etc.).
//
// Returns:
//   - LogEvent: the event with the added metadata.
//
// Note: any type required - metadata values are runtime-determined and type-heterogeneous.
//
//nolint:ktn-interface-anyuse // any required: heterogeneous log metadata (primitives, errors, custom types)
func (e LogEvent) WithMeta(key string, value any) LogEvent {
	newMeta := maps.Clone(e.Metadata)
	if newMeta == nil {
		newMeta = make(map[string]any, defaultMetadataCapacity)
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
//   - meta: the metadata map to add (uses any type to support diverse runtime metadata).
//
// Returns:
//   - LogEvent: the event with the added metadata.
func (e LogEvent) WithMetadata(meta map[string]any) LogEvent {
	if meta == nil {
		return e
	}

	newMeta := maps.Clone(e.Metadata)
	if newMeta == nil {
		newMeta = make(map[string]any, len(meta))
	}
	maps.Copy(newMeta, meta)

	return LogEvent{
		Timestamp: e.Timestamp,
		Level:     e.Level,
		Service:   e.Service,
		EventType: e.EventType,
		Message:   e.Message,
		Metadata:  newMeta,
	}
}
