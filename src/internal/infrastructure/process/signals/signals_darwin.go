//go:build darwin

package signals

// platformInit is a no-op on Darwin.
// Darwin uses the same signals as the base Unix implementation.
//
// Params:
//   - m: manager (unused on Darwin)
func platformInit(m *Manager) {
	// no platform-specific signals on Darwin.
}

// SetSubreaper is a no-op on Darwin.
// macOS does not support the PR_SET_CHILD_SUBREAPER functionality.
func (m *Manager) SetSubreaper() error {
	return ErrSignalNotSupported
}

// ClearSubreaper is a no-op on Darwin.
func (m *Manager) ClearSubreaper() error {
	return nil
}

// IsSubreaper always returns false on Darwin.
func (m *Manager) IsSubreaper() (bool, error) {
	return false, nil
}
