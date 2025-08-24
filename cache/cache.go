// Package cache provides a comprehensive Redis-based caching system for PlantUML diagrams
// and UML state machine definitions with enterprise-grade reliability features.
//
// This package implements a hierarchical caching layer that supports:
//   - PlantUML diagram storage and retrieval with configurable TTL
//   - Parsed UML state machine caching with version management
//   - Individual state machine entity caching with referential integrity
//   - Comprehensive error handling with typed errors and recovery strategies
//   - Configurable retry logic with exponential backoff and circuit breakers
//   - Health monitoring, diagnostics, and performance metrics
//   - Thread-safe concurrent access with connection pooling
//   - Security-focused input validation and sanitization
//
// # Key Structure
//
// The cache uses a hierarchical key structure in Redis:
//   - PlantUML Diagrams: /diagrams/puml/<diagram_name>
//   - Parsed State Machines: /machines/<uml_version>/<diagram_name>
//   - State Machine Entities: /machines/<uml_version>/<diagram_name>/entities/<entity_id>
//
// # Basic Usage
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
//	ctx := context.Background()
//
//	// Store and retrieve diagrams
//	err = redisCache.StoreDiagram(ctx, "my-diagram", pumlContent, time.Hour)
//	content, err := redisCache.GetDiagram(ctx, "my-diagram")
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
//	    case cache.IsConnectionError(err):
//	        // Handle Redis connection issues
//	    case cache.IsValidationError(err):
//	        // Handle input validation errors
//	    }
//	}
//
// # Health Monitoring
//
//	// Basic health check
//	err = redisCache.Health(ctx)
//
//	// Detailed health status
//	health, err := redisCache.HealthDetailed(ctx)
//	fmt.Printf("Status: %s, Response Time: %v\n", health.Status, health.ResponseTime)
//
// # API Separation
//
// This package provides the public API for the cache system. Internal implementation
// details are in the internal package and should not be used directly by applications.
//
// Public API (this package):
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
// # Thread Safety
//
// All operations are thread-safe and support concurrent access from multiple goroutines.
// The library handles internal synchronization and connection pooling automatically.
//
// # Performance
//
// The library is optimized for high-performance scenarios with:
//   - Connection pooling for efficient Redis connections
//   - Batch operations for bulk deletions
//   - Configurable timeouts and retry policies
//   - Memory usage monitoring and optimization
//
// # Security
//
// The library includes comprehensive security measures:
//   - Input validation and sanitization for all parameters
//   - SQL injection and XSS pattern detection
//   - Path traversal prevention
//   - Safe key generation with proper encoding
//
// For complete documentation, examples, and configuration guides, see:
//   - README.md - Comprehensive usage guide
//   - API.md - Complete API reference
//   - CONFIGURATION.md - Configuration and tuning guide
//   - examples/ directory - Working code examples
package cache

import (
	"context"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
)

// CleanupResult contains information about a cleanup operation
type CleanupResult struct {
	KeysScanned    int64         `json:"keys_scanned"`
	KeysDeleted    int64         `json:"keys_deleted"`
	BytesFreed     int64         `json:"bytes_freed"`
	Duration       time.Duration `json:"duration"`
	BatchesUsed    int           `json:"batches_used"`
	ErrorsOccurred int           `json:"errors_occurred"`
}

// CleanupOptions configures cleanup behavior
type CleanupOptions struct {
	BatchSize      int           `json:"batch_size"`      // Number of keys to delete per batch
	ScanCount      int64         `json:"scan_count"`      // Number of keys to scan per SCAN operation
	MaxKeys        int64         `json:"max_keys"`        // Maximum number of keys to delete (0 = no limit)
	DryRun         bool          `json:"dry_run"`         // If true, only scan but don't delete
	Timeout        time.Duration `json:"timeout"`         // Timeout for the entire cleanup operation
	CollectMetrics bool          `json:"collect_metrics"` // Whether to collect detailed metrics
}

// CacheSizeInfo contains information about cache size and usage
type CacheSizeInfo struct {
	TotalKeys         int64 `json:"total_keys"`
	DiagramCount      int64 `json:"diagram_count"`
	StateMachineCount int64 `json:"state_machine_count"`
	EntityCount       int64 `json:"entity_count"`
	MemoryUsed        int64 `json:"memory_used"`     // Bytes
	MemoryPeak        int64 `json:"memory_peak"`     // Bytes
	MemoryOverhead    int64 `json:"memory_overhead"` // Bytes
}

// HealthStatus represents the overall health status of the cache system
type HealthStatus struct {
	Status       string             `json:"status"`             // "healthy", "degraded", "unhealthy"
	Timestamp    time.Time          `json:"timestamp"`          // When the health check was performed
	ResponseTime time.Duration      `json:"response_time"`      // Time taken for health check
	Connection   ConnectionHealth   `json:"connection"`         // Connection-specific health info
	Performance  PerformanceMetrics `json:"performance"`        // Performance metrics
	Diagnostics  DiagnosticInfo     `json:"diagnostics"`        // Diagnostic information
	Errors       []string           `json:"errors,omitempty"`   // Any errors encountered
	Warnings     []string           `json:"warnings,omitempty"` // Any warnings
}

// ConnectionHealth contains connection-specific health information
type ConnectionHealth struct {
	Connected   bool                `json:"connected"`            // Whether Redis is connected
	Address     string              `json:"address"`              // Redis server address
	Database    int                 `json:"database"`             // Redis database number
	PingLatency time.Duration       `json:"ping_latency"`         // Latency of ping command
	PoolStats   ConnectionPoolStats `json:"pool_stats"`           // Connection pool statistics
	LastError   string              `json:"last_error,omitempty"` // Last connection error if any
}

// ConnectionPoolStats contains Redis connection pool statistics
type ConnectionPoolStats struct {
	TotalConnections  int `json:"total_connections"`  // Total connections in pool
	IdleConnections   int `json:"idle_connections"`   // Idle connections
	ActiveConnections int `json:"active_connections"` // Active connections
	Hits              int `json:"hits"`               // Pool hits
	Misses            int `json:"misses"`             // Pool misses
	Timeouts          int `json:"timeouts"`           // Pool timeouts
	StaleConnections  int `json:"stale_connections"`  // Stale connections
}

// PerformanceMetrics contains performance-related metrics
type PerformanceMetrics struct {
	MemoryUsage    MemoryMetrics    `json:"memory_usage"`    // Memory usage information
	KeyspaceInfo   KeyspaceMetrics  `json:"keyspace_info"`   // Keyspace statistics
	OperationStats OperationMetrics `json:"operation_stats"` // Operation statistics
	ServerInfo     ServerMetrics    `json:"server_info"`     // Server information
}

// MemoryMetrics contains memory usage information
type MemoryMetrics struct {
	UsedMemory          int64   `json:"used_memory"`          // Used memory in bytes
	UsedMemoryHuman     string  `json:"used_memory_human"`    // Human readable used memory
	UsedMemoryPeak      int64   `json:"used_memory_peak"`     // Peak memory usage
	MemoryFragmentation float64 `json:"memory_fragmentation"` // Memory fragmentation ratio
	MaxMemory           int64   `json:"max_memory"`           // Maximum memory limit
	MaxMemoryPolicy     string  `json:"max_memory_policy"`    // Memory eviction policy
}

// KeyspaceMetrics contains keyspace statistics
type KeyspaceMetrics struct {
	TotalKeys      int64   `json:"total_keys"`       // Total number of keys
	ExpiringKeys   int64   `json:"expiring_keys"`    // Keys with expiration
	AverageKeySize float64 `json:"average_key_size"` // Average key size in bytes
	KeyspaceHits   int64   `json:"keyspace_hits"`    // Keyspace hits
	KeyspaceMisses int64   `json:"keyspace_misses"`  // Keyspace misses
	HitRate        float64 `json:"hit_rate"`         // Cache hit rate percentage
}

// OperationMetrics contains operation statistics
type OperationMetrics struct {
	TotalCommands       int64   `json:"total_commands"`       // Total commands processed
	CommandsPerSecond   float64 `json:"commands_per_sec"`     // Commands per second
	ConnectedClients    int64   `json:"connected_clients"`    // Number of connected clients
	BlockedClients      int64   `json:"blocked_clients"`      // Number of blocked clients
	RejectedConnections int64   `json:"rejected_connections"` // Rejected connections
}

// ServerMetrics contains server information
type ServerMetrics struct {
	RedisVersion    string `json:"redis_version"`    // Redis server version
	UptimeSeconds   int64  `json:"uptime_seconds"`   // Server uptime in seconds
	UptimeDays      int64  `json:"uptime_days"`      // Server uptime in days
	ServerMode      string `json:"server_mode"`      // Server mode (standalone, sentinel, cluster)
	Role            string `json:"role"`             // Server role (master, slave)
	ConnectedSlaves int64  `json:"connected_slaves"` // Number of connected slaves
}

// DiagnosticInfo contains diagnostic information for troubleshooting
type DiagnosticInfo struct {
	ConfigurationCheck ConfigCheck        `json:"configuration"`             // Configuration validation
	NetworkCheck       NetworkCheck       `json:"network"`                   // Network connectivity check
	PerformanceCheck   PerformanceCheck   `json:"performance"`               // Performance validation
	DataIntegrityCheck DataIntegrityCheck `json:"data_integrity"`            // Data integrity validation
	Recommendations    []string           `json:"recommendations,omitempty"` // Performance recommendations
}

// ConfigCheck contains configuration validation results
type ConfigCheck struct {
	Valid           bool     `json:"valid"`                     // Whether configuration is valid
	Issues          []string `json:"issues,omitempty"`          // Configuration issues found
	Recommendations []string `json:"recommendations,omitempty"` // Configuration recommendations
}

// NetworkCheck contains network connectivity validation results
type NetworkCheck struct {
	Reachable  bool          `json:"reachable"`        // Whether Redis server is reachable
	Latency    time.Duration `json:"latency"`          // Network latency
	PacketLoss float64       `json:"packet_loss"`      // Packet loss percentage
	Issues     []string      `json:"issues,omitempty"` // Network issues found
}

// PerformanceCheck contains performance validation results
type PerformanceCheck struct {
	Acceptable      bool     `json:"acceptable"`                // Whether performance is acceptable
	Bottlenecks     []string `json:"bottlenecks,omitempty"`     // Performance bottlenecks identified
	Recommendations []string `json:"recommendations,omitempty"` // Performance recommendations
}

// DataIntegrityCheck contains data integrity validation results
type DataIntegrityCheck struct {
	Consistent   bool     `json:"consistent"`       // Whether data is consistent
	Issues       []string `json:"issues,omitempty"` // Data integrity issues found
	OrphanedKeys int64    `json:"orphaned_keys"`    // Number of orphaned keys found
}

// Cache defines the interface for caching PlantUML diagrams and parsed state machines
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
