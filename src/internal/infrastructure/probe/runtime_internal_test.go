//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRuntimeNames verifies all runtime types are mapped.
func TestRuntimeNames(t *testing.T) {
	tests := []struct {
		name    string
		runtime RuntimeType
	}{
		{name: "RuntimeNone", runtime: RuntimeNone},
		{name: "RuntimeDocker", runtime: RuntimeDocker},
		{name: "RuntimePodman", runtime: RuntimePodman},
		{name: "RuntimeContainerd", runtime: RuntimeContainerd},
		{name: "RuntimeCriO", runtime: RuntimeCriO},
		{name: "RuntimeLXC", runtime: RuntimeLXC},
		{name: "RuntimeLXD", runtime: RuntimeLXD},
		{name: "RuntimeSystemdNspawn", runtime: RuntimeSystemdNspawn},
		{name: "RuntimeFirecracker", runtime: RuntimeFirecracker},
		{name: "RuntimeFreeBSDJail", runtime: RuntimeFreeBSDJail},
		{name: "RuntimeKubernetes", runtime: RuntimeKubernetes},
		{name: "RuntimeNomad", runtime: RuntimeNomad},
		{name: "RuntimeDockerSwarm", runtime: RuntimeDockerSwarm},
		{name: "RuntimeOpenShift", runtime: RuntimeOpenShift},
		{name: "RuntimeAWSECS", runtime: RuntimeAWSECS},
		{name: "RuntimeAWSFargate", runtime: RuntimeAWSFargate},
		{name: "RuntimeGoogleGKE", runtime: RuntimeGoogleGKE},
		{name: "RuntimeAzureAKS", runtime: RuntimeAzureAKS},
		{name: "RuntimeUnknown", runtime: RuntimeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, ok := runtimeNames[tt.runtime]
			assert.True(t, ok, "runtime %d should be mapped", tt.runtime)
		})
	}
}

// TestRuntimeTypeConstants verifies runtime type constant values.
func TestRuntimeTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		runtime  RuntimeType
		expected RuntimeType
	}{
		{name: "RuntimeNone is 0", runtime: RuntimeNone, expected: RuntimeType(0)},
		{name: "RuntimeDocker is 1", runtime: RuntimeDocker, expected: RuntimeType(1)},
		{name: "RuntimePodman is 2", runtime: RuntimePodman, expected: RuntimeType(2)},
		{name: "RuntimeContainerd is 3", runtime: RuntimeContainerd, expected: RuntimeType(3)},
		{name: "RuntimeCriO is 4", runtime: RuntimeCriO, expected: RuntimeType(4)},
		{name: "RuntimeLXC is 5", runtime: RuntimeLXC, expected: RuntimeType(5)},
		{name: "RuntimeLXD is 6", runtime: RuntimeLXD, expected: RuntimeType(6)},
		{name: "RuntimeSystemdNspawn is 7", runtime: RuntimeSystemdNspawn, expected: RuntimeType(7)},
		{name: "RuntimeFirecracker is 8", runtime: RuntimeFirecracker, expected: RuntimeType(8)},
		{name: "RuntimeFreeBSDJail is 9", runtime: RuntimeFreeBSDJail, expected: RuntimeType(9)},
		{name: "RuntimeKubernetes is 20", runtime: RuntimeKubernetes, expected: RuntimeType(20)},
		{name: "RuntimeNomad is 21", runtime: RuntimeNomad, expected: RuntimeType(21)},
		{name: "RuntimeDockerSwarm is 22", runtime: RuntimeDockerSwarm, expected: RuntimeType(22)},
		{name: "RuntimeOpenShift is 23", runtime: RuntimeOpenShift, expected: RuntimeType(23)},
		{name: "RuntimeAWSECS is 40", runtime: RuntimeAWSECS, expected: RuntimeType(40)},
		{name: "RuntimeAWSFargate is 41", runtime: RuntimeAWSFargate, expected: RuntimeType(41)},
		{name: "RuntimeGoogleGKE is 42", runtime: RuntimeGoogleGKE, expected: RuntimeType(42)},
		{name: "RuntimeAzureAKS is 43", runtime: RuntimeAzureAKS, expected: RuntimeType(43)},
		{name: "RuntimeUnknown is 254", runtime: RuntimeUnknown, expected: RuntimeType(254)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.runtime)
		})
	}
}

// TestAvailableRuntime_Structure verifies AvailableRuntime struct fields.
func TestAvailableRuntime_Structure(t *testing.T) {
	tests := []struct {
		name       string
		runtime    RuntimeType
		socketPath string
		version    string
		isRunning  bool
	}{
		{
			name:       "docker runtime",
			runtime:    RuntimeDocker,
			socketPath: "/var/run/docker.sock",
			version:    "20.10.0",
			isRunning:  true,
		},
		{
			name:       "podman runtime",
			runtime:    RuntimePodman,
			socketPath: "/run/user/1000/podman/podman.sock",
			version:    "4.0.0",
			isRunning:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ar := AvailableRuntime{
				Runtime:    tt.runtime,
				SocketPath: tt.socketPath,
				Version:    tt.version,
				IsRunning:  tt.isRunning,
			}

			assert.Equal(t, tt.runtime, ar.Runtime)
			assert.Equal(t, tt.socketPath, ar.SocketPath)
			assert.Equal(t, tt.version, ar.Version)
			assert.Equal(t, tt.isRunning, ar.IsRunning)
		})
	}
}

// TestRuntimeInfo_Structure verifies RuntimeInfo struct fields.
func TestRuntimeInfo_Structure(t *testing.T) {
	tests := []struct {
		name             string
		isContainerized  bool
		containerRuntime RuntimeType
		orchestrator     RuntimeType
		containerID      string
		workloadID       string
		workloadName     string
		namespace        string
	}{
		{
			name:             "kubernetes pod",
			isContainerized:  true,
			containerRuntime: RuntimeDocker,
			orchestrator:     RuntimeKubernetes,
			containerID:      "abc123",
			workloadID:       "pod-uid",
			workloadName:     "my-pod",
			namespace:        "default",
		},
		{
			name:             "bare metal",
			isContainerized:  false,
			containerRuntime: RuntimeNone,
			orchestrator:     RuntimeNone,
			containerID:      "",
			workloadID:       "",
			workloadName:     "",
			namespace:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info := RuntimeInfo{
				IsContainerized:  tt.isContainerized,
				ContainerRuntime: tt.containerRuntime,
				Orchestrator:     tt.orchestrator,
				ContainerID:      tt.containerID,
				WorkloadID:       tt.workloadID,
				WorkloadName:     tt.workloadName,
				Namespace:        tt.namespace,
			}

			assert.Equal(t, tt.isContainerized, info.IsContainerized)
			assert.Equal(t, tt.containerRuntime, info.ContainerRuntime)
			assert.Equal(t, tt.orchestrator, info.Orchestrator)
			assert.Equal(t, tt.containerID, info.ContainerID)
			assert.Equal(t, tt.workloadID, info.WorkloadID)
			assert.Equal(t, tt.workloadName, info.WorkloadName)
			assert.Equal(t, tt.namespace, info.Namespace)
		})
	}
}
