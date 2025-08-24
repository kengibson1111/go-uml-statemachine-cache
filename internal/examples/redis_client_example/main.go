package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/internal"
)

func main() {
	fmt.Println("=== Redis Client Low-Level Example ===")
	fmt.Println("This example demonstrates low-level Redis client operations")
	fmt.Println()

	// Create Redis client configuration
	config := internal.DefaultConfig()
	config.RedisAddr = "localhost:6379"
	config.DefaultTTL = 30 * time.Minute

	// Configure retry behavior
	retryConfig := internal.DefaultRetryConfig()
	retryConfig.MaxAttempts = 5
	retryConfig.InitialDelay = 50 * time.Millisecond
	retryConfig.MaxDelay = 2 * time.Second
	retryConfig.Multiplier = 2.0
	retryConfig.Jitter = true
	config.RetryConfig = retryConfig

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Redis Address: %s\n", config.RedisAddr)
	fmt.Printf("  Database: %d\n", config.RedisDB)
	fmt.Printf("  Pool Size: %d\n", config.PoolSize)
	fmt.Printf("  Default TTL: %v\n", config.DefaultTTL)
	fmt.Printf("  Max Retries: %d\n", retryConfig.MaxAttempts)
	fmt.Println()

	// Create Redis client
	client, err := internal.NewRedisClient(config)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing client: %v", err)
		}
	}()

	ctx := context.Background()

	// Test basic connectivity
	fmt.Println("1. Testing basic connectivity...")
	if err := client.HealthWithRetry(ctx); err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Println("✓ Redis client connected successfully")
	fmt.Println()

	// Demonstrate basic operations with retry
	fmt.Println("2. Demonstrating basic operations with retry logic...")

	// SET operation
	key := "example:test-key"
	value := "Hello, Redis Cache!"
	ttl := 5 * time.Minute

	fmt.Printf("Setting key '%s' with TTL %v...\n", key, ttl)
	err = client.SetWithRetry(ctx, key, value, ttl)
	if err != nil {
		log.Fatalf("Failed to set key: %v", err)
	}
	fmt.Println("✓ Key set successfully")

	// GET operation
	fmt.Printf("Getting key '%s'...\n", key)
	retrievedValue, err := client.GetWithRetry(ctx, key)
	if err != nil {
		log.Fatalf("Failed to get key: %v", err)
	}
	fmt.Printf("✓ Retrieved value: '%s'\n", retrievedValue)

	// Verify value matches
	if retrievedValue == value {
		fmt.Println("✓ Value matches original")
	} else {
		fmt.Printf("✗ Value mismatch! Expected: '%s', Got: '%s'\n", value, retrievedValue)
	}
	fmt.Println()

	// Demonstrate EXISTS operation
	fmt.Println("3. Testing EXISTS operation...")
	exists, err := client.ExistsWithRetry(ctx, key)
	if err != nil {
		log.Fatalf("Failed to check key existence: %v", err)
	}
	fmt.Printf("✓ Key exists: %t (count: %d)\n", exists > 0, exists)
	fmt.Println()

	// Demonstrate batch operations
	fmt.Println("4. Demonstrating batch operations...")

	// Set multiple keys
	testKeys := []string{"batch:key1", "batch:key2", "batch:key3"}
	testValues := []string{"value1", "value2", "value3"}

	fmt.Printf("Setting %d keys...\n", len(testKeys))
	for i, testKey := range testKeys {
		err := client.SetWithRetry(ctx, testKey, testValues[i], 2*time.Minute)
		if err != nil {
			log.Printf("Warning: Failed to set key %s: %v", testKey, err)
		}
	}
	fmt.Println("✓ Batch set completed")

	// Check existence of multiple keys
	fmt.Printf("Checking existence of %d keys...\n", len(testKeys))
	existsCount, err := client.ExistsWithRetry(ctx, testKeys...)
	if err != nil {
		log.Printf("Warning: Failed to check batch existence: %v", err)
	} else {
		fmt.Printf("✓ %d out of %d keys exist\n", existsCount, len(testKeys))
	}
	fmt.Println()

	// Demonstrate SCAN operation
	fmt.Println("5. Demonstrating SCAN operation...")

	fmt.Println("Scanning for keys matching 'batch:*'...")
	keys, cursor, err := client.ScanWithRetry(ctx, 0, "batch:*", 10)
	if err != nil {
		log.Printf("Warning: SCAN operation failed: %v", err)
	} else {
		fmt.Printf("✓ Found %d keys (cursor: %d)\n", len(keys), cursor)
		for _, foundKey := range keys {
			fmt.Printf("  - %s\n", foundKey)
		}
	}
	fmt.Println()

	// Demonstrate database size monitoring
	fmt.Println("6. Demonstrating database monitoring...")

	dbSize, err := client.DBSizeWithRetry(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get database size: %v", err)
	} else {
		fmt.Printf("✓ Database contains %d keys\n", dbSize)
	}

	// Get Redis server info
	fmt.Println("Getting Redis server information...")
	info, err := client.InfoWithRetry(ctx, "server")
	if err != nil {
		log.Printf("Warning: Failed to get server info: %v", err)
	} else {
		// Parse and display key server information
		fmt.Println("✓ Server information retrieved")
		lines := parseInfoString(info)
		for _, line := range lines {
			if contains(line, "redis_version") || contains(line, "uptime_in_seconds") || contains(line, "tcp_port") {
				fmt.Printf("  %s\n", line)
			}
		}
	}
	fmt.Println()

	// Demonstrate connection information
	fmt.Println("7. Demonstrating connection information...")

	connInfo, err := client.GetConnectionInfo(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get connection info: %v", err)
	} else {
		fmt.Println("✓ Connection information:")
		fmt.Printf("  Address: %v\n", connInfo["addr"])
		fmt.Printf("  Database: %v\n", connInfo["db"])
		fmt.Printf("  Pool Size: %v\n", connInfo["pool_size"])
		fmt.Printf("  Pool Hits: %v\n", connInfo["pool_hits"])
		fmt.Printf("  Pool Misses: %v\n", connInfo["pool_misses"])
		fmt.Printf("  Total Connections: %v\n", connInfo["pool_total_conns"])
		fmt.Printf("  Idle Connections: %v\n", connInfo["pool_idle_conns"])
	}
	fmt.Println()

	// Cleanup operations
	fmt.Println("8. Cleaning up test data...")

	// Delete individual key
	fmt.Printf("Deleting key '%s'...\n", key)
	err = client.DelWithRetry(ctx, key)
	if err != nil {
		log.Printf("Warning: Failed to delete key: %v", err)
	} else {
		fmt.Println("✓ Key deleted successfully")
	}

	// Delete batch keys
	fmt.Printf("Deleting %d batch keys...\n", len(testKeys))
	deletedCount, err := client.DelBatchWithRetry(ctx, testKeys...)
	if err != nil {
		log.Printf("Warning: Failed to delete batch keys: %v", err)
	} else {
		fmt.Printf("✓ Deleted %d keys\n", deletedCount)
	}

	fmt.Println()
	fmt.Println("=== Redis Client Low-Level Example Complete ===")
}

// parseInfoString parses Redis INFO command output into lines
func parseInfoString(info string) []string {
	var lines []string
	var currentLine string

	for _, char := range info {
		if char == '\n' || char == '\r' {
			if len(currentLine) > 0 && currentLine[0] != '#' {
				lines = append(lines, currentLine)
			}
			currentLine = ""
		} else {
			currentLine += string(char)
		}
	}

	// Add the last line if it exists
	if len(currentLine) > 0 && currentLine[0] != '#' {
		lines = append(lines, currentLine)
	}

	return lines
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				match := true
				for j := 0; j < len(substr); j++ {
					if s[i+j] != substr[j] {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
			return false
		}()
}
