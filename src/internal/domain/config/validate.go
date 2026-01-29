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
	// check if services are configured
	if len(cfg.Services) == 0 {
		// return error when no services
		return ErrNoServices
	}

	seen := make(map[string]bool, len(cfg.Services))

	// validate each service
	for i := range cfg.Services {
		svc := &cfg.Services[i]

		// validate service configuration
		if err := validateService(svc); err != nil {
			// propagate validation error
			return fmt.Errorf("service %q: %w", svc.Name, err)
		}

		// check for duplicate service names
		if seen[svc.Name] {
			// return error on duplicate
			return fmt.Errorf("%w: %s", ErrDuplicateServiceName, svc.Name)
		}
		seen[svc.Name] = true
	}

	// validation passed
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
	// check if service has a name
	if svc.Name == "" {
		// return error when name is empty
		return ErrEmptyServiceName
	}

	// check if service has a command
	if svc.Command == "" {
		// return error when command is empty
		return ErrEmptyCommand
	}

	// validate each health check
	for i := range svc.HealthChecks {
		// validate health check configuration
		if err := validateHealthCheck(&svc.HealthChecks[i]); err != nil {
			// propagate validation error
			return err
		}
	}

	// validation passed
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
	// validate based on health check type
	switch hc.Type {
	// validate http health check
	case HealthCheckHTTP:
		// validate http check has endpoint
		if hc.Endpoint == "" {
			// return error when endpoint missing
			return ErrMissingHTTPEndpoint
		}
	// validate tcp health check
	case HealthCheckTCP:
		// validate tcp check has host
		if hc.Host == "" {
			// return error when host missing
			return ErrMissingTCPHost
		}
		// validate tcp check has port
		if hc.Port == 0 {
			// return error when port missing
			return ErrMissingTCPPort
		}
	// validate command health check
	case HealthCheckCommand:
		// validate command check has command
		if hc.Command == "" {
			// return error when command missing
			return ErrMissingHealthCommand
		}
	// handle unknown health check type
	default:
		// return error for unknown type
		return fmt.Errorf("%w: %s", ErrInvalidHealthCheckType, hc.Type)
	}

	// validation passed
	return nil
}
