package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate checks the configuration for errors.
func Validate(cfg *Config) error {
	var errs []error

	if len(cfg.Services) == 0 {
		errs = append(errs, ValidationError{
			Field:   "services",
			Message: "at least one service must be defined",
		})
	}

	serviceNames := make(map[string]bool)
	for i, svc := range cfg.Services {
		prefix := fmt.Sprintf("services[%d]", i)

		if err := validateService(&svc, prefix, serviceNames); err != nil {
			errs = append(errs, err)
		}

		if svc.Name != "" {
			serviceNames[svc.Name] = true
		}
	}

	// Validate dependencies exist
	for _, svc := range cfg.Services {
		for _, dep := range svc.DependsOn {
			if !serviceNames[dep] {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("services.%s.depends_on", svc.Name),
					Message: fmt.Sprintf("dependency '%s' not found", dep),
				})
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// validateService validates a single service configuration.
func validateService(svc *ServiceConfig, prefix string, existing map[string]bool) error {
	var errs []error

	if svc.Name == "" {
		errs = append(errs, ValidationError{
			Field:   prefix + ".name",
			Message: "name is required",
		})
	} else if existing[svc.Name] {
		errs = append(errs, ValidationError{
			Field:   prefix + ".name",
			Message: fmt.Sprintf("duplicate service name: %s", svc.Name),
		})
	}

	if svc.Command == "" {
		errs = append(errs, ValidationError{
			Field:   prefix + ".command",
			Message: "command is required",
		})
	}

	if err := validateRestartPolicy(svc.Restart.Policy, prefix+".restart.policy"); err != nil {
		errs = append(errs, err)
	}

	for j, hc := range svc.HealthChecks {
		hcPrefix := fmt.Sprintf("%s.health_checks[%d]", prefix, j)
		if err := validateHealthCheck(&hc, hcPrefix); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// validateRestartPolicy validates a restart policy value.
func validateRestartPolicy(policy RestartPolicy, field string) error {
	switch policy {
	case RestartAlways, RestartOnFailure, RestartNever, RestartUnless:
		return nil
	case "":
		return nil // Will use default
	default:
		return ValidationError{
			Field:   field,
			Message: fmt.Sprintf("invalid restart policy: %s (must be always, on-failure, never, or unless-stopped)", policy),
		}
	}
}

// validateHealthCheck validates a health check configuration.
func validateHealthCheck(hc *HealthCheckConfig, prefix string) error {
	var errs []error

	switch hc.Type {
	case HealthCheckHTTP:
		if err := validateHTTPHealthCheck(hc, prefix); err != nil {
			errs = append(errs, err)
		}
	case HealthCheckTCP:
		if err := validateTCPHealthCheck(hc, prefix); err != nil {
			errs = append(errs, err)
		}
	case HealthCheckCommand:
		if err := validateCommandHealthCheck(hc, prefix); err != nil {
			errs = append(errs, err)
		}
	case "":
		errs = append(errs, ValidationError{
			Field:   prefix + ".type",
			Message: "type is required (http, tcp, or command)",
		})
	default:
		errs = append(errs, ValidationError{
			Field:   prefix + ".type",
			Message: fmt.Sprintf("invalid health check type: %s", hc.Type),
		})
	}

	if hc.Interval == 0 {
		errs = append(errs, ValidationError{
			Field:   prefix + ".interval",
			Message: "interval is required",
		})
	}

	if hc.Timeout == 0 {
		errs = append(errs, ValidationError{
			Field:   prefix + ".timeout",
			Message: "timeout is required",
		})
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// validateHTTPHealthCheck validates HTTP health check specific fields.
func validateHTTPHealthCheck(hc *HealthCheckConfig, prefix string) error {
	if hc.Endpoint == "" {
		return ValidationError{
			Field:   prefix + ".endpoint",
			Message: "endpoint is required for HTTP health checks",
		}
	}

	u, err := url.Parse(hc.Endpoint)
	if err != nil {
		return ValidationError{
			Field:   prefix + ".endpoint",
			Message: fmt.Sprintf("invalid URL: %v", err),
		}
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return ValidationError{
			Field:   prefix + ".endpoint",
			Message: "endpoint must be http or https",
		}
	}

	return nil
}

// validateTCPHealthCheck validates TCP health check specific fields.
func validateTCPHealthCheck(hc *HealthCheckConfig, prefix string) error {
	var errs []error

	if hc.Host == "" {
		errs = append(errs, ValidationError{
			Field:   prefix + ".host",
			Message: "host is required for TCP health checks",
		})
	}

	if hc.Port <= 0 || hc.Port > 65535 {
		errs = append(errs, ValidationError{
			Field:   prefix + ".port",
			Message: "port must be between 1 and 65535",
		})
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// validateCommandHealthCheck validates command health check specific fields.
func validateCommandHealthCheck(hc *HealthCheckConfig, prefix string) error {
	if strings.TrimSpace(hc.Command) == "" {
		return ValidationError{
			Field:   prefix + ".command",
			Message: "command is required for command health checks",
		}
	}
	return nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	return Validate(c)
}
