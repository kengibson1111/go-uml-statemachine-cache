package internal

import (
	"strings"
	"testing"
)

func TestNewKeyGenerator(t *testing.T) {
	kg := NewKeyGenerator()
	if kg == nil {
		t.Fatal("NewKeyGenerator() returned nil")
	}

	// Verify it implements the interface
	var _ KeyGenerator = kg
}

func TestDiagramKey(t *testing.T) {
	kg := NewKeyGenerator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "simple",
			expected: "/diagrams/puml/simple",
		},
		{
			name:     "name with spaces",
			input:    "my diagram",
			expected: "/diagrams/puml/my_diagram",
		},
		{
			name:     "name with special characters",
			input:    "diagram/with\\special:chars",
			expected: "/diagrams/puml/diagram-with-special%3Achars",
		},
		{
			name:     "name with dots",
			input:    "diagram.v1.0",
			expected: "/diagrams/puml/diagram.v1.0",
		},
		{
			name:     "empty name",
			input:    "",
			expected: "/diagrams/puml/",
		},
		{
			name:     "name with unicode",
			input:    "диаграмма",
			expected: "/diagrams/puml/%D0%B4%D0%B8%D0%B0%D0%B3%D1%80%D0%B0%D0%BC%D0%BC%D0%B0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kg.DiagramKey(tt.input)
			if result != tt.expected {
				t.Errorf("DiagramKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStateMachineKey(t *testing.T) {
	kg := NewKeyGenerator()

	tests := []struct {
		name        string
		umlVersion  string
		diagramName string
		expected    string
	}{
		{
			name:        "simple version and name",
			umlVersion:  "2.5",
			diagramName: "simple",
			expected:    "/machines/2.5/simple",
		},
		{
			name:        "version and name with spaces",
			umlVersion:  "version 2.5",
			diagramName: "my diagram",
			expected:    "/machines/version_2.5/my_diagram",
		},
		{
			name:        "version and name with special characters",
			umlVersion:  "v2.5/beta",
			diagramName: "diagram\\with:special",
			expected:    "/machines/v2.5-beta/diagram-with%3Aspecial",
		},
		{
			name:        "empty values",
			umlVersion:  "",
			diagramName: "",
			expected:    "/machines//",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kg.StateMachineKey(tt.umlVersion, tt.diagramName)
			if result != tt.expected {
				t.Errorf("StateMachineKey(%q, %q) = %q, want %q", tt.umlVersion, tt.diagramName, result, tt.expected)
			}
		})
	}
}

func TestEntityKey(t *testing.T) {
	kg := NewKeyGenerator()

	tests := []struct {
		name        string
		umlVersion  string
		diagramName string
		entityID    string
		expected    string
	}{
		{
			name:        "simple values",
			umlVersion:  "2.5",
			diagramName: "simple",
			entityID:    "state1",
			expected:    "/machines/2.5/simple/entities/state1",
		},
		{
			name:        "values with spaces",
			umlVersion:  "version 2.5",
			diagramName: "my diagram",
			entityID:    "my state",
			expected:    "/machines/version_2.5/my_diagram/entities/my_state",
		},
		{
			name:        "values with special characters",
			umlVersion:  "v2.5/beta",
			diagramName: "diagram\\with:special",
			entityID:    "entity/id:1",
			expected:    "/machines/v2.5-beta/diagram-with%3Aspecial/entities/entity-id%3A1",
		},
		{
			name:        "empty values",
			umlVersion:  "",
			diagramName: "",
			entityID:    "",
			expected:    "/machines///entities/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kg.EntityKey(tt.umlVersion, tt.diagramName, tt.entityID)
			if result != tt.expected {
				t.Errorf("EntityKey(%q, %q, %q) = %q, want %q", tt.umlVersion, tt.diagramName, tt.entityID, result, tt.expected)
			}
		})
	}
}

func TestValidateKey(t *testing.T) {
	kg := NewKeyGenerator()

	validKeys := []string{
		"/diagrams/puml/simple",
		"/diagrams/puml/my_diagram",
		"/diagrams/puml/diagram.v1.0",
		"/machines/2.5/simple",
		"/machines/version_2.5/my_diagram",
		"/machines/2.5/simple/entities/state1",
		"/machines/v2.5-beta/diagram-with%3Aspecial/entities/entity-id%3A1",
	}

	for _, key := range validKeys {
		t.Run("valid_"+key, func(t *testing.T) {
			err := kg.ValidateKey(key)
			if err != nil {
				t.Errorf("ValidateKey(%q) returned error: %v", key, err)
			}
		})
	}

	invalidKeys := []struct {
		key           string
		expectedError string
	}{
		{
			key:           "",
			expectedError: "key cannot be empty",
		},
		{
			key:           "no-leading-slash",
			expectedError: "key must start with '/'",
		},
		{
			key:           "/diagrams/puml/name with spaces",
			expectedError: "key contains invalid characters",
		},
		{
			key:           "/diagrams//puml/double-slash",
			expectedError: "key contains double slashes",
		},
		{
			key:           "/diagrams/puml/" + strings.Repeat("a", 250),
			expectedError: "key exceeds maximum length",
		},
		{
			key:           "/invalid/pattern/here",
			expectedError: "key does not match any expected pattern",
		},
		{
			key:           "/diagrams/puml/",
			expectedError: "diagram name cannot be empty",
		},
		{
			key:           "/diagrams/wrong/format",
			expectedError: "key does not match any expected pattern",
		},
		{
			key:           "/machines/",
			expectedError: "invalid machine key format",
		},
		{
			key:           "/machines//diagram",
			expectedError: "key contains double slashes",
		},
		{
			key:           "/machines/version/",
			expectedError: "diagram name cannot be empty",
		},
		{
			key:           "/machines/version/diagram/entities/",
			expectedError: "entity ID cannot be empty",
		},
		{
			key:           "/machines/version/diagram/invalid/path",
			expectedError: "invalid machine key format",
		},
	}

	for _, tt := range invalidKeys {
		t.Run("invalid_"+tt.key, func(t *testing.T) {
			err := kg.ValidateKey(tt.key)
			if err == nil {
				t.Errorf("ValidateKey(%q) expected error but got nil", tt.key)
			} else if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("ValidateKey(%q) error = %q, want error containing %q", tt.key, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	kg := &DefaultKeyGenerator{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "simple",
			expected: "simple",
		},
		{
			name:     "name with spaces",
			input:    "my diagram",
			expected: "my_diagram",
		},
		{
			name:     "name with forward slash",
			input:    "path/to/diagram",
			expected: "path-to-diagram",
		},
		{
			name:     "name with backslash",
			input:    "path\\to\\diagram",
			expected: "path-to-diagram",
		},
		{
			name:     "name with mixed special chars",
			input:    "diagram/with\\spaces and:colons",
			expected: "diagram-with-spaces_and%3Acolons",
		},
		{
			name:     "empty name",
			input:    "",
			expected: "",
		},
		{
			name:     "name with dots and dashes",
			input:    "diagram.v1.0-beta",
			expected: "diagram.v1.0-beta",
		},
		{
			name:     "name with underscores",
			input:    "my_diagram_name",
			expected: "my_diagram_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kg.sanitizeName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestKeyGenerationPatterns(t *testing.T) {
	kg := NewKeyGenerator()

	// Test that generated keys are valid
	testCases := []struct {
		name        string
		keyFunc     func() string
		description string
	}{
		{
			name:        "diagram key validation",
			keyFunc:     func() string { return kg.DiagramKey("test diagram") },
			description: "diagram key should be valid",
		},
		{
			name:        "state machine key validation",
			keyFunc:     func() string { return kg.StateMachineKey("2.5", "test diagram") },
			description: "state machine key should be valid",
		},
		{
			name:        "entity key validation",
			keyFunc:     func() string { return kg.EntityKey("2.5", "test diagram", "entity1") },
			description: "entity key should be valid",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.keyFunc()
			err := kg.ValidateKey(key)
			if err != nil {
				t.Errorf("%s: generated key %q failed validation: %v", tt.description, key, err)
			}
		})
	}
}

func TestKeyGeneratorConsistency(t *testing.T) {
	kg := NewKeyGenerator()

	// Test that the same inputs always produce the same outputs
	name := "test diagram with spaces"
	version := "2.5"
	entityID := "state1"

	// Generate keys multiple times
	for i := 0; i < 5; i++ {
		diagramKey1 := kg.DiagramKey(name)
		diagramKey2 := kg.DiagramKey(name)
		if diagramKey1 != diagramKey2 {
			t.Errorf("DiagramKey not consistent: %q != %q", diagramKey1, diagramKey2)
		}

		machineKey1 := kg.StateMachineKey(version, name)
		machineKey2 := kg.StateMachineKey(version, name)
		if machineKey1 != machineKey2 {
			t.Errorf("StateMachineKey not consistent: %q != %q", machineKey1, machineKey2)
		}

		entityKey1 := kg.EntityKey(version, name, entityID)
		entityKey2 := kg.EntityKey(version, name, entityID)
		if entityKey1 != entityKey2 {
			t.Errorf("EntityKey not consistent: %q != %q", entityKey1, entityKey2)
		}
	}
}
