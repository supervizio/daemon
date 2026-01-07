//go:build unix

// Package adapters provides OS-specific implementations of kernel interfaces.
package adapters

import (
	"os"
	"os/signal"
	"syscall"
)

// UnixSignalManager implements SignalManager for Unix systems.
type UnixSignalManager struct {
	signalMap map[string]os.Signal
}

// NewSignalManager creates a new SignalManager for the current platform.
func NewSignalManager() *UnixSignalManager {
	sm := &UnixSignalManager{
		signalMap: map[string]os.Signal{
			"SIGHUP":  syscall.SIGHUP,
			"SIGINT":  syscall.SIGINT,
			"SIGQUIT": syscall.SIGQUIT,
			"SIGTERM": syscall.SIGTERM,
			"SIGKILL": syscall.SIGKILL,
			"SIGUSR1": syscall.SIGUSR1,
			"SIGUSR2": syscall.SIGUSR2,
			"SIGCHLD": syscall.SIGCHLD,
		},
	}
	// Platform-specific signals are added via init() in *_linux.go, etc.
	return sm
}

// Notify registers for signal notifications.
func (m *UnixSignalManager) Notify(signals ...os.Signal) <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	return ch
}

// Stop stops signal notifications on the channel.
func (m *UnixSignalManager) Stop(ch chan<- os.Signal) {
	signal.Stop(ch)
}

// Forward sends a signal to a process.
func (m *UnixSignalManager) Forward(pid int, sig os.Signal) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(sig)
}

// ForwardToGroup sends a signal to a process group.
func (m *UnixSignalManager) ForwardToGroup(pgid int, sig syscall.Signal) error {
	// Negative PID sends to process group
	return syscall.Kill(-pgid, sig)
}

// IsTermSignal returns true if the signal is a termination signal.
func (m *UnixSignalManager) IsTermSignal(sig os.Signal) bool {
	switch sig {
	case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL:
		return true
	default:
		return false
	}
}

// IsReloadSignal returns true if the signal is a reload signal.
func (m *UnixSignalManager) IsReloadSignal(sig os.Signal) bool {
	return sig == syscall.SIGHUP
}

// SignalByName returns a signal by its name.
func (m *UnixSignalManager) SignalByName(name string) (os.Signal, bool) {
	sig, ok := m.signalMap[name]
	return sig, ok
}

// AddSignal adds a platform-specific signal to the map.
func (m *UnixSignalManager) AddSignal(name string, sig os.Signal) {
	m.signalMap[name] = sig
}
