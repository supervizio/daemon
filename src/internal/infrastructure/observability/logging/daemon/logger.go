// Package daemon provides daemon event logging infrastructure.
// It implements the domain logging interfaces with multiple output writers.
package daemon

import (
	"sync"

	"github.com/kodflow/daemon/internal/domain/logging"
)

// MultiLogger aggregates multiple writers and dispatches events to all of them.
// It implements the logging.Logger interface.
type MultiLogger struct {
	mu      sync.RWMutex
	writers []logging.Writer
}

// New creates a new MultiLogger with the specified writers.
//
// Params:
//   - writers: the writers to dispatch events to.
//
// Returns:
//   - *MultiLogger: the created multi-logger.
func New(writers ...logging.Writer) *MultiLogger {
	// create logger with provided writers
	return &MultiLogger{
		writers: writers,
	}
}

// NewMultiLogger creates a new multi-logger.
// Alias for New for KTN-CTOR compliance.
//
// Params:
//   - writers: the writers to dispatch events to.
//
// Returns:
//   - *MultiLogger: the created multi-logger.
func NewMultiLogger(writers ...logging.Writer) *MultiLogger {
	// delegate to standard constructor
	return New(writers...)
}

// Log logs an event to all writers.
//
// Params:
//   - event: the log event to write.
func (l *MultiLogger) Log(event logging.LogEvent) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Write event to all writers (best effort, ignore individual errors).
	// dispatch event to all writers
	for _, w := range l.writers {
		_ = w.Write(event)
	}
}

// Debug logs a debug-level event.
//
// Params:
//   - service: the service name (empty for daemon-level).
//   - eventType: the event type.
//   - message: the event message.
//   - meta: optional metadata.
func (l *MultiLogger) Debug(service, eventType, message string, meta map[string]any) {
	event := logging.NewLogEvent(logging.LevelDebug, service, eventType, message).
		WithMetadata(meta)
	l.Log(event)
}

// Info logs an info-level event.
//
// Params:
//   - service: the service name (empty for daemon-level).
//   - eventType: the event type.
//   - message: the event message.
//   - meta: optional metadata.
func (l *MultiLogger) Info(service, eventType, message string, meta map[string]any) {
	event := logging.NewLogEvent(logging.LevelInfo, service, eventType, message).
		WithMetadata(meta)
	l.Log(event)
}

// Warn logs a warning-level event.
//
// Params:
//   - service: the service name (empty for daemon-level).
//   - eventType: the event type.
//   - message: the event message.
//   - meta: optional metadata.
func (l *MultiLogger) Warn(service, eventType, message string, meta map[string]any) {
	event := logging.NewLogEvent(logging.LevelWarn, service, eventType, message).
		WithMetadata(meta)
	l.Log(event)
}

// Error logs an error-level event.
//
// Params:
//   - service: the service name (empty for daemon-level).
//   - eventType: the event type.
//   - message: the event message.
//   - meta: optional metadata.
func (l *MultiLogger) Error(service, eventType, message string, meta map[string]any) {
	event := logging.NewLogEvent(logging.LevelError, service, eventType, message).
		WithMetadata(meta)
	l.Log(event)
}

// AddWriter adds a writer to the logger at runtime.
// This is useful for adding TUI writers after initial setup.
//
// Params:
//   - w: the writer to add.
func (l *MultiLogger) AddWriter(w logging.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writers = append(l.writers, w)
}

// Close closes all writers.
//
// Returns:
//   - error: the first error encountered, or nil if all closed successfully.
func (l *MultiLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var firstErr error
	// close all writers and track first error
	for _, w := range l.writers {
		// track first error encountered
		if err := w.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	// return first error or nil
	return firstErr
}

// Ensure MultiLogger implements logging.Logger.
var _ logging.Logger = (*MultiLogger)(nil)
