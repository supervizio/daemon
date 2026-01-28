//go:build unix

// Package signals provides signal handling for Unix systems.
package signals

import (
	"os"
	"os/signal"
	"syscall"
)

// Manager implements SignalManager for Unix systems.
// It provides signal handling capabilities including notification, forwarding, and signal identification.
type Manager struct {
	// signalMap maps signal names to their os.Signal values.
	signalMap map[string]os.Signal
}

// NewManager returns a Manager with common Unix signals pre-registered.
//
// Returns:
//   - *Manager: signal manager with POSIX signals mapped by name
func NewManager() *Manager {
	return &Manager{signalMap: map[string]os.Signal{
		"SIGHUP": syscall.SIGHUP, "SIGINT": syscall.SIGINT, "SIGQUIT": syscall.SIGQUIT,
		"SIGTERM": syscall.SIGTERM, "SIGKILL": syscall.SIGKILL, "SIGUSR1": syscall.SIGUSR1,
		"SIGUSR2": syscall.SIGUSR2, "SIGCHLD": syscall.SIGCHLD,
	}}
}

// New returns a Manager with common Unix signals pre-registered.
//
// Returns:
//   - *Manager: signal manager with POSIX signals mapped by name
func New() *Manager {
	return &Manager{signalMap: map[string]os.Signal{
		"SIGHUP": syscall.SIGHUP, "SIGINT": syscall.SIGINT, "SIGQUIT": syscall.SIGQUIT,
		"SIGTERM": syscall.SIGTERM, "SIGKILL": syscall.SIGKILL, "SIGUSR1": syscall.SIGUSR1,
		"SIGUSR2": syscall.SIGUSR2, "SIGCHLD": syscall.SIGCHLD,
	}}
}

// Notify returns a channel that receives the specified signals.
//
// Params:
//   - signals: list of signals to subscribe to
//
// Returns:
//   - <-chan os.Signal: buffered channel receiving signal notifications
func (m *Manager) Notify(signals ...os.Signal) <-chan os.Signal {
	// Buffer of 1 prevents signal loss during handler processing.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	return ch
}

// Stop unregisters the channel from signal notifications.
//
// Params:
//   - ch: channel to unregister from signal delivery
func (m *Manager) Stop(ch chan<- os.Signal) { signal.Stop(ch) }

// Forward delivers a signal to a process via os.Process handle.
//
// Params:
//   - pid: target process ID
//   - sig: signal to deliver
//
// Returns:
//   - error: if signal delivery fails (process not found, permission denied, etc)
func (m *Manager) Forward(pid int, sig os.Signal) error {
	// FindProcess always succeeds on Unix; error deferred to Signal call.
	process, _ := os.FindProcess(pid)
	return process.Signal(sig)
}

// ForwardToGroup delivers a signal to all processes in a group.
//
// Params:
//   - pgid: process group ID (negated internally for kill syscall)
//   - sig: signal to deliver to all group members
//
// Returns:
//   - error: if signal delivery fails
func (m *Manager) ForwardToGroup(pgid int, sig syscall.Signal) error {
	// Negative PID signals all processes in the group.
	return syscall.Kill(-pgid, sig)
}

// IsTermSignal checks for SIGTERM, SIGINT, SIGQUIT, or SIGKILL.
//
// Params:
//   - sig: signal to classify
//
// Returns:
//   - bool: true for termination signals
func (m *Manager) IsTermSignal(sig os.Signal) bool {
	// Standard termination signals per POSIX convention.
	switch sig {
	case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL:
		return true
	default:
		return false
	}
}

// IsReloadSignal checks for SIGHUP (config reload convention).
//
// Params:
//   - sig: signal to classify
//
// Returns:
//   - bool: true for SIGHUP (Unix reload convention)
func (m *Manager) IsReloadSignal(sig os.Signal) bool { return sig == syscall.SIGHUP }

// SignalByName looks up a signal by name from the registered map.
//
// Params:
//   - name: signal name (e.g., "SIGTERM", "SIGHUP")
//
// Returns:
//   - os.Signal: the signal if found
//   - bool: true if signal name was registered
func (m *Manager) SignalByName(name string) (os.Signal, bool) {
	sig, ok := m.signalMap[name]
	return sig, ok
}

// AddSignal registers a platform-specific signal for name lookup.
//
// Params:
//   - name: signal name to register (e.g., "SIGPWR" on Linux)
//   - sig: signal value to associate with name
func (m *Manager) AddSignal(name string, sig os.Signal) { m.signalMap[name] = sig }
