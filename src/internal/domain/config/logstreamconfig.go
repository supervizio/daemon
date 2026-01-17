// Package config provides domain value objects for service configuration.
package config

// LogStreamConfig defines configuration for a log stream.
// It specifies file path, timestamp format, and rotation settings.
type LogStreamConfig struct {
	// FilePath specifies the path to the log file for this stream.
	FilePath string
	// Format specifies the Go time format string for timestamps.
	Format string
	// RotationConfig defines log rotation settings for this stream.
	RotationConfig RotationConfig
}

// File returns the log file path.
//
// Returns:
//   - string: the configured file path for this log stream.
func (l *LogStreamConfig) File() string {
	// Return the configured file path.
	return l.FilePath
}

// TimestampFormat returns the timestamp format.
//
// Returns:
//   - string: the Go time format string for log timestamps.
func (l *LogStreamConfig) TimestampFormat() string {
	// Return the configured timestamp format.
	return l.Format
}

// Rotation returns the rotation configuration.
//
// Returns:
//   - RotationConfig: the log rotation settings for this stream.
func (l *LogStreamConfig) Rotation() RotationConfig {
	// Return the configured rotation settings.
	return l.RotationConfig
}

// NewLogStreamConfig creates a new LogStreamConfig with the given file path.
//
// Params:
//   - filePath: the path to the log file for this stream.
//
// Returns:
//   - LogStreamConfig: a log stream configuration with the given file path.
func NewLogStreamConfig(filePath string) LogStreamConfig {
	// Return a new LogStreamConfig with the specified file path.
	return LogStreamConfig{
		FilePath:       filePath,
		RotationConfig: DefaultRotationConfig(),
	}
}
