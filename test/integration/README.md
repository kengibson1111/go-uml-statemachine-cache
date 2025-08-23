# Integration Tests

This directory contains comprehensive integration tests that require a live Redis connection.

## Running Integration Tests

### Prerequisites

1. **Redis Server**: You need a running Redis server. You can:
   - Install Redis locally
   - Use Docker: `docker run -d -p 6379:6379 redis:latest`
   - Use Redis Cloud or other hosted Redis services

2. **Go Environment**: Ensure Go is installed and available in your PATH

### Running the Tests

#### Windows Users
Use the provided scripts for easy test execution:

```cmd
# Using Command Prompt
test\integration\run_integration_tests.bat

# Using PowerShell (recommended)
powershell -ExecutionPolicy Bypass -File test\integration\run_integration_tests.ps1
```

#### Manual Test Execution
```bash
# Run all integration tests
go test ./test/integration -v -timeout 300s

# Run specific test categories
go test ./test/integration -v -run "TestRedisCache_Concurrent.*" -timeout 60s
go test ./test/integration -v -run "TestRedisCache_Performance.*" -timeout 120s

# Run benchmarks
go test ./test/integration -bench=BenchmarkRedisCache_.* -benchtime=5s

# Run with custom Redis configuration
REDIS_ADDR=localhost:6379 REDIS_DB=15 go test ./test/integration -v
```

### Test Categories

#### Basic Integration Tests
- **Real Redis Operations**: Tests actual Redis SET, GET, DEL operations
- **TTL Functionality**: Tests time-to-live expiration with real Redis
- **Connection Handling**: Tests actual network connections and error scenarios
- **End-to-End Workflows**: Complete diagram storage, retrieval, and deletion cycles

#### Concurrent and Thread Safety Tests
- **Concurrent Diagram Operations**: Multiple goroutines performing diagram operations
- **Concurrent State Machine Operations**: Multiple goroutines with state machine data
- **Thread Safety**: Mixed operations testing thread safety
- **Race Condition Detection**: Ensures no data corruption under concurrent access

#### Performance and Stress Tests
- **High Throughput Operations**: Tests system behavior under high load
- **Large Payload Handling**: Tests with payloads from 1KB to 1MB
- **TTL Performance**: Performance testing with various TTL configurations
- **Memory Usage Patterns**: Monitors memory consumption and cleanup
- **Cleanup Performance**: Tests bulk deletion operations

#### Resilience Tests
- **Connection Resilience**: Tests behavior under connection issues
- **Timeout Handling**: Tests timeout scenarios and recovery
- **Error Recovery**: Tests error handling and retry mechanisms

### Test Configuration

Tests use the following configuration:
- **Redis Database**: 15 (to avoid conflicts with production data)
- **Default TTL**: 1 hour for test data
- **Automatic Cleanup**: All test data is cleaned up after completion
- **Timeout Handling**: Tests have appropriate timeouts to prevent hanging

### Test Behavior

- If Redis is not available, tests will be **skipped** (not failed)
- Tests use Redis database 15 to avoid conflicts with production data
- Tests automatically clean up test data after completion
- Performance tests include reasonable assertions for throughput and latency
- Stress tests can be skipped in short test mode (`go test -short`)

### Performance Expectations

The integration tests include performance assertions:
- **Basic Operations**: < 100ms for small payloads
- **Large Payloads (1MB)**: < 2s for store, < 1s for retrieval
- **Throughput**: > 100 operations/second under normal load
- **Error Rate**: < 5% under stress conditions
- **Cleanup**: < 5s for 1000 items

### Troubleshooting

#### Redis Connection Issues
1. Ensure Redis is running: `redis-cli ping`
2. Check Redis is accessible on localhost:6379
3. Verify no firewall blocking the connection
4. For Windows with Trend Micro: Ensure Redis executable is allowed

#### Test Failures
1. Check Redis logs for connection errors
2. Verify sufficient memory available for Redis
3. Ensure no other processes using Redis database 15
4. Run tests with `-v` flag for detailed output

#### Performance Issues
1. Ensure Redis is not under heavy load from other processes
2. Check system resources (CPU, memory, network)
3. Consider adjusting test timeouts for slower systems
4. Run performance tests individually to isolate issues

### Unit Tests vs Integration Tests

- **Unit Tests** (`cache/redis_cache_test.go`): Use mocks, no Redis required, fast
- **Integration Tests** (`test/integration/`): Use real Redis, slower, more comprehensive

Run unit tests for development and CI/CD pipelines. Run integration tests when you need to verify actual Redis behavior and performance characteristics.