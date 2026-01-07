//go:build darwin

package adapters

import "github.com/kodflow/daemon/internal/kernel/ports"

// SetSubreaper is a no-op on Darwin.
// macOS does not support the PR_SET_CHILD_SUBREAPER functionality.
func (m *UnixSignalManager) SetSubreaper() error {
	return ports.ErrNotSupported
}

// ClearSubreaper is a no-op on Darwin.
func (m *UnixSignalManager) ClearSubreaper() error {
	return nil
}

// IsSubreaper always returns false on Darwin.
func (m *UnixSignalManager) IsSubreaper() (bool, error) {
	return false, nil
}
