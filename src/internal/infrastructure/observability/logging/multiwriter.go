// Package logging provides log management with rotation and capture.
// It handles automatic file rotation based on size and file count.
package logging

import (
	"io"
)

// MultiWriter writes to multiple writers.
// It duplicates output to all underlying writers for logging to multiple destinations.
type MultiWriter struct {
	// writers is the slice of writers to duplicate output to.
	writers []io.WriteCloser
}

// NewMultiWriter creates a writer that duplicates output to multiple writers.
//
// Params:
//   - writers: writers to duplicate output to
//
// Returns:
//   - *MultiWriter: new multi-writer instance
func NewMultiWriter(writers ...io.WriteCloser) *MultiWriter {
	return &MultiWriter{writers: writers}
}

// Write writes to all writers.
//
// Params:
//   - p: bytes to write
//
// Returns:
//   - int: number of bytes written
//   - error: nil on success, error on failure
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}

	return len(p), nil
}

// Close closes all underlying writers and returns the first error encountered.
//
// Returns:
//   - error: nil on success, first error encountered on failure
func (mw *MultiWriter) Close() error {
	var firstErr error

	for _, w := range mw.writers {
		if err := w.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
