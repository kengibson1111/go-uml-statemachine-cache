package internal

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// KeyGenerator defines the interface for generating and validating cache keys
type KeyGenerator interface {
	DiagramKey(name string) string
	StateMachineKey(umlVersion, name string) string
	EntityKey(umlVersion, diagramName, entityID string) string
	ValidateKey(key string) error
}

// DefaultKeyGenerator implements the KeyGenerator interface
type DefaultKeyGenerator struct{}

// NewKeyGenerator creates a new DefaultKeyGenerator instance
func NewKeyGenerator() KeyGenerator {
	return &DefaultKeyGenerator{}
}

// DiagramKey generates a cache key for PlantUML diagrams
// Format: /diagrams/puml/<diagram_name>
func (kg *DefaultKeyGenerator) DiagramKey(name string) string {
	sanitizedName := kg.sanitizeName(name)
	return fmt.Sprintf("/diagrams/puml/%s", sanitizedName)
}

// StateMachineKey generates a cache key for parsed state machines
// Format: /machines/<uml_version>/<diagram_name>
func (kg *DefaultKeyGenerator) StateMachineKey(umlVersion, name string) string {
	sanitizedVersion := kg.sanitizeName(umlVersion)
	sanitizedName := kg.sanitizeName(name)
	return fmt.Sprintf("/machines/%s/%s", sanitizedVersion, sanitizedName)
}

// EntityKey generates a cache key for state machine entities
// Format: /machines/<uml_version>/<diagram_name>/entities/<entity_id>
func (kg *DefaultKeyGenerator) EntityKey(umlVersion, diagramName, entityID string) string {
	sanitizedVersion := kg.sanitizeName(umlVersion)
	sanitizedDiagramName := kg.sanitizeName(diagramName)
	sanitizedEntityID := kg.sanitizeName(entityID)
	return fmt.Sprintf("/machines/%s/%s/entities/%s", sanitizedVersion, sanitizedDiagramName, sanitizedEntityID)
}

// ValidateKey validates that a cache key follows the expected format and constraints
func (kg *DefaultKeyGenerator) ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if !strings.HasPrefix(key, "/") {
		return fmt.Errorf("key must start with '/'")
	}

	// Check for control characters and null bytes (security)
	for i, r := range key {
		if r < 32 || r == 127 { // Control characters and DEL
			return fmt.Errorf("key contains control character at position %d: %s", i, key)
		}
	}

	// Check for unencoded path traversal attempts
	// Encoded path traversal (like %2E%2E or ..%2F) is allowed as it's URL encoded
	if strings.Contains(key, "../") || strings.Contains(key, "..\\") || key == ".." || strings.HasSuffix(key, "/..") {
		return fmt.Errorf("key contains path traversal sequence: %s", key)
	}

	// Check for invalid characters that could cause issues in Redis
	// Allow alphanumeric, dash, underscore, forward slash, dot, and percent-encoded characters
	invalidChars := regexp.MustCompile(`[^\w\-_/.%]`)
	if invalidChars.MatchString(key) {
		return fmt.Errorf("key contains invalid characters: %s", key)
	}

	// Check for double slashes or other path issues
	if strings.Contains(key, "//") {
		return fmt.Errorf("key contains double slashes: %s", key)
	}

	// Check maximum key length (Redis has a 512MB limit, but we'll be more conservative)
	if len(key) > 250 {
		return fmt.Errorf("key exceeds maximum length of 250 characters")
	}

	// Validate specific key patterns
	if strings.HasPrefix(key, "/diagrams/puml/") {
		return kg.validateDiagramKey(key)
	} else if strings.HasPrefix(key, "/machines/") {
		return kg.validateMachineKey(key)
	} else {
		return fmt.Errorf("key does not match any expected pattern: %s", key)
	}
}

// sanitizeName sanitizes a name for use in cache keys by URL encoding special characters
func (kg *DefaultKeyGenerator) sanitizeName(name string) string {
	if name == "" {
		return ""
	}

	// URL encode the name to handle special characters
	encoded := url.QueryEscape(name)

	// Replace some URL-encoded characters with more readable alternatives
	encoded = strings.ReplaceAll(encoded, "+", "_")   // spaces (encoded as +) become underscores
	encoded = strings.ReplaceAll(encoded, "%20", "_") // spaces (encoded as %20) become underscores
	encoded = strings.ReplaceAll(encoded, "%2F", "-") // forward slashes become dashes
	encoded = strings.ReplaceAll(encoded, "%5C", "-") // backslashes become dashes

	return encoded
}

// validateDiagramKey validates diagram-specific key format
func (kg *DefaultKeyGenerator) validateDiagramKey(key string) error {
	parts := strings.Split(key, "/")
	if len(parts) != 4 || parts[0] != "" || parts[1] != "diagrams" || parts[2] != "puml" {
		return fmt.Errorf("invalid diagram key format: %s", key)
	}

	if parts[3] == "" {
		return fmt.Errorf("diagram name cannot be empty in key: %s", key)
	}

	return nil
}

// validateMachineKey validates state machine and entity key formats
func (kg *DefaultKeyGenerator) validateMachineKey(key string) error {
	parts := strings.Split(key, "/")
	if len(parts) < 4 || parts[0] != "" || parts[1] != "machines" {
		return fmt.Errorf("invalid machine key format: %s", key)
	}

	if parts[2] == "" {
		return fmt.Errorf("UML version cannot be empty in key: %s", key)
	}

	if parts[3] == "" {
		return fmt.Errorf("diagram name cannot be empty in key: %s", key)
	}

	// Check if it's an entity key
	if len(parts) == 6 && parts[4] == "entities" {
		if parts[5] == "" {
			return fmt.Errorf("entity ID cannot be empty in key: %s", key)
		}
	} else if len(parts) != 4 {
		return fmt.Errorf("invalid machine key format: %s", key)
	}

	return nil
}
