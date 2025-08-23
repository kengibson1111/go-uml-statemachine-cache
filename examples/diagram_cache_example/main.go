package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/cache"
	"github.com/kengibson1111/go-uml-statemachine-cache/internal"
)

func main() {
	// Create Redis cache configuration
	config := internal.DefaultConfig()
	config.RedisAddr = "localhost:6379"
	config.DefaultTTL = 1 * time.Hour

	// Create cache instance
	redisCache, err := cache.NewRedisCache(config)
	if err != nil {
		log.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer redisCache.Close()

	ctx := context.Background()

	// Test Redis connection
	if err := redisCache.Health(ctx); err != nil {
		log.Fatalf("Redis health check failed: %v", err)
	}
	fmt.Println("✓ Redis connection successful")

	// Example PlantUML diagram content
	diagramName := "user-authentication-flow"
	pumlContent := `@startuml
title User Authentication State Machine

[*] --> Unauthenticated
Unauthenticated --> Authenticating : login()
Authenticating --> Authenticated : success
Authenticating --> Unauthenticated : failure
Authenticated --> Unauthenticated : logout()
Authenticated --> [*]

@enduml`

	// Store the diagram with TTL
	fmt.Printf("Storing diagram '%s'...\n", diagramName)
	err = redisCache.StoreDiagram(ctx, diagramName, pumlContent, 30*time.Minute)
	if err != nil {
		log.Fatalf("Failed to store diagram: %v", err)
	}
	fmt.Println("✓ Diagram stored successfully")

	// Retrieve the diagram
	fmt.Printf("Retrieving diagram '%s'...\n", diagramName)
	retrievedContent, err := redisCache.GetDiagram(ctx, diagramName)
	if err != nil {
		log.Fatalf("Failed to retrieve diagram: %v", err)
	}
	fmt.Println("✓ Diagram retrieved successfully")

	// Verify content matches
	if retrievedContent == pumlContent {
		fmt.Println("✓ Retrieved content matches original")
	} else {
		fmt.Println("✗ Content mismatch!")
		fmt.Printf("Original length: %d\n", len(pumlContent))
		fmt.Printf("Retrieved length: %d\n", len(retrievedContent))
	}

	// Display retrieved content
	fmt.Println("\nRetrieved PlantUML content:")
	fmt.Println("---")
	fmt.Println(retrievedContent)
	fmt.Println("---")

	// Test diagram with special characters in name
	specialDiagramName := "order/processing-flow_v2.1"
	specialContent := `@startuml
state "Order Received" as OR
state "Processing" as P
state "Completed" as C

OR --> P : validate()
P --> C : process()
@enduml`

	fmt.Printf("\nStoring diagram with special characters: '%s'...\n", specialDiagramName)
	err = redisCache.StoreDiagram(ctx, specialDiagramName, specialContent, time.Hour)
	if err != nil {
		log.Fatalf("Failed to store special diagram: %v", err)
	}
	fmt.Println("✓ Special diagram stored successfully")

	// Retrieve the special diagram
	retrievedSpecial, err := redisCache.GetDiagram(ctx, specialDiagramName)
	if err != nil {
		log.Fatalf("Failed to retrieve special diagram: %v", err)
	}
	fmt.Println("✓ Special diagram retrieved successfully")

	// Verify special content matches
	if retrievedSpecial == specialContent {
		fmt.Println("✓ Special diagram content matches original")
	}

	// Clean up - delete the diagrams
	fmt.Println("\nCleaning up...")
	err = redisCache.DeleteDiagram(ctx, diagramName)
	if err != nil {
		log.Printf("Warning: Failed to delete diagram '%s': %v", diagramName, err)
	} else {
		fmt.Printf("✓ Deleted diagram '%s'\n", diagramName)
	}

	err = redisCache.DeleteDiagram(ctx, specialDiagramName)
	if err != nil {
		log.Printf("Warning: Failed to delete diagram '%s': %v", specialDiagramName, err)
	} else {
		fmt.Printf("✓ Deleted diagram '%s'\n", specialDiagramName)
	}

	// Verify deletion
	_, err = redisCache.GetDiagram(ctx, diagramName)
	if err != nil && internal.IsNotFoundError(err) {
		fmt.Printf("✓ Confirmed diagram '%s' was deleted\n", diagramName)
	} else {
		fmt.Printf("✗ Diagram '%s' still exists or unexpected error: %v\n", diagramName, err)
	}

	fmt.Println("\nDiagram cache example completed successfully!")
}
