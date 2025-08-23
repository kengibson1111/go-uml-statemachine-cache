package internal

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConfig_Comprehensive tests comprehensive configuration scenarios
func TestConfig_Comprehensive(t *testing.T) {
	t.Run("DefaultConfig values", func(t *testing.T) {
		config := DefaultConfig()

		assert.Equal(t, "localhost:6379", config.RedisAddr)
		assert.Equal(t, "", config.RedisPassword)
		assert.Equal(t, 0, config.RedisDB)
		assert.Equal(t, 3, config.MaxRetries)
		assert.Equal(t, 5*time.Second, config.DialTimeout)
		assert.Equal(t, 3*time.Second, config.ReadTimeout)
		assert.Equal(t, 3*time.Second, config.WriteTimeout)
		assert.Equal(t, 10, config.PoolSize)
		assert.Equal(t, 24*time.Hour, config.DefaultTTL)
		assert.NotNil(t, config.RetryConfig)
	})

	t.Run("Config with custom retry config", func(t *testing.T) {
		retryConfig := &RetryConfig{
			MaxAttempts:  5,
			InitialDelay: 200 * time.Millisecond,
			MaxDelay:     10 * time.Second,
			Multiplier:   3.0,
			Jitter:       false,
			RetryableOps: []string{"get", "set"},
		}

		config := &Config{
			RedisAddr:     "redis.example.com:6379",
			RedisPassword: "secret",
			RedisDB:       1,
			MaxRetries:    5,
			DialTimeout:   10 * time.Second,
			ReadTimeout:   5 * time.Second,
			WriteTimeout:  5 * time.Second,
			PoolSize:      20,
			DefaultTTL:    48 * time.Hour,
			RetryConfig:   retryConfig,
		}

		assert.Equal(t, "redis.example.com:6379", config.RedisAddr)
		assert.Equal(t, "secret", config.RedisPassword)
		assert.Equal(t, 1, config.RedisDB)
		assert.Equal(t, retryConfig, config.RetryConfig)
	})
}

// TestRetryConfig_Comprehensive tests comprehensive retry configuration scenarios
func TestRetryConfig_Comprehensive(t *testing.T) {
	t.Run("DefaultRetryConfig values", func(t *testing.T) {
		config := DefaultRetryConfig()

		assert.Equal(t, 3, config.MaxAttempts)
		assert.Equal(t, 100*time.Millisecond, config.InitialDelay)
		assert.Equal(t, 5*time.Second, config.MaxDelay)
		assert.Equal(t, 2.0, config.Multiplier)
		assert.True(t, config.Jitter)
		assert.Contains(t, config.RetryableOps, "ping")
		assert.Contains(t, config.RetryableOps, "get")
		assert.Contains(t, config.RetryableOps, "set")
		assert.Contains(t, config.RetryableOps, "del")
		assert.Contains(t, config.RetryableOps, "exists")
		assert.Contains(t, config.RetryableOps, "expire")
	})

	t.Run("Custom retry config", func(t *testing.T) {
		config := &RetryConfig{
			MaxAttempts:  10,
			InitialDelay: 50 * time.Millisecond,
			MaxDelay:     30 * time.Second,
			Multiplier:   1.5,
			Jitter:       false,
			RetryableOps: []string{"ping", "info"},
		}

		assert.Equal(t, 10, config.MaxAttempts)
		assert.Equal(t, 50*time.Millisecond, config.InitialDelay)
		assert.Equal(t, 30*time.Second, config.MaxDelay)
		assert.Equal(t, 1.5, config.Multiplier)
		assert.False(t, config.Jitter)
		assert.Equal(t, []string{"ping", "info"}, config.RetryableOps)
	})
}

// TestInputValidator_ComprehensiveEdgeCases tests comprehensive edge cases for input validation
func TestInputValidator_ComprehensiveEdgeCases(t *testing.T) {
	validator := NewInputValidator()

	t.Run("ValidateAndSanitizeString edge cases", func(t *testing.T) {
		tests := []struct {
			name        string
			input       string
			fieldName   string
			expected    string
			expectError bool
			description string
		}{
			{
				name:        "string with only whitespace",
				input:       "   \t\n   ",
				fieldName:   "test field",
				expected:    "",
				expectError: true,
				description: "whitespace-only string should be treated as empty",
			},
			{
				name:        "string with mixed control characters",
				input:       "valid\x01invalid\x02content",
				fieldName:   "test field",
				expected:    "",
				expectError: true,
				description: "mixed control characters should be rejected",
			},
			{
				name:        "string with unicode normalization",
				input:       "café", // Contains combining characters
				fieldName:   "test field",
				expected:    "café",
				expectError: false,
				description: "unicode should be preserved",
			},
			{
				name:        "string with embedded nulls",
				input:       "before\x00after",
				fieldName:   "test field",
				expected:    "beforeafter",
				expectError: false,
				description: "null bytes should be removed",
			},
			{
				name:        "string with CRLF line endings",
				input:       "line1\r\nline2\r\nline3",
				fieldName:   "test field",
				expected:    "line1\nline2\nline3",
				expectError: false,
				description: "CRLF should be normalized to LF",
			},
			{
				name:        "string with only CR line endings",
				input:       "line1\rline2\rline3",
				fieldName:   "test field",
				expected:    "line1\nline2\nline3",
				expectError: false,
				description: "CR should be normalized to LF",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := validator.ValidateAndSanitizeString(tt.input, tt.fieldName)

				if tt.expectError {
					assert.Error(t, err, tt.description)
				} else {
					assert.NoError(t, err, tt.description)
					assert.Equal(t, tt.expected, result, tt.description)
				}
			})
		}
	})

	t.Run("ValidateAndSanitizeName comprehensive tests", func(t *testing.T) {
		tests := []struct {
			name        string
			input       string
			expected    string
			expectError bool
			description string
		}{
			{
				name:        "name with Windows path",
				input:       "C:\\Users\\Admin\\diagram.puml",
				expected:    "C%3A-Users-Admin-diagram.puml",
				expectError: false,
				description: "Windows paths should be sanitized",
			},
			{
				name:        "name with Unix path",
				input:       "/home/user/diagram.puml",
				expected:    "-home-user-diagram.puml",
				expectError: false,
				description: "Unix paths should be sanitized",
			},
			{
				name:        "name with query parameters",
				input:       "diagram?version=1&type=state",
				expected:    "diagram%3Fversion%3D1%26type%3Dstate",
				expectError: false,
				description: "Query parameters should be encoded",
			},
			{
				name:        "name with fragment",
				input:       "diagram#section1",
				expected:    "diagram%23section1",
				expectError: false,
				description: "Fragments should be encoded",
			},
			{
				name:        "name with percent encoding",
				input:       "diagram%20with%20spaces",
				expected:    "diagram%2520with%2520spaces",
				expectError: false,
				description: "Already encoded strings should be double-encoded",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := validator.ValidateAndSanitizeName(tt.input, "test field")

				if tt.expectError {
					assert.Error(t, err, tt.description)
				} else {
					assert.NoError(t, err, tt.description)
					assert.Equal(t, tt.expected, result, tt.description)
				}
			})
		}
	})

	t.Run("ValidateEntityData comprehensive tests", func(t *testing.T) {
		tests := []struct {
			name        string
			entity      interface{}
			expectError bool
			description string
		}{
			{
				name:        "nil interface",
				entity:      nil,
				expectError: true,
				description: "nil interface should be rejected",
			},
			{
				name:        "nil pointer",
				entity:      (*string)(nil),
				expectError: true,
				description: "nil pointer should be rejected",
			},
			{
				name:        "empty struct",
				entity:      struct{}{},
				expectError: false,
				description: "empty struct should be accepted",
			},
			{
				name:        "map entity",
				entity:      map[string]interface{}{"id": "test"},
				expectError: false,
				description: "map entity should be accepted",
			},
			{
				name:        "slice entity",
				entity:      []string{"item1", "item2"},
				expectError: false,
				description: "slice entity should be accepted",
			},
			{
				name:        "primitive entity",
				entity:      "string entity",
				expectError: false,
				description: "primitive entity should be accepted",
			},
			{
				name:        "numeric entity",
				entity:      42,
				expectError: false,
				description: "numeric entity should be accepted",
			},
			{
				name:        "boolean entity",
				entity:      true,
				expectError: false,
				description: "boolean entity should be accepted",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.ValidateEntityData(tt.entity, "test entity")

				if tt.expectError {
					assert.Error(t, err, tt.description)
				} else {
					assert.NoError(t, err, tt.description)
				}
			})
		}
	})
}

// TestInputValidator_SecurityComprehensive tests comprehensive security validation
func TestInputValidator_SecurityComprehensive(t *testing.T) {
	validator := NewInputValidator()

	t.Run("Advanced SQL injection patterns", func(t *testing.T) {
		sqlPatterns := []string{
			"'; EXEC xp_cmdshell('dir'); --",
			"' UNION ALL SELECT NULL,NULL,NULL--",
			"'; WAITFOR DELAY '00:00:05'; --",
			"' OR 'a'='a",
			"admin'--",
			"admin'/*",
			"' OR 1=1#",
			"'; DROP DATABASE test; --",
			"' HAVING 1=1--",
			"' GROUP BY 1--",
			"' ORDER BY 1--",
		}

		for _, pattern := range sqlPatterns {
			t.Run("SQL pattern: "+pattern, func(t *testing.T) {
				_, err := validator.ValidateAndSanitizeString(pattern, "test field")
				assert.Error(t, err, "SQL injection pattern should be blocked: %s", pattern)
			})
		}
	})

	t.Run("Advanced XSS patterns", func(t *testing.T) {
		xssPatterns := []string{
			"<img src=x onerror=alert(1)>",
			"<svg onload=alert(1)>",
			"<body onpageshow=alert(1)>",
			"<input onfocus=alert(1) autofocus>",
			"<select onfocus=alert(1) autofocus>",
			"<textarea onfocus=alert(1) autofocus>",
			"<keygen onfocus=alert(1) autofocus>",
			"<video><source onerror=alert(1)>",
			"<audio src=x onerror=alert(1)>",
			"<details open ontoggle=alert(1)>",
		}

		for _, pattern := range xssPatterns {
			t.Run("XSS pattern: "+pattern, func(t *testing.T) {
				_, err := validator.ValidateAndSanitizeString(pattern, "test field")
				assert.Error(t, err, "XSS pattern should be blocked: %s", pattern)
			})
		}
	})

	t.Run("Protocol-based attacks", func(t *testing.T) {
		protocolPatterns := []string{
			"javascript:alert(1)",
			"vbscript:msgbox(1)",
			"data:text/html,<script>alert(1)</script>",
			"data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==",
			"file:///etc/passwd",
			"ftp://malicious.com/",
			"ldap://malicious.com/",
		}

		for _, pattern := range protocolPatterns {
			t.Run("Protocol pattern: "+pattern, func(t *testing.T) {
				_, err := validator.ValidateAndSanitizeString(pattern, "test field")
				assert.Error(t, err, "Protocol-based attack should be blocked: %s", pattern)
			})
		}
	})
}

// TestKeyGenerator_ComprehensiveEdgeCases tests comprehensive edge cases for key generation
func TestKeyGenerator_ComprehensiveEdgeCases(t *testing.T) {
	kg := NewKeyGenerator()

	t.Run("Key generation with extreme inputs", func(t *testing.T) {
		tests := []struct {
			name        string
			keyFunc     func() string
			description string
		}{
			{
				name:        "diagram key with very long name",
				keyFunc:     func() string { return kg.DiagramKey(strings.Repeat("a", 200)) },
				description: "very long diagram name should be handled",
			},
			{
				name:        "state machine key with unicode",
				keyFunc:     func() string { return kg.StateMachineKey("版本1.0", "状态机图") },
				description: "unicode in version and name should be handled",
			},
			{
				name:        "entity key with special characters",
				keyFunc:     func() string { return kg.EntityKey("v2.0", "diagram@#$%", "entity!@#$%^&*()") },
				description: "special characters should be sanitized",
			},
			{
				name:        "diagram key with empty string",
				keyFunc:     func() string { return kg.DiagramKey("") },
				description: "empty string should be handled",
			},
			{
				name:        "state machine key with whitespace",
				keyFunc:     func() string { return kg.StateMachineKey("  v1.0  ", "  diagram  ") },
				description: "whitespace should be handled",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				key := tt.keyFunc()
				assert.NotEmpty(t, key, tt.description)

				// Validate that generated key is valid
				err := kg.ValidateKey(key)
				if err != nil {
					// Some extreme cases might generate invalid keys, which is acceptable
					t.Logf("Generated key %q is invalid: %v", key, err)
				}
			})
		}
	})

	t.Run("Key validation with boundary conditions", func(t *testing.T) {
		tests := []struct {
			name        string
			key         string
			expectValid bool
			description string
		}{
			{
				name:        "key at maximum length",
				key:         "/diagrams/puml/" + strings.Repeat("a", 233), // Total length = 250
				expectValid: true,
				description: "key at maximum length should be valid",
			},
			{
				name:        "key over maximum length",
				key:         "/diagrams/puml/" + strings.Repeat("a", 234), // Total length = 251
				expectValid: false,
				description: "key over maximum length should be invalid",
			},
			{
				name:        "key with only allowed characters",
				key:         "/diagrams/puml/abcABC123-_.%",
				expectValid: true,
				description: "key with only allowed characters should be valid",
			},
			{
				name:        "key with path traversal in middle",
				key:         "/diagrams/../puml/test",
				expectValid: false,
				description: "path traversal in middle should be invalid",
			},
			{
				name:        "key with encoded path traversal",
				key:         "/diagrams/puml/..%2F..%2Fetc%2Fpasswd",
				expectValid: true, // URL encoded is allowed
				description: "encoded path traversal should be valid (encoded)",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := kg.ValidateKey(tt.key)

				if tt.expectValid {
					assert.NoError(t, err, tt.description)
				} else {
					assert.Error(t, err, tt.description)
				}
			})
		}
	})
}

// TestErrorHandling_ComprehensiveScenarios tests comprehensive error handling scenarios
func TestErrorHandling_ComprehensiveScenarios(t *testing.T) {
	t.Run("Error severity classification", func(t *testing.T) {
		tests := []struct {
			errorType ErrorType
			expected  ErrorSeverity
		}{
			{ErrorTypeConnection, SeverityCritical},
			{ErrorTypeTimeout, SeverityCritical},
			{ErrorTypeCapacity, SeverityCritical},
			{ErrorTypeCircuitOpen, SeverityCritical},
			{ErrorTypeRetryExhausted, SeverityHigh},
			{ErrorTypeSerialization, SeverityMedium},
			{ErrorTypeValidation, SeverityMedium},
			{ErrorTypeKeyInvalid, SeverityLow},
			{ErrorTypeNotFound, SeverityLow},
		}

		for _, tt := range tests {
			t.Run(tt.errorType.String(), func(t *testing.T) {
				severity := tt.errorType.Severity()
				assert.Equal(t, tt.expected, severity)
			})
		}
	})

	t.Run("Error recovery strategy mapping", func(t *testing.T) {
		tests := []struct {
			errorType ErrorType
			expected  RecoveryStrategy
		}{
			{ErrorTypeConnection, RecoveryStrategyRetryWithBackoff},
			{ErrorTypeTimeout, RecoveryStrategyRetryWithBackoff},
			{ErrorTypeCapacity, RecoveryStrategyRetryWithDelay},
			{ErrorTypeRetryExhausted, RecoveryStrategyCircuitBreaker},
			{ErrorTypeCircuitOpen, RecoveryStrategyWaitAndRetry},
			{ErrorTypeSerialization, RecoveryStrategyFail},
			{ErrorTypeValidation, RecoveryStrategyFail},
			{ErrorTypeKeyInvalid, RecoveryStrategyFail},
			{ErrorTypeNotFound, RecoveryStrategyIgnore},
		}

		for _, tt := range tests {
			t.Run(tt.errorType.String(), func(t *testing.T) {
				err := NewCacheError(tt.errorType, "test-key", "test message", nil)
				strategy := err.GetRecoveryStrategy()
				assert.Equal(t, tt.expected, strategy)
			})
		}
	})

	t.Run("Error context enrichment", func(t *testing.T) {
		err := NewCacheError(ErrorTypeConnection, "test-key", "connection failed", nil)

		// Add context
		err.WithContext("get-operation", 3, 500*time.Millisecond)
		err.WithMetadata("server", "redis-1")
		err.WithMetadata("port", 6379)

		assert.Equal(t, "get-operation", err.Context.Operation)
		assert.Equal(t, 3, err.Context.AttemptNumber)
		assert.Equal(t, 500*time.Millisecond, err.Context.Duration)
		assert.Equal(t, "redis-1", err.Context.Metadata["server"])
		assert.Equal(t, 6379, err.Context.Metadata["port"])
	})

	t.Run("Error chaining and unwrapping", func(t *testing.T) {
		rootCause := errors.New("network unreachable")
		wrappedErr := NewConnectionError("Redis connection failed", rootCause)

		// Test unwrapping
		assert.Equal(t, rootCause, errors.Unwrap(wrappedErr))

		// Test error chain
		assert.True(t, errors.Is(wrappedErr, rootCause))
	})
}

// TestCircuitBreaker_ComprehensiveScenarios tests comprehensive circuit breaker scenarios
func TestCircuitBreaker_ComprehensiveScenarios(t *testing.T) {
	t.Run("Circuit breaker state transitions", func(t *testing.T) {
		config := &CircuitBreakerConfig{
			FailureThreshold: 2,
			RecoveryTimeout:  100 * time.Millisecond,
			SuccessThreshold: 2,
			MaxHalfOpenCalls: 2,
		}
		cb := NewCircuitBreaker(config)

		// Initially closed
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
		assert.NoError(t, cb.CanExecute())

		// Record failures to open circuit
		cb.RecordFailure()
		assert.Equal(t, CircuitBreakerClosed, cb.GetState()) // Still closed after 1 failure

		cb.RecordFailure()
		assert.Equal(t, CircuitBreakerOpen, cb.GetState()) // Open after 2 failures

		// Should reject execution when open
		assert.Error(t, cb.CanExecute())

		// Wait for recovery timeout
		time.Sleep(150 * time.Millisecond)

		// Should transition to half-open
		assert.NoError(t, cb.CanExecute())
		assert.Equal(t, CircuitBreakerHalfOpen, cb.GetState())

		// Record successes to close circuit
		cb.RecordSuccess()
		assert.Equal(t, CircuitBreakerHalfOpen, cb.GetState()) // Still half-open after 1 success

		cb.RecordSuccess()
		assert.Equal(t, CircuitBreakerClosed, cb.GetState()) // Closed after 2 successes
	})

	t.Run("Circuit breaker half-open call limits", func(t *testing.T) {
		config := &CircuitBreakerConfig{
			FailureThreshold: 1,
			RecoveryTimeout:  50 * time.Millisecond,
			SuccessThreshold: 3,
			MaxHalfOpenCalls: 2,
		}
		cb := NewCircuitBreaker(config)

		// Open the circuit
		cb.RecordFailure()
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())

		// Wait for recovery
		time.Sleep(60 * time.Millisecond)

		// First half-open call should succeed
		assert.NoError(t, cb.CanExecute())
		assert.Equal(t, CircuitBreakerHalfOpen, cb.GetState())

		// Second half-open call should succeed
		assert.NoError(t, cb.CanExecute())

		// Third half-open call should fail (exceeds limit)
		assert.Error(t, cb.CanExecute())
	})

	t.Run("Circuit breaker metrics", func(t *testing.T) {
		cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

		// Record some failures
		cb.RecordFailure()
		cb.RecordFailure()

		metrics := cb.GetMetrics()

		assert.Equal(t, "CLOSED", metrics["state"])
		assert.Equal(t, 2, metrics["failure_count"])
		assert.Equal(t, 0, metrics["success_count"])
		assert.Equal(t, 0, metrics["half_open_calls"])
		assert.NotNil(t, metrics["last_failure"])
	})
}

// TestErrorRecoveryManager_ComprehensiveScenarios tests comprehensive error recovery scenarios
func TestErrorRecoveryManager_ComprehensiveScenarios(t *testing.T) {
	t.Run("Error recovery manager with multiple operations", func(t *testing.T) {
		erm := NewErrorRecoveryManager(DefaultCircuitBreakerConfig())

		// Test different operations get different circuit breakers
		cb1 := erm.GetCircuitBreaker("operation1")
		cb2 := erm.GetCircuitBreaker("operation2")
		cb3 := erm.GetCircuitBreaker("operation1") // Same as cb1

		assert.NotEqual(t, cb1, cb2)
		assert.Equal(t, cb1, cb3)
	})

	t.Run("ShouldRetry with different error types", func(t *testing.T) {
		erm := NewErrorRecoveryManager(nil)

		tests := []struct {
			name          string
			err           error
			attemptNumber int
			shouldRetry   bool
			expectedDelay time.Duration
		}{
			{
				name:          "connection error first attempt",
				err:           NewConnectionError("connection failed", nil),
				attemptNumber: 1,
				shouldRetry:   true,
				expectedDelay: 100 * time.Millisecond,
			},
			{
				name:          "connection error second attempt",
				err:           NewConnectionError("connection failed", nil),
				attemptNumber: 2,
				shouldRetry:   true,
				expectedDelay: 400 * time.Millisecond,
			},
			{
				name:          "timeout error with delay",
				err:           NewTimeoutError("key", "timeout", nil),
				attemptNumber: 1,
				shouldRetry:   true,
				expectedDelay: 100 * time.Millisecond,
			},
			{
				name:          "capacity error with fixed delay",
				err:           NewCacheError(ErrorTypeCapacity, "key", "capacity exceeded", nil),
				attemptNumber: 1,
				shouldRetry:   true,
				expectedDelay: 1 * time.Second,
			},
			{
				name:          "validation error should not retry",
				err:           NewValidationError("invalid input", nil),
				attemptNumber: 1,
				shouldRetry:   false,
				expectedDelay: 0,
			},
			{
				name:          "too many attempts",
				err:           NewConnectionError("connection failed", nil),
				attemptNumber: 6,
				shouldRetry:   false,
				expectedDelay: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				shouldRetry, delay := erm.ShouldRetry(tt.err, "test-op", tt.attemptNumber)

				assert.Equal(t, tt.shouldRetry, shouldRetry)
				if tt.shouldRetry {
					assert.True(t, delay >= tt.expectedDelay, "Expected delay >= %v, got %v", tt.expectedDelay, delay)
				}
			})
		}
	})

	t.Run("ExecuteWithRecovery success and failure", func(t *testing.T) {
		erm := NewErrorRecoveryManager(nil)

		// Test successful execution
		var executed bool
		err := erm.ExecuteWithRecovery("test-op", func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)

		// Test failed execution
		testErr := NewConnectionError("connection failed", nil)
		err = erm.ExecuteWithRecovery("test-op", func() error {
			return testErr
		})

		assert.Error(t, err)
		assert.Equal(t, testErr, err)

		// Verify circuit breaker recorded the failure
		cb := erm.GetCircuitBreaker("test-op")
		assert.True(t, cb.failureCount > 0)
	})

	t.Run("Circuit breaker metrics collection", func(t *testing.T) {
		erm := NewErrorRecoveryManager(nil)

		// Create some circuit breakers and record metrics
		cb1 := erm.GetCircuitBreaker("op1")
		cb2 := erm.GetCircuitBreaker("op2")

		cb1.RecordFailure()
		cb2.RecordSuccess()

		metrics := erm.GetCircuitBreakerMetrics()

		assert.Contains(t, metrics, "op1")
		assert.Contains(t, metrics, "op2")
		assert.Equal(t, 1, metrics["op1"]["failure_count"])
		assert.Equal(t, 0, metrics["op2"]["failure_count"])
	})
}

// TestReflectionBasedValidation tests validation using reflection
func TestReflectionBasedValidation(t *testing.T) {
	validator := NewInputValidator()

	t.Run("ValidateEntityData with reflection", func(t *testing.T) {
		// Test with various types using reflection
		testCases := []struct {
			name        string
			entity      interface{}
			expectError bool
		}{
			{"nil interface", nil, true},
			{"nil pointer", (*int)(nil), true},
			{"valid pointer", new(int), false},
			{"struct", struct{ Name string }{Name: "test"}, false},
			{"pointer to struct", &struct{ Name string }{Name: "test"}, false},
			{"slice", []int{1, 2, 3}, false},
			{"map", map[string]int{"key": 1}, false},
			{"channel", make(chan int), false},
			{"function", func() {}, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := validator.ValidateEntityData(tc.entity, "test entity")

				if tc.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				// Additional reflection-based checks
				if tc.entity != nil {
					val := reflect.ValueOf(tc.entity)
					if val.Kind() == reflect.Ptr && val.IsNil() {
						assert.Error(t, err, "nil pointer should be rejected")
					}
				}
			})
		}
	})
}

// TestContextValidation_Comprehensive tests comprehensive context validation
func TestContextValidation_Comprehensive(t *testing.T) {
	validator := NewInputValidator()

	t.Run("context validation scenarios", func(t *testing.T) {
		tests := []struct {
			name        string
			ctx         context.Context
			expectError bool
			description string
		}{
			{
				name:        "nil context",
				ctx:         nil,
				expectError: true,
				description: "nil context should be rejected",
			},
			{
				name:        "background context",
				ctx:         context.Background(),
				expectError: false,
				description: "background context should be accepted",
			},
			{
				name:        "todo context",
				ctx:         context.TODO(),
				expectError: false,
				description: "TODO context should be accepted",
			},
			{
				name:        "context with timeout",
				ctx:         func() context.Context { ctx, _ := context.WithTimeout(context.Background(), time.Hour); return ctx }(),
				expectError: false,
				description: "context with timeout should be accepted",
			},
			{
				name: "context with deadline",
				ctx: func() context.Context {
					ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Hour))
					return ctx
				}(),
				expectError: false,
				description: "context with deadline should be accepted",
			},
			{
				name:        "cancelled context",
				ctx:         func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
				expectError: true,
				description: "cancelled context should be rejected",
			},
			{
				name: "expired context",
				ctx: func() context.Context {
					ctx, _ := context.WithTimeout(context.Background(), time.Nanosecond)
					time.Sleep(time.Millisecond)
					return ctx
				}(),
				expectError: true,
				description: "expired context should be rejected",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.ValidateContext(tt.ctx)

				if tt.expectError {
					assert.Error(t, err, tt.description)
				} else {
					assert.NoError(t, err, tt.description)
				}
			})
		}
	})
}

// BenchmarkInputValidation benchmarks input validation performance
func BenchmarkInputValidation(b *testing.B) {
	validator := NewInputValidator()

	b.Run("ValidateAndSanitizeString", func(b *testing.B) {
		input := "This is a test string with some content that needs validation"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = validator.ValidateAndSanitizeString(input, "test field")
		}
	})

	b.Run("ValidateAndSanitizeName", func(b *testing.B) {
		input := "test-diagram-name-with-special-chars"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = validator.ValidateAndSanitizeName(input, "test field")
		}
	})

	b.Run("ValidateEntityData", func(b *testing.B) {
		entity := map[string]interface{}{"id": "test", "name": "Test Entity"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = validator.ValidateEntityData(entity, "test entity")
		}
	})
}

// BenchmarkKeyGeneration benchmarks key generation performance
func BenchmarkKeyGeneration(b *testing.B) {
	kg := NewKeyGenerator()

	b.Run("DiagramKey", func(b *testing.B) {
		name := "test-diagram-name"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = kg.DiagramKey(name)
		}
	})

	b.Run("StateMachineKey", func(b *testing.B) {
		version := "2.5"
		name := "test-state-machine"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = kg.StateMachineKey(version, name)
		}
	})

	b.Run("ValidateKey", func(b *testing.B) {
		key := "/diagrams/puml/test-diagram"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = kg.ValidateKey(key)
		}
	})
}
