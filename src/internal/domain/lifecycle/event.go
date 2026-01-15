// Package lifecycle provides domain types for daemon lifecycle management.
package lifecycle

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

// typeStrings maps event types to their string representations.
var typeStrings = map[Type]string{
	TypeUnknown:              unknownString,
	TypeProcessStarted:       "process.started",
	TypeProcessStopped:       "process.stopped",
	TypeProcessFailed:        "process.failed",
	TypeProcessRestarted:     "process.restarted",
	TypeProcessHealthy:       "process.healthy",
	TypeProcessUnhealthy:     "process.unhealthy",
	TypeMeshNodeUp:           "mesh.node.up",
	TypeMeshNodeDown:         "mesh.node.down",
	TypeMeshLeaderChanged:    "mesh.leader.changed",
	TypeMeshTopologyChanged:  "mesh.topology.changed",
	TypeK8sPodCreated:        "k8s.pod.created",
	TypeK8sPodDeleted:        "k8s.pod.deleted",
	TypeK8sPodReady:          "k8s.pod.ready",
	TypeK8sPodFailed:         "k8s.pod.failed",
	TypeSystemHighCPU:        "system.cpu.high",
	TypeSystemHighMemory:     "system.memory.high",
	TypeSystemDiskFull:       "system.disk.full",
	TypeDaemonStarted:        "daemon.started",
	TypeDaemonStopping:       "daemon.stopping",
	TypeDaemonConfigReloaded: "daemon.config.reloaded",
}

// String returns the string representation of the event type.
func (t Type) String() string {
	if s, ok := typeStrings[t]; ok {
		return s
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
