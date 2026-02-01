//go:build cgo

package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectRuntime verifies full runtime detection.
func TestDetectRuntime(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "detects runtime info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			info, err := probe.DetectRuntime()
			require.NoError(t, err)
			require.NotNil(t, info)

			// Log runtime info for debugging
			t.Logf("Runtime: IsContainerized=%v, Runtime=%v, Orchestrator=%v",
				info.IsContainerized, info.ContainerRuntime, info.Orchestrator)

			if len(info.AvailableRuntimes) > 0 {
				t.Logf("Available runtimes: %d", len(info.AvailableRuntimes))
				for _, ar := range info.AvailableRuntimes {
					t.Logf("  - %s (running=%v)", ar.Runtime, ar.IsRunning)
				}
			}
		})
	}
}

// TestIsContainerized verifies containerization detection.
func TestIsContainerized(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "detects containerization status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			// Just verify the function can be called
			result := probe.IsContainerized()
			t.Logf("IsContainerized: %v", result)
		})
	}
}

// TestGetRuntimeName verifies runtime name retrieval.
func TestGetRuntimeName(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns non-empty runtime name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			name := probe.GetRuntimeName()
			assert.NotEmpty(t, name)
			t.Logf("Runtime name: %s", name)
		})
	}
}

// TestRuntimeType_String verifies RuntimeType string conversion.
func TestRuntimeType_String(t *testing.T) {
	tests := []struct {
		name     string
		rt       probe.RuntimeType
		expected string
	}{
		{name: "none", rt: probe.RuntimeNone, expected: "none"},
		{name: "docker", rt: probe.RuntimeDocker, expected: "docker"},
		{name: "podman", rt: probe.RuntimePodman, expected: "podman"},
		{name: "containerd", rt: probe.RuntimeContainerd, expected: "containerd"},
		{name: "cri-o", rt: probe.RuntimeCriO, expected: "cri-o"},
		{name: "lxc", rt: probe.RuntimeLXC, expected: "lxc"},
		{name: "lxd", rt: probe.RuntimeLXD, expected: "lxd"},
		{name: "systemd-nspawn", rt: probe.RuntimeSystemdNspawn, expected: "systemd-nspawn"},
		{name: "firecracker", rt: probe.RuntimeFirecracker, expected: "firecracker"},
		{name: "freebsd-jail", rt: probe.RuntimeFreeBSDJail, expected: "freebsd-jail"},
		{name: "kubernetes", rt: probe.RuntimeKubernetes, expected: "kubernetes"},
		{name: "nomad", rt: probe.RuntimeNomad, expected: "nomad"},
		{name: "docker-swarm", rt: probe.RuntimeDockerSwarm, expected: "docker-swarm"},
		{name: "openshift", rt: probe.RuntimeOpenShift, expected: "openshift"},
		{name: "aws-ecs", rt: probe.RuntimeAWSECS, expected: "aws-ecs"},
		{name: "aws-fargate", rt: probe.RuntimeAWSFargate, expected: "aws-fargate"},
		{name: "google-gke", rt: probe.RuntimeGoogleGKE, expected: "google-gke"},
		{name: "azure-aks", rt: probe.RuntimeAzureAKS, expected: "azure-aks"},
		{name: "unknown", rt: probe.RuntimeUnknown, expected: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.rt.String())
		})
	}
}

// TestRuntimeType_IsOrchestrator verifies orchestrator detection.
func TestRuntimeType_IsOrchestrator(t *testing.T) {
	tests := []struct {
		name     string
		rt       probe.RuntimeType
		expected bool
	}{
		{name: "kubernetes is orchestrator", rt: probe.RuntimeKubernetes, expected: true},
		{name: "nomad is orchestrator", rt: probe.RuntimeNomad, expected: true},
		{name: "docker-swarm is orchestrator", rt: probe.RuntimeDockerSwarm, expected: true},
		{name: "openshift is orchestrator", rt: probe.RuntimeOpenShift, expected: true},
		{name: "aws-ecs is orchestrator", rt: probe.RuntimeAWSECS, expected: true},
		{name: "aws-fargate is orchestrator", rt: probe.RuntimeAWSFargate, expected: true},
		{name: "google-gke is orchestrator", rt: probe.RuntimeGoogleGKE, expected: true},
		{name: "azure-aks is orchestrator", rt: probe.RuntimeAzureAKS, expected: true},
		{name: "none is not orchestrator", rt: probe.RuntimeNone, expected: false},
		{name: "docker is not orchestrator", rt: probe.RuntimeDocker, expected: false},
		{name: "podman is not orchestrator", rt: probe.RuntimePodman, expected: false},
		{name: "containerd is not orchestrator", rt: probe.RuntimeContainerd, expected: false},
		{name: "lxc is not orchestrator", rt: probe.RuntimeLXC, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.rt.IsOrchestrator())
		})
	}
}

// TestRuntimeType_String_Unknown verifies unknown runtime handling.
func TestRuntimeType_String_Unknown(t *testing.T) {
	tests := []struct {
		name     string
		rt       probe.RuntimeType
		expected string
	}{
		{name: "unknown value 999", rt: probe.RuntimeType(999), expected: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.rt.String())
		})
	}
}
