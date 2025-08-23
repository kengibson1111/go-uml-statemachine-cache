# Internal Examples

This directory contains examples that demonstrate the internal APIs and components of the go-uml-statemachine-cache library. These examples are primarily for library developers and contributors.

## Internal API Examples

### redis_client_example
Demonstrates low-level Redis client configuration and usage:
- Default vs custom configuration
- Configuration validation
- Connection management
- Health checks and connection info

**Run:** `go run internal/examples/redis_client_example/main.go`

### retry_example
Shows the retry mechanism and error handling:
- Configurable retry policies
- Exponential backoff with jitter
- Operation-specific retry behavior
- Performance timing

**Run:** `go run internal/examples/retry_example/main.go`

## Usage Note

These examples use internal APIs that are not part of the public interface and may change without notice. For application development, use the examples in the `examples/` directory instead, which demonstrate the stable public APIs.

## Prerequisites

Both examples require a running Redis server on `localhost:6379` for full functionality, though they will demonstrate error handling when Redis is not available.

```bash
# Using Docker
docker run -d -p 6379:6379 redis:latest

# Or install Redis locally and run
redis-server
```