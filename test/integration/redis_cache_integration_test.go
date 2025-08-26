package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/cache"
	"github.com/kengibson1111/go-uml-statemachine-cache/internal"
	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_StoreDiagram_Integration(t *testing.T) {
	// Skip if Redis is not available
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name        string
		diagramType models.DiagramType
		diagramName string
		content     string
		ttl         time.Duration
		expectError bool
		errorType   internal.ErrorType
	}{
		{
			name:        "valid diagram storage",
			diagramType: models.DiagramTypePUML,
			diagramName: "test-diagram",
			content:     "@startuml\nstate A\nstate B\nA --> B\n@enduml",
			ttl:         time.Hour,
			expectError: false,
		},
		{
			name:        "valid diagram with default TTL",
			diagramType: models.DiagramTypePUML,
			diagramName: "test-diagram-default-ttl",
			content:     "@startuml\nstate X\nstate Y\nX --> Y\n@enduml",
			ttl:         0, // Should use default TTL
			expectError: false,
		},
		{
			name:        "diagram with special characters",
			diagramType: models.DiagramTypePUML,
			diagramName: "test/diagram-with-special_chars.puml",
			content:     "@startuml\nstate \"State with spaces\"\n@enduml",
			ttl:         time.Hour,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.StoreDiagram(ctx, tt.diagramType, tt.diagramName, tt.content, tt.ttl)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != 0 {
					cacheErr, ok := err.(*internal.CacheError)
					require.True(t, ok, "expected CacheError, got %T", err)
					assert.Equal(t, tt.errorType, cacheErr.Type)
				}
			} else {
				require.NoError(t, err)

				// Verify the diagram was stored by retrieving it
				retrieved, err := cache.GetDiagram(ctx, tt.diagramType, tt.diagramName)
				require.NoError(t, err)
				assert.Equal(t, tt.content, retrieved)
			}
		})
	}
}

func TestRedisCache_GetDiagram_Integration(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	// Store a test diagram first
	testName := "test-get-diagram"
	testContent := "@startuml\nstate A\nstate B\nA --> B\n@enduml"
	err := cache.StoreDiagram(ctx, models.DiagramTypePUML, testName, testContent, time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name        string
		diagramType models.DiagramType
		diagramName string
		expectError bool
		errorType   internal.ErrorType
		expected    string
	}{
		{
			name:        "retrieve existing diagram",
			diagramType: models.DiagramTypePUML,
			diagramName: testName,
			expectError: false,
			expected:    testContent,
		},
		{
			name:        "retrieve non-existent diagram",
			diagramType: models.DiagramTypePUML,
			diagramName: "non-existent-diagram",
			expectError: true,
			errorType:   internal.ErrorType(internal.ErrorTypeNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := cache.GetDiagram(ctx, tt.diagramType, tt.diagramName)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != 0 {
					cacheErr, ok := err.(*internal.CacheError)
					require.True(t, ok, "expected CacheError, got %T", err)
					assert.Equal(t, tt.errorType, cacheErr.Type)
				}
				assert.Empty(t, content)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, content)
			}
		})
	}
}

func TestRedisCache_DeleteDiagram_Integration(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	// Store a test diagram first
	testName := "test-delete-diagram"
	testContent := "@startuml\nstate A\n@enduml"
	err := cache.StoreDiagram(ctx, models.DiagramTypePUML, testName, testContent, time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name        string
		diagramType models.DiagramType
		diagramName string
		expectError bool
		errorType   internal.ErrorType
	}{
		{
			name:        "delete existing diagram",
			diagramType: models.DiagramTypePUML,
			diagramName: testName,
			expectError: false,
		},
		{
			name:        "delete non-existent diagram (should not error)",
			diagramType: models.DiagramTypePUML,
			diagramName: "non-existent-diagram",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.DeleteDiagram(ctx, tt.diagramType, tt.diagramName)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != 0 {
					cacheErr, ok := err.(*internal.CacheError)
					require.True(t, ok, "expected CacheError, got %T", err)
					assert.Equal(t, tt.errorType, cacheErr.Type)
				}
			} else {
				require.NoError(t, err)

				// If we deleted an existing diagram, verify it's gone
				if tt.diagramName == testName {
					_, err := cache.GetDiagram(ctx, tt.diagramType, tt.diagramName)
					require.Error(t, err)
					assert.True(t, internal.IsNotFoundError(err))
				}
			}
		})
	}
}

func TestRedisCache_DiagramTTL_Integration(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	// Store a diagram with a very short TTL
	testName := "test-ttl-diagram"
	testContent := "@startuml\nstate A\n@enduml"
	shortTTL := 100 * time.Millisecond

	err := cache.StoreDiagram(ctx, models.DiagramTypePUML, testName, testContent, shortTTL)
	require.NoError(t, err)

	// Verify it exists immediately
	content, err := cache.GetDiagram(ctx, models.DiagramTypePUML, testName)
	require.NoError(t, err)
	assert.Equal(t, testContent, content)

	// Wait for TTL to expire
	time.Sleep(shortTTL + 50*time.Millisecond)

	// Verify it's gone
	_, err = cache.GetDiagram(ctx, models.DiagramTypePUML, testName)
	require.Error(t, err)
	assert.True(t, internal.IsNotFoundError(err))
}

func TestRedisCache_DiagramOperationsIntegration(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	// Test complete workflow: store -> get -> delete -> verify gone
	diagramName := "integration-test-diagram"
	content := "@startuml\nstate Start\nstate End\nStart --> End\n@enduml"

	// Store
	err := cache.StoreDiagram(ctx, models.DiagramTypePUML, diagramName, content, time.Hour)
	require.NoError(t, err)

	// Get
	retrieved, err := cache.GetDiagram(ctx, models.DiagramTypePUML, diagramName)
	require.NoError(t, err)
	assert.Equal(t, content, retrieved)

	// Delete
	err = cache.DeleteDiagram(ctx, models.DiagramTypePUML, diagramName)
	require.NoError(t, err)

	// Verify gone
	_, err = cache.GetDiagram(ctx, models.DiagramTypePUML, diagramName)
	require.Error(t, err)
	assert.True(t, internal.IsNotFoundError(err))
}

// setupTestCache creates a test cache instance
// Returns the cache and a cleanup function
func setupTestCache(t *testing.T) (*cache.RedisCache, func()) {
	config := internal.DefaultConfig()
	config.RedisDB = 15 // Use a different DB for tests
	config.DefaultTTL = time.Hour

	cache, err := cache.NewRedisCache(config)
	require.NoError(t, err)

	// Test Redis connection
	ctx := context.Background()
	err = cache.Health(ctx)
	if err != nil {
		t.Skip("Redis not available for testing:", err)
	}

	// Cleanup function
	cleanup := func() {
		// Clean up any test data
		_ = cache.Cleanup(context.Background(), fmt.Sprintf("/diagrams/%s/*", models.DiagramTypePUML.String()))
		_ = cache.Close()
	}

	return cache, cleanup
}
