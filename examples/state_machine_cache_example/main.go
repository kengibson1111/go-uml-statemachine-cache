package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-models/models"

	"github.com/kengibson1111/go-uml-statemachine-cache/cache"
	"github.com/kengibson1111/go-uml-statemachine-cache/internal"
)

func main() {
	// Create cache configuration
	config := &internal.Config{
		RedisAddr:    "localhost:6379",
		RedisDB:      0,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		DefaultTTL:   time.Hour,
	}

	// Create Redis cache instance
	redisCache, err := cache.NewRedisCache(config)
	if err != nil {
		log.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer redisCache.Close()

	ctx := context.Background()

	// Check cache health
	if err := redisCache.Health(ctx); err != nil {
		log.Printf("Warning: Redis health check failed: %v", err)
		log.Println("Make sure Redis is running on localhost:6379")
		return
	}

	fmt.Println("=== State Machine Cache Example ===")

	// Create a sample state machine
	stateMachine := createSampleStateMachine()

	// Store the state machine
	umlVersion := "2.5.1"
	machineName := "OrderProcessing"
	ttl := 2 * time.Hour

	fmt.Printf("Storing state machine '%s' (version %s)...\n", machineName, umlVersion)
	err = redisCache.StoreStateMachine(ctx, umlVersion, machineName, stateMachine, ttl)
	if err != nil {
		log.Fatalf("Failed to store state machine: %v", err)
	}
	fmt.Println("✓ State machine stored successfully")

	// Retrieve the state machine
	fmt.Printf("Retrieving state machine '%s' (version %s)...\n", machineName, umlVersion)
	retrievedMachine, err := redisCache.GetStateMachine(ctx, umlVersion, machineName)
	if err != nil {
		log.Fatalf("Failed to retrieve state machine: %v", err)
	}
	fmt.Println("✓ State machine retrieved successfully")

	// Display retrieved state machine details
	fmt.Printf("\nRetrieved State Machine Details:\n")
	fmt.Printf("  ID: %s\n", retrievedMachine.ID)
	fmt.Printf("  Name: %s\n", retrievedMachine.Name)
	fmt.Printf("  Version: %s\n", retrievedMachine.Version)
	fmt.Printf("  Regions: %d\n", len(retrievedMachine.Regions))
	fmt.Printf("  Entities: %d\n", len(retrievedMachine.Entities))

	if len(retrievedMachine.Regions) > 0 {
		region := retrievedMachine.Regions[0]
		fmt.Printf("  First Region: %s (States: %d, Transitions: %d)\n",
			region.Name, len(region.States), len(region.Transitions))
	}

	// Demonstrate JSON serialization
	fmt.Printf("\nJSON Serialization Example:\n")
	jsonData, err := json.MarshalIndent(retrievedMachine, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal state machine to JSON: %v", err)
	}
	fmt.Printf("State machine as JSON (first 500 chars):\n%s...\n",
		string(jsonData[:min(500, len(jsonData))]))

	// Test versioned storage - store same machine with different version
	newVersion := "3.0.0"
	fmt.Printf("\nStoring same machine with different version (%s)...\n", newVersion)
	err = redisCache.StoreStateMachine(ctx, newVersion, machineName, stateMachine, ttl)
	if err != nil {
		log.Fatalf("Failed to store versioned state machine: %v", err)
	}
	fmt.Println("✓ Versioned state machine stored successfully")

	// Verify both versions exist
	fmt.Printf("Verifying both versions exist:\n")

	// Check version 2.5.1
	_, err = redisCache.GetStateMachine(ctx, umlVersion, machineName)
	if err != nil {
		log.Fatalf("Failed to retrieve original version: %v", err)
	}
	fmt.Printf("  ✓ Version %s exists\n", umlVersion)

	// Check version 3.0.0
	_, err = redisCache.GetStateMachine(ctx, newVersion, machineName)
	if err != nil {
		log.Fatalf("Failed to retrieve new version: %v", err)
	}
	fmt.Printf("  ✓ Version %s exists\n", newVersion)

	// Clean up - delete the state machines
	fmt.Printf("\nCleaning up...\n")
	err = redisCache.DeleteStateMachine(ctx, umlVersion, machineName)
	if err != nil {
		log.Printf("Warning: Failed to delete state machine (v%s): %v", umlVersion, err)
	} else {
		fmt.Printf("  ✓ Deleted version %s\n", umlVersion)
	}

	err = redisCache.DeleteStateMachine(ctx, newVersion, machineName)
	if err != nil {
		log.Printf("Warning: Failed to delete state machine (v%s): %v", newVersion, err)
	} else {
		fmt.Printf("  ✓ Deleted version %s\n", newVersion)
	}

	fmt.Println("\n=== Example completed successfully ===")
}

func createSampleStateMachine() *models.StateMachine {
	return &models.StateMachine{
		ID:      "order-processing-sm",
		Name:    "OrderProcessing",
		Version: "1.0",
		Regions: []*models.Region{
			{
				ID:   "main-region",
				Name: "MainRegion",
				States: []*models.State{
					{
						Vertex: models.Vertex{
							ID:   "pending",
							Name: "Pending",
							Type: "state",
						},
						IsSimple: true,
					},
					{
						Vertex: models.Vertex{
							ID:   "processing",
							Name: "Processing",
							Type: "state",
						},
						IsSimple: true,
					},
					{
						Vertex: models.Vertex{
							ID:   "completed",
							Name: "Completed",
							Type: "finalstate",
						},
						IsSimple: true,
					},
					{
						Vertex: models.Vertex{
							ID:   "cancelled",
							Name: "Cancelled",
							Type: "finalstate",
						},
						IsSimple: true,
					},
				},
				Transitions: []*models.Transition{
					{
						ID:   "start-processing",
						Name: "StartProcessing",
						Kind: models.TransitionKindExternal,
						Triggers: []*models.Trigger{
							{
								ID:   "process-trigger",
								Name: "ProcessOrder",
								Event: &models.Event{
									ID:   "process-event",
									Name: "ProcessOrderEvent",
									Type: models.EventTypeSignal,
								},
							},
						},
					},
					{
						ID:   "complete-order",
						Name: "CompleteOrder",
						Kind: models.TransitionKindExternal,
					},
					{
						ID:   "cancel-order",
						Name: "CancelOrder",
						Kind: models.TransitionKindExternal,
					},
				},
			},
		},
		Entities: map[string]string{
			"pending":          "/machines/2.5.1/OrderProcessing/entities/pending",
			"processing":       "/machines/2.5.1/OrderProcessing/entities/processing",
			"completed":        "/machines/2.5.1/OrderProcessing/entities/completed",
			"cancelled":        "/machines/2.5.1/OrderProcessing/entities/cancelled",
			"start-processing": "/machines/2.5.1/OrderProcessing/entities/start-processing",
			"complete-order":   "/machines/2.5.1/OrderProcessing/entities/complete-order",
			"cancel-order":     "/machines/2.5.1/OrderProcessing/entities/cancel-order",
		},
		Metadata: map[string]interface{}{
			"created_by":   "state_machine_example",
			"description":  "Sample order processing state machine",
			"domain":       "e-commerce",
			"complexity":   "simple",
			"states_count": 4,
		},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
