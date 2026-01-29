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
	// return configured multi-writer with all writers
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
	// write to each writer sequentially
	for _, w := range mw.writers {
		n, err = w.Write(p)
		// return on first error
		if err != nil {
			// propagate first error to caller
			return n, err
		}
	}

	// return total bytes written
	return len(p), nil
}

// Close closes all underlying writers and returns the first error encountered.
//
// Returns:
//   - error: nil on success, first error encountered on failure
func (mw *MultiWriter) Close() error {
	var firstErr error

	// close all writers and track first error
	for _, w := range mw.writers {
		// track first error encountered
		if err := w.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	// return first error or nil
	return firstErr
}
