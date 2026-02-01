// Package lifecycle provides domain types for daemon lifecycle management.
package lifecycle

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// SystemState contains system-wide resource metrics.
//
// This aggregates CPU and memory metrics for the entire system.
type SystemState struct {
	// CPU contains system CPU metrics.
	CPU metrics.SystemCPU `json:"cpu"`
	// Memory contains system memory metrics.
	Memory metrics.SystemMemory `json:"memory"`
}

// DaemonState represents the complete state of the daemon.
//
// This is a snapshot of all daemon state at a specific point in time,
// including host info, supervised processes, system metrics, and optional
// mesh/Kubernetes topology when available.
type DaemonState struct {
	// Timestamp is when this state snapshot was taken.
	Timestamp time.Time `json:"timestamp"`
	// Host contains host system information.
	Host HostInfo `json:"host"`
	// Processes contains metrics for all supervised processes.
	Processes []metrics.ProcessMetrics `json:"processes"`
	// System contains system-wide resource metrics.
	System SystemState `json:"system"`
	// Mesh contains mesh topology if available (optional).
	Mesh *MeshTopology `json:"mesh,omitempty"`
	// Kubernetes contains K8s state if available (optional).
	Kubernetes *KubernetesState `json:"kubernetes,omitempty"`
}

// MeshTopology represents the mesh network topology.
//
// Populated from IWFS mesh when available, containing information about
// all known nodes, their connections, and the current leader.
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
//
// Contains information about a single node's identity, state, and last contact.
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
//
// Contains information about a directed connection from one node to another,
// including latency and connection state.
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
//
// Contains information about the current K8s context, including namespace,
// pod identity, and discovered pods.
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
//
// Contains information about a single pod's identity, state, and location.
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
//
// Returns:
//   - int: total count of processes in this daemon state
func (d *DaemonState) ProcessCount() int {
	// return total number of processes
	return len(d.Processes)
}

// RunningProcessCount returns the number of running processes.
//
// Returns:
//   - int: count of processes currently in running state
func (d *DaemonState) RunningProcessCount() int {
	count := 0
	// iterate over all processes to check running state
	for i := range d.Processes {
		// count process if it is running
		if d.Processes[i].IsRunning() {
			count++
		}
	}
	// return total running process count
	return count
}

// HealthyProcessCount returns the number of healthy processes.
//
// Returns:
//   - int: count of processes currently marked as healthy
func (d *DaemonState) HealthyProcessCount() int {
	count := 0
	// iterate over all processes to check health
	for i := range d.Processes {
		// count process if it is healthy
		if d.Processes[i].Healthy {
			count++
		}
	}
	// return total healthy process count
	return count
}
