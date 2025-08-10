package gatussdk

import (
	"fmt"
)

// APIError represents an error returned by the Gatus API.
type APIError struct {
	// StatusCode is the HTTP status code returned by the API.
	StatusCode int
	// Message is a human-readable error message.
	Message string
	// Body contains the raw response body from the API.
	Body string
}

// Error returns a formatted error message.
func (e *APIError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("API error: status %d: %s (body: %s)", e.StatusCode, e.Message, e.Body)
	}
	return fmt.Sprintf("API error: status %d: %s", e.StatusCode, e.Message)
}

// ValidationError represents a validation error for input parameters.
type ValidationError struct {
	// Field is the name of the field that failed validation.
	Field string
	// Message is a human-readable error message.
	Message string
}

// Error returns a formatted validation error message.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: field '%s': %s", e.Field, e.Message)
}
