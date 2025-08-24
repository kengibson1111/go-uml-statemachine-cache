# Retry Mechanism Example

This example demonstrates the retry logic and error recovery mechanisms built into the Redis cache system.

## Features Demonstrated

- Multiple retry configuration strategies
- Exponential backoff with configurable parameters
- Jitter to prevent thundering herd problems
- Error classification and severity levels
- Recovery strategies for different error types
- Circuit breaker pattern implementation
- Error context and detailed error information
- Timing analysis of retry operations

## Prerequisites

- Redis server running on localhost:6379
- Go 1.24.4 or later

## Running the Example

### On Windows:

```cmd
cd internal\examples\retry_example
go run main.go
```

### Expected Output

The example will:
1. Test three different retry configurations
2. Demonstrate successful operations with timing
3. Show retry delay calculations
4. Demonstrate error recovery manager
5. Show error classification and type checking
6. Display error context information

## Retry Configurations

### 1. Conservative Retry
- **Max Attempts**: 3
- **Initial Delay**: 100ms
- **Max Delay**: 1s
- **Multiplier**: 2.0
- **Jitter**: Disabled

Best for: Production environments where stability is prioritized over speed.

### 2. Aggressive Retry
- **Max Attempts**: 7
- **Initial Delay**: 50ms
- **Max Delay**: 5s
- **Multiplier**: 1.5
- **Jitter**: Enabled

Best for: High-availability scenarios where temporary failures are common.

### 3. Fast Retry with Jitter
- **Max Attempts**: 5
- **Initial Delay**: 25ms
- **Max Delay**: 500ms
- **Multiplier**: 2.5
- **Jitter**: Enabled

Best for: Low-latency applications with good network conditions.

## Error Classification

The system classifies errors by:

### Severity Levels
- **Low**: Minor issues that don't affect functionality
- **Medium**: Issues that may impact performance
- **High**: Significant problems requiring attention
- **Critical**: System-threatening errors

### Recovery Strategies
- **Fail**: Immediately fail without retry
- **Ignore**: Log but continue operation
- **RetryWithBackoff**: Use exponential backoff retry
- **RetryWithDelay**: Use fixed delay retry
- **CircuitBreaker**: Use circuit breaker pattern
- **WaitAndRetry**: Wait for condition then retry

## Error Types

The system handles various error types:

1. **Connection Errors**: Network connectivity issues
2. **Timeout Errors**: Operation timeouts
3. **Validation Errors**: Input validation failures
4. **Not Found Errors**: Cache misses
5. **Serialization Errors**: JSON marshaling/unmarshaling issues
6. **Retry Exhausted Errors**: All retry attempts failed
7. **Circuit Open Errors**: Circuit breaker is open

## Circuit Breaker Pattern

The circuit breaker prevents cascading failures:

- **Closed**: Normal operation, requests pass through
- **Open**: Failures detected, requests fail fast
- **Half-Open**: Testing if service has recovered

Configuration parameters:
- **Failure Threshold**: Number of failures before opening
- **Recovery Timeout**: Time before attempting recovery
- **Half-Open Max Calls**: Maximum calls in half-open state

## Retry Delay Calculation

Delays are calculated using exponential backoff:

```
delay = initial_delay * (multiplier ^ attempt)
delay = min(delay, max_delay)
if jitter_enabled:
    delay += random(0, delay * 0.1)
```

## Error Context

Detailed error context includes:
- Operation being performed
- Cache key involved
- Current attempt number
- Maximum attempts configured
- Operation duration
- Timestamp of occurrence

## Best Practices

1. **Choose appropriate retry configuration** based on your use case
2. **Enable jitter** in high-concurrency scenarios
3. **Monitor error patterns** to tune retry parameters
4. **Use circuit breakers** to prevent cascading failures
5. **Log retry attempts** for debugging and monitoring
6. **Set reasonable timeouts** to avoid resource exhaustion

## Troubleshooting

Common retry scenarios:

1. **Network Instability**: Use aggressive retry with jitter
2. **Redis Overload**: Use conservative retry with circuit breaker
3. **Temporary Outages**: Use longer max delays
4. **High Concurrency**: Enable jitter to spread load
5. **Critical Operations**: Use higher max attempts