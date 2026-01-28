package daemon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONMapPool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "get and return map from pool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pooled := jsonMapPool.Get()
			entry, ok := pooled.(map[string]any)
			assert.True(t, ok)
			assert.NotNil(t, entry)

			entry["test"] = "value"
			clear(entry)
			assert.Empty(t, entry)

			jsonMapPool.Put(entry)
		})
	}
}
