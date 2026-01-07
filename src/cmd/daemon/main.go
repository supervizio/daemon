// Package main provides the entry point for the superviz.io process supervisor.
// It uses Domain-Driven Design (DDD) architecture with clean layer separation.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
	infraconfig "github.com/kodflow/daemon/internal/infrastructure/config/yaml"
	"github.com/kodflow/daemon/internal/infrastructure/kernel"
	infraprocess "github.com/kodflow/daemon/internal/infrastructure/process"
)

// signalHandler defines the interface for supervisor signal handling operations.
type signalHandler interface {
	Reload() error
	Stop() error
}

var (
	// version is the application version, set at build time via ldflags.
	version string = "dev"
	// configPath is the path to the YAML configuration file.
	configPath string = ""
)

// main is the entry point for the superviz.io process supervisor.
//
// Returns:
//   - none: this function does not return.
func main() {
	flag.StringVar(&configPath, "config", "/etc/daemon/config.yaml", "path to configuration file")
	showVersion := flag.Bool("version", false, "show version and exit")
	flag.Parse()

	// Check if version flag was provided to display version and exit.
	if *showVersion {
		fmt.Printf("daemon %s\n", version)
		os.Exit(0)
	}

	// Run the main supervisor logic and handle any errors.
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the main supervisor logic.
//
// Returns:
//   - error: nil on success, error on failure.
func run() error {
	// Infrastructure: Create YAML config loader.
	loader := infraconfig.NewLoader()

	// Infrastructure: Create Unix process executor.
	executor := infraprocess.NewUnixExecutor()

	// Infrastructure: Get zombie reaper if PID 1.
	reaper := kernel.Default.Reaper
	// Check if the process is not running as PID 1.
	if !reaper.IsPID1() {
		reaper = nil
	}

	// Application: Load configuration via loader port.
	cfg, err := loader.Load(configPath)
	// Check if configuration loading failed.
	if err != nil {
		// Return error with context about configuration loading failure.
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Application: Create supervisor with injected dependencies.
	sup, err := appsupervisor.NewSupervisor(cfg, loader, executor, reaper)
	// Check if supervisor creation failed.
	if err != nil {
		// Return error with context about supervisor creation failure.
		return fmt.Errorf("failed to create supervisor: %w", err)
	}

	// Setup context and signals for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	// Start supervisor and check for startup errors.
	if err := sup.Start(ctx); err != nil {
		// Return error with context about supervisor startup failure.
		return fmt.Errorf("failed to start supervisor: %w", err)
	}

	// Return the result of waiting for signals.
	return waitForSignals(ctx, cancel, sigCh, sup)
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
