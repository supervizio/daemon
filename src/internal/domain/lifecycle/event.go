// Package lifecycle provides domain types for daemon lifecycle management.
package lifecycle

import "time"

const (
	// unknownString is the string representation for unknown types.
	unknownString string = "unknown"

	// eventDataInitialCapacity is the initial capacity for event data maps.
	eventDataInitialCapacity int = 4
)

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

var (
	// typeStrings maps event types to their string representations.
	typeStrings map[Type]string = map[Type]string{
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

	// typeCategories maps event types to their category strings.
	typeCategories map[Type]string = map[Type]string{
		TypeUnknown:              unknownString,
		TypeProcessStarted:       "process",
		TypeProcessStopped:       "process",
		TypeProcessFailed:        "process",
		TypeProcessRestarted:     "process",
		TypeProcessHealthy:       "process",
		TypeProcessUnhealthy:     "process",
		TypeMeshNodeUp:           "mesh",
		TypeMeshNodeDown:         "mesh",
		TypeMeshLeaderChanged:    "mesh",
		TypeMeshTopologyChanged:  "mesh",
		TypeK8sPodCreated:        "kubernetes",
		TypeK8sPodDeleted:        "kubernetes",
		TypeK8sPodReady:          "kubernetes",
		TypeK8sPodFailed:         "kubernetes",
		TypeSystemHighCPU:        "system",
		TypeSystemHighMemory:     "system",
		TypeSystemDiskFull:       "system",
		TypeDaemonStarted:        "daemon",
		TypeDaemonStopping:       "daemon",
		TypeDaemonConfigReloaded: "daemon",
	}
)

// String returns the string representation of the event type.
//
// Returns:
//   - string: string representation of the type
func (t Type) String() string {
	// Check if type exists in map
	if s, ok := typeStrings[t]; ok {
		// Return mapped string
		return s
	}
	// Return unknown for unmapped types
	return unknownString
}

// Category returns the category of this event type.
//
// Returns:
//   - string: category name (process, mesh, kubernetes, system, daemon, or unknown)
func (t Type) Category() string {
	// Look up category in map
	if category, ok := typeCategories[t]; ok {
		// Return mapped category
		return category
	}
	// Return unknown for unmapped types
	return unknownString
}

// Event represents a daemon event.
//
// This is a value object containing all information about a lifecycle event,
// including metadata for process, mesh, and Kubernetes contexts.
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
//
// Params:
//   - t: the event type
//   - message: human-readable event description
//
// Returns:
//   - Event: newly created event with current timestamp
func NewEvent(t Type, message string) Event {
	// Create event with provided type and message
	return Event{
		Type:      t,
		Timestamp: time.Now(),
		Message:   message,
	}
}

// WithServiceName sets the service name for the event.
//
// Params:
//   - name: the service name to associate with this event
//
// Returns:
//   - Event: updated event with service name set
func (e Event) WithServiceName(name string) Event {
	// Set service name field
	e.ServiceName = name
	// Return modified event
	return e
}

// WithNodeID sets the node ID for the event.
//
// Params:
//   - id: the mesh node ID to associate with this event
//
// Returns:
//   - Event: updated event with node ID set
func (e Event) WithNodeID(id string) Event {
	// Set node ID field
	e.NodeID = id
	// Return modified event
	return e
}

// WithPodName sets the pod name for the event.
//
// Params:
//   - name: the Kubernetes pod name to associate with this event
//
// Returns:
//   - Event: updated event with pod name set
func (e Event) WithPodName(name string) Event {
	// Set pod name field
	e.PodName = name
	// Return modified event
	return e
}

// WithStringData adds a string value to the event data.
//
// Params:
//   - key: the data key
//   - value: the string value
//
// Returns:
//   - Event: updated event with additional data
func (e Event) WithStringData(key, value string) Event {
	// Initialize data map if nil
	if e.Data == nil {
		// Create map with initial capacity
		e.Data = make(map[string]any, eventDataInitialCapacity)
	}
	// Add key-value pair to data map
	e.Data[key] = value
	// Return modified event
	return e
}

// WithIntData adds an integer value to the event data.
//
// Params:
//   - key: the data key
//   - value: the integer value
//
// Returns:
//   - Event: updated event with additional data
func (e Event) WithIntData(key string, value int) Event {
	// Initialize data map if nil
	if e.Data == nil {
		// Create map with initial capacity
		e.Data = make(map[string]any, eventDataInitialCapacity)
	}
	// Add key-value pair to data map
	e.Data[key] = value
	// Return modified event
	return e
}

// WithBoolData adds a boolean value to the event data.
//
// Params:
//   - key: the data key
//   - value: the boolean value
//
// Returns:
//   - Event: updated event with additional data
func (e Event) WithBoolData(key string, value bool) Event {
	// Initialize data map if nil
	if e.Data == nil {
		// Create map with initial capacity
		e.Data = make(map[string]any, eventDataInitialCapacity)
	}
	// Add key-value pair to data map
	e.Data[key] = value
	// Return modified event
	return e
}
