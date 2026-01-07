package process

import (
	"os"
	"syscall"

	"github.com/kodflow/daemon/internal/kernel"
)

// SignalMap maps signal names to syscall signals.
// Deprecated: Use kernel.Default.Signals.SignalByName instead.
var SignalMap = map[string]os.Signal{
	"SIGHUP":  syscall.SIGHUP,
	"SIGINT":  syscall.SIGINT,
	"SIGQUIT": syscall.SIGQUIT,
	"SIGTERM": syscall.SIGTERM,
	"SIGUSR1": syscall.SIGUSR1,
	"SIGUSR2": syscall.SIGUSR2,
}

// ForwardSignal forwards a signal to a process.
func ForwardSignal(p *Process, sig os.Signal) error {
	if p == nil || p.State() != StateRunning {
		return nil
	}
	return p.Signal(sig)
}

// ForwardSignalToGroup forwards a signal to a process group.
func ForwardSignalToGroup(p *Process, sig syscall.Signal) error {
	if p == nil || p.State() != StateRunning {
		return nil
	}

	pid := p.PID()
	if pid <= 0 {
		return nil
	}

	return kernel.Default.Signals.ForwardToGroup(pid, sig)
}

// IsTermSignal returns true if the signal is a termination signal.
func IsTermSignal(sig os.Signal) bool {
	return kernel.Default.Signals.IsTermSignal(sig)
}

// IsReloadSignal returns true if the signal is a reload signal.
func IsReloadSignal(sig os.Signal) bool {
	return kernel.Default.Signals.IsReloadSignal(sig)
}
