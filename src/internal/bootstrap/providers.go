package bootstrap

import (
	appconfig "github.com/kodflow/daemon/internal/application/config"
	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
	domainconfig "github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/lifecycle"
	infrareaper "github.com/kodflow/daemon/internal/infrastructure/process/reaper"
)

// ProvideReaper returns the zombie reaper only if running as PID 1.
// When not running as PID 1, zombie reaping is not needed and nil is returned.
//
// Params:
//   - r: the reaper instance from infrastructure.
//
// Returns:
//   - lifecycle.Reaper: the reaper if PID 1, nil otherwise.
func ProvideReaper(r *infrareaper.Reaper) lifecycle.Reaper {
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
// Params:
//   - sup: the configured supervisor instance.
//
// Returns:
//   - *App: the application container with all dependencies wired.
func NewApp(sup *appsupervisor.Supervisor) *App {
	// Return the App with supervisor and optional cleanup.
	return &App{
		Supervisor: sup,
		Cleanup:    nil, // No cleanup needed currently; add if resources require it.
	}
}
