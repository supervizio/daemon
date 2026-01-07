// Package service provides domain value objects for service configuration.
package service

import "github.com/kodflow/daemon/internal/domain/shared"

// Default health check configuration values.
const (
	// defaultCheckInterval is the default interval between health checks (30 seconds).
	defaultCheckInterval int = 30
	// defaultCheckTimeout is the default timeout for health check responses (5 seconds).
	defaultCheckTimeout int = 5
	// defaultCheckRetries is the default number of failures before marking unhealthy.
	defaultCheckRetries int = 3
	// defaultHTTPStatusOK is the default expected HTTP status code for successful checks.
	defaultHTTPStatusOK int = 200
)

// HealthCheckConfig defines a health check for a service.
// It supports HTTP, TCP, and command-based health check types.
type HealthCheckConfig struct {
	// Name is an optional identifier for this health check.
	Name string
	// Type specifies the health check type (http, tcp, or command).
	Type HealthCheckType
	// Interval specifies the time between consecutive health checks.
	Interval shared.Duration
	// Timeout specifies the maximum time to wait for a health check response.
	Timeout shared.Duration
	// Retries specifies the number of failures before marking unhealthy.
	Retries int
	// Endpoint specifies the HTTP endpoint URL for HTTP health checks.
	Endpoint string
	// Method specifies the HTTP method for HTTP health checks (GET, POST, etc.).
	Method string
	// StatusCode specifies the expected HTTP status code for success.
	StatusCode int
	// Host specifies the hostname or IP for TCP health checks.
	Host string
	// Port specifies the port number for TCP health checks.
	Port int
	// Command specifies the command to execute for command health checks.
	Command string
}

// HealthCheckType defines the type of health check.
type HealthCheckType string

// Health check type constants.
const (
	// HealthCheckHTTP performs HTTP endpoint checks.
	HealthCheckHTTP HealthCheckType = "http"
	// HealthCheckTCP performs TCP connection checks.
	HealthCheckTCP HealthCheckType = "tcp"
	// HealthCheckCommand executes a command and checks its exit code.
	HealthCheckCommand HealthCheckType = "command"
)

// String returns the string representation of the health check type.
//
// Returns:
//   - string: the health check type as a string value.
func (t HealthCheckType) String() string {
	// Convert the type to its underlying string value.
	return string(t)
}

// IsHTTP returns true if the health check type is HTTP.
//
// Returns:
//   - bool: true if the type equals HealthCheckHTTP, false otherwise.
func (t HealthCheckType) IsHTTP() bool {
	// Compare against the HTTP constant.
	return t == HealthCheckHTTP
}

// IsTCP returns true if the health check type is TCP.
//
// Returns:
//   - bool: true if the type equals HealthCheckTCP, false otherwise.
func (t HealthCheckType) IsTCP() bool {
	// Compare against the TCP constant.
	return t == HealthCheckTCP
}

// IsCommand returns true if the health check type is command.
//
// Returns:
//   - bool: true if the type equals HealthCheckCommand, false otherwise.
func (t HealthCheckType) IsCommand() bool {
	// Compare against the command constant.
	return t == HealthCheckCommand
}

// DefaultHealthCheckConfig returns a HealthCheckConfig with sensible defaults.
//
// Params:
//   - checkType: the type of health check to configure (http, tcp, or command).
//
// Returns:
//   - HealthCheckConfig: a configuration struct with default values applied.
func DefaultHealthCheckConfig(checkType HealthCheckType) HealthCheckConfig {
	// Return a new config with default values for interval, timeout, retries, and HTTP settings.
	return HealthCheckConfig{
		Type:       checkType,
		Interval:   shared.Seconds(defaultCheckInterval),
		Timeout:    shared.Seconds(defaultCheckTimeout),
		Retries:    defaultCheckRetries,
		Method:     "GET",
		StatusCode: defaultHTTPStatusOK,
	}
}
