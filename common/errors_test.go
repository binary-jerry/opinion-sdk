package common

import (
	"errors"
	"testing"
)

func TestAPIErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name:     "with message",
			err:      &APIError{StatusCode: 400, Code: "BAD_REQUEST", Message: "Invalid input"},
			expected: "API error [400] BAD_REQUEST: Invalid input",
		},
		{
			name:     "without message",
			err:      &APIError{StatusCode: 500, Code: "INTERNAL_ERROR"},
			expected: "API error [500] INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("APIError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewAPIError(t *testing.T) {
	err := NewAPIError(404, "NOT_FOUND", "Resource not found")
	if err.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", err.StatusCode)
	}
	if err.Code != "NOT_FOUND" {
		t.Errorf("Code = %s, want NOT_FOUND", err.Code)
	}
	if err.Message != "Resource not found" {
		t.Errorf("Message = %s, want Resource not found", err.Message)
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrUnauthorized",
			err:      ErrUnauthorized,
			expected: true,
		},
		{
			name:     "API error 401",
			err:      &APIError{StatusCode: 401},
			expected: true,
		},
		{
			name:     "API error 403",
			err:      &APIError{StatusCode: 403},
			expected: true,
		},
		{
			name:     "API error 404",
			err:      &APIError{StatusCode: 404},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnauthorized(tt.err); got != tt.expected {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrNotFound",
			err:      ErrNotFound,
			expected: true,
		},
		{
			name:     "API error 404",
			err:      &APIError{StatusCode: 404},
			expected: true,
		},
		{
			name:     "API error 400",
			err:      &APIError{StatusCode: 400},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrRateLimited",
			err:      ErrRateLimited,
			expected: true,
		},
		{
			name:     "API error 429",
			err:      &APIError{StatusCode: 429},
			expected: true,
		},
		{
			name:     "API error 400",
			err:      &APIError{StatusCode: 400},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRateLimited(tt.err); got != tt.expected {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.expected)
			}
		})
	}
}
