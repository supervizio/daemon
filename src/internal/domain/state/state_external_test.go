// Package state_test provides external tests for the state package.
package state_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/domain/state"
)

func TestHostInfo_Uptime(t *testing.T) {
	t.Parallel()

	t.Run("returns positive uptime for started host", func(t *testing.T) {
		t.Parallel()

		host := state.HostInfo{
			StartTime: time.Now().Add(-5 * time.Minute),
		}

		uptime := host.Uptime()
		assert.True(t, uptime >= 5*time.Minute)
		assert.True(t, uptime < 6*time.Minute)
	})

	t.Run("returns zero for zero start time", func(t *testing.T) {
		t.Parallel()

		host := state.HostInfo{}
		assert.Equal(t, time.Duration(0), host.Uptime())
	})
}

func TestDaemonState_ProcessCount(t *testing.T) {
	t.Parallel()

	t.Run("returns correct count", func(t *testing.T) {
		t.Parallel()

		state := state.DaemonState{
			Processes: []metrics.ProcessMetrics{
				{ServiceName: "service1"},
				{ServiceName: "service2"},
				{ServiceName: "service3"},
			},
		}

		assert.Equal(t, 3, state.ProcessCount())
	})

	t.Run("returns zero for empty processes", func(t *testing.T) {
		t.Parallel()

		state := state.DaemonState{}
		assert.Equal(t, 0, state.ProcessCount())
	})
}

func TestDaemonState_RunningProcessCount(t *testing.T) {
	t.Parallel()

	t.Run("counts only running processes", func(t *testing.T) {
		t.Parallel()

		state := state.DaemonState{
			Processes: []metrics.ProcessMetrics{
				{ServiceName: "service1", State: process.StateRunning},
				{ServiceName: "service2", State: process.StateStopped},
				{ServiceName: "service3", State: process.StateRunning},
				{ServiceName: "service4", State: process.StateFailed},
			},
		}

		assert.Equal(t, 2, state.RunningProcessCount())
	})
}

func TestDaemonState_HealthyProcessCount(t *testing.T) {
	t.Parallel()

	t.Run("counts only healthy processes", func(t *testing.T) {
		t.Parallel()

		state := state.DaemonState{
			Processes: []metrics.ProcessMetrics{
				{ServiceName: "service1", Healthy: true},
				{ServiceName: "service2", Healthy: false},
				{ServiceName: "service3", Healthy: true},
			},
		}

		assert.Equal(t, 2, state.HealthyProcessCount())
	})
}

func TestDaemonState_Fields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	state := state.DaemonState{
		Timestamp: now,
		Host: state.HostInfo{
			Hostname:      "testhost",
			OS:            "linux",
			Arch:          "amd64",
			DaemonVersion: "1.0.0",
		},
		Processes: []metrics.ProcessMetrics{
			{ServiceName: "test-service"},
		},
		System: state.SystemState{
			CPU:    metrics.SystemCPU{User: 100},
			Memory: metrics.SystemMemory{Total: 1024},
		},
	}

	assert.Equal(t, now, state.Timestamp)
	assert.Equal(t, "testhost", state.Host.Hostname)
	assert.Equal(t, "linux", state.Host.OS)
	assert.Equal(t, "amd64", state.Host.Arch)
	assert.Equal(t, "1.0.0", state.Host.DaemonVersion)
	assert.Len(t, state.Processes, 1)
	assert.Equal(t, uint64(100), state.System.CPU.User)
	assert.Equal(t, uint64(1024), state.System.Memory.Total)
}

func TestMeshTopology_Fields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	mesh := state.MeshTopology{
		LocalNodeID: "node-1",
		LeaderID:    "node-2",
		Nodes: []state.MeshNode{
			{ID: "node-1", Address: "10.0.0.1:9090", State: "ready"},
			{ID: "node-2", Address: "10.0.0.2:9090", State: "ready", IsLeader: true},
		},
		Connections: []state.MeshConnection{
			{FromNodeID: "node-1", ToNodeID: "node-2", Latency: 5 * time.Millisecond, State: "connected"},
		},
		UpdatedAt: now,
	}

	assert.Equal(t, "node-1", mesh.LocalNodeID)
	assert.Equal(t, "node-2", mesh.LeaderID)
	assert.Len(t, mesh.Nodes, 2)
	assert.Len(t, mesh.Connections, 1)
	assert.Equal(t, 5*time.Millisecond, mesh.Connections[0].Latency)
}

func TestKubernetesState_Fields(t *testing.T) {
	t.Parallel()

	k8s := state.KubernetesState{
		Namespace: "default",
		PodName:   "daemon-pod-abc123",
		NodeName:  "worker-1",
		Pods: []state.KubernetesPod{
			{
				Name:      "app-pod-1",
				Namespace: "default",
				Phase:     "Running",
				IP:        "10.244.0.5",
				NodeName:  "worker-1",
				Labels:    map[string]string{"app": "myapp"},
			},
		},
	}

	assert.Equal(t, "default", k8s.Namespace)
	assert.Equal(t, "daemon-pod-abc123", k8s.PodName)
	assert.Len(t, k8s.Pods, 1)
	assert.Equal(t, "Running", k8s.Pods[0].Phase)
	assert.Equal(t, "myapp", k8s.Pods[0].Labels["app"])
}
