// Package state provides domain types for daemon state representation.
package state

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// SystemState contains system-wide resource probe.
type SystemState struct {
	// CPU contains system CPU probe.
	CPU probe.SystemCPU `json:"cpu"`
	// Memory contains system memory probe.
	Memory probe.SystemMemory `json:"memory"`
}

// DaemonState represents the complete state of the daemon.
type DaemonState struct {
	// Timestamp is when this state snapshot was taken.
	Timestamp time.Time `json:"timestamp"`
	// Host contains host system information.
	Host HostInfo `json:"host"`
	// Processes contains metrics for all supervised processes.
	Processes []probe.ProcessMetrics `json:"processes"`
	// System contains system-wide resource probe.
	System SystemState `json:"system"`
	// Mesh contains mesh topology if available (optional).
	Mesh *MeshTopology `json:"mesh,omitempty"`
	// Kubernetes contains K8s state if available (optional).
	Kubernetes *KubernetesState `json:"kubernetes,omitempty"`
}

// MeshTopology represents the mesh network topology.
// Populated from IWFS mesh when available.
type MeshTopology struct {
	// LocalNodeID is the ID of this node in the mesh.
	LocalNodeID string `json:"local_node_id"`
	// LeaderID is the current leader node ID.
	LeaderID string `json:"leader_id,omitempty"`
	// Nodes contains all known mesh nodes.
	Nodes []MeshNode `json:"nodes"`
	// Connections contains connections between nodes.
	Connections []MeshConnection `json:"connections"`
	// UpdatedAt is when the topology was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// MeshNode represents a node in the mesh network.
type MeshNode struct {
	// ID is the unique node identifier.
	ID string `json:"id"`
	// Address is the node's network address.
	Address string `json:"address"`
	// State is the node's current state.
	State string `json:"state"`
	// IsLeader indicates if this node is the leader.
	IsLeader bool `json:"is_leader"`
	// LastSeen is when the node was last seen.
	LastSeen time.Time `json:"last_seen"`
}

// MeshConnection represents a connection between mesh nodes.
type MeshConnection struct {
	// FromNodeID is the source node ID.
	FromNodeID string `json:"from_node_id"`
	// ToNodeID is the destination node ID.
	ToNodeID string `json:"to_node_id"`
	// Latency is the connection latency.
	Latency time.Duration `json:"latency"`
	// State is the connection state.
	State string `json:"state"`
}

// KubernetesState represents Kubernetes-related state.
type KubernetesState struct {
	// Namespace is the current namespace.
	Namespace string `json:"namespace"`
	// PodName is the name of this pod.
	PodName string `json:"pod_name,omitempty"`
	// NodeName is the K8s node name.
	NodeName string `json:"node_name,omitempty"`
	// Pods contains discovered pods.
	Pods []KubernetesPod `json:"pods,omitempty"`
	// UpdatedAt is when the K8s state was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// KubernetesPod represents a Kubernetes pod.
type KubernetesPod struct {
	// Name is the pod name.
	Name string `json:"name"`
	// Namespace is the pod namespace.
	Namespace string `json:"namespace"`
	// Phase is the pod phase (Pending, Running, etc.).
	Phase string `json:"phase"`
	// IP is the pod IP address.
	IP string `json:"ip,omitempty"`
	// NodeName is the node where the pod runs.
	NodeName string `json:"node_name"`
	// Labels contains pod labels.
	Labels map[string]string `json:"labels,omitempty"`
}

// ProcessCount returns the number of supervised processes.
func (d DaemonState) ProcessCount() int {
	return len(d.Processes)
}

// RunningProcessCount returns the number of running processes.
func (d DaemonState) RunningProcessCount() int {
	count := 0
	for _, p := range d.Processes {
		if p.IsRunning() {
			count++
		}
	}
	return count
}

// HealthyProcessCount returns the number of healthy processes.
func (d DaemonState) HealthyProcessCount() int {
	count := 0
	for _, p := range d.Processes {
		if p.Healthy {
			count++
		}
	}
	return count
}
