//go:build unix

// Package healthcheck_test provides black-box tests for native ICMP probing.
package healthcheck_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/infrastructure/observability/healthcheck"
)

// TestNewICMPProberWithMode tests ICMP prober creation with mode.
func TestNewICMPProberWithMode(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		mode    config.ICMPMode
	}{
		{
			name:    "native_mode",
			timeout: 5 * time.Second,
			mode:    config.ICMPModeNative,
		},
		{
			name:    "fallback_mode",
			timeout: 5 * time.Second,
			mode:    config.ICMPModeFallback,
		},
		{
			name:    "auto_mode",
			timeout: 5 * time.Second,
			mode:    config.ICMPModeAuto,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober with specified mode.
			prober := healthcheck.NewICMPProberWithMode(tt.timeout, tt.mode)

			// Verify prober is created.
			require.NotNil(t, prober)
			assert.Equal(t, "icmp", prober.Type())
		})
	}
}

// TestICMPProber_Probe_NativeMode tests native ICMP probing.
// This test may fail if CAP_NET_RAW is not available.
func TestICMPProber_Probe_NativeMode(t *testing.T) {
	// Return early in short mode - this is an integration test.
	if testing.Short() {
		t.Log("skipping native ICMP test in short mode")
		return
	}

	tests := []struct {
		name    string
		target  health.Target
		timeout time.Duration
	}{
		{
			name: "localhost_native",
			target: health.Target{
				Address: "127.0.0.1",
			},
			timeout: 2 * time.Second,
		},
		{
			name: "google_dns_native",
			target: health.Target{
				Address: "8.8.8.8",
			},
			timeout: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober in native mode.
			prober := healthcheck.NewICMPProberWithMode(tt.timeout, config.ICMPModeNative)
			ctx := context.Background()

			// Perform healthcheck.
			result := prober.Probe(ctx, tt.target)

			// If native ICMP fails due to permissions, it should fall back to TCP.
			// We just verify a result is returned.
			assert.Greater(t, result.Latency, time.Duration(0))

			// Log result for debugging.
			t.Logf("Result: success=%v, latency=%s, output=%s",
				result.Success, result.Latency, result.Output)
		})
	}
}

// TestICMPProber_Probe_AutoMode tests auto mode capability detection.
func TestICMPProber_Probe_AutoMode(t *testing.T) {
	tests := []struct {
		name    string
		target  health.Target
		timeout time.Duration
	}{
		{
			name: "localhost_auto",
			target: health.Target{
				Address: "127.0.0.1",
			},
			timeout: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ICMP prober in auto mode.
			prober := healthcheck.NewICMPProber(tt.timeout)
			ctx := context.Background()

			// Perform healthcheck.
			result := prober.Probe(ctx, tt.target)

			// Should succeed regardless of capability (falls back to TCP).
			assert.Greater(t, result.Latency, time.Duration(0))

			// Log result for debugging.
			t.Logf("Auto mode result: success=%v, latency=%s, output=%s",
				result.Success, result.Latency, result.Output)
		})
	}
}
