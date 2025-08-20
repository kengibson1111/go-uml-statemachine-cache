package internal

import (
	"errors"
	"testing"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{"Connection error", ErrorTypeConnection, "CONNECTION"},
		{"Key invalid error", ErrorTypeKeyInvalid, "KEY_INVALID"},
		{"Not found error", ErrorTypeNotFound, "NOT_FOUND"},
		{"Serialization error", ErrorTypeSerialization, "SERIALIZATION"},
		{"Timeout error", ErrorTypeTimeout, "TIMEOUT"},
		{"Capacity error", ErrorTypeCapacity, "CAPACITY"},
		{"Validation error", ErrorTypeValidation, "VALIDATION"},
		{"Unknown error", ErrorType(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errType.String(); got != tt.expected {
				t.Errorf("ErrorType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCacheError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *CacheError
		expected string
	}{
		{
			name: "Error with key",
			err: &CacheError{
				Type:    ErrorTypeNotFound,
				Key:     "test-key",
				Message: "key not found",
			},
			expected: "cache error [NOT_FOUND] for key 'test-key': key not found",
		},
		{
			name: "Error without key",
			err: &CacheError{
				Type:    ErrorTypeConnection,
				Message: "connection failed",
			},
			expected: "cache error [CONNECTION]: connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("CacheError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCacheError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &CacheError{
		Type:    ErrorTypeConnection,
		Message: "connection failed",
		Cause:   cause,
	}

	if got := err.Unwrap(); got != cause {
		t.Errorf("CacheError.Unwrap() = %v, want %v", got, cause)
	}
}

func TestCacheError_Is(t *testing.T) {
	err1 := &CacheError{Type: ErrorTypeConnection}
	err2 := &CacheError{Type: ErrorTypeConnection}
	err3 := &CacheError{Type: ErrorTypeNotFound}
	otherErr := errors.New("other error")

	if !err1.Is(err2) {
		t.Error("Expected err1.Is(err2) to be true")
	}

	if err1.Is(err3) {
		t.Error("Expected err1.Is(err3) to be false")
	}

	if err1.Is(otherErr) {
		t.Error("Expected err1.Is(otherErr) to be false")
	}
}

func TestNewCacheError(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewCacheError(ErrorTypeConnection, "test-key", "test message", cause)

	if err.Type != ErrorTypeConnection {
		t.Errorf("Expected Type to be %v, got %v", ErrorTypeConnection, err.Type)
	}
	if err.Key != "test-key" {
		t.Errorf("Expected Key to be 'test-key', got '%v'", err.Key)
	}
	if err.Message != "test message" {
		t.Errorf("Expected Message to be 'test message', got '%v'", err.Message)
	}
	if err.Cause != cause {
		t.Errorf("Expected Cause to be %v, got %v", cause, err.Cause)
	}
}

func TestNewConnectionError(t *testing.T) {
	cause := errors.New("connection failed")
	err := NewConnectionError("Redis connection failed", cause)

	if err.Type != ErrorTypeConnection {
		t.Errorf("Expected Type to be %v, got %v", ErrorTypeConnection, err.Type)
	}
	if err.Key != "" {
		t.Errorf("Expected Key to be empty, got '%v'", err.Key)
	}
	if err.Message != "Redis connection failed" {
		t.Errorf("Expected Message to be 'Redis connection failed', got '%v'", err.Message)
	}
	if err.Cause != cause {
		t.Errorf("Expected Cause to be %v, got %v", cause, err.Cause)
	}
}

func TestNewKeyInvalidError(t *testing.T) {
	err := NewKeyInvalidError("invalid-key", "key contains invalid characters")

	if err.Type != ErrorTypeKeyInvalid {
		t.Errorf("Expected Type to be %v, got %v", ErrorTypeKeyInvalid, err.Type)
	}
	if err.Key != "invalid-key" {
		t.Errorf("Expected Key to be 'invalid-key', got '%v'", err.Key)
	}
	if err.Message != "key contains invalid characters" {
		t.Errorf("Expected Message to be 'key contains invalid characters', got '%v'", err.Message)
	}
	if err.Cause != nil {
		t.Errorf("Expected Cause to be nil, got %v", err.Cause)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("missing-key")

	if err.Type != ErrorTypeNotFound {
		t.Errorf("Expected Type to be %v, got %v", ErrorTypeNotFound, err.Type)
	}
	if err.Key != "missing-key" {
		t.Errorf("Expected Key to be 'missing-key', got '%v'", err.Key)
	}
	if err.Message != "key not found in cache" {
		t.Errorf("Expected Message to be 'key not found in cache', got '%v'", err.Message)
	}
	if err.Cause != nil {
		t.Errorf("Expected Cause to be nil, got %v", err.Cause)
	}
}

func TestIsConnectionError(t *testing.T) {
	connErr := NewConnectionError("connection failed", nil)
	notFoundErr := NewNotFoundError("key")
	otherErr := errors.New("other error")

	if !IsConnectionError(connErr) {
		t.Error("Expected IsConnectionError(connErr) to be true")
	}

	if IsConnectionError(notFoundErr) {
		t.Error("Expected IsConnectionError(notFoundErr) to be false")
	}

	if IsConnectionError(otherErr) {
		t.Error("Expected IsConnectionError(otherErr) to be false")
	}
}

func TestIsNotFoundError(t *testing.T) {
	notFoundErr := NewNotFoundError("key")
	connErr := NewConnectionError("connection failed", nil)
	otherErr := errors.New("other error")

	if !IsNotFoundError(notFoundErr) {
		t.Error("Expected IsNotFoundError(notFoundErr) to be true")
	}

	if IsNotFoundError(connErr) {
		t.Error("Expected IsNotFoundError(connErr) to be false")
	}

	if IsNotFoundError(otherErr) {
		t.Error("Expected IsNotFoundError(otherErr) to be false")
	}
}

func TestIsValidationError(t *testing.T) {
	validationErr := NewValidationError("validation failed", nil)
	connErr := NewConnectionError("connection failed", nil)
	otherErr := errors.New("other error")

	if !IsValidationError(validationErr) {
		t.Error("Expected IsValidationError(validationErr) to be true")
	}

	if IsValidationError(connErr) {
		t.Error("Expected IsValidationError(connErr) to be false")
	}

	if IsValidationError(otherErr) {
		t.Error("Expected IsValidationError(otherErr) to be false")
	}
}
