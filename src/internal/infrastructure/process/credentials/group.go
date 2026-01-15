// Package credentials provides credential management interfaces and types.
package credentials

// Group represents a system group.
// It contains identification information from the OS group database.
type Group struct {
	// GID is the numeric group identifier.
	GID uint32
	// Name is the name of the group.
	Name string
}
