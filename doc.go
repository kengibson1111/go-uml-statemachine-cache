// Package cache provides a Redis-based caching system for PlantUML diagrams and UML state machine definitions.
//
// This package implements a comprehensive caching layer that supports:
//   - PlantUML diagram storage and retrieval with TTL support
//   - Parsed UML state machine caching with version management
//   - Individual state machine entity caching with hierarchical keys
//   - Comprehensive error handling with typed errors and recovery strategies
//   - Configurable retry logic with exponential backoff
//   - Health monitoring and diagnostics
//   - Thread-safe concurrent access
//
// # Architecture
//
// The cache system uses a hierarchical key structure in Redis:
//   - PlantUML Diagrams: /diagrams/puml/<diagram_name>
//   - Parsed State Machines: /machines/<uml_version>/<diagram_name>
//   - State Machine Entities: /machines/<uml_version>/<diagram_name>/entities/<entity_id>
//
// # Basic Usage
//
// Create a cache instance with default configuration:
//
//	config := cache.DefaultRedisConfig()
//	config.RedisAddr = "localhost:6379"
//
//	redisCache, err := cache.NewRedisCache(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer redisCache.Close()
//
// Store and retrieve a PlantUML diagram:
//
//	ctx := context.Background()
//	pumlContent := `@startuml
//	[*] --> State1
//	State1 --> [*]
//	@enduml`
//
//	// Store diagram with 1 hour TTL
//	err = redisCache.StoreDiagram(ctx, "my-diagram", pumlContent, time.Hour)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Retrieve diagram
//	retrieved, err := redisCache.GetDiagram(ctx, "my-diagram")
//	if err != nil {
//	    if cache.IsNotFoundError(err) {
//	        log.Println("Diagram not found")
//	    } else {
//	        log.Fatal(err)
//	    }
//	}
//
// # Configuration
//
// The package provides flexible configuration through RedisConfig:
//
//	config := &cache.RedisConfig{
//	    RedisAddr:    "localhost:6379",  // Redis server address
//	    RedisPassword: "",               // Redis password (optional)
//	    RedisDB:      0,                 // Redis database number
//	    MaxRetries:   3,                 // Maximum retry attempts
//	    DialTimeout:  5 * time.Second,   // Connection timeout
//	    ReadTimeout:  3 * time.Second,   // Read operation timeout
//	    WriteTimeout: 3 * time.Second,   // Write operation timeout
//	    PoolSize:     10,                // Connection pool size
//	    DefaultTTL:   24 * time.Hour,    // Default TTL for cache entries
//	    RetryConfig:  cache.DefaultRedisRetryConfig(), // Retry configuration
//	}
//
// # Error Handling
//
// The package provides comprehensive error handling with typed errors:
//
//	_, err := redisCache.GetDiagram(ctx, "nonexistent")
//	if err != nil {
//	    switch {
//	    case cache.IsNotFoundError(err):
//	        // Handle cache miss
//	        log.Println("Diagram not found in cache")
//	    case cache.IsConnectionError(err):
//	        // Handle Redis connection issues
//	        log.Println("Redis connection error:", err)
//	    case cache.IsValidationError(err):
//	        // Handle input validation errors
//	        log.Println("Invalid input:", err)
//	    case cache.IsRetryExhaustedError(err):
//	        // Handle retry exhaustion
//	        log.Println("Operation failed after retries:", err)
//	    default:
//	        // Handle other errors
//	        log.Println("Unexpected error:", err)
//	    }
//	}
//
// # State Machine Caching
//
// Store and retrieve parsed state machines with entity relationships:
//
//	// Assuming you have a parsed state machine
//	var stateMachine *models.StateMachine
//
//	// Store state machine (automatically creates entity cache entries)
//	err = redisCache.StoreStateMachine(ctx, "v1.0", "my-diagram", stateMachine, time.Hour)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Retrieve state machine
//	retrieved, err := redisCache.GetStateMachine(ctx, "v1.0", "my-diagram")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Access individual entities
//	entity, err := redisCache.GetEntity(ctx, "v1.0", "my-diagram", "state-id")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Type-safe entity retrieval
//	state, err := redisCache.GetEntityAsState(ctx, "v1.0", "my-diagram", "state-id")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Health Monitoring
//
// Monitor cache health and performance:
//
//	// Basic health check
//	err = redisCache.Health(ctx)
//	if err != nil {
//	    log.Println("Cache is unhealthy:", err)
//	}
//
//	// Detailed health status
//	health, err := redisCache.HealthDetailed(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Cache Status: %s\n", health.Status)
//	fmt.Printf("Response Time: %v\n", health.ResponseTime)
//	fmt.Printf("Memory Usage: %d bytes\n", health.Performance.MemoryUsage.UsedMemory)
//
// # Cache Management
//
// Clean up cache entries and monitor usage:
//
//	// Basic cleanup with pattern
//	err = redisCache.Cleanup(ctx, "/diagrams/puml/test-*")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Advanced cleanup with options
//	options := &cache.CleanupOptions{
//	    BatchSize:      100,
//	    ScanCount:      1000,
//	    MaxKeys:        10000,
//	    DryRun:         false,
//	    Timeout:        30 * time.Second,
//	    CollectMetrics: true,
//	}
//
//	result, err := redisCache.CleanupWithOptions(ctx, "/machines/v1.0/*", options)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Deleted %d keys in %v\n", result.KeysDeleted, result.Duration)
//
//	// Get cache size information
//	sizeInfo, err := redisCache.GetCacheSize(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Total Keys: %d\n", sizeInfo.TotalKeys)
//	fmt.Printf("Memory Used: %d bytes\n", sizeInfo.MemoryUsed)
//
// # Retry and Circuit Breaker
//
// The package includes built-in retry logic with exponential backoff and circuit breaker patterns:
//
//	retryConfig := &cache.RedisRetryConfig{
//	    MaxAttempts:  5,                    // Maximum retry attempts
//	    InitialDelay: 100 * time.Millisecond, // Initial delay
//	    MaxDelay:     5 * time.Second,      // Maximum delay
//	    Multiplier:   2.0,                  // Backoff multiplier
//	    Jitter:       true,                 // Add random jitter
//	    RetryableOps: []string{"get", "set", "del"}, // Operations to retry
//	}
//
//	config := cache.DefaultRedisConfig()
//	config.RetryConfig = retryConfig
//
// # Thread Safety
//
// All cache operations are thread-safe and support concurrent access:
//
//	// Safe to call from multiple goroutines
//	go func() {
//	    redisCache.StoreDiagram(ctx, "diagram1", content1, time.Hour)
//	}()
//
//	go func() {
//	    redisCache.StoreDiagram(ctx, "diagram2", content2, time.Hour)
//	}()
//
// # Security Considerations
//
// The package includes comprehensive input validation and sanitization:
//   - All inputs are validated for length, format, and content
//   - Special characters are properly encoded for Redis keys
//   - SQL injection and XSS patterns are detected and rejected
//   - Path traversal attempts are blocked
//   - Control characters and null bytes are filtered
//
// # Performance Optimization
//
// The package is optimized for performance:
//   - Connection pooling for efficient Redis connections
//   - Batch operations for bulk deletions
//   - Lazy loading for large state machine entities
//   - Configurable timeouts and retry policies
//   - Memory usage monitoring and optimization
//
// # API Separation
//
// The package maintains a clear separation between public and internal APIs:
//
// Public API (cache package):
//   - Cache interface and RedisCache implementation
//   - Configuration types (RedisConfig, RedisRetryConfig)
//   - Error types and utility functions
//   - Health monitoring and cleanup types
//
// Internal API (internal package):
//   - Low-level Redis client wrapper
//   - Key generation and validation
//   - Input validation and sanitization
//   - Error recovery and circuit breaker implementation
//
// Users should only import and use the public cache package. The internal package
// is for implementation details and may change without notice.
//
// # Examples
//
// See the examples directory for complete usage examples:
//   - examples/diagram_cache_example/ - Basic diagram caching
//   - examples/state_machine_cache_example/ - State machine and entity caching
//   - examples/health_monitoring_example/ - Health monitoring and diagnostics
//   - examples/cleanup_example/ - Cache cleanup and management
//   - examples/error_handling_example/ - Comprehensive error handling
//
// # Testing
//
// The package includes comprehensive test coverage:
//   - Unit tests for all components (no Redis required)
//   - Integration tests with real Redis instances
//   - Performance and stress tests
//   - Concurrent access tests
//
// Run tests with:
//
//	go test ./cache -v                    # Unit tests
//	go test ./test/integration -v         # Integration tests (requires Redis)
//
// # Dependencies
//
// The package depends on:
//   - github.com/redis/go-redis/v9 - Redis client library
//   - github.com/kengibson1111/go-uml-statemachine-models - UML state machine models
//
// # License
//
// This package is licensed under the terms specified in the LICENSE file.
package cache
