// Package health provides infrastructure adapters for health checking.
// It implements the health check interfaces for HTTP, TCP, and command checks.
package health

import (
	"errors"
	"fmt"

	apphealth "github.com/kodflow/daemon/internal/application/health"
	"github.com/kodflow/daemon/internal/domain/service"
)

// ErrUnknownHealthCheckType indicates an unknown health check type was specified.
// This error is returned when a configuration contains an unsupported health check type.
var ErrUnknownHealthCheckType error = errors.New("unknown health check type")

// Checker is the interface that all health checkers implement.
// This is an alias to the application port for convenience.
type Checker = apphealth.Checker

// NewChecker creates a checker from configuration.
// It acts as a factory function that instantiates the appropriate checker type.
//
// Params:
//   - cfg: the health check configuration containing type and parameters.
//
// Returns:
//   - Checker: the created health checker instance.
//   - error: an error if the health check type is unknown.
func NewChecker(cfg *service.HealthCheckConfig) (Checker, error) {
	// Switch on the health check type to create the appropriate checker.
	switch cfg.Type {
	// Handle HTTP health checks.
	case service.HealthCheckHTTP:
		// Return a new HTTP checker for endpoint monitoring.
		return NewHTTPChecker(cfg), nil
	// Handle TCP health checks.
	case service.HealthCheckTCP:
		// Return a new TCP checker for port connectivity.
		return NewTCPChecker(cfg), nil
	// Handle command health checks.
	case service.HealthCheckCommand:
		// Return a new command checker for script execution.
		return NewCommandChecker(cfg), nil
	// Handle unknown health check types.
	default:
		// Return an error for unsupported types.
		return nil, fmt.Errorf("%w: %s", ErrUnknownHealthCheckType, cfg.Type)
	}
}

// Factory implements Creator for creating health checkers.
// It provides a reusable factory pattern for health check instantiation.
type Factory struct{}

// NewFactory creates a new checker factory.
// It initializes and returns an empty factory instance.
//
// Returns:
//   - *Factory: a new factory instance ready for use.
func NewFactory() *Factory {
	// Return a new empty factory instance.
	return &Factory{}
}

// Create creates a checker from the given configuration.
// It converts the application config to domain config and delegates to NewChecker.
//
// Params:
//   - cfg: the application-level checker configuration.
//
// Returns:
//   - Checker: the created health checker instance.
//   - error: an error if the health check type is unknown.
func (f *Factory) Create(cfg apphealth.CheckerConfig) (Checker, error) {
	// Convert app config to domain config.
	domainCfg := &service.HealthCheckConfig{
		Name:       cfg.Name,
		Type:       service.HealthCheckType(cfg.Type),
		Endpoint:   cfg.Endpoint,
		Method:     cfg.Method,
		StatusCode: cfg.StatusCode,
		Host:       cfg.Host,
		Port:       cfg.Port,
		Command:    cfg.Command,
	}

	// Delegate to NewChecker for actual creation.
	return NewChecker(domainCfg)
}
