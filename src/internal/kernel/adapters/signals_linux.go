//go:build linux

package adapters

import (
	"os"
	"syscall"
	"unsafe"
)

// Linux prctl constants for child subreaper functionality.
// These are not exported by the syscall package, so we define them here.
// See: man 2 prctl, PR_SET_CHILD_SUBREAPER (since Linux 3.4)
const (
	prSetChildSubreaper = 36 // PR_SET_CHILD_SUBREAPER
	prGetChildSubreaper = 37 // PR_GET_CHILD_SUBREAPER
)

func init() {
	// Add Linux-specific signals
	sm := NewSignalManager()
	sm.AddSignal("SIGPWR", syscall.SIGPWR)
	sm.AddSignal("SIGSTKFLT", syscall.SIGSTKFLT)
}

// SetSubreaper sets the current process as a child subreaper.
// This allows orphaned child processes to be reparented to this process
// instead of init (PID 1). Available on Linux >= 3.4.
func (m *UnixSignalManager) SetSubreaper() error {
	return prctlSubreaper(1)
}

// ClearSubreaper clears the child subreaper flag.
func (m *UnixSignalManager) ClearSubreaper() error {
	return prctlSubreaper(0)
}

// IsSubreaper returns true if the current process is a child subreaper.
func (m *UnixSignalManager) IsSubreaper() (bool, error) {
	var flag int
	_, _, errno := syscall.RawSyscall(syscall.SYS_PRCTL, prGetChildSubreaper, uintptr(unsafe.Pointer(&flag)), 0)
	if errno != 0 {
		return false, os.NewSyscallError("prctl", errno)
	}
	return flag != 0, nil
}

func prctlSubreaper(flag int) error {
	_, _, errno := syscall.RawSyscall(syscall.SYS_PRCTL, prSetChildSubreaper, uintptr(flag), 0)
	if errno != 0 {
		return os.NewSyscallError("prctl", errno)
	}
	return nil
}
