package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/cache"
	"github.com/kengibson1111/go-uml-statemachine-models/models"
)

// Example PlantUML diagram content
const samplePUMLDiagram = `@startuml
!define PUML_VERSION 1.2024.8

state "Idle" as idle
state "Processing" as processing
state "Complete" as complete
state "Error" as error

[*] --> idle
idle --> processing : start
processing --> complete : success
processing --> error : failure
complete --> [*]
error --> idle : retry
@enduml`

func main() {
	fmt.Println("=== Redis Cache Diagram Example ===")
	fmt.Println("This example demonstrates caching PlantUML diagrams")
	fmt.Println()

	// Create Redis cache configuration
	config := cache.DefaultRedisConfig()
	config.RedisAddr = "localhost:6379"
	config.DefaultTTL = 1 * time.Hour

	// Create cache instance
	redisCache, err := cache.NewRedisCache(config)
	if err != nil {
		log.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer func() {
		if err := redisCache.Close(); err != nil {
			log.Printf("Error closing cache: %v", err)
		}
	}()

	ctx := context.Background()

	// Test Redis connection
	fmt.Println("1. Testing Redis connection...")
	if err := redisCache.Health(ctx); err != nil {
		log.Fatalf("Redis health check failed: %v", err)
	}
	fmt.Println("✓ Redis connection successful")
	fmt.Println()

	// Store a diagram
	diagramName := "simple-state-machine"
	fmt.Printf("2. Storing diagram '%s'...\n", diagramName)

	err = redisCache.StoreDiagram(ctx, models.DiagramTypePUML, diagramName, samplePUMLDiagram, 30*time.Minute)
	if err != nil {
		log.Fatalf("Failed to store diagram: %v", err)
	}
	fmt.Println("✓ Diagram stored successfully")
	fmt.Println()

	// Retrieve the diagram
	fmt.Printf("3. Retrieving diagram '%s'...\n", diagramName)
	retrievedDiagram, err := redisCache.GetDiagram(ctx, models.DiagramTypePUML, diagramName)
	if err != nil {
		log.Fatalf("Failed to retrieve diagram: %v", err)
	}
	fmt.Println("✓ Diagram retrieved successfully")
	fmt.Printf("Retrieved content length: %d characters\n", len(retrievedDiagram))
	fmt.Println()

	// Verify content matches
	fmt.Println("4. Verifying diagram content...")
	if retrievedDiagram == samplePUMLDiagram {
		fmt.Println("✓ Content matches original")
	} else {
		fmt.Println("✗ Content mismatch!")
		fmt.Printf("Expected length: %d, Got length: %d\n", len(samplePUMLDiagram), len(retrievedDiagram))
	}
	fmt.Println()

	// Demonstrate error handling - try to get non-existent diagram
	fmt.Println("5. Testing error handling with non-existent diagram...")
	_, err = redisCache.GetDiagram(ctx, models.DiagramTypePUML, "non-existent-diagram")
	if err != nil {
		if cache.IsNotFoundError(err) {
			fmt.Println("✓ Correctly handled not found error")
		} else {
			fmt.Printf("✗ Unexpected error type: %v\n", err)
		}
	} else {
		fmt.Println("✗ Expected error but got none")
	}
	fmt.Println()

	// Demonstrate cleanup operations
	fmt.Println("6. Testing cleanup operations...")

	// Store multiple diagrams for cleanup demo
	testDiagrams := []string{"test-diagram-1", "test-diagram-2", "test-diagram-3"}
	for _, name := range testDiagrams {
		err := redisCache.StoreDiagram(ctx, models.DiagramTypePUML, name, samplePUMLDiagram, 5*time.Minute)
		if err != nil {
			log.Printf("Warning: Failed to store test diagram %s: %v", name, err)
		}
	}

	// Clean up test diagrams using pattern
	fmt.Println("Cleaning up test diagrams...")
	err = redisCache.Cleanup(ctx, "/diagrams/puml/test-diagram-*")
	if err != nil {
		log.Printf("Warning: Cleanup failed: %v", err)
	} else {
		fmt.Println("✓ Cleanup completed successfully")
	}
	fmt.Println()

	// Final cleanup - delete the main example diagram
	fmt.Printf("7. Cleaning up example diagram '%s'...\n", diagramName)
	err = redisCache.DeleteDiagram(ctx, models.DiagramTypePUML, diagramName)
	if err != nil {
		log.Printf("Warning: Failed to delete diagram: %v", err)
	} else {
		fmt.Println("✓ Example diagram deleted successfully")
	}

	fmt.Println()
	fmt.Println("=== Diagram Cache Example Complete ===")
}
