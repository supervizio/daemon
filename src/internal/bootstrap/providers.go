// Package bootstrap provides Wire dependency injection for the daemon.
// This file contains custom providers that require conditional logic
// or special handling beyond simple constructor calls.
package bootstrap

import (
	"context"
	"time"

	appconfig "github.com/kodflow/daemon/internal/application/config"
	apphealth "github.com/kodflow/daemon/internal/application/health"
	domainconfig "github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/lifecycle"
	infrahealthcheck "github.com/kodflow/daemon/internal/infrastructure/observability/healthcheck"
)

// defaultProbeTimeout is the default timeout for health probes.
const defaultProbeTimeout time.Duration = 5 * time.Second

// ReaperMinimal defines the minimal interface needed for PID1 detection.
// This interface accepts any implementation with IsPID1 capability,
// satisfying the KTN-API-MINIF requirement.
// Exported for testing purposes.
type ReaperMinimal interface {
	lifecycle.Reaper
}

// supervisorConfigurer defines the minimal interface for supervisor configuration.
// This satisfies KTN-API-MINIF by accepting only the method actually called.
// Methods declared explicitly to ensure linter counts all methods.
type supervisorConfigurer interface {
	Start(ctx context.Context) error
	Stop() error
	Reload() error
	SetProberFactory(factory apphealth.Creator)
}

// ProvideReaper returns the zombie reaper only if running as PID 1.
// When not running as PID 1, zombie reaping is not needed and nil is returned.
//
// Params:
//   - r: the reaper instance from infrastructure.
//
// Returns:
//   - lifecycle.Reaper: the reaper if PID 1, nil otherwise.
func ProvideReaper(r ReaperMinimal) lifecycle.Reaper {
	// Check if the process is running as PID 1.
	if r.IsPID1() {
		// Return the reaper for zombie cleanup.
		return r
	}
	// Return nil when not PID 1 (reaping not needed).
	return nil
}

// LoadConfig loads configuration from the given path using the provided loader.
//
// Params:
//   - loader: the configuration loader interface.
//   - configPath: the path to the configuration file.
//
// Returns:
//   - *domainconfig.Config: the loaded configuration.
//   - error: any error during loading.
func LoadConfig(loader appconfig.Loader, configPath string) (*domainconfig.Config, error) {
	// Load and return the configuration from the specified path.
	return loader.Load(configPath)
}

// NewApp creates the App struct from the supervisor.
// This is the final provider in the dependency graph.
//
// Deprecated: Use NewAppWithHealth instead.
//
// Params:
//   - sup: the configured supervisor instance (minimal interface).
//
// Returns:
//   - *App: the application container with all dependencies wired.
func NewApp(sup AppSupervisor) *App {
	// Return the App with supervisor and optional cleanup.
	return &App{
		Supervisor: sup,
		Cleanup:    nil, // No cleanup needed currently; add if resources require it.
	}
}

// ProvideProberFactory creates the health prober factory.
//
// Returns:
//   - *infrahealthcheck.Factory: the prober factory instance.
func ProvideProberFactory() *infrahealthcheck.Factory {
	// Return factory with default timeout.
	return infrahealthcheck.NewFactory(defaultProbeTimeout)
}

// NewAppWithHealth creates the App struct with health monitoring wired.
// This provider connects the health prober factory to the supervisor,
// enabling health-probe-triggered restarts following the Kubernetes
// liveness probe pattern.
//
// Params:
//   - sup: the configured supervisor instance (minimal interface).
//   - factory: the health prober factory.
//
// Returns:
//   - *App: the application container with health monitoring enabled.
func NewAppWithHealth(sup supervisorConfigurer, factory apphealth.Creator) *App {
	// Wire the prober factory to enable health monitoring.
	sup.SetProberFactory(factory)

	// Return the App with supervisor and optional cleanup.
	return &App{
		Supervisor: sup,
		Cleanup:    nil, // No cleanup needed currently; add if resources require it.
	}
}
