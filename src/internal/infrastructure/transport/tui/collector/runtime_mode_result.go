// Package collector provides data collectors for TUI snapshot.
package collector

import (
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// runtimeModeResult is a private result struct for sync.OnceValue.
type runtimeModeResult struct {
	mode    model.RuntimeMode
	runtime string
}
