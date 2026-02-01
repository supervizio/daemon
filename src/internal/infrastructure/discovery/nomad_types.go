//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery

// nomadAllocationList represents a list of allocations from Nomad API response.
type nomadAllocationList []nomadAllocation

// nomadAllocation represents an allocation from Nomad API response.
// This is an internal DTO for JSON unmarshaling from the Nomad HTTP API.
type nomadAllocation struct {
	// ID is the allocation unique identifier.
	ID string `json:"ID"`

	// Name is the allocation name (includes job name and index).
	Name string `json:"Name"`

	// Namespace is the Nomad namespace containing this allocation.
	Namespace string `json:"Namespace"`

	// JobID is the job that created this allocation.
	JobID string `json:"JobID"`

	// TaskGroup is the task group name within the job.
	TaskGroup string `json:"TaskGroup"`

	// TaskStates contains the state of each task in the allocation.
	TaskStates map[string]nomadTaskState `json:"TaskStates"`

	// ClientStatus is the allocation status on the client (running, pending, dead).
	ClientStatus string `json:"ClientStatus"`
}

// nomadTaskState represents the state of a task within an allocation.
type nomadTaskState struct {
	// State is the task state (running, pending, dead).
	State string `json:"State"`

	// Failed indicates if the task has failed.
	Failed bool `json:"Failed"`
}

// nomadAllocationDetail represents detailed allocation information.
// This includes resource allocations and port mappings.
type nomadAllocationDetail struct {
	// Resources contains the allocated resources for this allocation.
	Resources nomadResources `json:"Resources"`
}

// nomadResources represents the resources allocated to an allocation.
type nomadResources struct {
	// Networks contains the network configurations for the allocation.
	Networks []nomadNetwork `json:"Networks"`
}

// nomadNetwork represents a network configuration for an allocation.
type nomadNetwork struct {
	// IP is the IP address allocated to the allocation.
	IP string `json:"IP"`

	// DynamicPorts are dynamically allocated ports.
	DynamicPorts []nomadPort `json:"DynamicPorts"`

	// ReservedPorts are statically reserved ports.
	ReservedPorts []nomadPort `json:"ReservedPorts"`
}

// nomadPort represents a port mapping in Nomad.
type nomadPort struct {
	// Label is the port label (from job spec).
	Label string `json:"Label"`

	// Value is the actual port number allocated.
	Value int `json:"Value"`
}
