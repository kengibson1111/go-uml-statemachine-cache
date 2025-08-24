# API Reference

This document provides a comprehensive reference for the go-uml-statemachine-cache library's public API.

## Table of Contents

- [Core Interfaces](#core-interfaces)
- [Configuration Types](#configuration-types)
- [Error Handling](#error-handling)
- [Health Monitoring](#health-monitoring)
- [Cache Management](#cache-management)
- [Type Definitions](#type-definitions)
- [Constants](#constants)
- [Functions](#functions)

## Core Interfaces

### Cache Interface

The main interface for all caching operations.

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
    UpdateStateMachineEntityMapping(ctx context.Context, umlVersion, name string, entityID, entityKey string, operation string) error

    // Type-safe entity retrieval methods
    GetEntityAsState(ctx context.Context, umlVersion, diagramName, entityID string) (*models.State, error)
    GetEntityAsTransition(ctx context.Context, umlVersion, diagramName, entityID string) (*models.Transition, error)
    GetEntityAsRegion(ctx context.Context, umlVersion, diagramName, entityID string) (*models.Region, error)
    GetEntityAsVertex(ctx context.Context, umlVersion, diagramName, entityID string) (*models.Vertex, error)

    // Management operations
    Cleanup(ctx context.Context, pattern string) error
    Health(ctx context.Context) error
    Close() error

    // Enhanced cleanup and monitoring operations
    CleanupWithOptions(ctx context.Context, pattern string, options *CleanupOptions) (*CleanupResult, error)
    GetCacheSize(ctx context.Context) (*CacheSizeInfo, error)

    // Enhanced health monitoring operations
    HealthDetailed(ctx context.Context) (*HealthStatus, error)
    GetConnectionHealth(ctx context.Context) (*ConnectionHealth, error)
    GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error)
    RunDiagnostics(ctx context.Context) (*DiagnosticInfo, error)
}
```

### RedisCache Implementation

The Redis-based implementation of the Cache interface.

```go
type RedisCache struct {
    // Private fields - use NewRedisCache() to create instances
}

// NewRedisCache creates a new Redis-backed cache implementation
func NewRedisCache(config *RedisConfig) (*RedisCache, error)
```

## Configuration Types

### RedisConfig

Main configuration for Redis connection and cache behavior.

```go
type RedisConfig struct {
    // Redis connection settings
    RedisAddr     string        // Redis server address (host:port)
    RedisPassword string        // Redis password (optional)
    RedisDB       int           // Redis database number (0-15)

    // Connection pool settings
    MaxRetries   int            // Maximum number of retries
    DialTimeout  time.Duration  // Timeout for establishing connection
    ReadTimeout  time.Duration  // Timeout for socket reads
    WriteTimeout time.Duration  // Timeout for socket writes
    PoolSize     int            // Maximum number of socket connections

    // Cache settings
    DefaultTTL time.Duration    // Default time-to-live for cache entries

    // Resilience settings
    RetryConfig *RedisRetryConfig // Retry configuration for operations
}

// DefaultRedisConfig returns a RedisConfig with sensible default values
func DefaultRedisConfig() *RedisConfig
```

**Default Values:**
- `RedisAddr`: "localhost:6379"
- `RedisPassword`: ""
- `RedisDB`: 0
- `MaxRetries`: 3
- `DialTimeout`: 5 seconds
- `ReadTimeout`: 3 seconds
- `WriteTimeout`: 3 seconds
- `PoolSize`: 10
- `DefaultTTL`: 24 hours

### RedisRetryConfig

Configuration for retry logic with exponential backoff.

```go
type RedisRetryConfig struct {
    MaxAttempts  int           // Maximum number of retry attempts
    InitialDelay time.Duration // Initial delay before first retry
    MaxDelay     time.Duration // Maximum delay between retries
    Multiplier   float64       // Backoff multiplier
    Jitter       bool          // Whether to add random jitter
    RetryableOps []string      // Operations that should be retried
}

// DefaultRedisRetryConfig returns a RedisRetryConfig with sensible default values
func DefaultRedisRetryConfig() *RedisRetryConfig
```

**Default Values:**
- `MaxAttempts`: 3
- `InitialDelay`: 100 milliseconds
- `MaxDelay`: 5 seconds
- `Multiplier`: 2.0
- `Jitter`: true
- `RetryableOps`: ["ping", "get", "set", "del", "exists", "expire"]

## Error Handling

### CacheError

Comprehensive error type with detailed context information.

```go
type CacheError struct {
    Type     CacheErrorType // Error type classification
    Key      string         // Cache key involved in the error
    Message  string         // Human-readable error message
    Cause    error          // Underlying error (not serialized)
    Context  *ErrorContext  // Additional error context
    Severity ErrorSeverity  // Error severity level
}

// Error implements the error interface
func (e *CacheError) Error() string

// Unwrap returns the underlying cause error
func (e *CacheError) Unwrap() error

// Is checks if the error matches the target error type
func (e *CacheError) Is(target error) bool

// IsRetryable returns true if this error should trigger retry logic
func (e *CacheError) IsRetryable() bool

// GetRecoveryStrategy returns the recommended recovery strategy for this error
func (e *CacheError) GetRecoveryStrategy() RecoveryStrategy

// WithContext adds context information to an existing CacheError
func (e *CacheError) WithContext(operation string, attemptNumber int, duration time.Duration) *CacheError

// WithMetadata adds metadata to an existing CacheError
func (e *CacheError) WithMetadata(key string, value any) *CacheError
```

### ErrorContext

Additional context information for errors.

```go
type ErrorContext struct {
    Operation     string         // The operation that failed
    Timestamp     time.Time      // When the error occurred
    AttemptNumber int            // Retry attempt number
    Duration      time.Duration  // How long the operation took
    Metadata      map[string]any // Additional context data
}
```

### Error Types

```go
type CacheErrorType int

const (
    CacheErrorTypeConnection     CacheErrorType // Redis connection error
    CacheErrorTypeKeyInvalid     CacheErrorType // Invalid cache key
    CacheErrorTypeNotFound       CacheErrorType // Cache miss or key not found
    CacheErrorTypeSerialization  CacheErrorType // JSON marshaling/unmarshaling error
    CacheErrorTypeTimeout        CacheErrorType // Timeout during cache operation
    CacheErrorTypeCapacity       CacheErrorType // Cache capacity or memory issues
    CacheErrorTypeValidation     CacheErrorType // Input validation failure
    CacheErrorTypeRetryExhausted CacheErrorType // All retry attempts exhausted
    CacheErrorTypeCircuitOpen    CacheErrorType // Circuit breaker is open
)
```

### Error Severity

```go
type ErrorSeverity int

const (
    SeverityLow      ErrorSeverity // Low-severity error (validation, not found)
    SeverityMedium   ErrorSeverity // Medium-severity error (serialization)
    SeverityHigh     ErrorSeverity // High-severity error (retry exhausted)
    SeverityCritical ErrorSeverity // Critical error (connection failure)
)
```

### Recovery Strategy

```go
type RecoveryStrategy int

const (
    RecoveryStrategyFail             RecoveryStrategy // Error should not be recovered from
    RecoveryStrategyIgnore           RecoveryStrategy // Error can be safely ignored
    RecoveryStrategyRetryWithBackoff RecoveryStrategy // Retry with exponential backoff
    RecoveryStrategyRetryWithDelay   RecoveryStrategy // Retry with fixed delay
    RecoveryStrategyCircuitBreaker   RecoveryStrategy // Circuit breaker should be activated
    RecoveryStrategyWaitAndRetry     RecoveryStrategy // Wait for circuit breaker to close
)
```

## Health Monitoring

### HealthStatus

Comprehensive health status information.

```go
type HealthStatus struct {
    Status       string             // "healthy", "degraded", "unhealthy"
    Timestamp    time.Time          // When the health check was performed
    ResponseTime time.Duration      // Time taken for health check
    Connection   ConnectionHealth   // Connection-specific health info
    Performance  PerformanceMetrics // Performance metrics
    Diagnostics  DiagnosticInfo     // Diagnostic information
    Errors       []string           // Any errors encountered
    Warnings     []string           // Any warnings
}
```

### ConnectionHealth

Connection-specific health information.

```go
type ConnectionHealth struct {
    Connected   bool                // Whether Redis is connected
    Address     string              // Redis server address
    Database    int                 // Redis database number
    PingLatency time.Duration       // Latency of ping command
    PoolStats   ConnectionPoolStats // Connection pool statistics
    LastError   string              // Last connection error if any
}
```

### ConnectionPoolStats

Redis connection pool statistics.

```go
type ConnectionPoolStats struct {
    TotalConnections  int // Total connections in pool
    IdleConnections   int // Idle connections
    ActiveConnections int // Active connections
    Hits              int // Pool hits
    Misses            int // Pool misses
    Timeouts          int // Pool timeouts
    StaleConnections  int // Stale connections
}
```

### PerformanceMetrics

Performance-related metrics.

```go
type PerformanceMetrics struct {
    MemoryUsage    MemoryMetrics    // Memory usage information
    KeyspaceInfo   KeyspaceMetrics  // Keyspace statistics
    OperationStats OperationMetrics // Operation statistics
    ServerInfo     ServerMetrics    // Server information
}
```

### MemoryMetrics

Memory usage information.

```go
type MemoryMetrics struct {
    UsedMemory          int64   // Used memory in bytes
    UsedMemoryHuman     string  // Human readable used memory
    UsedMemoryPeak      int64   // Peak memory usage
    MemoryFragmentation float64 // Memory fragmentation ratio
    MaxMemory           int64   // Maximum memory limit
    MaxMemoryPolicy     string  // Memory eviction policy
}
```

### KeyspaceMetrics

Keyspace statistics.

```go
type KeyspaceMetrics struct {
    TotalKeys      int64   // Total number of keys
    ExpiringKeys   int64   // Keys with expiration
    AverageKeySize float64 // Average key size in bytes
    KeyspaceHits   int64   // Keyspace hits
    KeyspaceMisses int64   // Keyspace misses
    HitRate        float64 // Cache hit rate percentage
}
```

### OperationMetrics

Operation statistics.

```go
type OperationMetrics struct {
    TotalCommands       int64   // Total commands processed
    CommandsPerSecond   float64 // Commands per second
    ConnectedClients    int64   // Number of connected clients
    BlockedClients      int64   // Number of blocked clients
    RejectedConnections int64   // Rejected connections
}
```

### ServerMetrics

Server information.

```go
type ServerMetrics struct {
    RedisVersion    string // Redis server version
    UptimeSeconds   int64  // Server uptime in seconds
    UptimeDays      int64  // Server uptime in days
    ServerMode      string // Server mode (standalone, sentinel, cluster)
    Role            string // Server role (master, slave)
    ConnectedSlaves int64  // Number of connected slaves
}
```

### DiagnosticInfo

Diagnostic information for troubleshooting.

```go
type DiagnosticInfo struct {
    ConfigurationCheck ConfigCheck        // Configuration validation
    NetworkCheck       NetworkCheck       // Network connectivity check
    PerformanceCheck   PerformanceCheck   // Performance validation
    DataIntegrityCheck DataIntegrityCheck // Data integrity validation
    Recommendations    []string           // Performance recommendations
}
```

## Cache Management

### CleanupOptions

Configuration for advanced cleanup operations.

```go
type CleanupOptions struct {
    BatchSize      int           // Number of keys to delete per batch
    ScanCount      int64         // Number of keys to scan per SCAN operation
    MaxKeys        int64         // Maximum number of keys to delete (0 = no limit)
    DryRun         bool          // If true, only scan but don't delete
    Timeout        time.Duration // Timeout for the entire cleanup operation
    CollectMetrics bool          // Whether to collect detailed metrics
}
```

### CleanupResult

Results from a cleanup operation.

```go
type CleanupResult struct {
    KeysScanned    int64         // Number of keys scanned
    KeysDeleted    int64         // Number of keys deleted
    BytesFreed     int64         // Estimated bytes freed
    Duration       time.Duration // Time taken for cleanup
    BatchesUsed    int           // Number of batches used
    ErrorsOccurred int           // Number of errors encountered
}
```

### CacheSizeInfo

Information about cache size and usage.

```go
type CacheSizeInfo struct {
    TotalKeys         int64 // Total number of keys
    DiagramCount      int64 // Number of diagram keys
    StateMachineCount int64 // Number of state machine keys
    EntityCount       int64 // Number of entity keys
    MemoryUsed        int64 // Memory used in bytes
    MemoryPeak        int64 // Peak memory usage in bytes
    MemoryOverhead    int64 // Memory overhead in bytes
}
```

## Functions

### Constructor Functions

```go
// NewRedisCache creates a new Redis-backed cache implementation
func NewRedisCache(config *RedisConfig) (*RedisCache, error)

// DefaultRedisConfig returns a RedisConfig with sensible default values
func DefaultRedisConfig() *RedisConfig

// DefaultRedisRetryConfig returns a RedisRetryConfig with sensible default values
func DefaultRedisRetryConfig() *RedisRetryConfig
```

### Error Creation Functions

```go
// NewCacheError creates a new CacheError with basic information
func NewCacheError(errType CacheErrorType, key, message string, cause error) *CacheError

// NewCacheErrorWithContext creates a new CacheError with detailed context
func NewCacheErrorWithContext(errType CacheErrorType, key, message string, cause error, context *ErrorContext) *CacheError

// NewConnectionError creates a connection-specific cache error
func NewConnectionError(message string, cause error) *CacheError

// NewKeyInvalidError creates a key validation error
func NewKeyInvalidError(key, message string) *CacheError

// NewNotFoundError creates a not found error
func NewNotFoundError(key string) *CacheError

// NewSerializationError creates a serialization error
func NewSerializationError(key, message string, cause error) *CacheError

// NewTimeoutError creates a timeout error
func NewTimeoutError(key, message string, cause error) *CacheError

// NewValidationError creates a validation error
func NewValidationError(message string, cause error) *CacheError

// NewRetryExhaustedError creates a retry exhausted error
func NewRetryExhaustedError(operation string, attempts int, lastError error) *CacheError

// NewCircuitOpenError creates a circuit breaker open error
func NewCircuitOpenError(operation string) *CacheError
```

### Error Checking Functions

```go
// IsConnectionError checks if the error is a connection error
func IsConnectionError(err error) bool

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool

// IsRetryExhaustedError checks if the error is a retry exhausted error
func IsRetryExhaustedError(err error) bool

// IsCircuitOpenError checks if the error is a circuit open error
func IsCircuitOpenError(err error) bool

// IsRetryableError checks if the error should trigger retry logic
func IsRetryableError(err error) bool
```

### Error Analysis Functions

```go
// GetErrorSeverity returns the severity level of an error
func GetErrorSeverity(err error) ErrorSeverity

// GetRecoveryStrategy returns the recommended recovery strategy for an error
func GetRecoveryStrategy(err error) RecoveryStrategy
```

### Circuit Breaker Functions

```go
// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig

// NewErrorRecoveryManager creates a new error recovery manager
func NewErrorRecoveryManager(config *CircuitBreakerConfig) *ErrorRecoveryManager
```

## Usage Examples

### Basic Cache Operations

```go
// Create cache
config := cache.DefaultRedisConfig()
config.RedisAddr = "localhost:6379"
redisCache, err := cache.NewRedisCache(config)
if err != nil {
    log.Fatal(err)
}
defer redisCache.Close()

ctx := context.Background()

// Store diagram
err = redisCache.StoreDiagram(ctx, "my-diagram", pumlContent, time.Hour)
if err != nil {
    log.Fatal(err)
}

// Get diagram
content, err := redisCache.GetDiagram(ctx, "my-diagram")
if err != nil {
    if cache.IsNotFoundError(err) {
        log.Println("Diagram not found")
    } else {
        log.Fatal(err)
    }
}
```

### Error Handling

```go
_, err := redisCache.GetDiagram(ctx, "nonexistent")
if err != nil {
    // Check error type
    if cache.IsNotFoundError(err) {
        // Handle cache miss
    } else if cache.IsConnectionError(err) {
        // Handle connection issues
    }
    
    // Get error details
    if cacheErr, ok := err.(*cache.CacheError); ok {
        fmt.Printf("Error Type: %v\n", cacheErr.Type)
        fmt.Printf("Severity: %v\n", cacheErr.Severity)
        fmt.Printf("Recovery Strategy: %v\n", cacheErr.GetRecoveryStrategy())
    }
}
```

### Health Monitoring

```go
// Basic health check
err = redisCache.Health(ctx)
if err != nil {
    log.Println("Cache unhealthy:", err)
}

// Detailed health status
health, err := redisCache.HealthDetailed(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Status: %s\n", health.Status)
fmt.Printf("Response Time: %v\n", health.ResponseTime)
```

### Cache Cleanup

```go
// Basic cleanup
err = redisCache.Cleanup(ctx, "/diagrams/puml/test-*")
if err != nil {
    log.Fatal(err)
}

// Advanced cleanup
options := &cache.CleanupOptions{
    BatchSize:      100,
    MaxKeys:        1000,
    CollectMetrics: true,
}

result, err := redisCache.CleanupWithOptions(ctx, "/machines/*", options)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Deleted %d keys\n", result.KeysDeleted)
```

## Best Practices

### Configuration

1. **Use appropriate timeouts** based on your network conditions
2. **Configure retry logic** for your specific use case
3. **Set reasonable TTL values** to balance performance and freshness
4. **Size connection pools** based on expected concurrent usage

### Error Handling

1. **Always check for specific error types** using the provided functions
2. **Implement appropriate recovery strategies** based on error severity
3. **Log errors with context** for better debugging
4. **Use circuit breakers** for critical operations

### Performance

1. **Use batch operations** for bulk deletions
2. **Monitor cache size** and implement cleanup strategies
3. **Configure appropriate pool sizes** for your workload
4. **Use health monitoring** to detect performance issues

### Security

1. **Validate all inputs** before caching
2. **Use secure Redis configurations** with authentication
3. **Implement proper key naming** to avoid conflicts
4. **Monitor for suspicious patterns** in error logs

## Thread Safety

All public API methods are thread-safe and can be called concurrently from multiple goroutines. The library handles internal synchronization and connection pooling automatically.

## Compatibility

- **Go Version**: 1.24.4 or later
- **Redis Version**: 6.0 or later
- **Operating Systems**: Windows, Linux, macOS
- **Architectures**: amd64, arm64

## Migration Guide

When upgrading between major versions, refer to the changelog and migration guide in the repository for breaking changes and upgrade instructions.