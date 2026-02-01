//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
// This file defines the kubeconfigCluster type for Kubernetes kubeconfig parsing.
package discovery

// kubeconfigCluster represents a cluster in kubeconfig.
type kubeconfigCluster struct {
	// CertificateAuthorityData is the base64-encoded CA certificate.
	CertificateAuthorityData string `dto:"in,priv,priv" yaml:"certificate-authority-data"`

	// Server is the API server URL.
	Server string `dto:"in,priv,pub" yaml:"server"`
}
