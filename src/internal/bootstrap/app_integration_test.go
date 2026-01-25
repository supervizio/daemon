//go:build integration
// +build integration

package bootstrap_test

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/bootstrap"
)

// TestRunWithConfig_Integration tests the full run path with signal handling.
// This test requires the integration build tag.
// Goroutine lifecycle:
//   - Main test goroutine spawns RunWithConfig in background.
//   - RunWithConfig starts supervisor and waits for signals.
//   - Test sends SIGINT after brief delay to trigger graceful shutdown.
//   - All goroutines clean up when supervisor stops.
func TestRunWithConfig_Integration(t *testing.T) {
	// Get absolute path to minimal test config.
	configPath, err := filepath.Abs(filepath.Join("testdata", "minimal.yaml"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Channel to capture result.
	done := make(chan error, 1)

	// Start RunWithConfig in background.
	go func() {
		done <- bootstrap.RunWithConfig(configPath)
	}()

	// Give supervisor time to start.
	time.Sleep(200 * time.Millisecond)

	// Send SIGINT to current process to trigger shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)

	// Trigger the signal.
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("Failed to send SIGINT: %v", err)
	}

	// Wait for completion with timeout.
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("RunWithConfig returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("RunWithConfig did not complete in time")
	}
}

// TestRun_Integration tests the full Run function with signal handling.
// This test requires the integration build tag.
// Goroutine lifecycle:
//   - Main test goroutine spawns Run in background.
//   - Run parses flags, starts supervisor, and waits for signals.
//   - Test sends SIGINT after brief delay to trigger graceful shutdown.
//   - All goroutines clean up when supervisor stops.
func TestRun_Integration(t *testing.T) {
	// Save and restore original command line args and flag set.
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	// Get absolute path to minimal test config.
	configPath, err := filepath.Abs(filepath.Join("testdata", "minimal.yaml"))
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Set args with minimal config.
	os.Args = []string{"cmd", "--config", configPath}

	// Reset flags to parse new args.
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Channel to capture result.
	done := make(chan int, 1)

	// Start Run in background.
	go func() {
		done <- bootstrap.Run()
	}()

	// Give supervisor time to start.
	time.Sleep(200 * time.Millisecond)

	// Send SIGINT to current process to trigger shutdown.
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("Failed to send SIGINT: %v", err)
	}

	// Wait for completion with timeout.
	select {
	case exitCode := <-done:
		if exitCode != 0 {
			t.Errorf("Run() exit code = %d, want 0", exitCode)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not complete in time")
	}
}
