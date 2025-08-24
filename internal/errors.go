package internal

import (
	"fmt"
	"strings"
	"time"
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
	// ErrorTypeRetryExhausted indicates all retry attempts have been exhausted
	ErrorTypeRetryExhausted
	// ErrorTypeCircuitOpen indicates circuit breaker is open
	ErrorTypeCircuitOpen
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
	case ErrorTypeRetryExhausted:
		return "RETRY_EXHAUSTED"
	case ErrorTypeCircuitOpen:
		return "CIRCUIT_OPEN"
	default:
		return "UNKNOWN"
	}
}

// IsRetryable returns true if this error type should trigger retry logic
func (e ErrorType) IsRetryable() bool {
	switch e {
	case ErrorTypeConnection, ErrorTypeTimeout, ErrorTypeCapacity:
		return true
	case ErrorTypeKeyInvalid, ErrorTypeNotFound, ErrorTypeSerialization, ErrorTypeValidation, ErrorTypeRetryExhausted, ErrorTypeCircuitOpen:
		return false
	default:
		return false
	}
}

// Severity returns the severity level of the error type
func (e ErrorType) Severity() ErrorSeverity {
	switch e {
	case ErrorTypeConnection, ErrorTypeTimeout, ErrorTypeCapacity, ErrorTypeCircuitOpen:
		return SeverityCritical
	case ErrorTypeRetryExhausted:
		return SeverityHigh
	case ErrorTypeSerialization, ErrorTypeValidation:
		return SeverityMedium
	case ErrorTypeKeyInvalid, ErrorTypeNotFound:
		return SeverityLow
	default:
		return SeverityMedium
	}
}

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	// SeverityLow indicates a low-severity error (e.g., validation, not found)
	SeverityLow ErrorSeverity = iota
	// SeverityMedium indicates a medium-severity error (e.g., serialization)
	SeverityMedium
	// SeverityHigh indicates a high-severity error (e.g., retry exhausted)
	SeverityHigh
	// SeverityCritical indicates a critical error (e.g., connection failure)
	SeverityCritical
)

// String returns the string representation of ErrorSeverity
func (s ErrorSeverity) String() string {
	switch s {
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ErrorContext provides additional context information for errors
type ErrorContext struct {
	Operation     string         `json:"operation,omitempty"`      // The operation that failed
	Timestamp     time.Time      `json:"timestamp"`                // When the error occurred
	AttemptNumber int            `json:"attempt_number,omitempty"` // Retry attempt number
	Duration      time.Duration  `json:"duration,omitempty"`       // How long the operation took
	Metadata      map[string]any `json:"metadata,omitempty"`       // Additional context data
}

// CacheError represents a cache-specific error with detailed context
type CacheError struct {
	Type     ErrorType     `json:"type"`              // Error type classification
	Key      string        `json:"key,omitempty"`     // Cache key involved in the error
	Message  string        `json:"message"`           // Human-readable error message
	Cause    error         `json:"-"`                 // Underlying error (not serialized)
	Context  *ErrorContext `json:"context,omitempty"` // Additional error context
	Severity ErrorSeverity `json:"severity"`          // Error severity level
}

// Error implements the error interface
func (e *CacheError) Error() string {
	var parts []string

	// Add severity and type
	parts = append(parts, fmt.Sprintf("[%s:%s]", e.Severity.String(), e.Type.String()))

	// Add key if present
	if e.Key != "" {
		parts = append(parts, fmt.Sprintf("key='%s'", e.Key))
	}

	// Add operation if present in context
	if e.Context != nil && e.Context.Operation != "" {
		parts = append(parts, fmt.Sprintf("op='%s'", e.Context.Operation))
	}

	// Add attempt number if this is a retry
	if e.Context != nil && e.Context.AttemptNumber > 0 {
		parts = append(parts, fmt.Sprintf("attempt=%d", e.Context.AttemptNumber))
	}

	// Add the main message
	errorMsg := fmt.Sprintf("cache error %s: %s", strings.Join(parts, " "), e.Message)

	// Add cause if present
	if e.Cause != nil {
		errorMsg += fmt.Sprintf(" (caused by: %v)", e.Cause)
	}

	return errorMsg
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

// NewCacheError creates a new CacheError with basic information
func NewCacheError(errType ErrorType, key, message string, cause error) *CacheError {
	return &CacheError{
		Type:     errType,
		Key:      key,
		Message:  message,
		Cause:    cause,
		Severity: errType.Severity(),
		Context: &ErrorContext{
			Timestamp: time.Now(),
		},
	}
}

// NewCacheErrorWithContext creates a new CacheError with detailed context
func NewCacheErrorWithContext(errType ErrorType, key, message string, cause error, context *ErrorContext) *CacheError {
	if context == nil {
		context = &ErrorContext{}
	}

	// Ensure timestamp is set
	if context.Timestamp.IsZero() {
		context.Timestamp = time.Now()
	}

	return &CacheError{
		Type:     errType,
		Key:      key,
		Message:  message,
		Cause:    cause,
		Severity: errType.Severity(),
		Context:  context,
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

// NewRetryExhaustedError creates a retry exhausted error
func NewRetryExhaustedError(operation string, attempts int, lastError error) *CacheError {
	context := &ErrorContext{
		Operation:     operation,
		AttemptNumber: attempts,
		Timestamp:     time.Now(),
	}

	message := fmt.Sprintf("operation '%s' failed after %d retry attempts", operation, attempts)
	return NewCacheErrorWithContext(ErrorTypeRetryExhausted, "", message, lastError, context)
}

// NewCircuitOpenError creates a circuit breaker open error
func NewCircuitOpenError(operation string) *CacheError {
	context := &ErrorContext{
		Operation: operation,
		Timestamp: time.Now(),
	}

	message := fmt.Sprintf("circuit breaker is open for operation '%s'", operation)
	return NewCacheErrorWithContext(ErrorTypeCircuitOpen, "", message, nil, context)
}

// WithContext adds context information to an existing CacheError
func (e *CacheError) WithContext(operation string, attemptNumber int, duration time.Duration) *CacheError {
	if e.Context == nil {
		e.Context = &ErrorContext{}
	}

	e.Context.Operation = operation
	e.Context.AttemptNumber = attemptNumber
	e.Context.Duration = duration

	return e
}

// WithMetadata adds metadata to an existing CacheError
func (e *CacheError) WithMetadata(key string, value any) *CacheError {
	if e.Context == nil {
		e.Context = &ErrorContext{}
	}

	if e.Context.Metadata == nil {
		e.Context.Metadata = make(map[string]any)
	}

	e.Context.Metadata[key] = value
	return e
}

// IsRetryable returns true if this error should trigger retry logic
func (e *CacheError) IsRetryable() bool {
	return e.Type.IsRetryable()
}

// GetRecoveryStrategy returns the recommended recovery strategy for this error
func (e *CacheError) GetRecoveryStrategy() RecoveryStrategy {
	switch e.Type {
	case ErrorTypeConnection:
		return RecoveryStrategyRetryWithBackoff
	case ErrorTypeTimeout:
		return RecoveryStrategyRetryWithBackoff
	case ErrorTypeCapacity:
		return RecoveryStrategyRetryWithDelay
	case ErrorTypeRetryExhausted:
		return RecoveryStrategyCircuitBreaker
	case ErrorTypeCircuitOpen:
		return RecoveryStrategyWaitAndRetry
	case ErrorTypeSerialization, ErrorTypeValidation, ErrorTypeKeyInvalid:
		return RecoveryStrategyFail
	case ErrorTypeNotFound:
		return RecoveryStrategyIgnore
	default:
		return RecoveryStrategyFail
	}
}

// RecoveryStrategy represents different error recovery approaches
type RecoveryStrategy int

const (
	// RecoveryStrategyFail indicates the error should not be recovered from
	RecoveryStrategyFail RecoveryStrategy = iota
	// RecoveryStrategyIgnore indicates the error can be safely ignored
	RecoveryStrategyIgnore
	// RecoveryStrategyRetryWithBackoff indicates retry with exponential backoff
	RecoveryStrategyRetryWithBackoff
	// RecoveryStrategyRetryWithDelay indicates retry with fixed delay
	RecoveryStrategyRetryWithDelay
	// RecoveryStrategyCircuitBreaker indicates circuit breaker should be activated
	RecoveryStrategyCircuitBreaker
	// RecoveryStrategyWaitAndRetry indicates wait for circuit breaker to close
	RecoveryStrategyWaitAndRetry
)

// String returns the string representation of RecoveryStrategy
func (r RecoveryStrategy) String() string {
	switch r {
	case RecoveryStrategyFail:
		return "FAIL"
	case RecoveryStrategyIgnore:
		return "IGNORE"
	case RecoveryStrategyRetryWithBackoff:
		return "RETRY_WITH_BACKOFF"
	case RecoveryStrategyRetryWithDelay:
		return "RETRY_WITH_DELAY"
	case RecoveryStrategyCircuitBreaker:
		return "CIRCUIT_BREAKER"
	case RecoveryStrategyWaitAndRetry:
		return "WAIT_AND_RETRY"
	default:
		return "UNKNOWN"
	}
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

// IsRetryExhaustedError checks if the error is a retry exhausted error
func IsRetryExhaustedError(err error) bool {
	if cacheErr, ok := err.(*CacheError); ok {
		return cacheErr.Type == ErrorTypeRetryExhausted
	}
	return false
}

// IsCircuitOpenError checks if the error is a circuit open error
func IsCircuitOpenError(err error) bool {
	if cacheErr, ok := err.(*CacheError); ok {
		return cacheErr.Type == ErrorTypeCircuitOpen
	}
	return false
}

// IsRetryableError checks if the error should trigger retry logic
func IsRetryableError(err error) bool {
	if cacheErr, ok := err.(*CacheError); ok {
		return cacheErr.IsRetryable()
	}
	return false
}

// GetErrorSeverity returns the severity level of an error
func GetErrorSeverity(err error) ErrorSeverity {
	if cacheErr, ok := err.(*CacheError); ok {
		return cacheErr.Severity
	}
	return SeverityMedium
}

// GetRecoveryStrategy returns the recommended recovery strategy for an error
func GetRecoveryStrategy(err error) RecoveryStrategy {
	if cacheErr, ok := err.(*CacheError); ok {
		return cacheErr.GetRecoveryStrategy()
	}
	return RecoveryStrategyFail
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	// CircuitBreakerClosed indicates normal operation
	CircuitBreakerClosed CircuitBreakerState = iota
	// CircuitBreakerOpen indicates circuit is open (failing fast)
	CircuitBreakerOpen
	// CircuitBreakerHalfOpen indicates circuit is testing if service recovered
	CircuitBreakerHalfOpen
)

// String returns the string representation of CircuitBreakerState
func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitBreakerClosed:
		return "CLOSED"
	case CircuitBreakerOpen:
		return "OPEN"
	case CircuitBreakerHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig defines circuit breaker behavior
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`   // Number of failures before opening
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`    // Time to wait before trying half-open
	SuccessThreshold int           `json:"success_threshold"`   // Successes needed to close from half-open
	MaxHalfOpenCalls int           `json:"max_half_open_calls"` // Max calls allowed in half-open state
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		RecoveryTimeout:  30 * time.Second,
		SuccessThreshold: 3,
		MaxHalfOpenCalls: 3,
	}
}

// CircuitBreaker implements the circuit breaker pattern for error recovery
type CircuitBreaker struct {
	config          *CircuitBreakerConfig
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	halfOpenCalls   int
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		config: config,
		state:  CircuitBreakerClosed,
	}
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() error {
	switch cb.state {
	case CircuitBreakerClosed:
		return nil
	case CircuitBreakerOpen:
		if time.Since(cb.lastFailureTime) >= cb.config.RecoveryTimeout {
			cb.state = CircuitBreakerHalfOpen
			cb.halfOpenCalls = 0
			cb.successCount = 0
			// Fall through to half-open logic
		} else {
			return NewCircuitOpenError("circuit breaker is open")
		}
		fallthrough
	case CircuitBreakerHalfOpen:
		if cb.halfOpenCalls >= cb.config.MaxHalfOpenCalls {
			return NewCircuitOpenError("circuit breaker half-open call limit exceeded")
		}
		cb.halfOpenCalls++
		return nil
	default:
		return NewCircuitOpenError("circuit breaker in unknown state")
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	switch cb.state {
	case CircuitBreakerClosed:
		cb.failureCount = 0
	case CircuitBreakerHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.state = CircuitBreakerClosed
			cb.failureCount = 0
			cb.successCount = 0
			cb.halfOpenCalls = 0
		}
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case CircuitBreakerClosed:
		cb.failureCount++
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.state = CircuitBreakerOpen
		}
	case CircuitBreakerHalfOpen:
		cb.state = CircuitBreakerOpen
		cb.halfOpenCalls = 0
		cb.successCount = 0
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return cb.state
}

// GetMetrics returns current circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]any {
	return map[string]any{
		"state":           cb.state.String(),
		"failure_count":   cb.failureCount,
		"success_count":   cb.successCount,
		"half_open_calls": cb.halfOpenCalls,
		"last_failure":    cb.lastFailureTime,
	}
}

// ErrorRecoveryManager manages error recovery strategies and circuit breakers
type ErrorRecoveryManager struct {
	circuitBreakers map[string]*CircuitBreaker
	config          *CircuitBreakerConfig
}

// NewErrorRecoveryManager creates a new error recovery manager
func NewErrorRecoveryManager(config *CircuitBreakerConfig) *ErrorRecoveryManager {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &ErrorRecoveryManager{
		circuitBreakers: make(map[string]*CircuitBreaker),
		config:          config,
	}
}

// GetCircuitBreaker returns or creates a circuit breaker for the given operation
func (erm *ErrorRecoveryManager) GetCircuitBreaker(operation string) *CircuitBreaker {
	if cb, exists := erm.circuitBreakers[operation]; exists {
		return cb
	}

	cb := NewCircuitBreaker(erm.config)
	erm.circuitBreakers[operation] = cb
	return cb
}

// ShouldRetry determines if an operation should be retried based on the error and recovery strategy
func (erm *ErrorRecoveryManager) ShouldRetry(err error, operation string, attemptNumber int) (bool, time.Duration) {
	if err == nil {
		return false, 0
	}

	// Check if it's a cache error with recovery strategy
	if cacheErr, ok := err.(*CacheError); ok {
		strategy := cacheErr.GetRecoveryStrategy()

		switch strategy {
		case RecoveryStrategyFail, RecoveryStrategyIgnore:
			return false, 0
		case RecoveryStrategyRetryWithBackoff:
			if attemptNumber >= 5 { // Max retry attempts
				return false, 0
			}
			delay := time.Duration(attemptNumber*attemptNumber) * 100 * time.Millisecond
			if delay > 5*time.Second {
				delay = 5 * time.Second
			}
			return true, delay
		case RecoveryStrategyRetryWithDelay:
			if attemptNumber >= 3 { // Max retry attempts for fixed delay
				return false, 0
			}
			return true, 1 * time.Second
		case RecoveryStrategyCircuitBreaker, RecoveryStrategyWaitAndRetry:
			return false, 0 // Circuit breaker handles this
		default:
			return false, 0
		}
	}

	// For non-cache errors, use basic retry logic
	if attemptNumber >= 3 {
		return false, 0
	}

	return true, time.Duration(attemptNumber) * 500 * time.Millisecond
}

// ExecuteWithRecovery executes an operation with error recovery mechanisms
func (erm *ErrorRecoveryManager) ExecuteWithRecovery(operation string, fn func() error) error {
	cb := erm.GetCircuitBreaker(operation)

	// Check if circuit breaker allows execution
	if err := cb.CanExecute(); err != nil {
		return err
	}

	startTime := time.Now()
	err := fn()
	duration := time.Since(startTime)

	if err != nil {
		cb.RecordFailure()

		// Enhance error with context if it's a CacheError
		if cacheErr, ok := err.(*CacheError); ok {
			cacheErr.WithContext(operation, 1, duration)
		}

		return err
	}

	cb.RecordSuccess()
	return nil
}

// GetCircuitBreakerMetrics returns metrics for all circuit breakers
func (erm *ErrorRecoveryManager) GetCircuitBreakerMetrics() map[string]map[string]any {
	metrics := make(map[string]map[string]any)

	for operation, cb := range erm.circuitBreakers {
		metrics[operation] = cb.GetMetrics()
	}

	return metrics
}
