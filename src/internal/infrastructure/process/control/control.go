// Package control provides process control interfaces.
package control

import "os/exec"

// ProcessControl handles process-level OS operations.
type ProcessControl interface {
	// SetProcessGroup configures a command to run in its own process group.
	SetProcessGroup(cmd *exec.Cmd)

	// GetProcessGroup returns the process group ID for a process.
	GetProcessGroup(pid int) (int, error)
}
