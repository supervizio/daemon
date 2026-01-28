// Package config_test provides black-box tests for the config package.
// It tests the public API of configuration types.
package config_test

import (
	"github.com/kodflow/daemon/internal/domain/config"
	"testing"

	"github.com/stretchr/testify/assert"

)

// TestDefaultDaemonLogging tests config.DefaultDaemonLogging function.
// It verifies that default daemon logging configuration is correctly initialized.
//
// Params:
//   - t: testing context
func TestDefaultDaemonLogging(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name                string
		expectedWriterCount int
		expectedWriterType  string
		expectedWriterLevel string
	}{
		{
			name:                "default configuration",
			expectedWriterCount: 1,
			expectedWriterType:  "console",
			expectedWriterLevel: "info",
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := config.DefaultDaemonLogging()

			// Verify writer count.
			assert.Len(t, result.Writers, testCase.expectedWriterCount)

			// Verify first writer configuration.
			if len(result.Writers) > 0 {
				assert.Equal(t, testCase.expectedWriterType, result.Writers[0].Type)
				assert.Equal(t, testCase.expectedWriterLevel, result.Writers[0].Level)
			}
		})
	}
}
