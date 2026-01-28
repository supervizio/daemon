//go:build unix

// Package control provides platform-specific implementations of process control interfaces.
// This file implements process control functionality for Unix systems.
package control

import (
	"os/exec"
	"syscall"

	"github.com/kodflow/daemon/internal/infrastructure/process"
)

// Control implements ProcessControl for Unix systems.
// It provides process group management capabilities using Unix syscalls.
type Control struct{}

// NewControl returns a new Control for managing process groups.
//
// Returns:
//   - *Control: initialized process control for Unix systems
func NewControl() *Control { return &Control{} }

// New returns a new Control for managing process groups.
//
// Returns:
//   - *Control: initialized process control for Unix systems
func New() *Control { return &Control{} }

// SetProcessGroup configures a command to run in its own process group,
// enabling signal forwarding to all child processes.
//
// Params:
//   - cmd: the command to configure
func (m *Control) SetProcessGroup(cmd *exec.Cmd) {
	// Initialize SysProcAttr if not set to avoid nil pointer dereference.
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// GetProcessGroup retrieves the process group ID via syscall.
//
// Params:
//   - pid: the process ID to query
//
// Returns:
//   - int: the process group ID
//   - error: wrapped syscall error if the query fails
func (m *Control) GetProcessGroup(pid int) (int, error) {
	pgid, err := syscall.Getpgid(pid)
	// Check for syscall failure to provide meaningful error context.
	if err != nil {
		// Early exit: wrap syscall error with operation name for debugging.
		return 0, process.WrapError("getpgid", err)
	}
	// Success: return the process group ID to the caller.
	return pgid, nil
}
