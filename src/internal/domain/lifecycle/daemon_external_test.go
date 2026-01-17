// Package lifecycle_test provides external tests for daemon.go.
package lifecycle_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/lifecycle"
	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
)

func TestSystemState_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state lifecycle.SystemState
	}{
		{
			name:  "empty system state",
			state: lifecycle.SystemState{},
		},
		{
			name: "system state with CPU",
			state: lifecycle.SystemState{
				CPU: metrics.SystemCPU{User: 100, System: 50},
			},
		},
		{
			name: "system state with memory",
			state: lifecycle.SystemState{
				Memory: metrics.SystemMemory{Total: 1024, Available: 512},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.NotNil(t, tt.state)
		})
	}
}

func TestDaemonState_ProcessCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		processes []metrics.ProcessMetrics
		want      int
	}{
		{
			name:      "returns zero for empty processes",
			processes: nil,
			want:      0,
		},
		{
			name: "returns correct count for single process",
			processes: []metrics.ProcessMetrics{
				{ServiceName: "service1"},
			},
			want: 1,
		},
		{
			name: "returns correct count for multiple processes",
			processes: []metrics.ProcessMetrics{
				{ServiceName: "service1"},
				{ServiceName: "service2"},
				{ServiceName: "service3"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state := lifecycle.DaemonState{
				Processes: tt.processes,
			}
			assert.Equal(t, tt.want, state.ProcessCount())
		})
	}
}

func TestDaemonState_RunningProcessCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		processes []metrics.ProcessMetrics
		want      int
	}{
		{
			name:      "returns zero for empty processes",
			processes: nil,
			want:      0,
		},
		{
			name: "counts only running processes",
			processes: []metrics.ProcessMetrics{
				{ServiceName: "service1", State: process.StateRunning},
				{ServiceName: "service2", State: process.StateStopped},
				{ServiceName: "service3", State: process.StateRunning},
				{ServiceName: "service4", State: process.StateFailed},
			},
			want: 2,
		},
		{
			name: "returns zero when no processes running",
			processes: []metrics.ProcessMetrics{
				{ServiceName: "service1", State: process.StateStopped},
				{ServiceName: "service2", State: process.StateFailed},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state := lifecycle.DaemonState{
				Processes: tt.processes,
			}
			assert.Equal(t, tt.want, state.RunningProcessCount())
		})
	}
}

func TestDaemonState_HealthyProcessCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		processes []metrics.ProcessMetrics
		want      int
	}{
		{
			name:      "returns zero for empty processes",
			processes: nil,
			want:      0,
		},
		{
			name: "counts only healthy processes",
			processes: []metrics.ProcessMetrics{
				{ServiceName: "service1", Healthy: true},
				{ServiceName: "service2", Healthy: false},
				{ServiceName: "service3", Healthy: true},
			},
			want: 2,
		},
		{
			name: "returns zero when no healthy processes",
			processes: []metrics.ProcessMetrics{
				{ServiceName: "service1", Healthy: false},
				{ServiceName: "service2", Healthy: false},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state := lifecycle.DaemonState{
				Processes: tt.processes,
			}
			assert.Equal(t, tt.want, state.HealthyProcessCount())
		})
	}
}

func TestDaemonState_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		timestamp     time.Time
		host          lifecycle.HostInfo
		processes     []metrics.ProcessMetrics
		system        lifecycle.SystemState
		wantHostname  string
		wantOS        string
		wantArch      string
		wantVersion   string
		wantProcCount int
	}{
		{
			name:          "empty daemon state",
			timestamp:     time.Time{},
			host:          lifecycle.HostInfo{},
			processes:     nil,
			system:        lifecycle.SystemState{},
			wantHostname:  "",
			wantOS:        "",
			wantArch:      "",
			wantVersion:   "",
			wantProcCount: 0,
		},
		{
			name:      "full daemon state",
			timestamp: time.Now(),
			host: lifecycle.HostInfo{
				Hostname:      "testhost",
				OS:            "linux",
				Arch:          "amd64",
				DaemonVersion: "1.0.0",
			},
			processes: []metrics.ProcessMetrics{
				{ServiceName: "test-service"},
			},
			system: lifecycle.SystemState{
				CPU:    metrics.SystemCPU{User: 100},
				Memory: metrics.SystemMemory{Total: 1024},
			},
			wantHostname:  "testhost",
			wantOS:        "linux",
			wantArch:      "amd64",
			wantVersion:   "1.0.0",
			wantProcCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state := lifecycle.DaemonState{
				Timestamp: tt.timestamp,
				Host:      tt.host,
				Processes: tt.processes,
				System:    tt.system,
			}

			assert.Equal(t, tt.timestamp, state.Timestamp)
			assert.Equal(t, tt.wantHostname, state.Host.Hostname)
			assert.Equal(t, tt.wantOS, state.Host.OS)
			assert.Equal(t, tt.wantArch, state.Host.Arch)
			assert.Equal(t, tt.wantVersion, state.Host.DaemonVersion)
			assert.Len(t, state.Processes, tt.wantProcCount)
		})
	}
}

func TestMeshTopology_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		localNodeID   string
		leaderID      string
		nodes         []lifecycle.MeshNode
		connections   []lifecycle.MeshConnection
		wantNodeCount int
		wantConnCount int
		wantLatency   time.Duration
	}{
		{
			name:          "empty mesh topology",
			localNodeID:   "",
			leaderID:      "",
			nodes:         nil,
			connections:   nil,
			wantNodeCount: 0,
			wantConnCount: 0,
			wantLatency:   0,
		},
		{
			name:        "mesh with two nodes",
			localNodeID: "node-1",
			leaderID:    "node-2",
			nodes: []lifecycle.MeshNode{
				{ID: "node-1", Address: "10.0.0.1:9090", State: "ready"},
				{ID: "node-2", Address: "10.0.0.2:9090", State: "ready", IsLeader: true},
			},
			connections: []lifecycle.MeshConnection{
				{FromNodeID: "node-1", ToNodeID: "node-2", Latency: 5 * time.Millisecond, State: "connected"},
			},
			wantNodeCount: 2,
			wantConnCount: 1,
			wantLatency:   5 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mesh := lifecycle.MeshTopology{
				LocalNodeID: tt.localNodeID,
				LeaderID:    tt.leaderID,
				Nodes:       tt.nodes,
				Connections: tt.connections,
			}

			assert.Equal(t, tt.localNodeID, mesh.LocalNodeID)
			assert.Equal(t, tt.leaderID, mesh.LeaderID)
			assert.Len(t, mesh.Nodes, tt.wantNodeCount)
			assert.Len(t, mesh.Connections, tt.wantConnCount)
			if tt.wantConnCount > 0 {
				assert.Equal(t, tt.wantLatency, mesh.Connections[0].Latency)
			}
		})
	}
}

func TestKubernetesState_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		namespace    string
		podName      string
		nodeName     string
		pods         []lifecycle.KubernetesPod
		wantPodCount int
		wantPhase    string
		wantLabel    string
	}{
		{
			name:         "empty kubernetes state",
			namespace:    "",
			podName:      "",
			nodeName:     "",
			pods:         nil,
			wantPodCount: 0,
			wantPhase:    "",
			wantLabel:    "",
		},
		{
			name:      "kubernetes state with pod",
			namespace: "default",
			podName:   "daemon-pod-abc123",
			nodeName:  "worker-1",
			pods: []lifecycle.KubernetesPod{
				{
					Name:      "app-pod-1",
					Namespace: "default",
					Phase:     "Running",
					IP:        "10.244.0.5",
					NodeName:  "worker-1",
					Labels:    map[string]string{"app": "myapp"},
				},
			},
			wantPodCount: 1,
			wantPhase:    "Running",
			wantLabel:    "myapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			k8s := lifecycle.KubernetesState{
				Namespace: tt.namespace,
				PodName:   tt.podName,
				NodeName:  tt.nodeName,
				Pods:      tt.pods,
			}

			assert.Equal(t, tt.namespace, k8s.Namespace)
			assert.Equal(t, tt.podName, k8s.PodName)
			assert.Len(t, k8s.Pods, tt.wantPodCount)
			if tt.wantPodCount > 0 {
				assert.Equal(t, tt.wantPhase, k8s.Pods[0].Phase)
				assert.Equal(t, tt.wantLabel, k8s.Pods[0].Labels["app"])
			}
		})
	}
}
