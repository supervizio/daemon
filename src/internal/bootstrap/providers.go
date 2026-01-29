// Package bootstrap provides Wire dependency injection for the daemon.
// This file contains custom providers that require conditional logic
// or special handling beyond simple constructor calls.
package bootstrap

import (
	"context"
	"time"

	appconfig "github.com/kodflow/daemon/internal/application/config"
	apphealth "github.com/kodflow/daemon/internal/application/health"
	appmetrics "github.com/kodflow/daemon/internal/application/metrics"
	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
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
	SetMetricsTracker(tracker appmetrics.ProcessTracker)
	SetEventHandler(handler appsupervisor.EventHandler)
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
	// check if process is PID 1
	if r.IsPID1() {
		// return reaper for PID 1
		return r
	}
	// return nil for non-PID1 processes
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
	// load config from path
	return loader.Load(configPath)
}

// NewApp creates the App struct from the supervisor.
// This is the final provider in the dependency graph.
//
// Params:
//   - sup: the configured supervisor instance (minimal interface).
//
// Returns:
//   - *App: the application container with all dependencies wired.
//
// Deprecated: Use NewAppWithHealth instead.
func NewApp(sup AppSupervisor) *App {
	// construct app with supervisor only
	return &App{
		Supervisor: sup,
		Cleanup:    nil,
	}
}

// ProvideProberFactory creates the health prober factory.
//
// Returns:
//   - *infrahealthcheck.Factory: the prober factory instance.
func ProvideProberFactory() *infrahealthcheck.Factory {
	// construct prober factory with default timeout
	return infrahealthcheck.NewFactory(defaultProbeTimeout)
}

// ProvideMetricsTracker creates a metrics tracker with a platform-specific collector.
//
// Params:
//   - collector: the process metrics collector.
//
// Returns:
//   - *appmetrics.Tracker: the metrics tracker instance.
func ProvideMetricsTracker(collector appmetrics.Collector) *appmetrics.Tracker {
	// construct tracker with platform collector
	return appmetrics.NewTracker(collector)
}

// NewAppWithHealth creates the App struct with health monitoring and metrics wired.
// This provider connects the health prober factory and metrics tracker to the supervisor,
// enabling health-probe-triggered restarts following the Kubernetes
// liveness probe pattern and process CPU/memory tracking.
//
// Params:
//   - sup: the configured supervisor instance (minimal interface).
//   - factory: the health prober factory.
//   - tracker: the metrics tracker for CPU/memory monitoring.
//   - cfg: the domain configuration for daemon logging.
//
// Returns:
//   - *App: the application container with health monitoring and metrics enabled.
func NewAppWithHealth(sup supervisorConfigurer, factory apphealth.Creator, tracker *appmetrics.Tracker, cfg *domainconfig.Config) *App {
	// configure supervisor with prober factory
	sup.SetProberFactory(factory)
	// configure supervisor with metrics tracker
	sup.SetMetricsTracker(tracker)

	// construct app with all components
	return &App{
		Supervisor:     sup,
		Config:         cfg,
		MetricsTracker: tracker,
		Cleanup:        nil,
	}
}
