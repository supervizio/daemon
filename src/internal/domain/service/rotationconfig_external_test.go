// Package service_test provides black-box tests for RotationConfig.
package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/service"
)

// TestDefaultRotationConfig verifies the DefaultRotationConfig function.
//
// Params:
//   - t: testing context for assertions.
func TestDefaultRotationConfig(t *testing.T) {
	// defaultMaxFiles is the expected default number of rotated log files.
	const defaultMaxFiles int = 10

	tests := []struct {
		name            string
		expectedMaxSize string
		expectedMaxFile int
	}{
		{"default values", "100MB", defaultMaxFiles},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := service.DefaultRotationConfig()
			assert.Equal(t, tt.expectedMaxSize, cfg.MaxSize)
			assert.Equal(t, tt.expectedMaxFile, cfg.MaxFiles)
		})
	}
}
