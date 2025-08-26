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

func TestRedisCache_ReferentialIntegrity_EntityMappingConsistency(t *testing.T) {
	ctx := context.Background()

	// Create a complex state machine with multiple entities
	complexStateMachine := &models.StateMachine{
		ID:      "complex-machine",
		Name:    "ComplexMachine",
		Version: "1.0",
		Regions: []*models.Region{
			{
				ID:   "main-region",
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
							ID:   "composite-state",
							Name: "CompositeState",
							Type: "state",
						},
						IsComposite: true,
						Regions: []*models.Region{
							{
								ID:   "nested-region",
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
								Transitions: []*models.Transition{
									{
										ID:   "nested-transition",
										Name: "NestedTransition",
										Kind: models.TransitionKindInternal,
									},
								},
							},
						},
					},
				},
				Transitions: []*models.Transition{
					{
						ID:   "main-transition",
						Name: "MainTransition",
						Kind: models.TransitionKindExternal,
					},
				},
				Vertices: []*models.Vertex{
					{
						ID:   "initial-vertex",
						Name: "InitialVertex",
						Type: "pseudostate",
					},
				},
			},
		},
	}

	t.Run("entities mapping is populated correctly during storage", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := &RedisConfig{DefaultTTL: time.Hour}

		// Setup mocks for diagram existence check
		mockKeyGen.On("DiagramKey", "ComplexMachine").Return(fmt.Sprintf("/diagrams/%s/ComplexMachine", models.DiagramTypePUML.String()))
		mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/ComplexMachine", models.DiagramTypePUML.String())).Return(nil)
		mockClient.On("GetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/ComplexMachine", models.DiagramTypePUML.String())).Return("@startuml\nstate A\n@enduml", nil)

		// Setup mocks for state machine key
		mockKeyGen.On("StateMachineKey", "1.0", "ComplexMachine").Return("/machines/1.0/ComplexMachine")
		mockKeyGen.On("ValidateKey", "/machines/1.0/ComplexMachine").Return(nil)

		// Expected entities that should be extracted
		expectedEntities := map[string]string{
			"main-region":       "/machines/1.0/ComplexMachine/entities/main-region",
			"state-1":           "/machines/1.0/ComplexMachine/entities/state-1",
			"composite-state":   "/machines/1.0/ComplexMachine/entities/composite-state",
			"nested-region":     "/machines/1.0/ComplexMachine/entities/nested-region",
			"nested-state-1":    "/machines/1.0/ComplexMachine/entities/nested-state-1",
			"nested-transition": "/machines/1.0/ComplexMachine/entities/nested-transition",
			"main-transition":   "/machines/1.0/ComplexMachine/entities/main-transition",
			"initial-vertex":    "/machines/1.0/ComplexMachine/entities/initial-vertex",
		}

		// Setup mocks for entity storage
		for entityID, entityKey := range expectedEntities {
			mockKeyGen.On("EntityKey", "1.0", "ComplexMachine", entityID).Return(entityKey)
			mockKeyGen.On("ValidateKey", entityKey).Return(nil)
			mockClient.On("SetWithRetry", ctx, entityKey, mock.AnythingOfType("[]uint8"), time.Hour).Return(nil)
		}

		// Mock the final state machine storage
		mockClient.On("SetWithRetry", ctx, "/machines/1.0/ComplexMachine", mock.AnythingOfType("[]uint8"), time.Hour).Return(nil)

		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		err := cache.StoreStateMachine(ctx, "1.0", models.DiagramTypePUML, "ComplexMachine", complexStateMachine, time.Hour)
		require.NoError(t, err)

		// Verify that all expected entities are in the Entities map
		assert.NotNil(t, complexStateMachine.Entities)
		assert.Len(t, complexStateMachine.Entities, len(expectedEntities))

		for entityID, expectedKey := range expectedEntities {
			actualKey, exists := complexStateMachine.Entities[entityID]
			assert.True(t, exists, "Entity %s should be in Entities map", entityID)
			assert.Equal(t, expectedKey, actualKey, "Entity %s should have correct cache key", entityID)
		}

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})

	t.Run("entities mapping is cleared and repopulated on re-storage", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := &RedisConfig{DefaultTTL: time.Hour}

		// Pre-populate the Entities map with some existing (potentially stale) entries
		complexStateMachine.Entities = map[string]string{
			"old-entity-1": "/machines/1.0/ComplexMachine/entities/old-entity-1",
			"old-entity-2": "/machines/1.0/ComplexMachine/entities/old-entity-2",
		}

		// Setup mocks for diagram existence check
		mockKeyGen.On("DiagramKey", "ComplexMachine").Return(fmt.Sprintf("/diagrams/%s/ComplexMachine", models.DiagramTypePUML.String()))
		mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/ComplexMachine", models.DiagramTypePUML.String())).Return(nil)
		mockClient.On("GetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/ComplexMachine", models.DiagramTypePUML.String())).Return("@startuml\nstate A\n@enduml", nil)

		// Setup mocks for state machine key
		mockKeyGen.On("StateMachineKey", "1.0", "ComplexMachine").Return("/machines/1.0/ComplexMachine")
		mockKeyGen.On("ValidateKey", "/machines/1.0/ComplexMachine").Return(nil)

		// Expected entities that should be extracted (same as before)
		expectedEntities := map[string]string{
			"main-region":       "/machines/1.0/ComplexMachine/entities/main-region",
			"state-1":           "/machines/1.0/ComplexMachine/entities/state-1",
			"composite-state":   "/machines/1.0/ComplexMachine/entities/composite-state",
			"nested-region":     "/machines/1.0/ComplexMachine/entities/nested-region",
			"nested-state-1":    "/machines/1.0/ComplexMachine/entities/nested-state-1",
			"nested-transition": "/machines/1.0/ComplexMachine/entities/nested-transition",
			"main-transition":   "/machines/1.0/ComplexMachine/entities/main-transition",
			"initial-vertex":    "/machines/1.0/ComplexMachine/entities/initial-vertex",
		}

		// Setup mocks for entity storage
		for entityID, entityKey := range expectedEntities {
			mockKeyGen.On("EntityKey", "1.0", "ComplexMachine", entityID).Return(entityKey)
			mockKeyGen.On("ValidateKey", entityKey).Return(nil)
			mockClient.On("SetWithRetry", ctx, entityKey, mock.AnythingOfType("[]uint8"), time.Hour).Return(nil)
		}

		// Mock the final state machine storage
		mockClient.On("SetWithRetry", ctx, "/machines/1.0/ComplexMachine", mock.AnythingOfType("[]uint8"), time.Hour).Return(nil)

		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		err := cache.StoreStateMachine(ctx, "1.0", models.DiagramTypePUML, "ComplexMachine", complexStateMachine, time.Hour)
		require.NoError(t, err)

		// Verify that old entities are gone and new entities are present
		assert.NotNil(t, complexStateMachine.Entities)
		assert.Len(t, complexStateMachine.Entities, len(expectedEntities))

		// Verify old entities are no longer present
		_, exists := complexStateMachine.Entities["old-entity-1"]
		assert.False(t, exists, "Old entity should be removed from Entities map")
		_, exists = complexStateMachine.Entities["old-entity-2"]
		assert.False(t, exists, "Old entity should be removed from Entities map")

		// Verify new entities are present
		for entityID, expectedKey := range expectedEntities {
			actualKey, exists := complexStateMachine.Entities[entityID]
			assert.True(t, exists, "Entity %s should be in Entities map", entityID)
			assert.Equal(t, expectedKey, actualKey, "Entity %s should have correct cache key", entityID)
		}

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})
}

func TestRedisCache_ReferentialIntegrity_PartialStorageFailure(t *testing.T) {
	ctx := context.Background()

	sampleStateMachine := &models.StateMachine{
		ID:      "test-machine",
		Name:    "TestMachine",
		Version: "1.0",
		Regions: []*models.Region{
			{
				ID:   "region-1",
				Name: "Region1",
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
			},
		},
	}

	t.Run("cleanup entities when entity key validation fails", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := &RedisConfig{DefaultTTL: time.Hour}

		// Setup mocks for diagram existence check
		mockKeyGen.On("DiagramKey", "TestMachine").Return(fmt.Sprintf("/diagrams/%s/TestMachine", models.DiagramTypePUML.String()))
		mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/TestMachine", models.DiagramTypePUML.String())).Return(nil)
		mockClient.On("GetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/TestMachine", models.DiagramTypePUML.String())).Return("@startuml\nstate A\n@enduml", nil)

		// Setup mocks for state machine key
		mockKeyGen.On("StateMachineKey", "1.0", "TestMachine").Return("/machines/1.0/TestMachine")
		mockKeyGen.On("ValidateKey", "/machines/1.0/TestMachine").Return(nil)

		// Setup mocks for entity key generation - first entity key validation fails
		mockKeyGen.On("EntityKey", "1.0", "TestMachine", "region-1").Return("/invalid/entity/key")
		mockKeyGen.On("ValidateKey", "/invalid/entity/key").Return(NewKeyInvalidError("/invalid/entity/key", "invalid key format"))

		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		err := cache.StoreStateMachine(ctx, "1.0", models.DiagramTypePUML, "TestMachine", sampleStateMachine, time.Hour)
		require.Error(t, err)
		cacheErr, ok := err.(*CacheError)
		require.True(t, ok, "Expected CacheError, got %T", err)
		assert.Equal(t, CacheErrorTypeKeyInvalid, cacheErr.Type)
		assert.Contains(t, err.Error(), "invalid entity key generated for 'region-1'")

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})

	t.Run("cleanup entities when state machine storage fails", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := &RedisConfig{DefaultTTL: time.Hour}

		// Setup mocks for diagram existence check
		mockKeyGen.On("DiagramKey", "TestMachine").Return(fmt.Sprintf("/diagrams/%s/TestMachine", models.DiagramTypePUML.String()))
		mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/TestMachine", models.DiagramTypePUML.String())).Return(nil)
		mockClient.On("GetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/TestMachine", models.DiagramTypePUML.String())).Return("@startuml\nstate A\n@enduml", nil)

		// Setup mocks for state machine key
		mockKeyGen.On("StateMachineKey", "1.0", "TestMachine").Return("/machines/1.0/TestMachine")
		mockKeyGen.On("ValidateKey", "/machines/1.0/TestMachine").Return(nil)

		// Setup mocks for successful entity storage
		mockKeyGen.On("EntityKey", "1.0", "TestMachine", "region-1").Return("/machines/1.0/TestMachine/entities/region-1")
		mockKeyGen.On("ValidateKey", "/machines/1.0/TestMachine/entities/region-1").Return(nil)
		mockClient.On("SetWithRetry", ctx, "/machines/1.0/TestMachine/entities/region-1", mock.AnythingOfType("[]uint8"), time.Hour).Return(nil)

		mockKeyGen.On("EntityKey", "1.0", "TestMachine", "state-1").Return("/machines/1.0/TestMachine/entities/state-1")
		mockKeyGen.On("ValidateKey", "/machines/1.0/TestMachine/entities/state-1").Return(nil)
		mockClient.On("SetWithRetry", ctx, "/machines/1.0/TestMachine/entities/state-1", mock.AnythingOfType("[]uint8"), time.Hour).Return(nil)

		// State machine storage fails
		mockClient.On("SetWithRetry", ctx, "/machines/1.0/TestMachine", mock.AnythingOfType("[]uint8"), time.Hour).Return(assert.AnError)

		// Expect cleanup of all stored entities (use flexible matcher for order)
		mockClient.On("DelWithRetry", ctx, mock.MatchedBy(func(keys []string) bool {
			// Should contain exactly 2 keys (both entities)
			if len(keys) != 2 {
				return false
			}
			// Check that both expected keys are present (order doesn't matter)
			hasRegion := false
			hasState := false
			for _, key := range keys {
				if key == "/machines/1.0/TestMachine/entities/region-1" {
					hasRegion = true
				} else if key == "/machines/1.0/TestMachine/entities/state-1" {
					hasState = true
				}
			}
			return hasRegion && hasState
		})).Return(nil)

		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		err := cache.StoreStateMachine(ctx, "1.0", models.DiagramTypePUML, "TestMachine", sampleStateMachine, time.Hour)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to store state machine 'TestMachine'")

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})
}

func TestRedisCache_ReferentialIntegrity_CascadeDeletion(t *testing.T) {
	ctx := context.Background()

	sampleStateMachineWithEntities := &models.StateMachine{
		ID:      "test-machine-cascade",
		Name:    "TestMachineCascade",
		Version: "1.0",
		Entities: map[string]string{
			"region-1":     "/machines/1.0/TestMachineCascade/entities/region-1",
			"state-1":      "/machines/1.0/TestMachineCascade/entities/state-1",
			"transition-1": "/machines/1.0/TestMachineCascade/entities/transition-1",
		},
	}

	t.Run("cascade deletion removes all tracked entities", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := &RedisConfig{DefaultTTL: time.Hour}

		// Setup mocks for state machine key
		mockKeyGen.On("StateMachineKey", "1.0", "TestMachineCascade").Return("/machines/1.0/TestMachineCascade")
		mockKeyGen.On("ValidateKey", "/machines/1.0/TestMachineCascade").Return(nil)

		// Mock getting the state machine for cascade deletion
		expectedData, _ := json.Marshal(sampleStateMachineWithEntities)
		mockClient.On("GetWithRetry", ctx, "/machines/1.0/TestMachineCascade").Return(string(expectedData), nil)

		// Mock entity key generation for cascade deletion
		mockKeyGen.On("EntityKey", "1.0", "TestMachineCascade", "region-1").Return("/machines/1.0/TestMachineCascade/entities/region-1")
		mockKeyGen.On("EntityKey", "1.0", "TestMachineCascade", "state-1").Return("/machines/1.0/TestMachineCascade/entities/state-1")
		mockKeyGen.On("EntityKey", "1.0", "TestMachineCascade", "transition-1").Return("/machines/1.0/TestMachineCascade/entities/transition-1")

		// Mock entity key generation for orphaned entity scan
		mockKeyGen.On("EntityKey", "1.0", "TestMachineCascade", "*").Return("/machines/1.0/TestMachineCascade/entities/*")

		// Mock Redis client for orphaned entity scan - return nil to skip the scan
		mockClient.On("Client").Return(nil)

		// Mock the deletion of state machine and entities
		expectedKeys := []string{
			"/machines/1.0/TestMachineCascade",
			"/machines/1.0/TestMachineCascade/entities/region-1",
			"/machines/1.0/TestMachineCascade/entities/state-1",
			"/machines/1.0/TestMachineCascade/entities/transition-1",
		}
		mockClient.On("DelWithRetry", ctx, expectedKeys).Return(nil)

		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		err := cache.DeleteStateMachine(ctx, "1.0", "TestMachineCascade")
		require.NoError(t, err)

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})

	t.Run("cascade deletion handles missing state machine gracefully", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := &RedisConfig{DefaultTTL: time.Hour}

		// Setup mocks for state machine key
		mockKeyGen.On("StateMachineKey", "1.0", "NonExistentMachine").Return("/machines/1.0/NonExistentMachine")
		mockKeyGen.On("ValidateKey", "/machines/1.0/NonExistentMachine").Return(nil)

		// Mock getting the state machine - not found
		mockClient.On("GetWithRetry", ctx, "/machines/1.0/NonExistentMachine").Return("", redis.Nil)

		// Mock entity key generation for orphaned entity scan
		mockKeyGen.On("EntityKey", "1.0", "NonExistentMachine", "*").Return("/machines/1.0/NonExistentMachine/entities/*")

		// Mock Redis client for orphaned entity scan - return nil to skip the scan
		mockClient.On("Client").Return(nil)

		// Mock the deletion of just the state machine key
		mockClient.On("DelWithRetry", ctx, []string{"/machines/1.0/NonExistentMachine"}).Return(nil)

		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		err := cache.DeleteStateMachine(ctx, "1.0", "NonExistentMachine")
		require.NoError(t, err)

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})
}

func TestRedisCache_UpdateStateMachineEntityMapping(t *testing.T) {
	ctx := context.Background()

	sampleStateMachine := &models.StateMachine{
		ID:      "test-machine-update",
		Name:    "TestMachineUpdate",
		Version: "1.0",
		Entities: map[string]string{
			"existing-entity": "/machines/1.0/TestMachineUpdate/entities/existing-entity",
		},
	}

	t.Run("add entity to mapping", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := &RedisConfig{DefaultTTL: time.Hour}

		// Mock getting the current state machine
		mockKeyGen.On("StateMachineKey", "1.0", "TestMachineUpdate").Return("/machines/1.0/TestMachineUpdate")
		mockKeyGen.On("ValidateKey", "/machines/1.0/TestMachineUpdate").Return(nil)
		expectedData, _ := json.Marshal(sampleStateMachine)
		mockClient.On("GetWithRetry", ctx, "/machines/1.0/TestMachineUpdate").Return(string(expectedData), nil)

		// Mock storing the updated state machine
		mockClient.On("SetWithRetry", ctx, "/machines/1.0/TestMachineUpdate", mock.AnythingOfType("[]uint8"), time.Hour).Return(nil)

		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		err := cache.UpdateStateMachineEntityMapping(ctx, "1.0", "TestMachineUpdate", "new-entity", "/machines/1.0/TestMachineUpdate/entities/new-entity", "add")
		require.NoError(t, err)

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})

	t.Run("remove entity from mapping", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := &RedisConfig{DefaultTTL: time.Hour}

		// Mock getting the current state machine
		mockKeyGen.On("StateMachineKey", "1.0", "TestMachineUpdate").Return("/machines/1.0/TestMachineUpdate")
		mockKeyGen.On("ValidateKey", "/machines/1.0/TestMachineUpdate").Return(nil)
		expectedData, _ := json.Marshal(sampleStateMachine)
		mockClient.On("GetWithRetry", ctx, "/machines/1.0/TestMachineUpdate").Return(string(expectedData), nil)

		// Mock storing the updated state machine
		mockClient.On("SetWithRetry", ctx, "/machines/1.0/TestMachineUpdate", mock.AnythingOfType("[]uint8"), time.Hour).Return(nil)

		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		err := cache.UpdateStateMachineEntityMapping(ctx, "1.0", "TestMachineUpdate", "existing-entity", "", "remove")
		require.NoError(t, err)

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})

	t.Run("validation errors", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := &RedisConfig{DefaultTTL: time.Hour}

		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		// Test empty UML version
		err := cache.UpdateStateMachineEntityMapping(ctx, "", "TestMachine", "entity-1", "key", "add")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "UML version cannot be empty")

		// Test empty state machine name
		err = cache.UpdateStateMachineEntityMapping(ctx, "1.0", "", "entity-1", "key", "add")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "state machine name cannot be empty")

		// Test empty entity ID
		err = cache.UpdateStateMachineEntityMapping(ctx, "1.0", "TestMachine", "", "key", "add")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "entity ID cannot be empty")

		// Test invalid operation
		err = cache.UpdateStateMachineEntityMapping(ctx, "1.0", "TestMachine", "entity-1", "key", "invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation must be 'add' or 'remove'")

		// Test empty entity key for add operation
		err = cache.UpdateStateMachineEntityMapping(ctx, "1.0", "TestMachine", "entity-1", "", "add")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "entity key cannot be empty for add operation")
	})
}

func TestRedisCache_ReferentialIntegrity_EntityKeyValidation(t *testing.T) {
	ctx := context.Background()

	sampleStateMachine := &models.StateMachine{
		ID:      "test-machine-validation",
		Name:    "TestMachineValidation",
		Version: "1.0",
		Regions: []*models.Region{
			{
				ID:   "region-1",
				Name: "Region1",
			},
		},
	}

	t.Run("entity key validation failure during storage", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := &RedisConfig{DefaultTTL: time.Hour}

		// Setup mocks for diagram existence check
		mockKeyGen.On("DiagramKey", "TestMachineValidation").Return(fmt.Sprintf("/diagrams/%s/TestMachineValidation", models.DiagramTypePUML.String()))
		mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/TestMachineValidation", models.DiagramTypePUML.String())).Return(nil)
		mockClient.On("GetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/TestMachineValidation", models.DiagramTypePUML.String())).Return("@startuml\nstate A\n@enduml", nil)

		// Setup mocks for state machine key
		mockKeyGen.On("StateMachineKey", "1.0", "TestMachineValidation").Return("/machines/1.0/TestMachineValidation")
		mockKeyGen.On("ValidateKey", "/machines/1.0/TestMachineValidation").Return(nil)

		// Setup mock for entity key generation - return invalid key
		mockKeyGen.On("EntityKey", "1.0", "TestMachineValidation", "region-1").Return("/invalid/entity/key")
		mockKeyGen.On("ValidateKey", "/invalid/entity/key").Return(NewKeyInvalidError("/invalid/entity/key", "invalid key format"))

		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		err := cache.StoreStateMachine(ctx, "1.0", models.DiagramTypePUML, "TestMachineValidation", sampleStateMachine, time.Hour)
		require.Error(t, err)
		cacheErr, ok := err.(*CacheError)
		require.True(t, ok, "Expected CacheError, got %T", err)
		assert.Equal(t, CacheErrorTypeKeyInvalid, cacheErr.Type)
		assert.Contains(t, err.Error(), "invalid entity key generated for 'region-1'")

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})
}
