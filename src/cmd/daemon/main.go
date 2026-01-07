// Package main provides the entry point for the daemon process supervisor.
// daemon is a PID1-capable process supervisor designed to run in containers
// and on Linux/BSD systems. It manages multiple services with health checks,
// restart policies, and log rotation.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kodflow/daemon/internal/config"
	"github.com/kodflow/daemon/internal/supervisor"
)

var (
	version   = "dev"
	configPath string
)

func main() {
	flag.StringVar(&configPath, "config", "/etc/daemon/config.yaml", "path to configuration file")
	showVersion := flag.Bool("version", false, "show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("daemon %s\n", version)
		os.Exit(0)
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create supervisor
	sup, err := supervisor.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create supervisor: %w", err)
	}

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	// Start supervisor
	if err := sup.Start(ctx); err != nil {
		return fmt.Errorf("failed to start supervisor: %w", err)
	}

	// Wait for signals
	for {
		select {
		case sig := <-sigCh:
			switch sig {
			case syscall.SIGHUP:
				// Reload configuration
				if err := sup.Reload(); err != nil {
					fmt.Fprintf(os.Stderr, "reload failed: %v\n", err)
				}
			case syscall.SIGTERM, syscall.SIGINT:
				// Graceful shutdown
				cancel()
				return sup.Stop()
			}
		case <-ctx.Done():
			return sup.Stop()
		}
	}
}
