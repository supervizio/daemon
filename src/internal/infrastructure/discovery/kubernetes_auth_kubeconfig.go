//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
// This file defines the kubeconfig type for parsing Kubernetes kubeconfig files.
package discovery

// kubeconfig represents a simplified kubeconfig file.
type kubeconfig struct {
	// CurrentContext is the active context name.
	CurrentContext string `dto:"in,priv,pub" yaml:"current-context"`

	// Clusters are the available clusters.
	Clusters []struct {
		Name    string            `dto:"in,priv,pub" yaml:"name"`
		Cluster kubeconfigCluster `dto:"in,priv,priv" yaml:"cluster"`
	} `dto:"in,priv,priv" yaml:"clusters"`

	// Contexts are the available contexts.
	Contexts []struct {
		Name    string            `dto:"in,priv,pub" yaml:"name"`
		Context kubeconfigContext `dto:"in,priv,pub" yaml:"context"`
	} `dto:"in,priv,pub" yaml:"contexts"`

	// Users are the available users.
	Users []struct {
		Name string         `dto:"in,priv,pub" yaml:"name"`
		User kubeconfigUser `dto:"in,priv,secret" yaml:"user"`
	} `dto:"in,priv,secret" yaml:"users"`
}
