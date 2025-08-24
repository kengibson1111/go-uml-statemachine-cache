# go-uml-statemachine-cache

A comprehensive Go library providing Redis-based caching for PlantUML diagrams and UML state machine definitions with enterprise-grade error handling, retry mechanisms, health monitoring, and versioned storage.

## Features

### Core Functionality
- **Diagram Caching**: Store and retrieve PlantUML diagram content with configurable TTL support
- **State Machine Caching**: Cache parsed UML state machine objects with version management and entity relationships
- **Entity Caching**: Store individual state machine entities with hierarchical keys and referential integrity
- **Redis Backend**: High-performance Redis-based storage with connection pooling and cluster support

### Reliability & Performance
- **Error Handling**: Comprehensive typed error system with detailed context and recovery strategies
- **Retry Logic**: Configurable exponential backoff with jitter and circuit breaker patterns
- **Health Monitoring**: Built-in health checks, performance metrics, and diagnostic tools
- **Thread Safety**: Full concurrent access support with proper synchronization
- **Input Validation**: Comprehensive security-focused input sanitization and validation

### Monitoring & Management
- **Cache Cleanup**: Pattern-based cleanup with batch operations and metrics collection
- **Size Monitoring**: Real-time cache size and memory usage tracking
- **Performance Metrics**: Detailed Redis performance and connection pool statistics
- **Diagnostics**: Built-in troubleshooting tools and configuration validation

## Installation

```bash
go get github.com/kengibson1111/go-uml-statemachine-cache
```

### Prerequisites

- Go 1.24.4 or later
- Redis 6.0 or later
- Windows 10/11 or Windows Server 2019/2022 (for Windows users)

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/kengibson1111/go-uml-statemachine-cache/cache"
)

func main() {
    // Create cache with default configuration
    config := cache.DefaultRedisConfig()
    config.RedisAddr = "localhost:6379"
    
    redisCache, err := cache.NewRedisCache(config)
    if err != nil {
        log.Fatal("Failed to create cache:", err)
    }
    defer redisCache.Close()
    
    ctx := context.Background()
    
    // Store a PlantUML diagram
    pumlContent := `@startuml
    [*] --> Idle
    Idle --> Processing : start
    Processing --> Idle : complete
    Processing --> Error : fail
    Error --> Idle : reset
    @enduml`
    
    err = redisCache.StoreDiagram(ctx, "workflow-diagram", pumlContent, time.Hour)
    if err != nil {
        log.Fatal("Failed to store diagram:", err)
    }
    
    // Retrieve the diagram
    retrieved, err := redisCache.GetDiagram(ctx, "workflow-diagram")
    if err != nil {
        log.Fatal("Failed to retrieve diagram:", err)
    }
    
    fmt.Println("Retrieved diagram:", retrieved)
}
```

### Advanced Configuration

```go
config := &cache.RedisConfig{
    // Redis connection settings
    RedisAddr:     "localhost:6379",
    RedisPassword: "",                    // Set if Redis requires authentication
    RedisDB:       0,                     // Redis database number (0-15)
    
    // Connection pool settings
    MaxRetries:   3,                      // Maximum retry attempts
    DialTimeout:  5 * time.Second,        // Connection establishment timeout
    ReadTimeout:  3 * time.Second,        // Read operation timeout
    WriteTimeout: 3 * time.Second,        // Write operation timeout
    PoolSize:     10,                     // Maximum connections in pool
    
    // Cache settings
    DefaultTTL:   24 * time.Hour,         // Default expiration time
    
    // Retry configuration
    RetryConfig: &cache.RedisRetryConfig{
        MaxAttempts:  5,                  // Maximum retry attempts
        InitialDelay: 100 * time.Millisecond, // Initial retry delay
        MaxDelay:     5 * time.Second,    // Maximum retry delay
        Multiplier:   2.0,                // Exponential backoff multiplier
        Jitter:       true,               // Add random jitter to delays
        RetryableOps: []string{"get", "set", "del", "exists"}, // Operations to retry
    },
}

redisCache, err := cache.NewRedisCache(config)
if err != nil {
    log.Fatal("Failed to create cache:", err)
}
```

## Comprehensive Error Handling

The library provides a sophisticated error handling system with typed errors and recovery strategies:

```go
_, err := redisCache.GetDiagram(ctx, "nonexistent-diagram")
if err != nil {
    switch {
    case cache.IsNotFoundError(err):
        fmt.Println("Diagram not found in cache")
        // Handle cache miss - maybe load from source
        
    case cache.IsConnectionError(err):
        fmt.Println("Redis connection error:", err)
        // Handle connection issues - maybe use fallback
        
    case cache.IsValidationError(err):
        fmt.Println("Invalid input provided:", err)
        // Handle validation errors - fix input
        
    case cache.IsTimeoutError(err):
        fmt.Println("Operation timed out:", err)
        // Handle timeout - maybe retry with longer timeout
        
    case cache.IsRetryExhaustedError(err):
        fmt.Println("All retry attempts failed:", err)
        // Handle retry exhaustion - maybe use circuit breaker
        
    default:
        fmt.Println("Unexpected error:", err)
        // Handle other errors
    }
    
    // Get error severity and recovery strategy
    severity := cache.GetErrorSeverity(err)
    strategy := cache.GetRecoveryStrategy(err)
    
    fmt.Printf("Error severity: %v, Recovery strategy: %v\n", severity, strategy)
}
```

## State Machine Caching

Store and retrieve complex state machines with automatic entity relationship management:

```go
import "github.com/kengibson1111/go-uml-statemachine-models/models"

// Assuming you have a parsed state machine
var stateMachine *models.StateMachine

// Store state machine (automatically creates entity cache entries)
err = redisCache.StoreStateMachine(ctx, "v1.0", "workflow-diagram", stateMachine, time.Hour)
if err != nil {
    log.Fatal("Failed to store state machine:", err)
}

// Retrieve state machine
retrieved, err := redisCache.GetStateMachine(ctx, "v1.0", "workflow-diagram")
if err != nil {
    log.Fatal("Failed to retrieve state machine:", err)
}

// Access individual entities
entity, err := redisCache.GetEntity(ctx, "v1.0", "workflow-diagram", "idle-state")
if err != nil {
    log.Fatal("Failed to retrieve entity:", err)
}

// Type-safe entity retrieval
state, err := redisCache.GetEntityAsState(ctx, "v1.0", "workflow-diagram", "idle-state")
if err != nil {
    log.Fatal("Failed to retrieve state:", err)
}

transition, err := redisCache.GetEntityAsTransition(ctx, "v1.0", "workflow-diagram", "start-transition")
if err != nil {
    log.Fatal("Failed to retrieve transition:", err)
}
```

## Health Monitoring and Diagnostics

Monitor cache health and performance with built-in tools:

```go
// Basic health check
err = redisCache.Health(ctx)
if err != nil {
    log.Println("Cache is unhealthy:", err)
}

// Detailed health status with metrics
health, err := redisCache.HealthDetailed(ctx)
if err != nil {
    log.Fatal("Failed to get health status:", err)
}

fmt.Printf("Cache Status: %s\n", health.Status)
fmt.Printf("Response Time: %v\n", health.ResponseTime)
fmt.Printf("Redis Version: %s\n", health.Performance.ServerInfo.RedisVersion)
fmt.Printf("Memory Usage: %s\n", health.Performance.MemoryUsage.UsedMemoryHuman)
fmt.Printf("Hit Rate: %.2f%%\n", health.Performance.KeyspaceInfo.HitRate)

// Connection health
connHealth, err := redisCache.GetConnectionHealth(ctx)
if err != nil {
    log.Fatal("Failed to get connection health:", err)
}

fmt.Printf("Connected: %v\n", connHealth.Connected)
fmt.Printf("Ping Latency: %v\n", connHealth.PingLatency)
fmt.Printf("Pool Stats - Total: %d, Idle: %d, Active: %d\n",
    connHealth.PoolStats.TotalConnections,
    connHealth.PoolStats.IdleConnections,
    connHealth.PoolStats.ActiveConnections)

// Run diagnostics
diagnostics, err := redisCache.RunDiagnostics(ctx)
if err != nil {
    log.Fatal("Failed to run diagnostics:", err)
}

if !diagnostics.ConfigurationCheck.Valid {
    fmt.Println("Configuration issues found:")
    for _, issue := range diagnostics.ConfigurationCheck.Issues {
        fmt.Printf("  - %s\n", issue)
    }
}

if len(diagnostics.Recommendations) > 0 {
    fmt.Println("Performance recommendations:")
    for _, rec := range diagnostics.Recommendations {
        fmt.Printf("  - %s\n", rec)
    }
}
```

## Cache Management and Cleanup

Efficiently manage cache size and cleanup with advanced options:

```go
// Basic cleanup with pattern
err = redisCache.Cleanup(ctx, "/diagrams/puml/test-*")
if err != nil {
    log.Fatal("Cleanup failed:", err)
}

// Advanced cleanup with detailed options
options := &cache.CleanupOptions{
    BatchSize:      100,                  // Keys to delete per batch
    ScanCount:      1000,                 // Keys to scan per operation
    MaxKeys:        10000,                // Maximum keys to delete (0 = no limit)
    DryRun:         false,                // Set to true to simulate cleanup
    Timeout:        30 * time.Second,     // Total cleanup timeout
    CollectMetrics: true,                 // Collect detailed metrics
}

result, err := redisCache.CleanupWithOptions(ctx, "/machines/v1.0/*", options)
if err != nil {
    log.Fatal("Advanced cleanup failed:", err)
}

fmt.Printf("Cleanup Results:\n")
fmt.Printf("  Keys Scanned: %d\n", result.KeysScanned)
fmt.Printf("  Keys Deleted: %d\n", result.KeysDeleted)
fmt.Printf("  Bytes Freed: %d\n", result.BytesFreed)
fmt.Printf("  Duration: %v\n", result.Duration)
fmt.Printf("  Batches Used: %d\n", result.BatchesUsed)

// Monitor cache size
sizeInfo, err := redisCache.GetCacheSize(ctx)
if err != nil {
    log.Fatal("Failed to get cache size:", err)
}

fmt.Printf("Cache Size Information:\n")
fmt.Printf("  Total Keys: %d\n", sizeInfo.TotalKeys)
fmt.Printf("  Diagrams: %d\n", sizeInfo.DiagramCount)
fmt.Printf("  State Machines: %d\n", sizeInfo.StateMachineCount)
fmt.Printf("  Entities: %d\n", sizeInfo.EntityCount)
fmt.Printf("  Memory Used: %d bytes\n", sizeInfo.MemoryUsed)
```

## Windows-Specific Setup and Usage

### Redis Installation on Windows

1. **Using Chocolatey** (recommended):
   ```cmd
   choco install redis-64
   ```

2. **Using Windows Subsystem for Linux (WSL)**:
   ```cmd
   wsl --install
   wsl
   sudo apt update
   sudo apt install redis-server
   redis-server
   ```

3. **Using Docker Desktop**:
   ```cmd
   docker run --name redis-cache -p 6379:6379 -d redis:latest
   ```

### Running Examples on Windows

Navigate to the project directory and run examples using Go directly:

```cmd
# Navigate to project directory
cd go-uml-statemachine-cache

# Run diagram caching example
go run examples\diagram_cache_example\main.go

# Run state machine caching example
go run examples\state_machine_cache_example\main.go

# Run health monitoring example
go run examples\health_monitoring_example\main.go
```

### Testing on Windows

```cmd
# Run unit tests (no Redis required)
go test .\cache -v

# Run integration tests (requires Redis running)
go test .\test\integration -v

# Run all tests with coverage
go test -cover .\cache .\internal

# Run specific test
go test -run TestRedisCache_StoreDiagram .\cache -v
```

### Building on Windows

```cmd
# Build the library
go build .\cache

# Build examples
go build -o diagram_example.exe .\examples\diagram_cache_example
go build -o statemachine_example.exe .\examples\state_machine_cache_example

# Run built examples
.\diagram_example.exe
.\statemachine_example.exe
```

## Examples and Documentation

The repository includes comprehensive examples demonstrating all features:

### Public API Examples
- **`examples\diagram_cache_example\`** - Basic diagram caching operations with error handling
- **`examples\state_machine_cache_example\`** - State machine and entity caching workflows
- **`examples\health_monitoring_example\`** - Health monitoring and diagnostics
- **`examples\cleanup_example\`** - Cache cleanup and management operations
- **`examples\error_handling_example\`** - Comprehensive error handling patterns

### Internal API Examples (for library developers)
- **`internal\examples\redis_client_example\`** - Low-level Redis client usage
- **`internal\examples\retry_example\`** - Retry mechanism and circuit breaker demonstration

## API Reference

### Public API (cache package)

The public API provides all functionality needed for application development:

#### Core Interfaces
- `Cache` - Main caching interface with all operations
- `RedisCache` - Redis-based implementation of Cache interface

#### Configuration Types
- `RedisConfig` - Redis connection and cache configuration
- `RedisRetryConfig` - Retry logic configuration
- `CleanupOptions` - Advanced cleanup operation configuration

#### Error Types and Functions
- `CacheError` - Comprehensive error type with context
- `CacheErrorType` - Error type enumeration
- `IsConnectionError()`, `IsNotFoundError()`, etc. - Error type checking functions

#### Health and Monitoring Types
- `HealthStatus` - Detailed health information
- `ConnectionHealth` - Connection-specific health data
- `PerformanceMetrics` - Performance and usage metrics
- `CleanupResult` - Cleanup operation results
- `CacheSizeInfo` - Cache size and usage information

### Internal API (internal package)

The internal API contains implementation details and should not be used directly:

#### Low-Level Components
- `RedisClientInterface` - Low-level Redis operations
- `KeyGenerator` - Cache key generation and validation
- `InputValidator` - Input validation and sanitization
- `ErrorRecoveryManager` - Error recovery and circuit breaker logic

**Important**: Only use the public cache package API. The internal package is for implementation details and may change without notice.

## Testing

The library includes comprehensive test coverage with multiple test types:

### Unit Tests
```cmd
# Run all unit tests (no external dependencies)
go test .\cache -v

# Run with coverage
go test -cover .\cache

# Run specific test
go test -run TestRedisCache_StoreDiagram .\cache -v
```

### Integration Tests
```cmd
# Requires Redis server running on localhost:6379
go test .\test\integration -v

# Run with custom Redis address
REDIS_ADDR=localhost:6380 go test .\test\integration -v
```

### Performance Tests
```cmd
# Run performance benchmarks
go test -bench=. .\cache

# Run stress tests
go test -run TestStress .\test\integration -v
```

## Performance Optimization

The library is optimized for high-performance scenarios:

### Connection Management
- Connection pooling with configurable pool size
- Automatic connection health monitoring
- Graceful connection recovery and retry logic

### Batch Operations
- Bulk deletion operations for efficient cleanup
- Pipelined Redis commands where appropriate
- Configurable batch sizes for optimal performance

### Memory Optimization
- Lazy loading for large state machine entities
- Efficient JSON serialization with minimal overhead
- Memory usage monitoring and reporting

### Caching Strategies
- Configurable TTL values for different data types
- Hierarchical key structure for efficient lookups
- Entity relationship caching for complex state machines

## Security Features

The library includes comprehensive security measures:

### Input Validation
- Length validation for all string inputs
- Character encoding validation (UTF-8)
- Control character filtering
- Path traversal prevention

### Security Threat Detection
- SQL injection pattern detection
- XSS (Cross-Site Scripting) pattern detection
- Protocol-based attack prevention
- Script injection detection

### Key Security
- Safe key generation with URL encoding
- Special character sanitization
- Key format validation
- Maximum key length enforcement

## Project Structure

```
go-uml-statemachine-cache/
├── cache/                           # Public API - Main cache interface and implementation
│   ├── cache.go                    # Cache interface definition
│   ├── redis_cache.go              # Redis implementation
│   └── redis_cache_test.go         # Unit tests
├── internal/                        # Internal API - Implementation details
│   ├── errors.go                   # Error types and recovery logic
│   ├── keygen.go                   # Key generation and validation
│   ├── redis_client.go             # Low-level Redis client wrapper
│   ├── validation.go               # Input validation and sanitization
│   └── examples/                   # Internal API usage examples
├── examples/                        # Public API usage examples
│   ├── diagram_cache_example/      # Basic diagram caching
│   ├── state_machine_cache_example/ # State machine caching
│   ├── health_monitoring_example/  # Health monitoring
│   ├── cleanup_example/            # Cache cleanup
│   └── error_handling_example/     # Error handling patterns
├── test/                           # Test suites
│   └── integration/                # Integration tests requiring Redis
├── original-specs/                 # UML specification files
├── doc.go                          # Package documentation
├── README.md                       # This file
├── go.mod                          # Go module definition
└── LICENSE                         # License file
```

## Dependencies

The library has minimal external dependencies:

### Required Dependencies
- **github.com/redis/go-redis/v9** - Official Redis client for Go
- **github.com/kengibson1111/go-uml-statemachine-models** - UML state machine data models

### Development Dependencies
- **github.com/stretchr/testify** - Testing framework and assertions

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork the repository** and create a feature branch
2. **Write tests** for new functionality
3. **Follow Go conventions** and run `go fmt`
4. **Update documentation** for API changes
5. **Test on Windows** if making platform-specific changes
6. **Submit a pull request** with a clear description

### Development Setup on Windows

```cmd
# Clone the repository
git clone https://github.com/kengibson1111/go-uml-statemachine-cache.git
cd go-uml-statemachine-cache

# Install dependencies
go mod download

# Run tests
go test .\cache -v

# Install Redis for integration tests
choco install redis-64

# Run integration tests
go test .\test\integration -v
```

## License

This project is licensed under the terms specified in the LICENSE file.

## Support

For questions, issues, or contributions:

1. **Check the examples** in the `examples/` directory
2. **Review the documentation** in `doc.go`
3. **Search existing issues** on GitHub
4. **Create a new issue** with detailed information
5. **Include system information** (OS, Go version, Redis version) for bug reports

## Changelog

See the Git commit history for detailed changes. Major version changes will include migration guides and breaking change documentation.
