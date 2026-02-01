//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strings"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
)

// Kubernetes probe and pod phase constants.
const (
	// kubernetesProbeTypeTCP is the default probe type for K8s pods.
	kubernetesProbeTypeTCP string = "tcp"

	// kubernetesPodPhaseRunning is the pod phase for running pods.
	kubernetesPodPhaseRunning string = "Running"
)

// errUnexpectedK8sStatus is returned when Kubernetes API returns non-OK status.
var errUnexpectedK8sStatus error = errors.New("unexpected status code")

// KubernetesDiscoverer discovers Kubernetes pods via the Kubernetes API.
// It connects to the K8s API server and queries pods across namespaces.
// Pods are filtered by namespace and label selector.
type KubernetesDiscoverer struct {
	// auth contains API server URL, token, and CA cert.
	auth *k8sAuth

	// namespaces are the namespaces to query (empty means all).
	namespaces []string

	// labelSelector is the label selector filter (e.g., "app=nginx").
	labelSelector string

	// client is the HTTP client for K8s API requests.
	client *http.Client
}

// NewKubernetesDiscoverer creates a new Kubernetes discoverer.
// It loads authentication from kubeconfig or in-cluster config and creates an HTTP client.
//
// Params:
//   - cfg: the Kubernetes discovery configuration.
//
// Returns:
//   - *KubernetesDiscoverer: a new Kubernetes discoverer.
//   - error: any error during initialization.
func NewKubernetesDiscoverer(cfg *config.KubernetesDiscoveryConfig) (*KubernetesDiscoverer, error) {
	// Load authentication from kubeconfig or in-cluster.
	auth, err := loadKubeconfigOrInCluster(cfg.KubeconfigPath)
	// Check for authentication loading error.
	if err != nil {
		// Return error with authentication context.
		return nil, fmt.Errorf("load kubernetes auth: %w", err)
	}

	// Create HTTP client with TLS configuration.
	client, err := newHTTPClient(auth)
	// Check for HTTP client creation error.
	if err != nil {
		// Return error with client context.
		return nil, fmt.Errorf("create kubernetes http client: %w", err)
	}

	// Construct discoverer with auth and filters.
	return &KubernetesDiscoverer{
		auth:          auth,
		namespaces:    cfg.Namespaces,
		labelSelector: cfg.LabelSelector,
		client:        client,
	}, nil
}

// Type returns the target type for Kubernetes.
//
// Returns:
//   - target.Type: TypeKubernetes.
func (d *KubernetesDiscoverer) Type() target.Type {
	// Return Kubernetes type constant for this discoverer.
	return target.TypeKubernetes
}

// Discover finds all running Kubernetes pods matching the filters.
// It queries the K8s API across all configured namespaces and converts matching pods to ExternalTargets.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []target.ExternalTarget: the discovered pods.
//   - error: any error during discovery.
func (d *KubernetesDiscoverer) Discover(ctx context.Context) ([]target.ExternalTarget, error) {
	// Use all namespaces if none specified.
	namespaces := d.namespaces
	// Check if no namespaces were configured.
	if len(namespaces) == 0 {
		namespaces = []string{"default"}
	}

	// Collect all discovered targets across namespaces.
	var allTargets []target.ExternalTarget

	// Query each namespace sequentially.
	for _, ns := range namespaces {
		targets, err := d.discoverNamespace(ctx, ns)
		// Check for namespace discovery error.
		if err != nil {
			// Return error with namespace context.
			return nil, fmt.Errorf("discover namespace %s: %w", ns, err)
		}
		allTargets = append(allTargets, targets...)
	}

	// Return all discovered targets.
	return allTargets, nil
}

// discoverNamespace discovers pods in a single namespace.
//
// Params:
//   - ctx: context for cancellation.
//   - namespace: the namespace to query.
//
// Returns:
//   - []target.ExternalTarget: the discovered pods.
//   - error: any error during discovery.
func (d *KubernetesDiscoverer) discoverNamespace(ctx context.Context, namespace string) ([]target.ExternalTarget, error) {
	// Build API URL for pod list.
	apiURL := d.buildPodListURL(namespace)

	// Fetch pods from K8s API.
	podList, err := d.fetchPods(ctx, apiURL)
	// Check for pod fetch error.
	if err != nil {
		// Return error from pod fetch.
		return nil, err
	}

	// Convert pods to external targets.
	return d.filterAndConvertPods(podList), nil
}

// buildPodListURL constructs the API URL for listing pods in a namespace.
//
// Params:
//   - namespace: the namespace to query.
//
// Returns:
//   - string: the full API URL with optional label selector.
func (d *KubernetesDiscoverer) buildPodListURL(namespace string) string {
	// Build base API URL for pod list.
	apiURL := fmt.Sprintf("%s/api/v1/namespaces/%s/pods", d.auth.apiServer, namespace)

	// Add label selector query parameter if set.
	if d.labelSelector != "" {
		params := url.Values{}
		params.Add("labelSelector", d.labelSelector)
		apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())
	}

	// Return the constructed API URL.
	return apiURL
}

// fetchPods retrieves the pod list from the Kubernetes API.
//
// Params:
//   - ctx: context for cancellation.
//   - apiURL: the API endpoint URL.
//
// Returns:
//   - *k8sPodList: the retrieved pod list.
//   - error: any error during fetch.
func (d *KubernetesDiscoverer) fetchPods(ctx context.Context, apiURL string) (*k8sPodList, error) {
	// Build HTTP request with bearer token.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	// Check for request creation error.
	if err != nil {
		// Return error with request context.
		return nil, fmt.Errorf("create kubernetes request: %w", err)
	}

	// Add bearer token header.
	req.Header.Set("Authorization", "Bearer "+d.auth.token)
	req.Header.Set("Accept", "application/json")

	// Execute request against K8s API.
	resp, err := d.client.Do(req)
	// Check for API request error.
	if err != nil {
		// Return error with API context.
		return nil, fmt.Errorf("kubernetes api request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Verify successful response from K8s API.
	if resp.StatusCode != http.StatusOK {
		// Return error for unexpected status code.
		return nil, fmt.Errorf("kubernetes api: %w (status %d)", errUnexpectedK8sStatus, resp.StatusCode)
	}

	// Parse JSON response into pod list.
	var podList k8sPodList
	// Check for JSON decode error.
	if err := json.NewDecoder(resp.Body).Decode(&podList); err != nil {
		// Return error with decode context.
		return nil, fmt.Errorf("decode kubernetes response: %w", err)
	}

	// Return the parsed pod list.
	return &podList, nil
}

// filterAndConvertPods filters running pods with IPs and converts them to targets.
//
// Params:
//   - podList: the pod list to filter and convert.
//
// Returns:
//   - []target.ExternalTarget: the filtered and converted targets.
func (d *KubernetesDiscoverer) filterAndConvertPods(podList *k8sPodList) []target.ExternalTarget {
	var targets []target.ExternalTarget

	// Iterate through each pod in the list.
	for _, pod := range podList.Items {
		// Skip non-running pods.
		if pod.Status.Phase != kubernetesPodPhaseRunning {
			continue
		}

		// Skip pods without IP address.
		if pod.Status.PodIP == "" {
			continue
		}

		t := d.podToTarget(pod)
		targets = append(targets, t)
	}

	// Return filtered and converted targets.
	return targets
}

// podToTarget converts a Kubernetes pod to an ExternalTarget.
// It extracts metadata, configures probes from container ports, and sets default thresholds.
//
// Params:
//   - pod: the Kubernetes pod.
//
// Returns:
//   - target.ExternalTarget: the external target.
func (d *KubernetesDiscoverer) podToTarget(pod k8sPod) target.ExternalTarget {
	// Build unique ID from namespace and pod name.
	id := fmt.Sprintf("kubernetes:%s/%s", pod.Metadata.Namespace, pod.Metadata.Name)

	// Initialize target with K8s-specific configuration.
	t := target.ExternalTarget{
		ID:               id,
		Name:             pod.Metadata.Name,
		Type:             target.TypeKubernetes,
		Source:           target.SourceDiscovered,
		Labels:           make(map[string]string, len(pod.Metadata.Labels)+kubernetesMetadataLabels),
		Interval:         defaultProbeInterval,
		Timeout:          defaultProbeTimeout,
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
	}

	// Copy all pod labels to target labels.
	maps.Copy(t.Labels, pod.Metadata.Labels)

	// Add K8s-specific metadata as labels.
	t.Labels["kubernetes.namespace"] = pod.Metadata.Namespace
	t.Labels["kubernetes.pod"] = pod.Metadata.Name
	t.Labels["kubernetes.phase"] = pod.Status.Phase

	// Configure TCP probe based on container ports.
	d.configureProbe(&t, pod)

	// Return fully configured target with probe.
	return t
}

// configureProbe configures the probe for a pod based on its container ports.
// It uses the first TCP port found across all containers.
//
// Params:
//   - t: the target to configure.
//   - pod: the Kubernetes pod.
func (d *KubernetesDiscoverer) configureProbe(t *target.ExternalTarget, pod k8sPod) {
	// Find first TCP port across all containers.
	for _, container := range pod.Spec.Containers {
		// Iterate through ports for this container.
		for _, port := range container.Ports {
			// Check for TCP port (default if protocol not specified).
			protocol := strings.ToUpper(port.Protocol)
			// Configure probe if this is a TCP port.
			if protocol == "" || protocol == "TCP" {
				addr := fmt.Sprintf("%s:%d", pod.Status.PodIP, port.ContainerPort)
				t.ProbeType = kubernetesProbeTypeTCP
				t.ProbeTarget = health.NewTCPTarget(addr)
				// Return after configuring with first TCP port.
				return
			}
		}
	}

	// No ports found - leave probe unconfigured.
	// The target will be discovered but not health-checked.
}
