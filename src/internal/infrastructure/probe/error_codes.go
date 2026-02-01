// Package probe provides CGO bindings to the Rust probe library for
// cross-platform system metrics collection.
package probe

// Go-only error code constants matching probe.h.
// These are duplicated from bindings.go to allow testing without CGO.
const (
	// probeCodeOK indicates success.
	probeCodeOK int = 0
	// probeCodeNotSupported indicates the operation is not supported.
	probeCodeNotSupported int = 1
	// probeCodePermission indicates permission denied.
	probeCodePermission int = 2
	// probeCodeNotFound indicates resource not found.
	probeCodeNotFound int = 3
	// probeCodeInvalidParam indicates invalid parameter.
	probeCodeInvalidParam int = 4
	// probeCodeIO indicates I/O error.
	probeCodeIO int = 5
	// probeCodeInternal indicates internal error.
	probeCodeInternal int = 99
)

// convertResultToError converts result fields to a Go error.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - success: whether the operation succeeded
//   - code: the error code if not successful
//   - message: the error message if available
//
// Returns:
//   - error: nil on success, appropriate error on failure
func convertResultToError(success bool, code int, message string) error {
	// Check if the result indicates success.
	if success {
		// Return nil for successful operations.
		return nil
	}

	// Map error code to Go error using the Go-only function.
	if err := mapProbeErrorCode(code); err != nil {
		// Return known error.
		return err
	}

	// Handle unknown error codes with message.
	if message != "" {
		// Build error with code and message.
		return newProbeError(code, message)
	}
	// Fallback to generic internal error.
	return ErrInternal
}

// bytesToStringWithNull converts a byte slice to a string, stopping at null.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - data: the byte slice to convert
//
// Returns:
//   - string: the converted string up to the first null byte
func bytesToStringWithNull(data []byte) string {
	// Start with full length as default.
	length := len(data)
	// Find null terminator by iterating bytes.
	for i, b := range data {
		// Check for null terminator.
		if b == 0 {
			length = i
			break
		}
	}
	// Return the string up to null terminator.
	return string(data[:length])
}
