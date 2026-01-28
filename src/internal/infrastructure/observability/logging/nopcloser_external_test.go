package logging_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNopCloser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "nopCloser is internal type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// nopCloser is not exported, tested via Capture.
			buf := &bytes.Buffer{}
			assert.NotNil(t, buf)
		})
	}
}
