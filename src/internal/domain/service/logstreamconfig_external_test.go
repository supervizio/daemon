// Package service_test provides black-box tests for LogStreamConfig.
package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/service"
)

// TestLogStreamConfig_File verifies the File method returns the configured file path.
//
// Params:
//   - t: testing context for assertions.
func TestLogStreamConfig_File(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{"empty path", "", ""},
		{"absolute path", "/var/log/app.log", "/var/log/app.log"},
		{"relative path", "logs/app.log", "logs/app.log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := service.NewLogStreamConfig(tt.filePath)
			assert.Equal(t, tt.expected, cfg.File())
		})
	}
}

// TestLogStreamConfig_TimestampFormat verifies the TimestampFormat method.
//
// Params:
//   - t: testing context for assertions.
func TestLogStreamConfig_TimestampFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{"empty format", "", ""},
		{"iso8601 format", "iso8601", "iso8601"},
		{"rfc3339 format", "2006-01-02T15:04:05Z07:00", "2006-01-02T15:04:05Z07:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := service.LogStreamConfig{Format: tt.format}
			assert.Equal(t, tt.expected, cfg.TimestampFormat())
		})
	}
}

// TestLogStreamConfig_Rotation verifies the Rotation method returns configured rotation.
//
// Params:
//   - t: testing context for assertions.
func TestLogStreamConfig_Rotation(t *testing.T) {
	tests := []struct {
		name        string
		maxFiles    int
		expectedMax int
	}{
		{"default rotation", 0, 0},
		{"custom max files", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := service.LogStreamConfig{
				RotationConfig: service.RotationConfig{MaxFiles: tt.maxFiles},
			}
			assert.Equal(t, tt.expectedMax, cfg.Rotation().MaxFiles)
		})
	}
}

// TestNewLogStreamConfig verifies the NewLogStreamConfig constructor.
//
// Params:
//   - t: testing context for assertions.
func TestNewLogStreamConfig(t *testing.T) {
	// defaultMaxFiles is the expected default number of rotated log files.
	const defaultMaxFiles int = 10

	tests := []struct {
		name     string
		filePath string
	}{
		{"empty path", ""},
		{"valid path", "/var/log/test.log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := service.NewLogStreamConfig(tt.filePath)
			assert.Equal(t, tt.filePath, cfg.File())
			assert.Equal(t, defaultMaxFiles, cfg.Rotation().MaxFiles)
		})
	}
}
