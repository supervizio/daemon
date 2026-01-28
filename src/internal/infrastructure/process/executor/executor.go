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

// NewExecutor returns an Executor with production dependencies.
//
// Returns:
//   - *Executor: initialized executor with default credential and process managers
func NewExecutor() *Executor {
	return &Executor{credentials: credentials.New(), process: control.New(), findProcess: defaultFindProcess}
}

// New returns an Executor with production dependencies.
//
// Returns:
//   - *Executor: initialized executor with default credential and process managers
func New() *Executor {
	return &Executor{credentials: credentials.New(), process: control.New(), findProcess: defaultFindProcess}
}

// NewWithDeps returns an Executor with Wire-injected dependencies.
//
// Params:
//   - creds: credential manager for user/group resolution
//   - proc: process control for group management
//
// Returns:
//   - *Executor: initialized executor with provided dependencies
func NewWithDeps(creds credentials.CredentialManager, proc control.ProcessControl) *Executor {
	return &Executor{credentials: creds, process: proc, findProcess: defaultFindProcess}
}

// NewWithOptions returns an Executor with custom dependencies for testing.
//
// Params:
//   - creds: credential manager for user/group resolution
//   - proc: process control for group management
//   - finder: custom process finder for mock injection
//
// Returns:
//   - *Executor: initialized executor with all custom dependencies
func NewWithOptions(creds credentials.CredentialManager, proc control.ProcessControl, finder ProcessFinder) *Executor {
	return &Executor{credentials: creds, process: proc, findProcess: finder}
}

// Start spawns a process and returns a channel for exit notification.
// The background goroutine terminates when the process exits.
//
// Params:
//   - ctx: context for process cancellation
//   - spec: process specification including command, args, env, and credentials
//
// Returns:
//   - pid: process ID of the started process
//   - wait: channel that receives exit result when process terminates
//   - err: error if command build, credential setup, or start fails
func (e *Executor) Start(ctx context.Context, spec domain.Spec) (pid int, wait <-chan domain.ExitResult, err error) {
	cmd, err := e.buildCommand(ctx, spec)
	// Command parsing or environment setup failed.
	if err != nil {
		return 0, nil, err
	}
	// Credential resolution or application failed.
	if err := e.configureCredentials(cmd, spec.User, spec.Group); err != nil {
		return 0, nil, err
	}
	// Fork/exec failed.
	if err := cmd.Start(); err != nil {
		return 0, nil, fmt.Errorf("starting process: %w", err)
	}
	// Buffer of 1 prevents goroutine leak if receiver abandons channel.
	waitCh := make(chan domain.ExitResult, 1)
	go e.waitForProcess(cmd, waitCh)
	return cmd.Process.Pid, waitCh, nil
}

// waitForProcess collects the exit result and signals completion via channel.
//
// Params:
//   - cmd: waiter interface (typically *exec.Cmd) to wait on
//   - wait: channel to send exit result when process terminates
func (e *Executor) waitForProcess(cmd Waiter, wait chan<- domain.ExitResult) {
	err := cmd.Wait()
	result := domain.ExitResult{}
	// Process exited with error or non-zero status.
	if err != nil {
		var exitErr *exec.ExitError
		// Normal exit with non-zero code.
		if errors.As(err, &exitErr) {
			result.Code = exitErr.ExitCode()
		} else {
			// Abnormal termination (signal, resource limit, etc).
			result.Code = -1
			result.Error = err
		}
	}
	wait <- result
	close(wait)
}

// Stop sends SIGTERM and waits for graceful exit, then SIGKILL on timeout.
//
// Params:
//   - pid: process ID to stop
//   - timeout: maximum time to wait for graceful shutdown before SIGKILL
//
// Returns:
//   - error: if process cannot be found or signal delivery fails
//
// Goroutine lifecycle: Spawns one goroutine to wait for process exit.
// Termination: Goroutine exits when proc.Wait() returns (process exits or is killed).
// Cleanup: Done channel is buffered (size 1) to prevent goroutine leak if caller abandons.
func (e *Executor) Stop(pid int, timeout time.Duration) error {
	proc, err := e.findProcess(pid)
	// Process handle acquisition failed.
	if err != nil {
		return fmt.Errorf("finding process: %w", err)
	}
	// Request graceful shutdown.
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("sending SIGTERM: %w", err)
	}
	// Wait for process exit in background.
	// Goroutine lifecycle: Waits for process termination, sends result to done channel.
	// Termination guarantee: proc.Wait() always returns when process exits (naturally or killed).
	done := make(chan error, 1)
	go func() {
		_, err := proc.Wait()
		done <- err
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	// Process exited gracefully within timeout.
	case <-done:
		return nil
	// Timeout expired; force kill.
	case <-timer.C:
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("killing process: %w", err)
		}
		// Wait for kill to complete.
		<-done
		return nil
	}
}

// Signal delivers a signal to the specified process.
//
// Params:
//   - pid: process ID to signal
//   - sig: signal to deliver (e.g., syscall.SIGHUP for reload)
//
// Returns:
//   - error: if process cannot be found or signal delivery fails
func (e *Executor) Signal(pid int, sig os.Signal) error {
	proc, err := e.findProcess(pid)
	// Process handle acquisition failed.
	if err != nil {
		return fmt.Errorf("finding process: %w", err)
	}
	return proc.Signal(sig)
}

// buildCommand constructs an exec.Cmd with environment and process group setup.
//
// Params:
//   - ctx: context for cancellation support
//   - spec: process specification with command, args, dir, and env
//
// Returns:
//   - *exec.Cmd: configured command with environment and process group
//   - error: ErrEmptyCommand if command string is empty
func (e *Executor) buildCommand(ctx context.Context, spec domain.Spec) (*exec.Cmd, error) {
	parts := strings.Fields(spec.Command)
	// Empty command string after whitespace split.
	if len(parts) == 0 {
		return nil, shared.ErrEmptyCommand
	}
	// Combine inline args from command string with explicit args.
	args := make([]string, 0, len(parts)-1+len(spec.Args))
	args = append(args, parts[1:]...)
	args = append(args, spec.Args...)
	cmd := TrustedCommand(ctx, parts[0], args...)
	// Set working directory if specified.
	if spec.Dir != "" {
		cmd.Dir = spec.Dir
	}
	// Inherit current environment and merge spec-provided vars.
	baseEnv := os.Environ()
	cmd.Env = make([]string, 0, len(baseEnv)+len(spec.Env))
	cmd.Env = append(cmd.Env, baseEnv...)
	for k, v := range spec.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	// Enable process group for clean signal delivery.
	e.process.SetProcessGroup(cmd)
	return cmd, nil
}

// configureCredentials applies user/group credentials for privilege drop.
//
// Params:
//   - cmd: exec.Cmd to configure with credentials
//   - user: username or UID (empty to skip)
//   - group: group name or GID (empty to inherit from user)
//
// Returns:
//   - error: if credential resolution or application fails
func (e *Executor) configureCredentials(cmd *exec.Cmd, user, group string) error {
	// Skip credential setup when running as invoking user.
	if user == "" && group == "" {
		return nil
	}
	// Resolve names to numeric IDs.
	uid, gid, err := e.credentials.ResolveCredentials(user, group)
	if err != nil {
		return fmt.Errorf("resolving credentials: %w", err)
	}
	// Apply credentials to SysProcAttr for privilege drop.
	if err := e.credentials.ApplyCredentials(cmd, uid, gid); err != nil {
		return fmt.Errorf("applying credentials: %w", err)
	}
	return nil
}
