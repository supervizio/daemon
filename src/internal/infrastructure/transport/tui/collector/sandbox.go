// Package collector provides data collectors for TUI snapshot.
package collector

import (
	"os"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// SandboxCollector detects container runtimes.
type SandboxCollector struct{}

// NewSandboxCollector creates a sandbox collector.
func NewSandboxCollector() *SandboxCollector {
	return &SandboxCollector{}
}

// sandboxCheck defines a runtime detection check.
type sandboxCheck struct {
	name      string
	endpoints []string
}

// CollectInto detects container runtimes.
func (c *SandboxCollector) CollectInto(snap *model.Snapshot) error {
	checks := []sandboxCheck{
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
				"/run/user/1000/podman/podman.sock",
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

	snap.Sandboxes = make([]model.SandboxInfo, 0, len(checks))

	for _, check := range checks {
		info := model.SandboxInfo{
			Name:     check.name,
			Detected: false,
		}

		for _, endpoint := range check.endpoints {
			if _, err := os.Stat(endpoint); err == nil {
				info.Detected = true
				info.Endpoint = endpoint
				break
			}
		}

		snap.Sandboxes = append(snap.Sandboxes, info)
	}

	return nil
}
