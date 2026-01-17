//go:build darwin

package signals

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
