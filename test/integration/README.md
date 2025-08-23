# Integration Tests

This directory contains integration tests that require a live Redis connection.

## Running Integration Tests

### Prerequisites

1. **Redis Server**: You need a running Redis server. You can:
   - Install Redis locally
   - Use Docker: `docker run -d -p 6379:6379 redis:latest`
   - Use Redis Cloud or other hosted Redis services

### Running the Tests

```bash
# Run integration tests (will skip if Redis is not available)
go test ./test/integration -v

# Run integration tests with Redis connection details
REDIS_ADDR=localhost:6379 go test ./test/integration -v
```

### Test Behavior

- If Redis is not available, tests will be **skipped** (not failed)
- Tests use Redis database 15 to avoid conflicts with production data
- Tests automatically clean up test data after completion

### What These Tests Cover

- **Real Redis Operations**: Tests actual Redis SET, GET, DEL operations
- **TTL Functionality**: Tests time-to-live expiration with real Redis
- **Connection Handling**: Tests actual network connections and error scenarios
- **End-to-End Workflows**: Complete diagram storage, retrieval, and deletion cycles

### Unit Tests vs Integration Tests

- **Unit Tests** (`cache/redis_cache_test.go`): Use mocks, no Redis required, fast
- **Integration Tests** (`test/integration/`): Use real Redis, slower, more comprehensive

Run unit tests for development and CI/CD pipelines. Run integration tests when you need to verify actual Redis behavior.