// Package shared provides common domain types used across multiple domain packages.
package shared

import "errors"

// Error variables for domain operations.
var (
	// ErrNotFound indicates a requested resource was not found.
	// This error is returned when a lookup operation fails to find the target.
	ErrNotFound error = errors.New("not found")

	// ErrAlreadyExists indicates a resource already exists.
	// This error is returned when attempting to create a duplicate resource.
	ErrAlreadyExists error = errors.New("already exists")

	// ErrInvalidState indicates an invalid state transition.
	// This error is returned when an operation is not valid for the current state.
	ErrInvalidState error = errors.New("invalid state")

	// ErrInvalidArgument indicates an invalid argument was provided.
	// This error is returned when a function receives an argument that is not valid.
	ErrInvalidArgument error = errors.New("invalid argument")

	// ErrEmptyCommand indicates the command configuration is empty.
	// This error is returned when a command is required but not provided.
	ErrEmptyCommand error = errors.New("empty command")
)
