package internal

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// InputValidator provides comprehensive input validation and sanitization
type InputValidator struct {
	maxStringLength     int
	maxKeyLength        int
	allowedCharPattern  *regexp.Regexp
	sqlInjectionPattern *regexp.Regexp
	xssPattern          *regexp.Regexp
}

// NewInputValidator creates a new input validator with default settings
func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxStringLength:     10000, // 10KB max for string inputs
		maxKeyLength:        250,   // Redis key length limit
		allowedCharPattern:  regexp.MustCompile(`^[\w\-_/.%\s]+$`),
		sqlInjectionPattern: regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|vbscript|onload|onerror|onclick)`),
		xssPattern:          regexp.MustCompile(`(?i)(<script|javascript:|vbscript:|onload=|onerror=|onclick=|<iframe|<object|<embed)`),
	}
}

// ValidateAndSanitizeString validates and sanitizes a general string input
func (v *InputValidator) ValidateAndSanitizeString(input, fieldName string) (string, error) {
	if input == "" {
		return "", NewValidationError(fmt.Sprintf("%s cannot be empty", fieldName), nil)
	}

	// Check length
	if len(input) > v.maxStringLength {
		return "", NewValidationError(fmt.Sprintf("%s exceeds maximum length of %d characters", fieldName, v.maxStringLength), nil)
	}

	// Check for valid UTF-8
	if !utf8.ValidString(input) {
		return "", NewValidationError(fmt.Sprintf("%s contains invalid UTF-8 characters", fieldName), nil)
	}

	// Sanitize first, then check for remaining control characters
	sanitized := v.sanitizeString(input)

	// Check for control characters in sanitized string (except tab, newline, carriage return)
	for i, r := range sanitized {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return "", NewValidationError(fmt.Sprintf("%s contains control character at position %d", fieldName, i), nil)
		}
	}

	// Check for potential security threats in sanitized string
	if err := v.checkSecurityThreats(sanitized, fieldName); err != nil {
		return "", err
	}

	return sanitized, nil
}

// ValidateAndSanitizeName validates and sanitizes names (diagram names, entity IDs, etc.)
func (v *InputValidator) ValidateAndSanitizeName(name, fieldName string) (string, error) {
	if name == "" {
		return "", NewValidationError(fmt.Sprintf("%s cannot be empty", fieldName), nil)
	}

	// Names have stricter length limits
	if len(name) > 100 {
		return "", NewValidationError(fmt.Sprintf("%s exceeds maximum length of 100 characters", fieldName), nil)
	}

	// Check for valid UTF-8
	if !utf8.ValidString(name) {
		return "", NewValidationError(fmt.Sprintf("%s contains invalid UTF-8 characters", fieldName), nil)
	}

	// Check for control characters
	for i, r := range name {
		if unicode.IsControl(r) {
			return "", NewValidationError(fmt.Sprintf("%s contains control character at position %d", fieldName, i), nil)
		}
	}

	// Check for path traversal attempts
	if strings.Contains(name, "..") {
		return "", NewValidationError(fmt.Sprintf("%s contains path traversal sequence", fieldName), nil)
	}

	// Check for null bytes (security)
	if strings.Contains(name, "\x00") {
		return "", NewValidationError(fmt.Sprintf("%s contains null bytes", fieldName), nil)
	}

	// Sanitize the name for use in keys
	sanitized := v.sanitizeName(name)

	return sanitized, nil
}

// ValidateAndSanitizeVersion validates and sanitizes version strings
func (v *InputValidator) ValidateAndSanitizeVersion(version, fieldName string) (string, error) {
	if version == "" {
		return "", NewValidationError(fmt.Sprintf("%s cannot be empty", fieldName), nil)
	}

	// Version strings should be shorter
	if len(version) > 50 {
		return "", NewValidationError(fmt.Sprintf("%s exceeds maximum length of 50 characters", fieldName), nil)
	}

	// Check for valid UTF-8
	if !utf8.ValidString(version) {
		return "", NewValidationError(fmt.Sprintf("%s contains invalid UTF-8 characters", fieldName), nil)
	}

	// Sanitize version first (trim spaces)
	sanitized := strings.TrimSpace(version)

	// Version strings should only contain alphanumeric, dots, dashes, and underscores
	versionPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !versionPattern.MatchString(sanitized) {
		return "", NewValidationError(fmt.Sprintf("%s contains invalid characters (only alphanumeric, dots, dashes, and underscores allowed)", fieldName), nil)
	}

	return sanitized, nil
}

// ValidateAndSanitizeContent validates and sanitizes content (PUML, JSON, etc.)
func (v *InputValidator) ValidateAndSanitizeContent(content, fieldName string) (string, error) {
	if content == "" {
		return "", NewValidationError(fmt.Sprintf("%s cannot be empty", fieldName), nil)
	}

	// Content can be larger but still has limits
	maxContentLength := 1024 * 1024 // 1MB max for content
	if len(content) > maxContentLength {
		return "", NewValidationError(fmt.Sprintf("%s exceeds maximum length of %d bytes", fieldName, maxContentLength), nil)
	}

	// Check for valid UTF-8
	if !utf8.ValidString(content) {
		return "", NewValidationError(fmt.Sprintf("%s contains invalid UTF-8 characters", fieldName), nil)
	}

	// Check for potential security threats in content
	if err := v.checkSecurityThreats(content, fieldName); err != nil {
		return "", err
	}

	// Minimal sanitization for content (preserve formatting)
	sanitized := v.sanitizeContent(content)

	return sanitized, nil
}

// ValidateOperation validates operation strings for entity mapping
func (v *InputValidator) ValidateOperation(operation string) error {
	if operation == "" {
		return NewValidationError("operation cannot be empty", nil)
	}

	validOperations := map[string]bool{
		"add":    true,
		"remove": true,
	}

	if !validOperations[operation] {
		return NewValidationError("operation must be 'add' or 'remove'", nil)
	}

	return nil
}

// ValidateContext validates context for timeout and cancellation
func (v *InputValidator) ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return NewValidationError("context cannot be nil", nil)
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return NewValidationError("context is already cancelled", ctx.Err())
	default:
		return nil
	}
}

// ValidateTTL validates time-to-live duration
func (v *InputValidator) ValidateTTL(ttl time.Duration, allowZero bool) error {
	if ttl < 0 {
		return NewValidationError("TTL cannot be negative", nil)
	}

	if !allowZero && ttl == 0 {
		return NewValidationError("TTL cannot be zero", nil)
	}

	// Check for reasonable upper bound (1 year)
	maxTTL := 365 * 24 * time.Hour
	if ttl > maxTTL {
		return NewValidationError(fmt.Sprintf("TTL exceeds maximum allowed duration of %v", maxTTL), nil)
	}

	return nil
}

// ValidateCleanupPattern validates cleanup patterns for security
func (v *InputValidator) ValidateCleanupPattern(pattern string) error {
	if pattern == "" {
		return NewValidationError("cleanup pattern cannot be empty", nil)
	}

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"*",  // Too broad
		"/*", // Too broad
		"",   // Empty
		"..", // Path traversal
	}

	for _, dangerous := range dangerousPatterns {
		if pattern == dangerous {
			return NewValidationError(fmt.Sprintf("cleanup pattern '%s' is too dangerous", pattern), nil)
		}
	}

	// Pattern should start with a specific prefix
	validPrefixes := []string{
		"/diagrams/",
		"/machines/",
	}

	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(pattern, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return NewValidationError("cleanup pattern must start with a valid prefix (/diagrams/ or /machines/)", nil)
	}

	return nil
}

// checkSecurityThreats checks for common security threats in input
func (v *InputValidator) checkSecurityThreats(input, fieldName string) error {
	// Check for SQL injection patterns
	if v.sqlInjectionPattern.MatchString(input) {
		return NewValidationError(fmt.Sprintf("%s contains potential SQL injection patterns", fieldName), nil)
	}

	// Check for XSS patterns
	if v.xssPattern.MatchString(input) {
		return NewValidationError(fmt.Sprintf("%s contains potential XSS patterns", fieldName), nil)
	}

	// Check for script injection
	scriptPatterns := []string{
		"<script",
		"javascript:",
		"vbscript:",
		"data:text/html",
		"data:application/javascript",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range scriptPatterns {
		if strings.Contains(lowerInput, pattern) {
			return NewValidationError(fmt.Sprintf("%s contains potential script injection pattern: %s", fieldName, pattern), nil)
		}
	}

	return nil
}

// sanitizeString performs general string sanitization
func (v *InputValidator) sanitizeString(input string) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(input)

	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// Normalize line endings
	sanitized = strings.ReplaceAll(sanitized, "\r\n", "\n")
	sanitized = strings.ReplaceAll(sanitized, "\r", "\n")

	return sanitized
}

// sanitizeName performs name-specific sanitization for use in cache keys
func (v *InputValidator) sanitizeName(name string) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(name)

	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// URL encode for safe use in keys, then apply custom replacements
	encoded := url.QueryEscape(sanitized)

	// Replace some URL-encoded characters with more readable alternatives
	encoded = strings.ReplaceAll(encoded, "+", "_")   // spaces (encoded as +) become underscores
	encoded = strings.ReplaceAll(encoded, "%20", "_") // spaces (encoded as %20) become underscores
	encoded = strings.ReplaceAll(encoded, "%2F", "-") // forward slashes become dashes
	encoded = strings.ReplaceAll(encoded, "%5C", "-") // backslashes become dashes

	return encoded
}

// sanitizeContent performs content-specific sanitization
func (v *InputValidator) sanitizeContent(content string) string {
	// Minimal sanitization to preserve content structure

	// Remove null bytes
	sanitized := strings.ReplaceAll(content, "\x00", "")

	// Normalize line endings but preserve them
	sanitized = strings.ReplaceAll(sanitized, "\r\n", "\n")
	sanitized = strings.ReplaceAll(sanitized, "\r", "\n")

	// Trim only leading/trailing whitespace, preserve internal formatting
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// ValidateEntityData validates entity data before storage
func (v *InputValidator) ValidateEntityData(entity interface{}, fieldName string) error {
	if entity == nil {
		return NewValidationError(fmt.Sprintf("%s cannot be nil", fieldName), nil)
	}

	// Additional validation could be added here for specific entity types
	// For now, we just check that it's not nil

	return nil
}
