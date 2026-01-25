// Package daemon provides daemon event logging infrastructure.
package daemon

import (
	"sync"

	"github.com/kodflow/daemon/internal/domain/logging"
)

// BufferedWriter buffers log events until Flush is called.
// This is useful for delaying console output until after the MOTD banner is displayed.
type BufferedWriter struct {
	// mu protects concurrent access.
	mu sync.Mutex
	// inner is the underlying writer to delegate to after flush.
	inner logging.Writer
	// buffer holds events until Flush is called.
	buffer []logging.LogEvent
	// flushed indicates whether Flush has been called.
	flushed bool
}

// NewBufferedWriter creates a new buffered writer wrapping the given writer.
//
// Params:
//   - inner: the underlying writer to delegate to after flush.
//
// Returns:
//   - *BufferedWriter: the created buffered writer.
func NewBufferedWriter(inner logging.Writer) *BufferedWriter {
	return &BufferedWriter{
		inner:  inner,
		buffer: make([]logging.LogEvent, 0, 64),
	}
}

// Write buffers or writes an event depending on whether Flush has been called.
//
// Params:
//   - event: the log event to write.
//
// Returns:
//   - error: nil on success, error on failure (only after flush).
func (w *BufferedWriter) Write(event logging.LogEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.flushed {
		// After flush, write directly to inner writer.
		return w.inner.Write(event)
	}

	// Before flush, buffer the event.
	w.buffer = append(w.buffer, event)
	return nil
}

// Flush writes all buffered events to the inner writer and enables direct writes.
// After calling Flush, all subsequent Write calls go directly to the inner writer.
//
// Returns:
//   - error: the first error encountered during flush, or nil if all succeeded.
func (w *BufferedWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.flushed {
		// Already flushed.
		return nil
	}

	// Write all buffered events.
	var firstErr error
	for _, event := range w.buffer {
		if err := w.inner.Write(event); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	// Clear buffer and mark as flushed.
	w.buffer = nil
	w.flushed = true

	return firstErr
}

// Close closes the inner writer.
//
// Returns:
//   - error: error from closing the inner writer, or nil.
func (w *BufferedWriter) Close() error {
	return w.inner.Close()
}

// Ensure BufferedWriter implements logging.Writer.
var _ logging.Writer = (*BufferedWriter)(nil)
