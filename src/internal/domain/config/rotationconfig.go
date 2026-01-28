// Package config provides domain value objects for service configuration.
package config

// defaultMaxFilesToRetain defines the default number of rotated log files to keep.
const defaultMaxFilesToRetain int = 10

// RotationConfig defines log rotation settings.
// It controls file size limits, age limits, and compression for rotated logs.
type RotationConfig struct {
	// MaxSize specifies the maximum size of a log file before rotation (e.g., "100MB").
	MaxSize string
	// MaxAge specifies the maximum age of log files before deletion (e.g., "7d").
	MaxAge string
	// MaxFiles specifies the maximum number of rotated log files to retain.
	MaxFiles int
	// Compress indicates whether rotated log files should be gzip compressed.
	Compress bool
}

// DefaultRotationConfig returns a RotationConfig with sensible defaults.
//
// Returns:
//   - RotationConfig: a configuration with standard size limits and file retention.
func DefaultRotationConfig() RotationConfig {
	return RotationConfig{
		MaxSize:  "100MB",
		MaxFiles: defaultMaxFilesToRetain,
	}
}
