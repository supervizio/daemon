// Package kernel provides OS abstraction for the daemon.
package kernel

import (
	"github.com/kodflow/daemon/internal/kernel/adapters"
	"github.com/kodflow/daemon/internal/kernel/ports"
)

// Kernel provides access to all OS abstraction interfaces.
// It aggregates platform-specific implementations for signals, credentials, process, and zombie reaping.
type Kernel struct {
	// Signals handles signal notification and forwarding operations.
	Signals ports.SignalManager
	// Credentials handles user and group credential resolution and application.
	Credentials ports.CredentialManager
	// Process handles process group operations.
	Process ports.ProcessControl
	// Reaper handles zombie process cleanup.
	Reaper ports.ZombieReaper
}

// New creates a new Kernel with platform-specific implementations.
//
// Returns:
//   - *Kernel: a new kernel instance with all interfaces initialized
func New() *Kernel {
	// Return a new Kernel with all platform-specific adapters initialized.
	return &Kernel{
		Signals:     adapters.NewUnixSignalManager(),
		Credentials: adapters.NewUnixCredentialManager(),
		Process:     adapters.NewUnixProcessControl(),
		Reaper:      adapters.NewUnixZombieReaper(),
	}
}

// Default is the default kernel instance.
var Default *Kernel = New()
