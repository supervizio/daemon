// Package health provides the application service for health monitoring.
package health

import "errors"

var (
	// ErrProberFactoryMissing indicates that a prober factory was not configured.
	// This error is returned when AddListener is called with a listener that has
	// probe configuration, but no factory was provided to create the prober.
	ErrProberFactoryMissing error = errors.New("prober factory is not configured")
)
