//go:build linux

package process

import (
	"os"
	"syscall"
	"unsafe"
)

// Additional Linux-specific signals.
func init() {
	SignalMap["SIGPWR"] = syscall.SIGPWR
	SignalMap["SIGSTKFLT"] = syscall.SIGSTKFLT
}

// SetChildSubreaper sets the current process as a child subreaper.
// This allows orphaned child processes to be reparented to this process
// instead of init (PID 1). Available on Linux >= 3.4.
func SetChildSubreaper() error {
	return prctlSubreaper(1)
}

// ClearChildSubreaper clears the child subreaper flag.
func ClearChildSubreaper() error {
	return prctlSubreaper(0)
}

func prctlSubreaper(flag int) error {
	_, _, errno := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_SET_CHILD_SUBREAPER, uintptr(flag), 0)
	if errno != 0 {
		return os.NewSyscallError("prctl", errno)
	}
	return nil
}

// IsChildSubreaper returns true if the current process is a child subreaper.
func IsChildSubreaper() (bool, error) {
	var flag int
	_, _, errno := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_GET_CHILD_SUBREAPER, uintptr(unsafe.Pointer(&flag)), 0)
	if errno != 0 {
		return false, os.NewSyscallError("prctl", errno)
	}
	return flag != 0, nil
}
