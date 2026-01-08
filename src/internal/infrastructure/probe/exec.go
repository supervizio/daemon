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
		// Return failure for empty command.
		return probe.NewFailureResult(
			time.Since(start),
			"empty command",
			probe.ErrEmptyCommand,
		)
	}

	// Parse command if Args is empty (command may contain full command line).
	command := target.Command
	args := target.Args
	if len(args) == 0 {
		// Split command line into parts.
		parts := strings.Fields(command)
		if len(parts) == 0 {
			// Return failure for empty command.
			return probe.NewFailureResult(
				time.Since(start),
				"empty command",
				probe.ErrEmptyCommand,
			)
		}
		// First part is the command, rest are args.
		command = parts[0]
		if len(parts) > 1 {
			args = parts[1:]
		}
	}

	// Create context with timeout.
	execCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Create and execute command using TrustedCommand.
	cmd := process.TrustedCommand(execCtx, command, args...)
	output, err := cmd.CombinedOutput()
	latency := time.Since(start)

	// Handle execution errors.
	if err != nil {
		// Return failure result with error details.
		return probe.NewFailureResult(
			latency,
			fmt.Sprintf("command failed: %v (output: %s)", err, string(output)),
			err,
		)
	}

	// Return success result with command output.
	return probe.NewSuccessResult(
		latency,
		strings.TrimSpace(string(output)),
	)
}
