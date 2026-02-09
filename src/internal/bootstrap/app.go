// Package bootstrap provides dependency injection wiring using Google Wire.
// It isolates all dependency construction from the main entry point,
// allowing for a minimal main.go and better testability.
package bootstrap

import (
	"context"
	"errors"
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
	"github.com/kodflow/daemon/internal/infrastructure/probe"
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
	// ErrUnsupportedTUIMode indicates an unknown TUI mode was requested.
	ErrUnsupportedTUIMode error = errors.New("unsupported TUI mode")
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
	probeMode := flag.Bool("probe", false, "collect all system metrics and output as JSON")
	flag.Parse()

	// print version and exit early if requested
	if *showVersion {
		fmt.Printf("supervizio %s\n", version)
		// return success after showing version
		return 0
	}

	// run probe mode if requested
	if *probeMode {
		// return exit code from probe mode
		return runProbeMode()
	}

	tuiMode := determineTUIMode(*forceInteractive)

	// run main application logic with error handling
	if err := run(configPath, tuiMode); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		// return error code on failure
		return 1
	}
	// return success code on clean exit
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
	// enable interactive mode when explicitly requested
	if forceInteractive {
		// return interactive mode for full TUI experience
		return tui.ModeInteractive
	}
	// return raw mode as default for minimal output
	return tui.ModeRaw
}

// runProbeMode collects all system metrics and outputs them as JSON.
// This is a standalone mode that doesn't start the supervisor.
//
// Returns:
//   - int: exit code (0 for success, 1 for error).
func runProbeMode() int {
	// initialize the probe library
	if err := probe.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to initialize probe: %v\n", err)
		// return error code on probe init failure
		return 1
	}
	defer probe.Shutdown()

	// collect all metrics and output as JSON
	ctx := context.Background()
	cfg := domainconfig.DefaultMetricsConfig()
	jsonStr, err := probe.CollectAllMetricsJSON(ctx, &cfg)
	// return early if metrics collection failed
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to collect metrics: %v\n", err)
		// return error code on collection failure
		return 1
	}

	// output JSON to stdout
	fmt.Println(jsonStr)
	// return success code
	return 0
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
	// delegate to main run function with default mode
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
	// return early if initialization failed
	if err != nil {
		// propagate initialization error with context
		return fmt.Errorf("failed to initialize: %w", err)
	}
	// schedule cleanup when function returns
	if app.Cleanup != nil {
		defer app.Cleanup()
	}

	logger, bufferedConsole := setupLoggingAndEvents(app, logAdapter, tuiMode)
	defer func() { _ = logger.Close() }()

	ctx, cancel, sigCh := setupContextAndSignals()
	defer cancel()

	// start all core services before TUI
	if err := startSupervisorAndMetrics(ctx, app, logger); err != nil {
		// propagate startup error
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
	// delegate to mode-specific execution logic
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
	// propagate wire initialization errors
	if err != nil {
		// return error with all nil values
		return nil, nil, err
	}

	logAdapter := tui.NewLogAdapter()

	// attempt to load log history for TUI display
	if logFilePath := findLogFilePath(app.Config); logFilePath != "" {
		// warn if history loading fails but continue execution
		if err := logAdapter.LoadLogHistory(logFilePath, defaultLogHistoryLines); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load log history: %v\n", err)
		}
	}

	// return initialized components
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
	// warn on logger initialization failure but continue
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to build daemon logger: %v\n", err)
	}

	attachTUIWriter(logger, logAdapter)

	app.Supervisor.SetEventHandler(func(serviceName string, event *domainprocess.Event, stats *appsupervisor.ServiceStatsSnapshot) {
		logEvent := convertProcessEventToLogEvent(serviceName, event, stats)
		logger.Log(logEvent)
	})

	// return configured logging infrastructure
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

	// return context and signal infrastructure
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

	// start supervisor before metrics tracking
	if err := app.Supervisor.Start(ctx); err != nil {
		// propagate supervisor start error with context
		return fmt.Errorf("failed to start supervisor: %w", err)
	}

	// start metrics tracker if configured
	if app.MetricsTracker != nil {
		_ = app.MetricsTracker.Start(ctx)
	}

	// return nil on successful startup
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

	// select logger type based on TUI mode requirements
	if tuiMode == tui.ModeInteractive {
		logger, loggerErr = daemonlogger.BuildLoggerWithoutConsole(
			cfg.Logging.Daemon,
			cfg.Logging.BaseDir,
		)
		// use buffered console for raw mode
	} else {
		logger, bufferedConsole, loggerErr = daemonlogger.BuildLoggerWithBufferedConsole(
			cfg.Logging.Daemon,
			cfg.Logging.BaseDir,
		)
	}

	// fallback to default logger if initialization failed
	if loggerErr != nil {
		// use silent logger in interactive mode to avoid pollution
		if tuiMode == tui.ModeInteractive {
			logger = daemonlogger.NewSilentLogger()
			// use default console logger for raw mode
		} else {
			logger = daemonlogger.DefaultLogger()
		}
	}

	// return logger and optional buffered console
	return logger, bufferedConsole, loggerErr
}

// attachTUIWriter adds TUI writer to logger if it's a MultiLogger.
//
// Params:
//   - logger: the logger to attach writer to.
//   - logAdapter: the log adapter for TUI.
func attachTUIWriter(logger domainlogging.Logger, logAdapter *tui.LogAdapter) {
	// attach TUI writer if logger supports multiple writers
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

	// attach service lister if supervisor provides snapshots
	if sup, ok := supervisor.(ServiceSnapshotsForTUIer); ok {
		lister := &supervisorServiceLister{sup: sup}
		t.SetServiceLister(lister)
	}

	t.SetSummarizeer(logAdapter)
	t.SetConfigPath(cfgPath)

	// return fully configured TUI instance
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
	// execute mode-specific logic
	switch cfg.tuiMode {
	// handle raw mode with MOTD banner
	case tui.ModeRaw:
		// warn if TUI rendering fails but continue
		if err := cfg.tui.Run(cfg.ctx); err != nil {
			fmt.Fprintf(os.Stderr, "warning: TUI error: %v\n", err)
		}
		// flush buffered logs after MOTD display
		if cfg.bufferedConsole != nil {
			_ = cfg.bufferedConsole.Flush()
		}
		// return result from signal handling loop
		return WaitForSignals(cfg.ctx, cfg.cancel, cfg.sigCh, cfg.sup)

	// handle interactive mode with full TUI
	case tui.ModeInteractive:
		tuiDone := make(chan error, 1)
		// run TUI in background goroutine until completion
		go func() {
			tuiDone <- cfg.tui.Run(cfg.ctx)
		}()

		// return result from combined TUI and signal handling
		return waitForTUIOrSignals(cfg.ctx, cfg.cancel, cfg.sigCh, tuiDone, cfg.sup)

	// handle unsupported TUI modes
	default:
		// return error for unsupported mode
		return fmt.Errorf("%w: %v", ErrUnsupportedTUIMode, cfg.tuiMode)
	}
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
	// wait for first event to occur
	for {
		select {
		case sig := <-sigCh:
			// process signal and propagate shutdown errors
			if err := handleSignal(sig, cancel, sup); err != nil {
				// return error from signal handler
				return err
			}
		case err := <-tuiDone:
			// log TUI errors but continue shutdown
			if err != nil {
				fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			}
			cancel()
			// return result of graceful shutdown
			return sup.Stop()
		case <-ctx.Done():
			// return result of graceful shutdown
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
	// route signal to appropriate handler
	switch sig {
	// reload configuration on SIGHUP
	case syscall.SIGHUP:
		// attempt config reload but continue on failure
		if err := sup.Reload(); err != nil {
			fmt.Fprintf(os.Stderr, "reload failed: %v\n", err)
		}
		// return nil to continue signal loop
		return nil
	// graceful shutdown on SIGTERM or SIGINT
	case syscall.SIGTERM, syscall.SIGINT:
		cancel()
		// return stop result to terminate loop
		return sup.Stop()
	}
	// return nil for unknown signals
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
	// wait for shutdown signal or context cancellation
	for {
		select {
		case sig := <-sigCh:
			// process signal and propagate shutdown errors
			if err := handleSignal(sig, cancel, sup); err != nil {
				// return error to terminate loop
				return err
			}
		case <-ctx.Done():
			// return result of graceful shutdown
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
	// return event enriched with metadata
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
	// map event types to severity levels
	switch eventType {
	// warn level for recoverable failures
	case domainprocess.EventFailed, domainprocess.EventUnhealthy:
		// return warning for recoverable failures
		return domainlogging.LevelWarn
	// error level for permanent failures
	case domainprocess.EventExhausted:
		// return error for permanent failures
		return domainlogging.LevelError
	// info level for normal lifecycle events
	case domainprocess.EventStarted, domainprocess.EventStopped,
		domainprocess.EventRestarting, domainprocess.EventHealthy:
		// return info for normal lifecycle events
		return domainlogging.LevelInfo
	// safe default for unknown events
	default:
		// return info as safe default
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
	// generate context-specific message for event type
	switch event.Type {
	// service started event
	case domainprocess.EventStarted:
		// return message with restart context if available
		return buildStartedMessage(stats)
	// service stopped event
	case domainprocess.EventStopped:
		// return message with exit code details
		return buildStoppedMessage(event.ExitCode)
	// service failed event
	case domainprocess.EventFailed:
		// return message with failure count if available
		return buildFailedMessage(stats)
	// service restarting event
	case domainprocess.EventRestarting:
		// return message with restart attempt number
		return buildRestartingMessage(stats)
	// service became healthy
	case domainprocess.EventHealthy:
		// return simple health recovery message
		return "Service became healthy"
	// service became unhealthy
	case domainprocess.EventUnhealthy:
		// return simple health degradation message
		return "Service became unhealthy"
	// service exhausted restart attempts
	case domainprocess.EventExhausted:
		// return message with total restart count
		return buildExhaustedMessage(stats)
	// unknown or custom event
	default:
		// return generic message for unknown events
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
	// include restart count if this is not first start
	if stats != nil && stats.RestartCount > 0 {
		// return message showing restart number
		return fmt.Sprintf("Service started (restart #%d)", stats.RestartCount)
	}
	// return simple message for initial start
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
	// indicate clean shutdown for zero exit code
	if exitCode == cleanExitCode {
		// return success message
		return "Service stopped cleanly"
	}
	// return message with non-zero exit code
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
	// include failure count if multiple failures occurred
	if stats != nil && stats.FailCount > 1 {
		// return message showing failure number
		return fmt.Sprintf("Service failed (failure #%d)", stats.FailCount)
	}
	// return simple message for first failure
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
	// include next attempt number if stats available
	if stats != nil {
		// return message showing next restart attempt number
		return fmt.Sprintf("Service restarting (attempt #%d)", stats.RestartCount+1)
	}
	// return simple message without attempt number
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
	// include total restart count if stats available
	if stats != nil {
		// return message showing total restart attempts
		return fmt.Sprintf("Service abandoned after %d restarts (max exceeded)", stats.RestartCount)
	}
	// return generic message without count
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
	// ensure type safety before enrichment
	result, ok := logEvent.(domainlogging.LogEvent)
	// return empty event if type assertion fails
	if !ok {
		// return zero value for incompatible types
		return domainlogging.LogEvent{}
	}

	result = addPIDMetadata(result, event)
	result = addExitMetadata(result, event)
	result = addRestartMetadata(result, event, stats)

	// return fully enriched event
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
	// add PID metadata if process was running
	if event.PID > 0 {
		// return event enriched with PID
		return result.WithMeta("pid", event.PID)
	}
	logEvent, _ := result.(domainlogging.LogEvent)
	// return unchanged event if no PID
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
	enriched := result
	// add exit code for terminal events
	if event.Type == domainprocess.EventStopped || event.Type == domainprocess.EventFailed {
		enriched = enriched.WithMeta("exit_code", event.ExitCode)
	}

	// add error message if available
	if event.Error != nil {
		enriched = enriched.WithMeta("error", event.Error.Error())
	}

	logEvent, _ := enriched.(domainlogging.LogEvent)
	// return event with exit metadata
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
	// add restart count for restart-related events
	if stats != nil && (event.Type == domainprocess.EventRestarting || event.Type == domainprocess.EventExhausted) {
		// return event enriched with restart count
		return result.WithMeta("restarts", stats.RestartCount)
	}
	logEvent, _ := result.(domainlogging.LogEvent)
	// return unchanged event if not restart-related
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
	// return early if config not available
	if cfg == nil {
		// return empty string for nil config
		return ""
	}

	baseDir := cfg.Logging.BaseDir
	// search for first file writer in config
	for i := range cfg.Logging.Daemon.Writers {
		w := &cfg.Logging.Daemon.Writers[i]
		// check if writer is file type with valid path
		if w.Type == "file" && w.File.Path != "" {
			path := w.File.Path
			// convert relative path to absolute using base dir
			if !filepath.IsAbs(path) && baseDir != "" {
				path = filepath.Join(baseDir, path)
			}
			// return first valid file path found
			return path
		}
	}
	// return empty string if no file writer found
	return ""
}
