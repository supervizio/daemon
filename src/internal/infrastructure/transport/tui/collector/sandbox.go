// Package collector provides data collectors for TUI snapshot.
package collector

import (
	"fmt"
	"os"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// SandboxCollector detects container runtimes.
// It checks for socket files and tokens to identify Docker, Podman, Kubernetes, etc.
type SandboxCollector struct{}

// NewSandboxCollector creates a sandbox collector.
//
// Returns:
//   - *SandboxCollector: new sandbox collector instance
func NewSandboxCollector() *SandboxCollector {
	// Return empty struct (no state needed).
	return &SandboxCollector{}
}

// Gather detects container runtimes.
//
// Params:
//   - snap: snapshot to populate with sandbox information
//
// Returns:
//   - error: nil (always succeeds)
func (c *SandboxCollector) Gather(snap *model.Snapshot) error {
	checks := getSandboxChecks()
	snap.Sandboxes = make([]model.SandboxInfo, 0, len(checks))

	// Process each sandbox check.
	for _, check := range checks {
		info := c.detectSandbox(check)
		snap.Sandboxes = append(snap.Sandboxes, info)
	}

	// Successfully collected sandbox info.
	return nil
}

// getSandboxChecks returns the list of sandbox checks to perform.
//
// Returns:
//   - []sandboxCheck: list of sandbox detection configurations
func getSandboxChecks() []sandboxCheck {
	// Return all sandbox check configurations.
	return []sandboxCheck{
		{
			name: "Docker",
			endpoints: []string{
				"/var/run/docker.sock",
				"/run/docker.sock",
			},
		},
		{
			name: "Podman",
			endpoints: []string{
				"/var/run/podman/podman.sock",
				"/run/podman/podman.sock",
				fmt.Sprintf("/run/user/%d/podman/podman.sock", os.Getuid()),
			},
		},
		{
			name: "containerd",
			endpoints: []string{
				"/var/run/containerd/containerd.sock",
				"/run/containerd/containerd.sock",
			},
		},
		{
			name: "Kubernetes",
			endpoints: []string{
				"/var/run/secrets/kubernetes.io/serviceaccount/token",
			},
		},
		{
			name: "LXC/LXD",
			endpoints: []string{
				"/var/lib/lxd/unix.socket",
				"/var/snap/lxd/common/lxd/unix.socket",
			},
		},
		{
			name: "CRI-O",
			endpoints: []string{
				"/var/run/crio/crio.sock",
				"/run/crio/crio.sock",
			},
		},
	}
}

// detectSandbox checks if a sandbox is present by checking endpoints.
//
// Params:
//   - check: sandbox check configuration
//
// Returns:
//   - model.SandboxInfo: sandbox detection result
func (c *SandboxCollector) detectSandbox(check sandboxCheck) model.SandboxInfo {
	info := model.SandboxInfo{
		Name:     check.name,
		Detected: false,
	}

	// Check each endpoint for this sandbox.
	for _, endpoint := range check.endpoints {
		// Check if endpoint exists.
		if _, err := os.Stat(endpoint); err == nil {
			info.Detected = true
			info.Endpoint = endpoint
			// Stop after first match.
			break
		}
	}

	// Return sandbox info.
	return info
}
