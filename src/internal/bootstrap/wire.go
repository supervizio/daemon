//go:build wireinject

package bootstrap

import (
	"github.com/google/wire"
	appconfig "github.com/kodflow/daemon/internal/application/config"
	apphealth "github.com/kodflow/daemon/internal/application/health"
	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
	domainprocess "github.com/kodflow/daemon/internal/domain/process"
	infrahealthcheck "github.com/kodflow/daemon/internal/infrastructure/observability/healthcheck"
	infraconfig "github.com/kodflow/daemon/internal/infrastructure/persistence/config/yaml"
	"github.com/kodflow/daemon/internal/infrastructure/process/control"
	"github.com/kodflow/daemon/internal/infrastructure/process/credentials"
	"github.com/kodflow/daemon/internal/infrastructure/process/executor"
	infrareaper "github.com/kodflow/daemon/internal/infrastructure/process/reaper"
)

// InitializeApp creates the application with all dependencies wired.
// This function is the injector that Wire will generate code for.
//
// Params:
//   - configPath: the path to the YAML configuration file.
//
// Returns:
//   - *App: the fully wired application.
//   - error: any error during dependency construction.
func InitializeApp(configPath string) (*App, error) {
	wire.Build(
		// Infrastructure: Configuration loader.
		infraconfig.NewLoader,
		wire.Bind(new(appconfig.Loader), new(*infraconfig.Loader)),

		// Infrastructure: Process credentials manager.
		credentials.New,
		wire.Bind(new(credentials.CredentialManager), new(*credentials.Manager)),

		// Infrastructure: Process control.
		control.New,
		wire.Bind(new(control.ProcessControl), new(*control.Control)),

		// Infrastructure: Zombie reaper (conditional via ProvideReaper).
		infrareaper.New,
		wire.Bind(new(ReaperMinimal), new(*infrareaper.Reaper)),

		// Infrastructure: Process executor.
		executor.NewWithDeps,
		wire.Bind(new(domainprocess.Executor), new(*executor.Executor)),

		// Infrastructure: Health prober factory.
		ProvideProberFactory,
		wire.Bind(new(apphealth.Creator), new(*infrahealthcheck.Factory)),

		// Providers: Custom provider functions.
		ProvideReaper,
		LoadConfig,

		// Application: Supervisor orchestrates all services.
		appsupervisor.NewSupervisor,
		wire.Bind(new(supervisorConfigurer), new(*appsupervisor.Supervisor)),

		// Bootstrap: Final App struct with health monitoring.
		NewAppWithHealth,
	)
	return nil, nil
}
