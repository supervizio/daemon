// Package config provides domain value objects for service configuration.
package config

// FileWriterConfig defines configuration for file writers.
// It specifies the output path and rotation settings for plain text log files.
type FileWriterConfig struct {
	// Path specifies the log file path (relative to BaseDir or absolute).
	Path string
	// Rotation specifies log rotation settings.
	Rotation RotationConfig
}
