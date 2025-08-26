package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/cache"
	"github.com/kengibson1111/go-uml-statemachine-models/models"
)

// Sample PlantUML diagram for the state machine
const samplePUMLDiagram = `@startuml
!define PUML_VERSION 1.2024.8

state "Order Processing" as OrderProcessing {
  state "Pending" as pending
  state "Validated" as validated
  state "Shipped" as shipped
  state "Delivered" as delivered
  state "Cancelled" as cancelled
  
  [*] --> pending
  pending --> validated : validate
  pending --> cancelled : cancel
  validated --> shipped : ship
  validated --> cancelled : cancel
  shipped --> delivered : deliver
  shipped --> cancelled : cancel
}

OrderProcessing --> [*]
@enduml`

func main() {
	fmt.Println("=== Redis Cache State Machine Example ===")
	fmt.Println("This example demonstrates caching state machines with entities")
	fmt.Println()

	// Create Redis cache configuration
	config := cache.DefaultRedisConfig()
	config.RedisAddr = "localhost:6379"
	config.DefaultTTL = 2 * time.Hour

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

	// First, store the diagram (required before storing state machine)
	diagramName := "order-processing-workflow"
	umlVersion := "1.2024.8"

	fmt.Printf("2. Storing diagram '%s'...\n", diagramName)
	err = redisCache.StoreDiagram(ctx, models.DiagramTypePUML, diagramName, samplePUMLDiagram, 1*time.Hour)
	if err != nil {
		log.Fatalf("Failed to store diagram: %v", err)
	}
	fmt.Println("✓ Diagram stored successfully")
	fmt.Println()

	// Create a sample state machine with entities
	fmt.Println("3. Creating sample state machine with entities...")
	stateMachine := createSampleStateMachine()
	fmt.Printf("✓ Created state machine with %d regions and %d states\n",
		len(stateMachine.Regions), countStates(stateMachine))
	fmt.Println()

	// Store the state machine (this will also store individual entities)
	fmt.Printf("4. Storing state machine '%s' (version %s)...\n", diagramName, umlVersion)
	err = redisCache.StoreStateMachine(ctx, umlVersion, models.DiagramTypePUML, diagramName, stateMachine, 1*time.Hour)
	if err != nil {
		log.Fatalf("Failed to store state machine: %v", err)
	}
	fmt.Println("✓ State machine and entities stored successfully")
	fmt.Printf("Entity mappings created: %d\n", len(stateMachine.Entities))
	fmt.Println()

	// Retrieve the state machine
	fmt.Printf("5. Retrieving state machine '%s'...\n", diagramName)
	retrievedMachine, err := redisCache.GetStateMachine(ctx, umlVersion, diagramName)
	if err != nil {
		log.Fatalf("Failed to retrieve state machine: %v", err)
	}
	fmt.Println("✓ State machine retrieved successfully")
	fmt.Printf("Retrieved machine has %d regions and %d entity mappings\n",
		len(retrievedMachine.Regions), len(retrievedMachine.Entities))
	fmt.Println()

	// Demonstrate entity retrieval
	fmt.Println("6. Demonstrating entity retrieval...")
	if len(retrievedMachine.Entities) > 0 {
		// Get the first entity as an example
		var firstEntityID string
		for entityID := range retrievedMachine.Entities {
			firstEntityID = entityID
			break
		}

		fmt.Printf("Retrieving entity '%s'...\n", firstEntityID)
		entity, err := redisCache.GetEntity(ctx, umlVersion, diagramName, firstEntityID)
		if err != nil {
			log.Printf("Warning: Failed to retrieve entity: %v", err)
		} else {
			fmt.Printf("✓ Entity retrieved successfully (type: %T)\n", entity)
		}

		// Try type-safe retrieval if it's a state
		if state, err := redisCache.GetEntityAsState(ctx, umlVersion, diagramName, firstEntityID); err == nil {
			fmt.Printf("✓ Successfully retrieved as State: %s\n", state.Name)
		}
	}
	fmt.Println()

	// Demonstrate error handling
	fmt.Println("7. Testing error handling...")

	// Try to store state machine without diagram
	fmt.Println("Attempting to store state machine without corresponding diagram...")
	err = redisCache.StoreStateMachine(ctx, umlVersion, models.DiagramTypePUML, "non-existent-diagram", stateMachine, 1*time.Hour)
	if err != nil {
		if cache.IsValidationError(err) {
			fmt.Println("✓ Correctly prevented storing state machine without diagram")
		} else {
			fmt.Printf("✗ Unexpected error type: %v\n", err)
		}
	} else {
		fmt.Println("✗ Expected validation error but got none")
	}

	// Try to get non-existent state machine
	fmt.Println("Attempting to retrieve non-existent state machine...")
	_, err = redisCache.GetStateMachine(ctx, umlVersion, "non-existent-machine")
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

	// Demonstrate cache monitoring
	fmt.Println("8. Demonstrating cache monitoring...")

	// Get cache size information
	sizeInfo, err := redisCache.GetCacheSize(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get cache size: %v", err)
	} else {
		fmt.Printf("Cache statistics:\n")
		fmt.Printf("  Total keys: %d\n", sizeInfo.TotalKeys)
		fmt.Printf("  Diagrams: %d\n", sizeInfo.DiagramCount)
		fmt.Printf("  State machines: %d\n", sizeInfo.StateMachineCount)
		fmt.Printf("  Entities: %d\n", sizeInfo.EntityCount)
	}
	fmt.Println()

	// Cleanup - delete state machine (cascade delete will remove entities)
	fmt.Printf("9. Cleaning up state machine '%s'...\n", diagramName)
	err = redisCache.DeleteStateMachine(ctx, umlVersion, diagramName)
	if err != nil {
		log.Printf("Warning: Failed to delete state machine: %v", err)
	} else {
		fmt.Println("✓ State machine and entities deleted successfully")
	}

	// Cleanup diagram
	fmt.Printf("Cleaning up diagram '%s'...\n", diagramName)
	err = redisCache.DeleteDiagram(ctx, models.DiagramTypePUML, diagramName)
	if err != nil {
		log.Printf("Warning: Failed to delete diagram: %v", err)
	} else {
		fmt.Println("✓ Diagram deleted successfully")
	}

	fmt.Println()
	fmt.Println("=== State Machine Cache Example Complete ===")
}

// createSampleStateMachine creates a sample state machine for demonstration
func createSampleStateMachine() *models.StateMachine {
	// Create states
	pendingState := &models.State{
		Vertex: models.Vertex{
			ID:   "pending",
			Name: "Pending",
			Type: "state",
		},
		IsSimple: true,
	}

	validatedState := &models.State{
		Vertex: models.Vertex{
			ID:   "validated",
			Name: "Validated",
			Type: "state",
		},
		IsSimple: true,
	}

	shippedState := &models.State{
		Vertex: models.Vertex{
			ID:   "shipped",
			Name: "Shipped",
			Type: "state",
		},
		IsSimple: true,
	}

	deliveredState := &models.State{
		Vertex: models.Vertex{
			ID:   "delivered",
			Name: "Delivered",
			Type: "state",
		},
		IsSimple: true,
	}

	cancelledState := &models.State{
		Vertex: models.Vertex{
			ID:   "cancelled",
			Name: "Cancelled",
			Type: "state",
		},
		IsSimple: true,
	}

	// Create transitions
	validateTransition := &models.Transition{
		ID:     "t1",
		Name:   "validate",
		Source: &pendingState.Vertex,
		Target: &validatedState.Vertex,
		Kind:   models.TransitionKindExternal,
	}

	cancelFromPendingTransition := &models.Transition{
		ID:     "t2",
		Name:   "cancel",
		Source: &pendingState.Vertex,
		Target: &cancelledState.Vertex,
		Kind:   models.TransitionKindExternal,
	}

	shipTransition := &models.Transition{
		ID:     "t3",
		Name:   "ship",
		Source: &validatedState.Vertex,
		Target: &shippedState.Vertex,
		Kind:   models.TransitionKindExternal,
	}

	deliverTransition := &models.Transition{
		ID:     "t4",
		Name:   "deliver",
		Source: &shippedState.Vertex,
		Target: &deliveredState.Vertex,
		Kind:   models.TransitionKindExternal,
	}

	// Create main region
	mainRegion := &models.Region{
		ID:   "main-region",
		Name: "Order Processing Region",
		States: []*models.State{
			pendingState,
			validatedState,
			shippedState,
			deliveredState,
			cancelledState,
		},
		Transitions: []*models.Transition{
			validateTransition,
			cancelFromPendingTransition,
			shipTransition,
			deliverTransition,
		},
		Vertices: []*models.Vertex{
			&pendingState.Vertex,
			&validatedState.Vertex,
			&shippedState.Vertex,
			&deliveredState.Vertex,
			&cancelledState.Vertex,
		},
	}

	// Create state machine
	stateMachine := &models.StateMachine{
		ID:       "order-processing-sm",
		Name:     "Order Processing State Machine",
		Version:  "1.0.0",
		Regions:  []*models.Region{mainRegion},
		Entities: make(map[string]string), // Will be populated by cache
		Metadata: map[string]interface{}{
			"description": "Sample order processing workflow",
			"author":      "Cache Example",
			"created":     time.Now().Format(time.RFC3339),
		},
		CreatedAt: time.Now(),
	}

	return stateMachine
}

// countStates counts the total number of states in a state machine
func countStates(sm *models.StateMachine) int {
	count := 0
	for _, region := range sm.Regions {
		count += len(region.States)
	}
	return count
}
