//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery_test

import (
	"context"
	"fmt"
	"log"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
)

// ExampleNewKubernetesDiscoverer demonstrates basic Kubernetes discovery.
func ExampleNewKubernetesDiscoverer() {
	// Configure Kubernetes discovery.
	cfg := &config.KubernetesDiscoveryConfig{
		Enabled:        true,
		KubeconfigPath: "", // Uses ~/.kube/config or in-cluster
		Namespaces:     []string{"default"},
		LabelSelector:  "app=nginx",
	}

	// Create discoverer.
	discoverer, err := discovery.NewKubernetesDiscoverer(cfg)
	if err != nil {
		log.Fatalf("create discoverer: %v", err)
	}

	// Discover pods.
	targets, err := discoverer.Discover(context.Background())
	if err != nil {
		log.Fatalf("discover: %v", err)
	}

	// Print discovered targets.
	for _, target := range targets {
		fmt.Printf("Pod: %s (namespace=%s, ip=%s)\n",
			target.Name,
			target.Labels["kubernetes.namespace"],
			target.ProbeTarget.Address)
	}
}

// ExampleNewKubernetesDiscoverer_inCluster demonstrates in-cluster discovery.
func ExampleNewKubernetesDiscoverer_inCluster() {
	// In-cluster config - no kubeconfig needed.
	cfg := &config.KubernetesDiscoveryConfig{
		Enabled:    true,
		Namespaces: []string{"production", "staging"},
	}

	discoverer, err := discovery.NewKubernetesDiscoverer(cfg)
	if err != nil {
		log.Fatalf("create discoverer: %v", err)
	}

	targets, err := discoverer.Discover(context.Background())
	if err != nil {
		log.Fatalf("discover: %v", err)
	}

	fmt.Printf("Discovered %d pods across namespaces\n", len(targets))
}

// ExampleNewKubernetesDiscoverer_labelSelector demonstrates label filtering.
func ExampleNewKubernetesDiscoverer_labelSelector() {
	// Filter by multiple labels.
	cfg := &config.KubernetesDiscoveryConfig{
		Enabled:       true,
		Namespaces:    []string{"default"},
		LabelSelector: "app=web,tier=frontend,version=v2",
	}

	discoverer, err := discovery.NewKubernetesDiscoverer(cfg)
	if err != nil {
		log.Fatalf("create discoverer: %v", err)
	}

	targets, err := discoverer.Discover(context.Background())
	if err != nil {
		log.Fatalf("discover: %v", err)
	}

	for _, target := range targets {
		fmt.Printf("Found: %s with labels %v\n", target.Name, target.Labels)
	}
}

// ExampleKubernetesDiscoverer_Type demonstrates type checking.
func ExampleKubernetesDiscoverer_Type() {
	cfg := &config.KubernetesDiscoveryConfig{
		Enabled: true,
	}

	discoverer, _ := discovery.NewKubernetesDiscoverer(cfg)

	fmt.Printf("Discoverer type: %s\n", discoverer.Type())
	// Output: Discoverer type: kubernetes
}
