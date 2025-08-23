# Examples

This directory contains examples demonstrating how to use the public APIs of the go-uml-statemachine-cache library.

## Public API Examples

These examples show how to use the library in your applications:

### diagram_cache_example
Demonstrates basic PlantUML diagram caching operations:
- Storing diagrams with TTL
- Retrieving cached diagrams
- Handling special characters in diagram names
- Error handling and cleanup

**Run:** `go run examples/diagram_cache_example/main.go`

### state_machine_cache_example
Shows state machine caching with versioning:
- Storing complex state machine objects
- Version-based retrieval
- JSON serialization
- Multi-version management

**Run:** `go run examples/state_machine_cache_example/main.go`

### health_monitoring_example
Demonstrates comprehensive health monitoring:
- System health checks and status monitoring
- Performance metrics collection
- Network connectivity validation
- Data integrity checks
- Automated troubleshooting recommendations

**Run:** `go run examples/health_monitoring_example/main.go`

### health_monitoring_diagnostics
Command-line diagnostic tool for troubleshooting cache issues:
- Verbose diagnostic output
- System health analysis
- Performance troubleshooting

**Run:** `go run examples/health_monitoring_diagnostics/main.go -verbose`

## Prerequisites

Both examples require a running Redis server on `localhost:6379`. You can start Redis using:

```bash
# Using Docker
docker run -d -p 6379:6379 redis:latest

# Or install Redis locally and run
redis-server
```

## Internal Examples

For examples using internal APIs (for library development), see `internal/examples/`:
- `redis_client_example/` - Low-level Redis client configuration
- `retry_example/` - Retry mechanism and error handling

These internal examples are useful for understanding the library's internals but should not be used as reference for application development.