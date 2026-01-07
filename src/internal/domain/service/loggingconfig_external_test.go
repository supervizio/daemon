// Package service_test provides black-box tests for LoggingConfig.
package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/service"
)

// TestDefaultLoggingConfig verifies the DefaultLoggingConfig function.
//
// Params:
//   - t: testing context for assertions.
func TestDefaultLoggingConfig(t *testing.T) {
	tests := []struct {
		name            string
		expectedBaseDir string
		expectedFormat  string
	}{
		{"default values", "/var/log/daemon", "iso8601"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := service.DefaultLoggingConfig()
			assert.Equal(t, tt.expectedBaseDir, cfg.BaseDir)
			assert.Equal(t, tt.expectedFormat, cfg.Defaults.TimestampFormat)
		})
	}
}
