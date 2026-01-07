//go:build freebsd || openbsd || netbsd || dragonfly

package process

// SetChildSubreaper is a no-op on BSD systems.
// BSD does not have the PR_SET_CHILD_SUBREAPER functionality.
// On BSD, daemon must run as PID 1 to reap zombies.
func SetChildSubreaper() error {
	return nil
}

// ClearChildSubreaper is a no-op on BSD systems.
func ClearChildSubreaper() error {
	return nil
}

// IsChildSubreaper always returns false on BSD systems.
func IsChildSubreaper() (bool, error) {
	return false, nil
}
