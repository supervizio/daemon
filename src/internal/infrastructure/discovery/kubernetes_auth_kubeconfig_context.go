//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
// This file defines the kubeconfigContext type for Kubernetes kubeconfig parsing.
package discovery

// kubeconfigContext represents a context in kubeconfig.
type kubeconfigContext struct {
	// Cluster is the cluster name.
	Cluster string `dto:"in,priv,pub" yaml:"cluster"`

	// User is the user name.
	User string `dto:"in,priv,pub" yaml:"user"`
}
