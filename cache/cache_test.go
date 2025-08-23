package cache

import (
	"context"
	"testing"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
)

// MockCache is a simple mock implementation of the Cache interface for testing
type MockCache struct{}

func (m *MockCache) StoreDiagram(ctx context.Context, name string, pumlContent string, ttl time.Duration) error {
	return nil
}

func (m *MockCache) GetDiagram(ctx context.Context, name string) (string, error) {
	return "", nil
}

func (m *MockCache) DeleteDiagram(ctx context.Context, name string) error {
	return nil
}

func (m *MockCache) StoreStateMachine(ctx context.Context, umlVersion, name string, machine *models.StateMachine, ttl time.Duration) error {
	return nil
}

func (m *MockCache) GetStateMachine(ctx context.Context, umlVersion, name string) (*models.StateMachine, error) {
	return nil, nil
}

func (m *MockCache) DeleteStateMachine(ctx context.Context, umlVersion, name string) error {
	return nil
}

func (m *MockCache) StoreEntity(ctx context.Context, umlVersion, diagramName, entityID string, entity interface{}, ttl time.Duration) error {
	return nil
}

func (m *MockCache) GetEntity(ctx context.Context, umlVersion, diagramName, entityID string) (interface{}, error) {
	return nil, nil
}

func (m *MockCache) GetEntityAsState(ctx context.Context, umlVersion, diagramName, entityID string) (*models.State, error) {
	return nil, nil
}

func (m *MockCache) GetEntityAsTransition(ctx context.Context, umlVersion, diagramName, entityID string) (*models.Transition, error) {
	return nil, nil
}

func (m *MockCache) GetEntityAsRegion(ctx context.Context, umlVersion, diagramName, entityID string) (*models.Region, error) {
	return nil, nil
}

func (m *MockCache) GetEntityAsVertex(ctx context.Context, umlVersion, diagramName, entityID string) (*models.Vertex, error) {
	return nil, nil
}

func (m *MockCache) Cleanup(ctx context.Context, pattern string) error {
	return nil
}

func (m *MockCache) Health(ctx context.Context) error {
	return nil
}

func (m *MockCache) Close() error {
	return nil
}

// TestCacheInterface verifies that MockCache implements the Cache interface
func TestCacheInterface(t *testing.T) {
	var _ Cache = &MockCache{}

	// Test that we can create a mock cache instance
	cache := &MockCache{}

	// Test basic interface compliance by calling methods
	ctx := context.Background()

	if err := cache.StoreDiagram(ctx, "test", "content", time.Hour); err != nil {
		t.Errorf("StoreDiagram failed: %v", err)
	}

	if _, err := cache.GetDiagram(ctx, "test"); err != nil {
		t.Errorf("GetDiagram failed: %v", err)
	}

	if err := cache.Health(ctx); err != nil {
		t.Errorf("Health failed: %v", err)
	}

	if err := cache.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
