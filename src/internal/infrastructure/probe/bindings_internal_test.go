//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckInitialized(t *testing.T) {
	tests := []struct {
		name           string
		setInitialized bool
		wantErr        bool
	}{
		{
			name:           "ReturnsErrorWhenNotInitialized",
			setInitialized: false,
			wantErr:        true,
		},
		{
			name:           "ReturnsNilWhenInitialized",
			setInitialized: true,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set initialization state
			initMu.Lock()
			oldValue := initialized
			initialized = tt.setInitialized
			initMu.Unlock()

			// Restore after test
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			err := checkInitialized()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrNotInitialized)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
