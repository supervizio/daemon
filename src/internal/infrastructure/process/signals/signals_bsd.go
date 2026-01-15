//go:build freebsd || openbsd || netbsd || dragonfly

package signals

import "github.com/kodflow/daemon/internal/infrastructure/kernel/ports"

// SetSubreaper is a no-op on BSD systems.
// BSD does not have the PR_SET_CHILD_SUBREAPER functionality.
// On BSD, daemon must run as PID 1 to reap zombies.
func (m *Manager) SetSubreaper() error {
	return ports.ErrNotSupported
}

// ClearSubreaper is a no-op on BSD systems.
func (m *Manager) ClearSubreaper() error {
	return nil
}

// IsSubreaper always returns false on BSD systems.
func (m *Manager) IsSubreaper() (bool, error) {
	return false, nil
}
