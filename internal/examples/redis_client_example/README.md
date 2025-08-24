# Redis Client Low-Level Example

This example demonstrates low-level Redis client operations using the internal Redis client wrapper.

## Features Demonstrated

- Redis client configuration and initialization
- Retry logic with exponential backoff
- Basic Redis operations (SET, GET, EXISTS, DEL)
- Batch operations for efficiency
- SCAN operations for key discovery
- Database monitoring and statistics
- Connection pool information
- Server information retrieval
- Proper error handling and cleanup

## Prerequisites

- Redis server running on localhost:6379
- Go 1.24.4 or later

## Running the Example

### On Windows:

```cmd
cd internal\examples\redis_client_example
go run main.go
```

### Expected Output

The example will:
1. Configure and connect to Redis with retry settings
2. Perform basic SET/GET operations
3. Demonstrate batch operations
4. Show SCAN functionality for key discovery
5. Display database and server monitoring information
6. Show connection pool statistics
7. Clean up all test data

## Key Concepts

### Client Configuration
```go
config := internal.DefaultConfig()
config.RedisAddr = "localhost:6379"
config.RetryConfig.MaxAttempts = 5
config.RetryConfig.InitialDelay = 50 * time.Millisecond
```

### Retry Operations
```go
// All operations have retry variants
err = client.SetWithRetry(ctx, key, value, ttl)
value, err := client.GetWithRetry(ctx, key)
err = client.DelWithRetry(ctx, keys...)
```

### Batch Operations
```go
// Check multiple keys at once
existsCount, err := client.ExistsWithRetry(ctx, key1, key2, key3)

// Delete multiple keys and get count
deletedCount, err := client.DelBatchWithRetry(ctx, keys...)
```

### Monitoring Operations
```go
// Database size
dbSize, err := client.DBSizeWithRetry(ctx)

// Server information
info, err := client.InfoWithRetry(ctx, "server")

// Connection pool stats
connInfo, err := client.GetConnectionInfo(ctx)
```

## Retry Configuration

The example demonstrates configurable retry behavior:

- **MaxAttempts**: Maximum number of retry attempts
- **InitialDelay**: Starting delay before first retry
- **MaxDelay**: Maximum delay between retries
- **Multiplier**: Exponential backoff multiplier
- **Jitter**: Random variation to prevent thundering herd

## Error Handling

The client automatically handles:

- Connection failures with retry logic
- Network timeouts and interruptions
- Redis server temporary unavailability
- Pool exhaustion scenarios

## Performance Considerations

- Connection pooling for efficient resource usage
- Batch operations to reduce network round trips
- Configurable timeouts for different scenarios
- Monitoring capabilities for performance tuning

## Troubleshooting

Common issues and solutions:

1. **Connection Refused**: Ensure Redis server is running
2. **Timeout Errors**: Adjust timeout settings in configuration
3. **Pool Exhaustion**: Increase pool size or reduce connection hold time
4. **Memory Issues**: Monitor Redis memory usage and configure eviction policies