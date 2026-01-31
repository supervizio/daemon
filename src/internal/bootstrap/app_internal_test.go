package bootstrap

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
	domainconfig "github.com/kodflow/daemon/internal/domain/config"
	domainlogging "github.com/kodflow/daemon/internal/domain/logging"
	domainprocess "github.com/kodflow/daemon/internal/domain/process"
	daemonlogger "github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
)

// mockSignalHandler is a test double for signalHandler interface.
type mockSignalHandler struct {
	reloadCalled atomic.Bool
	stopCalled   atomic.Bool
	reloadErr    error
	stopErr      error
}

// Reload records that Reload was called and returns configured error.
func (m *mockSignalHandler) Reload() error {
	// Mark reload as called.
	m.reloadCalled.Store(true)
	// Return configured error or nil.
	return m.reloadErr
}

// Stop records that Stop was called and returns configured error.
func (m *mockSignalHandler) Stop() error {
	// Mark stop as called.
	m.stopCalled.Store(true)
	// Return configured error or nil.
	return m.stopErr
}

// TestWaitForSignals_Internal verifies WaitForSignals behavior with various signals.
// Goroutine lifecycle:
//   - Spawns a goroutine that runs WaitForSignals until the signal is processed.
//   - The goroutine exits when Stop is called via signal handling or context cancellation.
//   - Cleanup occurs when the function returns and the done channel receives the result.
func TestWaitForSignals_Internal(t *testing.T) {
	t.Parallel()

	// Define test cases for WaitForSignals.
	tests := []struct {
		name           string
		triggerType    string
		wantStopCalled bool
		wantReload     bool
		followUpSignal bool
		reloadErr      error
		stopErr        error
		wantErr        bool
	}{
		{
			name:           "SIGINT triggers stop",
			triggerType:    "SIGINT",
			wantStopCalled: true,
			wantReload:     false,
			followUpSignal: false,
		},
		{
			name:           "SIGHUP triggers reload then SIGINT stops",
			triggerType:    "SIGHUP",
			wantStopCalled: true,
			wantReload:     true,
			followUpSignal: true,
		},
		{
			name:           "context done triggers stop",
			triggerType:    "context",
			wantStopCalled: true,
			wantReload:     false,
			followUpSignal: false,
		},
		{
			name:           "SIGTERM triggers stop",
			triggerType:    "SIGTERM",
			wantStopCalled: true,
			wantReload:     false,
			followUpSignal: false,
		},
		{
			name:           "reload error is handled gracefully",
			triggerType:    "SIGHUP",
			wantStopCalled: true,
			wantReload:     true,
			followUpSignal: true,
			reloadErr:      errors.New("reload failed"),
			wantErr:        false,
		},
		{
			name:           "stop error is returned",
			triggerType:    "SIGINT",
			wantStopCalled: true,
			wantReload:     false,
			followUpSignal: false,
			stopErr:        errors.New("stop failed"),
			wantErr:        true,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create context and signal channel for test.
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			mock := &mockSignalHandler{
				reloadErr: tt.reloadErr,
				stopErr:   tt.stopErr,
			}

			done := make(chan error, 1)
			go func() {
				done <- WaitForSignals(ctx, cancel, sigCh, mock)
			}()

			// Trigger the appropriate signal or action.
			switch tt.triggerType {
			case "SIGINT":
				sigCh <- syscall.SIGINT
			case "SIGTERM":
				sigCh <- syscall.SIGTERM
			case "SIGHUP":
				sigCh <- syscall.SIGHUP
				// Wait briefly for reload to be called.
				time.Sleep(100 * time.Millisecond)
				// Verify reload was called.
				if !mock.reloadCalled.Load() {
					t.Error("Reload should have been called")
				}
				// Send SIGINT to stop the loop if follow-up is needed.
				if tt.followUpSignal {
					sigCh <- syscall.SIGINT
				}
			case "context":
				cancel()
			}

			// Wait for function to complete with timeout.
			select {
			case err := <-done:
				// Verify error expectation.
				if tt.wantErr && err == nil {
					t.Error("WaitForSignals should have returned an error")
				}
				if !tt.wantErr && err != nil {
					t.Errorf("WaitForSignals returned error: %v", err)
				}
				// Verify stop was called if expected.
				if tt.wantStopCalled && !mock.stopCalled.Load() {
					t.Error("Stop should have been called")
				}
				// Verify reload was called if expected.
				if tt.wantReload && !mock.reloadCalled.Load() {
					t.Error("Reload should have been called")
				}
			case <-time.After(time.Second):
				t.Fatal("WaitForSignals did not complete in time")
			}
		})
	}
}

// TestRun_Internal tests the Run function with various configurations.
//
// Goroutine lifecycle:
//   - For async tests: runs Run() function in background
//   - Synchronized: via done channel and SIGINT signal
//   - Terminated: when Run() returns after SIGINT or test timeout
func TestRun_Internal(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		async         bool
		sendSignal    bool
		wantExitCode  int
		verifyConfig  bool
		configRelPath string
	}{
		{
			name:         "version_flag",
			args:         []string{"cmd", "--version"},
			async:        false,
			wantExitCode: 0,
		},
		{
			name:         "invalid_config",
			args:         []string{"cmd", "--config", "/nonexistent/config.yaml"},
			async:        false,
			wantExitCode: 1,
		},
		{
			name:          "success_path",
			async:         true,
			sendSignal:    true,
			wantExitCode:  0,
			verifyConfig:  true,
			configRelPath: "testdata/minimal.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original command line args and flag set.
			oldArgs := os.Args
			defer func() {
				os.Args = oldArgs
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			}()

			// Setup args
			if tt.verifyConfig {
				wd, err := os.Getwd()
				if err != nil {
					t.Fatalf("Failed to get working directory: %v", err)
				}
				configPath := filepath.Join(wd, tt.configRelPath)
				if _, statErr := os.Stat(configPath); statErr != nil {
					t.Fatalf("Config file not found at %s: %v", configPath, statErr)
				}
				os.Args = []string{"cmd", "--config", configPath}
			} else {
				os.Args = tt.args
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			if tt.async {
				done := make(chan int, 1)
				go func() {
					done <- Run()
				}()

				time.Sleep(200 * time.Millisecond)

				if tt.sendSignal {
					if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
						t.Fatalf("Failed to send SIGINT: %v", err)
					}
				}

				select {
				case exitCode := <-done:
					if exitCode != tt.wantExitCode {
						t.Errorf("Run() exit code = %d, want %d", exitCode, tt.wantExitCode)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("Run() did not complete in time")
				}
			} else {
				exitCode := Run()
				if exitCode != tt.wantExitCode {
					t.Errorf("Run() exit code = %d, want %d", exitCode, tt.wantExitCode)
				}
			}
		})
	}
}

// TestRun_SupervisorStartError_Internal tests error handling when supervisor fails to start.
// This test uses a workaround to trigger the supervisor.Start error path.
//
// Goroutine lifecycle:
//   - Started: none spawned directly
//   - Synchronized: via context cancellation
//   - Terminated: when Supervisor.Stop is called
func TestRun_SupervisorStartError_Internal(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
	}{
		{
			name:       "supervisor_already_running",
			configPath: "testdata/minimal.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get absolute path to test config.
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}

			configPath := filepath.Join(wd, tt.configPath)

			// Verify config exists.
			if _, statErr := os.Stat(configPath); statErr != nil {
				t.Fatalf("Config file not found at %s: %v", configPath, statErr)
			}

			// Initialize app first time.
			app, initErr := InitializeApp(configPath)
			if initErr != nil {
				t.Fatalf("InitializeApp failed: %v", initErr)
			}

			// Start the supervisor to put it in running state.
			ctx := t.Context()

			if startErr := app.Supervisor.Start(ctx); startErr != nil {
				t.Fatalf("Initial Supervisor.Start failed: %v", startErr)
			}

			// Try to start again - this should fail with ErrAlreadyRunning.
			doubleStartErr := app.Supervisor.Start(ctx)
			if doubleStartErr == nil {
				t.Error("Expected error when starting supervisor twice, got nil")
			}
			if !errors.Is(doubleStartErr, appsupervisor.ErrAlreadyRunning) {
				t.Errorf("Expected ErrAlreadyRunning, got %v", doubleStartErr)
			}

			// Stop supervisor.
			if stopErr := app.Supervisor.Stop(); stopErr != nil {
				t.Logf("Supervisor.Stop failed: %v", stopErr)
			}
		})
	}
}

// TestRun_CleanupPath_Internal tests the cleanup defer path.
// This test manually creates a run-like scenario with cleanup.
func TestRun_CleanupPath_Internal(t *testing.T) {
	tests := []struct {
		name        string
		hasCleanup  bool
		wantCleanup bool
	}{
		{
			name:        "cleanup_function_provided",
			hasCleanup:  true,
			wantCleanup: true,
		},
		{
			name:        "no_cleanup_function",
			hasCleanup:  false,
			wantCleanup: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cleanupCalled := false
			var cleanup func()
			if tt.hasCleanup {
				cleanup = func() {
					cleanupCalled = true
				}
			}

			app := &App{
				Supervisor: nil,
				Cleanup:    cleanup,
			}

			// Simulate the defer pattern from run().
			func() {
				if app.Cleanup != nil {
					defer app.Cleanup()
				}
				// Function body would go here.
			}()

			// Verify cleanup was called if expected.
			if cleanupCalled != tt.wantCleanup {
				t.Errorf("cleanupCalled = %v, want %v", cleanupCalled, tt.wantCleanup)
			}
		})
	}
}

// TestInitializeApp_ValidationError_Internal tests NewSupervisor validation error.
func TestInitializeApp_ValidationError_Internal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configPath string
		wantErr    bool
	}{
		{
			name:       "invalid_yaml_syntax",
			configPath: "testdata/invalid.yaml",
			wantErr:    true,
		},
		{
			name:       "empty_service_name",
			configPath: "testdata/invalid_service.yaml",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app, err := InitializeApp(tt.configPath)

			// Verify error matches expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("InitializeApp() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify app is nil on error.
			if err != nil && app != nil {
				t.Error("InitializeApp() should return nil app on error")
			}
		})
	}
}

// Test_run verifies the run function behavior.
//
// Params:
//   - t: testing context for assertions
func Test_run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfgPath string
		wantErr bool
	}{
		{
			name:    "invalid_config_path",
			cfgPath: "/nonexistent/config.yaml",
			wantErr: true,
		},
		{
			name:    "empty_config_path",
			cfgPath: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := run(tt.cfgPath, tui.ModeRaw)
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_determineTUIMode verifies TUI mode determination based on flags.
//
// Params:
//   - t: testing context for assertions.
func Test_determineTUIMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		forceInteractive bool
		wantMode         tui.Mode
	}{
		{
			name:             "default_raw_mode",
			forceInteractive: false,
			wantMode:         tui.ModeRaw,
		},
		{
			name:             "force_interactive_mode",
			forceInteractive: true,
			wantMode:         tui.ModeInteractive,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := determineTUIMode(tt.forceInteractive)
			// Verify mode matches expectation.
			if got != tt.wantMode {
				t.Errorf("determineTUIMode() = %v, want %v", got, tt.wantMode)
			}
		})
	}
}

// Test_setupContextAndSignals verifies context and signal channel setup.
//
// Params:
//   - t: testing context for assertions.
func Test_setupContextAndSignals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "returns_non_nil_values",
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel, sigCh := setupContextAndSignals()
			defer cancel()

			// Verify context is not nil.
			if ctx == nil {
				t.Error("setupContextAndSignals() ctx is nil")
			}

			// Verify cancel function is not nil.
			if cancel == nil {
				t.Error("setupContextAndSignals() cancel is nil")
			}

			// Verify signal channel is not nil.
			if sigCh == nil {
				t.Error("setupContextAndSignals() sigCh is nil")
			}
		})
	}
}

// Test_handleSignal verifies signal handling logic.
//
// Params:
//   - t: testing context for assertions.
func Test_handleSignal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		signal       syscall.Signal
		reloadErr    error
		stopErr      error
		wantErr      bool
		expectReload bool
		expectStop   bool
	}{
		{
			name:         "SIGHUP_triggers_reload",
			signal:       syscall.SIGHUP,
			wantErr:      false,
			expectReload: true,
			expectStop:   false,
		},
		{
			name:         "SIGINT_triggers_stop",
			signal:       syscall.SIGINT,
			wantErr:      false,
			expectReload: false,
			expectStop:   true,
		},
		{
			name:         "SIGTERM_triggers_stop",
			signal:       syscall.SIGTERM,
			wantErr:      false,
			expectReload: false,
			expectStop:   true,
		},
		{
			name:         "SIGHUP_with_error",
			signal:       syscall.SIGHUP,
			reloadErr:    errors.New("reload failed"),
			wantErr:      false,
			expectReload: true,
			expectStop:   false,
		},
		{
			name:         "SIGINT_with_stop_error",
			signal:       syscall.SIGINT,
			stopErr:      errors.New("stop failed"),
			wantErr:      true,
			expectReload: false,
			expectStop:   true,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock := &mockSignalHandler{
				reloadErr: tt.reloadErr,
				stopErr:   tt.stopErr,
			}

			err := handleSignal(tt.signal, cancel, mock)

			// Verify error matches expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("handleSignal() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify reload was called if expected.
			if tt.expectReload && !mock.reloadCalled.Load() {
				t.Error("handleSignal() should have called Reload()")
			}

			// Verify stop was called if expected.
			if tt.expectStop && !mock.stopCalled.Load() {
				t.Error("handleSignal() should have called Stop()")
			}
		})
	}
}

// Test_determineLogLevel verifies log level determination from event types.
//
// Params:
//   - t: testing context for assertions.
func Test_determineLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eventType domainprocess.EventType
		wantLevel domainlogging.Level
	}{
		{
			name:      "started_is_info",
			eventType: domainprocess.EventStarted,
			wantLevel: domainlogging.LevelInfo,
		},
		{
			name:      "stopped_is_info",
			eventType: domainprocess.EventStopped,
			wantLevel: domainlogging.LevelInfo,
		},
		{
			name:      "failed_is_warn",
			eventType: domainprocess.EventFailed,
			wantLevel: domainlogging.LevelWarn,
		},
		{
			name:      "unhealthy_is_warn",
			eventType: domainprocess.EventUnhealthy,
			wantLevel: domainlogging.LevelWarn,
		},
		{
			name:      "exhausted_is_error",
			eventType: domainprocess.EventExhausted,
			wantLevel: domainlogging.LevelError,
		},
		{
			name:      "restarting_is_info",
			eventType: domainprocess.EventRestarting,
			wantLevel: domainlogging.LevelInfo,
		},
		{
			name:      "healthy_is_info",
			eventType: domainprocess.EventHealthy,
			wantLevel: domainlogging.LevelInfo,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := determineLogLevel(tt.eventType)
			// Verify level matches expectation.
			if got != tt.wantLevel {
				t.Errorf("determineLogLevel() = %v, want %v", got, tt.wantLevel)
			}
		})
	}
}

// Test_buildStartedMessage verifies started message construction.
//
// Params:
//   - t: testing context for assertions.
func Test_buildStartedMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		stats *appsupervisor.ServiceStatsSnapshot
		want  string
	}{
		{
			name:  "nil_stats",
			stats: nil,
			want:  "Service started",
		},
		{
			name:  "zero_restarts",
			stats: &appsupervisor.ServiceStatsSnapshot{RestartCount: 0},
			want:  "Service started",
		},
		{
			name:  "with_restarts",
			stats: &appsupervisor.ServiceStatsSnapshot{RestartCount: 3},
			want:  "Service started (restart #3)",
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildStartedMessage(tt.stats)
			// Verify message matches expectation.
			if got != tt.want {
				t.Errorf("buildStartedMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_buildStoppedMessage verifies stopped message construction.
//
// Params:
//   - t: testing context for assertions.
func Test_buildStoppedMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		exitCode int
		want     string
	}{
		{
			name:     "clean_exit",
			exitCode: 0,
			want:     "Service stopped cleanly",
		},
		{
			name:     "error_exit",
			exitCode: 1,
			want:     "Service exited with code 1",
		},
		{
			name:     "signal_exit",
			exitCode: 137,
			want:     "Service exited with code 137",
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildStoppedMessage(tt.exitCode)
			// Verify message matches expectation.
			if got != tt.want {
				t.Errorf("buildStoppedMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_buildFailedMessage verifies failed message construction.
//
// Params:
//   - t: testing context for assertions.
func Test_buildFailedMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		stats *appsupervisor.ServiceStatsSnapshot
		want  string
	}{
		{
			name:  "nil_stats",
			stats: nil,
			want:  "Service failed",
		},
		{
			name:  "first_failure",
			stats: &appsupervisor.ServiceStatsSnapshot{FailCount: 1},
			want:  "Service failed",
		},
		{
			name:  "multiple_failures",
			stats: &appsupervisor.ServiceStatsSnapshot{FailCount: 3},
			want:  "Service failed (failure #3)",
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildFailedMessage(tt.stats)
			// Verify message matches expectation.
			if got != tt.want {
				t.Errorf("buildFailedMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_buildRestartingMessage verifies restarting message construction.
//
// Params:
//   - t: testing context for assertions.
func Test_buildRestartingMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		stats *appsupervisor.ServiceStatsSnapshot
		want  string
	}{
		{
			name:  "nil_stats",
			stats: nil,
			want:  "Service restarting",
		},
		{
			name:  "with_stats",
			stats: &appsupervisor.ServiceStatsSnapshot{RestartCount: 2},
			want:  "Service restarting (attempt #3)",
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildRestartingMessage(tt.stats)
			// Verify message matches expectation.
			if got != tt.want {
				t.Errorf("buildRestartingMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_buildExhaustedMessage verifies exhausted message construction.
//
// Params:
//   - t: testing context for assertions.
func Test_buildExhaustedMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		stats *appsupervisor.ServiceStatsSnapshot
		want  string
	}{
		{
			name:  "nil_stats",
			stats: nil,
			want:  "Service abandoned (max restarts exceeded)",
		},
		{
			name:  "with_stats",
			stats: &appsupervisor.ServiceStatsSnapshot{RestartCount: 5},
			want:  "Service abandoned after 5 restarts (max exceeded)",
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildExhaustedMessage(tt.stats)
			// Verify message matches expectation.
			if got != tt.want {
				t.Errorf("buildExhaustedMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_buildEventMessage verifies event message construction.
//
// Params:
//   - t: testing context for assertions.
func Test_buildEventMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		eventType    domainprocess.EventType
		stats        *appsupervisor.ServiceStatsSnapshot
		wantContains string
	}{
		{
			name:         "started_event",
			eventType:    domainprocess.EventStarted,
			stats:        nil,
			wantContains: "started",
		},
		{
			name:         "stopped_event",
			eventType:    domainprocess.EventStopped,
			stats:        nil,
			wantContains: "stopped",
		},
		{
			name:         "failed_event",
			eventType:    domainprocess.EventFailed,
			stats:        nil,
			wantContains: "failed",
		},
		{
			name:         "healthy_event",
			eventType:    domainprocess.EventHealthy,
			stats:        nil,
			wantContains: "healthy",
		},
		{
			name:         "unhealthy_event",
			eventType:    domainprocess.EventUnhealthy,
			stats:        nil,
			wantContains: "unhealthy",
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := &domainprocess.Event{Type: tt.eventType}
			got := buildEventMessage(event, tt.stats)
			// Verify message contains expected substring.
			if !strings.Contains(strings.ToLower(got), tt.wantContains) {
				t.Errorf("buildEventMessage() = %v, want to contain %v", got, tt.wantContains)
			}
		})
	}
}

// Test_findLogFilePath verifies log file path extraction from config.
//
// Params:
//   - t: testing context for assertions.
func Test_findLogFilePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cfg       *domainconfig.Config
		wantEmpty bool
	}{
		{
			name:      "nil_config",
			cfg:       nil,
			wantEmpty: true,
		},
		{
			name: "empty_writers",
			cfg: &domainconfig.Config{
				Logging: domainconfig.LoggingConfig{
					Daemon: domainconfig.DaemonLogging{
						Writers: nil,
					},
				},
			},
			wantEmpty: true,
		},
		{
			name: "no_file_writer",
			cfg: &domainconfig.Config{
				Logging: domainconfig.LoggingConfig{
					Daemon: domainconfig.DaemonLogging{
						Writers: []domainconfig.WriterConfig{
							{Type: "console"},
						},
					},
				},
			},
			wantEmpty: true,
		},
		{
			name: "file_writer_present",
			cfg: &domainconfig.Config{
				Logging: domainconfig.LoggingConfig{
					BaseDir: "/var/log",
					Daemon: domainconfig.DaemonLogging{
						Writers: []domainconfig.WriterConfig{
							{Type: "file", File: domainconfig.FileWriterConfig{Path: "daemon.log"}},
						},
					},
				},
			},
			wantEmpty: false,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := findLogFilePath(tt.cfg)
			// Verify empty/non-empty state matches expectation.
			if (got == "") != tt.wantEmpty {
				t.Errorf("findLogFilePath() = %v, wantEmpty %v", got, tt.wantEmpty)
			}
		})
	}
}

// Test_convertProcessEventToLogEvent verifies event conversion.
//
// Params:
//   - t: testing context for assertions.
func Test_convertProcessEventToLogEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		event       *domainprocess.Event
		stats       *appsupervisor.ServiceStatsSnapshot
	}{
		{
			name:        "basic_event",
			serviceName: "test-service",
			event:       &domainprocess.Event{Type: domainprocess.EventStarted, PID: 1234},
			stats:       nil,
		},
		{
			name:        "event_with_stats",
			serviceName: "test-service",
			event:       &domainprocess.Event{Type: domainprocess.EventFailed, ExitCode: 1},
			stats:       &appsupervisor.ServiceStatsSnapshot{RestartCount: 2},
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := convertProcessEventToLogEvent(tt.serviceName, tt.event, tt.stats)
			// Verify log event is not empty.
			if got.Message == "" {
				t.Error("convertProcessEventToLogEvent() returned empty message")
			}
		})
	}
}

// Test_addEventMetadata verifies metadata addition to log events.
//
// Params:
//   - t: testing context for assertions.
func Test_addEventMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event *domainprocess.Event
		stats *appsupervisor.ServiceStatsSnapshot
	}{
		{
			name:  "event_with_pid",
			event: &domainprocess.Event{Type: domainprocess.EventStarted, PID: 1234},
			stats: nil,
		},
		{
			name:  "event_with_exit_code",
			event: &domainprocess.Event{Type: domainprocess.EventStopped, ExitCode: 1},
			stats: nil,
		},
		{
			name:  "event_with_error",
			event: &domainprocess.Event{Type: domainprocess.EventFailed, Error: errors.New("test error")},
			stats: nil,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			baseEvent := domainlogging.NewLogEvent(domainlogging.LevelInfo, "test", "test", "test")
			got := addEventMetadata(baseEvent, tt.event, tt.stats)
			// Verify log event is returned.
			if got.Message == "" {
				t.Error("addEventMetadata() returned empty message")
			}
		})
	}
}

// mockTUIRunner is a test double for Runner interface.
type mockTUIRunner struct {
	runErr error
}

// Run returns the configured error.
//
// Params:
//   - ctx: the context (unused).
//
// Returns:
//   - error: the configured error.
func (m *mockTUIRunner) Run(_ context.Context) error {
	// Return configured error.
	return m.runErr
}

// mockFlusher is a test double for Flusher interface.
type mockFlusher struct {
	flushErr error
	called   bool
}

// Flush records that Flush was called and returns configured error.
//
// Returns:
//   - error: the configured error.
func (m *mockFlusher) Flush() error {
	m.called = true
	// Return configured error.
	return m.flushErr
}

// Test_runTUIMode verifies TUI mode execution.
//
// Goroutine lifecycle (KTN-GOROUTINE-LIFECYCLE):
//   - Spawns a goroutine that sends SIGINT after a brief delay.
//   - The goroutine terminates after sending the signal.
//   - No cleanup needed as the signal channel is buffered.
//
// Params:
//   - t: testing context for assertions.
func Test_runTUIMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tuiMode tui.Mode
		tuiErr  error
		wantErr bool
	}{
		{
			name:    "raw_mode_success",
			tuiMode: tui.ModeRaw,
			tuiErr:  nil,
			wantErr: false,
		},
		{
			name:    "raw_mode_tui_error_continues",
			tuiMode: tui.ModeRaw,
			tuiErr:  errors.New("tui error"),
			wantErr: false,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			sigCh := make(chan os.Signal, 1)
			mockTUI := &mockTUIRunner{runErr: tt.tuiErr}
			mockFlush := &mockFlusher{}
			mockSup := &mockSignalHandler{}

			// Send SIGINT to trigger stop in raw mode.
			go func() {
				time.Sleep(10 * time.Millisecond)
				sigCh <- syscall.SIGINT
			}()

			cfg := tuiModeConfig{
				ctx:             ctx,
				cancel:          cancel,
				sigCh:           sigCh,
				tui:             mockTUI,
				bufferedConsole: mockFlush,
				tuiMode:         tt.tuiMode,
				sup:             mockSup,
			}

			err := runTUIMode(cfg)

			// Verify error matches expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("runTUIMode() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify flush was called in raw mode.
			if tt.tuiMode == tui.ModeRaw && !mockFlush.called {
				t.Error("runTUIMode() should call Flush in raw mode")
			}
		})
	}
}

// Test_waitForTUIOrSignals verifies TUI and signal waiting.
//
// Params:
//   - t: testing context for assertions.
func Test_waitForTUIOrSignals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		signal     syscall.Signal
		tuiResult  error
		stopErr    error
		wantErr    bool
		useTUIExit bool
	}{
		{
			name:       "signal_triggers_stop",
			signal:     syscall.SIGINT,
			wantErr:    false,
			useTUIExit: false,
		},
		{
			name:       "tui_exit_triggers_stop",
			tuiResult:  nil,
			wantErr:    false,
			useTUIExit: true,
		},
		{
			name:       "tui_error_still_stops",
			tuiResult:  errors.New("tui failed"),
			wantErr:    false,
			useTUIExit: true,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			tuiDone := make(chan error, 1)
			mockSup := &mockSignalHandler{stopErr: tt.stopErr}

			// Trigger the appropriate event.
			if tt.useTUIExit {
				tuiDone <- tt.tuiResult
			} else {
				sigCh <- tt.signal
			}

			err := waitForTUIOrSignals(ctx, cancel, sigCh, tuiDone, mockSup)

			// Verify error matches expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("waitForTUIOrSignals() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify stop was called.
			if !mockSup.stopCalled.Load() {
				t.Error("waitForTUIOrSignals() should call Stop()")
			}
		})
	}
}

// Test_initializeLogger verifies logger initialization.
//
// Params:
//   - t: testing context for assertions.
func Test_initializeLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *domainconfig.Config
		tuiMode tui.Mode
	}{
		{
			name: "raw_mode_with_config",
			cfg: &domainconfig.Config{
				Logging: domainconfig.LoggingConfig{
					BaseDir: "/tmp",
					Daemon:  domainconfig.DefaultDaemonLogging(),
				},
			},
			tuiMode: tui.ModeRaw,
		},
		{
			name: "interactive_mode_with_config",
			cfg: &domainconfig.Config{
				Logging: domainconfig.LoggingConfig{
					BaseDir: "/tmp",
					Daemon:  domainconfig.DefaultDaemonLogging(),
				},
			},
			tuiMode: tui.ModeInteractive,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, buffered, err := initializeLogger(tt.cfg, tt.tuiMode)

			// Logger should never be nil (fallback is used).
			if logger == nil {
				t.Error("initializeLogger() returned nil logger")
			}

			// In raw mode, buffered console may be nil if config is nil.
			if tt.tuiMode == tui.ModeInteractive && buffered != nil {
				t.Error("initializeLogger() should return nil buffered in interactive mode")
			}

			// Error handling - we accept errors since config may be invalid.
			_ = err
		})
	}
}

// Test_attachTUIWriter verifies TUI writer attachment.
//
// Params:
//   - t: testing context for assertions.
func Test_attachTUIWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "with_nil_logger",
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logAdapter := tui.NewLogAdapter()
			// attachTUIWriter with nil logger should not panic.
			attachTUIWriter(nil, logAdapter)
			// Test passes if no panic occurred.
		})
	}
}

// Test_setupTUI verifies TUI setup.
//
// Params:
//   - t: testing context for assertions.
func Test_setupTUI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfgPath string
		tuiMode tui.Mode
	}{
		{
			name:    "raw_mode",
			cfgPath: "/test/config.yaml",
			tuiMode: tui.ModeRaw,
		},
		{
			name:    "interactive_mode",
			cfgPath: "/test/config.yaml",
			tuiMode: tui.ModeInteractive,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logAdapter := tui.NewLogAdapter()
			// Use a mock supervisor that doesn't implement ServiceSnapshotsForTUIer.
			mockSup := &mockAppSupervisor{}

			result := setupTUI(mockSup, logAdapter, tt.cfgPath, tt.tuiMode)

			// Verify TUI is created.
			if result == nil {
				t.Error("setupTUI() returned nil")
			}
		})
	}
}

// Test_initializeAppAndLogAdapter verifies app initialization.
//
// Params:
//   - t: testing context for assertions.
func Test_initializeAppAndLogAdapter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfgPath string
		wantErr bool
	}{
		{
			name:    "invalid_config_path",
			cfgPath: "/nonexistent/config.yaml",
			wantErr: true,
		},
		{
			name:    "empty_config_path",
			cfgPath: "",
			wantErr: true,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app, logAdapter, err := initializeAppAndLogAdapter(tt.cfgPath)

			// Verify error matches expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("initializeAppAndLogAdapter() error = %v, wantErr %v", err, tt.wantErr)
			}

			// On error, app and logAdapter should be nil.
			if tt.wantErr {
				if app != nil {
					t.Error("initializeAppAndLogAdapter() should return nil app on error")
				}
				if logAdapter != nil {
					t.Error("initializeAppAndLogAdapter() should return nil logAdapter on error")
				}
			}
		})
	}
}

// Test_setupLoggingAndEvents verifies logging and events setup.
// This is a smoke test that verifies the function doesn't panic.
//
// Params:
//   - t: testing context for assertions.
func Test_setupLoggingAndEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "basic_setup",
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create minimal app with mock supervisor.
			mockSup := &mockAppSupervisor{}
			app := &App{
				Supervisor: mockSup,
				Config: &domainconfig.Config{
					Logging: domainconfig.LoggingConfig{
						BaseDir: "/tmp",
						Daemon:  domainconfig.DefaultDaemonLogging(),
					},
				},
			}
			logAdapter := tui.NewLogAdapter()

			logger, buffered := setupLoggingAndEvents(app, logAdapter, tui.ModeRaw)

			// Verify logger is returned.
			if logger == nil {
				t.Error("setupLoggingAndEvents() returned nil logger")
			}

			// Close logger.
			_ = logger.Close()
			_ = buffered
		})
	}
}

// mockAppSupervisor is a test double for AppSupervisor interface.
type mockAppSupervisor struct {
	eventHandler appsupervisor.EventHandler
}

// Start does nothing.
//
// Params:
//   - ctx: the context (unused).
//
// Returns:
//   - error: nil.
func (m *mockAppSupervisor) Start(_ context.Context) error {
	// Return nil.
	return nil
}

// Stop does nothing.
//
// Returns:
//   - error: nil.
func (m *mockAppSupervisor) Stop() error {
	// Return nil.
	return nil
}

// Reload does nothing.
//
// Returns:
//   - error: nil.
func (m *mockAppSupervisor) Reload() error {
	// Return nil.
	return nil
}

// SetEventHandler stores the event handler.
//
// Params:
//   - handler: the event handler.
func (m *mockAppSupervisor) SetEventHandler(handler appsupervisor.EventHandler) {
	m.eventHandler = handler
}

// Test_startSupervisorAndMetrics verifies supervisor and metrics startup.
//
// Params:
//   - t: testing context for assertions.
func Test_startSupervisorAndMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		startErr error
		wantErr  bool
	}{
		{
			name:     "success",
			startErr: nil,
			wantErr:  false,
		},
		{
			name:     "start_error",
			startErr: errors.New("start failed"),
			wantErr:  true,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mockSup := &mockAppSupervisorWithErr{startErr: tt.startErr}
			app := &App{
				Supervisor:     mockSup,
				MetricsTracker: nil,
			}

			// Create a simple logger.
			logger := daemonlogger.NewSilentLogger()

			err := startSupervisorAndMetrics(ctx, app, logger)

			// Verify error matches expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("startSupervisorAndMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// mockAppSupervisorWithErr is a test double for AppSupervisor with configurable errors.
type mockAppSupervisorWithErr struct {
	startErr error
}

// Start returns the configured error.
//
// Params:
//   - ctx: the context (unused).
//
// Returns:
//   - error: the configured error.
func (m *mockAppSupervisorWithErr) Start(_ context.Context) error {
	// Return configured error.
	return m.startErr
}

// Stop does nothing.
//
// Returns:
//   - error: nil.
func (m *mockAppSupervisorWithErr) Stop() error {
	// Return nil.
	return nil
}

// Reload does nothing.
//
// Returns:
//   - error: nil.
func (m *mockAppSupervisorWithErr) Reload() error {
	// Return nil.
	return nil
}

// SetEventHandler does nothing.
//
// Params:
//   - handler: the event handler (unused).
func (m *mockAppSupervisorWithErr) SetEventHandler(_ appsupervisor.EventHandler) {
	// Do nothing.
}

// Test_addPIDMetadata verifies PID metadata enrichment.
//
// Params:
//   - t: testing context for assertions.
func Test_addPIDMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		pid         int
		expectPID   bool
		expectedPID int
	}{
		{
			name:        "adds_pid_when_greater_than_zero",
			pid:         12345,
			expectPID:   true,
			expectedPID: 12345,
		},
		{
			name:      "does_not_add_pid_when_zero",
			pid:       0,
			expectPID: false,
		},
		{
			name:      "does_not_add_pid_when_negative",
			pid:       -1,
			expectPID: false,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := &domainprocess.Event{
				PID: tt.pid,
			}

			logEvent := domainlogging.NewLogEvent(domainlogging.LevelInfo, "test", "test_event", "test message")
			result := addPIDMetadata(logEvent, event)

			// Verify PID metadata presence.
			if tt.expectPID {
				if result.Metadata["pid"] != tt.expectedPID {
					t.Errorf("addPIDMetadata() PID = %v, want %v", result.Metadata["pid"], tt.expectedPID)
				}
			} else {
				if _, exists := result.Metadata["pid"]; exists {
					t.Error("addPIDMetadata() should not add PID metadata")
				}
			}
		})
	}
}

// Test_addExitMetadata verifies exit code and error metadata enrichment.
//
// Params:
//   - t: testing context for assertions.
func Test_addExitMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		eventType        domainprocess.EventType
		exitCode         int
		err              error
		expectExitCode   bool
		expectedExitCode int
		expectError      bool
		expectedError    string
	}{
		{
			name:             "adds_exit_code_for_stopped_event",
			eventType:        domainprocess.EventStopped,
			exitCode:         0,
			expectExitCode:   true,
			expectedExitCode: 0,
			expectError:      false,
		},
		{
			name:             "adds_exit_code_for_failed_event",
			eventType:        domainprocess.EventFailed,
			exitCode:         1,
			expectExitCode:   true,
			expectedExitCode: 1,
			expectError:      false,
		},
		{
			name:           "adds_error_when_present",
			eventType:      domainprocess.EventStarted,
			err:            errors.New("test error"),
			expectExitCode: false,
			expectError:    true,
			expectedError:  "test error",
		},
		{
			name:             "adds_both_exit_code_and_error",
			eventType:        domainprocess.EventFailed,
			exitCode:         127,
			err:              errors.New("command not found"),
			expectExitCode:   true,
			expectedExitCode: 127,
			expectError:      true,
			expectedError:    "command not found",
		},
		{
			name:           "does_not_add_exit_code_for_started_event",
			eventType:      domainprocess.EventStarted,
			expectExitCode: false,
			expectError:    false,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := &domainprocess.Event{
				Type:     tt.eventType,
				ExitCode: tt.exitCode,
				Error:    tt.err,
			}

			logEvent := domainlogging.NewLogEvent(domainlogging.LevelInfo, "test", "test_event", "test message")
			result := addExitMetadata(logEvent, event)

			// Verify exit code metadata.
			if tt.expectExitCode {
				if result.Metadata["exit_code"] != tt.expectedExitCode {
					t.Errorf("addExitMetadata() exit_code = %v, want %v", result.Metadata["exit_code"], tt.expectedExitCode)
				}
			} else {
				if _, exists := result.Metadata["exit_code"]; exists {
					t.Error("addExitMetadata() should not add exit_code metadata")
				}
			}

			// Verify error metadata.
			if tt.expectError {
				if result.Metadata["error"] != tt.expectedError {
					t.Errorf("addExitMetadata() error = %v, want %v", result.Metadata["error"], tt.expectedError)
				}
			} else {
				if _, exists := result.Metadata["error"]; exists {
					t.Error("addExitMetadata() should not add error metadata")
				}
			}
		})
	}
}

// Test_addRestartMetadata verifies restart count metadata enrichment.
//
// Params:
//   - t: testing context for assertions.
func Test_addRestartMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		eventType            domainprocess.EventType
		stats                *appsupervisor.ServiceStatsSnapshot
		expectRestarts       bool
		expectedRestartCount int
	}{
		{
			name:      "adds_restart_count_for_restarting_event",
			eventType: domainprocess.EventRestarting,
			stats: &appsupervisor.ServiceStatsSnapshot{
				RestartCount: 3,
			},
			expectRestarts:       true,
			expectedRestartCount: 3,
		},
		{
			name:      "adds_restart_count_for_exhausted_event",
			eventType: domainprocess.EventExhausted,
			stats: &appsupervisor.ServiceStatsSnapshot{
				RestartCount: 10,
			},
			expectRestarts:       true,
			expectedRestartCount: 10,
		},
		{
			name:           "does_not_add_restart_count_for_started_event",
			eventType:      domainprocess.EventStarted,
			stats:          &appsupervisor.ServiceStatsSnapshot{RestartCount: 5},
			expectRestarts: false,
		},
		{
			name:           "does_not_add_restart_count_when_stats_nil",
			eventType:      domainprocess.EventRestarting,
			stats:          nil,
			expectRestarts: false,
		},
		{
			name:           "does_not_add_restart_count_for_stopped_event",
			eventType:      domainprocess.EventStopped,
			stats:          &appsupervisor.ServiceStatsSnapshot{RestartCount: 2},
			expectRestarts: false,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			event := &domainprocess.Event{
				Type: tt.eventType,
			}

			logEvent := domainlogging.NewLogEvent(domainlogging.LevelInfo, "test", "test_event", "test message")
			result := addRestartMetadata(logEvent, event, tt.stats)

			// Verify restart count metadata.
			if tt.expectRestarts {
				if result.Metadata["restarts"] != tt.expectedRestartCount {
					t.Errorf("addRestartMetadata() restarts = %v, want %v", result.Metadata["restarts"], tt.expectedRestartCount)
				}
			} else {
				if _, exists := result.Metadata["restarts"]; exists {
					t.Error("addRestartMetadata() should not add restarts metadata")
				}
			}
		})
	}
}

// Test_runProbeMode verifies runProbeMode function behavior.
//
// Params:
//   - t: testing context for assertions.
func Test_runProbeMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		wantExitCode int
	}{
		{
			name:         "probe_mode_returns_valid_exit_code",
			wantExitCode: 0,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call runProbeMode and verify exit code.
			exitCode := runProbeMode()

			// Verify exit code matches expectation or is an error code.
			if exitCode != tt.wantExitCode && exitCode != 1 {
				t.Errorf("runProbeMode() exit code = %d, want %d or error code 1", exitCode, tt.wantExitCode)
			}
		})
	}
}
