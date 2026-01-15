//go:build unix

// Package process provides platform-specific implementations of kernel interfaces.
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

// New creates a new process Control instance.
//
// Returns:
//   - *Control: a new process control instance
func New() *Control {
	// Return a new instance of Control.
	return &Control{}
}

// SetProcessGroup configures a command to run in its own process group.
//
// Params:
//   - cmd: the command to configure
func (m *Control) SetProcessGroup(cmd *exec.Cmd) {
	// Check if SysProcAttr is nil and initialize it.
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// GetProcessGroup returns the process group ID for a process.
//
// Params:
//   - pid: the process ID to get the group for
//
// Returns:
//   - int: the process group ID
//   - error: an error if the process group could not be retrieved
func (m *Control) GetProcessGroup(pid int) (int, error) {
	pgid, err := syscall.Getpgid(pid)
	// Check if getpgid syscall failed.
	if err != nil {
		// Return zero and the wrapped error.
		return 0, process.WrapError("getpgid", err)
	}
	// Return the process group ID.
	return pgid, nil
}
