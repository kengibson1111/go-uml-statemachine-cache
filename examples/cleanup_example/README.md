# Cache Cleanup Example

This example demonstrates the enhanced cache cleanup functionality in the Redis cache system.

## Features Demonstrated

1. **Basic Cleanup**: Simple pattern-based cache cleanup
2. **Advanced Cleanup**: Cleanup with configurable options
3. **Dry Run Mode**: Preview what would be deleted without actually deleting
4. **Cache Size Monitoring**: Get detailed information about cache usage
5. **Bulk Operations**: Efficient cleanup of large numbers of keys
6. **Metrics Collection**: Detailed statistics about cleanup operations

## Prerequisites

- Redis server running on localhost:6379
- Go 1.21 or later

## Running the Example

1. Start Redis server:
   ```bash
   redis-server
   ```

2. Run the example:
   ```bash
   go run main.go
   ```

## Cleanup Options

The example demonstrates various cleanup options:

- **BatchSize**: Number of keys to delete per batch (default: 100)
- **ScanCount**: Number of keys to scan per SCAN operation (default: 100)
- **MaxKeys**: Maximum number of keys to delete (0 = no limit)
- **DryRun**: Preview mode - scan but don't delete
- **Timeout**: Maximum time for cleanup operation
- **CollectMetrics**: Whether to collect detailed performance metrics

## Output

The example will show:

1. Initial cache population with test data
2. Cache size information before cleanup
3. Results of different cleanup operations
4. Performance metrics and statistics
5. Final cache size after cleanup

## Cleanup Patterns

The example uses various Redis key patterns:

- `/diagrams/puml/test-*` - Test PUML diagram keys
- `/machines/test-*` - Test state machine keys  
- `/test-*` - General test keys

## Error Handling

The example demonstrates proper error handling for:

- Connection failures
- Timeout errors
- Invalid cleanup options
- Redis operation failures

## Performance Considerations

The cleanup functionality is designed for efficiency:

- **Batched Operations**: Keys are deleted in configurable batches
- **Scan Optimization**: Uses Redis SCAN for memory-efficient key discovery
- **Timeout Protection**: Prevents long-running operations from blocking
- **Metrics Collection**: Optional detailed statistics with minimal overhead
- **Memory Monitoring**: Tracks memory usage and freed space

## Use Cases

This cleanup functionality is useful for:

- **Cache Maintenance**: Regular cleanup of expired or unused data
- **Testing**: Cleaning up test data after test runs
- **Memory Management**: Freeing up Redis memory when approaching limits
- **Pattern-based Cleanup**: Removing specific types of cached data
- **Monitoring**: Understanding cache usage patterns and growth