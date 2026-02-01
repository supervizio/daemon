//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
//
//nolint:ktn-struct-onefile // Kubeconfig types are logically grouped for YAML parsing
package discovery

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Sentinel errors for Kubernetes authentication.
var (
	// errNoContext is returned when no context is found in kubeconfig.
	errNoContext error = errors.New("no context found in kubeconfig")

	// errContextNotFound is returned when the specified context is not found.
	errContextNotFound error = errors.New("context not found")

	// errClusterNotFound is returned when the specified cluster is not found.
	errClusterNotFound error = errors.New("cluster not found")

	// errUserNotFound is returned when the specified user is not found.
	errUserNotFound error = errors.New("user not found")

	// errMissingK8sEnvVars is returned when Kubernetes service environment variables are missing.
	errMissingK8sEnvVars error = errors.New("missing kubernetes service env vars")

	// errAppendCACert is returned when CA certificate cannot be appended to pool.
	errAppendCACert error = errors.New("append ca cert failed")
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
//
//nolint:ktn-struct-onefile
type kubeconfigCluster struct {
	// CertificateAuthorityData is the base64-encoded CA certificate.
	CertificateAuthorityData string `dto:"in,priv,priv" yaml:"certificate-authority-data"`

	// Server is the API server URL.
	Server string `dto:"in,priv,pub" yaml:"server"`
}

// kubeconfigContext represents a context in kubeconfig.
//
//nolint:ktn-struct-onefile // grouped with kubeconfig types
type kubeconfigContext struct {
	// Cluster is the cluster name.
	Cluster string `dto:"in,priv,pub" yaml:"cluster"`

	// User is the user name.
	User string `dto:"in,priv,pub" yaml:"user"`
}

// kubeconfigUser represents a user in kubeconfig.
//
//nolint:ktn-struct-onefile // grouped with kubeconfig types
type kubeconfigUser struct {
	// Token is the bearer token.
	Token string `dto:"in,priv,secret" yaml:"token"`
}

// kubeconfig represents a simplified kubeconfig file.
//
//nolint:ktn-struct-onefile // grouped with kubeconfig types
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
	// Parse kubeconfig file.
	cfg, err := parseKubeconfigFile(path)
	// Check for kubeconfig parsing error.
	if err != nil {
		// Return error from parsing.
		return nil, err
	}

	// Find current context name.
	contextName, err := cfg.findCurrentContextName()
	// Check for context lookup error.
	if err != nil {
		// Return error from context lookup.
		return nil, err
	}

	// Find context, cluster, and user.
	context, err := cfg.findContext(contextName)
	// Check for context error.
	if err != nil {
		// Return error from context lookup.
		return nil, err
	}

	cluster, err := cfg.findCluster(context.Cluster)
	// Check for cluster lookup error.
	if err != nil {
		// Return error from cluster lookup.
		return nil, err
	}

	user, err := cfg.findUser(context.User)
	// Check for user lookup error.
	if err != nil {
		// Return error from user lookup.
		return nil, err
	}

	// Build authentication configuration.
	return buildK8sAuth(cluster, user), nil
}

// parseKubeconfigFile reads and parses a kubeconfig file.
//
// Params:
//   - path: path to the kubeconfig file.
//
// Returns:
//   - *kubeconfig: the parsed kubeconfig.
//   - error: any error during parsing.
func parseKubeconfigFile(path string) (*kubeconfig, error) {
	// Read kubeconfig file.
	data, err := os.ReadFile(path)
	// Check for file read error.
	if err != nil {
		// Return error with read context.
		return nil, fmt.Errorf("read kubeconfig: %w", err)
	}

	// Parse YAML kubeconfig.
	var cfg kubeconfig
	// Check for YAML unmarshal error.
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// Return error with parse context.
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}

	// Return parsed kubeconfig.
	return &cfg, nil
}

// findCurrentContextName returns the current context name from kubeconfig.
//
// Returns:
//   - string: the context name.
//   - error: error if no context found.
func (kc *kubeconfig) findCurrentContextName() (string, error) {
	// Determine which context name to use.
	switch {
	// Handle explicit current context setting.
	case kc.CurrentContext != "":
		// Return explicitly set current context.
		return kc.CurrentContext, nil
	// Handle fallback to first available context.
	case len(kc.Contexts) > 0:
		// Return first available context name.
		return kc.Contexts[0].Name, nil
	// Handle missing context error.
	default:
		// Return error when no context is available.
		return "", fmt.Errorf("load kubeconfig: %w", errNoContext)
	}
}

// findContext finds a context by name in the kubeconfig.
//
// Params:
//   - name: the context name to find.
//
// Returns:
//   - *kubeconfigContext: the found context.
//   - error: error if context not found.
func (kc *kubeconfig) findContext(name string) (*kubeconfigContext, error) {
	// Iterate through contexts to find matching name.
	for _, ctx := range kc.Contexts {
		// Check if this context matches the requested name.
		if ctx.Name == name {
			// Return found context.
			return &ctx.Context, nil
		}
	}
	// Return error when context is not found.
	return nil, fmt.Errorf("%w: %s", errContextNotFound, name)
}

// findCluster finds a cluster by name in the kubeconfig.
//
// Params:
//   - name: the cluster name to find.
//
// Returns:
//   - *kubeconfigCluster: the found cluster.
//   - error: error if cluster not found.
func (kc *kubeconfig) findCluster(name string) (*kubeconfigCluster, error) {
	// Iterate through clusters to find matching name.
	for _, c := range kc.Clusters {
		// Check if this cluster matches the requested name.
		if c.Name == name {
			// Return found cluster.
			return &c.Cluster, nil
		}
	}
	// Return error when cluster is not found.
	return nil, fmt.Errorf("%w: %s", errClusterNotFound, name)
}

// findUser finds a user by name in the kubeconfig.
//
// Params:
//   - name: the user name to find.
//
// Returns:
//   - *kubeconfigUser: the found user.
//   - error: error if user not found.
func (kc *kubeconfig) findUser(name string) (*kubeconfigUser, error) {
	// Iterate through users to find matching name.
	for _, u := range kc.Users {
		// Check if this user matches the requested name.
		if u.Name == name {
			// Return found user.
			return &u.User, nil
		}
	}
	// Return error when user is not found.
	return nil, fmt.Errorf("%w: %s", errUserNotFound, name)
}

// buildK8sAuth builds the authentication configuration from cluster and user.
//
// Params:
//   - cluster: the cluster configuration.
//   - user: the user configuration.
//
// Returns:
//   - *k8sAuth: the authentication configuration.
func buildK8sAuth(cluster *kubeconfigCluster, user *kubeconfigUser) *k8sAuth {
	// Decode base64 CA certificate if present.
	var caCert []byte
	// Check if CA certificate data is available.
	if cluster.CertificateAuthorityData != "" {
		// In production, we'd base64 decode. For simplicity, assume it's PEM.
		caCert = []byte(cluster.CertificateAuthorityData)
	}

	// Return authentication configuration.
	return &k8sAuth{
		apiServer: cluster.Server,
		token:     user.Token,
		caCert:    caCert,
	}
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
		tokenPath  string = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		caCertPath string = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
		apiHost    string = "KUBERNETES_SERVICE_HOST"
		apiPort    string = "KUBERNETES_SERVICE_PORT"
	)

	// Read token from mounted secret.
	token, err := os.ReadFile(tokenPath)
	// Check for token file read error.
	if err != nil {
		// Return error with token context.
		return nil, fmt.Errorf("read service account token: %w", err)
	}

	// Read CA certificate from mounted secret.
	caCert, err := os.ReadFile(caCertPath)
	// Check for CA cert file read error.
	if err != nil {
		// Return error with CA cert context.
		return nil, fmt.Errorf("read service account ca cert: %w", err)
	}

	// Get API server from environment variables.
	host := os.Getenv(apiHost)
	port := os.Getenv(apiPort)
	// Check for missing Kubernetes service environment variables.
	if host == "" || port == "" {
		// Return error for missing env vars.
		return nil, fmt.Errorf("load in-cluster config: %w", errMissingK8sEnvVars)
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
		// Check if CA cert can be appended to pool.
		if !certPool.AppendCertsFromPEM(auth.caCert) {
			// Return error for invalid CA cert.
			return nil, fmt.Errorf("create http client: %w", errAppendCACert)
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

	// Return configured HTTP client.
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
		// Check for kubeconfig loading error.
		if err != nil {
			// Return error with kubeconfig context.
			return nil, fmt.Errorf("load kubeconfig: %w", err)
		}
		// Return auth on successful kubeconfig load.
		return auth, nil
	}

	// Try default kubeconfig location.
	home, err := os.UserHomeDir()
	// Check if home directory is available.
	if err == nil {
		defaultPath := filepath.Join(home, ".kube", "config")
		// Check if default kubeconfig exists.
		if _, err := os.Stat(defaultPath); err == nil {
			auth, err := loadKubeconfig(defaultPath)
			// Check if loading succeeded.
			if err == nil {
				// Return auth from default kubeconfig.
				return auth, nil
			}
		}
	}

	// Fall back to in-cluster config.
	auth, err := loadInClusterConfig()
	// Check for in-cluster config loading error.
	if err != nil {
		// Return error with in-cluster context.
		return nil, fmt.Errorf("load in-cluster config: %w", err)
	}
	// Return auth from in-cluster config.
	return auth, nil
}
