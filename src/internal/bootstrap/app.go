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
	"path/filepath"
	"syscall"
	"time"

	appmetrics "github.com/kodflow/daemon/internal/application/metrics"
	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
	domainconfig "github.com/kodflow/daemon/internal/domain/config"
	domainhealth "github.com/kodflow/daemon/internal/domain/health"
	domainlogging "github.com/kodflow/daemon/internal/domain/logging"
	domainprocess "github.com/kodflow/daemon/internal/domain/process"
	daemonlogger "github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

var (
	// version is the application version, set at build time via ldflags.
	version string = "dev"
	// configPath is the path to the YAML configuration file.
	configPath string = ""
)

// AppSupervisor defines the interface for supervisor operations used by the application.
// This allows for minimal interface usage in providers (KTN-API-MINIF).
// Methods are declared explicitly to ensure linter counts all methods.
type AppSupervisor interface {
	Start(ctx context.Context) error
	Stop() error
	Reload() error
	SetEventHandler(handler appsupervisor.EventHandler)
}

// App holds all application dependencies injected by Wire.
// It is the root object of the dependency graph.
type App struct {
	// Supervisor is the main service orchestrator.
	Supervisor AppSupervisor
	// Config holds the domain configuration for daemon logging.
	Config *domainconfig.Config
	// MetricsTracker tracks process CPU and memory metrics.
	MetricsTracker *appmetrics.Tracker
	// Cleanup is the cleanup function for all resources.
	Cleanup func()
}

// SignalHandler defines the interface for supervisor signal handling operations.
// Exported for testing purposes.
type SignalHandler interface {
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
	forceInteractive := flag.Bool("tui", false, "enable interactive TUI mode")
	flag.Parse()

	// Check if version flag was provided to display version and exit.
	if *showVersion {
		fmt.Printf("supervizio %s\n", version)
		// Return success exit code after displaying version.
		return 0
	}

	// Determine TUI mode: raw by default, interactive only with --tui flag.
	tuiMode := determineTUIMode(*forceInteractive)

	// Run the main application logic and handle any errors.
	if err := run(configPath, tuiMode); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		// Return failure exit code due to application error.
		return 1
	}
	// Return success exit code after clean shutdown.
	return 0
}

// determineTUIMode determines the TUI mode based on flags.
// Raw mode is the default; interactive mode requires explicit --tui flag.
func determineTUIMode(forceInteractive bool) tui.Mode {
	if forceInteractive {
		return tui.ModeInteractive
	}
	// Raw mode is the default (no TTY auto-detection).
	return tui.ModeRaw
}

// RunWithConfig executes the main application logic with a specified config path.
// This function is exported for testing purposes.
//
// Params:
//   - cfgPath: the path to the configuration file.
//
// Returns:
//   - error: nil on success, error on failure.
func RunWithConfig(cfgPath string) error {
	// Delegate to internal run function with default mode (raw).
	return run(cfgPath, determineTUIMode(false))
}

// TUISnapshotsProvider provides service snapshots for TUI.
type TUISnapshotsProvider interface {
	ServiceSnapshotsForTUI() []appsupervisor.ServiceSnapshotForTUI
}

// supervisorServiceProvider wraps a supervisor to provide services to TUI.
type supervisorServiceProvider struct {
	sup TUISnapshotsProvider
}

// Services implements tui.ServiceProvider.
// Services are returned in definition order (as configured), not sorted.
func (p *supervisorServiceProvider) Services() []model.ServiceSnapshot {
	snapshots := p.sup.ServiceSnapshotsForTUI()
	result := make([]model.ServiceSnapshot, 0, len(snapshots))

	for _, snap := range snapshots {
		// Convert listeners.
		listeners := make([]model.ListenerSnapshot, 0, len(snap.Listeners))
		for _, l := range snap.Listeners {
			listeners = append(listeners, model.ListenerSnapshot{
				Name:      l.Name,
				Port:      l.Port,
				Protocol:  l.Protocol,
				Exposed:   l.Exposed,
				Listening: l.Listening,
				Status:    model.PortStatus(l.StatusInt),
			})
		}

		result = append(result, model.ServiceSnapshot{
			Name:            snap.Name,
			State:           domainprocess.State(snap.StateInt),
			Health:          domainhealth.Status(snap.HealthInt),
			HasHealthChecks: snap.HasHealthChecks,
			PID:             snap.PID,
			Uptime:          time.Duration(snap.UptimeSecs) * time.Second,
			CPUPercent:      snap.CPUPercent,
			MemoryRSS:       snap.MemoryRSS,
			RestartCount:    snap.RestartCount,
			Ports:           snap.Ports,
			Listeners:       listeners,
		})
	}

	// Keep definition order - no sorting (stable display, values update in place).
	return result
}

// run executes the main application logic with the specified TUI mode.
//
// Params:
//   - cfgPath: the path to the configuration file.
//   - tuiMode: the TUI display mode (raw or interactive).
//
// Returns:
//   - error: nil on success, error on failure.
func run(cfgPath string, tuiMode tui.Mode) error {
	// Initialize the application using Wire-generated code.
	app, err := InitializeApp(cfgPath)
	// Check if initialization failed.
	if err != nil {
		// Return error with context about initialization failure.
		return fmt.Errorf("failed to initialize: %w", err)
	}
	// Ensure cleanup is called when run exits.
	if app.Cleanup != nil {
		defer app.Cleanup()
	}

	// Create log adapter for TUI (captures logs for display).
	logAdapter := tui.NewLogAdapter()

	// Load recent log history from file if available (tail -n 100).
	if logFilePath := findLogFilePath(app.Config); logFilePath != "" {
		if err := logAdapter.LoadLogHistory(logFilePath, 100); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load log history: %v\n", err)
		}
	}

	// Build daemon event logger based on TUI mode.
	// In interactive mode: no console output (TUI handles display).
	// In raw mode: buffered console output (logs displayed after MOTD banner).
	var logger domainlogging.Logger
	var bufferedConsole *daemonlogger.BufferedWriter
	var loggerErr error

	if tuiMode == tui.ModeInteractive {
		// Interactive mode: only file writers + TUI writer (no console pollution).
		logger, loggerErr = daemonlogger.BuildLoggerWithoutConsole(
			app.Config.Logging.Daemon,
			app.Config.Logging.BaseDir,
		)
	} else {
		// Raw mode: use buffered console writer so logs appear after MOTD banner.
		logger, bufferedConsole, loggerErr = daemonlogger.BuildLoggerWithBufferedConsole(
			app.Config.Logging.Daemon,
			app.Config.Logging.BaseDir,
		)
	}

	if loggerErr != nil {
		// Log warning but continue - daemon can run without event logging.
		fmt.Fprintf(os.Stderr, "warning: failed to build daemon logger: %v\n", loggerErr)
		if tuiMode == tui.ModeInteractive {
			// In interactive mode, use a silent logger (TUI only).
			logger = daemonlogger.NewSilentLogger()
		} else {
			logger = daemonlogger.DefaultLogger()
		}
	}
	defer func() { _ = logger.Close() }()

	// Add TUI writer to logger to capture events for TUI display.
	if ml, ok := logger.(*daemonlogger.MultiLogger); ok {
		tuiWriter := tui.NewTUILogWriter(logAdapter)
		ml.AddWriter(tuiWriter)
	}

	// Set up event handler to log service events.
	app.Supervisor.SetEventHandler(func(serviceName string, event *domainprocess.Event, stats *appsupervisor.ServiceStatsSnapshot) {
		logEvent := convertProcessEventToLogEvent(serviceName, event, stats)
		logger.Log(logEvent)
	})

	// Setup context and signals for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	// Log daemon startup.
	logger.Info("", "daemon_started", "Supervisor started", map[string]any{"version": version})

	// Start supervisor and check for startup errors.
	if err := app.Supervisor.Start(ctx); err != nil {
		// Return error with context about supervisor startup failure.
		return fmt.Errorf("failed to start supervisor: %w", err)
	}

	// Start metrics tracker for CPU/memory monitoring.
	if app.MetricsTracker != nil {
		_ = app.MetricsTracker.Start(ctx)
		defer app.MetricsTracker.Stop()
	}

	// Create TUI with the configured mode.
	tuiConfig := tui.DefaultConfig(version)
	tuiConfig.Mode = tuiMode
	t := tui.New(tuiConfig)

	// Set service provider if supervisor supports ServiceSnapshotsForTUI.
	if sup, ok := app.Supervisor.(TUISnapshotsProvider); ok {
		// Create a service provider that queries the supervisor.
		provider := &supervisorServiceProvider{sup: sup}
		t.SetServiceProvider(provider)
	}

	// Set log adapter as health provider to display logs in TUI.
	t.SetHealthProvider(logAdapter)

	// Set config path for TUI display.
	t.SetConfigPath(cfgPath)

	// Run TUI based on mode.
	switch tuiMode {
	case tui.ModeRaw:
		// Raw mode: show MOTD once, then flush buffered logs and wait for signals.
		if err := t.Run(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "warning: TUI error: %v\n", err)
		}
		// Flush buffered logs after MOTD banner is displayed.
		if bufferedConsole != nil {
			_ = bufferedConsole.Flush()
		}
		// Continue to signal handling.
		return WaitForSignals(ctx, cancel, sigCh, app.Supervisor)

	case tui.ModeInteractive:
		// Interactive mode: run TUI in parallel with signal handling.
		tuiDone := make(chan error, 1)
		go func() {
			tuiDone <- t.Run(ctx)
		}()

		// Wait for TUI exit or signals.
		return waitForTUIOrSignals(ctx, cancel, sigCh, tuiDone, app.Supervisor)
	}

	// Fallback: wait for signals.
	return WaitForSignals(ctx, cancel, sigCh, app.Supervisor)
}

// waitForTUIOrSignals handles both TUI events and OS signals.
func waitForTUIOrSignals(ctx context.Context, cancel context.CancelFunc, sigCh <-chan os.Signal, tuiDone <-chan error, sup SignalHandler) error {
	for {
		select {
		case sig := <-sigCh:
			switch sig {
			case syscall.SIGHUP:
				if err := sup.Reload(); err != nil {
					fmt.Fprintf(os.Stderr, "reload failed: %v\n", err)
				}
			case syscall.SIGTERM, syscall.SIGINT:
				cancel()
				return sup.Stop()
			}
		case err := <-tuiDone:
			// TUI exited (user pressed q or error).
			if err != nil {
				fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			}
			// Stop the supervisor when TUI exits.
			cancel()
			return sup.Stop()
		case <-ctx.Done():
			return sup.Stop()
		}
	}
}

// WaitForSignals handles OS signals in a continuous loop until shutdown.
// Exported for testing purposes.
//
// Params:
//   - ctx: the context for cancellation.
//   - cancel: the cancel function for the context.
//   - sigCh: the channel receiving OS signals.
//   - sup: the signal handler interface for supervisor operations.
//
// Returns:
//   - error: nil on success, error on failure.
func WaitForSignals(ctx context.Context, cancel context.CancelFunc, sigCh <-chan os.Signal, sup SignalHandler) error {
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

// convertProcessEventToLogEvent converts a domain process event to a log event.
//
// Params:
//   - serviceName: the name of the service.
//   - event: the process event.
//   - stats: optional atomic service statistics snapshot for enriched logging.
//
// Returns:
//   - domainlogging.LogEvent: the converted log event.
func convertProcessEventToLogEvent(serviceName string, event *domainprocess.Event, stats *appsupervisor.ServiceStatsSnapshot) domainlogging.LogEvent {
	// Determine log level based on event type.
	var level domainlogging.Level
	switch event.Type {
	case domainprocess.EventFailed, domainprocess.EventUnhealthy:
		level = domainlogging.LevelWarn
	case domainprocess.EventExhausted:
		level = domainlogging.LevelError
	case domainprocess.EventStarted, domainprocess.EventStopped,
		domainprocess.EventRestarting, domainprocess.EventHealthy:
		level = domainlogging.LevelInfo
	}

	// Build event type string.
	eventType := event.Type.String()

	// Build message based on event type and exit code.
	var message string
	switch event.Type {
	case domainprocess.EventStarted:
		if stats != nil && stats.RestartCount > 0 {
			message = fmt.Sprintf("Service started (restart #%d)", stats.RestartCount)
		} else {
			message = "Service started"
		}
	case domainprocess.EventStopped:
		// Differentiate clean exit from non-clean exit.
		if event.ExitCode == 0 {
			message = "Service stopped cleanly"
		} else {
			message = fmt.Sprintf("Service exited with code %d", event.ExitCode)
		}
	case domainprocess.EventFailed:
		if stats != nil && stats.FailCount > 1 {
			message = fmt.Sprintf("Service failed (failure #%d)", stats.FailCount)
		} else {
			message = "Service failed"
		}
	case domainprocess.EventRestarting:
		if stats != nil {
			message = fmt.Sprintf("Service restarting (attempt #%d)", stats.RestartCount+1)
		} else {
			message = "Service restarting"
		}
	case domainprocess.EventHealthy:
		message = "Service became healthy"
	case domainprocess.EventUnhealthy:
		message = "Service became unhealthy"
	case domainprocess.EventExhausted:
		if stats != nil {
			message = fmt.Sprintf("Service abandoned after %d restarts (max exceeded)", stats.RestartCount)
		} else {
			message = "Service abandoned (max restarts exceeded)"
		}
	default:
		message = "Service event"
	}

	// Create the log event.
	logEvent := domainlogging.NewLogEvent(level, serviceName, eventType, message)

	// Add metadata.
	if event.PID > 0 {
		logEvent = logEvent.WithMeta("pid", event.PID)
	}
	// Always show exit_code for stopped/failed events (even 0).
	if event.Type == domainprocess.EventStopped || event.Type == domainprocess.EventFailed {
		logEvent = logEvent.WithMeta("exit_code", event.ExitCode)
	}
	if event.Error != nil {
		logEvent = logEvent.WithMeta("error", event.Error.Error())
	}
	// Add restart count for restarting/exhausted events.
	if stats != nil && (event.Type == domainprocess.EventRestarting || event.Type == domainprocess.EventExhausted) {
		logEvent = logEvent.WithMeta("restarts", stats.RestartCount)
	}

	return logEvent
}

// findLogFilePath finds the first file writer's path from the config.
// Returns the absolute path to the log file, or empty string if not configured.
//
// Params:
//   - app: the application containing the config.
//
// Returns:
//   - string: the log file path, or empty string if not found.
func findLogFilePath(cfg *domainconfig.Config) string {
	if cfg == nil {
		return ""
	}

	baseDir := cfg.Logging.BaseDir
	for _, w := range cfg.Logging.Daemon.Writers {
		if w.Type == "file" && w.File.Path != "" {
			path := w.File.Path
			if !filepath.IsAbs(path) && baseDir != "" {
				path = filepath.Join(baseDir, path)
			}
			return path
		}
	}
	return ""
}
