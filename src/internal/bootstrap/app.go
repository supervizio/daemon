// Package bootstrap provides dependency injection wiring using Google Wire.
// It isolates all dependency construction from the main entry point,
// allowing for a minimal main.go and better testability.
package bootstrap

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
)

var (
	// version is the application version, set at build time via ldflags.
	version string = "dev"
	// configPath is the path to the YAML configuration file.
	configPath string = ""
)

// App holds all application dependencies injected by Wire.
// It is the root object of the dependency graph.
type App struct {
	// Supervisor is the main service orchestrator.
	Supervisor *appsupervisor.Supervisor
	// Cleanup is the cleanup function for all resources.
	Cleanup func()
}

// signalHandler defines the interface for supervisor signal handling operations.
type signalHandler interface {
	Reload() error
	Stop() error
}

// Run is the main entry point called from cmd/daemon/main.go.
// It parses flags, initializes the application via Wire, and runs the main loop.
//
// Returns:
//   - int: exit code (0 for success, 1 for error).
func Run() int {
	flag.StringVar(&configPath, "config", "/etc/daemon/config.yaml", "path to configuration file")
	showVersion := flag.Bool("version", false, "show version and exit")
	flag.Parse()

	// Check if version flag was provided to display version and exit.
	if *showVersion {
		fmt.Printf("daemon %s\n", version)
		return 0
	}

	// Run the main application logic and handle any errors.
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

// run executes the main application logic.
//
// Returns:
//   - error: nil on success, error on failure.
func run() error {
	// Initialize the application using Wire-generated code.
	app, err := InitializeApp(configPath)
	// Check if initialization failed.
	if err != nil {
		// Return error with context about initialization failure.
		return fmt.Errorf("failed to initialize: %w", err)
	}
	// Ensure cleanup is called when run exits.
	if app.Cleanup != nil {
		defer app.Cleanup()
	}

	// Setup context and signals for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	// Start supervisor and check for startup errors.
	if err := app.Supervisor.Start(ctx); err != nil {
		// Return error with context about supervisor startup failure.
		return fmt.Errorf("failed to start supervisor: %w", err)
	}

	// Return the result of waiting for signals.
	return waitForSignals(ctx, cancel, sigCh, app.Supervisor)
}

// waitForSignals handles OS signals in a continuous loop until shutdown.
//
// Params:
//   - ctx: the context for cancellation.
//   - cancel: the cancel function for the context.
//   - sigCh: the channel receiving OS signals.
//   - sup: the signal handler interface for supervisor operations.
//
// Returns:
//   - error: nil on success, error on failure.
func waitForSignals(ctx context.Context, cancel context.CancelFunc, sigCh <-chan os.Signal, sup signalHandler) error {
	// Loop forever until a shutdown signal is received.
	for {
		// Select on signal channel or context done.
		select {
		case sig := <-sigCh:
			// Switch on the received signal type.
			switch sig {
			// Case for SIGHUP signal to reload configuration.
			case syscall.SIGHUP:
				// Check if reload operation failed.
				if err := sup.Reload(); err != nil {
					fmt.Fprintf(os.Stderr, "reload failed: %v\n", err)
				}
			// Case for SIGTERM or SIGINT signals to initiate shutdown.
			case syscall.SIGTERM, syscall.SIGINT:
				cancel()
				// Return the result of stopping the supervisor.
				return sup.Stop()
			}
		case <-ctx.Done():
			// Return the result of stopping the supervisor when context is done.
			return sup.Stop()
		}
	}
}
