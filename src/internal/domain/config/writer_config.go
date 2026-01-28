// Package config provides domain value objects for service configuration.
package config

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
