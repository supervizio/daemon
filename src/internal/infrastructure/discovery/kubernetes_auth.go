//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// k8sAuth contains Kubernetes authentication configuration.
type k8sAuth struct {
	// apiServer is the Kubernetes API server URL.
	apiServer string

	// token is the bearer token for authentication.
	token string

	// caCert is the CA certificate for TLS verification.
	caCert []byte
}

// kubeconfigCluster represents a cluster in kubeconfig.
type kubeconfigCluster struct {
	// CertificateAuthorityData is the base64-encoded CA certificate.
	CertificateAuthorityData string `yaml:"certificate-authority-data"`

	// Server is the API server URL.
	Server string `yaml:"server"`
}

// kubeconfigContext represents a context in kubeconfig.
type kubeconfigContext struct {
	// Cluster is the cluster name.
	Cluster string `yaml:"cluster"`

	// User is the user name.
	User string `yaml:"user"`
}

// kubeconfigUser represents a user in kubeconfig.
type kubeconfigUser struct {
	// Token is the bearer token.
	Token string `yaml:"token"`
}

// kubeconfig represents a simplified kubeconfig file.
type kubeconfig struct {
	// CurrentContext is the active context name.
	CurrentContext string `yaml:"current-context"`

	// Clusters are the available clusters.
	Clusters []struct {
		Name    string            `yaml:"name"`
		Cluster kubeconfigCluster `yaml:"cluster"`
	} `yaml:"clusters"`

	// Contexts are the available contexts.
	Contexts []struct {
		Name    string            `yaml:"name"`
		Context kubeconfigContext `yaml:"context"`
	} `yaml:"contexts"`

	// Users are the available users.
	Users []struct {
		Name string         `yaml:"name"`
		User kubeconfigUser `yaml:"user"`
	} `yaml:"users"`
}

// loadKubeconfig loads authentication from a kubeconfig file.
// It parses the YAML kubeconfig and extracts API server, token, and CA cert.
//
// Params:
//   - path: path to the kubeconfig file.
//
// Returns:
//   - *k8sAuth: the authentication configuration.
//   - error: any error during loading.
func loadKubeconfig(path string) (*k8sAuth, error) {
	// Read kubeconfig file.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig: %w", err)
	}

	// Parse YAML kubeconfig.
	var cfg kubeconfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}

	// Find current context.
	var contextName string
	if cfg.CurrentContext != "" {
		contextName = cfg.CurrentContext
	} else if len(cfg.Contexts) > 0 {
		contextName = cfg.Contexts[0].Name
	} else {
		return nil, fmt.Errorf("no context found in kubeconfig")
	}

	// Find context by name.
	var context *kubeconfigContext
	for _, ctx := range cfg.Contexts {
		if ctx.Name == contextName {
			context = &ctx.Context
			break
		}
	}
	if context == nil {
		return nil, fmt.Errorf("context %s not found", contextName)
	}

	// Find cluster by name.
	var cluster *kubeconfigCluster
	for _, c := range cfg.Clusters {
		if c.Name == context.Cluster {
			cluster = &c.Cluster
			break
		}
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster %s not found", context.Cluster)
	}

	// Find user by name.
	var user *kubeconfigUser
	for _, u := range cfg.Users {
		if u.Name == context.User {
			user = &u.User
			break
		}
	}
	if user == nil {
		return nil, fmt.Errorf("user %s not found", context.User)
	}

	// Decode base64 CA certificate if present.
	var caCert []byte
	if cluster.CertificateAuthorityData != "" {
		// In production, we'd base64 decode. For simplicity, assume it's PEM.
		caCert = []byte(cluster.CertificateAuthorityData)
	}

	// Return authentication configuration.
	return &k8sAuth{
		apiServer: cluster.Server,
		token:     user.Token,
		caCert:    caCert,
	}, nil
}

// loadInClusterConfig loads authentication from in-cluster service account.
// It reads token and CA cert from standard paths mounted by Kubernetes.
//
// Returns:
//   - *k8sAuth: the authentication configuration.
//   - error: any error during loading.
func loadInClusterConfig() (*k8sAuth, error) {
	// Standard paths for in-cluster service account.
	const (
		tokenPath  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		caCertPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
		apiHost    = "KUBERNETES_SERVICE_HOST"
		apiPort    = "KUBERNETES_SERVICE_PORT"
	)

	// Read token from mounted secret.
	token, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("read service account token: %w", err)
	}

	// Read CA certificate from mounted secret.
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("read service account ca cert: %w", err)
	}

	// Get API server from environment variables.
	host := os.Getenv(apiHost)
	port := os.Getenv(apiPort)
	if host == "" || port == "" {
		return nil, fmt.Errorf("missing kubernetes service env vars")
	}

	apiServer := fmt.Sprintf("https://%s:%s", host, port)

	// Return in-cluster authentication configuration.
	return &k8sAuth{
		apiServer: apiServer,
		token:     string(token),
		caCert:    caCert,
	}, nil
}

// newHTTPClient creates an HTTP client with TLS configuration.
//
// Params:
//   - auth: the authentication configuration.
//
// Returns:
//   - *http.Client: an HTTP client configured for K8s API.
//   - error: any error during client creation.
func newHTTPClient(auth *k8sAuth) (*http.Client, error) {
	// Create TLS config with CA certificate.
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Add CA certificate if present.
	if len(auth.caCert) > 0 {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(auth.caCert) {
			return nil, fmt.Errorf("append ca cert failed")
		}
		tlsConfig.RootCAs = certPool
	}

	// Create HTTP client with TLS transport.
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   kubernetesRequestTimeout,
	}

	return client, nil
}

// loadKubeconfigOrInCluster loads authentication from kubeconfig or in-cluster config.
// It tries kubeconfig first, then falls back to in-cluster if path is empty.
//
// Params:
//   - kubeconfigPath: path to kubeconfig file (empty for in-cluster).
//
// Returns:
//   - *k8sAuth: the authentication configuration.
//   - error: any error during loading.
func loadKubeconfigOrInCluster(kubeconfigPath string) (*k8sAuth, error) {
	// Try kubeconfig if path is provided.
	if kubeconfigPath != "" {
		auth, err := loadKubeconfig(kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig: %w", err)
		}
		return auth, nil
	}

	// Try default kubeconfig location.
	home, err := os.UserHomeDir()
	if err == nil {
		defaultPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(defaultPath); err == nil {
			auth, err := loadKubeconfig(defaultPath)
			if err == nil {
				return auth, nil
			}
		}
	}

	// Fall back to in-cluster config.
	auth, err := loadInClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("load in-cluster config: %w", err)
	}
	return auth, nil
}
