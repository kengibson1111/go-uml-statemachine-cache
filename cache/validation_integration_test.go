package cache

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_EnhancedValidation_StoreDiagram(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		diagramType models.DiagramType
		diagramName string
		content     string
		ttl         time.Duration
		expectError bool
		errorType   CacheErrorType
	}{
		{
			name:        "empty diagram name",
			diagramType: models.DiagramTypePUML,
			diagramName: "",
			content:     "@startuml\nstate A\n@enduml",
			ttl:         time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "empty content",
			diagramType: models.DiagramTypePUML,
			diagramName: "test-diagram",
			content:     "",
			ttl:         time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "negative TTL",
			diagramType: models.DiagramTypePUML,
			diagramName: "test-diagram",
			content:     "@startuml\nstate A\n@enduml",
			ttl:         -time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "diagram name with special characters",
			diagramType: models.DiagramTypePUML,
			diagramName: "test/diagram with spaces & symbols",
			content:     "@startuml\nstate A\n@enduml",
			ttl:         time.Hour,
			expectError: false,
		},
		{
			name:        "content with security threat",
			diagramType: models.DiagramTypePUML,
			diagramName: "test-diagram",
			content:     "@startuml\n<script>alert('xss')</script>\n@enduml",
			ttl:         time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "diagram name too long",
			diagramType: models.DiagramTypePUML,
			diagramName: strings.Repeat("a", 101),
			content:     "@startuml\nstate A\n@enduml",
			ttl:         time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "content too large",
			diagramType: models.DiagramTypePUML,
			diagramName: "test-diagram",
			content:     strings.Repeat("a", 1024*1024+1),
			ttl:         time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "valid inputs",
			diagramType: models.DiagramTypePUML,
			diagramName: "test-diagram",
			content:     "@startuml\nstate A\nstate B\nA --> B\n@enduml",
			ttl:         time.Hour,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()
			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, DefaultRedisConfig())

			if !tt.expectError {
				// Setup mocks for successful case - need to account for input sanitization
				// The diagram name will be sanitized, so we need to predict the sanitized version
				sanitizedName := tt.diagramName
				if tt.diagramName == "test/diagram with spaces & symbols" {
					sanitizedName = "test-diagram_with_spaces_%26_symbols"
				}
				expectedKey := fmt.Sprintf("/diagrams/%s/%s", tt.diagramType.String(), sanitizedName)
				mockKeyGen.On("DiagramKey", sanitizedName).Return(expectedKey)
				mockKeyGen.On("ValidateKey", expectedKey).Return(nil)
				mockClient.On("SetWithRetry", ctx, expectedKey, tt.content, tt.ttl).Return(nil)
			}

			err := cache.StoreDiagram(ctx, tt.diagramType, tt.diagramName, tt.content, tt.ttl)

			if tt.expectError {
				require.Error(t, err)
				if cacheErr, ok := err.(*CacheError); ok {
					assert.Equal(t, tt.errorType, cacheErr.Type)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRedisCache_EnhancedValidation_StoreStateMachine(t *testing.T) {
	ctx := context.Background()

	validMachine := &models.StateMachine{
		ID:      "test-machine",
		Name:    "Test Machine",
		Version: "1.0",
		Regions: []*models.Region{
			{
				ID:   "region1",
				Name: "Main Region",
			},
		},
	}

	tests := []struct {
		name        string
		umlVersion  string
		diagramType models.DiagramType
		machineName string
		machine     *models.StateMachine
		ttl         time.Duration
		expectError bool
		errorType   CacheErrorType
	}{
		{
			name:        "empty UML version",
			umlVersion:  "",
			diagramType: models.DiagramTypePUML,
			machineName: "test-machine",
			machine:     validMachine,
			ttl:         time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "empty machine name",
			umlVersion:  "1.0",
			diagramType: models.DiagramTypePUML,
			machineName: "",
			machine:     validMachine,
			ttl:         time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "nil machine",
			umlVersion:  "1.0",
			diagramType: models.DiagramTypePUML,
			machineName: "test-machine",
			machine:     nil,
			ttl:         time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "invalid version format",
			umlVersion:  "1.0@beta",
			diagramType: models.DiagramTypePUML,
			machineName: "test-machine",
			machine:     validMachine,
			ttl:         time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "machine name with path traversal",
			umlVersion:  "1.0",
			diagramType: models.DiagramTypePUML,
			machineName: "../../../etc/passwd",
			machine:     validMachine,
			ttl:         time.Hour,
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "TTL too large",
			umlVersion:  "1.0",
			diagramType: models.DiagramTypePUML,
			machineName: "test-machine",
			machine:     validMachine,
			ttl:         400 * 24 * time.Hour, // More than 1 year
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()
			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, DefaultRedisConfig())

			// Setup mocks for diagram existence check for all cases that have valid version and name
			// Some validation errors happen after the diagram check
			if tt.umlVersion != "" && tt.machineName != "" {
				// Setup mocks for diagram existence check
				sanitizedName := tt.machineName
				if tt.machineName == "test-machine" {
					sanitizedName = "test-machine" // No sanitization needed for this name
				}
				diagramKey := fmt.Sprintf("/diagrams/%s/%s", tt.diagramType.String(), sanitizedName)
				mockKeyGen.On("DiagramKey", sanitizedName).Return(diagramKey)
				mockKeyGen.On("ValidateKey", diagramKey).Return(nil)
				mockClient.On("GetWithRetry", ctx, diagramKey).Return("@startuml\nstate A\n@enduml", nil)

				if !tt.expectError {
					// Setup mocks for state machine storage
					sanitizedVersion := tt.umlVersion
					machineKey := "/machines/" + sanitizedVersion + "/" + sanitizedName
					mockKeyGen.On("StateMachineKey", sanitizedVersion, sanitizedName).Return(machineKey)
					mockKeyGen.On("ValidateKey", machineKey).Return(nil)
					mockClient.On("SetWithRetry", ctx, machineKey, mock.Anything, tt.ttl).Return(nil)
				}
			}

			err := cache.StoreStateMachine(ctx, tt.umlVersion, tt.diagramType, tt.machineName, tt.machine, tt.ttl)

			if tt.expectError {
				require.Error(t, err)
				if cacheErr, ok := err.(*CacheError); ok {
					assert.Equal(t, tt.errorType, cacheErr.Type)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRedisCache_EnhancedValidation_CleanupWithOptions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		pattern     string
		options     *CleanupOptions
		expectError bool
		errorType   CacheErrorType
	}{
		{
			name:        "empty pattern",
			pattern:     "",
			options:     DefaultCleanupOptions(),
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "dangerous wildcard pattern",
			pattern:     "*",
			options:     DefaultCleanupOptions(),
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "pattern without valid prefix",
			pattern:     "/invalid/test*",
			options:     DefaultCleanupOptions(),
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "nil options",
			pattern:     fmt.Sprintf("/diagrams/%s/test*", models.DiagramTypePUML.String()),
			options:     nil,
			expectError: false, // Should use defaults
		},
		{
			name:    "invalid batch size",
			pattern: fmt.Sprintf("/diagrams/%s/test*", models.DiagramTypePUML.String()),
			options: &CleanupOptions{
				BatchSize:      -1,
				ScanCount:      100,
				MaxKeys:        0,
				DryRun:         false,
				Timeout:        5 * time.Minute,
				CollectMetrics: true,
			},
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:    "batch size too large",
			pattern: fmt.Sprintf("/diagrams/%s/test*", models.DiagramTypePUML.String()),
			options: &CleanupOptions{
				BatchSize:      1001,
				ScanCount:      100,
				MaxKeys:        0,
				DryRun:         false,
				Timeout:        5 * time.Minute,
				CollectMetrics: true,
			},
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:    "timeout too large",
			pattern: fmt.Sprintf("/diagrams/%s/test*", models.DiagramTypePUML.String()),
			options: &CleanupOptions{
				BatchSize:      100,
				ScanCount:      100,
				MaxKeys:        0,
				DryRun:         false,
				Timeout:        31 * time.Minute,
				CollectMetrics: true,
			},
			expectError: true,
			errorType:   CacheErrorTypeValidation,
		},
		{
			name:        "valid pattern and options",
			pattern:     fmt.Sprintf("/diagrams/%s/test*", models.DiagramTypePUML.String()),
			options:     DefaultCleanupOptions(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()
			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, DefaultRedisConfig())

			if !tt.expectError {
				// Setup mocks for successful cleanup
				mockClient.On("ScanWithRetry", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, uint64(0), nil)
			}

			_, err := cache.CleanupWithOptions(ctx, tt.pattern, tt.options)

			if tt.expectError {
				require.Error(t, err)
				if cacheErr, ok := err.(*CacheError); ok {
					assert.Equal(t, tt.errorType, cacheErr.Type)
				}
			} else {
				// Note: This might still fail due to mock setup, but validation should pass
				// The important thing is that validation errors are caught before Redis operations
				if err != nil && IsValidationError(err) {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestRedisCache_EnhancedValidation_ContextValidation(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		expectError bool
	}{
		{
			name:        "nil context",
			ctx:         nil,
			expectError: true,
		},
		{
			name:        "cancelled context",
			ctx:         func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			expectError: true,
		},
		{
			name:        "valid context",
			ctx:         context.Background(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()
			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, DefaultRedisConfig())

			if !tt.expectError {
				// Setup mocks for successful operation
				mockKeyGen.On("DiagramKey", "test").Return(fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String()))
				mockKeyGen.On("ValidateKey", fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String())).Return(nil)
				mockClient.On("GetWithRetry", tt.ctx, fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String())).Return("content", nil)
			}

			_, err := cache.GetDiagram(tt.ctx, models.DiagramTypePUML, "test")

			if tt.expectError {
				require.Error(t, err)
				assert.True(t, IsValidationError(err), "Expected validation error")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRedisCache_EnhancedValidation_SecurityThreats(t *testing.T) {
	ctx := context.Background()

	securityTests := []struct {
		name     string
		input    string
		field    string
		expected bool // true if should be blocked
	}{
		// SQL Injection in diagram names
		{"SQL injection in name", "'; DROP TABLE users; --", "name", true},
		{"SQL injection in content", "@startuml\n'; UNION SELECT * FROM users; --\n@enduml", "content", true},

		// XSS in various fields
		{"XSS in diagram name", "<script>alert('xss')</script>", "name", true},
		{"XSS in content", "@startuml\n<script>alert('xss')</script>\n@enduml", "content", true},

		// Path traversal
		{"Path traversal in name", "../../../etc/passwd", "name", true},

		// Safe content
		{"Safe diagram name", "user-authentication-flow", "name", false},
		{"Safe PUML content", "@startuml\nactor User\nUser -> System : authenticate\n@enduml", "content", false},
	}

	for _, tt := range securityTests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()
			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, DefaultRedisConfig())

			// Setup mocks for safe content tests
			if !tt.expected {
				mockKeyGen.On("DiagramKey", mock.Anything).Return(fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String()))
				mockKeyGen.On("ValidateKey", mock.Anything).Return(nil)
				mockClient.On("SetWithRetry", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			}

			var err error
			switch tt.field {
			case "name":
				err = cache.StoreDiagram(ctx, models.DiagramTypePUML, tt.input, "@startuml\nstate A\n@enduml", time.Hour)
			case "content":
				err = cache.StoreDiagram(ctx, models.DiagramTypePUML, "test-diagram", tt.input, time.Hour)
			}

			if tt.expected {
				require.Error(t, err, "Expected security threat to be detected: %s", tt.input)
				assert.True(t, IsValidationError(err), "Expected validation error for security threat")
			} else {
				// For safe content, we should not get validation errors
				if err != nil && IsValidationError(err) {
					t.Errorf("Safe content should not trigger validation error: %s", tt.input)
				}
			}
		})
	}
}

func TestRedisCache_EnhancedValidation_InputSanitization(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, DefaultRedisConfig())

	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name:           "spaces in name",
			input:          "test diagram with spaces",
			expectedOutput: "test_diagram_with_spaces",
		},
		{
			name:           "special characters in name",
			input:          "test/diagram-with_special.chars",
			expectedOutput: "test-diagram-with_special.chars",
		},
		{
			name:           "content with mixed line endings",
			input:          "@startuml\r\nstate A\rstate B\n@enduml",
			expectedOutput: "@startuml\nstate A\nstate B\n@enduml",
		},
		{
			name:           "content with null bytes",
			input:          "@startuml\x00\nstate A\n@enduml",
			expectedOutput: "@startuml\nstate A\n@enduml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test name sanitization
			if strings.Contains(tt.name, "name") {
				expectedKey := fmt.Sprintf("/diagrams/%s/%s", models.DiagramTypePUML.String(), tt.expectedOutput)
				mockKeyGen.On("DiagramKey", tt.expectedOutput).Return(expectedKey)
				mockKeyGen.On("ValidateKey", expectedKey).Return(nil)
				mockClient.On("SetWithRetry", ctx, expectedKey, "@startuml\nstate A\n@enduml", mock.Anything).Return(nil)

				err := cache.StoreDiagram(ctx, models.DiagramTypePUML, tt.input, "@startuml\nstate A\n@enduml", time.Hour)
				assert.NoError(t, err)
			}

			// Test content sanitization
			if strings.Contains(tt.name, "content") {
				expectedKey := fmt.Sprintf("/diagrams/%s/test", models.DiagramTypePUML.String())
				mockKeyGen.On("DiagramKey", "test").Return(expectedKey)
				mockKeyGen.On("ValidateKey", expectedKey).Return(nil)
				mockClient.On("SetWithRetry", ctx, expectedKey, tt.expectedOutput, mock.Anything).Return(nil)

				err := cache.StoreDiagram(ctx, models.DiagramTypePUML, "test", tt.input, time.Hour)
				assert.NoError(t, err)
			}
		})
	}
}
