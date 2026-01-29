// Package executor provides infrastructure adapters for process execution.
package executor

import (
	"context"
	"os/exec"
)

// TrustedCommand wraps exec.CommandContext for admin-controlled config commands.
// SECURITY: Only use for commands from validated YAML configs, not user input.
// Commands are trusted as they come from administrator-controlled files.
//
// Params:
//   - ctx: context for cancellation support
//   - name: executable path or name to run
//   - args: command-line arguments to pass
//
// Returns:
//   - *exec.Cmd: configured command ready to start
//
// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
// nosemgrep: go_subproc_rule-subproc
func TrustedCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	// create command from trusted configuration source.
	return exec.CommandContext(ctx, name, args...)
}
