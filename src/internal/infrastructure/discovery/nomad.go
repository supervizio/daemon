//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
)

// nomadProbeTypeTCP is the TCP probe type constant.
const nomadProbeTypeTCP string = "tcp"

// defaultNomadAddress is the default Nomad API address.
const defaultNomadAddress string = "http://localhost:4646"

// nomadRequestTimeout is the timeout for Nomad API requests.
const nomadRequestTimeout time.Duration = 10 * time.Second

// nomadMetadataLabels is the number of Nomad-specific metadata labels added to targets.
const nomadMetadataLabels int = 5

// allocIDDisplayLength is the truncated length for allocation IDs in names/labels.
const allocIDDisplayLength int = 8

// NomadDiscoverer discovers Nomad allocations via the Nomad HTTP API.
// It connects to the Nomad API and queries running allocations.
// Allocations are filtered by namespace and job pattern, then configured with TCP probes.
type NomadDiscoverer struct {
	// address is the Nomad API address.
	address string

	// namespace limits discovery to a specific namespace.
	namespace string

	// jobFilter filters allocations by job name pattern.
	jobFilter string

	// client is the HTTP client for Nomad API requests.
	client *http.Client
}

// NewNomadDiscoverer creates a new Nomad discoverer.
// It configures an HTTP client for Nomad API communication.
//
// Params:
//   - cfg: the Nomad discovery configuration.
//
// Returns:
//   - *NomadDiscoverer: a new Nomad discoverer.
func NewNomadDiscoverer(cfg *config.NomadDiscoveryConfig) *NomadDiscoverer {
	// Use default address if not specified in configuration.
	address := cfg.Address
	if address == "" {
		address = defaultNomadAddress
	}

	// Create HTTP client with timeout for API requests.
	client := &http.Client{
		Timeout: nomadRequestTimeout,
	}

	// Construct discoverer with API address and filters.
	return &NomadDiscoverer{
		address:   address,
		namespace: cfg.Namespace,
		jobFilter: cfg.JobFilter,
		client:    client,
	}
}

// Discover finds all running Nomad allocations matching the filters.
// It queries the Nomad HTTP API and converts matching allocations to ExternalTargets.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []target.ExternalTarget: the discovered allocations.
//   - error: any error during discovery.
func (d *NomadDiscoverer) Discover(ctx context.Context) ([]target.ExternalTarget, error) {
	// Build API endpoint for allocation list with namespace parameter.
	url := fmt.Sprintf("%s/v1/allocations", d.address)
	// Create request with context for cancellation support.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	// Check for request creation error.
	if err != nil {
		// Return error with request context.
		return nil, fmt.Errorf("create nomad request: %w", err)
	}

	// Add namespace query parameter if specified.
	if d.namespace != "" {
		q := req.URL.Query()
		q.Add("namespace", d.namespace)
		req.URL.RawQuery = q.Encode()
	}

	// Execute request against Nomad API.
	resp, err := d.client.Do(req)
	// Check for API communication error.
	if err != nil {
		// Return error with API context.
		return nil, fmt.Errorf("nomad api request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Verify successful response from Nomad API.
	if resp.StatusCode != http.StatusOK {
		// Return error for non-OK status.
		return nil, fmt.Errorf("nomad api returned status %d", resp.StatusCode)
	}

	// Parse JSON response into allocation structs.
	var allocations nomadAllocationList
	// Check for JSON decoding error.
	if err := json.NewDecoder(resp.Body).Decode(&allocations); err != nil {
		// Return error with decode context.
		return nil, fmt.Errorf("decode nomad response: %w", err)
	}

	// Convert matching allocations to external targets.
	var targets []target.ExternalTarget
	// Iterate over discovered allocations.
	for _, alloc := range allocations {
		// Skip allocations that don't match filters.
		if !d.matchesFilters(alloc) {
			continue
		}

		// Fetch detailed allocation info for port mappings.
		allocDetail, err := d.fetchAllocationDetail(ctx, alloc.ID)
		// Skip allocation on error (continue with others).
		if err != nil {
			continue
		}

		// Convert allocation to external targets (one per task).
		taskTargets := d.allocationToTargets(alloc, allocDetail)
		targets = append(targets, taskTargets...)
	}

	// Return all discovered and converted targets.
	return targets, nil
}

// Type returns the target type for Nomad.
//
// Returns:
//   - target.Type: TypeNomad.
func (d *NomadDiscoverer) Type() target.Type {
	// Return Nomad type constant for this discoverer.
	return target.TypeNomad
}

// matchesFilters checks if an allocation matches the configured filters.
// Returns true if allocation is running and matches job filter.
//
// Params:
//   - alloc: the allocation to check.
//
// Returns:
//   - bool: true if allocation matches filters.
func (d *NomadDiscoverer) matchesFilters(alloc nomadAllocation) bool {
	// Only include running allocations.
	if alloc.ClientStatus != "running" {
		return false
	}

	// Check job filter pattern if specified.
	if d.jobFilter != "" {
		// Simple prefix matching for job names.
		if !strings.HasPrefix(alloc.JobID, d.jobFilter) {
			return false
		}
	}

	// Allocation matches all filters.
	return true
}

// fetchAllocationDetail retrieves detailed allocation information from Nomad API.
// This includes resource allocations and port mappings.
//
// Params:
//   - ctx: context for cancellation.
//   - allocID: the allocation ID.
//
// Returns:
//   - *nomadAllocationDetail: the allocation details.
//   - error: any error during fetch.
func (d *NomadDiscoverer) fetchAllocationDetail(ctx context.Context, allocID string) (*nomadAllocationDetail, error) {
	// Build API endpoint for allocation detail.
	url := fmt.Sprintf("%s/v1/allocation/%s", d.address, allocID)
	// Create request with context for cancellation support.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	// Check for request creation error.
	if err != nil {
		// Return error with request context.
		return nil, fmt.Errorf("create detail request: %w", err)
	}

	// Execute request against Nomad API.
	resp, err := d.client.Do(req)
	// Check for API communication error.
	if err != nil {
		// Return error with API context.
		return nil, fmt.Errorf("fetch detail: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Verify successful response from Nomad API.
	if resp.StatusCode != http.StatusOK {
		// Return error for non-OK status.
		return nil, fmt.Errorf("detail api returned status %d", resp.StatusCode)
	}

	// Parse JSON response into allocation detail struct.
	var detail nomadAllocationDetail
	// Check for JSON decoding error.
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		// Return error with decode context.
		return nil, fmt.Errorf("decode detail response: %w", err)
	}

	// Return parsed allocation details.
	return &detail, nil
}

// allocationToTargets converts a Nomad allocation to ExternalTargets.
// It creates one target per running task with TCP probes for exposed ports.
//
// Params:
//   - alloc: the Nomad allocation.
//   - detail: the allocation detail with port mappings.
//
// Returns:
//   - []target.ExternalTarget: the external targets (one per task).
func (d *NomadDiscoverer) allocationToTargets(
	alloc nomadAllocation,
	detail *nomadAllocationDetail,
) []target.ExternalTarget {
	// Prepare targets slice for all running tasks.
	var targets []target.ExternalTarget

	// Iterate over tasks in the allocation.
	for taskName, taskState := range alloc.TaskStates {
		// Skip non-running tasks.
		if taskState.State != "running" {
			continue
		}

		// Create target for this task.
		t := d.taskToTarget(alloc, taskName, detail)
		targets = append(targets, t)
	}

	// Return all task targets.
	return targets
}

// taskToTarget converts a Nomad task to an ExternalTarget.
// It extracts metadata, configures probes from network ports, and sets default thresholds.
//
// Params:
//   - alloc: the Nomad allocation.
//   - taskName: the task name.
//   - detail: the allocation detail with port mappings.
//
// Returns:
//   - target.ExternalTarget: the external target.
func (d *NomadDiscoverer) taskToTarget(
	alloc nomadAllocation,
	taskName string,
	detail *nomadAllocationDetail,
) target.ExternalTarget {
	// Truncate allocation ID for display.
	allocIDShort := alloc.ID
	if len(allocIDShort) > allocIDDisplayLength {
		allocIDShort = allocIDShort[:allocIDDisplayLength]
	}

	// Initialize target with Nomad-specific configuration.
	t := target.ExternalTarget{
		ID:               fmt.Sprintf("nomad:%s/%s", allocIDShort, taskName),
		Name:             fmt.Sprintf("%s/%s", alloc.JobID, taskName),
		Type:             target.TypeNomad,
		Source:           target.SourceDiscovered,
		Labels:           make(map[string]string, nomadMetadataLabels),
		Interval:         defaultProbeInterval,
		Timeout:          defaultProbeTimeout,
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
	}

	// Add Nomad-specific metadata as labels.
	t.Labels["nomad.alloc_id"] = allocIDShort
	t.Labels["nomad.job"] = alloc.JobID
	t.Labels["nomad.task_group"] = alloc.TaskGroup
	t.Labels["nomad.task"] = taskName
	t.Labels["nomad.namespace"] = alloc.Namespace

	// Configure TCP probe based on network ports.
	d.configureProbe(&t, detail)

	// Return fully configured target with probe.
	return t
}

// configureProbe configures the probe for an allocation based on its ports.
// It uses the first available port (reserved or dynamic) for TCP probes.
//
// Params:
//   - t: the target to configure.
//   - detail: the allocation detail with port mappings.
func (d *NomadDiscoverer) configureProbe(t *target.ExternalTarget, detail *nomadAllocationDetail) {
	// Check if allocation has network configurations.
	if len(detail.Resources.Networks) == 0 {
		return
	}

	// Use first network configuration.
	network := detail.Resources.Networks[0]

	// Collect all available ports (reserved first, then dynamic).
	var ports []nomadPort
	ports = append(ports, network.ReservedPorts...)
	ports = append(ports, network.DynamicPorts...)

	// Configure probe with first available port.
	if len(ports) > 0 {
		port := ports[0]
		// Use network IP or fallback to localhost.
		ip := network.IP
		if ip == "" {
			ip = "127.0.0.1"
		}
		addr := fmt.Sprintf("%s:%d", ip, port.Value)
		t.ProbeType = nomadProbeTypeTCP
		t.ProbeTarget = health.NewTCPTarget(addr)

		// Add port label to target labels.
		t.Labels["nomad.port_label"] = port.Label
	}
}
