package internal

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInputValidator_ValidateAndSanitizeString(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		input       string
		fieldName   string
		expected    string
		expectError bool
		errorType   ErrorType
	}{
		{
			name:        "empty string",
			input:       "",
			fieldName:   "test field",
			expected:    "",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "valid string",
			input:       "valid content",
			fieldName:   "test field",
			expected:    "valid content",
			expectError: false,
		},
		{
			name:        "string with whitespace",
			input:       "  content with spaces  ",
			fieldName:   "test field",
			expected:    "content with spaces",
			expectError: false,
		},
		{
			name:        "string with null bytes",
			input:       "content\x00with\x00nulls",
			fieldName:   "test field",
			expected:    "contentwithnulls",
			expectError: false,
		},
		{
			name:        "string with control characters",
			input:       "content\x01with\x02control",
			fieldName:   "test field",
			expected:    "",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "string with allowed control characters",
			input:       "content\twith\ntabs\rand\nlines",
			fieldName:   "test field",
			expected:    "content\twith\ntabs\nand\nlines",
			expectError: false,
		},
		{
			name:        "string too long",
			input:       strings.Repeat("a", 10001),
			fieldName:   "test field",
			expected:    "",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "string with SQL injection pattern",
			input:       "'; DROP TABLE users; --",
			fieldName:   "test field",
			expected:    "",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "string with XSS pattern",
			input:       "<script>alert('xss')</script>",
			fieldName:   "test field",
			expected:    "",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "string with javascript protocol",
			input:       "javascript:alert('xss')",
			fieldName:   "test field",
			expected:    "",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateAndSanitizeString(tt.input, tt.fieldName)

			if tt.expectError {
				require.Error(t, err)
				if cacheErr, ok := err.(*CacheError); ok {
					assert.Equal(t, tt.errorType, cacheErr.Type)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestInputValidator_ValidateAndSanitizeName(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		input       string
		fieldName   string
		expected    string
		expectError bool
	}{
		{
			name:        "empty name",
			input:       "",
			fieldName:   "diagram name",
			expected:    "",
			expectError: true,
		},
		{
			name:        "valid simple name",
			input:       "test-diagram",
			fieldName:   "diagram name",
			expected:    "test-diagram",
			expectError: false,
		},
		{
			name:        "name with spaces",
			input:       "test diagram with spaces",
			fieldName:   "diagram name",
			expected:    "test_diagram_with_spaces",
			expectError: false,
		},
		{
			name:        "name with special characters",
			input:       "test/diagram-with_special.chars",
			fieldName:   "diagram name",
			expected:    "test-diagram-with_special.chars",
			expectError: false,
		},
		{
			name:        "name too long",
			input:       strings.Repeat("a", 101),
			fieldName:   "diagram name",
			expected:    "",
			expectError: true,
		},
		{
			name:        "name with path traversal",
			input:       "../../../etc/passwd",
			fieldName:   "diagram name",
			expected:    "",
			expectError: true,
		},
		{
			name:        "name with null bytes",
			input:       "test\x00name",
			fieldName:   "diagram name",
			expected:    "",
			expectError: true,
		},
		{
			name:        "name with control characters",
			input:       "test\x01name",
			fieldName:   "diagram name",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateAndSanitizeName(tt.input, tt.fieldName)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestInputValidator_ValidateAndSanitizeVersion(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		input       string
		fieldName   string
		expected    string
		expectError bool
	}{
		{
			name:        "empty version",
			input:       "",
			fieldName:   "UML version",
			expected:    "",
			expectError: true,
		},
		{
			name:        "valid version",
			input:       "1.0.0",
			fieldName:   "UML version",
			expected:    "1.0.0",
			expectError: false,
		},
		{
			name:        "version with dashes and underscores",
			input:       "1.0-beta_1",
			fieldName:   "UML version",
			expected:    "1.0-beta_1",
			expectError: false,
		},
		{
			name:        "version too long",
			input:       strings.Repeat("1", 51),
			fieldName:   "UML version",
			expected:    "",
			expectError: true,
		},
		{
			name:        "version with invalid characters",
			input:       "1.0@beta",
			fieldName:   "UML version",
			expected:    "",
			expectError: true,
		},
		{
			name:        "version with spaces",
			input:       " 1.0.0 ",
			fieldName:   "UML version",
			expected:    "1.0.0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateAndSanitizeVersion(tt.input, tt.fieldName)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestInputValidator_ValidateAndSanitizeContent(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		input       string
		fieldName   string
		expected    string
		expectError bool
	}{
		{
			name:        "empty content",
			input:       "",
			fieldName:   "diagram content",
			expected:    "",
			expectError: true,
		},
		{
			name:        "valid PUML content",
			input:       "@startuml\nstate A\nstate B\nA --> B\n@enduml",
			fieldName:   "diagram content",
			expected:    "@startuml\nstate A\nstate B\nA --> B\n@enduml",
			expectError: false,
		},
		{
			name:        "content with mixed line endings",
			input:       "@startuml\r\nstate A\rstate B\n@enduml",
			fieldName:   "diagram content",
			expected:    "@startuml\nstate A\nstate B\n@enduml",
			expectError: false,
		},
		{
			name:        "content with null bytes",
			input:       "@startuml\x00\nstate A\n@enduml",
			fieldName:   "diagram content",
			expected:    "@startuml\nstate A\n@enduml",
			expectError: false,
		},
		{
			name:        "content too large",
			input:       strings.Repeat("a", 1024*1024+1),
			fieldName:   "diagram content",
			expected:    "",
			expectError: true,
		},
		{
			name:        "content with script injection",
			input:       "@startuml\n<script>alert('xss')</script>\n@enduml",
			fieldName:   "diagram content",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateAndSanitizeContent(tt.input, tt.fieldName)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestInputValidator_ValidateOperation(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		operation   string
		expectError bool
	}{
		{
			name:        "empty operation",
			operation:   "",
			expectError: true,
		},
		{
			name:        "valid add operation",
			operation:   "add",
			expectError: false,
		},
		{
			name:        "valid remove operation",
			operation:   "remove",
			expectError: false,
		},
		{
			name:        "invalid operation",
			operation:   "update",
			expectError: true,
		},
		{
			name:        "case sensitive operation",
			operation:   "ADD",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateOperation(tt.operation)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInputValidator_ValidateContext(t *testing.T) {
	validator := NewInputValidator()

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
			name:        "valid context",
			ctx:         context.Background(),
			expectError: false,
		},
		{
			name:        "cancelled context",
			ctx:         func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateContext(tt.ctx)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInputValidator_ValidateTTL(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		ttl         time.Duration
		allowZero   bool
		expectError bool
	}{
		{
			name:        "negative TTL",
			ttl:         -1 * time.Second,
			allowZero:   false,
			expectError: true,
		},
		{
			name:        "zero TTL not allowed",
			ttl:         0,
			allowZero:   false,
			expectError: true,
		},
		{
			name:        "zero TTL allowed",
			ttl:         0,
			allowZero:   true,
			expectError: false,
		},
		{
			name:        "valid TTL",
			ttl:         1 * time.Hour,
			allowZero:   false,
			expectError: false,
		},
		{
			name:        "TTL too large",
			ttl:         400 * 24 * time.Hour, // More than 1 year
			allowZero:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTTL(tt.ttl, tt.allowZero)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInputValidator_ValidateCleanupPattern(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		pattern     string
		expectError bool
	}{
		{
			name:        "empty pattern",
			pattern:     "",
			expectError: true,
		},
		{
			name:        "dangerous wildcard pattern",
			pattern:     "*",
			expectError: true,
		},
		{
			name:        "dangerous root pattern",
			pattern:     "/*",
			expectError: true,
		},
		{
			name:        "path traversal pattern",
			pattern:     "..",
			expectError: true,
		},
		{
			name:        "valid diagram pattern",
			pattern:     fmt.Sprintf("/diagrams/%s/test*", models.DiagramTypePUML.String()),
			expectError: false,
		},
		{
			name:        "valid machine pattern",
			pattern:     "/machines/1.0/test*",
			expectError: false,
		},
		{
			name:        "invalid prefix pattern",
			pattern:     "/invalid/test*",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCleanupPattern(tt.pattern)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInputValidator_ValidateEntityData(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		entity      interface{}
		fieldName   string
		expectError bool
	}{
		{
			name:        "nil entity",
			entity:      nil,
			fieldName:   "entity",
			expectError: true,
		},
		{
			name:        "valid entity",
			entity:      map[string]interface{}{"id": "test", "name": "Test Entity"},
			fieldName:   "entity",
			expectError: false,
		},
		{
			name:        "string entity",
			entity:      "test entity",
			fieldName:   "entity",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateEntityData(tt.entity, tt.fieldName)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Note: CleanupOptions validation is now handled in the cache package

func TestInputValidator_SecurityThreats(t *testing.T) {
	validator := NewInputValidator()

	securityTests := []struct {
		name     string
		input    string
		expected bool // true if should be blocked
	}{
		// SQL Injection patterns
		{"SQL injection - union", "'; UNION SELECT * FROM users; --", true},
		{"SQL injection - drop", "'; DROP TABLE users; --", true},
		{"SQL injection - insert", "'; INSERT INTO users VALUES ('admin', 'password'); --", true},
		{"SQL injection - update", "'; UPDATE users SET password='hacked'; --", true},
		{"SQL injection - delete", "'; DELETE FROM users; --", true},
		{"SQL injection - select", "' OR 1=1; SELECT * FROM users; --", true},

		// XSS patterns
		{"XSS - script tag", "<script>alert('xss')</script>", true},
		{"XSS - javascript protocol", "javascript:alert('xss')", true},
		{"XSS - vbscript protocol", "vbscript:msgbox('xss')", true},
		{"XSS - onload event", "<body onload='alert(1)'>", true},
		{"XSS - onerror event", "<img src=x onerror='alert(1)'>", true},
		{"XSS - onclick event", "<div onclick='alert(1)'>", true},
		{"XSS - iframe tag", "<iframe src='javascript:alert(1)'></iframe>", true},
		{"XSS - object tag", "<object data='javascript:alert(1)'></object>", true},
		{"XSS - embed tag", "<embed src='javascript:alert(1)'></embed>", true},

		// Data URI attacks
		{"Data URI - HTML", "data:text/html,<script>alert(1)</script>", true},
		{"Data URI - JavaScript", "data:application/javascript,alert(1)", true},

		// Safe content
		{"Safe PlantUML", "@startuml\nstate A\nstate B\nA --> B\n@enduml", false},
		{"Safe JSON", `{"id": "test", "name": "Test Entity"}`, false},
		{"Safe text", "This is normal text content", false},
		{"Safe HTML-like", "A < B and B > C", false},
	}

	for _, tt := range securityTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validator.ValidateAndSanitizeString(tt.input, "test field")

			if tt.expected {
				require.Error(t, err, "Expected security threat to be detected: %s", tt.input)
				if cacheErr, ok := err.(*CacheError); ok {
					assert.Equal(t, ErrorTypeValidation, cacheErr.Type)
				}
			} else {
				require.NoError(t, err, "Safe content should not be blocked: %s", tt.input)
			}
		})
	}
}

func TestInputValidator_EdgeCases(t *testing.T) {
	validator := NewInputValidator()

	t.Run("unicode handling", func(t *testing.T) {
		// Test various Unicode characters
		unicodeTests := []struct {
			name     string
			input    string
			expected bool // true if should pass validation
		}{
			{"ASCII text", "Hello World", true},
			{"UTF-8 emoji", "Hello 👋 World", true},
			{"UTF-8 accents", "Café résumé naïve", true},
			{"UTF-8 Chinese", "你好世界", true},
			{"UTF-8 Arabic", "مرحبا بالعالم", true},
			{"Invalid UTF-8", string([]byte{0xff, 0xfe, 0xfd}), false},
		}

		for _, tt := range unicodeTests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := validator.ValidateAndSanitizeName(tt.input, "test field")

				if tt.expected {
					assert.NoError(t, err, "Valid Unicode should pass: %s", tt.input)
				} else {
					assert.Error(t, err, "Invalid Unicode should fail: %s", tt.input)
				}
			})
		}
	})

	t.Run("boundary conditions", func(t *testing.T) {
		// Test exact boundary conditions
		maxNameLength := strings.Repeat("a", 100)
		tooLongName := strings.Repeat("a", 101)

		_, err := validator.ValidateAndSanitizeName(maxNameLength, "test field")
		assert.NoError(t, err, "Name at max length should pass")

		_, err = validator.ValidateAndSanitizeName(tooLongName, "test field")
		assert.Error(t, err, "Name over max length should fail")
	})
}
