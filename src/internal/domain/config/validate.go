// Package config provides domain value objects for service configuration.
package config

import (
	"errors"
	"fmt"
)

// Validation errors.
var (
	// ErrNoServices indicates no services are configured.
	ErrNoServices error = errors.New("no services configured")
	// ErrEmptyServiceName indicates a service has no name.
	ErrEmptyServiceName error = errors.New("service name is required")
	// ErrEmptyCommand indicates a service has no command.
	ErrEmptyCommand error = errors.New("service command is required")
	// ErrDuplicateServiceName indicates duplicate service names.
	ErrDuplicateServiceName error = errors.New("duplicate service name")
	// ErrInvalidHealthCheckType indicates an invalid health check type.
	ErrInvalidHealthCheckType error = errors.New("invalid health check type")
	// ErrMissingHTTPEndpoint indicates HTTP check missing endpoint.
	ErrMissingHTTPEndpoint error = errors.New("http health check requires endpoint")
	// ErrMissingTCPHost indicates TCP check missing host.
	ErrMissingTCPHost error = errors.New("tcp health check requires host")
	// ErrMissingTCPPort indicates TCP check missing port.
	ErrMissingTCPPort error = errors.New("tcp health check requires port")
	// ErrMissingHealthCommand indicates command check missing command.
	ErrMissingHealthCommand error = errors.New("command health check requires command")
)

// Validate validates the configuration.
//
// Params:
//   - cfg: configuration to validate
//
// Returns:
//   - error: validation error if any
func Validate(cfg *Config) error {
	// Check if at least one service is configured.
	if len(cfg.Services) == 0 {
		// Return error when no services are defined.
		return ErrNoServices
	}

	seen := make(map[string]bool, len(cfg.Services))

	// Iterate through all services to validate each one.
	for i := range cfg.Services {
		svc := &cfg.Services[i]

		// Validate the service configuration.
		if err := validateService(svc); err != nil {
			// Return wrapped error with service name context.
			return fmt.Errorf("service %q: %w", svc.Name, err)
		}

		// Check for duplicate service names.
		if seen[svc.Name] {
			// Return error for duplicate service name.
			return fmt.Errorf("%w: %s", ErrDuplicateServiceName, svc.Name)
		}
		seen[svc.Name] = true
	}

	// Return nil when all validations pass.
	return nil
}

// validateService validates a single service configuration.
//
// Params:
//   - svc: service configuration to validate
//
// Returns:
//   - error: validation error if any
func validateService(svc *ServiceConfig) error {
	// Check if service name is provided.
	if svc.Name == "" {
		// Return error when service name is empty.
		return ErrEmptyServiceName
	}

	// Check if service command is provided.
	if svc.Command == "" {
		// Return error when service command is empty.
		return ErrEmptyCommand
	}

	// Iterate through all health checks to validate each one.
	for i := range svc.HealthChecks {
		// Validate the health check configuration.
		if err := validateHealthCheck(&svc.HealthChecks[i]); err != nil {
			// Return the health check validation error.
			return err
		}
	}

	// Return nil when all validations pass.
	return nil
}

// validateHealthCheck validates a health check configuration.
//
// Params:
//   - hc: health check configuration to validate
//
// Returns:
//   - error: validation error if any
func validateHealthCheck(hc *HealthCheckConfig) error {
	// Switch on health check type to apply type-specific validation.
	switch hc.Type {
	// Handle HTTP health check type.
	case HealthCheckHTTP:
		// Check if HTTP endpoint is provided.
		if hc.Endpoint == "" {
			// Return error when HTTP endpoint is missing.
			return ErrMissingHTTPEndpoint
		}
	// Handle TCP health check type.
	case HealthCheckTCP:
		// Check if TCP host is provided.
		if hc.Host == "" {
			// Return error when TCP host is missing.
			return ErrMissingTCPHost
		}
		// Check if TCP port is provided.
		if hc.Port == 0 {
			// Return error when TCP port is missing.
			return ErrMissingTCPPort
		}
	// Handle command health check type.
	case HealthCheckCommand:
		// Check if health command is provided.
		if hc.Command == "" {
			// Return error when health command is missing.
			return ErrMissingHealthCommand
		}
	// Handle unknown health check type.
	default:
		// Return error for invalid health check type.
		return fmt.Errorf("%w: %s", ErrInvalidHealthCheckType, hc.Type)
	}

	// Return nil when validation passes.
	return nil
}
