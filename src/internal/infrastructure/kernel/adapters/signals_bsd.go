//go:build freebsd || openbsd || netbsd || dragonfly

package adapters

import "github.com/kodflow/daemon/internal/infrastructure/kernel/ports"

// SetSubreaper is a no-op on BSD systems.
// BSD does not have the PR_SET_CHILD_SUBREAPER functionality.
// On BSD, daemon must run as PID 1 to reap zombies.
func (m *UnixSignalManager) SetSubreaper() error {
	return ports.ErrNotSupported
}

// ClearSubreaper is a no-op on BSD systems.
func (m *UnixSignalManager) ClearSubreaper() error {
	return nil
}

// IsSubreaper always returns false on BSD systems.
func (m *UnixSignalManager) IsSubreaper() (bool, error) {
	return false, nil
}
