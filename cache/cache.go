package cache

import (
	"context"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
)

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
}
