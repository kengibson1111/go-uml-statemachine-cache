package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/internal"
	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestRedisCache_NewRedisCache tests the constructor function
func TestRedisCache_NewRedisCache(t *testing.T) {
	tests := []struct {
		name        string
		config      *RedisConfig
		expectError bool
		description string
	}{
		{
			name:        "valid config",
			config:      DefaultRedisConfig(),
			expectError: false,
			description: "should create cache with valid config",
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: false,
			description: "should use default config when nil",
		},
		{
			name: "invalid config - empty address",
			config: &RedisConfig{
				RedisAddr:    "",
				RedisDB:      0,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10,
				DefaultTTL:   24 * time.Hour,
			},
			expectError: true,
			description: "should fail with empty Redis address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewRedisCache(tt.config)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, cache)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, cache)
				if cache != nil {
					assert.NoError(t, cache.Close())
				}
			}
		})
	}
}

// TestRedisCache_NewRedisCacheWithDependencies tests the dependency injection constructor
func TestRedisCache_NewRedisCacheWithDependencies(t *testing.T) {
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := DefaultRedisConfig()

	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	assert.NotNil(t, cache)
	assert.Equal(t, mockClient, cache.client)
	assert.Equal(t, mockKeyGen, cache.keyGen)
	assert.Equal(t, config, cache.config)
}

// TestRedisCache_PublicTypeAliases tests that public type aliases work correctly
func TestRedisCache_PublicTypeAliases(t *testing.T) {
	t.Run("RedisConfig alias", func(t *testing.T) {
		config := DefaultRedisConfig()
		assert.NotNil(t, config)
		assert.Equal(t, "localhost:6379", config.RedisAddr)
	})

	t.Run("RedisRetryConfig alias", func(t *testing.T) {
		retryConfig := DefaultRedisRetryConfig()
		assert.NotNil(t, retryConfig)
		assert.Equal(t, 3, retryConfig.MaxAttempts)
	})

	t.Run("CacheError alias", func(t *testing.T) {
		err := internal.NewCacheError(ErrorTypeConnection, "test-key", "test message", nil)
		assert.NotNil(t, err)
		assert.Equal(t, ErrorTypeConnection, err.Type)
	})
}

// TestRedisCache_ErrorHelperFunctions tests the public error helper functions
func TestRedisCache_ErrorHelperFunctions(t *testing.T) {
	t.Run("IsConnectionError", func(t *testing.T) {
		connErr := internal.NewConnectionError("connection failed", nil)
		otherErr := internal.NewNotFoundError("key")
		regularErr := errors.New("regular error")

		assert.True(t, IsConnectionError(connErr))
		assert.False(t, IsConnectionError(otherErr))
		assert.False(t, IsConnectionError(regularErr))
	})

	t.Run("IsNotFoundError", func(t *testing.T) {
		notFoundErr := internal.NewNotFoundError("key")
		otherErr := internal.NewConnectionError("connection failed", nil)
		regularErr := errors.New("regular error")

		assert.True(t, IsNotFoundError(notFoundErr))
		assert.False(t, IsNotFoundError(otherErr))
		assert.False(t, IsNotFoundError(regularErr))
	})

	t.Run("IsValidationError", func(t *testing.T) {
		validationErr := internal.NewValidationError("validation failed", nil)
		otherErr := internal.NewConnectionError("connection failed", nil)
		regularErr := errors.New("regular error")

		assert.True(t, IsValidationError(validationErr))
		assert.False(t, IsValidationError(otherErr))
		assert.False(t, IsValidationError(regularErr))
	})

	t.Run("IsRetryExhaustedError", func(t *testing.T) {
		retryErr := internal.NewRetryExhaustedError("test-op", 3, nil)
		otherErr := internal.NewConnectionError("connection failed", nil)

		assert.True(t, IsRetryExhaustedError(retryErr))
		assert.False(t, IsRetryExhaustedError(otherErr))
	})

	t.Run("IsCircuitOpenError", func(t *testing.T) {
		circuitErr := internal.NewCircuitOpenError("test-op")
		otherErr := internal.NewConnectionError("connection failed", nil)

		assert.True(t, IsCircuitOpenError(circuitErr))
		assert.False(t, IsCircuitOpenError(otherErr))
	})

	t.Run("IsRetryableError", func(t *testing.T) {
		retryableErr := internal.NewConnectionError("connection failed", nil)
		nonRetryableErr := internal.NewValidationError("validation failed", nil)

		assert.True(t, IsRetryableError(retryableErr))
		assert.False(t, IsRetryableError(nonRetryableErr))
	})

	t.Run("GetErrorSeverity", func(t *testing.T) {
		criticalErr := internal.NewConnectionError("connection failed", nil)
		lowErr := internal.NewNotFoundError("key")

		assert.Equal(t, SeverityCritical, GetErrorSeverity(criticalErr))
		assert.Equal(t, SeverityLow, GetErrorSeverity(lowErr))
	})

	t.Run("GetErrorSeverity", func(t *testing.T) {
		connErr := internal.NewConnectionError("connection failed", nil)
		validationErr := internal.NewValidationError("validation failed", nil)

		assert.Equal(t, SeverityCritical, GetErrorSeverity(connErr))
		assert.Equal(t, SeverityMedium, GetErrorSeverity(validationErr))
	})
}

// TestRedisCache_ErrorConstructors tests all error constructor functions
func TestRedisCache_ErrorConstructors(t *testing.T) {
	t.Run("NewCacheError", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := internal.NewCacheError(ErrorTypeConnection, "test-key", "test message", cause)

		assert.Equal(t, ErrorTypeConnection, err.Type)
		assert.Equal(t, "test-key", err.Key)
		assert.Equal(t, "test message", err.Message)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("NewConnectionError", func(t *testing.T) {
		cause := errors.New("connection failed")
		err := internal.NewConnectionError("Redis connection failed", cause)

		assert.Equal(t, ErrorTypeConnection, err.Type)
		assert.Equal(t, "", err.Key)
		assert.Equal(t, "Redis connection failed", err.Message)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("NewKeyInvalidError", func(t *testing.T) {
		err := internal.NewKeyInvalidError("invalid-key", "key contains invalid characters")

		assert.Equal(t, ErrorTypeKeyInvalid, err.Type)
		assert.Equal(t, "invalid-key", err.Key)
		assert.Equal(t, "key contains invalid characters", err.Message)
		assert.Nil(t, err.Cause)
	})

	t.Run("NewNotFoundError", func(t *testing.T) {
		err := internal.NewNotFoundError("missing-key")

		assert.Equal(t, ErrorTypeNotFound, err.Type)
		assert.Equal(t, "missing-key", err.Key)
		assert.Equal(t, "key not found in cache", err.Message)
		assert.Nil(t, err.Cause)
	})

	t.Run("NewSerializationError", func(t *testing.T) {
		cause := errors.New("json marshal failed")
		err := internal.NewSerializationError("test-key", "failed to serialize", cause)

		assert.Equal(t, ErrorTypeSerialization, err.Type)
		assert.Equal(t, "test-key", err.Key)
		assert.Equal(t, "failed to serialize", err.Message)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("NewTimeoutError", func(t *testing.T) {
		cause := errors.New("context deadline exceeded")
		err := internal.NewTimeoutError("test-key", "operation timed out", cause)

		assert.Equal(t, ErrorTypeTimeout, err.Type)
		assert.Equal(t, "test-key", err.Key)
		assert.Equal(t, "operation timed out", err.Message)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("NewValidationError", func(t *testing.T) {
		cause := errors.New("invalid input")
		err := internal.NewValidationError("validation failed", cause)

		assert.Equal(t, ErrorTypeValidation, err.Type)
		assert.Equal(t, "", err.Key)
		assert.Equal(t, "validation failed", err.Message)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("NewRetryExhaustedError", func(t *testing.T) {
		lastError := errors.New("connection failed")
		err := internal.NewRetryExhaustedError("get-operation", 5, lastError)

		assert.Equal(t, ErrorTypeRetryExhausted, err.Type)
		assert.Equal(t, "get-operation", err.Context.Operation)
		assert.Equal(t, 5, err.Context.AttemptNumber)
		assert.Equal(t, lastError, err.Cause)
	})

	t.Run("NewCircuitOpenError", func(t *testing.T) {
		err := internal.NewCircuitOpenError("set-operation")

		assert.Equal(t, ErrorTypeCircuitOpen, err.Type)
		assert.Equal(t, "set-operation", err.Context.Operation)
	})

	t.Run("NewCacheErrorWithContext", func(t *testing.T) {
		cause := errors.New("underlying error")
		context := &internal.ErrorContext{
			Operation:     "test-op",
			AttemptNumber: 2,
			Duration:      100 * time.Millisecond,
		}

		err := internal.NewCacheErrorWithContext(ErrorTypeTimeout, "test-key", "test message", cause, context)

		assert.Equal(t, ErrorTypeTimeout, err.Type)
		assert.Equal(t, "test-key", err.Key)
		assert.Equal(t, "test message", err.Message)
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, "test-op", err.Context.Operation)
		assert.Equal(t, 2, err.Context.AttemptNumber)
	})
}

// TestRedisCache_CircuitBreakerIntegration tests circuit breaker functionality
func TestRedisCache_CircuitBreakerIntegration(t *testing.T) {
	t.Run("DefaultCircuitBreakerConfig", func(t *testing.T) {
		config := internal.DefaultCircuitBreakerConfig()

		assert.NotNil(t, config)
		assert.Equal(t, 5, config.FailureThreshold)
		assert.Equal(t, 30*time.Second, config.RecoveryTimeout)
		assert.Equal(t, 3, config.SuccessThreshold)
		assert.Equal(t, 3, config.MaxHalfOpenCalls)
	})

	t.Run("NewErrorRecoveryManager", func(t *testing.T) {
		config := internal.DefaultCircuitBreakerConfig()
		erm := internal.NewErrorRecoveryManager(config)

		assert.NotNil(t, erm)

		// Test circuit breaker creation
		cb := erm.GetCircuitBreaker("test-operation")
		assert.NotNil(t, cb)

		// Test that same operation returns same circuit breaker
		cb2 := erm.GetCircuitBreaker("test-operation")
		assert.Equal(t, cb, cb2)
	})
}

// TestRedisCache_SerializationHelpers tests serialization edge cases
func TestRedisCache_SerializationHelpers(t *testing.T) {
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := DefaultRedisConfig()
	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	t.Run("extractEntitiesFromStateMachine with complex structure", func(t *testing.T) {
		// Create a complex state machine with nested regions
		machine := &models.StateMachine{
			ID:   "test-machine",
			Name: "Test Machine",
			Regions: []*models.Region{
				{
					ID:   "region1",
					Name: "Region 1",
					States: []*models.State{
						{
							Vertex: models.Vertex{ID: "state1", Name: "State 1"},
							Regions: []*models.Region{
								{
									ID:   "nested-region",
									Name: "Nested Region",
									States: []*models.State{
										{Vertex: models.Vertex{ID: "nested-state", Name: "Nested State"}},
									},
									Transitions: []*models.Transition{
										{ID: "nested-transition", Name: "Nested Transition"},
									},
									Vertices: []*models.Vertex{
										{ID: "nested-vertex", Name: "Nested Vertex"},
									},
								},
							},
						},
					},
					Transitions: []*models.Transition{
						{ID: "transition1", Name: "Transition 1"},
					},
					Vertices: []*models.Vertex{
						{ID: "vertex1", Name: "Vertex 1"},
					},
				},
			},
		}

		entities := cache.extractEntitiesFromStateMachine(machine)

		// Verify all entities are extracted
		assert.Contains(t, entities, "region1")
		assert.Contains(t, entities, "state1")
		assert.Contains(t, entities, "transition1")
		assert.Contains(t, entities, "vertex1")
		assert.Contains(t, entities, "nested-region")
		assert.Contains(t, entities, "nested-state")
		assert.Contains(t, entities, "nested-transition")
		assert.Contains(t, entities, "nested-vertex")

		// Verify entity types
		assert.IsType(t, &models.Region{}, entities["region1"])
		assert.IsType(t, &models.State{}, entities["state1"])
		assert.IsType(t, &models.Transition{}, entities["transition1"])
		assert.IsType(t, &models.Vertex{}, entities["vertex1"])
	})

	t.Run("extractEntitiesFromStateMachine with empty machine", func(t *testing.T) {
		machine := &models.StateMachine{
			ID:      "empty-machine",
			Name:    "Empty Machine",
			Regions: []*models.Region{},
		}

		entities := cache.extractEntitiesFromStateMachine(machine)
		assert.Empty(t, entities)
	})

	t.Run("extractEntitiesFromStateMachine with nil regions", func(t *testing.T) {
		machine := &models.StateMachine{
			ID:      "nil-regions-machine",
			Name:    "Nil Regions Machine",
			Regions: nil,
		}

		entities := cache.extractEntitiesFromStateMachine(machine)
		assert.Empty(t, entities)
	})
}

// TestRedisCache_TypeSafeEntityRetrieval tests type-safe entity retrieval methods
func TestRedisCache_TypeSafeEntityRetrieval(t *testing.T) {
	ctx := context.Background()
	umlVersion := "2.5"
	diagramName := "test-diagram"
	entityID := "test-entity"

	t.Run("GetEntityAsState success", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := DefaultRedisConfig()
		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		state := &models.State{
			Vertex:      models.Vertex{ID: "state1", Name: "State 1"},
			IsComposite: true,
		}
		stateJSON, _ := json.Marshal(state)

		mockKeyGen.On("EntityKey", umlVersion, diagramName, entityID).Return("/test/state/key")
		mockKeyGen.On("ValidateKey", "/test/state/key").Return(nil)
		mockClient.On("GetWithRetry", ctx, "/test/state/key").Return(string(stateJSON), nil)

		result, err := cache.GetEntityAsState(ctx, umlVersion, diagramName, entityID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "state1", result.ID)
		assert.Equal(t, "State 1", result.Name)
		assert.True(t, result.IsComposite)

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})

	t.Run("GetEntityAsTransition success", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := DefaultRedisConfig()
		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		transition := &models.Transition{
			ID:   "transition1",
			Name: "Transition 1",
			Kind: models.TransitionKindExternal,
		}
		transitionJSON, _ := json.Marshal(transition)

		mockKeyGen.On("EntityKey", umlVersion, diagramName, entityID).Return("/test/transition/key")
		mockKeyGen.On("ValidateKey", "/test/transition/key").Return(nil)
		mockClient.On("GetWithRetry", ctx, "/test/transition/key").Return(string(transitionJSON), nil)

		result, err := cache.GetEntityAsTransition(ctx, umlVersion, diagramName, entityID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "transition1", result.ID)
		assert.Equal(t, "Transition 1", result.Name)
		assert.Equal(t, models.TransitionKindExternal, result.Kind)

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})

	t.Run("GetEntityAsRegion success", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := DefaultRedisConfig()
		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		region := &models.Region{
			ID:   "region1",
			Name: "Region 1",
		}
		regionJSON, _ := json.Marshal(region)

		mockKeyGen.On("EntityKey", umlVersion, diagramName, entityID).Return("/test/region/key")
		mockKeyGen.On("ValidateKey", "/test/region/key").Return(nil)
		mockClient.On("GetWithRetry", ctx, "/test/region/key").Return(string(regionJSON), nil)

		result, err := cache.GetEntityAsRegion(ctx, umlVersion, diagramName, entityID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "region1", result.ID)
		assert.Equal(t, "Region 1", result.Name)

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})

	t.Run("GetEntityAsVertex success", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := DefaultRedisConfig()
		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		vertex := &models.Vertex{
			ID:   "vertex1",
			Name: "Vertex 1",
			Type: "state",
		}
		vertexJSON, _ := json.Marshal(vertex)

		mockKeyGen.On("EntityKey", umlVersion, diagramName, entityID).Return("/test/vertex/key")
		mockKeyGen.On("ValidateKey", "/test/vertex/key").Return(nil)
		mockClient.On("GetWithRetry", ctx, "/test/vertex/key").Return(string(vertexJSON), nil)

		result, err := cache.GetEntityAsVertex(ctx, umlVersion, diagramName, entityID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "vertex1", result.ID)
		assert.Equal(t, "Vertex 1", result.Name)
		assert.Equal(t, "state", result.Type)

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})

	t.Run("GetEntityAsState with serialization error", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := DefaultRedisConfig()
		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		mockKeyGen.On("EntityKey", umlVersion, diagramName, entityID).Return("/test/error/key")
		mockKeyGen.On("ValidateKey", "/test/error/key").Return(nil)
		mockClient.On("GetWithRetry", ctx, "/test/error/key").Return("invalid json", nil)

		result, err := cache.GetEntityAsState(ctx, umlVersion, diagramName, entityID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, IsSerializationError(err))

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})
}

// TestRedisCache_EdgeCases tests various edge cases and error conditions
func TestRedisCache_EdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("StoreDiagram with Redis Nil error", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := DefaultRedisConfig()
		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		mockKeyGen.On("DiagramKey", "test").Return(fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String()))
		mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String())).Return(nil)
		mockClient.On("SetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String()), "content", config.DefaultTTL).Return(redis.Nil)

		err := cache.StoreDiagram(ctx, models.DiagramTypePUML, "test", "content", 0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to store diagram")

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})

	t.Run("GetDiagram with connection error", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := DefaultRedisConfig()
		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		connErr := errors.New("connection refused")
		mockKeyGen.On("DiagramKey", "test").Return(fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String()))
		mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String())).Return(nil)
		mockClient.On("GetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String())).Return("", connErr)

		result, err := cache.GetDiagram(ctx, models.DiagramTypePUML, "test")

		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "failed to retrieve diagram")

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})

	t.Run("StoreEntity with serialization error", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := DefaultRedisConfig()
		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		// Create an entity that will cause JSON marshaling to fail
		// Using a channel which cannot be marshaled to JSON
		entity := make(chan int)

		mockKeyGen.On("EntityKey", "2.5", "test", "entity1").Return("/test/entity/key")
		mockKeyGen.On("ValidateKey", "/test/entity/key").Return(nil)

		err := cache.StoreEntity(ctx, "2.5", "test", "entity1", entity, time.Hour)

		assert.Error(t, err)
		assert.True(t, IsSerializationError(err))

		mockKeyGen.AssertExpectations(t)
	})

	t.Run("StoreEntity with key validation error", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := DefaultRedisConfig()
		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		keyErr := errors.New("invalid key format")
		mockKeyGen.On("EntityKey", "2.5", "test", "entity1").Return("/invalid/key")
		mockKeyGen.On("ValidateKey", "/invalid/key").Return(keyErr)

		err := cache.StoreEntity(ctx, "2.5", "test", "entity1", map[string]string{"id": "entity1"}, time.Hour)

		assert.Error(t, err)
		assert.True(t, IsKeyInvalidError(err))

		mockKeyGen.AssertExpectations(t)
	})
}

// TestRedisCache_ConcurrencyAndThreadSafety tests thread safety aspects
func TestRedisCache_ConcurrencyAndThreadSafety(t *testing.T) {
	t.Run("concurrent diagram operations", func(t *testing.T) {
		mockClient := NewMockRedisClient()
		mockKeyGen := NewMockKeyGenerator()
		config := DefaultRedisConfig()
		cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

		ctx := context.Background()

		// Set up mocks for concurrent operations
		for i := 0; i < 10; i++ {
			mockKeyGen.On("DiagramKey", mock.AnythingOfType("string")).Return(fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String()))
			mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String())).Return(nil)
			mockClient.On("SetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String()), mock.AnythingOfType("string"), config.DefaultTTL).Return(nil)
		}

		// Run concurrent operations
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				err := cache.StoreDiagram(ctx, models.DiagramTypePUML, "test", "content", 0)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all operations to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		mockClient.AssertExpectations(t)
		mockKeyGen.AssertExpectations(t)
	})
}

// Helper function to check if error is a serialization error
func IsSerializationError(err error) bool {
	if cacheErr, ok := err.(*Error); ok {
		return cacheErr.Type == ErrorTypeSerialization
	}
	return false
}

// Helper function to check if error is a key invalid error
func IsKeyInvalidError(err error) bool {
	if cacheErr, ok := err.(*Error); ok {
		return cacheErr.Type == ErrorTypeKeyInvalid
	}
	return false
}
