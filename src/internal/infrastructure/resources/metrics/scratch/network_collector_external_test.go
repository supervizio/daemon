//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

// Package scratch_test provides black-box tests for scratch metrics collectors.
package scratch_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/scratch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNetworkCollector_ListInterfaces tests interface listing.
//
// Params:
//   - t: the testing context
func TestNetworkCollector_ListInterfaces(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewNetworkCollector()
			require.NotNil(t, collector)

			_, err := collector.ListInterfaces(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestNetworkCollector_CollectStats tests interface stats collection.
//
// Params:
//   - t: the testing context
func TestNetworkCollector_CollectStats(t *testing.T) {
	tests := []struct {
		name  string
		iface string
	}{
		{name: "eth0 interface", iface: "eth0"},
		{name: "lo interface", iface: "lo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewNetworkCollector()

			_, err := collector.CollectStats(context.Background(), tt.iface)

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}

// TestNetworkCollector_CollectAllStats tests all interface stats collection.
//
// Params:
//   - t: the testing context
func TestNetworkCollector_CollectAllStats(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns ErrNotSupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := scratch.NewNetworkCollector()

			_, err := collector.CollectAllStats(context.Background())

			// Verify ErrNotSupported is returned
			assert.True(t, errors.Is(err, scratch.ErrNotSupported))
		})
	}
}
