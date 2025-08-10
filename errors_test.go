package gatussdk

import (
	"strings"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "error with body",
			err: &APIError{
				StatusCode: 404,
				Message:    "Not Found",
				Body:       `{"error": "endpoint not found"}`,
			},
			expected: `API error: status 404: Not Found (body: {"error": "endpoint not found"})`,
		},
		{
			name: "error without body",
			err: &APIError{
				StatusCode: 500,
				Message:    "Internal Server Error",
				Body:       "",
			},
			expected: "API error: status 500: Internal Server Error",
		},
		{
			name: "unauthorized error",
			err: &APIError{
				StatusCode: 401,
				Message:    "Unauthorized",
				Body:       `{"message": "invalid token"}`,
			},
			expected: `API error: status 401: Unauthorized (body: {"message": "invalid token"})`,
		},
		{
			name: "bad request error",
			err: &APIError{
				StatusCode: 400,
				Message:    "Bad Request",
				Body:       "malformed request",
			},
			expected: "API error: status 400: Bad Request (body: malformed request)",
		},
		{
			name: "rate limit error",
			err: &APIError{
				StatusCode: 429,
				Message:    "Too Many Requests",
				Body:       `{"retry_after": 60}`,
			},
			expected: `API error: status 429: Too Many Requests (body: {"retry_after": 60})`,
		},
		{
			name: "service unavailable",
			err: &APIError{
				StatusCode: 503,
				Message:    "Service Unavailable",
				Body:       "",
			},
			expected: "API error: status 503: Service Unavailable",
		},
		{
			name: "empty message with body",
			err: &APIError{
				StatusCode: 400,
				Message:    "",
				Body:       "error details",
			},
			expected: "API error: status 400:  (body: error details)",
		},
		{
			name: "multiline body",
			err: &APIError{
				StatusCode: 500,
				Message:    "Error",
				Body:       "line1\nline2\nline3",
			},
			expected: "API error: status 500: Error (body: line1\nline2\nline3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "field required error",
			err: &ValidationError{
				Field:   "key",
				Message: "cannot be empty",
			},
			expected: "validation error: field 'key': cannot be empty",
		},
		{
			name: "invalid duration",
			err: &ValidationError{
				Field:   "duration",
				Message: "must be one of: [1h 24h 7d 30d]",
			},
			expected: "validation error: field 'duration': must be one of: [1h 24h 7d 30d]",
		},
		{
			name: "invalid format",
			err: &ValidationError{
				Field:   "timestamp",
				Message: "invalid time format",
			},
			expected: "validation error: field 'timestamp': invalid time format",
		},
		{
			name: "field too long",
			err: &ValidationError{
				Field:   "name",
				Message: "exceeds maximum length of 255 characters",
			},
			expected: "validation error: field 'name': exceeds maximum length of 255 characters",
		},
		{
			name: "empty field name",
			err: &ValidationError{
				Field:   "",
				Message: "validation failed",
			},
			expected: "validation error: field '': validation failed",
		},
		{
			name: "empty message",
			err: &ValidationError{
				Field:   "value",
				Message: "",
			},
			expected: "validation error: field 'value': ",
		},
		{
			name: "special characters in field",
			err: &ValidationError{
				Field:   "user.email",
				Message: "invalid email format",
			},
			expected: "validation error: field 'user.email': invalid email format",
		},
		{
			name: "numeric validation",
			err: &ValidationError{
				Field:   "port",
				Message: "must be between 1 and 65535",
			},
			expected: "validation error: field 'port': must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("ValidationError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	t.Run("APIError implements error interface", func(t *testing.T) {
		var err error = &APIError{
			StatusCode: 404,
			Message:    "Not Found",
			Body:       "",
		}

		if _, ok := err.(*APIError); !ok {
			t.Error("APIError does not implement error interface properly")
		}

		if !strings.Contains(err.Error(), "404") {
			t.Error("APIError.Error() should contain status code")
		}
	})

	t.Run("ValidationError implements error interface", func(t *testing.T) {
		var err error = &ValidationError{
			Field:   "test",
			Message: "test message",
		}

		if _, ok := err.(*ValidationError); !ok {
			t.Error("ValidationError does not implement error interface properly")
		}

		if !strings.Contains(err.Error(), "test") {
			t.Error("ValidationError.Error() should contain field name")
		}
	})

	t.Run("errors can be compared", func(t *testing.T) {
		err1 := &APIError{StatusCode: 404, Message: "Not Found", Body: ""}
		err2 := &APIError{StatusCode: 404, Message: "Not Found", Body: ""}
		err3 := &APIError{StatusCode: 500, Message: "Server Error", Body: ""}

		if err1.Error() != err2.Error() {
			t.Error("identical APIErrors should produce identical error strings")
		}

		if err1.Error() == err3.Error() {
			t.Error("different APIErrors should produce different error strings")
		}
	})

	t.Run("nil checks", func(t *testing.T) {
		// Ensure methods don't panic on edge cases
		apiErr := &APIError{}
		_ = apiErr.Error()

		valErr := &ValidationError{}
		_ = valErr.Error()
	})
}
