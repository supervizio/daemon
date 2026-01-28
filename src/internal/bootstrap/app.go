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

	if *showVersion {
		fmt.Printf("supervizio %s\n", version)
		return 0
	}

	tuiMode := determineTUIMode(*forceInteractive)

	if err := run(configPath, tuiMode); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
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
	if forceInteractive {
		return tui.ModeInteractive
	}
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
	app, logAdapter, err := initializeAppAndLogAdapter(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	if app.Cleanup != nil {
		defer app.Cleanup()
	}

	logger, bufferedConsole := setupLoggingAndEvents(app, logAdapter, tuiMode)
	defer func() { _ = logger.Close() }()

	ctx, cancel, sigCh := setupContextAndSignals()
	defer cancel()

	if err := startSupervisorAndMetrics(ctx, app, logger); err != nil {
		return err
	}

	t := setupTUI(app.Supervisor, logAdapter, cfgPath, tuiMode)

	cfg := tuiModeConfig{
		ctx:             ctx,
		cancel:          cancel,
		sigCh:           sigCh,
		tui:             t,
		bufferedConsole: bufferedConsole,
		tuiMode:         tuiMode,
		sup:             app.Supervisor,
	}
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
	app, err := InitializeApp(cfgPath)
	if err != nil {
		return nil, nil, err
	}

	logAdapter := tui.NewLogAdapter()

	// Load recent log history from file if available for TUI display.
	if logFilePath := findLogFilePath(app.Config); logFilePath != "" {
		if err := logAdapter.LoadLogHistory(logFilePath, defaultLogHistoryLines); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load log history: %v\n", err)
		}
	}

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
	logger, bufferedConsole, err := initializeLogger(app.Config, tuiMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to build daemon logger: %v\n", err)
	}

	attachTUIWriter(logger, logAdapter)

	app.Supervisor.SetEventHandler(func(serviceName string, event *domainprocess.Event, stats *appsupervisor.ServiceStatsSnapshot) {
		logEvent := convertProcessEventToLogEvent(serviceName, event, stats)
		logger.Log(logEvent)
	})

	return logger, bufferedConsole
}

// setupContextAndSignals creates context and signal channel.
//
// Returns:
//   - context.Context: the context for cancellation.
//   - context.CancelFunc: the cancel function.
//   - chan os.Signal: the signal channel.
func setupContextAndSignals() (context.Context, context.CancelFunc, chan os.Signal) {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

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
	logger.Info("", "daemon_started", "Supervisor started", map[string]any{"version": version})

	if err := app.Supervisor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start supervisor: %w", err)
	}

	if app.MetricsTracker != nil {
		_ = app.MetricsTracker.Start(ctx)
	}

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

	// Interactive mode: only file writers + TUI writer (no console pollution).
	// Raw mode: use buffered console writer so logs appear after MOTD banner.
	if tuiMode == tui.ModeInteractive {
		logger, loggerErr = daemonlogger.BuildLoggerWithoutConsole(
			cfg.Logging.Daemon,
			cfg.Logging.BaseDir,
		)
	} else {
		logger, bufferedConsole, loggerErr = daemonlogger.BuildLoggerWithBufferedConsole(
			cfg.Logging.Daemon,
			cfg.Logging.BaseDir,
		)
	}

	// Fallback to silent logger (interactive) or default console logger (raw).
	if loggerErr != nil {
		if tuiMode == tui.ModeInteractive {
			logger = daemonlogger.NewSilentLogger()
		} else {
			logger = daemonlogger.DefaultLogger()
		}
	}

	return logger, bufferedConsole, loggerErr
}

// attachTUIWriter adds TUI writer to logger if it's a MultiLogger.
//
// Params:
//   - logger: the logger to attach writer to.
//   - logAdapter: the log adapter for TUI.
func attachTUIWriter(logger domainlogging.Logger, logAdapter *tui.LogAdapter) {
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
	tuiConfig := tui.DefaultConfig(version)
	tuiConfig.Mode = tuiMode
	t := tui.NewTUI(tuiConfig)

	if sup, ok := supervisor.(ServiceSnapshotsForTUIer); ok {
		lister := &supervisorServiceLister{sup: sup}
		t.SetServiceLister(lister)
	}

	t.SetSummarizeer(logAdapter)
	t.SetConfigPath(cfgPath)

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
	switch cfg.tuiMode {
	case tui.ModeRaw:
		// Raw mode: show MOTD once, then flush buffered logs and wait for signals.
		if err := cfg.tui.Run(cfg.ctx); err != nil {
			fmt.Fprintf(os.Stderr, "warning: TUI error: %v\n", err)
		}
		if cfg.bufferedConsole != nil {
			_ = cfg.bufferedConsole.Flush()
		}
		return WaitForSignals(cfg.ctx, cfg.cancel, cfg.sigCh, cfg.sup)

	case tui.ModeInteractive:
		// Interactive mode: run TUI in parallel with signal handling.
		tuiDone := make(chan error, 1)
		// Goroutine lifecycle: runs until TUI exits or context cancels, then sends result.
		go func() {
			tuiDone <- cfg.tui.Run(cfg.ctx)
		}()

		return waitForTUIOrSignals(cfg.ctx, cfg.cancel, cfg.sigCh, tuiDone, cfg.sup)
	}

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
	for {
		select {
		case sig := <-sigCh:
			if err := handleSignal(sig, cancel, sup); err != nil {
				return err
			}
		case err := <-tuiDone:
			if err != nil {
				fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			}
			cancel()
			return sup.Stop()
		case <-ctx.Done():
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
	switch sig {
	case syscall.SIGHUP:
		if err := sup.Reload(); err != nil {
			fmt.Fprintf(os.Stderr, "reload failed: %v\n", err)
		}
		return nil
	case syscall.SIGTERM, syscall.SIGINT:
		cancel()
		return sup.Stop()
	}
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
	for {
		select {
		case sig := <-sigCh:
			if err := handleSignal(sig, cancel, sup); err != nil {
				return err
			}
		case <-ctx.Done():
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
	level := determineLogLevel(event.Type)
	eventType := event.Type.String()
	message := buildEventMessage(event, stats)
	logEvent := domainlogging.NewLogEvent(level, serviceName, eventType, message)
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
	switch eventType {
	case domainprocess.EventFailed, domainprocess.EventUnhealthy:
		return domainlogging.LevelWarn
	case domainprocess.EventExhausted:
		return domainlogging.LevelError
	case domainprocess.EventStarted, domainprocess.EventStopped,
		domainprocess.EventRestarting, domainprocess.EventHealthy:
		return domainlogging.LevelInfo
	default:
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
	switch event.Type {
	case domainprocess.EventStarted:
		return buildStartedMessage(stats)
	case domainprocess.EventStopped:
		return buildStoppedMessage(event.ExitCode)
	case domainprocess.EventFailed:
		return buildFailedMessage(stats)
	case domainprocess.EventRestarting:
		return buildRestartingMessage(stats)
	case domainprocess.EventHealthy:
		return "Service became healthy"
	case domainprocess.EventUnhealthy:
		return "Service became unhealthy"
	case domainprocess.EventExhausted:
		return buildExhaustedMessage(stats)
	default:
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
	if stats != nil && stats.RestartCount > 0 {
		return fmt.Sprintf("Service started (restart #%d)", stats.RestartCount)
	}
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
	if exitCode == cleanExitCode {
		return "Service stopped cleanly"
	}
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
	if stats != nil && stats.FailCount > 1 {
		return fmt.Sprintf("Service failed (failure #%d)", stats.FailCount)
	}
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
	if stats != nil {
		return fmt.Sprintf("Service restarting (attempt #%d)", stats.RestartCount+1)
	}
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
	if stats != nil {
		return fmt.Sprintf("Service abandoned after %d restarts (max exceeded)", stats.RestartCount)
	}
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
	// Type assertion for enrichment (KTN-VAR-TYPEASSERT).
	// Graceful fallback avoids panic when caller passes incompatible interface type.
	result, ok := logEvent.(domainlogging.LogEvent)
	if !ok {
		return domainlogging.LogEvent{}
	}

	result = addPIDMetadata(result, event)
	result = addExitMetadata(result, event)
	result = addRestartMetadata(result, event, stats)

	return result
}

// addPIDMetadata adds PID to log event if process was running.
//
// Params:
//   - result: the log event to enrich (uses WithMetaer interface).
//   - event: the process event.
//
// Returns:
//   - domainlogging.LogEvent: the enriched log event.
func addPIDMetadata(result WithMetaer, event *domainprocess.Event) domainlogging.LogEvent {
	if event.PID > 0 {
		return result.WithMeta("pid", event.PID)
	}
	// Type assertion safe: all callers pass LogEvent.
	logEvent, _ := result.(domainlogging.LogEvent)
	return logEvent
}

// addExitMetadata adds exit code and error to log event.
//
// Params:
//   - result: the log event to enrich (uses WithMetaer interface).
//   - event: the process event.
//
// Returns:
//   - domainlogging.LogEvent: the enriched log event.
func addExitMetadata(result WithMetaer, event *domainprocess.Event) domainlogging.LogEvent {
	// Track the enriched result through the chain.
	enriched := result
	if event.Type == domainprocess.EventStopped || event.Type == domainprocess.EventFailed {
		enriched = enriched.WithMeta("exit_code", event.ExitCode)
	}

	if event.Error != nil {
		enriched = enriched.WithMeta("error", event.Error.Error())
	}

	// Type assertion safe: WithMeta returns LogEvent which implements WithMetaer.
	logEvent, _ := enriched.(domainlogging.LogEvent)
	return logEvent
}

// addRestartMetadata adds restart count to log event for restarting/exhausted events.
//
// Params:
//   - result: the log event to enrich (uses WithMetaer interface).
//   - event: the process event.
//   - stats: optional service statistics.
//
// Returns:
//   - domainlogging.LogEvent: the enriched log event.
func addRestartMetadata(result WithMetaer, event *domainprocess.Event, stats *appsupervisor.ServiceStatsSnapshot) domainlogging.LogEvent {
	if stats != nil && (event.Type == domainprocess.EventRestarting || event.Type == domainprocess.EventExhausted) {
		return result.WithMeta("restarts", stats.RestartCount)
	}
	// Type assertion safe: all callers pass LogEvent.
	logEvent, _ := result.(domainlogging.LogEvent)
	return logEvent
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
