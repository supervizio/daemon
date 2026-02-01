//go:build linux

// Package signals provides platform-specific implementations.
package signals

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

// prctlFunc is the function type for prctl syscalls.
// Extracted as a variable to enable error injection in tests.
type prctlFunc func(trap uintptr, a1 uintptr, a2 uintptr, a3 uintptr) (uintptr, uintptr, syscall.Errno)

// prctlSyscall is the syscall function, can be overridden in tests.
var prctlSyscall prctlFunc = syscall.RawSyscall

// platformInit adds Linux-specific signals to the manager during construction.
//
// Params:
//   - m: manager to register signals to
func platformInit(m *Manager) {
	// register power failure signal (SIGPWR).
	m.AddSignal("SIGPWR", syscall.SIGPWR)
	// register stack fault signal (SIGSTKFLT).
	m.AddSignal("SIGSTKFLT", syscall.SIGSTKFLT)
}

// SetSubreaper enables orphan adoption via prctl (Linux >= 3.4).
//
// Returns:
//   - error: syscall error if prctl fails
func (m *Manager) SetSubreaper() error {
	// set subreaper flag to 1 (enabled).
	return prctlSubreaper(1)
}

// ClearSubreaper disables orphan adoption.
//
// Returns:
//   - error: syscall error if prctl fails
func (m *Manager) ClearSubreaper() error {
	// set subreaper flag to 0 (disabled).
	return prctlSubreaper(0)
}

// IsSubreaper queries the subreaper flag via prctl.
//
// Returns:
//   - bool: true if process is currently a subreaper for orphaned children
//   - error: syscall error if prctl fails
func (m *Manager) IsSubreaper() (bool, error) {
	var flag int
	// #nosec G103 - unsafe.Pointer required for prctl syscall interface
	// query current subreaper status from kernel.
	_, _, errno := prctlSyscall(syscall.SYS_PRCTL, prGetChildSubreaper, uintptr(unsafe.Pointer(&flag)), 0)
	// prctl failed (unlikely on modern kernels).
	if errno != 0 {
		// return error with syscall context.
		return false, os.NewSyscallError("prctl", errno)
	}
	// Flag is non-zero when subreaper is enabled.
	// convert flag to boolean result.
	return flag != 0, nil
}

// prctlSubreaper modifies the subreaper flag via prctl syscall.
//
// Params:
//   - flag: 1 to enable subreaper, 0 to disable
//
// Returns:
//   - error: syscall error if prctl fails
func prctlSubreaper(flag int) error {
	// execute prctl syscall to set subreaper flag.
	_, _, errno := prctlSyscall(syscall.SYS_PRCTL, prSetChildSubreaper, uintptr(flag), 0)
	// prctl failed (unlikely on modern kernels).
	if errno != 0 {
		// return error with syscall context.
		return os.NewSyscallError("prctl", errno)
	}
	// subreaper flag successfully modified.
	return nil
}
