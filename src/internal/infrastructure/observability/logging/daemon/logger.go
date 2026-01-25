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
	// mu protects concurrent access to the writers slice.
	mu sync.RWMutex
	// writers is the list of writers to dispatch events to.
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
	return &MultiLogger{
		writers: writers,
	}
}

// Log logs an event to all writers.
//
// Params:
//   - event: the log event to write.
func (l *MultiLogger) Log(event logging.LogEvent) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, w := range l.writers {
		// Ignore errors from individual writers - best effort logging.
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

// Close closes all writers.
//
// Returns:
//   - error: the first error encountered, or nil if all closed successfully.
func (l *MultiLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var firstErr error
	for _, w := range l.writers {
		if err := w.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Ensure MultiLogger implements logging.Logger.
var _ logging.Logger = (*MultiLogger)(nil)
