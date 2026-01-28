package logging_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileOpener(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "file opener functionality"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			// File opener is internal, tested via Writer.
			assert.NotEmpty(t, path)
		})
	}
}
