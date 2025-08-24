package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/internal"
)

func main() {
	fmt.Println("=== Redis Retry Mechanism Example ===")
	fmt.Println("This example demonstrates retry logic and error recovery")
	fmt.Println()

	// Create configurations with different retry strategies
	configs := createRetryConfigurations()

	for i, config := range configs {
		fmt.Printf("=== Configuration %d: %s ===\n", i+1, config.name)
		fmt.Printf("Max Attempts: %d\n", config.config.RetryConfig.MaxAttempts)
		fmt.Printf("Initial Delay: %v\n", config.config.RetryConfig.InitialDelay)
		fmt.Printf("Max Delay: %v\n", config.config.RetryConfig.MaxDelay)
		fmt.Printf("Multiplier: %.1f\n", config.config.RetryConfig.Multiplier)
		fmt.Printf("Jitter: %t\n", config.config.RetryConfig.Jitter)
		fmt.Println()

		demonstrateRetryBehavior(config.config, config.name)
		fmt.Println()
	}

	// Demonstrate error recovery manager
	fmt.Println("=== Error Recovery Manager Example ===")
	demonstrateErrorRecoveryManager()

	fmt.Println()
	fmt.Println("=== Retry Mechanism Example Complete ===")
}

type retryConfig struct {
	name   string
	config *internal.Config
}

func createRetryConfigurations() []retryConfig {
	// Configuration 1: Conservative retry
	conservativeConfig := internal.DefaultConfig()
	conservativeConfig.RedisAddr = "localhost:6379"
	conservativeConfig.RetryConfig = &internal.RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
		RetryableOps: []string{"ping", "get", "set", "del"},
	}

	// Configuration 2: Aggressive retry
	aggressiveConfig := internal.DefaultConfig()
	aggressiveConfig.RedisAddr = "localhost:6379"
	aggressiveConfig.RetryConfig = &internal.RetryConfig{
		MaxAttempts:  7,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   1.5,
		Jitter:       true,
		RetryableOps: []string{"ping", "get", "set", "del", "exists", "scan"},
	}

	// Configuration 3: Fast retry with jitter
	fastConfig := internal.DefaultConfig()
	fastConfig.RedisAddr = "localhost:6379"
	fastConfig.RetryConfig = &internal.RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 25 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.5,
		Jitter:       true,
		RetryableOps: []string{"ping", "get", "set"},
	}

	return []retryConfig{
		{"Conservative Retry", conservativeConfig},
		{"Aggressive Retry", aggressiveConfig},
		{"Fast Retry with Jitter", fastConfig},
	}
}

func demonstrateRetryBehavior(config *internal.Config, configName string) {
	// Create client with the specific configuration
	client, err := internal.NewRedisClient(config)
	if err != nil {
		log.Printf("Failed to create client for %s: %v", configName, err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Test successful operation (should work immediately)
	fmt.Println("1. Testing successful operation...")
	start := time.Now()
	err = client.HealthWithRetry(ctx)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("✗ Health check failed: %v (took %v)\n", err, duration)
	} else {
		fmt.Printf("✓ Health check succeeded on first attempt (took %v)\n", duration)
	}

	// Test operation with valid Redis (should succeed quickly)
	fmt.Println("2. Testing SET/GET operations with retry...")
	testKey := fmt.Sprintf("retry-test:%s:%d", configName, time.Now().Unix())
	testValue := "retry-test-value"

	start = time.Now()
	err = client.SetWithRetry(ctx, testKey, testValue, 1*time.Minute)
	setDuration := time.Since(start)

	if err != nil {
		fmt.Printf("✗ SET operation failed: %v (took %v)\n", err, setDuration)
	} else {
		fmt.Printf("✓ SET operation succeeded (took %v)\n", setDuration)

		// Test GET operation
		start = time.Now()
		retrievedValue, err := client.GetWithRetry(ctx, testKey)
		getDuration := time.Since(start)

		if err != nil {
			fmt.Printf("✗ GET operation failed: %v (took %v)\n", err, getDuration)
		} else if retrievedValue == testValue {
			fmt.Printf("✓ GET operation succeeded (took %v)\n", getDuration)
		} else {
			fmt.Printf("✗ GET operation returned wrong value: got '%s', expected '%s'\n", retrievedValue, testValue)
		}

		// Cleanup
		_ = client.DelWithRetry(ctx, testKey)
	}

	// Demonstrate retry timing calculation
	fmt.Println("3. Demonstrating retry delay calculation...")
	demonstrateRetryDelays(config.RetryConfig)
}

func demonstrateRetryDelays(retryConfig *internal.RetryConfig) {
	fmt.Println("Calculated retry delays for failed attempts:")

	for attempt := 0; attempt < retryConfig.MaxAttempts; attempt++ {
		delay := calculateRetryDelay(retryConfig, attempt)
		fmt.Printf("  Attempt %d: %v\n", attempt+1, delay)
	}
}

// calculateRetryDelay mimics the internal retry delay calculation
func calculateRetryDelay(config *internal.RetryConfig, attempt int) time.Duration {
	if config == nil {
		return time.Second
	}

	// Calculate exponential backoff
	delay := float64(config.InitialDelay)
	for i := 0; i < attempt; i++ {
		delay *= config.Multiplier
	}

	// Cap at max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// Note: We don't add jitter in this demo to show consistent values
	return time.Duration(delay)
}

func demonstrateErrorRecoveryManager() {
	fmt.Println("Creating error recovery manager with circuit breaker...")

	// Create circuit breaker configuration
	circuitConfig := internal.DefaultCircuitBreakerConfig()
	circuitConfig.FailureThreshold = 3
	circuitConfig.RecoveryTimeout = 5 * time.Second
	circuitConfig.MaxHalfOpenCalls = 2

	fmt.Printf("Circuit Breaker Configuration:\n")
	fmt.Printf("  Failure Threshold: %d\n", circuitConfig.FailureThreshold)
	fmt.Printf("  Recovery Timeout: %v\n", circuitConfig.RecoveryTimeout)
	fmt.Printf("  Max Half-Open Calls: %d\n", circuitConfig.MaxHalfOpenCalls)
	fmt.Println()

	// Create error recovery manager
	recoveryManager := internal.NewErrorRecoveryManager(circuitConfig)

	// Demonstrate error classification
	fmt.Println("Demonstrating error classification...")

	// Create sample errors
	sampleErrors := []error{
		internal.NewConnectionError("connection refused", nil),
		internal.NewNotFoundError("test-key"),
		internal.NewValidationError("invalid input", nil),
		internal.NewTimeoutError("test-key", "operation timeout", nil),
		internal.NewRetryExhaustedError("test-operation", 3, nil),
		internal.NewCircuitOpenError("test-operation"),
	}

	for _, err := range sampleErrors {
		severity := internal.GetErrorSeverity(err)
		strategy := internal.GetRecoveryStrategy(err)
		retryable := internal.IsRetryableError(err)

		fmt.Printf("Error: %v\n", err)
		fmt.Printf("  Severity: %v\n", severity)
		fmt.Printf("  Recovery Strategy: %v\n", strategy)
		fmt.Printf("  Retryable: %t\n", retryable)
		fmt.Println()
	}

	// Demonstrate error type checking
	fmt.Println("Demonstrating error type checking...")

	connectionErr := internal.NewConnectionError("test connection error", nil)
	notFoundErr := internal.NewNotFoundError("missing-key")
	validationErr := internal.NewValidationError("test validation error", nil)

	fmt.Printf("Connection Error Check: %t\n", internal.IsConnectionError(connectionErr))
	fmt.Printf("Not Found Error Check: %t\n", internal.IsNotFoundError(notFoundErr))
	fmt.Printf("Validation Error Check: %t\n", internal.IsValidationError(validationErr))
	fmt.Printf("Retry Exhausted Error Check: %t\n", internal.IsRetryExhaustedError(internal.NewRetryExhaustedError("test", 3, nil)))
	fmt.Printf("Circuit Open Error Check: %t\n", internal.IsCircuitOpenError(internal.NewCircuitOpenError("test")))
	fmt.Println()

	// Demonstrate error context
	fmt.Println("Demonstrating error context...")

	errorContext := &internal.ErrorContext{
		Operation:     "test-operation",
		AttemptNumber: 2,
		Duration:      150 * time.Millisecond,
		Timestamp:     time.Now(),
		Metadata: map[string]any{
			"max_attempts": 3,
			"key":          "test-key",
		},
	}

	contextualError := internal.NewCacheErrorWithContext(
		internal.ErrorTypeTimeout,
		"test-key",
		"operation timed out with context",
		nil,
		errorContext,
	)

	fmt.Printf("Contextual Error: %v\n", contextualError)
	fmt.Printf("Error has context: %t\n", contextualError.Context != nil)
	if contextualError.Context != nil {
		fmt.Printf("  Operation: %s\n", contextualError.Context.Operation)
		fmt.Printf("  Attempt: %d\n", contextualError.Context.AttemptNumber)
		fmt.Printf("  Duration: %v\n", contextualError.Context.Duration)
		if maxAttempts, ok := contextualError.Context.Metadata["max_attempts"]; ok {
			fmt.Printf("  Max Attempts: %v\n", maxAttempts)
		}
	}

	fmt.Println()
	fmt.Printf("Recovery manager created successfully\n")

	// Demonstrate recovery manager functionality
	fmt.Println("Testing recovery manager...")
	testOperation := "test-cache-operation"

	// Test successful operation
	err := recoveryManager.ExecuteWithRecovery(testOperation, func() error {
		return nil // Simulate successful operation
	})
	if err != nil {
		fmt.Printf("✗ Recovery manager test failed: %v\n", err)
	} else {
		fmt.Printf("✓ Recovery manager executed successful operation\n")
	}

	// Get circuit breaker metrics
	metrics := recoveryManager.GetCircuitBreakerMetrics()
	if len(metrics) > 0 {
		fmt.Printf("Circuit breaker metrics: %d operations tracked\n", len(metrics))
	}
}
