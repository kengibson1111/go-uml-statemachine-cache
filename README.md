# go-uml-statemachine-cache

A Go library providing Redis-based caching for PlantUML diagrams and UML state machine definitions with comprehensive error handling, retry mechanisms, and versioned storage.

## Features

- **Diagram Caching**: Store and retrieve PlantUML diagram content with TTL support
- **State Machine Caching**: Cache parsed UML state machine objects with version management
- **Entity Caching**: Store individual state machine entities with hierarchical keys
- **Redis Backend**: High-performance Redis-based storage with connection pooling
- **Error Handling**: Comprehensive error types with detailed context
- **Retry Logic**: Configurable exponential backoff with jitter
- **Health Monitoring**: Built-in health checks and connection monitoring
- **Thread Safe**: Concurrent access support with proper synchronization

## Installation

```bash
go get github.com/kengibson1111/go-uml-statemachine-cache
```

## Quick Start

```go
package main

import (
    "context"
    "time"
    "github.com/kengibson1111/go-uml-statemachine-cache/cache"
)

func main() {
    // Create cache with default configuration
    config := cache.DefaultRedisConfig()
    config.RedisAddr = "localhost:6379"
    
    redisCache, err := cache.NewRedisCache(config)
    if err != nil {
        panic(err)
    }
    defer redisCache.Close()
    
    ctx := context.Background()
    
    // Store a PlantUML diagram
    pumlContent := `@startuml
    [*] --> State1
    State1 --> [*]
    @enduml`
    
    err = redisCache.StoreDiagram(ctx, "my-diagram", pumlContent, time.Hour)
    if err != nil {
        panic(err)
    }
    
    // Retrieve the diagram
    retrieved, err := redisCache.GetDiagram(ctx, "my-diagram")
    if err != nil {
        panic(err)
    }
    
    println("Retrieved:", retrieved)
}
```

## Configuration

The library provides flexible configuration options:

```go
config := &cache.RedisConfig{
    RedisAddr:    "localhost:6379",
    RedisDB:      0,
    MaxRetries:   3,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    PoolSize:     10,
    DefaultTTL:   time.Hour,
}
```

## Error Handling

The library provides typed errors for better error handling:

```go
_, err := redisCache.GetDiagram(ctx, "nonexistent")
if cache.IsNotFoundError(err) {
    // Handle not found case
} else if cache.IsConnectionError(err) {
    // Handle connection issues
}
```

## Examples

The repository includes several examples:

### Public Examples (using public APIs)
- `examples/diagram_cache_example/` - Basic diagram caching operations
- `examples/state_machine_cache_example/` - State machine storage and retrieval

### Internal Examples (using internal APIs)
- `internal/examples/redis_client_example/` - Low-level Redis client usage
- `internal/examples/retry_example/` - Retry mechanism demonstration

## Testing

The library includes comprehensive test coverage:

```bash
# Run unit tests (no Redis required)
go test ./cache -v

# Run integration tests (requires Redis)
go test ./test/integration -v
```

## Project Structure

```
├── cache/                    # Public cache interface and Redis implementation
├── internal/                 # Internal utilities and low-level components
│   ├── examples/            # Internal API usage examples
│   ├── errors.go            # Error types and utilities
│   ├── keygen.go            # Key generation utilities
│   ├── redis_client.go      # Low-level Redis client
│   └── validation_test.go   # Validation tests
├── examples/                # Public API usage examples
├── test/integration/        # Integration tests requiring Redis
└── original-specs/          # UML specification files
```

## License

This project is licensed under the terms specified in the LICENSE file.
