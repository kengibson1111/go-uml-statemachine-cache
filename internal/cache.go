package internal

import (
	"context"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/internal/models"
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

	// Management operations
	Cleanup(ctx context.Context, pattern string) error
	Health(ctx context.Context) error
	Close() error
}
