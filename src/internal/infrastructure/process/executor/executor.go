//go:build unix

// Package executor provides infrastructure adapters for OS process execution.
// It implements the domain process interfaces using Unix system calls.
package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	domain "github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/domain/shared"
	"github.com/kodflow/daemon/internal/infrastructure/process/control"
	"github.com/kodflow/daemon/internal/infrastructure/process/credentials"
)

// Waiter is a minimal interface for waiting on commands.
// It abstracts exec.Cmd.Wait() for testability.
type Waiter interface {
	Wait() error
}

// Process is an interface for process operations.
// It abstracts os.Process methods for testability.
type Process interface {
	Signal(sig os.Signal) error
	Kill() error
	Wait() (*os.ProcessState, error)
}

// ProcessFinder is a function type for finding processes.
// It abstracts os.FindProcess for testability.
type ProcessFinder func(pid int) (Process, error)

// Executor implements the domain.Executor interface for Unix systems.
// It wraps the standard library exec.Cmd to provide process lifecycle management
// with support for credentials, environment variables, and signal handling.
type Executor struct {
	credentials credentials.CredentialManager
	process     control.ProcessControl
	findProcess ProcessFinder
}

// NewExecutor creates a new Unix process executor with default dependencies.
//
// Returns:
//   - *Executor: a configured executor instance.
func NewExecutor() *Executor {
	// Create and return executor with default dependencies.
	return &Executor{
		credentials: credentials.New(),
		process:     control.New(),
		findProcess: defaultFindProcess,
	}
}

// New creates a new Unix process executor with default dependencies.
//
// Returns:
//   - *Executor: a configured executor instance.
func New() *Executor {
	// Create and return executor with default dependencies.
	return &Executor{
		credentials: credentials.New(),
		process:     control.New(),
		findProcess: defaultFindProcess,
	}
}

// NewWithDeps creates a new Unix process executor with injected dependencies.
// This is the primary constructor for Wire dependency injection.
//
// Params:
//   - creds: the credential manager for user/group resolution.
//   - proc: the process control for process group management.
//
// Returns:
//   - *Executor: a configured executor instance.
func NewWithDeps(creds credentials.CredentialManager, proc control.ProcessControl) *Executor {
	// Create and return executor with injected dependencies.
	return &Executor{
		credentials: creds,
		process:     proc,
		findProcess: defaultFindProcess,
	}
}

// NewWithOptions creates a new Unix process executor with custom options.
// This constructor is useful for testing with mock implementations.
//
// Params:
//   - creds: the credential manager for user/group resolution.
//   - proc: the process control for process group management.
//   - finder: the process finder function to use.
//
// Returns:
//   - *Executor: a configured executor instance.
func NewWithOptions(creds credentials.CredentialManager, proc control.ProcessControl, finder ProcessFinder) *Executor {
	// Create and return executor with custom options for testing.
	return &Executor{
		credentials: creds,
		process:     proc,
		findProcess: finder,
	}
}

// Start starts a process with the given specification and returns its PID.
// This method spawns a background goroutine to monitor the process lifecycle.
// The goroutine terminates when the spawned process exits (normally or via signal).
// Resources: The goroutine uses a buffered channel (size 1) for the exit result.
// Thread-safety: The wait channel is safe to read from any goroutine.
// Cleanup: The channel is closed after sending the result.
//
// Params:
//   - ctx: context for command cancellation.
//   - spec: process specification containing command, args, env, and credentials.
//
// Returns:
//   - int: the process ID of the started process.
//   - <-chan domain.ExitResult: channel that receives exit result when process terminates.
//   - error: any error encountered during process start.
func (e *Executor) Start(ctx context.Context, spec domain.Spec) (pid int, wait <-chan domain.ExitResult, err error) {
	// Build the command from specification
	cmd, err := e.buildCommand(ctx, spec)
	// Check if command building failed
	if err != nil {
		// Return error if command could not be built
		return 0, nil, err
	}

	// Configure user and group credentials if specified
	if err := e.configureCredentials(cmd, spec.User, spec.Group); err != nil {
		// Return error if credentials configuration failed
		return 0, nil, err
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		// Return error if process failed to start
		return 0, nil, fmt.Errorf("starting process: %w", err)
	}

	// Create channel for exit result notification
	waitCh := make(chan domain.ExitResult, 1)
	// Launch goroutine to wait for process completion
	go e.waitForProcess(cmd, waitCh)

	// Return process ID and wait channel
	return cmd.Process.Pid, waitCh, nil
}

// waitForProcess waits for the command to complete and sends the result.
//
// Params:
//   - cmd: the command waiter interface to wait on
//   - wait: channel to send the exit result
//
// Returns:
//   - None (sends result via channel)
func (e *Executor) waitForProcess(cmd Waiter, wait chan<- domain.ExitResult) {
	// Wait for process to complete
	err := cmd.Wait()
	// Initialize result with zero exit code
	result := domain.ExitResult{}
	// Check if process exited with an error
	if err != nil {
		// Try to extract exit code from error
		var exitErr *exec.ExitError
		// Handle exit error type to extract exit code
		if errors.As(err, &exitErr) {
			// Set exit code from exit error
			result.Code = exitErr.ExitCode()
		} else {
			// Set error code for non-exit errors
			result.Code = -1
			result.Error = err
		}
	}
	// Send result to wait channel
	wait <- result
	// Close channel to signal completion
	close(wait)
}

// Stop gracefully stops the process with the given PID using SIGTERM.
// If the process does not exit within the timeout, it is forcefully killed.
// This method launches a background goroutine to wait for process exit.
// The goroutine terminates when the process exits or is killed.
// Resources: Uses a buffered channel (size 1) for completion signaling.
// Thread-safety: The done channel is managed internally by this method.
// Cleanup: The goroutine always terminates within the timeout duration.
//
// Params:
//   - pid: the process ID to stop.
//   - timeout: maximum time to wait for graceful shutdown before killing.
//
// Returns:
//   - error: any error encountered during stop operation.
func (e *Executor) Stop(pid int, timeout time.Duration) error {
	// Find the process by PID
	proc, err := e.findProcess(pid)
	// Check if process lookup failed
	if err != nil {
		// Return error if process not found
		return fmt.Errorf("finding process: %w", err)
	}

	// Send SIGTERM for graceful shutdown
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		// Return error if signal could not be sent
		return fmt.Errorf("sending SIGTERM: %w", err)
	}

	// Create channel for process wait result
	done := make(chan error, 1)
	// Launch goroutine to wait for process exit using inline adapter
	go func() {
		// Wait for process to exit and capture any error
		_, err := proc.Wait()
		// Send completion signal with error status
		done <- err
	}()

	// Use NewTimer instead of time.After to allow proper cleanup.
	// time.After creates a timer that won't be GC'd until it fires.
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	// Wait for process to exit or timeout
	select {
	// Handle process exit completion
	case <-done:
		// Return nil on successful exit
		return nil
	// Handle timeout case
	case <-timer.C:
		// Force kill after timeout exceeded
		if err := proc.Kill(); err != nil {
			// Return error if kill failed
			return fmt.Errorf("killing process: %w", err)
		}
		// Wait for process to actually terminate
		<-done
		// Return nil after forced kill
		return nil
	}
}

// Signal sends a signal to the process with the given PID.
//
// Params:
//   - pid: the process ID to signal
//   - sig: the signal to send
//
// Returns:
//   - error: any error encountered during signal delivery
func (e *Executor) Signal(pid int, sig os.Signal) error {
	// Find the process by PID
	proc, err := e.findProcess(pid)
	// Check if process lookup failed
	if err != nil {
		// Return error if process not found
		return fmt.Errorf("finding process: %w", err)
	}
	// Send the signal and return result
	return proc.Signal(sig)
}

// buildCommand creates an exec.Cmd from the specification.
//
// Params:
//   - ctx: context for command cancellation
//   - spec: process specification with command, args, and environment
//
// Returns:
//   - *exec.Cmd: configured command ready to execute
//   - error: any error encountered during command building
func (e *Executor) buildCommand(ctx context.Context, spec domain.Spec) (*exec.Cmd, error) {
	// Split command string into parts
	parts := strings.Fields(spec.Command)
	// Check if command is empty
	if len(parts) == 0 {
		// Return error for empty command
		return nil, shared.ErrEmptyCommand
	}

	// Initialize args slice with preallocated capacity
	args := make([]string, 0, len(parts)-1+len(spec.Args))
	args = append(args, parts[1:]...)
	// Append additional args from spec
	args = append(args, spec.Args...)

	// Create command with context for cancellation support
	cmd := TrustedCommand(ctx, parts[0], args...)

	// Set working directory if specified
	if spec.Dir != "" {
		// Configure command working directory
		cmd.Dir = spec.Dir
	}

	// Initialize environment with current process environment.
	// Pre-allocate capacity for spec.Env additions.
	baseEnv := os.Environ()
	cmd.Env = make([]string, len(baseEnv), len(baseEnv)+len(spec.Env))
	copy(cmd.Env, baseEnv)
	// Append custom environment variables using string concatenation (faster than fmt.Sprintf).
	for k, v := range spec.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	// Note: stdout/stderr are inherited from parent process by default.
	// I/O configuration is handled at infrastructure level (e.g., logging/capture)
	// rather than through domain Spec, following hexagonal architecture.

	// Set process group for signal forwarding
	e.process.SetProcessGroup(cmd)

	// Return configured command
	return cmd, nil
}

// configureCredentials sets up user/group credentials on the command.
//
// Params:
//   - cmd: the exec.Cmd to configure
//   - user: username or UID to run as (empty string to skip)
//   - group: group name or GID to run as (empty string to skip)
//
// Returns:
//   - error: any error encountered during credentials configuration
func (e *Executor) configureCredentials(cmd *exec.Cmd, user, group string) error {
	// Check if credentials are specified
	if user == "" && group == "" {
		// Skip configuration if no credentials specified
		return nil
	}

	// Resolve user and group to UID and GID
	uid, gid, err := e.credentials.ResolveCredentials(user, group)
	// Check if resolution failed
	if err != nil {
		// Return error if credentials could not be resolved
		return fmt.Errorf("resolving credentials: %w", err)
	}

	// Apply resolved credentials to command
	if err := e.credentials.ApplyCredentials(cmd, uid, gid); err != nil {
		// Return error if credentials could not be applied
		return fmt.Errorf("applying credentials: %w", err)
	}

	// Return nil on success
	return nil
}
