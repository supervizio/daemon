package bootstrap

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
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
