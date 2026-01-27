// Package logging provides domain types for daemon event logging.
package logging

// Writer is the port interface for log event writers.
// Infrastructure layer implements this interface for different output targets.
type Writer interface {
	// Write writes a log event to the output target.
	//
	// Params:
	//   - event: the log event to write.
	//
	// Returns:
	//   - error: nil on success, error on failure.
	Write(event LogEvent) error

	// Close closes the writer and releases any resources.
	//
	// Returns:
	//   - error: nil on success, error on failure.
	Close() error
}
