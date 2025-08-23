package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/internal"
)

func main() {
	// Create a configuration with custom retry settings
	config := internal.DefaultConfig()

	// Customize retry behavior
	config.RetryConfig.MaxAttempts = 5
	config.RetryConfig.InitialDelay = 50 * time.Millisecond
	config.RetryConfig.MaxDelay = 2 * time.Second
	config.RetryConfig.Multiplier = 1.5
	config.RetryConfig.Jitter = true

	fmt.Println("Redis Cache Retry Example")
	fmt.Println("========================")
	fmt.Printf("Retry Configuration:\n")
	fmt.Printf("  Max Attempts: %d\n", config.RetryConfig.MaxAttempts)
	fmt.Printf("  Initial Delay: %v\n", config.RetryConfig.InitialDelay)
	fmt.Printf("  Max Delay: %v\n", config.RetryConfig.MaxDelay)
	fmt.Printf("  Multiplier: %.1f\n", config.RetryConfig.Multiplier)
	fmt.Printf("  Jitter: %v\n", config.RetryConfig.Jitter)
	fmt.Printf("  Retryable Operations: %v\n", config.RetryConfig.RetryableOps)
	fmt.Println()

	// Create Redis client
	client, err := internal.NewRedisClient(config)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Demonstrate health check with retry
	fmt.Println("Testing Health Check with Retry...")
	start := time.Now()
	err = client.HealthWithRetry(ctx)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Health check failed after %v: %v\n", duration, err)
		fmt.Println("   This is expected if Redis is not running")
	} else {
		fmt.Printf("✅ Health check succeeded in %v\n", duration)
	}
	fmt.Println()

	// Demonstrate ping with retry
	fmt.Println("Testing Ping with Retry...")
	start = time.Now()
	err = client.PingWithRetry(ctx)
	duration = time.Since(start)

	if err != nil {
		fmt.Printf("❌ Ping failed after %v: %v\n", duration, err)
		fmt.Println("   This is expected if Redis is not running")
	} else {
		fmt.Printf("✅ Ping succeeded in %v\n", duration)
	}
	fmt.Println()

	// If Redis is running, demonstrate successful operations
	if err == nil {
		fmt.Println("Redis is available! Testing successful operations...")

		// Test SET with retry
		fmt.Println("Testing SET with retry...")
		start = time.Now()
		err = client.SetWithRetry(ctx, "test:retry:key", "test-value", time.Hour)
		duration = time.Since(start)

		if err != nil {
			fmt.Printf("❌ SET failed after %v: %v\n", duration, err)
		} else {
			fmt.Printf("✅ SET succeeded in %v\n", duration)
		}

		// Test GET with retry
		fmt.Println("Testing GET with retry...")
		start = time.Now()
		value, err := client.GetWithRetry(ctx, "test:retry:key")
		duration = time.Since(start)

		if err != nil {
			fmt.Printf("❌ GET failed after %v: %v\n", duration, err)
		} else {
			fmt.Printf("✅ GET succeeded in %v, value: %s\n", duration, value)
		}

		// Test EXISTS with retry
		fmt.Println("Testing EXISTS with retry...")
		start = time.Now()
		exists, err := client.ExistsWithRetry(ctx, "test:retry:key")
		duration = time.Since(start)

		if err != nil {
			fmt.Printf("❌ EXISTS failed after %v: %v\n", duration, err)
		} else {
			fmt.Printf("✅ EXISTS succeeded in %v, exists: %d\n", duration, exists)
		}

		// Test DEL with retry
		fmt.Println("Testing DEL with retry...")
		start = time.Now()
		err = client.DelWithRetry(ctx, "test:retry:key")
		duration = time.Since(start)

		if err != nil {
			fmt.Printf("❌ DEL failed after %v: %v\n", duration, err)
		} else {
			fmt.Printf("✅ DEL succeeded in %v\n", duration)
		}
	}

	fmt.Println()
	fmt.Println("Example completed!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("- Exponential backoff with configurable parameters")
	fmt.Println("- Jitter to prevent thundering herd")
	fmt.Println("- Configurable retry attempts and delays")
	fmt.Println("- Operation-specific retry policies")
	fmt.Println("- Graceful handling of connection failures")
	fmt.Println("- Context-aware cancellation support")
}
