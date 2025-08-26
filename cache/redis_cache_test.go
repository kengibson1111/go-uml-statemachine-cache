package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_StoreDiagram(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		diagramType models.DiagramType
		diagramName string
		content     string
		ttl         time.Duration
		expectError bool
		errorType   CacheErrorType
		setupMocks  func(*MockRedisClient, *MockKeyGenerator)
	}{
		{
			name:        "valid diagram storage",
			diagramType: models.DiagramTypePUML,
			diagramName: "test-diagram",
			content:     "@startuml\nstate A\nstate B\nA --> B\n@enduml",
			ttl:         time.Hour,
			expectError: false,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				mockKeyGen.On("DiagramKey", "test-diagram").Return(fmt.Sprintf("/diagrams/%s/test-diagram", models.DiagramTypePUML.String()))
				mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/test-diagram", models.DiagramTypePUML.String())).Return(nil)
				mockClient.On("SetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/test-diagram", models.DiagramTypePUML.String()), "@startuml\nstate A\nstate B\nA --> B\n@enduml", time.Hour).Return(nil)
			},
		},
		{
			name:        "valid diagram with default TTL",
			diagramType: models.DiagramTypePUML,
			diagramName: "test-diagram-default-ttl",
			content:     "@startuml\nstate X\nstate Y\nX --> Y\n@enduml",
			ttl:         0, // Should use default TTL
			expectError: false,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				mockKeyGen.On("DiagramKey", "test-diagram-default-ttl").Return(fmt.Sprintf("/diagrams/%s/test-diagram-default-ttl", models.DiagramTypePUML.String()))
				mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/test-diagram-default-ttl", models.DiagramTypePUML.String())).Return(nil)
				mockClient.On("SetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/test-diagram-default-ttl", models.DiagramTypePUML.String()), "@startuml\nstate X\nstate Y\nX --> Y\n@enduml", time.Hour).Return(nil)
			},
		},
		{
			name:        "empty diagram name",
			diagramType: models.DiagramTypePUML,
			diagramName: "",
			content:     "@startuml\nstate A\n@enduml",
			ttl:         time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// No mocks needed for validation errors
			},
		},
		{
			name:        "empty content",
			diagramType: models.DiagramTypePUML,
			diagramName: "empty-content",
			content:     "",
			ttl:         time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// No mocks needed for validation errors
			},
		},
		{
			name:        "diagram with special characters",
			diagramType: models.DiagramTypePUML,
			diagramName: "test/diagram-with-special_chars.puml",
			content:     "@startuml\nstate \"State with spaces\"\n@enduml",
			ttl:         time.Hour,
			expectError: false,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				mockKeyGen.On("DiagramKey", "test-diagram-with-special_chars.puml").Return(fmt.Sprintf("/diagrams/%s/test-diagram-with-special_chars.puml", models.DiagramTypePUML.String()))
				mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/test-diagram-with-special_chars.puml", models.DiagramTypePUML.String())).Return(nil)
				mockClient.On("SetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/test-diagram-with-special_chars.puml", models.DiagramTypePUML.String()), "@startuml\nstate \"State with spaces\"\n@enduml", time.Hour).Return(nil)
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

			err := cache.StoreDiagram(ctx, tt.diagramType, tt.diagramName, tt.content, tt.ttl)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != 0 {
					cacheErr, ok := err.(*CacheError)
					require.True(t, ok, "expected CacheError, got %T", err)
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

func TestRedisCache_GetDiagram(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		diagramType models.DiagramType
		diagramName string
		expectError bool
		errorType   CacheErrorType
		expected    string
		setupMocks  func(*MockRedisClient, *MockKeyGenerator)
	}{
		{
			name:        "retrieve existing diagram",
			diagramType: models.DiagramTypePUML,
			diagramName: "test-get-diagram",
			expectError: false,
			expected:    "@startuml\nstate A\nstate B\nA --> B\n@enduml",
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				mockKeyGen.On("DiagramKey", "test-get-diagram").Return(fmt.Sprintf("/diagrams/%s/test-get-diagram", models.DiagramTypePUML.String()))
				mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/test-get-diagram", models.DiagramTypePUML.String())).Return(nil)
				mockClient.On("GetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/test-get-diagram", models.DiagramTypePUML.String())).Return("@startuml\nstate A\nstate B\nA --> B\n@enduml", nil)
			},
		},
		{
			name:        "retrieve non-existent diagram",
			diagramType: models.DiagramTypePUML,
			diagramName: "non-existent-diagram",
			expectError: true,
			errorType:   CacheErrorTypeNotFound,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				mockKeyGen.On("DiagramKey", "non-existent-diagram").Return(fmt.Sprintf("/diagrams/%s/non-existent-diagram", models.DiagramTypePUML.String()))
				mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/non-existent-diagram", models.DiagramTypePUML.String())).Return(nil)
				mockClient.On("GetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/non-existent-diagram", models.DiagramTypePUML.String())).Return("", redis.Nil)
			},
		},
		{
			name:        "empty diagram name",
			diagramType: models.DiagramTypePUML,
			diagramName: "",
			expectError: true,
			errorType:   CacheErrorTypeValidation,
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

			content, err := cache.GetDiagram(ctx, tt.diagramType, tt.diagramName)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != 0 {
					cacheErr, ok := err.(*CacheError)
					require.True(t, ok, "expected CacheError, got %T", err)
					assert.Equal(t, tt.errorType, cacheErr.Type)
				}
				assert.Empty(t, content)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, content)
			}

			mockClient.AssertExpectations(t)
			mockKeyGen.AssertExpectations(t)
		})
	}
}

func TestRedisCache_DeleteDiagram(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		diagramType models.DiagramType
		diagramName string
		expectError bool
		errorType   CacheErrorType
		setupMocks  func(*MockRedisClient, *MockKeyGenerator)
	}{
		{
			name:        "delete existing diagram",
			diagramType: models.DiagramTypePUML,
			diagramName: "test-delete-diagram",
			expectError: false,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				mockKeyGen.On("DiagramKey", "test-delete-diagram").Return(fmt.Sprintf("/diagrams/%s/test-delete-diagram", models.DiagramTypePUML.String()))
				mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/test-delete-diagram", models.DiagramTypePUML.String())).Return(nil)
				mockClient.On("DelWithRetry", ctx, []string{fmt.Sprintf("/diagrams/%s/test-delete-diagram", models.DiagramTypePUML.String())}).Return(nil)
			},
		},
		{
			name:        "delete non-existent diagram (should not error)",
			diagramType: models.DiagramTypePUML,
			diagramName: "non-existent-diagram",
			expectError: false,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				mockKeyGen.On("DiagramKey", "non-existent-diagram").Return(fmt.Sprintf("/diagrams/%s/non-existent-diagram", models.DiagramTypePUML.String()))
				mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/non-existent-diagram", models.DiagramTypePUML.String())).Return(nil)
				mockClient.On("DelWithRetry", ctx, []string{fmt.Sprintf("/diagrams/%s/non-existent-diagram", models.DiagramTypePUML.String())}).Return(nil)
			},
		},
		{
			name:        "empty diagram name",
			diagramType: models.DiagramTypePUML,
			diagramName: "",
			expectError: true,
			errorType:   CacheErrorTypeValidation,
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

			err := cache.DeleteDiagram(ctx, tt.diagramType, tt.diagramName)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != 0 {
					cacheErr, ok := err.(*CacheError)
					require.True(t, ok, "expected CacheError, got %T", err)
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

func TestRedisCache_KeyValidationError(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := &RedisConfig{DefaultTTL: time.Hour}

	// Setup mock to return key validation error
	mockKeyGen.On("DiagramKey", "test-diagram").Return("/invalid/key")
	mockKeyGen.On("ValidateKey", "/invalid/key").Return(NewKeyInvalidError("/invalid/key", "invalid key format"))

	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	err := cache.StoreDiagram(ctx, models.DiagramTypePUML, "test-diagram", "@startuml\nstate A\n@enduml", time.Hour)

	require.Error(t, err)
	cacheErr, ok := err.(*CacheError)
	require.True(t, ok, "expected CacheError, got %T", err)
	assert.Equal(t, CacheErrorTypeKeyInvalid, cacheErr.Type)

	mockKeyGen.AssertExpectations(t)
}

func TestRedisCache_RedisConnectionError(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := &RedisConfig{DefaultTTL: time.Hour}

	// Setup mocks for successful key generation but Redis connection error
	mockKeyGen.On("DiagramKey", "test-diagram").Return(fmt.Sprintf("/diagrams/%s/test-diagram", models.DiagramTypePUML.String()))
	mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/test-diagram", models.DiagramTypePUML.String())).Return(nil)
	mockClient.On("SetWithRetry", ctx, fmt.Sprintf("/diagrams/%s/test-diagram", models.DiagramTypePUML.String()), "@startuml\nstate A\n@enduml", time.Hour).Return(assert.AnError)

	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	err := cache.StoreDiagram(ctx, models.DiagramTypePUML, "test-diagram", "@startuml\nstate A\n@enduml", time.Hour)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store diagram")

	mockClient.AssertExpectations(t)
	mockKeyGen.AssertExpectations(t)
}

func TestRedisCache_Health(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := &RedisConfig{DefaultTTL: time.Hour}

	mockClient.On("HealthWithRetry", ctx).Return(nil)

	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	err := cache.Health(ctx)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestRedisCache_Close(t *testing.T) {
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := &RedisConfig{DefaultTTL: time.Hour}

	mockClient.On("Close").Return(nil)

	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	err := cache.Close()

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}
