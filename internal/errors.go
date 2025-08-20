package internal

import (
	"fmt"
)

// ErrorType represents the type of cache error
type ErrorType int

const (
	// ErrorTypeConnection indicates a Redis connection error
	ErrorTypeConnection ErrorType = iota
	// ErrorTypeKeyInvalid indicates an invalid cache key
	ErrorTypeKeyInvalid
	// ErrorTypeNotFound indicates a cache miss or key not found
	ErrorTypeNotFound
	// ErrorTypeSerialization indicates JSON marshaling/unmarshaling error
	ErrorTypeSerialization
	// ErrorTypeTimeout indicates a timeout during cache operation
	ErrorTypeTimeout
	// ErrorTypeCapacity indicates cache capacity or memory issues
	ErrorTypeCapacity
	// ErrorTypeValidation indicates input validation failure
	ErrorTypeValidation
)

// String returns the string representation of ErrorType
func (e ErrorType) String() string {
	switch e {
	case ErrorTypeConnection:
		return "CONNECTION"
	case ErrorTypeKeyInvalid:
		return "KEY_INVALID"
	case ErrorTypeNotFound:
		return "NOT_FOUND"
	case ErrorTypeSerialization:
		return "SERIALIZATION"
	case ErrorTypeTimeout:
		return "TIMEOUT"
	case ErrorTypeCapacity:
		return "CAPACITY"
	case ErrorTypeValidation:
		return "VALIDATION"
	default:
		return "UNKNOWN"
	}
}

// CacheError represents a cache-specific error with context
type CacheError struct {
	Type    ErrorType
	Key     string
	Message string
	Cause   error
}

// Error implements the error interface
func (e *CacheError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("cache error [%s] for key '%s': %s", e.Type.String(), e.Key, e.Message)
	}
	return fmt.Sprintf("cache error [%s]: %s", e.Type.String(), e.Message)
}

// Unwrap returns the underlying cause error
func (e *CacheError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target error type
func (e *CacheError) Is(target error) bool {
	if t, ok := target.(*CacheError); ok {
		return e.Type == t.Type
	}
	return false
}

// NewCacheError creates a new CacheError
func NewCacheError(errType ErrorType, key, message string, cause error) *CacheError {
	return &CacheError{
		Type:    errType,
		Key:     key,
		Message: message,
		Cause:   cause,
	}
}

// NewConnectionError creates a connection-specific cache error
func NewConnectionError(message string, cause error) *CacheError {
	return NewCacheError(ErrorTypeConnection, "", message, cause)
}

// NewKeyInvalidError creates a key validation error
func NewKeyInvalidError(key, message string) *CacheError {
	return NewCacheError(ErrorTypeKeyInvalid, key, message, nil)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(key string) *CacheError {
	return NewCacheError(ErrorTypeNotFound, key, "key not found in cache", nil)
}

// NewSerializationError creates a serialization error
func NewSerializationError(key, message string, cause error) *CacheError {
	return NewCacheError(ErrorTypeSerialization, key, message, cause)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(key, message string, cause error) *CacheError {
	return NewCacheError(ErrorTypeTimeout, key, message, cause)
}

// NewValidationError creates a validation error
func NewValidationError(message string, cause error) *CacheError {
	return NewCacheError(ErrorTypeValidation, "", message, cause)
}

// IsConnectionError checks if the error is a connection error
func IsConnectionError(err error) bool {
	if cacheErr, ok := err.(*CacheError); ok {
		return cacheErr.Type == ErrorTypeConnection
	}
	return false
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	if cacheErr, ok := err.(*CacheError); ok {
		return cacheErr.Type == ErrorTypeNotFound
	}
	return false
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	if cacheErr, ok := err.(*CacheError); ok {
		return cacheErr.Type == ErrorTypeValidation
	}
	return false
}
