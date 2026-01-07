package process

import (
	"os"
	"syscall"
)

// SignalMap maps signal names to syscall signals.
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

	// Send to process group (negative PID)
	return syscall.Kill(-pid, sig)
}

// IsTermSignal returns true if the signal is a termination signal.
func IsTermSignal(sig os.Signal) bool {
	switch sig {
	case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL:
		return true
	default:
		return false
	}
}

// IsReloadSignal returns true if the signal is a reload signal.
func IsReloadSignal(sig os.Signal) bool {
	return sig == syscall.SIGHUP
}
