package logging

// Logger is the port interface for daemon event logging.
// Infrastructure layer implements this interface to provide logging capabilities.
type Logger interface {
	// Log logs an event directly.
	//
	// Params:
	//   - event: the log event to write.
	Log(event LogEvent)

	// Debug logs a debug-level event.
	//
	// Params:
	//   - service: the service name (empty for daemon-level).
	//   - eventType: the event type.
	//   - message: the event message.
	//   - meta: optional metadata.
	Debug(service, eventType, message string, meta map[string]any)

	// Info logs an info-level event.
	//
	// Params:
	//   - service: the service name (empty for daemon-level).
	//   - eventType: the event type.
	//   - message: the event message.
	//   - meta: optional metadata.
	Info(service, eventType, message string, meta map[string]any)

	// Warn logs a warning-level event.
	//
	// Params:
	//   - service: the service name (empty for daemon-level).
	//   - eventType: the event type.
	//   - message: the event message.
	//   - meta: optional metadata.
	Warn(service, eventType, message string, meta map[string]any)

	// Error logs an error-level event.
	//
	// Params:
	//   - service: the service name (empty for daemon-level).
	//   - eventType: the event type.
	//   - message: the event message.
	//   - meta: optional metadata.
	Error(service, eventType, message string, meta map[string]any)

	// Close closes the logger and all underlying writers.
	//
	// Returns:
	//   - error: nil on success, error on failure.
	Close() error
}
