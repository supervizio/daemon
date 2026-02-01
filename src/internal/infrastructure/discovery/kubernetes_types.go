//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery

// k8sPodList is the response from Kubernetes API for listing pods.
type k8sPodList struct {
	// Items contains the list of pods returned by the API.
	Items []k8sPod `dto:"out,api,pub" json:"items"`
}

// k8sPod represents a Kubernetes pod from the API.
type k8sPod struct {
	// Metadata contains pod metadata like name, namespace, labels.
	Metadata k8sMetadata `dto:"out,api,pub" json:"metadata"`

	// Spec contains the pod specification.
	Spec k8sPodSpec `dto:"out,api,pub" json:"spec"`

	// Status contains the current pod status.
	Status k8sPodStatus `dto:"out,api,pub" json:"status"`
}

// k8sMetadata contains pod metadata.
type k8sMetadata struct {
	// Name is the pod name.
	Name string `dto:"out,api,pub" json:"name"`

	// Namespace is the pod namespace.
	Namespace string `dto:"out,api,pub" json:"namespace"`

	// Labels are pod labels.
	Labels map[string]string `dto:"out,api,pub" json:"labels"`
}

// k8sPodSpec contains the pod specification.
type k8sPodSpec struct {
	// Containers are the containers in the pod.
	Containers []k8sContainer `dto:"out,api,pub" json:"containers"`
}

// k8sContainer represents a container in a pod.
type k8sContainer struct {
	// Name is the container name.
	Name string `dto:"out,api,pub" json:"name"`

	// Ports are the container ports.
	Ports []k8sPort `dto:"out,api,pub" json:"ports,omitempty"`
}

// k8sPort represents a container port.
type k8sPort struct {
	// ContainerPort is the port number.
	ContainerPort int `dto:"out,api,pub" json:"containerPort"`

	// Protocol is the port protocol (TCP or UDP).
	Protocol string `dto:"out,api,pub" json:"protocol,omitempty"`
}

// k8sPodStatus contains the current pod status.
type k8sPodStatus struct {
	// Phase is the pod lifecycle phase (Pending, Running, Succeeded, Failed, Unknown).
	Phase string `dto:"out,api,pub" json:"phase"`

	// PodIP is the pod IP address.
	PodIP string `dto:"out,api,pub" json:"podIP,omitempty"`
}
