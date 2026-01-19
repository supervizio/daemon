// Package config provides the application port for configuration loading.
package config

import "github.com/kodflow/daemon/internal/domain/config"

// Loader loads configuration from external sources.
// This is the port that infrastructure adapters implement.
type Loader interface {
	// Load loads configuration from the given path.
	Load(path string) (*config.Config, error)
}

// Reloader can reload configuration at runtime.
type Reloader interface {
	// Reload reloads configuration from its original source.
	Reload() (*config.Config, error)
}
