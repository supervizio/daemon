//go:build linux

// Package adapters provides platform-specific implementations.
package adapters

import (
	"os"
	"syscall"
	"unsafe" // #nosec G103 - required for low-level prctl syscall
)

// Linux prctl constants for child subreaper functionality.
// These are not exported by the syscall package, so we define them here.
// See: man 2 prctl, PR_SET_CHILD_SUBREAPER (since Linux 3.4)
const (
	// prSetChildSubreaper is the prctl option to set the child subreaper flag.
	prSetChildSubreaper uintptr = 36
	// prGetChildSubreaper is the prctl option to get the child subreaper flag.
	prGetChildSubreaper uintptr = 37
)

// initLinuxSignals adds Linux-specific signals to the signal manager.
// Called from NewUnixSignalManager when on Linux platform.
//
//nolint:gochecknoinits // Platform-specific signal registration requires init
func init() {
	// Add Linux-specific signals.
	sm := NewUnixSignalManager()
	sm.AddSignal("SIGPWR", syscall.SIGPWR)
	sm.AddSignal("SIGSTKFLT", syscall.SIGSTKFLT)
}

// SetSubreaper sets the current process as a child subreaper.
// This allows orphaned child processes to be reparented to this process
// instead of init (PID 1). Available on Linux >= 3.4.
//
// Returns:
//   - error: an error if the prctl syscall fails
func (m *UnixSignalManager) SetSubreaper() error {
	// Return the result of enabling subreaper mode.
	return prctlSubreaper(1)
}

// ClearSubreaper clears the child subreaper flag.
//
// Returns:
//   - error: an error if the prctl syscall fails
func (m *UnixSignalManager) ClearSubreaper() error {
	// Return the result of disabling subreaper mode.
	return prctlSubreaper(0)
}

// IsSubreaper returns true if the current process is a child subreaper.
//
// Returns:
//   - bool: true if the current process is a child subreaper
//   - error: an error if the prctl syscall fails
func (m *UnixSignalManager) IsSubreaper() (bool, error) {
	var flag int
	// #nosec G103 - unsafe.Pointer required for prctl syscall interface
	_, _, errno := syscall.RawSyscall(syscall.SYS_PRCTL, prGetChildSubreaper, uintptr(unsafe.Pointer(&flag)), 0)
	// Check if the syscall returned an error.
	if errno != 0 {
		// Return false and wrap the syscall error.
		return false, os.NewSyscallError("prctl", errno)
	}
	// Return whether the subreaper flag is set.
	return flag != 0, nil
}

// prctlSubreaper sets or clears the child subreaper flag using prctl.
//
// Params:
//   - flag: 1 to enable subreaper, 0 to disable
//
// Returns:
//   - error: an error if the prctl syscall fails
func prctlSubreaper(flag int) error {
	_, _, errno := syscall.RawSyscall(syscall.SYS_PRCTL, prSetChildSubreaper, uintptr(flag), 0)
	// Check if the syscall returned an error.
	if errno != 0 {
		// Return the wrapped syscall error.
		return os.NewSyscallError("prctl", errno)
	}
	// Return nil indicating success.
	return nil
}
