package internal

import (
	"errors"
	"testing"
	"time"
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
		{"Retry exhausted error", ErrorTypeRetryExhausted, "RETRY_EXHAUSTED"},
		{"Circuit open error", ErrorTypeCircuitOpen, "CIRCUIT_OPEN"},
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

func TestErrorType_IsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected bool
	}{
		{"Connection error is retryable", ErrorTypeConnection, true},
		{"Timeout error is retryable", ErrorTypeTimeout, true},
		{"Capacity error is retryable", ErrorTypeCapacity, true},
		{"Key invalid error is not retryable", ErrorTypeKeyInvalid, false},
		{"Not found error is not retryable", ErrorTypeNotFound, false},
		{"Serialization error is not retryable", ErrorTypeSerialization, false},
		{"Validation error is not retryable", ErrorTypeValidation, false},
		{"Retry exhausted error is not retryable", ErrorTypeRetryExhausted, false},
		{"Circuit open error is not retryable", ErrorTypeCircuitOpen, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errType.IsRetryable(); got != tt.expected {
				t.Errorf("ErrorType.IsRetryable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorType_Severity(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected ErrorSeverity
	}{
		{"Connection error is critical", ErrorTypeConnection, SeverityCritical},
		{"Timeout error is critical", ErrorTypeTimeout, SeverityCritical},
		{"Capacity error is critical", ErrorTypeCapacity, SeverityCritical},
		{"Circuit open error is critical", ErrorTypeCircuitOpen, SeverityCritical},
		{"Retry exhausted error is high", ErrorTypeRetryExhausted, SeverityHigh},
		{"Serialization error is medium", ErrorTypeSerialization, SeverityMedium},
		{"Validation error is medium", ErrorTypeValidation, SeverityMedium},
		{"Key invalid error is low", ErrorTypeKeyInvalid, SeverityLow},
		{"Not found error is low", ErrorTypeNotFound, SeverityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errType.Severity(); got != tt.expected {
				t.Errorf("ErrorType.Severity() = %v, want %v", got, tt.expected)
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
			name: "Error with key and context",
			err: &CacheError{
				Type:     ErrorTypeNotFound,
				Key:      "test-key",
				Message:  "key not found",
				Severity: SeverityLow,
				Context: &ErrorContext{
					Operation:     "get",
					AttemptNumber: 1,
				},
			},
			expected: "cache error [LOW:NOT_FOUND] key='test-key' op='get' attempt=1: key not found",
		},
		{
			name: "Error without key",
			err: &CacheError{
				Type:     ErrorTypeConnection,
				Message:  "connection failed",
				Severity: SeverityCritical,
				Context: &ErrorContext{
					Operation: "connect",
				},
			},
			expected: "cache error [CRITICAL:CONNECTION] op='connect': connection failed",
		},
		{
			name: "Error with cause",
			err: &CacheError{
				Type:     ErrorTypeTimeout,
				Key:      "timeout-key",
				Message:  "operation timed out",
				Severity: SeverityCritical,
				Cause:    errors.New("context deadline exceeded"),
			},
			expected: "cache error [CRITICAL:TIMEOUT] key='timeout-key': operation timed out (caused by: context deadline exceeded)",
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
func TestNewCacheErrorWithContext(t *testing.T) {
	cause := errors.New("underlying error")
	context := &ErrorContext{
		Operation:     "test-op",
		AttemptNumber: 2,
		Duration:      100 * time.Millisecond,
		Metadata:      map[string]any{"key": "value"},
	}

	err := NewCacheErrorWithContext(ErrorTypeConnection, "test-key", "test message", cause, context)

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
	if err.Severity != SeverityCritical {
		t.Errorf("Expected Severity to be %v, got %v", SeverityCritical, err.Severity)
	}
	if err.Context.Operation != "test-op" {
		t.Errorf("Expected Context.Operation to be 'test-op', got '%v'", err.Context.Operation)
	}
	if err.Context.AttemptNumber != 2 {
		t.Errorf("Expected Context.AttemptNumber to be 2, got %v", err.Context.AttemptNumber)
	}
}

func TestCacheError_WithContext(t *testing.T) {
	err := NewCacheError(ErrorTypeTimeout, "test-key", "timeout error", nil)

	err.WithContext("get-operation", 3, 500*time.Millisecond)

	if err.Context.Operation != "get-operation" {
		t.Errorf("Expected Context.Operation to be 'get-operation', got '%v'", err.Context.Operation)
	}
	if err.Context.AttemptNumber != 3 {
		t.Errorf("Expected Context.AttemptNumber to be 3, got %v", err.Context.AttemptNumber)
	}
	if err.Context.Duration != 500*time.Millisecond {
		t.Errorf("Expected Context.Duration to be 500ms, got %v", err.Context.Duration)
	}
}

func TestCacheError_WithMetadata(t *testing.T) {
	err := NewCacheError(ErrorTypeValidation, "", "validation error", nil)

	err.WithMetadata("field", "username").WithMetadata("value", "invalid@")

	if err.Context.Metadata["field"] != "username" {
		t.Errorf("Expected metadata 'field' to be 'username', got '%v'", err.Context.Metadata["field"])
	}
	if err.Context.Metadata["value"] != "invalid@" {
		t.Errorf("Expected metadata 'value' to be 'invalid@', got '%v'", err.Context.Metadata["value"])
	}
}

func TestCacheError_IsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected bool
	}{
		{"Connection error is retryable", ErrorTypeConnection, true},
		{"Timeout error is retryable", ErrorTypeTimeout, true},
		{"Validation error is not retryable", ErrorTypeValidation, false},
		{"Not found error is not retryable", ErrorTypeNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCacheError(tt.errType, "test-key", "test message", nil)
			if got := err.IsRetryable(); got != tt.expected {
				t.Errorf("CacheError.IsRetryable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCacheError_GetRecoveryStrategy(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected RecoveryStrategy
	}{
		{"Connection error uses retry with backoff", ErrorTypeConnection, RecoveryStrategyRetryWithBackoff},
		{"Timeout error uses retry with backoff", ErrorTypeTimeout, RecoveryStrategyRetryWithBackoff},
		{"Capacity error uses retry with delay", ErrorTypeCapacity, RecoveryStrategyRetryWithDelay},
		{"Retry exhausted uses circuit breaker", ErrorTypeRetryExhausted, RecoveryStrategyCircuitBreaker},
		{"Circuit open uses wait and retry", ErrorTypeCircuitOpen, RecoveryStrategyWaitAndRetry},
		{"Validation error fails", ErrorTypeValidation, RecoveryStrategyFail},
		{"Not found error is ignored", ErrorTypeNotFound, RecoveryStrategyIgnore},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCacheError(tt.errType, "test-key", "test message", nil)
			if got := err.GetRecoveryStrategy(); got != tt.expected {
				t.Errorf("CacheError.GetRecoveryStrategy() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewRetryExhaustedError(t *testing.T) {
	lastError := errors.New("connection failed")
	err := NewRetryExhaustedError("get-operation", 5, lastError)

	if err.Type != ErrorTypeRetryExhausted {
		t.Errorf("Expected Type to be %v, got %v", ErrorTypeRetryExhausted, err.Type)
	}
	if err.Context.Operation != "get-operation" {
		t.Errorf("Expected Context.Operation to be 'get-operation', got '%v'", err.Context.Operation)
	}
	if err.Context.AttemptNumber != 5 {
		t.Errorf("Expected Context.AttemptNumber to be 5, got %v", err.Context.AttemptNumber)
	}
	if err.Cause != lastError {
		t.Errorf("Expected Cause to be %v, got %v", lastError, err.Cause)
	}
}

func TestNewCircuitOpenError(t *testing.T) {
	err := NewCircuitOpenError("set-operation")

	if err.Type != ErrorTypeCircuitOpen {
		t.Errorf("Expected Type to be %v, got %v", ErrorTypeCircuitOpen, err.Type)
	}
	if err.Context.Operation != "set-operation" {
		t.Errorf("Expected Context.Operation to be 'set-operation', got '%v'", err.Context.Operation)
	}
}

func TestIsRetryExhaustedError(t *testing.T) {
	retryErr := NewRetryExhaustedError("test-op", 3, nil)
	connErr := NewConnectionError("connection failed", nil)
	otherErr := errors.New("other error")

	if !IsRetryExhaustedError(retryErr) {
		t.Error("Expected IsRetryExhaustedError(retryErr) to be true")
	}

	if IsRetryExhaustedError(connErr) {
		t.Error("Expected IsRetryExhaustedError(connErr) to be false")
	}

	if IsRetryExhaustedError(otherErr) {
		t.Error("Expected IsRetryExhaustedError(otherErr) to be false")
	}
}

func TestIsCircuitOpenError(t *testing.T) {
	circuitErr := NewCircuitOpenError("test-op")
	connErr := NewConnectionError("connection failed", nil)
	otherErr := errors.New("other error")

	if !IsCircuitOpenError(circuitErr) {
		t.Error("Expected IsCircuitOpenError(circuitErr) to be true")
	}

	if IsCircuitOpenError(connErr) {
		t.Error("Expected IsCircuitOpenError(connErr) to be false")
	}

	if IsCircuitOpenError(otherErr) {
		t.Error("Expected IsCircuitOpenError(otherErr) to be false")
	}
}

func TestCircuitBreaker_CanExecute(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  100 * time.Millisecond,
		SuccessThreshold: 2,
		MaxHalfOpenCalls: 2,
	}
	cb := NewCircuitBreaker(config)

	// Initially closed, should allow execution
	if err := cb.CanExecute(); err != nil {
		t.Errorf("Expected CanExecute() to return nil for closed circuit, got %v", err)
	}

	// Record failures to open circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Should be open now, should not allow execution
	if err := cb.CanExecute(); err == nil {
		t.Error("Expected CanExecute() to return error for open circuit")
	}

	// Wait for recovery timeout
	time.Sleep(150 * time.Millisecond)

	// Should be half-open now, should allow limited execution
	if err := cb.CanExecute(); err != nil {
		t.Errorf("Expected CanExecute() to return nil for half-open circuit, got %v", err)
	}
}

func TestCircuitBreaker_RecordSuccess(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  100 * time.Millisecond,
		SuccessThreshold: 2,
		MaxHalfOpenCalls: 2,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait for recovery timeout to enter half-open
	time.Sleep(150 * time.Millisecond)
	cb.CanExecute() // This transitions to half-open

	// Record successes to close circuit
	cb.RecordSuccess()
	cb.RecordSuccess()

	if cb.GetState() != CircuitBreakerClosed {
		t.Errorf("Expected circuit to be closed after successful operations, got %v", cb.GetState())
	}
}

func TestErrorRecoveryManager_ShouldRetry(t *testing.T) {
	erm := NewErrorRecoveryManager(nil)

	tests := []struct {
		name          string
		err           error
		operation     string
		attemptNumber int
		shouldRetry   bool
	}{
		{
			name:          "Connection error should retry",
			err:           NewConnectionError("connection failed", nil),
			operation:     "get",
			attemptNumber: 1,
			shouldRetry:   true,
		},
		{
			name:          "Validation error should not retry",
			err:           NewValidationError("invalid input", nil),
			operation:     "set",
			attemptNumber: 1,
			shouldRetry:   false,
		},
		{
			name:          "Too many attempts should not retry",
			err:           NewConnectionError("connection failed", nil),
			operation:     "get",
			attemptNumber: 6,
			shouldRetry:   false,
		},
		{
			name:          "No error should not retry",
			err:           nil,
			operation:     "get",
			attemptNumber: 1,
			shouldRetry:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldRetry, _ := erm.ShouldRetry(tt.err, tt.operation, tt.attemptNumber)
			if shouldRetry != tt.shouldRetry {
				t.Errorf("ShouldRetry() = %v, want %v", shouldRetry, tt.shouldRetry)
			}
		})
	}
}

func TestErrorRecoveryManager_ExecuteWithRecovery(t *testing.T) {
	erm := NewErrorRecoveryManager(nil)

	// Test successful execution
	err := erm.ExecuteWithRecovery("test-op", func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected successful execution to return nil, got %v", err)
	}

	// Test failed execution
	testErr := NewConnectionError("connection failed", nil)
	err = erm.ExecuteWithRecovery("test-op", func() error {
		return testErr
	})
	if err == nil {
		t.Error("Expected failed execution to return error")
	}

	// Verify circuit breaker recorded the failure
	cb := erm.GetCircuitBreaker("test-op")
	if cb.failureCount == 0 {
		t.Error("Expected circuit breaker to record failure")
	}
}
