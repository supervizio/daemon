// Package probe provides infrastructure adapters for service probing.
package probe

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kodflow/daemon/internal/domain/probe"
	"github.com/kodflow/daemon/internal/infrastructure/process"
)

// proberTypeExec is the type identifier for exec probers.
const proberTypeExec string = "exec"

// ExecProber performs command execution probes.
// It verifies service health by executing commands and checking exit codes.
type ExecProber struct {
	// timeout is the maximum duration for command execution.
	timeout time.Duration
}

// NewExecProber creates a new exec prober.
//
// Params:
//   - timeout: the maximum duration for command execution.
//
// Returns:
//   - *ExecProber: a configured exec prober ready to perform probes.
func NewExecProber(timeout time.Duration) *ExecProber {
	// Return configured exec prober.
	return &ExecProber{
		timeout: timeout,
	}
}

// Type returns the prober type.
//
// Returns:
//   - string: the constant "exec" identifying the prober type.
func (p *ExecProber) Type() string {
	// Return the exec prober type identifier.
	return proberTypeExec
}

// Probe performs a command execution probe.
// It executes the configured command and checks the exit code.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - target: the target containing command and args.
//
// Returns:
//   - probe.Result: the probe result with output and exit status.
func (p *ExecProber) Probe(ctx context.Context, target probe.Target) probe.Result {
	start := time.Now()

	// Validate command is not empty.
	if target.Command == "" {
		// Return failure for missing command configuration.
		return probe.NewFailureResult(
			time.Since(start),
			"empty command",
			probe.ErrEmptyCommand,
		)
	}

	// Extract command and arguments from target.
	command := target.Command
	args := target.Args

	// Check if args need to be parsed from command string.
	if len(args) == 0 {
		// Split command line into parts to separate command from arguments.
		parts := strings.Fields(command)
		// Check if splitting resulted in empty parts.
		if len(parts) == 0 {
			// Return failure for whitespace-only command.
			return probe.NewFailureResult(
				time.Since(start),
				"empty command",
				probe.ErrEmptyCommand,
			)
		}
		// First part is the command, rest are args.
		command = parts[0]
		// Check if there are additional arguments after the command.
		if len(parts) > 1 {
			// Extract remaining parts as arguments.
			args = parts[1:]
		}
	}

	// Execute the command and return result.
	return p.executeCommand(ctx, command, args, start)
}

// executeCommand runs the command and returns the probe result.
//
// Params:
//   - ctx: context for cancellation and timeout control.
//   - command: the command to execute.
//   - args: the command arguments.
//   - start: the start time for latency measurement.
//
// Returns:
//   - probe.Result: the probe result with output and exit status.
func (p *ExecProber) executeCommand(ctx context.Context, command string, args []string, start time.Time) probe.Result {
	// Create context with timeout only if timeout is positive.
	// Zero or negative timeout would create an already-canceled context.
	execCtx := ctx
	cancel := func() {}
	// Check if timeout is configured before creating timeout context.
	if p.timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, p.timeout)
	}
	defer cancel()

	// Create and execute command using TrustedCommand for security.
	cmd := process.TrustedCommand(execCtx, command, args...)
	output, err := cmd.CombinedOutput()
	latency := time.Since(start)

	// Handle execution errors from command.
	if err != nil {
		// Return failure result with error details and captured output.
		return probe.NewFailureResult(
			latency,
			fmt.Sprintf("command failed: %v (output: %s)", err, string(output)),
			err,
		)
	}

	// Return success result with trimmed command output.
	return probe.NewSuccessResult(
		latency,
		strings.TrimSpace(string(output)),
	)
}
