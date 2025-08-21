package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/internal"
)

func main() {
	// Example 1: Using default configuration
	fmt.Println("=== Example 1: Default Configuration ===")
	defaultClient, err := internal.NewRedisClient(nil)
	if err != nil {
		log.Printf("Failed to create Redis client with default config: %v", err)
	} else {
		fmt.Printf("Created Redis client with default config: %s\n", defaultClient.Config().RedisAddr)

		// Test health check
		ctx := context.Background()
		if err := defaultClient.Health(ctx); err != nil {
			log.Printf("Health check failed (Redis may not be running): %v", err)
		} else {
			fmt.Println("Health check passed!")
		}

		defaultClient.Close()
	}

	// Example 2: Using custom configuration
	fmt.Println("\n=== Example 2: Custom Configuration ===")
	customConfig := &internal.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       1, // Use database 1 instead of 0
		MaxRetries:    5,
		DialTimeout:   10 * time.Second,
		ReadTimeout:   5 * time.Second,
		WriteTimeout:  5 * time.Second,
		PoolSize:      20,
		DefaultTTL:    12 * time.Hour,
	}

	customClient, err := internal.NewRedisClient(customConfig)
	if err != nil {
		log.Printf("Failed to create Redis client with custom config: %v", err)
	} else {
		fmt.Printf("Created Redis client with custom config: DB=%d, PoolSize=%d\n",
			customClient.Config().RedisDB, customClient.Config().PoolSize)

		// Get connection info
		ctx := context.Background()
		if info, err := customClient.GetConnectionInfo(ctx); err != nil {
			log.Printf("Failed to get connection info: %v", err)
		} else {
			fmt.Printf("Connection info: %+v\n", info)
		}

		customClient.Close()
	}

	// Example 3: Configuration validation
	fmt.Println("\n=== Example 3: Configuration Validation ===")
	invalidConfig := &internal.Config{
		RedisAddr:    "", // Invalid: empty address
		RedisDB:      -1, // Invalid: negative database
		MaxRetries:   -1, // Invalid: negative retries
		DialTimeout:  0,  // Invalid: zero timeout
		ReadTimeout:  0,  // Invalid: zero timeout
		WriteTimeout: 0,  // Invalid: zero timeout
		PoolSize:     0,  // Invalid: zero pool size
		DefaultTTL:   0,  // Invalid: zero TTL
	}

	invalidClient, err := internal.NewRedisClient(invalidConfig)
	if err != nil {
		fmt.Printf("Expected validation error: %v\n", err)
	} else {
		fmt.Println("Unexpected: invalid config was accepted")
		invalidClient.Close()
	}

	// Example 4: Using DefaultConfig() as a starting point
	fmt.Println("\n=== Example 4: Modifying Default Configuration ===")
	modifiedConfig := internal.DefaultConfig()
	modifiedConfig.RedisDB = 2
	modifiedConfig.PoolSize = 15
	modifiedConfig.DefaultTTL = 6 * time.Hour

	modifiedClient, err := internal.NewRedisClient(modifiedConfig)
	if err != nil {
		log.Printf("Failed to create Redis client with modified config: %v", err)
	} else {
		fmt.Printf("Created Redis client with modified config: DB=%d, TTL=%v\n",
			modifiedClient.Config().RedisDB, modifiedClient.Config().DefaultTTL)
		modifiedClient.Close()
	}

	fmt.Println("\n=== Redis Client Configuration Examples Complete ===")
}
