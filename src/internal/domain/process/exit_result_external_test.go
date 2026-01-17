// Package process_test provides black-box tests for the exit_result.go file.
// These tests validate the public API behavior of ExitResult without accessing internal state.
package process_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/process"
)

// TestExitResult_Fields validates ExitResult struct field access.
// It ensures the struct correctly holds exit code and error information.
//
// Params:
//   - t: the testing context
func TestExitResult_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   process.ExitResult
		wantCode int
		wantErr  bool
	}{
		{
			name: "successful exit with code 0",
			result: process.ExitResult{
				Code:  0,
				Error: nil,
			},
			wantCode: 0,
			wantErr:  false,
		},
		{
			name: "failed exit with code 1",
			result: process.ExitResult{
				Code:  1,
				Error: nil,
			},
			wantCode: 1,
			wantErr:  false,
		},
		{
			name: "failed exit with code 127 (command not found)",
			result: process.ExitResult{
				Code:  127,
				Error: nil,
			},
			wantCode: 127,
			wantErr:  false,
		},
		{
			name: "failed exit with code 255",
			result: process.ExitResult{
				Code:  255,
				Error: nil,
			},
			wantCode: 255,
			wantErr:  false,
		},
		{
			name: "negative exit code",
			result: process.ExitResult{
				Code:  -1,
				Error: nil,
			},
			wantCode: -1,
			wantErr:  false,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify exit code matches expected value.
			assert.Equal(t, tt.wantCode, tt.result.Code, "exit code should match")

			// Verify error field based on expectation.
			if tt.wantErr {
				assert.NotNil(t, tt.result.Error, "error should not be nil")
			} else {
				assert.Nil(t, tt.result.Error, "error should be nil")
			}
		})
	}
}

// TestExitResult_ZeroValue validates the zero value of ExitResult.
//
// Params:
//   - t: the testing context
func TestExitResult_ZeroValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "zero value has code 0 and nil error",
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create zero value ExitResult.
			var result process.ExitResult

			// Verify zero values.
			assert.Equal(t, 0, result.Code, "zero value code should be 0")
			assert.Nil(t, result.Error, "zero value error should be nil")
		})
	}
}
