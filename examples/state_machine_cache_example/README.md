# State Machine Cache Example

This example demonstrates how to use the Redis cache system to store and retrieve parsed state machines with their entities.

## Features Demonstrated

- Storing PlantUML diagrams as prerequisites for state machines
- Creating and storing complex state machines with multiple states and transitions
- Automatic entity extraction and caching
- Retrieving state machines and individual entities
- Type-safe entity retrieval methods
- Error handling and validation
- Cache monitoring and statistics
- Cascade deletion of state machines and entities

## Prerequisites

- Redis server running on localhost:6379
- Go 1.24.4 or later

## Running the Example

### On Windows:

```cmd
cd examples\state_machine_cache_example
go run main.go
```

### Expected Output

The example will:
1. Test Redis connection
2. Store a PlantUML diagram
3. Create and store a state machine with entities
4. Retrieve and verify the state machine
5. Demonstrate entity retrieval
6. Show error handling scenarios
7. Display cache monitoring information
8. Clean up all resources

## Key Concepts

### State Machine Storage
```go
// Diagram must be stored first
err = redisCache.StoreDiagram(ctx, diagramName, pumlContent, ttl)

// Then store the state machine (entities are automatically extracted)
err = redisCache.StoreStateMachine(ctx, umlVersion, diagramName, stateMachine, ttl)
```

### Entity Retrieval
```go
// Generic entity retrieval
entity, err := redisCache.GetEntity(ctx, umlVersion, diagramName, entityID)

// Type-safe retrieval
state, err := redisCache.GetEntityAsState(ctx, umlVersion, diagramName, entityID)
```

### Cache Monitoring
```go
sizeInfo, err := redisCache.GetCacheSize(ctx)
fmt.Printf("Total keys: %d, Entities: %d\n", sizeInfo.TotalKeys, sizeInfo.EntityCount)
```

### Cascade Deletion
```go
// Deletes state machine and all its entities
err = redisCache.DeleteStateMachine(ctx, umlVersion, diagramName)
```

## State Machine Structure

The example creates a sample order processing workflow with:

- **States**: Pending, Validated, Shipped, Delivered, Cancelled
- **Transitions**: validate, cancel, ship, deliver
- **Entities**: All states, transitions, and regions are cached as individual entities
- **Referential Integrity**: Entity mappings maintain relationships between components

## Error Handling

The example demonstrates several error scenarios:

1. **Validation Errors**: Attempting to store state machine without corresponding diagram
2. **Not Found Errors**: Retrieving non-existent state machines or entities
3. **Connection Errors**: Handled with retry logic and proper error reporting

## Cache Key Structure

The cache uses a hierarchical key structure:

- Diagrams: `/diagrams/puml/{diagram-name}`
- State Machines: `/machines/{uml-version}/{diagram-name}`
- Entities: `/machines/{uml-version}/{diagram-name}/entities/{entity-id}`

## Troubleshooting

If you encounter issues:

1. **Redis Connection**: Ensure Redis is running and accessible
2. **Memory Usage**: Monitor cache size for large state machines
3. **TTL Settings**: Adjust time-to-live values based on your needs
4. **Entity Relationships**: Verify entity mappings are correctly maintained