package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/cache"
)

func main() {
	fmt.Println("Redis Cache Error Handling Example")
	fmt.Println("==================================")

	// Create a cache configuration
	config := cache.DefaultRedisConfig()
	config.RedisAddr = "localhost:6379" // Ensure Redis is running

	// Create cache instance
	redisCache, err := cache.NewRedisCache(config)
	if err != nil {
		log.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer redisCache.Close()

	ctx := context.Background()

	// Example 1: Demonstrate validation errors
	fmt.Println("\n1. Validation Error Example:")
	err = redisCache.StoreDiagram(ctx, "", "some content", time.Hour)
	if err != nil {
		if cache.IsValidationError(err) {
			fmt.Printf("   Validation error detected: %v\n", err)
			fmt.Printf("   Error severity: %v\n", cache.GetErrorSeverity(err))
			fmt.Printf("   Recovery strategy: %v\n", cache.GetRecoveryStrategy(err))
		}
	}

	// Example 2: Demonstrate not found errors
	fmt.Println("\n2. Not Found Error Example:")
	_, err = redisCache.GetDiagram(ctx, "non-existent-diagram")
	if err != nil {
		if cache.IsNotFoundError(err) {
			fmt.Printf("   Not found error detected: %v\n", err)
			fmt.Printf("   Error severity: %v\n", cache.GetErrorSeverity(err))
			fmt.Printf("   Recovery strategy: %v\n", cache.GetRecoveryStrategy(err))
		}
	}

	// Example 3: Demonstrate error context and metadata
	fmt.Println("\n3. Enhanced Error Context Example:")
	context := &cache.ErrorContext{
		Operation:     "store_diagram",
		AttemptNumber: 1,
		Duration:      50 * time.Millisecond,
		Metadata: map[string]any{
			"diagram_size": 1024,
			"user_id":      "user123",
		},
	}

	enhancedErr := cache.NewCacheErrorWithContext(
		cache.CacheErrorTypeTimeout,
		"test-diagram",
		"operation timed out during storage",
		fmt.Errorf("context deadline exceeded"),
		context,
	)

	fmt.Printf("   Enhanced error: %v\n", enhancedErr)
	fmt.Printf("   Error severity: %v\n", enhancedErr.Severity)
	fmt.Printf("   Is retryable: %v\n", enhancedErr.IsRetryable())
	fmt.Printf("   Recovery strategy: %v\n", enhancedErr.GetRecoveryStrategy())

	// Example 4: Demonstrate circuit breaker functionality
	fmt.Println("\n4. Circuit Breaker Example:")
	cbConfig := cache.DefaultCircuitBreakerConfig()
	cbConfig.FailureThreshold = 2
	cbConfig.RecoveryTimeout = 1 * time.Second

	erm := cache.NewErrorRecoveryManager(cbConfig)

	// Simulate failures to open circuit breaker
	for i := 0; i < 3; i++ {
		err := erm.ExecuteWithRecovery("test-operation", func() error {
			return cache.NewConnectionError("simulated connection failure", nil)
		})
		fmt.Printf("   Attempt %d: %v\n", i+1, err)
	}

	// Try to execute when circuit is open
	err = erm.ExecuteWithRecovery("test-operation", func() error {
		return nil // This won't be executed due to open circuit
	})
	if cache.IsCircuitOpenError(err) {
		fmt.Printf("   Circuit breaker is open: %v\n", err)
	}

	// Example 5: Demonstrate retry logic
	fmt.Println("\n5. Retry Logic Example:")
	connectionErr := cache.NewConnectionError("connection failed", nil)
	shouldRetry, delay := erm.ShouldRetry(connectionErr, "get-operation", 1)
	fmt.Printf("   Connection error should retry: %v (delay: %v)\n", shouldRetry, delay)

	validationErr := cache.NewValidationError("invalid input", nil)
	shouldRetry, delay = erm.ShouldRetry(validationErr, "set-operation", 1)
	fmt.Printf("   Validation error should retry: %v (delay: %v)\n", shouldRetry, delay)

	// Example 6: Demonstrate error categorization
	fmt.Println("\n6. Error Categorization Example:")
	errors := []error{
		cache.NewConnectionError("connection failed", nil),
		cache.NewTimeoutError("timeout-key", "operation timed out", nil),
		cache.NewValidationError("invalid input", nil),
		cache.NewNotFoundError("missing-key"),
		cache.NewRetryExhaustedError("failed-operation", 5, nil),
		cache.NewCircuitOpenError("blocked-operation"),
	}

	for _, err := range errors {
		fmt.Printf("   Error: %v\n", err)
		fmt.Printf("     - Severity: %v\n", cache.GetErrorSeverity(err))
		fmt.Printf("     - Retryable: %v\n", cache.IsRetryableError(err))
		fmt.Printf("     - Recovery: %v\n", cache.GetRecoveryStrategy(err))
		fmt.Println()
	}

	fmt.Println("Error handling example completed successfully!")
}
