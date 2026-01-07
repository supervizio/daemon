package health

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/kodflow/daemon/internal/config"
)

// CommandChecker performs command-based health checks.
type CommandChecker struct {
	name    string
	command string
	timeout time.Duration
}

// NewCommandChecker creates a new command health checker.
func NewCommandChecker(cfg *config.HealthCheckConfig) *CommandChecker {
	name := cfg.Name
	if name == "" {
		name = fmt.Sprintf("cmd-%s", strings.Fields(cfg.Command)[0])
	}

	return &CommandChecker{
		name:    name,
		command: cfg.Command,
		timeout: cfg.Timeout.Duration(),
	}
}

// Name returns the checker name.
func (c *CommandChecker) Name() string {
	return c.name
}

// Type returns the checker type.
func (c *CommandChecker) Type() string {
	return "command"
}

// Check performs a command health check.
func (c *CommandChecker) Check(ctx context.Context) Result {
	start := time.Now()

	// Parse command
	parts := strings.Fields(c.command)
	if len(parts) == 0 {
		return Result{
			Status:    StatusUnhealthy,
			Message:   "empty command",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
			Error:     fmt.Errorf("empty command"),
		}
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return Result{
			Status:    StatusUnhealthy,
			Message:   fmt.Sprintf("command failed: %v (output: %s)", err, string(output)),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
			Error:     err,
		}
	}

	return Result{
		Status:    StatusHealthy,
		Message:   strings.TrimSpace(string(output)),
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}
}
