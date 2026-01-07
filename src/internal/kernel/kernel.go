// Package kernel provides OS abstraction for the daemon.
package kernel

import (
	"github.com/kodflow/daemon/internal/kernel/adapters"
	"github.com/kodflow/daemon/internal/kernel/ports"
)

// Kernel provides access to all OS abstraction interfaces.
type Kernel struct {
	Signals     ports.SignalManager
	Credentials ports.CredentialManager
	Process     ports.ProcessControl
	Reaper      ports.ZombieReaper
}

// New creates a new Kernel with platform-specific implementations.
func New() *Kernel {
	return &Kernel{
		Signals:     adapters.NewSignalManager(),
		Credentials: adapters.NewCredentialManager(),
		Process:     adapters.NewProcessControl(),
		Reaper:      adapters.NewZombieReaper(),
	}
}

// Default is the default kernel instance.
var Default = New()
