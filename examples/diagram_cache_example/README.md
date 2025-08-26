# Diagram Cache Example

This example demonstrates how to use the Redis cache system to store and retrieve PlantUML diagrams.

## Features Demonstrated

- Creating a Redis cache instance with configuration
- Storing PlantUML diagrams with TTL (time-to-live)
- Retrieving cached diagrams
- Error handling for non-existent diagrams
- Cleanup operations using patterns
- Proper resource management

## Prerequisites

- Redis server running on localhost:6379
- Go 1.24.4 or later

## Running the Example

### On Windows:

```cmd
cd examples\diagram_cache_example
go run main.go
```

### Expected Output

The example will:
1. Test Redis connection
2. Store a sample PlantUML diagram
3. Retrieve and verify the diagram
4. Demonstrate error handling
5. Show cleanup operations
6. Clean up resources

## Key Concepts

### Configuration
```go
config := cache.DefaultRedisConfig()
config.RedisAddr = "localhost:6379"
config.DefaultTTL = 1 * time.Hour
```

### Storing Diagrams
```go
err = redisCache.StoreDiagram(ctx, models.DiagramTypePUML, diagramName, pumlContent, 30*time.Minute)
```

### Retrieving Diagrams
```go
diagram, err := redisCache.GetDiagram(ctx, models.DiagramTypePUML, diagramName)
```

### Error Handling
```go
if cache.IsNotFoundError(err) {
    // Handle cache miss
}
```

### Cleanup for PUML diagrams
```go
err = redisCache.Cleanup(ctx, "/diagrams/puml/pattern-*")
```

## Troubleshooting

If you encounter connection errors:
1. Ensure Redis is running: `redis-server`
2. Check Redis is accessible: `redis-cli ping`
3. Verify the Redis address in the configuration