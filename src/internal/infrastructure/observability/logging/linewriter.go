// Package logging provides linewriter.go implementing a line-buffered writer with prefix.
// It buffers input and writes complete lines with the configured prefix.
package logging

import (
	"bytes"
	"io"
)

// LineWriter constants for line buffering and processing.
const (
	// newlineChar is the newline character used to detect line endings.
	newlineChar byte = '\n'
	// indexNotFound is the sentinel value indicating no newline was found in buffer.
	indexNotFound int = -1
	// zeroBytes is the return value when no bytes were written due to an error.
	zeroBytes int = 0
	// defaultWriteBufCapacity is the default capacity for the write buffer to handle typical line lengths.
	defaultWriteBufCapacity int = 256
)

// LineWriter writes lines with optional prefix.
// It buffers input and writes complete lines with the configured prefix.
// Uses a reusable write buffer to minimize allocations.
type LineWriter struct {
	// writer is the underlying writer that receives formatted output.
	writer io.Writer
	// prefix is the string prepended to each line.
	prefix string
	// prefixBytes is the prefix as bytes (cached to avoid conversion).
	prefixBytes []byte
	// buf holds incomplete line data until a newline is received.
	buf []byte
	// writeBuf is a reusable buffer for batching prefix + line writes.
	writeBuf []byte
}

// NewLineWriter creates a writer that prefixes each line.
// It wraps an existing writer and adds prefix support with line buffering.
// Pre-allocates buffers for efficient writes.
//
// Params:
//   - w: the underlying writer to write to.
//   - prefix: the string to prepend to each line.
//
// Returns:
//   - *LineWriter: the initialized line writer instance.
func NewLineWriter(w io.Writer, prefix string) *LineWriter {
	// Return a new LineWriter with pre-allocated write buffer.
	return &LineWriter{
		writer:      w,
		prefix:      prefix,
		prefixBytes: []byte(prefix),
		writeBuf:    make([]byte, 0, defaultWriteBufCapacity), // Pre-allocate for typical line length.
	}
}

// Write implements io.Writer with line buffering.
// It buffers data until complete lines are available, then writes them with prefix.
// Batches prefix + line into a single write call for efficiency.
//
// Params:
//   - p: the byte slice to write.
//
// Returns:
//   - int: the number of bytes from p that were processed.
//   - error: an error if writing to the underlying writer fails.
func (lw *LineWriter) Write(p []byte) (n int, err error) {
	lw.buf = append(lw.buf, p...)

	// Iterate through the buffer to find and process complete lines.
	for {
		// Use bytes.IndexByte for optimized newline search (assembly-optimized).
		idx := bytes.IndexByte(lw.buf, newlineChar)

		// Check if no newline was found, meaning no complete line is available.
		if idx == indexNotFound {
			break
		}

		line := lw.buf[:idx+1]
		lw.buf = lw.buf[idx+1:]

		// Batch prefix + line into single write to reduce syscalls.
		if len(lw.prefixBytes) > 0 {
			// Reuse write buffer to avoid allocations.
			lw.writeBuf = lw.writeBuf[:0]
			lw.writeBuf = append(lw.writeBuf, lw.prefixBytes...)
			lw.writeBuf = append(lw.writeBuf, line...)
			// Check if write operation succeeded for prefixed line.
			if _, err := lw.writer.Write(lw.writeBuf); err != nil {
				// Write failed - return zero bytes and propagate error.
				return zeroBytes, err
			}
		} else {
			// No prefix, write line directly.
			// Check if write operation succeeded for plain line.
			if _, err := lw.writer.Write(line); err != nil {
				// Write failed - return zero bytes and propagate error.
				return zeroBytes, err
			}
		}
	}

	// Return the total number of bytes processed from the input.
	return len(p), nil
}

// Flush writes any remaining buffered data.
// It ensures all data is written even if no trailing newline was received.
// Batches prefix + content + newline into a single write call.
//
// Returns:
//   - error: an error if writing to the underlying writer fails.
func (lw *LineWriter) Flush() error {
	// Check if there is any buffered data remaining to be flushed.
	if len(lw.buf) > zeroBytes {
		// Batch prefix + content + newline into single write.
		lw.writeBuf = lw.writeBuf[:0]
		// Check if prefix should be prepended to the flushed content.
		if len(lw.prefixBytes) > 0 {
			lw.writeBuf = append(lw.writeBuf, lw.prefixBytes...)
		}
		lw.writeBuf = append(lw.writeBuf, lw.buf...)
		lw.writeBuf = append(lw.writeBuf, newlineChar)

		// Check if write operation succeeded for the flushed content.
		if _, err := lw.writer.Write(lw.writeBuf); err != nil {
			// Write failed - return error to caller.
			return err
		}
		lw.buf = nil
	}
	// Return nil indicating successful flush or no data to flush.
	return nil
}
