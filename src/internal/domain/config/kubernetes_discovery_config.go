// Package config provides domain value objects for service configuration.
package config

// KubernetesDiscoveryConfig configures Kubernetes discovery.
// Kubernetes discovery monitors pods and services in a Kubernetes cluster.
type KubernetesDiscoveryConfig struct {
	// Enabled activates Kubernetes discovery.
	Enabled bool

	// KubeconfigPath is the path to kubeconfig file.
	// Empty uses in-cluster config or default kubeconfig.
	KubeconfigPath string

	// Namespaces limits discovery to specific namespaces.
	// Empty means all namespaces.
	Namespaces []string

	// LabelSelector filters resources by label selector.
	// Example: "app=nginx,version=v1".
	LabelSelector string
}
