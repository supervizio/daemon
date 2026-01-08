// Package process provides infrastructure adapters for process execution.
package process

import (
	"context"
	"os/exec"
)

// TrustedCommand creates an exec.Cmd from trusted configuration sources.
//
// SECURITY: This function is intended for commands loaded from validated
// configuration files (YAML configs, health check definitions), NOT for
// user-provided input. The commands are trusted because they come from
// administrator-controlled configuration files loaded at startup.
//
// This is expected behavior for a process supervisor whose core function
// is to execute configured services and health check commands.
//
// The nosemgrep annotations acknowledge this is intentional design,
// not a security vulnerability.
//
// Params:
//   - ctx: context for cancellation support
//   - name: the command executable name or path
//   - args: command arguments
//
// Returns:
//   - *exec.Cmd: configured command ready for execution
//
// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
// nosemgrep: go_subproc_rule-subproc
func TrustedCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}
