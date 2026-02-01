package discovery_test

import (
	"testing"
	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
	"github.com/kodflow/daemon/internal/domain/target"
)

func TestFactory_WithPortScan(t *testing.T) {
	// Create discovery config with port scan enabled
	discoveryCfg := &config.DiscoveryConfig{
		PortScan: &config.PortScanDiscoveryConfig{
			Enabled: true,
			ExcludePorts: []int{22},
		},
	}
	
	// Create factory
	factory := discovery.NewFactory(discoveryCfg)
	
	// Get all discoverers
	discoverers := factory.CreateDiscoverers()
	
	// Should have at least one discoverer (port scan)
	if len(discoverers) == 0 {
		t.Error("Expected at least one discoverer")
	}
	
	// Find port scan discoverer
	var foundPortScan bool
	for _, d := range discoverers {
		if d.Type() == target.TypeCustom {
			foundPortScan = true
			break
		}
	}
	
	if !foundPortScan {
		t.Error("Port scan discoverer not found in factory output")
	}
}
