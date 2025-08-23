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
}
