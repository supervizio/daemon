// Package event provides domain types for event handling.
package event

import "time"

// unknownString is the string representation for unknown types.
const unknownString = "unknown"

// Type represents the type of event.
type Type int

// Event types for daemon lifecycle and state changes.
const (
	TypeUnknown Type = iota
	// Process events
	TypeProcessStarted
	TypeProcessStopped
	TypeProcessFailed
	TypeProcessRestarted
	TypeProcessHealthy
	TypeProcessUnhealthy
	// Mesh events
	TypeMeshNodeUp
	TypeMeshNodeDown
	TypeMeshLeaderChanged
	TypeMeshTopologyChanged
	// Kubernetes events
	TypeK8sPodCreated
	TypeK8sPodDeleted
	TypeK8sPodReady
	TypeK8sPodFailed
	// System events
	TypeSystemHighCPU
	TypeSystemHighMemory
	TypeSystemDiskFull
	// Daemon events
	TypeDaemonStarted
	TypeDaemonStopping
	TypeDaemonConfigReloaded
)

// String returns the string representation of the event type.
func (t Type) String() string {
	switch t {
	case TypeUnknown:
		return unknownString
	case TypeProcessStarted:
		return "process.started"
	case TypeProcessStopped:
		return "process.stopped"
	case TypeProcessFailed:
		return "process.failed"
	case TypeProcessRestarted:
		return "process.restarted"
	case TypeProcessHealthy:
		return "process.healthy"
	case TypeProcessUnhealthy:
		return "process.unhealthy"
	case TypeMeshNodeUp:
		return "mesh.node.up"
	case TypeMeshNodeDown:
		return "mesh.node.down"
	case TypeMeshLeaderChanged:
		return "mesh.leader.changed"
	case TypeMeshTopologyChanged:
		return "mesh.topology.changed"
	case TypeK8sPodCreated:
		return "k8s.pod.created"
	case TypeK8sPodDeleted:
		return "k8s.pod.deleted"
	case TypeK8sPodReady:
		return "k8s.pod.ready"
	case TypeK8sPodFailed:
		return "k8s.pod.failed"
	case TypeSystemHighCPU:
		return "system.cpu.high"
	case TypeSystemHighMemory:
		return "system.memory.high"
	case TypeSystemDiskFull:
		return "system.disk.full"
	case TypeDaemonStarted:
		return "daemon.started"
	case TypeDaemonStopping:
		return "daemon.stopping"
	case TypeDaemonConfigReloaded:
		return "daemon.config.reloaded"
	}
	return unknownString
}

// Category returns the category of this event type.
func (t Type) Category() string {
	switch {
	case t >= TypeProcessStarted && t <= TypeProcessUnhealthy:
		return "process"
	case t >= TypeMeshNodeUp && t <= TypeMeshTopologyChanged:
		return "mesh"
	case t >= TypeK8sPodCreated && t <= TypeK8sPodFailed:
		return "kubernetes"
	case t >= TypeSystemHighCPU && t <= TypeSystemDiskFull:
		return "system"
	case t >= TypeDaemonStarted && t <= TypeDaemonConfigReloaded:
		return "daemon"
	default:
		return unknownString
	}
}

// Event represents a daemon event.
type Event struct {
	// ID is a unique event identifier.
	ID string `json:"id"`
	// Type is the event type.
	Type Type `json:"type"`
	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
	// ServiceName is the related service name (for process events).
	ServiceName string `json:"service_name,omitempty"`
	// NodeID is the related mesh node ID (for mesh events).
	NodeID string `json:"node_id,omitempty"`
	// PodName is the related pod name (for K8s events).
	PodName string `json:"pod_name,omitempty"`
	// Message is a human-readable event message.
	Message string `json:"message"`
	// Data contains additional event-specific data.
	Data map[string]any `json:"data,omitempty"`
}

// NewEvent creates a new event with the given type and message.
func NewEvent(t Type, message string) Event {
	return Event{
		Type:      t,
		Timestamp: time.Now(),
		Message:   message,
	}
}

// WithServiceName sets the service name for the event.
func (e Event) WithServiceName(name string) Event {
	e.ServiceName = name
	return e
}

// WithNodeID sets the node ID for the event.
func (e Event) WithNodeID(id string) Event {
	e.NodeID = id
	return e
}

// WithPodName sets the pod name for the event.
func (e Event) WithPodName(name string) Event {
	e.PodName = name
	return e
}

// WithData adds data to the event.
func (e Event) WithData(key string, value any) Event {
	if e.Data == nil {
		e.Data = make(map[string]any)
	}
	e.Data[key] = value
	return e
}
