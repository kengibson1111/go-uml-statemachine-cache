package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_StoreEntity(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		umlVersion  string
		diagramName string
		entityID    string
		entity      interface{}
		ttl         time.Duration
		expectError bool
		errorType   ErrorType
		setupMocks  func(*MockRedisClient, *MockKeyGenerator)
	}{
		{
			name:        "valid entity storage",
			umlVersion:  "1.0",
			diagramName: "TestDiagram",
			entityID:    "state-1",
			entity: &models.State{
				Vertex: models.Vertex{
					ID:   "state-1",
					Name: "State1",
					Type: "state",
				},
				IsSimple: true,
			},
			ttl:         time.Hour,
			expectError: false,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				expectedKey := "/machines/1.0/TestDiagram/entities/state-1"
				mockKeyGen.On("EntityKey", "1.0", "TestDiagram", "state-1").Return(expectedKey)
				mockKeyGen.On("ValidateKey", expectedKey).Return(nil)

				expectedEntity := &models.State{
					Vertex: models.Vertex{
						ID:   "state-1",
						Name: "State1",
						Type: "state",
					},
					IsSimple: true,
				}
				expectedData, _ := json.Marshal(expectedEntity)
				mockClient.On("SetWithRetry", ctx, expectedKey, expectedData, time.Hour).Return(nil)
			},
		},
		{
			name:        "empty UML version",
			umlVersion:  "",
			diagramName: "TestDiagram",
			entityID:    "state-1",
			entity:      &models.State{},
			ttl:         time.Hour,
			expectError: true,
			errorType:   ErrorTypeValidation,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// No mocks needed for validation errors
			},
		},
		{
			name:        "empty diagram name",
			umlVersion:  "1.0",
			diagramName: "",
			entityID:    "state-1",
			entity:      &models.State{},
			ttl:         time.Hour,
			expectError: true,
			errorType:   ErrorTypeValidation,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// No mocks needed for validation errors
			},
		},
		{
			name:        "empty entity ID",
			umlVersion:  "1.0",
			diagramName: "TestDiagram",
			entityID:    "",
			entity:      &models.State{},
			ttl:         time.Hour,
			expectError: true,
			errorType:   ErrorTypeValidation,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// No mocks needed for validation errors
			},
		},
		{
			name:        "nil entity",
			umlVersion:  "1.0",
			diagramName: "TestDiagram",
			entityID:    "state-1",
			entity:      nil,
			ttl:         time.Hour,
			expectError: true,
			errorType:   ErrorTypeValidation,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// No mocks needed for validation errors
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()
			config := &RedisConfig{DefaultTTL: time.Hour}

			tt.setupMocks(mockClient, mockKeyGen)

			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

			err := cache.StoreEntity(ctx, tt.umlVersion, tt.diagramName, tt.entityID, tt.entity, tt.ttl)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != 0 {
					cacheErr, ok := err.(*Error)
					require.True(t, ok, "Expected CacheError, got %T", err)
					assert.Equal(t, tt.errorType, cacheErr.Type)
				}
			} else {
				require.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
			mockKeyGen.AssertExpectations(t)
		})
	}
}

func TestRedisCache_GetEntity(t *testing.T) {
	ctx := context.Background()

	sampleState := &models.State{
		Vertex: models.Vertex{
			ID:   "state-1",
			Name: "State1",
			Type: "state",
		},
		IsSimple: true,
	}

	tests := []struct {
		name        string
		umlVersion  string
		diagramName string
		entityID    string
		expectError bool
		errorType   ErrorType
		setupMocks  func(*MockRedisClient, *MockKeyGenerator)
		expected    interface{}
	}{
		{
			name:        "valid entity retrieval",
			umlVersion:  "1.0",
			diagramName: "TestDiagram",
			entityID:    "state-1",
			expectError: false,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				expectedKey := "/machines/1.0/TestDiagram/entities/state-1"
				mockKeyGen.On("EntityKey", "1.0", "TestDiagram", "state-1").Return(expectedKey)
				mockKeyGen.On("ValidateKey", expectedKey).Return(nil)

				expectedData, _ := json.Marshal(sampleState)
				mockClient.On("GetWithRetry", ctx, expectedKey).Return(string(expectedData), nil)
			},
			expected: sampleState,
		},
		{
			name:        "entity not found",
			umlVersion:  "1.0",
			diagramName: "TestDiagram",
			entityID:    "nonexistent",
			expectError: true,
			errorType:   ErrorTypeNotFound,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				expectedKey := "/machines/1.0/TestDiagram/entities/nonexistent"
				mockKeyGen.On("EntityKey", "1.0", "TestDiagram", "nonexistent").Return(expectedKey)
				mockKeyGen.On("ValidateKey", expectedKey).Return(nil)
				mockClient.On("GetWithRetry", ctx, expectedKey).Return("", redis.Nil)
			},
		},
		{
			name:        "empty UML version",
			umlVersion:  "",
			diagramName: "TestDiagram",
			entityID:    "state-1",
			expectError: true,
			errorType:   ErrorTypeValidation,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// No mocks needed for validation errors
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()
			config := &RedisConfig{DefaultTTL: time.Hour}

			tt.setupMocks(mockClient, mockKeyGen)

			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

			result, err := cache.GetEntity(ctx, tt.umlVersion, tt.diagramName, tt.entityID)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != 0 {
					cacheErr, ok := err.(*Error)
					require.True(t, ok, "Expected CacheError, got %T", err)
					assert.Equal(t, tt.errorType, cacheErr.Type)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				// Note: Due to JSON marshaling/unmarshaling, we need to compare the JSON representations
				expectedJSON, _ := json.Marshal(tt.expected)
				resultJSON, _ := json.Marshal(result)
				assert.JSONEq(t, string(expectedJSON), string(resultJSON))
			}

			mockClient.AssertExpectations(t)
			mockKeyGen.AssertExpectations(t)
		})
	}
}

func TestRedisCache_GetEntityAsState(t *testing.T) {
	ctx := context.Background()

	sampleState := &models.State{
		Vertex: models.Vertex{
			ID:   "state-1",
			Name: "State1",
			Type: "state",
		},
		IsSimple: true,
	}

	tests := []struct {
		name        string
		umlVersion  string
		diagramName string
		entityID    string
		expectError bool
		setupMocks  func(*MockRedisClient, *MockKeyGenerator)
		expected    *models.State
	}{
		{
			name:        "valid state retrieval",
			umlVersion:  "1.0",
			diagramName: "TestDiagram",
			entityID:    "state-1",
			expectError: false,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				expectedKey := "/machines/1.0/TestDiagram/entities/state-1"
				mockKeyGen.On("EntityKey", "1.0", "TestDiagram", "state-1").Return(expectedKey)
				mockKeyGen.On("ValidateKey", expectedKey).Return(nil)

				expectedData, _ := json.Marshal(sampleState)
				mockClient.On("GetWithRetry", ctx, expectedKey).Return(string(expectedData), nil)
			},
			expected: sampleState,
		},
		{
			name:        "state not found",
			umlVersion:  "1.0",
			diagramName: "TestDiagram",
			entityID:    "nonexistent",
			expectError: true,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				expectedKey := "/machines/1.0/TestDiagram/entities/nonexistent"
				mockKeyGen.On("EntityKey", "1.0", "TestDiagram", "nonexistent").Return(expectedKey)
				mockKeyGen.On("ValidateKey", expectedKey).Return(nil)
				mockClient.On("GetWithRetry", ctx, expectedKey).Return("", redis.Nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()
			config := &RedisConfig{DefaultTTL: time.Hour}

			tt.setupMocks(mockClient, mockKeyGen)

			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

			result, err := cache.GetEntityAsState(ctx, tt.umlVersion, tt.diagramName, tt.entityID)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.Type, result.Type)
				assert.Equal(t, tt.expected.IsSimple, result.IsSimple)
			}

			mockClient.AssertExpectations(t)
			mockKeyGen.AssertExpectations(t)
		})
	}
}

func TestRedisCache_GetEntityAsTransition(t *testing.T) {
	ctx := context.Background()

	sampleTransition := &models.Transition{
		ID:   "transition-1",
		Name: "Transition1",
		Kind: models.TransitionKindExternal,
	}

	tests := []struct {
		name        string
		umlVersion  string
		diagramName string
		entityID    string
		expectError bool
		setupMocks  func(*MockRedisClient, *MockKeyGenerator)
		expected    *models.Transition
	}{
		{
			name:        "valid transition retrieval",
			umlVersion:  "1.0",
			diagramName: "TestDiagram",
			entityID:    "transition-1",
			expectError: false,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				expectedKey := "/machines/1.0/TestDiagram/entities/transition-1"
				mockKeyGen.On("EntityKey", "1.0", "TestDiagram", "transition-1").Return(expectedKey)
				mockKeyGen.On("ValidateKey", expectedKey).Return(nil)

				expectedData, _ := json.Marshal(sampleTransition)
				mockClient.On("GetWithRetry", ctx, expectedKey).Return(string(expectedData), nil)
			},
			expected: sampleTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()
			config := &RedisConfig{DefaultTTL: time.Hour}

			tt.setupMocks(mockClient, mockKeyGen)

			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

			result, err := cache.GetEntityAsTransition(ctx, tt.umlVersion, tt.diagramName, tt.entityID)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.Kind, result.Kind)
			}

			mockClient.AssertExpectations(t)
			mockKeyGen.AssertExpectations(t)
		})
	}
}

func TestRedisCache_ExtractEntitiesFromStateMachine(t *testing.T) {
	// Create a sample state machine with nested entities
	sampleStateMachine := &models.StateMachine{
		ID:      "test-machine",
		Name:    "TestMachine",
		Version: "1.0",
		Regions: []*models.Region{
			{
				ID:   "region-1",
				Name: "MainRegion",
				States: []*models.State{
					{
						Vertex: models.Vertex{
							ID:   "state-1",
							Name: "State1",
							Type: "state",
						},
						IsSimple: true,
					},
					{
						Vertex: models.Vertex{
							ID:   "state-2",
							Name: "State2",
							Type: "state",
						},
						IsComposite: true,
						Regions: []*models.Region{
							{
								ID:   "nested-region-1",
								Name: "NestedRegion",
								States: []*models.State{
									{
										Vertex: models.Vertex{
											ID:   "nested-state-1",
											Name: "NestedState1",
											Type: "state",
										},
										IsSimple: true,
									},
								},
							},
						},
					},
				},
				Transitions: []*models.Transition{
					{
						ID:   "transition-1",
						Name: "Transition1",
						Kind: models.TransitionKindExternal,
					},
				},
				Vertices: []*models.Vertex{
					{
						ID:   "vertex-1",
						Name: "Vertex1",
						Type: "pseudostate",
					},
				},
			},
		},
	}

	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := &RedisConfig{DefaultTTL: time.Hour}

	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	entities := cache.extractEntitiesFromStateMachine(sampleStateMachine)

	// Verify that all entities were extracted
	expectedEntityIDs := []string{
		"region-1",
		"state-1",
		"state-2",
		"nested-region-1",
		"nested-state-1",
		"transition-1",
		"vertex-1",
	}

	assert.Len(t, entities, len(expectedEntityIDs))

	for _, entityID := range expectedEntityIDs {
		assert.Contains(t, entities, entityID, "Entity %s should be extracted", entityID)
	}

	// Verify specific entity types
	region, ok := entities["region-1"].(*models.Region)
	assert.True(t, ok, "region-1 should be a Region")
	assert.Equal(t, "MainRegion", region.Name)

	state, ok := entities["state-1"].(*models.State)
	assert.True(t, ok, "state-1 should be a State")
	assert.Equal(t, "State1", state.Name)
	assert.True(t, state.IsSimple)

	transition, ok := entities["transition-1"].(*models.Transition)
	assert.True(t, ok, "transition-1 should be a Transition")
	assert.Equal(t, "Transition1", transition.Name)

	vertex, ok := entities["vertex-1"].(*models.Vertex)
	assert.True(t, ok, "vertex-1 should be a Vertex")
	assert.Equal(t, "Vertex1", vertex.Name)
}

func TestRedisCache_StoreStateMachine_WithEntityExtraction(t *testing.T) {
	ctx := context.Background()

	// Create a sample state machine with entities
	sampleStateMachine := &models.StateMachine{
		ID:      "test-machine",
		Name:    "TestMachine",
		Version: "1.0",
		Regions: []*models.Region{
			{
				ID:   "region-1",
				Name: "MainRegion",
				States: []*models.State{
					{
						Vertex: models.Vertex{
							ID:   "state-1",
							Name: "State1",
							Type: "state",
						},
						IsSimple: true,
					},
				},
				Transitions: []*models.Transition{
					{
						ID:   "transition-1",
						Name: "Transition1",
						Kind: models.TransitionKindExternal,
					},
				},
			},
		},
	}

	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := &RedisConfig{DefaultTTL: time.Hour}

	// Setup mocks for diagram existence check
	mockKeyGen.On("DiagramKey", "TestMachine").Return(fmt.Sprintf("/diagrams/%s/TestMachine", models.DiagramTypePUML.String()))
	mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/TestMachine", models.DiagramTypePUML.String())).Return(nil)
	mockClient.On("GetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/TestMachine", models.DiagramTypePUML.String())).Return("@startuml\nstate A\n@enduml", nil)

	// Setup mocks for state machine key generation and validation
	mockKeyGen.On("StateMachineKey", "1.0", "TestMachine").Return("/machines/1.0/TestMachine")
	mockKeyGen.On("ValidateKey", "/machines/1.0/TestMachine").Return(nil)

	// Setup mocks for entity key generation and storage
	expectedEntityKeys := map[string]string{
		"region-1":     "/machines/1.0/TestMachine/entities/region-1",
		"state-1":      "/machines/1.0/TestMachine/entities/state-1",
		"transition-1": "/machines/1.0/TestMachine/entities/transition-1",
	}

	for entityID, entityKey := range expectedEntityKeys {
		mockKeyGen.On("EntityKey", "1.0", "TestMachine", entityID).Return(entityKey)
		mockKeyGen.On("ValidateKey", entityKey).Return(nil)
		// Mock the entity storage calls
		mockClient.On("SetWithRetry", ctx, entityKey, mock.AnythingOfType("[]uint8"), time.Hour).Return(nil)
	}

	// Mock the final state machine storage (with populated Entities map)
	mockClient.On("SetWithRetry", ctx, "/machines/1.0/TestMachine", mock.AnythingOfType("[]uint8"), time.Hour).Return(nil)

	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	err := cache.StoreStateMachine(ctx, "1.0", models.DiagramTypePUML, "TestMachine", sampleStateMachine, time.Hour)

	require.NoError(t, err)

	// Verify that the Entities map was populated
	assert.NotNil(t, sampleStateMachine.Entities)
	assert.Len(t, sampleStateMachine.Entities, 3)

	for entityID, expectedKey := range expectedEntityKeys {
		actualKey, exists := sampleStateMachine.Entities[entityID]
		assert.True(t, exists, "Entity %s should be in Entities map", entityID)
		assert.Equal(t, expectedKey, actualKey, "Entity %s should have correct cache key", entityID)
	}

	mockClient.AssertExpectations(t)
	mockKeyGen.AssertExpectations(t)
}
