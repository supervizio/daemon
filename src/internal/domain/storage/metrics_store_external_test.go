// Package storage provides domain interfaces for metrics persistence.
package storage_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/storage"
)

// TestDefaultStoreConfig verifies the default configuration values.
//
// Params:
//   - t: testing context for assertions
func TestDefaultStoreConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedPath  string
		expectedRet   time.Duration
		expectedPrune time.Duration
	}{
		{
			name:          "default values",
			expectedPath:  "/var/lib/supervizio/metrics.db",
			expectedRet:   24 * time.Hour,
			expectedPrune: time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call the function under test.
			cfg := storage.DefaultStoreConfig()

			// Verify all default values.
			assert.Equal(t, tt.expectedPath, cfg.Path, "default path should be set")
			assert.Equal(t, tt.expectedRet, cfg.Retention, "default retention should be 24 hours")
			assert.Equal(t, tt.expectedPrune, cfg.PruneInterval, "default prune interval should be 1 hour")
		})
	}
}

// TestStoreConfig_Fields verifies StoreConfig can be constructed with custom values.
//
// Params:
//   - t: testing context for assertions
func TestStoreConfig_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		cfg           storage.StoreConfig
		expectedPath  string
		expectedRet   time.Duration
		expectedPrune time.Duration
	}{
		{
			name: "custom configuration",
			cfg: storage.StoreConfig{
				Path:          "/custom/path/metrics.db",
				Retention:     48 * time.Hour,
				PruneInterval: 2 * time.Hour,
			},
			expectedPath:  "/custom/path/metrics.db",
			expectedRet:   48 * time.Hour,
			expectedPrune: 2 * time.Hour,
		},
		{
			name: "minimal retention",
			cfg: storage.StoreConfig{
				Path:          "/tmp/metrics.db",
				Retention:     time.Hour,
				PruneInterval: 15 * time.Minute,
			},
			expectedPath:  "/tmp/metrics.db",
			expectedRet:   time.Hour,
			expectedPrune: 15 * time.Minute,
		},
		{
			name: "zero values allowed",
			cfg: storage.StoreConfig{
				Path:          "",
				Retention:     0,
				PruneInterval: 0,
			},
			expectedPath:  "",
			expectedRet:   0,
			expectedPrune: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify all fields are set correctly.
			assert.Equal(t, tt.expectedPath, tt.cfg.Path, "path should match")
			assert.Equal(t, tt.expectedRet, tt.cfg.Retention, "retention should match")
			assert.Equal(t, tt.expectedPrune, tt.cfg.PruneInterval, "prune interval should match")
		})
	}
}
