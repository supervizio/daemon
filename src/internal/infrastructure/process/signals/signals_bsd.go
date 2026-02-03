//go:build freebsd || openbsd || netbsd

package signals

// platformInit is a no-op on BSD systems.
// BSD uses the same signals as the base Unix implementation.
//
// Params:
//   - m: manager (unused on BSD)
func platformInit(m *Manager) {
	// no platform-specific signals on BSD.
}

// SetSubreaper is a no-op on BSD systems.
// BSD does not have the PR_SET_CHILD_SUBREAPER functionality.
// On BSD, daemon must run as PID 1 to reap zombies.
func (m *Manager) SetSubreaper() error {
	return ErrSignalNotSupported
}

// ClearSubreaper is a no-op on BSD systems.
func (m *Manager) ClearSubreaper() error {
	return nil
}

// IsSubreaper always returns false on BSD systems.
func (m *Manager) IsSubreaper() (bool, error) {
	return false, nil
}
