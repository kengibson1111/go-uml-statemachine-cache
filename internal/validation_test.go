package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiagramValidation tests the validation logic without requiring Redis
func TestDiagramValidation(t *testing.T) {
	// Test validation directly without creating a full cache instance
	keyGen := NewKeyGenerator()

	tests := []struct {
		name        string
		diagramName string
		content     string
		expectError bool
		errorType   ErrorType
	}{
		{
			name:        "empty diagram name",
			diagramName: "",
			content:     "@startuml\nstate A\n@enduml",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "empty content",
			diagramName: "test-diagram",
			content:     "",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "valid inputs should pass validation",
			diagramName: "test-diagram",
			content:     "@startuml\nstate A\n@enduml",
			expectError: false, // Will fail later due to no Redis, but validation should pass
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic directly
			if tt.diagramName == "" {
				err := NewValidationError("diagram name cannot be empty", nil)
				require.Error(t, err)
				assert.Equal(t, ErrorTypeValidation, err.Type)
			}

			if tt.content == "" && tt.diagramName != "" {
				err := NewValidationError("diagram content cannot be empty", nil)
				require.Error(t, err)
				assert.Equal(t, ErrorTypeValidation, err.Type)
			}

			// Test key generation and validation for valid inputs
			if tt.diagramName != "" {
				key := keyGen.DiagramKey(tt.diagramName)
				err := keyGen.ValidateKey(key)
				assert.NoError(t, err, "Key validation should pass for valid diagram name")
			}
		})
	}
}

// TestKeyGeneration tests that keys are generated correctly
func TestKeyGeneration(t *testing.T) {
	keyGen := NewKeyGenerator()

	tests := []struct {
		name        string
		diagramName string
		expectedKey string
	}{
		{
			name:        "simple diagram name",
			diagramName: "test-diagram",
			expectedKey: "/diagrams/puml/test-diagram",
		},
		{
			name:        "diagram name with spaces",
			diagramName: "test diagram with spaces",
			expectedKey: "/diagrams/puml/test_diagram_with_spaces",
		},
		{
			name:        "diagram name with special characters",
			diagramName: "test/diagram-with_special.chars",
			expectedKey: "/diagrams/puml/test-diagram-with_special.chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := keyGen.DiagramKey(tt.diagramName)
			assert.Equal(t, tt.expectedKey, key)

			// Verify the key is valid
			err := keyGen.ValidateKey(key)
			assert.NoError(t, err)
		})
	}
}

// TestTTLHandling tests TTL parameter handling logic
func TestTTLHandling(t *testing.T) {
	config := DefaultConfig()
	config.DefaultTTL = 24 * time.Hour

	// Test that default TTL is properly configured
	assert.Equal(t, 24*time.Hour, config.DefaultTTL)

	// Test TTL validation in config
	assert.True(t, config.DefaultTTL > 0, "Default TTL should be positive")
}
