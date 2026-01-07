// Package service provides domain value objects for service configuration.
package service

// ServiceLogging defines per-service logging configuration.
// It configures separate settings for stdout and stderr streams.
type ServiceLogging struct {
	// Stdout configures logging for the service's standard output stream.
	Stdout LogStreamConfig
	// Stderr configures logging for the service's standard error stream.
	Stderr LogStreamConfig
}
