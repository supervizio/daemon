// Package config provides domain value objects for service configuration.
package config

// JSONWriterConfig defines configuration for JSON writers.
// It specifies the output path and rotation settings for structured JSON log files.
type JSONWriterConfig struct {
	// Path specifies the JSON log file path (relative to BaseDir or absolute).
	Path string
	// Rotation specifies log rotation settings.
	Rotation RotationConfig
}
