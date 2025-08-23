# Public API Reference

This document describes the public API surface of the go-uml-statemachine-cache library.

## Core Interface

### Cache Interface

The main interface for all caching operations:

```go
type Cache interface {
    // Diagram operations
    StoreDiagram(ctx context.Context, name string, pumlContent string, ttl time.Duration) error
    GetDiagram(ctx context.Context, name string) (string, error)
    DeleteDiagram(ctx context.Context, name string) error

    // State machine operations
    StoreStateMachine(ctx context.Context, umlVersion, name string, machine *models.StateMachine, ttl time.Duration) error
    GetStateMachine(ctx context.Context, umlVersion, name string) (*models.StateMachine, error)
    DeleteStateMachine(ctx context.Context, umlVersion, name string) error

    // Entity operations
    StoreEntity(ctx context.Context, umlVersion, diagramName, entityID string, entity interface{}, ttl time.Duration) error
    GetEntity(ctx context.Context, umlVersion, diagramName, entityID string) (interface{}, error)

    // Management operations
    Cleanup(ctx context.Context, pattern string) error
    Health(ctx context.Context) error
    Close() error
}
```

## Configuration Types

### RedisConfig

Configuration for Redis cache instances:

```go
type RedisConfig = internal.Config
```

### RedisRetryConfig

Configuration for retry behavior:

```go
type RedisRetryConfig = internal.RetryConfig
```

## Error Types

### CacheError

Main error type with context:

```go
type CacheError = internal.CacheError
type CacheErrorType = internal.ErrorType
```

### Error Constants

```go
const (
    ErrorTypeConnection    CacheErrorType = "connection"
    ErrorTypeKeyInvalid    CacheErrorType = "key_invalid"
    ErrorTypeNotFound      CacheErrorType = "not_found"
    ErrorTypeSerialization CacheErrorType = "serialization"
    ErrorTypeTimeout       CacheErrorType = "timeout"
    ErrorTypeValidation    CacheErrorType = "validation"
)
```

## Constructor Functions

### NewRedisCache

Create a new Redis cache instance:

```go
func NewRedisCache(config *RedisConfig) (*RedisCache, error)
```

### Configuration Helpers

```go
func DefaultRedisConfig() *RedisConfig
func DefaultRedisRetryConfig() *RedisRetryConfig
```

## Error Creation Functions

```go
func NewCacheError(errType CacheErrorType, key, message string, cause error) *CacheError
func NewConnectionError(message string, cause error) *CacheError
func NewKeyInvalidError(key, message string) *CacheError
func NewNotFoundError(key string) *CacheError
func NewSerializationError(key, message string, cause error) *CacheError
func NewTimeoutError(key, message string, cause error) *CacheError
func NewValidationError(message string, cause error) *CacheError
```

## Error Checking Functions

```go
func IsConnectionError(err error) bool
func IsNotFoundError(err error) bool
func IsValidationError(err error) bool
```

## Implementation

### RedisCache

The Redis implementation of the Cache interface:

```go
type RedisCache struct {
    // Internal fields (not exported)
}
```

## Usage Patterns

### Basic Usage

```go
// Create cache
config := cache.DefaultRedisConfig()
config.RedisAddr = "localhost:6379"
redisCache, err := cache.NewRedisCache(config)
if err != nil {
    return err
}
defer redisCache.Close()

// Use cache
ctx := context.Background()
err = redisCache.StoreDiagram(ctx, "my-diagram", content, time.Hour)
```

### Error Handling

```go
diagram, err := redisCache.GetDiagram(ctx, "my-diagram")
if err != nil {
    if cache.IsNotFoundError(err) {
        // Handle not found
    } else if cache.IsConnectionError(err) {
        // Handle connection issues
    } else {
        // Handle other errors
    }
}
```

### Configuration

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

## Thread Safety

All public APIs are thread-safe and can be used concurrently from multiple goroutines.

## Dependencies

The library depends on:
- `github.com/kengibson1111/go-uml-statemachine-models` for state machine models
- Redis server for storage backend