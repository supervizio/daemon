// Package health provides infrastructure adapters for health checking.
package health

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/infrastructure/process"
)

// checkerTypeCommand is the type identifier for command health checkers.
const checkerTypeCommand string = "command"

// emptyCommandName is the default name used when command is empty.
const emptyCommandName string = "cmd-empty"

// ErrEmptyCommand indicates the command string is empty.
var ErrEmptyCommand error = errors.New("empty command")

// CommandChecker performs command-based health checks.
// It executes shell commands to determine the health status of a service.
type CommandChecker struct {
	name    string
	command string
	timeout time.Duration
}

// NewCommandChecker creates a new command health checker.
// It initializes the checker with the provided configuration.
//
// Params:
//   - cfg: the health check configuration containing command and timeout settings
//
// Returns:
//   - *CommandChecker: the initialized command checker instance
func NewCommandChecker(cfg *service.HealthCheckConfig) *CommandChecker {
	name := cfg.Name
	// Check if a custom name was provided in the configuration.
	if name == "" {
		parts := strings.Fields(cfg.Command)
		// Use the command executable name as the checker name.
		if len(parts) > 0 {
			name = fmt.Sprintf("cmd-%s", parts[0])
		} else {
			// Fall back to default name for empty commands.
			name = emptyCommandName
		}
	}

	// Return the initialized command checker.
	return &CommandChecker{
		name:    name,
		command: cfg.Command,
		timeout: cfg.Timeout.Duration(),
	}
}

// Name returns the checker name.
// It provides the identifier for this health checker instance.
//
// Returns:
//   - string: the name of this health checker
func (c *CommandChecker) Name() string {
	// Return the stored checker name.
	return c.name
}

// Type returns the checker type.
// It identifies this checker as a command-based health checker.
//
// Returns:
//   - string: the type identifier "command"
func (c *CommandChecker) Type() string {
	// Return the command checker type constant.
	return checkerTypeCommand
}

// Check performs a command health check.
// It executes the configured command and returns the health result.
//
// Params:
//   - ctx: the context for cancellation and timeout control
//
// Returns:
//   - domain.Result: the health check result indicating healthy or unhealthy status
func (c *CommandChecker) Check(ctx context.Context) domain.Result {
	start := time.Now()

	parts := strings.Fields(c.command)
	// Validate that the command is not empty.
	if len(parts) == 0 {
		// Return unhealthy result for empty command.
		return domain.NewUnhealthyResult(
			"empty command",
			time.Since(start),
			ErrEmptyCommand,
		)
	}

	cmd := process.TrustedCommand(ctx, parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()

	// Check if the command execution failed.
	if err != nil {
		// Return unhealthy result with error details.
		return domain.NewUnhealthyResult(
			fmt.Sprintf("command failed: %v (output: %s)", err, string(output)),
			time.Since(start),
			err,
		)
	}

	// Return healthy result with command output.
	return domain.NewHealthyResult(
		strings.TrimSpace(string(output)),
		time.Since(start),
	)
}
