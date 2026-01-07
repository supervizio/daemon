//go:build unix

package adapters

import (
	"os/exec"
	"syscall"

	"github.com/kodflow/daemon/internal/kernel/ports"
)

// UnixProcessControl implements ProcessControl for Unix systems.
type UnixProcessControl struct{}

// NewProcessControl creates a new ProcessControl.
func NewProcessControl() *UnixProcessControl {
	return &UnixProcessControl{}
}

// SetProcessGroup configures a command to run in its own process group.
func (m *UnixProcessControl) SetProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// GetProcessGroup returns the process group ID for a process.
func (m *UnixProcessControl) GetProcessGroup(pid int) (int, error) {
	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		return 0, ports.WrapError("getpgid", err)
	}
	return pgid, nil
}
