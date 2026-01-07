// Package ports defines the interfaces for OS abstraction.
package ports

// Group represents a system group.
// It contains identification information from the OS group database.
type Group struct {
	// GID is the numeric group identifier.
	GID uint32
	// Name is the name of the group.
	Name string
}
