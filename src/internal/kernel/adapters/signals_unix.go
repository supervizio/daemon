//go:build unix

// Package adapters provides OS-specific implementations of kernel interfaces.
package adapters

import (
	"os"
	"os/signal"
	"syscall"
)

// UnixSignalManager implements SignalManager for Unix systems.
// It provides signal handling capabilities including notification, forwarding, and signal identification.
type UnixSignalManager struct {
	// signalMap maps signal names to their os.Signal values.
	signalMap map[string]os.Signal
}

// NewUnixSignalManager creates a new SignalManager for the current platform.
//
// Returns:
//   - *UnixSignalManager: a new signal manager instance with common Unix signals registered
func NewUnixSignalManager() *UnixSignalManager {
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
//
// Params:
//   - signals: the signals to listen for
//
// Returns:
//   - <-chan os.Signal: a channel that receives the specified signals
func (m *UnixSignalManager) Notify(signals ...os.Signal) <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	// Return the channel that will receive the specified signals.
	return ch
}

// Stop stops signal notifications on the channel.
//
// Params:
//   - ch: the channel to stop receiving signals on
func (m *UnixSignalManager) Stop(ch chan<- os.Signal) {
	signal.Stop(ch)
}

// Forward sends a signal to a process.
//
// Params:
//   - pid: the process ID to send the signal to
//   - sig: the signal to send
//
// Returns:
//   - error: an error if the signal could not be sent
func (m *UnixSignalManager) Forward(pid int, sig os.Signal) error {
	process, err := os.FindProcess(pid)
	// Check if the process could not be found.
	if err != nil {
		// Return the error from FindProcess.
		return err
	}
	// Return the result of sending the signal to the process.
	return process.Signal(sig)
}

// ForwardToGroup sends a signal to a process group.
//
// Params:
//   - pgid: the process group ID to send the signal to
//   - sig: the signal to send
//
// Returns:
//   - error: an error if the signal could not be sent
func (m *UnixSignalManager) ForwardToGroup(pgid int, sig syscall.Signal) error {
	// Return the result of sending the signal to the process group.
	// Negative PID sends to process group.
	return syscall.Kill(-pgid, sig)
}

// IsTermSignal returns true if the signal is a termination signal.
//
// Params:
//   - sig: the signal to check
//
// Returns:
//   - bool: true if the signal is SIGTERM, SIGINT, SIGQUIT, or SIGKILL
func (m *UnixSignalManager) IsTermSignal(sig os.Signal) bool {
	// Switch on the signal to determine if it is a termination signal.
	switch sig {
	// Case SIGTERM handles the standard termination signal.
	case syscall.SIGTERM:
		// Return true indicating this is a termination signal.
		return true
	// Case SIGINT handles the interrupt signal (Ctrl+C).
	case syscall.SIGINT:
		// Return true indicating this is a termination signal.
		return true
	// Case SIGQUIT handles the quit signal with core dump.
	case syscall.SIGQUIT:
		// Return true indicating this is a termination signal.
		return true
	// Case SIGKILL handles the forced termination signal.
	case syscall.SIGKILL:
		// Return true indicating this is a termination signal.
		return true
	// Case default handles all non-termination signals.
	default:
		// Return false indicating this is not a termination signal.
		return false
	}
}

// IsReloadSignal returns true if the signal is a reload signal.
//
// Params:
//   - sig: the signal to check
//
// Returns:
//   - bool: true if the signal is SIGHUP
func (m *UnixSignalManager) IsReloadSignal(sig os.Signal) bool {
	// Return true if the signal is SIGHUP, the standard reload signal.
	return sig == syscall.SIGHUP
}

// SignalByName returns a signal by its name.
//
// Params:
//   - name: the name of the signal (e.g., "SIGTERM")
//
// Returns:
//   - os.Signal: the signal if found
//   - bool: true if the signal was found
func (m *UnixSignalManager) SignalByName(name string) (os.Signal, bool) {
	sig, ok := m.signalMap[name]
	// Return the signal and whether it was found in the map.
	return sig, ok
}

// AddSignal adds a platform-specific signal to the map.
//
// Params:
//   - name: the name of the signal
//   - sig: the signal value
func (m *UnixSignalManager) AddSignal(name string, sig os.Signal) {
	m.signalMap[name] = sig
}
