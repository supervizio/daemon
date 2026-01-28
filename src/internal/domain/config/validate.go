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
	if len(cfg.Services) == 0 {
		return ErrNoServices
	}

	seen := make(map[string]bool, len(cfg.Services))

	for i := range cfg.Services {
		svc := &cfg.Services[i]

		if err := validateService(svc); err != nil {
			return fmt.Errorf("service %q: %w", svc.Name, err)
		}

		if seen[svc.Name] {
			return fmt.Errorf("%w: %s", ErrDuplicateServiceName, svc.Name)
		}
		seen[svc.Name] = true
	}

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
	if svc.Name == "" {
		return ErrEmptyServiceName
	}

	if svc.Command == "" {
		return ErrEmptyCommand
	}

	for i := range svc.HealthChecks {
		if err := validateHealthCheck(&svc.HealthChecks[i]); err != nil {
			return err
		}
	}

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
	switch hc.Type {
	case HealthCheckHTTP:
		if hc.Endpoint == "" {
			return ErrMissingHTTPEndpoint
		}
	case HealthCheckTCP:
		if hc.Host == "" {
			return ErrMissingTCPHost
		}
		if hc.Port == 0 {
			return ErrMissingTCPPort
		}
	case HealthCheckCommand:
		if hc.Command == "" {
			return ErrMissingHealthCommand
		}
	default:
		return fmt.Errorf("%w: %s", ErrInvalidHealthCheckType, hc.Type)
	}

	return nil
}
