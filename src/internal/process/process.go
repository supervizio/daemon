// Package process provides process management for daemon.
package process

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/kodflow/daemon/internal/config"
	"github.com/kodflow/daemon/internal/kernel"
)

// State represents the current state of a process.
type State int

const (
	StateStopped State = iota
	StateStarting
	StateRunning
	StateStopping
	StateFailed
)

func (s State) String() string {
	switch s {
	case StateStopped:
		return "stopped"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// Process represents a managed process.
type Process struct {
	Config *config.ServiceConfig

	mu        sync.RWMutex
	cmd       *exec.Cmd
	state     State
	pid       int
	exitCode  int
	startTime time.Time
	stopTime  time.Time
	restarts  int

	stdout io.Writer
	stderr io.Writer

	done chan struct{}
}

// New creates a new Process from service configuration.
func New(cfg *config.ServiceConfig) *Process {
	return &Process{
		Config: cfg,
		state:  StateStopped,
		done:   make(chan struct{}),
	}
}

// SetOutput sets the stdout and stderr writers for the process.
func (p *Process) SetOutput(stdout, stderr io.Writer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stdout = stdout
	p.stderr = stderr
}

// State returns the current state of the process.
func (p *Process) State() State {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// PID returns the process ID, or 0 if not running.
func (p *Process) PID() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pid
}

// ExitCode returns the last exit code.
func (p *Process) ExitCode() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.exitCode
}

// Restarts returns the number of times this process has been restarted.
func (p *Process) Restarts() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.restarts
}

// Uptime returns how long the process has been running.
func (p *Process) Uptime() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.state != StateRunning {
		return 0
	}
	return time.Since(p.startTime)
}

// Start starts the process.
func (p *Process) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == StateRunning || p.state == StateStarting {
		return fmt.Errorf("process already running")
	}

	p.state = StateStarting
	p.done = make(chan struct{})

	// Parse command and args
	cmdParts := parseCommand(p.Config.Command)
	if len(cmdParts) == 0 {
		p.state = StateFailed
		return fmt.Errorf("empty command")
	}

	args := append(cmdParts[1:], p.Config.Args...)
	cmd := exec.CommandContext(ctx, cmdParts[0], args...)

	// Set working directory
	if p.Config.WorkingDirectory != "" {
		cmd.Dir = p.Config.WorkingDirectory
	}

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range p.Config.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set output
	if p.stdout != nil {
		cmd.Stdout = p.stdout
	}
	if p.stderr != nil {
		cmd.Stderr = p.stderr
	}

	// Set process group for signal forwarding
	kernel.Default.Process.SetProcessGroup(cmd)

	// Apply credentials if configured
	if p.Config.User != "" || p.Config.Group != "" {
		uid, gid, err := kernel.Default.Credentials.ResolveCredentials(p.Config.User, p.Config.Group)
		if err != nil {
			p.state = StateFailed
			return fmt.Errorf("resolving credentials: %w", err)
		}
		if err := kernel.Default.Credentials.ApplyCredentials(cmd, uid, gid); err != nil {
			p.state = StateFailed
			return fmt.Errorf("applying credentials: %w", err)
		}
	}

	if err := cmd.Start(); err != nil {
		p.state = StateFailed
		return fmt.Errorf("starting process: %w", err)
	}

	p.cmd = cmd
	p.pid = cmd.Process.Pid
	p.startTime = time.Now()
	p.state = StateRunning

	// Monitor process in background
	go p.monitor()

	return nil
}

// parseCommand splits a command string into parts.
func parseCommand(cmd string) []string {
	return strings.Fields(cmd)
}

// monitor waits for the process to exit and updates state.
func (p *Process) monitor() {
	err := p.cmd.Wait()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.stopTime = time.Now()
	p.pid = 0

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			p.exitCode = exitErr.ExitCode()
		} else {
			p.exitCode = -1
		}
		p.state = StateFailed
	} else {
		p.exitCode = 0
		p.state = StateStopped
	}

	close(p.done)
}

// Stop gracefully stops the process.
func (p *Process) Stop(timeout time.Duration) error {
	p.mu.Lock()
	if p.state != StateRunning {
		p.mu.Unlock()
		return nil
	}

	p.state = StateStopping
	cmd := p.cmd
	done := p.done
	p.mu.Unlock()

	// Send SIGTERM
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("sending SIGTERM: %w", err)
	}

	// Wait for graceful shutdown or timeout
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		// Force kill
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("killing process: %w", err)
		}
		<-done
		return nil
	}
}

// Signal sends a signal to the process.
func (p *Process) Signal(sig os.Signal) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("process not running")
	}

	return p.cmd.Process.Signal(sig)
}

// Wait returns a channel that's closed when the process exits.
func (p *Process) Wait() <-chan struct{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.done
}

// Reload sends SIGHUP to the process.
func (p *Process) Reload() error {
	return p.Signal(syscall.SIGHUP)
}

// IncrementRestarts increments the restart counter.
func (p *Process) IncrementRestarts() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.restarts++
}

// ResetRestarts resets the restart counter.
func (p *Process) ResetRestarts() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.restarts = 0
}
