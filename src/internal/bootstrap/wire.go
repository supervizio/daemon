//go:build wireinject

package bootstrap

import (
	"github.com/google/wire"
	appconfig "github.com/kodflow/daemon/internal/application/config"
	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
	domainprocess "github.com/kodflow/daemon/internal/domain/process"
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
		wire.Bind(new(reaper), new(*infrareaper.Reaper)),

		// Infrastructure: Process executor.
		executor.NewWithDeps,
		wire.Bind(new(domainprocess.Executor), new(*executor.Executor)),

		// Providers: Custom provider functions.
		ProvideReaper,
		LoadConfig,

		// Application: Supervisor orchestrates all services.
		appsupervisor.NewSupervisor,

		// Bootstrap: Final App struct.
		NewApp,
	)
	return nil, nil
}
