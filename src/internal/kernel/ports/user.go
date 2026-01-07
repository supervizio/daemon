// Package ports defines the interfaces for OS abstraction.
package ports

// User represents a system user.
// It contains identification and profile information from the OS user database.
type User struct {
	// UID is the numeric user identifier.
	UID uint32
	// GID is the numeric primary group identifier for this user.
	GID uint32
	// Username is the login name of the user.
	Username string
	// HomeDir is the path to the user's home directory.
	HomeDir string
}
