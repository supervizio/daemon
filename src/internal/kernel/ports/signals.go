// Package ports defines the interfaces for OS abstraction.
package ports

import (
	"os"
	"syscall"
)

// SignalManager handles signal operations across platforms.
type SignalManager interface {
	// Notify registers for signal notifications and returns a channel.
	Notify(signals ...os.Signal) <-chan os.Signal

	// Stop stops signal notifications on the channel.
	Stop(ch chan<- os.Signal)

	// Forward sends a signal to a process.
	Forward(pid int, sig os.Signal) error

	// ForwardToGroup sends a signal to a process group.
	ForwardToGroup(pgid int, sig syscall.Signal) error

	// IsTermSignal returns true if the signal is a termination signal.
	IsTermSignal(sig os.Signal) bool

	// IsReloadSignal returns true if the signal is a reload signal.
	IsReloadSignal(sig os.Signal) bool

	// SignalByName returns a signal by its name (e.g., "SIGTERM").
	SignalByName(name string) (os.Signal, bool)

	// SetSubreaper sets the current process as a child subreaper (Linux only).
	SetSubreaper() error

	// ClearSubreaper clears the child subreaper flag.
	ClearSubreaper() error

	// IsSubreaper returns true if the current process is a subreaper.
	IsSubreaper() (bool, error)
}
