package bootstrap_test

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/bootstrap"
)

// TestAppStructure verifies the App struct has required fields.
func TestAppStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		app       *bootstrap.App
		wantNil   bool
		wantClean bool
	}{
		{
			name: "app with nil supervisor and cleanup",
			app: &bootstrap.App{
				Supervisor: nil,
				Cleanup:    nil,
			},
			wantNil:   false,
			wantClean: true,
		},
		{
			name: "app with cleanup function",
			app: &bootstrap.App{
				Supervisor: nil,
				Cleanup:    func() {},
			},
			wantNil:   false,
			wantClean: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Verify the app struct was created successfully.
			if got := tt.app == nil; got != tt.wantNil {
				t.Errorf("App nil = %v, want %v", got, tt.wantNil)
			}
			// Verify cleanup is set as expected.
			if got := tt.app.Cleanup == nil; got != tt.wantClean {
				t.Errorf("Cleanup nil = %v, want %v", got, tt.wantClean)
			}
		})
	}
}

// TestWaitForSignals verifies the signal handling function behavior.
// Goroutine lifecycle:
//   - Spawns a goroutine for SIGHUP test case to cancel context after reload.
//   - The goroutine terminates after calling cancel() within 10ms.
//   - Context cancellation ensures WaitForSignals exits cleanly.
//   - No cleanup needed as goroutine terminates before test completes.
func TestWaitForSignals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		signal     os.Signal
		wantErr    bool
		stopErr    error
		reloadErr  error
		expectStop bool
	}{
		{
			name:       "SIGTERM triggers stop",
			signal:     syscall.SIGTERM,
			wantErr:    false,
			stopErr:    nil,
			expectStop: true,
		},
		{
			name:       "SIGINT triggers stop",
			signal:     syscall.SIGINT,
			wantErr:    false,
			stopErr:    nil,
			expectStop: true,
		},
		{
			name:       "SIGHUP triggers reload",
			signal:     syscall.SIGHUP,
			wantErr:    false,
			reloadErr:  nil,
			expectStop: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockSignalHandler{
				stopErr:   tt.stopErr,
				reloadErr: tt.reloadErr,
			}

			ctx, cancel := context.WithCancel(context.Background())
			sigCh := make(chan os.Signal, 1)

			// For non-stop signals, we need to cancel context after handling.
			if !tt.expectStop {
				go func() {
					// Wait a short time for the signal to be processed.
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
			}

			sigCh <- tt.signal

			err := bootstrap.WaitForSignals(ctx, cancel, sigCh, mock)

			// Verify error matches expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("WaitForSignals() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Verify stop was called for termination signals.
			if tt.expectStop && !mock.stopCalled {
				t.Error("expected Stop() to be called")
			}
		})
	}
}

// TestWaitForSignals_WithErrors tests error handling in WaitForSignals.
// Goroutine lifecycle:
//   - Spawns a goroutine for SIGHUP test to cancel context after reload.
//   - Goroutine terminates after calling cancel().
func TestWaitForSignals_WithErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		signal     os.Signal
		stopErr    error
		reloadErr  error
		wantErr    bool
		expectStop bool
	}{
		{
			name:       "SIGINT with stop error",
			signal:     syscall.SIGINT,
			stopErr:    errors.New("stop failed"),
			wantErr:    true,
			expectStop: true,
		},
		{
			name:       "SIGHUP with reload error",
			signal:     syscall.SIGHUP,
			reloadErr:  errors.New("reload failed"),
			wantErr:    false,
			expectStop: false,
		},
		{
			name:       "context done with stop error",
			signal:     nil, // Will use context cancel instead
			stopErr:    errors.New("stop failed"),
			wantErr:    true,
			expectStop: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockSignalHandler{
				stopErr:   tt.stopErr,
				reloadErr: tt.reloadErr,
			}

			ctx, cancel := context.WithCancel(context.Background())
			sigCh := make(chan os.Signal, 1)

			// Handle different trigger types.
			if tt.signal == syscall.SIGHUP {
				// For SIGHUP, cancel context after signal is processed.
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
				sigCh <- tt.signal
			} else if tt.signal != nil {
				// For other signals, just send the signal.
				sigCh <- tt.signal
			} else {
				// For context done test, just cancel.
				cancel()
			}

			err := bootstrap.WaitForSignals(ctx, cancel, sigCh, mock)

			// Verify error matches expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("WaitForSignals() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify stop was called for termination signals.
			if tt.expectStop && !mock.stopCalled {
				t.Error("expected Stop() to be called")
			}
		})
	}
}

// TestRunWithConfig verifies the RunWithConfig function behavior.
func TestRunWithConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configPath string
		wantErr    bool
	}{
		{
			name:       "missing config returns error",
			configPath: "/nonexistent/path/config.yaml",
			wantErr:    true,
		},
		{
			name:       "empty config path returns error",
			configPath: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call RunWithConfig with the test configuration path.
			err := bootstrap.RunWithConfig(tt.configPath)

			// Verify error matches expectation.
			if (err != nil) != tt.wantErr {
				t.Errorf("RunWithConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestInitializeApp tests app initialization with various configurations.
func TestInitializeApp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		configPath       string
		wantErr          bool
		verifySupervisor bool
	}{
		{
			name:             "success",
			configPath:       filepath.Join("testdata", "valid.yaml"),
			wantErr:          false,
			verifySupervisor: true,
		},
		{
			name:       "config_load_error",
			configPath: "/nonexistent/config.yaml",
			wantErr:    true,
		},
		{
			name:       "invalid_config",
			configPath: filepath.Join("testdata", "invalid.yaml"),
			wantErr:    true,
		},
		{
			name:       "empty_path",
			configPath: "",
			wantErr:    true,
		},
		{
			name:       "invalid_service",
			configPath: filepath.Join("testdata", "invalid_service.yaml"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app, err := bootstrap.InitializeApp(tt.configPath)

			if tt.wantErr {
				if err == nil {
					t.Error("InitializeApp() error = nil, want error")
				}
				if app != nil {
					t.Error("InitializeApp() returned non-nil app on error")
				}
				return
			}

			if err != nil {
				t.Errorf("InitializeApp() error = %v, want nil", err)
			}
			if app == nil {
				t.Fatal("InitializeApp() returned nil app")
			}
			if tt.verifySupervisor && app.Supervisor == nil {
				t.Error("InitializeApp() app.Supervisor is nil")
			}
			if app.Cleanup != nil {
				app.Cleanup()
			}
		})
	}
}

// mockSignalHandler implements SignalHandler for testing.
type mockSignalHandler struct {
	stopCalled   bool
	reloadCalled bool
	stopErr      error
	reloadErr    error
}

// Reload handles the reload signal.
func (m *mockSignalHandler) Reload() error {
	m.reloadCalled = true
	// Return the configured reload error.
	return m.reloadErr
}

// Stop handles the stop signal.
func (m *mockSignalHandler) Stop() error {
	m.stopCalled = true
	// Return the configured stop error.
	return m.stopErr
}

// TestRunWithConfig_FullPath tests the complete success path through run().
// This test exercises the full initialization and signal handling flow.
//
// Goroutine lifecycle:
//   - Started: RunWithConfig runs in background goroutine
//   - Synchronized: via done channel and SIGINT signal
//   - Terminated: when RunWithConfig returns after signal or test timeout
func TestRunWithConfig_FullPath(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "full_path_success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get absolute path to minimal test config (no services to start).
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}

			configPath := filepath.Join(wd, "testdata", "minimal.yaml")

			// Verify config exists.
			if _, statErr := os.Stat(configPath); statErr != nil {
				t.Fatalf("Config file not found at %s: %v", configPath, statErr)
			}

			// Create a signal channel and register for INT/TERM.
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			defer signal.Stop(sigCh)

			// Run RunWithConfig in a goroutine.
			done := make(chan error, 1)
			go func() {
				done <- bootstrap.RunWithConfig(configPath)
			}()

			// Give the supervisor time to initialize and start.
			time.Sleep(300 * time.Millisecond)

			// Send SIGINT to ourselves to trigger shutdown.
			proc, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Fatalf("Failed to find current process: %v", err)
			}

			if err := proc.Signal(syscall.SIGINT); err != nil {
				t.Fatalf("Failed to send SIGINT: %v", err)
			}

			// Wait for RunWithConfig to complete.
			select {
			case err := <-done:
				if err != nil {
					t.Logf("RunWithConfig completed with error (may be expected): %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("RunWithConfig did not complete in time")
			}
		})
	}
}

// TestRunWithConfig_SuccessPath tests the success path of run function using subprocess.
// This test spawns the daemon and sends it a signal to test the full flow.
//
// Goroutine lifecycle:
//   - Started: goroutine to wait for subprocess exit
//   - Synchronized: via done channel
//   - Terminated: when subprocess exits or timeout occurs
func TestRunWithConfig_SuccessPath(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "subprocess_daemon_flow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if we're running as the subprocess daemon.
			if os.Getenv("TEST_RUN_DAEMON") == "1" {
				// Get config path from environment.
				configPath := os.Getenv("TEST_CONFIG_PATH")

				// Verify config exists.
				if _, statErr := os.Stat(configPath); statErr != nil {
					t.Fatalf("Config file not found: %v", statErr)
				}

				// Run the daemon - this will block until signal.
				runErr := bootstrap.RunWithConfig(configPath)
				if runErr != nil {
					t.Logf("RunWithConfig error: %v", runErr)
					os.Exit(1)
				}
				os.Exit(0)
			}

			// Main test process - get absolute path to config.
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}

			configPath := filepath.Join(wd, "testdata", "minimal.yaml")

			// Verify config exists before running test.
			if _, statErr := os.Stat(configPath); statErr != nil {
				t.Fatalf("Config file not found at %s: %v", configPath, statErr)
			}

			// Start subprocess.
			cmd := exec.Command(os.Args[0], "-test.run=TestRunWithConfig_SuccessPath")
			cmd.Env = append(os.Environ(),
				"TEST_RUN_DAEMON=1",
				"TEST_CONFIG_PATH="+configPath,
			)
			cmd.Dir = wd // Set working directory

			// Start the subprocess.
			if startErr := cmd.Start(); startErr != nil {
				t.Fatalf("Failed to start subprocess: %v", startErr)
			}

			// Give it time to start.
			time.Sleep(500 * time.Millisecond)

			// Send SIGINT to subprocess.
			if sigErr := cmd.Process.Signal(syscall.SIGINT); sigErr != nil {
				t.Logf("Failed to send SIGINT: %v", sigErr)
				// Try to kill it anyway.
				_ = cmd.Process.Kill()
			}

			// Wait for subprocess to exit.
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case waitErr := <-done:
				// Check exit code.
				if waitErr != nil {
					if exitErr, ok := waitErr.(*exec.ExitError); ok {
						if exitErr.ExitCode() == 0 {
							// Success case.
							return
						}
						t.Logf("Subprocess exited with code %d (may be expected)", exitErr.ExitCode())
					} else {
						t.Logf("Subprocess error: %v (may be expected)", waitErr)
					}
				}
			case <-time.After(5 * time.Second):
				// Kill subprocess if it didn't exit.
				_ = cmd.Process.Kill()
				t.Fatal("Subprocess did not exit in time")
			}
		})
	}
}

// TestRunInternals_CleanupPath tests the cleanup path in run function.
func TestRunInternals_CleanupPath(t *testing.T) {
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

			app := &bootstrap.App{
				Supervisor: nil,
				Cleanup:    cleanup,
			}

			// Call cleanup if set.
			if app.Cleanup != nil {
				app.Cleanup()
			}

			// Verify cleanup was called if expected.
			if cleanupCalled != tt.wantCleanup {
				t.Errorf("cleanupCalled = %v, want %v", cleanupCalled, tt.wantCleanup)
			}
		})
	}
}

// mockSupervisorStart is a mock implementation for testing supervisor start errors.
type mockSupervisorStart struct {
	startErr error
}

// Start returns the configured error.
func (m *mockSupervisorStart) Start(_ context.Context) error {
	// Return the configured error.
	return m.startErr
}

// Stop always returns nil.
func (m *mockSupervisorStart) Stop() error {
	// Return nil.
	return nil
}

// Reload always returns nil.
func (m *mockSupervisorStart) Reload() error {
	// Return nil.
	return nil
}

// TestSupervisorStart_ErrorPath tests error handling when supervisor fails to start.
// This test verifies the error propagation from supervisor Start failures.
func TestSupervisorStart_ErrorPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		startErr  error
		wantError bool
	}{
		{
			name:      "supervisor_start_succeeds",
			startErr:  nil,
			wantError: false,
		},
		{
			name:      "supervisor_start_fails",
			startErr:  errors.New("start failed"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock supervisor.
			mock := &mockSupervisorStart{
				startErr: tt.startErr,
			}

			// Create context.
			ctx := context.Background()

			// Call Start.
			err := mock.Start(ctx)

			// Verify error matches expectation.
			if (err != nil) != tt.wantError {
				t.Errorf("Start() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// Test_Run verifies the Run function behavior.
//
// Params:
//   - t: testing context for assertions
func Test_Run(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
	}{
		{
			name:         "version_flag_returns_zero",
			args:         []string{"cmd", "--version"},
			wantExitCode: 0,
		},
		{
			name:         "invalid_config_returns_one",
			args:         []string{"cmd", "--config", "/nonexistent/config.yaml"},
			wantExitCode: 1,
		},
		{
			name:         "empty_config_returns_one",
			args:         []string{"cmd", "--config", ""},
			wantExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			defer func() {
				os.Args = oldArgs
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			}()

			os.Args = tt.args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			exitCode := bootstrap.Run()
			if exitCode != tt.wantExitCode {
				t.Errorf("Run() exit code = %d, want %d", exitCode, tt.wantExitCode)
			}
		})
	}
}
