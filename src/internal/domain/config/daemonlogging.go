// Package config provides domain value objects for service configuration.
package config

// DaemonLogging defines daemon-level event logging configuration.
// It specifies writers and their individual settings.
type DaemonLogging struct {
	// Writers specifies the list of writer configurations.
	Writers []WriterConfig
}

// DefaultDaemonLogging returns a DaemonLogging with sensible defaults.
// Default is console only with INFO level, stdout for INFO/DEBUG and stderr for WARN/ERROR.
//
// Returns:
//   - DaemonLogging: default daemon logging configuration.
func DefaultDaemonLogging() DaemonLogging {
	return DaemonLogging{
		Writers: []WriterConfig{
			{Type: "console", Level: "info"},
		},
	}
}
