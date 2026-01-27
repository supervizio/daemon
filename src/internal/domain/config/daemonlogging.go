// Package config provides domain value objects for service configuration.
package config

// DaemonLogging defines daemon-level event logging configuration.
// It specifies writers and their individual settings.
type DaemonLogging struct {
	// Writers specifies the list of writer configurations.
	Writers []WriterConfig
}

// WriterConfig defines configuration for a single log writer.
// It supports multiple writer types (console, file, json) with individual level filtering.
type WriterConfig struct {
	// Type specifies the writer type: "console", "file", "json".
	Type string
	// Level specifies the minimum log level for this writer.
	Level string
	// File contains file writer specific configuration.
	File FileWriterConfig
	// JSON contains JSON writer specific configuration.
	JSON JSONWriterConfig
}

// FileWriterConfig defines configuration for file writers.
// It specifies the output path and rotation settings for plain text log files.
type FileWriterConfig struct {
	// Path specifies the log file path (relative to BaseDir or absolute).
	Path string
	// Rotation specifies log rotation settings.
	Rotation RotationConfig
}

// JSONWriterConfig defines configuration for JSON writers.
// It specifies the output path and rotation settings for structured JSON log files.
type JSONWriterConfig struct {
	// Path specifies the JSON log file path (relative to BaseDir or absolute).
	Path string
	// Rotation specifies log rotation settings.
	Rotation RotationConfig
}

// DefaultDaemonLogging returns a DaemonLogging with sensible defaults.
// Default is console only with INFO level, stdout for INFO/DEBUG and stderr for WARN/ERROR.
//
// Returns:
//   - DaemonLogging: default daemon logging configuration.
func DefaultDaemonLogging() DaemonLogging {
	// Return default configuration with console writer at INFO level.
	return DaemonLogging{
		Writers: []WriterConfig{
			{Type: "console", Level: "info"},
		},
	}
}
