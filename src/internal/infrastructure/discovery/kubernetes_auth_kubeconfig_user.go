//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
// This file defines the kubeconfigUser type for Kubernetes kubeconfig parsing.
package discovery

// kubeconfigUser represents a user in kubeconfig.
type kubeconfigUser struct {
	// Token is the bearer token.
	Token string `dto:"in,priv,secret" yaml:"token"`
}
