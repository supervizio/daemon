//go:build unix

// Package adapters provides platform-specific implementations of kernel interfaces.
// This file implements process control functionality for Unix systems.
package adapters

import (
	"os/exec"
	"syscall"

	"github.com/kodflow/daemon/internal/infrastructure/kernel/ports"
)

// UnixProcessControl implements ProcessControl for Unix systems.
// It provides process group management capabilities using Unix syscalls.
type UnixProcessControl struct{}

// NewUnixProcessControl creates a new ProcessControl.
//
// Returns:
//   - *UnixProcessControl: a new process control instance
func NewUnixProcessControl() *UnixProcessControl {
	// Return a new instance of UnixProcessControl.
	return &UnixProcessControl{}
}

// SetProcessGroup configures a command to run in its own process group.
//
// Params:
//   - cmd: the command to configure
func (m *UnixProcessControl) SetProcessGroup(cmd *exec.Cmd) {
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
func (m *UnixProcessControl) GetProcessGroup(pid int) (int, error) {
	pgid, err := syscall.Getpgid(pid)
	// Check if getpgid syscall failed.
	if err != nil {
		// Return zero and the wrapped error.
		return 0, ports.WrapError("getpgid", err)
	}
	// Return the process group ID.
	return pgid, nil
}
