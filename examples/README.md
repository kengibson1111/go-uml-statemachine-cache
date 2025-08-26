# Redis Cache System Examples

This directory contains comprehensive examples demonstrating how to use the Redis cache system for PlantUML diagrams and state machines.

## Examples Overview

### Public Examples (examples/)

These examples demonstrate the public API and are intended for end users of the cache system.

#### 1. Diagram Cache Example (`diagram_cache_example/`)
- **Purpose**: Demonstrates basic diagram caching operations
- **Features**: Store, retrieve, and manage PlantUML diagrams
- **Key Concepts**: TTL management, error handling, cleanup operations

#### 2. State Machine Cache Example (`state_machine_cache_example/`)
- **Purpose**: Shows advanced state machine caching with entity management
- **Features**: Complex state machine storage, entity extraction, referential integrity
- **Key Concepts**: Entity relationships, cascade operations, type-safe retrieval

### Internal Examples (internal/examples/)

These examples demonstrate low-level functionality and are intended for developers working on the cache system itself.

#### 3. Redis Client Example (`internal/examples/redis_client_example/`)
- **Purpose**: Low-level Redis client operations
- **Features**: Connection management, batch operations, monitoring
- **Key Concepts**: Connection pooling, performance metrics, server information

#### 4. Retry Mechanism Example (`internal/examples/retry_example/`)
- **Purpose**: Demonstrates retry logic and error recovery
- **Features**: Multiple retry strategies, circuit breaker pattern, error classification
- **Key Concepts**: Exponential backoff, jitter, error recovery strategies

## Prerequisites

All examples require:
- **Redis Server**: Running on localhost:6379 (default configuration)
- **Go**: Version 1.24.4 or later
- **Network Access**: To connect to Redis server

### Installing Redis

#### Windows (using Chocolatey):
```cmd
choco install redis-64
redis-server
```

#### Windows (using WSL):
```bash
sudo apt update
sudo apt install redis-server
sudo service redis-server start
```

#### Verify Redis Installation:
```cmd
redis-cli ping
```
Expected response: `PONG`

## Running Examples

### Windows Commands

Navigate to any example directory and run:

```cmd
cd examples\diagram_cache_example
go run main.go
```

Or build and run:

```cmd
go build -o example.exe .
.\example.exe
```

### Example Execution Order

For learning purposes, run examples in this order:

1. **Diagram Cache Example**: Learn basic caching concepts
2. **State Machine Cache Example**: Understand complex data relationships
3. **Redis Client Example**: Explore low-level operations
4. **Retry Example**: Master error handling and resilience

## Common Configuration

All examples use similar Redis configuration:

```go
config := cache.DefaultRedisConfig()
config.RedisAddr = "localhost:6379"
config.DefaultTTL = 1 * time.Hour
```

### Customizing Configuration

You can modify examples to use different Redis settings:

```go
config.RedisAddr = "your-redis-host:6379"
config.RedisPassword = "your-password"
config.RedisDB = 1
config.PoolSize = 20
```

## Key Learning Concepts

### 1. Cache Hierarchy
- **Diagrams**: Base level storage for PlantUML content
- **State Machines**: Parsed representations with metadata
- **Entities**: Individual components (states, transitions, regions)

### 2. Error Handling
- **Connection Errors**: Network and Redis connectivity issues
- **Validation Errors**: Input validation and data integrity
- **Not Found Errors**: Cache misses and missing keys
- **Timeout Errors**: Operation timeouts and performance issues

### 3. Performance Optimization
- **TTL Management**: Automatic expiration of cached data
- **Batch Operations**: Efficient multi-key operations
- **Connection Pooling**: Resource management and scalability
- **Retry Logic**: Resilience against temporary failures

### 4. Monitoring and Diagnostics
- **Health Checks**: System status and connectivity
- **Performance Metrics**: Throughput and latency monitoring
- **Cache Statistics**: Memory usage and key counts
- **Error Tracking**: Failure patterns and recovery

## Troubleshooting

### Common Issues

#### Redis Connection Failed
```
Error: dial tcp 127.0.0.1:6379: connect: connection refused
```
**Solution**: Ensure Redis server is running
```cmd
redis-server
```

#### Permission Denied (Windows)
```
Error: Access is denied
```
**Solution**: Run command prompt as Administrator or use Windows Defender exclusions

#### Module Not Found
```
Error: cannot find module
```
**Solution**: Ensure you're in the correct directory and run:
```cmd
go mod tidy
```

#### Timeout Errors
```
Error: context deadline exceeded
```
**Solution**: Check Redis server performance and network connectivity

### Performance Tips

1. **Use appropriate TTL values** based on your data lifecycle
2. **Enable connection pooling** for high-concurrency scenarios
3. **Monitor memory usage** to prevent Redis OOM errors
4. **Use batch operations** when working with multiple keys
5. **Configure retry logic** based on your reliability requirements

## Advanced Usage

### Custom Error Handling
```go
if cache.IsConnectionError(err) {
    // Handle connection issues
} else if cache.IsNotFoundError(err) {
    // Handle cache misses
}
```

### Monitoring Integration
```go
healthStatus, err := cache.HealthDetailed(ctx)
if err == nil {
    log.Printf("Cache health: %s", healthStatus.Status)
}
```

### Cleanup Automation
```go
// Clean up old PUML entries
err = cache.CleanupWithOptions(ctx, "/diagrams/puml/*", &cache.CleanupOptions{
    MaxKeys: 1000,
    DryRun:  false,
})
```

## Contributing

When adding new examples:

1. **Follow naming conventions**: Use descriptive directory names
2. **Include README files**: Document purpose and usage
3. **Add error handling**: Demonstrate proper error management
4. **Use Windows-compatible commands**: Ensure examples work on Windows
5. **Test thoroughly**: Verify examples compile and run correctly

## Support

For issues with examples:

1. **Check Prerequisites**: Ensure Redis is running and accessible
2. **Review Error Messages**: Most errors include helpful context
3. **Check Configuration**: Verify Redis connection settings
4. **Monitor Resources**: Ensure sufficient memory and network capacity