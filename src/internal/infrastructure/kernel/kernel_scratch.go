//go:build unix

// Package kernel provides OS abstraction for the daemon.
package kernel

import (
	"github.com/kodflow/daemon/internal/infrastructure/kernel/adapters"
	"github.com/kodflow/daemon/internal/infrastructure/kernel/ports"
)

// NewWithScratchDetection creates a new Kernel with automatic scratch detection.
// If running in a scratch container (no /etc/passwd), it uses ScratchCredentialManager.
// Otherwise, it uses the standard UnixCredentialManager.
//
// Returns:
//   - *Kernel: a new kernel instance with appropriate credential manager
func NewWithScratchDetection() *Kernel {
	var credentials ports.CredentialManager
	if adapters.IsScratchEnvironment() {
		credentials = adapters.NewScratchCredentialManager()
	} else {
		credentials = adapters.NewUnixCredentialManager()
	}

	return &Kernel{
		Signals:     adapters.NewUnixSignalManager(),
		Credentials: credentials,
		Process:     adapters.NewUnixProcessControl(),
		Reaper:      adapters.NewUnixZombieReaper(),
	}
}

// IsScratchEnvironment returns true if running in a scratch container.
// This is a convenience wrapper around adapters.IsScratchEnvironment().
func IsScratchEnvironment() bool {
	return adapters.IsScratchEnvironment()
}
