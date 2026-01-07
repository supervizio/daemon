// Package health provides the application service for health monitoring.
// It defines ports for health checking infrastructure adapters.
package health

import (
	"context"

	domain "github.com/kodflow/daemon/internal/domain/health"
)

// Checker performs health checks for a service.
// This is the port that infrastructure adapters implement to provide
// specific health check mechanisms (HTTP, TCP, command).
type Checker interface {
	// Check performs a health check and returns the result.
	//
	// Params:
	//   - ctx: the context for cancellation and timeout control.
	//
	// Returns:
	//   - domain.Result: the result of the health check.
	Check(ctx context.Context) domain.Result

	// Name returns the name of this checker.
	//
	// Returns:
	//   - string: the unique name identifying this checker.
	Name() string

	// Type returns the type of this checker (http, tcp, command).
	//
	// Returns:
	//   - string: the type of health check performed.
	Type() string
}

// Creator creates health checkers from configuration.
// It abstracts the creation of different checker types based on
// the provided configuration.
type Creator interface {
	// Create creates a checker from the given configuration.
	//
	// Params:
	//   - cfg: the configuration for the checker to create.
	//
	// Returns:
	//   - Checker: the created checker instance.
	//   - error: an error if the checker could not be created.
	Create(cfg CheckerConfig) (Checker, error)
}

// CheckerConfig contains the configuration needed to create a checker.
// This is a simplified view of the domain service.HealthCheckConfig
// containing only the fields relevant to checker creation.
type CheckerConfig struct {
	// Name is the checker name.
	Name string
	// Type is the checker type.
	Type string
	// Endpoint for HTTP checks.
	Endpoint string
	// Method for HTTP checks.
	Method string
	// StatusCode expected for HTTP checks.
	StatusCode int
	// Host for TCP checks.
	Host string
	// Port for TCP checks.
	Port int
	// Command for command checks.
	Command string
	// TimeoutSeconds for checks.
	TimeoutSeconds int
}
