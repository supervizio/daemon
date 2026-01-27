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

	appmetrics "github.com/kodflow/daemon/internal/application/metrics"
	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
	domainconfig "github.com/kodflow/daemon/internal/domain/config"
	domainlogging "github.com/kodflow/daemon/internal/domain/logging"
	domainprocess "github.com/kodflow/daemon/internal/domain/process"
	daemonlogger "github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
)

const (
	// defaultLogHistoryLines is the number of log lines to load from history file.
	defaultLogHistoryLines int = 100
	// cleanExitCode is the exit code for a clean process termination.
	cleanExitCode int = 0
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

// WithMetaer defines the interface for enriching log events (KTN-API-MINIF).
type WithMetaer interface {
	WithMeta(key string, value any) domainlogging.LogEvent
}

// Flusher defines the interface for flushing buffered output (KTN-API-MINIF).
type Flusher interface {
	Flush() error
}

// Runner defines the interface for running a component (KTN-API-MINIF).
type Runner interface {
	Run(ctx context.Context) error
}

// tuiModeConfig holds configuration for TUI mode execution.
type tuiModeConfig struct {
	ctx             context.Context
	cancel          context.CancelFunc
	sigCh           <-chan os.Signal
	tui             Runner
	bufferedConsole Flusher
	tuiMode         tui.Mode
	sup             SignalHandler
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
//
// Params:
//   - forceInteractive: whether to force interactive mode.
//
// Returns:
//   - tui.Mode: the determined TUI mode.
func determineTUIMode(forceInteractive bool) tui.Mode {
	// Check if interactive mode is forced via flag.
	if forceInteractive {
		// Return interactive mode when explicitly requested.
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

// run executes the main application logic with the specified TUI mode.
//
// Params:
//   - cfgPath: the path to the configuration file.
//   - tuiMode: the TUI display mode (raw or interactive).
//
// Returns:
//   - error: nil on success, error on failure.
func run(cfgPath string, tuiMode tui.Mode) error {
	// Initialize the application and its dependencies.
	app, logAdapter, err := initializeAppAndLogAdapter(cfgPath)
	// Check if initialization failed.
	if err != nil {
		// Return error with context about initialization failure.
		return fmt.Errorf("failed to initialize: %w", err)
	}
	// Ensure cleanup is called when run exits.
	if app.Cleanup != nil {
		defer app.Cleanup()
	}

	// Initialize logger and setup event handling.
	logger, bufferedConsole := setupLoggingAndEvents(app, logAdapter, tuiMode)
	defer func() { _ = logger.Close() }()

	// Setup supervisor context and start services.
	ctx, cancel, sigCh := setupContextAndSignals()
	defer cancel()

	// Start supervisor and metrics tracking.
	if err := startSupervisorAndMetrics(ctx, app, logger); err != nil {
		// Return error with context about startup failure.
		return err
	}

	// Create and configure TUI with service and log providers.
	t := setupTUI(app.Supervisor, logAdapter, cfgPath, tuiMode)

	// Run TUI based on mode and handle signals.
	cfg := tuiModeConfig{
		ctx:             ctx,
		cancel:          cancel,
		sigCh:           sigCh,
		tui:             t,
		bufferedConsole: bufferedConsole,
		tuiMode:         tuiMode,
		sup:             app.Supervisor,
	}
	// Execute TUI mode with the configured parameters.
	return runTUIMode(cfg)
}

// initializeAppAndLogAdapter initializes the app and log adapter.
//
// Params:
//   - cfgPath: the path to the configuration file.
//
// Returns:
//   - *App: the initialized application.
//   - *tui.LogAdapter: the log adapter for TUI.
//   - error: initialization error if any.
func initializeAppAndLogAdapter(cfgPath string) (*App, *tui.LogAdapter, error) {
	// Initialize the application using Wire-generated code.
	app, err := InitializeApp(cfgPath)
	// Check if initialization failed.
	if err != nil {
		// Return error with nil values.
		return nil, nil, err
	}

	// Create log adapter for TUI (captures logs for display).
	logAdapter := tui.NewLogAdapter()

	// Attempt to load recent log history from file if available.
	if logFilePath := findLogFilePath(app.Config); logFilePath != "" {
		// Load the last N lines of log history for TUI display.
		if err := logAdapter.LoadLogHistory(logFilePath, defaultLogHistoryLines); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load log history: %v\n", err)
		}
	}

	// Return initialized app and log adapter.
	return app, logAdapter, nil
}

// setupLoggingAndEvents initializes logger and sets up event handling.
//
// Params:
//   - app: the application instance.
//   - logAdapter: the log adapter for TUI.
//   - tuiMode: the TUI display mode.
//
// Returns:
//   - domainlogging.Logger: the configured logger.
//   - *daemonlogger.BufferedWriter: buffered console writer (nil in interactive mode).
func setupLoggingAndEvents(app *App, logAdapter *tui.LogAdapter, tuiMode tui.Mode) (domainlogging.Logger, *daemonlogger.BufferedWriter) {
	// Initialize logger and buffered console writer.
	logger, bufferedConsole, err := initializeLogger(app.Config, tuiMode)
	// Check if logger initialization failed (continue with fallback).
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to build daemon logger: %v\n", err)
	}

	// Attach TUI writer to logger to capture events for display.
	attachTUIWriter(logger, logAdapter)

	// Set up event handler to log service events.
	app.Supervisor.SetEventHandler(func(serviceName string, event *domainprocess.Event, stats *appsupervisor.ServiceStatsSnapshot) {
		logEvent := convertProcessEventToLogEvent(serviceName, event, stats)
		logger.Log(logEvent)
	})

	// Return configured logger and buffered console.
	return logger, bufferedConsole
}

// setupContextAndSignals creates context and signal channel.
//
// Returns:
//   - context.Context: the context for cancellation.
//   - context.CancelFunc: the cancel function.
//   - chan os.Signal: the signal channel.
func setupContextAndSignals() (context.Context, context.CancelFunc, chan os.Signal) {
	// Setup context and signals for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	// Return context, cancel function, and signal channel.
	return ctx, cancel, sigCh
}

// startSupervisorAndMetrics starts the supervisor and metrics tracker.
//
// Params:
//   - ctx: the context for cancellation.
//   - app: the application instance.
//   - logger: the logger instance.
//
// Returns:
//   - error: startup error if any.
func startSupervisorAndMetrics(ctx context.Context, app *App, logger domainlogging.Logger) error {
	// Log daemon startup.
	logger.Info("", "daemon_started", "Supervisor started", map[string]any{"version": version})

	// Start supervisor and check for startup errors.
	if err := app.Supervisor.Start(ctx); err != nil {
		// Return error with context about supervisor startup failure.
		return fmt.Errorf("failed to start supervisor: %w", err)
	}

	// Start metrics tracker for CPU/memory monitoring if available.
	if app.MetricsTracker != nil {
		_ = app.MetricsTracker.Start(ctx)
	}

	// Return nil indicating successful startup.
	return nil
}

// initializeLogger creates and configures the logger based on TUI mode.
//
// Params:
//   - cfg: the application configuration.
//   - tuiMode: the TUI display mode.
//
// Returns:
//   - domainlogging.Logger: the configured logger.
//   - *daemonlogger.BufferedWriter: buffered console writer (nil in interactive mode).
//   - error: initialization error if any.
func initializeLogger(cfg *domainconfig.Config, tuiMode tui.Mode) (domainlogging.Logger, *daemonlogger.BufferedWriter, error) {
	var logger domainlogging.Logger
	var bufferedConsole *daemonlogger.BufferedWriter
	var loggerErr error

	// Check if interactive mode to avoid console pollution.
	if tuiMode == tui.ModeInteractive {
		// Interactive mode: only file writers + TUI writer (no console pollution).
		logger, loggerErr = daemonlogger.BuildLoggerWithoutConsole(
			cfg.Logging.Daemon,
			cfg.Logging.BaseDir,
		)
	} else {
		// Raw mode: use buffered console writer so logs appear after MOTD banner.
		logger, bufferedConsole, loggerErr = daemonlogger.BuildLoggerWithBufferedConsole(
			cfg.Logging.Daemon,
			cfg.Logging.BaseDir,
		)
	}

	// Check if logger creation failed (fallback to default logger).
	if loggerErr != nil {
		// Use silent logger in interactive mode or default logger in raw mode.
		if tuiMode == tui.ModeInteractive {
			// In interactive mode, use a silent logger (TUI only).
			logger = daemonlogger.NewSilentLogger()
		} else {
			// In raw mode, use default console logger as fallback.
			logger = daemonlogger.DefaultLogger()
		}
	}

	// Return configured logger and any initialization error.
	return logger, bufferedConsole, loggerErr
}

// attachTUIWriter adds TUI writer to logger if it's a MultiLogger.
//
// Params:
//   - logger: the logger to attach writer to.
//   - logAdapter: the log adapter for TUI.
func attachTUIWriter(logger domainlogging.Logger, logAdapter *tui.LogAdapter) {
	// Check if logger is a MultiLogger that supports adding writers.
	if ml, ok := logger.(*daemonlogger.MultiLogger); ok {
		tuiWriter := tui.NewTUILogWriter(logAdapter)
		ml.AddWriter(tuiWriter)
	}
}

// setupTUI creates and configures the TUI with all providers.
//
// Params:
//   - supervisor: the application supervisor.
//   - logAdapter: the log adapter for health display.
//   - cfgPath: the configuration file path.
//   - tuiMode: the TUI display mode.
//
// Returns:
//   - *tui.TUI: the configured TUI instance.
func setupTUI(supervisor AppSupervisor, logAdapter *tui.LogAdapter, cfgPath string, tuiMode tui.Mode) *tui.TUI {
	// Create TUI with the configured mode.
	tuiConfig := tui.DefaultConfig(version)
	tuiConfig.Mode = tuiMode
	t := tui.New(tuiConfig)

	// Set service provider if supervisor supports ServiceSnapshotsForTUI.
	if sup, ok := supervisor.(ServiceSnapshotsForTUIer); ok {
		// Create a service provider that queries the supervisor.
		provider := &supervisorServiceProvider{sup: sup}
		t.SetServiceProvider(provider)
	}

	// Set log adapter as health provider to display logs in TUI.
	t.SetHealthProvider(logAdapter)

	// Set config path for TUI display.
	t.SetConfigPath(cfgPath)

	// Return fully configured TUI instance.
	return t
}

// runTUIMode executes TUI and handles signals based on the configured mode.
//
// Params:
//   - cfg: the TUI mode configuration containing all required parameters.
//
// Returns:
//   - error: nil on success, error on failure.
//
// Goroutine lifecycle (KTN-GOROUTINE-LIFECYCLE):
//   - In interactive mode, a goroutine runs t.Run(ctx) until TUI exits or context cancels.
//   - The goroutine sends its result to tuiDone channel and terminates.
//   - Cleanup occurs when waitForTUIOrSignals returns, ensuring the goroutine has exited.
func runTUIMode(cfg tuiModeConfig) error {
	// Switch on TUI mode to determine execution flow.
	switch cfg.tuiMode {
	// Handle raw mode: show MOTD once then wait for signals.
	case tui.ModeRaw:
		// Raw mode: show MOTD once, then flush buffered logs and wait for signals.
		if err := cfg.tui.Run(cfg.ctx); err != nil {
			fmt.Fprintf(os.Stderr, "warning: TUI error: %v\n", err)
		}
		// Flush buffered logs after MOTD banner is displayed.
		if cfg.bufferedConsole != nil {
			_ = cfg.bufferedConsole.Flush()
		}
		// Continue to signal handling.
		return WaitForSignals(cfg.ctx, cfg.cancel, cfg.sigCh, cfg.sup)

	// Handle interactive mode: run TUI in parallel with signal handling.
	case tui.ModeInteractive:
		// Interactive mode: run TUI in parallel with signal handling.
		tuiDone := make(chan error, 1)
		// Goroutine lifecycle: runs until TUI exits or context cancels, then sends result.
		go func() {
			tuiDone <- cfg.tui.Run(cfg.ctx)
		}()

		// Wait for TUI exit or signals.
		return waitForTUIOrSignals(cfg.ctx, cfg.cancel, cfg.sigCh, tuiDone, cfg.sup)
	}

	// Fallback: wait for signals.
	return WaitForSignals(cfg.ctx, cfg.cancel, cfg.sigCh, cfg.sup)
}

// waitForTUIOrSignals handles both TUI events and OS signals.
//
// Params:
//   - ctx: the context for cancellation.
//   - cancel: the cancel function for the context.
//   - sigCh: the channel receiving OS signals.
//   - tuiDone: channel signaling TUI completion.
//   - sup: the signal handler interface.
//
// Returns:
//   - error: nil on success, error on failure.
func waitForTUIOrSignals(ctx context.Context, cancel context.CancelFunc, sigCh <-chan os.Signal, tuiDone <-chan error, sup SignalHandler) error {
	// Loop forever until TUI exits or shutdown signal received.
	for {
		// Select on signals, TUI completion, or context cancellation.
		select {
		// Handle OS signals for reload or shutdown.
		case sig := <-sigCh:
			// Handle signal based on type.
			if err := handleSignal(sig, cancel, sup); err != nil {
				// Return error from signal handling.
				return err
			}
		// Handle TUI exit (user pressed q or error occurred).
		case err := <-tuiDone:
			// TUI exited (user pressed q or error).
			if err != nil {
				fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			}
			// Stop the supervisor when TUI exits.
			cancel()
			// Stop supervisor and return result.
			return sup.Stop()
		// Handle context cancellation.
		case <-ctx.Done():
			// Stop supervisor when context is cancelled.
			return sup.Stop()
		}
	}
}

// handleSignal processes a single OS signal.
//
// Params:
//   - sig: the OS signal received.
//   - cancel: the cancel function for context.
//   - sup: the signal handler interface.
//
// Returns:
//   - error: error from stop operation if shutdown signal, nil otherwise.
func handleSignal(sig os.Signal, cancel context.CancelFunc, sup SignalHandler) error {
	// Switch on signal type to determine action.
	switch sig {
	// Handle SIGHUP for configuration reload.
	case syscall.SIGHUP:
		// Attempt to reload configuration.
		if err := sup.Reload(); err != nil {
			fmt.Fprintf(os.Stderr, "reload failed: %v\n", err)
		}
		// Return nil to continue loop.
		return nil
	// Handle SIGTERM or SIGINT for graceful shutdown.
	case syscall.SIGTERM, syscall.SIGINT:
		cancel()
		// Stop supervisor and return result.
		return sup.Stop()
	}
	// Return nil for unhandled signals.
	return nil
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
		// Handle OS signals received on signal channel.
		case sig := <-sigCh:
			// Handle signal and return error if shutdown requested.
			if err := handleSignal(sig, cancel, sup); err != nil {
				// Return error from signal handling (shutdown).
				return err
			}
		// Handle context cancellation.
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
	// Determine appropriate log level based on event type.
	level := determineLogLevel(event.Type)

	// Build event type string.
	eventType := event.Type.String()

	// Build descriptive message based on event type and context.
	message := buildEventMessage(event, stats)

	// Create the log event with determined level and message.
	logEvent := domainlogging.NewLogEvent(level, serviceName, eventType, message)

	// Add relevant metadata to enrich log event.
	return addEventMetadata(logEvent, event, stats)
}

// determineLogLevel maps process event type to appropriate log level.
//
// Params:
//   - eventType: the type of process event.
//
// Returns:
//   - domainlogging.Level: the determined log level.
func determineLogLevel(eventType domainprocess.EventType) domainlogging.Level {
	// Switch on event type to determine appropriate log level.
	switch eventType {
	// Failed and unhealthy events are warnings.
	case domainprocess.EventFailed, domainprocess.EventUnhealthy:
		// Return warning level for failure and unhealthy events.
		return domainlogging.LevelWarn
	// Exhausted events are errors (max restarts exceeded).
	case domainprocess.EventExhausted:
		// Return error level for exhausted restart attempts.
		return domainlogging.LevelError
	// Started, stopped, restarting, and healthy events are informational.
	case domainprocess.EventStarted, domainprocess.EventStopped,
		domainprocess.EventRestarting, domainprocess.EventHealthy:
		// Return info level for normal lifecycle events.
		return domainlogging.LevelInfo
	// Default to info level for unknown event types.
	default:
		// Return info level as safe default.
		return domainlogging.LevelInfo
	}
}

// buildEventMessage constructs a descriptive message for the process event.
//
// Params:
//   - event: the process event.
//   - stats: optional service statistics for context.
//
// Returns:
//   - string: the formatted event message.
func buildEventMessage(event *domainprocess.Event, stats *appsupervisor.ServiceStatsSnapshot) string {
	// Switch on event type to build appropriate message.
	switch event.Type {
	// Handle service started event.
	case domainprocess.EventStarted:
		// Return message for service start (with restart count if applicable).
		return buildStartedMessage(stats)

	// Handle service stopped event.
	case domainprocess.EventStopped:
		// Return message for service stop (with exit code).
		return buildStoppedMessage(event.ExitCode)

	// Handle service failed event.
	case domainprocess.EventFailed:
		// Return message for service failure (with failure count if applicable).
		return buildFailedMessage(stats)

	// Handle service restarting event.
	case domainprocess.EventRestarting:
		// Return message for service restart (with attempt number if applicable).
		return buildRestartingMessage(stats)

	// Handle service became healthy event.
	case domainprocess.EventHealthy:
		// Return message indicating service is now healthy.
		return "Service became healthy"

	// Handle service became unhealthy event.
	case domainprocess.EventUnhealthy:
		// Return message indicating service is now unhealthy.
		return "Service became unhealthy"

	// Handle service exhausted (max restarts exceeded) event.
	case domainprocess.EventExhausted:
		// Return message for service abandonment (with restart count).
		return buildExhaustedMessage(stats)

	// Handle unknown event types.
	default:
		// Return generic message for unknown event types.
		return "Service event"
	}
}

// buildStartedMessage creates message for service start event.
//
// Params:
//   - stats: optional service statistics.
//
// Returns:
//   - string: the formatted message.
func buildStartedMessage(stats *appsupervisor.ServiceStatsSnapshot) string {
	// Check if this is a restart (not initial start).
	if stats != nil && stats.RestartCount > 0 {
		// Return message showing restart count.
		return fmt.Sprintf("Service started (restart #%d)", stats.RestartCount)
	}
	// Return simple started message for initial start.
	return "Service started"
}

// buildStoppedMessage creates message for service stop event.
//
// Params:
//   - exitCode: the process exit code.
//
// Returns:
//   - string: the formatted message.
func buildStoppedMessage(exitCode int) string {
	// Differentiate clean exit from non-clean exit.
	if exitCode == cleanExitCode {
		// Return message for clean shutdown.
		return "Service stopped cleanly"
	}
	// Return message showing non-zero exit code.
	return fmt.Sprintf("Service exited with code %d", exitCode)
}

// buildFailedMessage creates message for service failure event.
//
// Params:
//   - stats: optional service statistics.
//
// Returns:
//   - string: the formatted message.
func buildFailedMessage(stats *appsupervisor.ServiceStatsSnapshot) string {
	// Check if this is a repeated failure.
	if stats != nil && stats.FailCount > 1 {
		// Return message showing failure count.
		return fmt.Sprintf("Service failed (failure #%d)", stats.FailCount)
	}
	// Return simple failure message for first failure.
	return "Service failed"
}

// buildRestartingMessage creates message for service restart event.
//
// Params:
//   - stats: optional service statistics.
//
// Returns:
//   - string: the formatted message.
func buildRestartingMessage(stats *appsupervisor.ServiceStatsSnapshot) string {
	// Check if stats available to show restart attempt number.
	if stats != nil {
		// Return message showing restart attempt number.
		return fmt.Sprintf("Service restarting (attempt #%d)", stats.RestartCount+1)
	}
	// Return simple restarting message.
	return "Service restarting"
}

// buildExhaustedMessage creates message for exhausted restart event.
//
// Params:
//   - stats: optional service statistics.
//
// Returns:
//   - string: the formatted message.
func buildExhaustedMessage(stats *appsupervisor.ServiceStatsSnapshot) string {
	// Check if stats available to show total restart count.
	if stats != nil {
		// Return message showing restart count when abandoned.
		return fmt.Sprintf("Service abandoned after %d restarts (max exceeded)", stats.RestartCount)
	}
	// Return simple exhausted message.
	return "Service abandoned (max restarts exceeded)"
}

// addEventMetadata enriches log event with relevant metadata fields.
//
// Params:
//   - logEvent: the log event to enrich (implements WithMetaer).
//   - event: the process event.
//   - stats: optional service statistics.
//
// Returns:
//   - domainlogging.LogEvent: the enriched log event.
func addEventMetadata(logEvent WithMetaer, event *domainprocess.Event, stats *appsupervisor.ServiceStatsSnapshot) domainlogging.LogEvent {
	// Start with the input event cast to LogEvent for enrichment (KTN-VAR-TYPEASSERT).
	result, ok := logEvent.(domainlogging.LogEvent)
	if !ok {
		// Return zero value if type assertion fails.
		return domainlogging.LogEvent{}
	}

	// Add PID if process was running.
	if event.PID > 0 {
		result = result.WithMeta("pid", event.PID)
	}

	// Always show exit_code for stopped/failed events (even 0).
	if event.Type == domainprocess.EventStopped || event.Type == domainprocess.EventFailed {
		result = result.WithMeta("exit_code", event.ExitCode)
	}

	// Add error message if present.
	if event.Error != nil {
		result = result.WithMeta("error", event.Error.Error())
	}

	// Add restart count for restarting/exhausted events.
	if stats != nil && (event.Type == domainprocess.EventRestarting || event.Type == domainprocess.EventExhausted) {
		result = result.WithMeta("restarts", stats.RestartCount)
	}

	// Return enriched log event.
	return result
}

// findLogFilePath finds the first file writer's path from the config.
// Returns the absolute path to the log file, or empty string if not configured.
//
// Params:
//   - cfg: the application configuration.
//
// Returns:
//   - string: the log file path, or empty string if not found.
func findLogFilePath(cfg *domainconfig.Config) string {
	// Check if config is nil to avoid nil pointer dereference.
	if cfg == nil {
		// Return empty string if config is nil.
		return ""
	}

	baseDir := cfg.Logging.BaseDir
	// Iterate over configured writers to find file writer.
	for _, w := range cfg.Logging.Daemon.Writers {
		// Check if this is a file writer with a path.
		if w.Type == "file" && w.File.Path != "" {
			path := w.File.Path
			// Convert relative path to absolute using baseDir.
			if !filepath.IsAbs(path) && baseDir != "" {
				path = filepath.Join(baseDir, path)
			}
			// Return the first file writer path found.
			return path
		}
	}
	// Return empty string if no file writer found.
	return ""
}
