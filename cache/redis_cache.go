package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/internal"
	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/redis/go-redis/v9"
)

// types and funcs for external use.
type RedisConfig = internal.Config
type RedisRetryConfig = internal.RetryConfig
type CacheErrorType = internal.ErrorType
type CacheError = internal.CacheError
type ErrorContext = internal.ErrorContext
type ErrorSeverity = internal.ErrorSeverity
type RecoveryStrategy = internal.RecoveryStrategy
type CircuitBreakerConfig = internal.CircuitBreakerConfig
type ErrorRecoveryManager = internal.ErrorRecoveryManager

const (
	// ErrorTypeConnection indicates a Redis connection error
	CacheErrorTypeConnection CacheErrorType = internal.ErrorTypeConnection
	// ErrorTypeKeyInvalid indicates an invalid cache key
	CacheErrorTypeKeyInvalid = internal.ErrorTypeKeyInvalid
	// ErrorTypeNotFound indicates a cache miss or key not found
	CacheErrorTypeNotFound = internal.ErrorTypeNotFound
	// ErrorTypeSerialization indicates JSON marshaling/unmarshaling error
	CacheErrorTypeSerialization = internal.ErrorTypeSerialization
	// ErrorTypeTimeout indicates a timeout during cache operation
	CacheErrorTypeTimeout = internal.ErrorTypeTimeout
	// ErrorTypeCapacity indicates cache capacity or memory issues
	CacheErrorTypeCapacity = internal.ErrorTypeCapacity
	// ErrorTypeValidation indicates input validation failure
	CacheErrorTypeValidation = internal.ErrorTypeValidation
	// ErrorTypeRetryExhausted indicates all retry attempts have been exhausted
	CacheErrorTypeRetryExhausted = internal.ErrorTypeRetryExhausted
	// ErrorTypeCircuitOpen indicates circuit breaker is open
	CacheErrorTypeCircuitOpen = internal.ErrorTypeCircuitOpen

	// Error severity levels
	SeverityLow      ErrorSeverity = internal.SeverityLow
	SeverityMedium   ErrorSeverity = internal.SeverityMedium
	SeverityHigh     ErrorSeverity = internal.SeverityHigh
	SeverityCritical ErrorSeverity = internal.SeverityCritical

	// Recovery strategies
	RecoveryStrategyFail             RecoveryStrategy = internal.RecoveryStrategyFail
	RecoveryStrategyIgnore           RecoveryStrategy = internal.RecoveryStrategyIgnore
	RecoveryStrategyRetryWithBackoff RecoveryStrategy = internal.RecoveryStrategyRetryWithBackoff
	RecoveryStrategyRetryWithDelay   RecoveryStrategy = internal.RecoveryStrategyRetryWithDelay
	RecoveryStrategyCircuitBreaker   RecoveryStrategy = internal.RecoveryStrategyCircuitBreaker
	RecoveryStrategyWaitAndRetry     RecoveryStrategy = internal.RecoveryStrategyWaitAndRetry
)

// NewCacheError creates a new CacheError
func NewCacheError(errType CacheErrorType, key, message string, cause error) *CacheError {
	return internal.NewCacheError(errType, key, message, cause)
}

// NewConnectionError creates a connection-specific cache error
func NewConnectionError(message string, cause error) *CacheError {
	return internal.NewConnectionError(message, cause)
}

// NewKeyInvalidError creates a key validation error
func NewKeyInvalidError(key, message string) *CacheError {
	return internal.NewKeyInvalidError(key, message)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(key string) *CacheError {
	return internal.NewNotFoundError(key)
}

// NewSerializationError creates a serialization error
func NewSerializationError(key, message string, cause error) *CacheError {
	return internal.NewSerializationError(key, message, cause)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(key, message string, cause error) *CacheError {
	return internal.NewTimeoutError(key, message, cause)
}

// NewValidationError creates a validation error
func NewValidationError(message string, cause error) *CacheError {
	return internal.NewValidationError(message, cause)
}

// IsConnectionError checks if the error is a connection error
func IsConnectionError(err error) bool {
	return internal.IsConnectionError(err)
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	return internal.IsNotFoundError(err)
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	return internal.IsValidationError(err)
}

// IsRetryExhaustedError checks if the error is a retry exhausted error
func IsRetryExhaustedError(err error) bool {
	return internal.IsRetryExhaustedError(err)
}

// IsCircuitOpenError checks if the error is a circuit open error
func IsCircuitOpenError(err error) bool {
	return internal.IsCircuitOpenError(err)
}

// IsRetryableError checks if the error should trigger retry logic
func IsRetryableError(err error) bool {
	return internal.IsRetryableError(err)
}

// GetErrorSeverity returns the severity level of an error
func GetErrorSeverity(err error) ErrorSeverity {
	return internal.GetErrorSeverity(err)
}

// GetRecoveryStrategy returns the recommended recovery strategy for an error
func GetRecoveryStrategy(err error) RecoveryStrategy {
	return internal.GetRecoveryStrategy(err)
}

// NewCacheErrorWithContext creates a new CacheError with detailed context
func NewCacheErrorWithContext(errType CacheErrorType, key, message string, cause error, context *ErrorContext) *CacheError {
	return internal.NewCacheErrorWithContext(errType, key, message, cause, context)
}

// NewRetryExhaustedError creates a retry exhausted error
func NewRetryExhaustedError(operation string, attempts int, lastError error) *CacheError {
	return internal.NewRetryExhaustedError(operation, attempts, lastError)
}

// NewCircuitOpenError creates a circuit breaker open error
func NewCircuitOpenError(operation string) *CacheError {
	return internal.NewCircuitOpenError(operation)
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return internal.DefaultCircuitBreakerConfig()
}

// NewErrorRecoveryManager creates a new error recovery manager
func NewErrorRecoveryManager(config *CircuitBreakerConfig) *ErrorRecoveryManager {
	return internal.NewErrorRecoveryManager(config)
}

func DefaultRedisConfig() *RedisConfig {
	return internal.DefaultConfig()
}

func DefaultRedisRetryConfig() *RedisRetryConfig {
	return internal.DefaultRetryConfig()
}

// RedisCache implements the Cache interface using Redis as the backend
type RedisCache struct {
	client    internal.RedisClientInterface
	keyGen    internal.KeyGenerator
	config    *RedisConfig
	validator *internal.InputValidator
}

// NewRedisCache creates a new Redis-backed cache implementation
func NewRedisCache(config *RedisConfig) (*RedisCache, error) {
	client, err := internal.NewRedisClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	return &RedisCache{
		client:    client,
		keyGen:    internal.NewKeyGenerator(),
		config:    config,
		validator: internal.NewInputValidator(),
	}, nil
}

// NewRedisCacheWithDependencies creates a new Redis cache with injected dependencies for testing
func NewRedisCacheWithDependencies(client internal.RedisClientInterface, keyGen internal.KeyGenerator, config *RedisConfig) *RedisCache {
	return &RedisCache{
		client:    client,
		keyGen:    keyGen,
		config:    config,
		validator: internal.NewInputValidator(),
	}
}

// StoreDiagram stores a PlantUML diagram with TTL support
func (rc *RedisCache) StoreDiagram(ctx context.Context, name string, pumlContent string, ttl time.Duration) error {
	// Validate context
	if err := rc.validator.ValidateContext(ctx); err != nil {
		return err
	}

	// Validate and sanitize inputs
	sanitizedName, err := rc.validator.ValidateAndSanitizeName(name, "diagram name")
	if err != nil {
		return err
	}

	sanitizedContent, err := rc.validator.ValidateAndSanitizeContent(pumlContent, "diagram content")
	if err != nil {
		return err
	}

	// Validate TTL
	if err := rc.validator.ValidateTTL(ttl, true); err != nil {
		return err
	}

	// Generate cache key using sanitized name
	key := rc.keyGen.DiagramKey(sanitizedName)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Use default TTL if not specified
	if ttl <= 0 {
		ttl = rc.config.DefaultTTL
	}

	// Store the sanitized diagram content in Redis
	err = rc.client.SetWithRetry(ctx, key, sanitizedContent, ttl)
	if err != nil {
		if isTimeoutError(err) {
			return NewTimeoutError(key, "timeout storing diagram", err)
		}
		if isConnectionError(err) {
			return NewConnectionError("failed to store diagram", err)
		}
		return fmt.Errorf("failed to store diagram '%s': %w", sanitizedName, err)
	}

	return nil
}

// GetDiagram retrieves a PlantUML diagram with error handling
func (rc *RedisCache) GetDiagram(ctx context.Context, name string) (string, error) {
	// Validate context
	if err := rc.validator.ValidateContext(ctx); err != nil {
		return "", err
	}

	// Validate and sanitize name
	sanitizedName, err := rc.validator.ValidateAndSanitizeName(name, "diagram name")
	if err != nil {
		return "", err
	}

	// Generate cache key using sanitized name
	key := rc.keyGen.DiagramKey(sanitizedName)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return "", NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Retrieve the diagram content from Redis
	content, err := rc.client.GetWithRetry(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return "", NewNotFoundError(key)
		}
		if isTimeoutError(err) {
			return "", NewTimeoutError(key, "timeout retrieving diagram", err)
		}
		if isConnectionError(err) {
			return "", NewConnectionError("failed to retrieve diagram", err)
		}
		return "", fmt.Errorf("failed to retrieve diagram '%s': %w", sanitizedName, err)
	}

	return content, nil
}

// DeleteDiagram removes a diagram from the cache for cleanup
func (rc *RedisCache) DeleteDiagram(ctx context.Context, name string) error {
	// Validate context
	if err := rc.validator.ValidateContext(ctx); err != nil {
		return err
	}

	// Validate and sanitize name
	sanitizedName, err := rc.validator.ValidateAndSanitizeName(name, "diagram name")
	if err != nil {
		return err
	}

	// Generate cache key using sanitized name
	key := rc.keyGen.DiagramKey(sanitizedName)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Delete the diagram from Redis
	err = rc.client.DelWithRetry(ctx, key)
	if err != nil {
		if isTimeoutError(err) {
			return NewTimeoutError(key, "timeout deleting diagram", err)
		}
		if isConnectionError(err) {
			return NewConnectionError("failed to delete diagram", err)
		}
		return fmt.Errorf("failed to delete diagram '%s': %w", sanitizedName, err)
	}

	return nil
}

// StoreStateMachine stores a parsed state machine with TTL support and creates entity cache paths
func (rc *RedisCache) StoreStateMachine(ctx context.Context, umlVersion, name string, machine *models.StateMachine, ttl time.Duration) error {
	// Validate context
	if err := rc.validator.ValidateContext(ctx); err != nil {
		return err
	}

	// Validate and sanitize inputs
	sanitizedVersion, err := rc.validator.ValidateAndSanitizeVersion(umlVersion, "UML version")
	if err != nil {
		return err
	}

	sanitizedName, err := rc.validator.ValidateAndSanitizeName(name, "state machine name")
	if err != nil {
		return err
	}

	if err := rc.validator.ValidateEntityData(machine, "state machine"); err != nil {
		return err
	}

	// Validate TTL
	if err := rc.validator.ValidateTTL(ttl, true); err != nil {
		return err
	}

	// Validate that the corresponding diagram exists before storing state machine
	_, err = rc.GetDiagram(ctx, sanitizedName)
	if err != nil {
		if IsNotFoundError(err) {
			return NewValidationError(fmt.Sprintf("cannot store state machine: corresponding diagram '%s' does not exist", sanitizedName), err)
		}
		return fmt.Errorf("failed to validate diagram existence for state machine '%s': %w", sanitizedName, err)
	}

	// Generate cache key using sanitized inputs
	key := rc.keyGen.StateMachineKey(sanitizedVersion, sanitizedName)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Use default TTL if not specified
	if ttl <= 0 {
		ttl = rc.config.DefaultTTL
	}

	// Extract and store entities from the state machine, populate Entities map
	entities := rc.extractEntitiesFromStateMachine(machine)

	// Initialize or clear the Entities map to ensure referential integrity
	if machine.Entities == nil {
		machine.Entities = make(map[string]string)
	} else {
		// Clear existing entities to ensure consistency
		for entityID := range machine.Entities {
			delete(machine.Entities, entityID)
		}
	}

	// Sort entity IDs to ensure deterministic processing order
	var entityIDs []string
	for entityID := range entities {
		entityIDs = append(entityIDs, entityID)
	}

	// Sort to ensure consistent order
	for i := 0; i < len(entityIDs); i++ {
		for j := i + 1; j < len(entityIDs); j++ {
			if entityIDs[i] > entityIDs[j] {
				entityIDs[i], entityIDs[j] = entityIDs[j], entityIDs[i]
			}
		}
	}

	// Store each entity and populate the Entities map with cache paths
	for _, entityID := range entityIDs {
		entity := entities[entityID]
		entityKey := rc.keyGen.EntityKey(sanitizedVersion, sanitizedName, entityID)

		// Validate entity key before storing
		if err := rc.keyGen.ValidateKey(entityKey); err != nil {
			return NewKeyInvalidError(entityKey, fmt.Sprintf("invalid entity key generated for '%s': %v", entityID, err))
		}

		// Update the Entities mapping for referential integrity
		machine.Entities[entityID] = entityKey

		// Store the entity using sanitized values
		err := rc.StoreEntity(ctx, sanitizedVersion, sanitizedName, entityID, entity, ttl)
		if err != nil {
			// If entity storage fails, clean up any previously stored entities to maintain consistency
			rc.cleanupPartialEntityStorage(ctx, sanitizedVersion, sanitizedName, machine.Entities)
			return fmt.Errorf("failed to store entity '%s': %w", entityID, err)
		}
	}

	// Serialize the state machine to JSON (now with populated Entities map)
	data, err := json.Marshal(machine)
	if err != nil {
		// Clean up entities if state machine serialization fails
		rc.cleanupPartialEntityStorage(ctx, sanitizedVersion, sanitizedName, machine.Entities)
		return NewSerializationError(key, "failed to marshal state machine", err)
	}

	// Store the serialized state machine in Redis
	err = rc.client.SetWithRetry(ctx, key, data, ttl)
	if err != nil {
		// Clean up entities if state machine storage fails
		rc.cleanupPartialEntityStorage(ctx, sanitizedVersion, sanitizedName, machine.Entities)

		if isTimeoutError(err) {
			return NewTimeoutError(key, "timeout storing state machine", err)
		}
		if isConnectionError(err) {
			return NewConnectionError("failed to store state machine", err)
		}
		return fmt.Errorf("failed to store state machine '%s': %w", sanitizedName, err)
	}

	return nil
}

// GetStateMachine retrieves a parsed state machine
func (rc *RedisCache) GetStateMachine(ctx context.Context, umlVersion, name string) (*models.StateMachine, error) {
	// Validate context
	if err := rc.validator.ValidateContext(ctx); err != nil {
		return nil, err
	}

	// Validate and sanitize inputs
	sanitizedVersion, err := rc.validator.ValidateAndSanitizeVersion(umlVersion, "UML version")
	if err != nil {
		return nil, err
	}

	sanitizedName, err := rc.validator.ValidateAndSanitizeName(name, "state machine name")
	if err != nil {
		return nil, err
	}

	// Generate cache key using sanitized inputs
	key := rc.keyGen.StateMachineKey(sanitizedVersion, sanitizedName)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return nil, NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Retrieve the serialized state machine from Redis
	data, err := rc.client.GetWithRetry(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, NewNotFoundError(key)
		}
		if isTimeoutError(err) {
			return nil, NewTimeoutError(key, "timeout retrieving state machine", err)
		}
		if isConnectionError(err) {
			return nil, NewConnectionError("failed to retrieve state machine", err)
		}
		return nil, fmt.Errorf("failed to retrieve state machine '%s': %w", sanitizedName, err)
	}

	// Deserialize the state machine from JSON
	var machine models.StateMachine
	err = json.Unmarshal([]byte(data), &machine)
	if err != nil {
		return nil, NewSerializationError(key, "failed to unmarshal state machine", err)
	}

	return &machine, nil
}

// DeleteStateMachine removes a state machine from the cache with cascade deletion
func (rc *RedisCache) DeleteStateMachine(ctx context.Context, umlVersion, name string) error {
	// Validate context
	if err := rc.validator.ValidateContext(ctx); err != nil {
		return err
	}

	// Validate and sanitize inputs
	sanitizedVersion, err := rc.validator.ValidateAndSanitizeVersion(umlVersion, "UML version")
	if err != nil {
		return err
	}

	sanitizedName, err := rc.validator.ValidateAndSanitizeName(name, "state machine name")
	if err != nil {
		return err
	}

	// Generate cache key for the state machine using sanitized inputs
	key := rc.keyGen.StateMachineKey(sanitizedVersion, sanitizedName)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// First, try to retrieve the state machine to get entity information for cascade deletion
	machine, err := rc.GetStateMachine(ctx, sanitizedVersion, sanitizedName)
	if err != nil && !IsNotFoundError(err) {
		return fmt.Errorf("failed to retrieve state machine for cascade deletion: %w", err)
	}

	// Collect all keys to delete (state machine + entities)
	var keysToDelete []string
	keysToDelete = append(keysToDelete, key)

	// If state machine exists and has entities, add entity keys for cascade deletion
	if machine != nil && machine.Entities != nil {
		// Sort entity IDs to ensure deterministic order for testing
		var entityIDs []string
		for entityID := range machine.Entities {
			entityIDs = append(entityIDs, entityID)
		}

		// Sort to ensure consistent order
		for i := 0; i < len(entityIDs); i++ {
			for j := i + 1; j < len(entityIDs); j++ {
				if entityIDs[i] > entityIDs[j] {
					entityIDs[i], entityIDs[j] = entityIDs[j], entityIDs[i]
				}
			}
		}

		for _, entityID := range entityIDs {
			entityKey := rc.keyGen.EntityKey(sanitizedVersion, sanitizedName, entityID)
			keysToDelete = append(keysToDelete, entityKey)
		}
	}

	// Additional safety check: scan for any orphaned entity keys that might exist
	// but aren't tracked in the Entities map to ensure complete cleanup
	entityPattern := rc.keyGen.EntityKey(sanitizedVersion, sanitizedName, "*")
	orphanedKeys, err := rc.scanForOrphanedEntities(ctx, entityPattern, machine)
	if err != nil {
		// Log the error but don't fail the deletion - this is a best-effort cleanup
		// In a production system, you might want to log this for monitoring
	} else {
		keysToDelete = append(keysToDelete, orphanedKeys...)
	}

	// Delete all keys (state machine + entities) in a single operation
	if len(keysToDelete) > 0 {
		err = rc.client.DelWithRetry(ctx, keysToDelete...)
		if err != nil {
			if isTimeoutError(err) {
				return NewTimeoutError(key, "timeout deleting state machine and entities", err)
			}
			if isConnectionError(err) {
				return NewConnectionError("failed to delete state machine and entities", err)
			}
			return fmt.Errorf("failed to delete state machine '%s' and its entities: %w", sanitizedName, err)
		}
	}

	return nil
}

// scanForOrphanedEntities scans for entity keys that might exist but aren't tracked in the Entities map
func (rc *RedisCache) scanForOrphanedEntities(ctx context.Context, pattern string, machine *models.StateMachine) ([]string, error) {
	client := rc.client.Client()
	if client == nil {
		return nil, fmt.Errorf("Redis client not available")
	}

	iter := client.Scan(ctx, 0, pattern, 0).Iterator()
	var orphanedKeys []string

	// Track keys that are already in the Entities map
	trackedKeys := make(map[string]bool)
	if machine != nil && machine.Entities != nil {
		for _, entityKey := range machine.Entities {
			trackedKeys[entityKey] = true
		}
	}

	for iter.Next(ctx) {
		key := iter.Val()
		// Only add keys that aren't already tracked in the Entities map
		if !trackedKeys[key] {
			orphanedKeys = append(orphanedKeys, key)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan for orphaned entities: %w", err)
	}

	return orphanedKeys, nil
}

// StoreEntity stores a state machine entity with TTL support
func (rc *RedisCache) StoreEntity(ctx context.Context, umlVersion, diagramName, entityID string, entity interface{}, ttl time.Duration) error {
	// Validate context
	if err := rc.validator.ValidateContext(ctx); err != nil {
		return err
	}

	// Validate and sanitize inputs
	sanitizedVersion, err := rc.validator.ValidateAndSanitizeVersion(umlVersion, "UML version")
	if err != nil {
		return err
	}

	sanitizedDiagramName, err := rc.validator.ValidateAndSanitizeName(diagramName, "diagram name")
	if err != nil {
		return err
	}

	sanitizedEntityID, err := rc.validator.ValidateAndSanitizeName(entityID, "entity ID")
	if err != nil {
		return err
	}

	if err := rc.validator.ValidateEntityData(entity, "entity"); err != nil {
		return err
	}

	// Validate TTL
	if err := rc.validator.ValidateTTL(ttl, true); err != nil {
		return err
	}

	// Generate cache key using sanitized inputs
	key := rc.keyGen.EntityKey(sanitizedVersion, sanitizedDiagramName, sanitizedEntityID)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Serialize the entity to JSON
	data, err := json.Marshal(entity)
	if err != nil {
		return NewSerializationError(key, "failed to marshal entity", err)
	}

	// Use default TTL if not specified
	if ttl <= 0 {
		ttl = rc.config.DefaultTTL
	}

	// Store the serialized entity in Redis
	err = rc.client.SetWithRetry(ctx, key, data, ttl)
	if err != nil {
		if isTimeoutError(err) {
			return NewTimeoutError(key, "timeout storing entity", err)
		}
		if isConnectionError(err) {
			return NewConnectionError("failed to store entity", err)
		}
		return fmt.Errorf("failed to store entity '%s': %w", sanitizedEntityID, err)
	}

	return nil
}

// UpdateStateMachineEntityMapping updates the Entities mapping in a stored state machine
// This ensures referential integrity when entities are added or removed independently
func (rc *RedisCache) UpdateStateMachineEntityMapping(ctx context.Context, umlVersion, name string, entityID, entityKey string, operation string) error {
	// Validate context
	if err := rc.validator.ValidateContext(ctx); err != nil {
		return err
	}

	// Validate and sanitize inputs
	sanitizedVersion, err := rc.validator.ValidateAndSanitizeVersion(umlVersion, "UML version")
	if err != nil {
		return err
	}

	sanitizedName, err := rc.validator.ValidateAndSanitizeName(name, "state machine name")
	if err != nil {
		return err
	}

	sanitizedEntityID, err := rc.validator.ValidateAndSanitizeName(entityID, "entity ID")
	if err != nil {
		return err
	}

	if err := rc.validator.ValidateOperation(operation); err != nil {
		return err
	}

	if operation == "add" {
		if entityKey == "" {
			return NewValidationError("entity key cannot be empty for add operation", nil)
		}
		sanitizedEntityKey, err := rc.validator.ValidateAndSanitizeString(entityKey, "entity key")
		if err != nil {
			return err
		}
		entityKey = sanitizedEntityKey
	}

	// Get the current state machine using sanitized inputs
	machine, err := rc.GetStateMachine(ctx, sanitizedVersion, sanitizedName)
	if err != nil {
		return fmt.Errorf("failed to get state machine for entity mapping update: %w", err)
	}

	// Initialize Entities map if it doesn't exist
	if machine.Entities == nil {
		machine.Entities = make(map[string]string)
	}

	// Update the mapping based on operation
	switch operation {
	case "add":
		machine.Entities[sanitizedEntityID] = entityKey
	case "remove":
		delete(machine.Entities, sanitizedEntityID)
	}

	// Store the updated state machine using sanitized inputs
	key := rc.keyGen.StateMachineKey(sanitizedVersion, sanitizedName)
	data, err := json.Marshal(machine)
	if err != nil {
		return NewSerializationError(key, "failed to marshal updated state machine", err)
	}

	err = rc.client.SetWithRetry(ctx, key, data, rc.config.DefaultTTL)
	if err != nil {
		if isTimeoutError(err) {
			return NewTimeoutError(key, "timeout updating state machine entity mapping", err)
		}
		if isConnectionError(err) {
			return NewConnectionError("failed to update state machine entity mapping", err)
		}
		return fmt.Errorf("failed to update state machine entity mapping: %w", err)
	}

	return nil
}

// extractEntitiesFromStateMachine extracts all entities from a state machine for caching
func (rc *RedisCache) extractEntitiesFromStateMachine(machine *models.StateMachine) map[string]interface{} {
	entities := make(map[string]interface{})

	// Extract entities from all regions
	for _, region := range machine.Regions {
		rc.extractEntitiesFromRegion(region, entities)
	}

	return entities
}

// cleanupPartialEntityStorage removes entities that were stored during a failed state machine storage operation
func (rc *RedisCache) cleanupPartialEntityStorage(ctx context.Context, umlVersion, diagramName string, entityMap map[string]string) {
	if len(entityMap) == 0 {
		return
	}

	// Collect all entity keys to delete
	var keysToDelete []string
	for _, entityKey := range entityMap {
		keysToDelete = append(keysToDelete, entityKey)
	}

	// Attempt to delete the entities (best effort - don't propagate errors)
	if len(keysToDelete) > 0 {
		_ = rc.client.DelWithRetry(ctx, keysToDelete...)
	}
}

// extractEntitiesFromRegion recursively extracts entities from a region
func (rc *RedisCache) extractEntitiesFromRegion(region *models.Region, entities map[string]interface{}) {
	// Add the region itself as an entity
	if region.ID != "" {
		entities[region.ID] = region
	}

	// Extract states
	for _, state := range region.States {
		if state.ID != "" {
			entities[state.ID] = state
		}
		// Recursively extract from nested regions in composite states
		for _, nestedRegion := range state.Regions {
			rc.extractEntitiesFromRegion(nestedRegion, entities)
		}
	}

	// Extract transitions
	for _, transition := range region.Transitions {
		if transition.ID != "" {
			entities[transition.ID] = transition
		}
	}

	// Extract vertices (pseudostates, final states, etc.)
	for _, vertex := range region.Vertices {
		if vertex.ID != "" {
			entities[vertex.ID] = vertex
		}
	}
}

// GetEntity retrieves a state machine entity
func (rc *RedisCache) GetEntity(ctx context.Context, umlVersion, diagramName, entityID string) (interface{}, error) {
	// Validate context
	if err := rc.validator.ValidateContext(ctx); err != nil {
		return nil, err
	}

	// Validate and sanitize inputs
	sanitizedVersion, err := rc.validator.ValidateAndSanitizeVersion(umlVersion, "UML version")
	if err != nil {
		return nil, err
	}

	sanitizedDiagramName, err := rc.validator.ValidateAndSanitizeName(diagramName, "diagram name")
	if err != nil {
		return nil, err
	}

	sanitizedEntityID, err := rc.validator.ValidateAndSanitizeName(entityID, "entity ID")
	if err != nil {
		return nil, err
	}

	// Generate cache key using sanitized inputs
	key := rc.keyGen.EntityKey(sanitizedVersion, sanitizedDiagramName, sanitizedEntityID)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return nil, NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Retrieve the serialized entity from Redis
	data, err := rc.client.GetWithRetry(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, NewNotFoundError(key)
		}
		if isTimeoutError(err) {
			return nil, NewTimeoutError(key, "timeout retrieving entity", err)
		}
		if isConnectionError(err) {
			return nil, NewConnectionError("failed to retrieve entity", err)
		}
		return nil, fmt.Errorf("failed to retrieve entity '%s': %w", sanitizedEntityID, err)
	}

	// Deserialize the entity from JSON
	var entity interface{}
	err = json.Unmarshal([]byte(data), &entity)
	if err != nil {
		return nil, NewSerializationError(key, "failed to unmarshal entity", err)
	}

	return entity, nil
}

// GetEntityAsState retrieves a state machine entity and attempts to unmarshal it as a State
func (rc *RedisCache) GetEntityAsState(ctx context.Context, umlVersion, diagramName, entityID string) (*models.State, error) {
	entity, err := rc.GetEntity(ctx, umlVersion, diagramName, entityID)
	if err != nil {
		return nil, err
	}

	// Re-marshal and unmarshal to convert to specific type
	data, err := json.Marshal(entity)
	if err != nil {
		key := rc.keyGen.EntityKey(umlVersion, diagramName, entityID)
		return nil, NewSerializationError(key, "failed to marshal entity for type conversion", err)
	}

	var state models.State
	err = json.Unmarshal(data, &state)
	if err != nil {
		key := rc.keyGen.EntityKey(umlVersion, diagramName, entityID)
		return nil, NewSerializationError(key, "failed to unmarshal entity as State", err)
	}

	return &state, nil
}

// GetEntityAsTransition retrieves a state machine entity and attempts to unmarshal it as a Transition
func (rc *RedisCache) GetEntityAsTransition(ctx context.Context, umlVersion, diagramName, entityID string) (*models.Transition, error) {
	entity, err := rc.GetEntity(ctx, umlVersion, diagramName, entityID)
	if err != nil {
		return nil, err
	}

	// Re-marshal and unmarshal to convert to specific type
	data, err := json.Marshal(entity)
	if err != nil {
		key := rc.keyGen.EntityKey(umlVersion, diagramName, entityID)
		return nil, NewSerializationError(key, "failed to marshal entity for type conversion", err)
	}

	var transition models.Transition
	err = json.Unmarshal(data, &transition)
	if err != nil {
		key := rc.keyGen.EntityKey(umlVersion, diagramName, entityID)
		return nil, NewSerializationError(key, "failed to unmarshal entity as Transition", err)
	}

	return &transition, nil
}

// GetEntityAsRegion retrieves a state machine entity and attempts to unmarshal it as a Region
func (rc *RedisCache) GetEntityAsRegion(ctx context.Context, umlVersion, diagramName, entityID string) (*models.Region, error) {
	entity, err := rc.GetEntity(ctx, umlVersion, diagramName, entityID)
	if err != nil {
		return nil, err
	}

	// Re-marshal and unmarshal to convert to specific type
	data, err := json.Marshal(entity)
	if err != nil {
		key := rc.keyGen.EntityKey(umlVersion, diagramName, entityID)
		return nil, NewSerializationError(key, "failed to marshal entity for type conversion", err)
	}

	var region models.Region
	err = json.Unmarshal(data, &region)
	if err != nil {
		key := rc.keyGen.EntityKey(umlVersion, diagramName, entityID)
		return nil, NewSerializationError(key, "failed to unmarshal entity as Region", err)
	}

	return &region, nil
}

// GetEntityAsVertex retrieves a state machine entity and attempts to unmarshal it as a Vertex
func (rc *RedisCache) GetEntityAsVertex(ctx context.Context, umlVersion, diagramName, entityID string) (*models.Vertex, error) {
	entity, err := rc.GetEntity(ctx, umlVersion, diagramName, entityID)
	if err != nil {
		return nil, err
	}

	// Re-marshal and unmarshal to convert to specific type
	data, err := json.Marshal(entity)
	if err != nil {
		key := rc.keyGen.EntityKey(umlVersion, diagramName, entityID)
		return nil, NewSerializationError(key, "failed to marshal entity for type conversion", err)
	}

	var vertex models.Vertex
	err = json.Unmarshal(data, &vertex)
	if err != nil {
		key := rc.keyGen.EntityKey(umlVersion, diagramName, entityID)
		return nil, NewSerializationError(key, "failed to unmarshal entity as Vertex", err)
	}

	return &vertex, nil
}

// DefaultCleanupOptions returns sensible default cleanup options
func DefaultCleanupOptions() *CleanupOptions {
	return &CleanupOptions{
		BatchSize:      100,
		ScanCount:      100,
		MaxKeys:        0, // No limit
		DryRun:         false,
		Timeout:        5 * time.Minute,
		CollectMetrics: true,
	}
}

// Cleanup removes cache entries matching a pattern with enhanced functionality
func (rc *RedisCache) Cleanup(ctx context.Context, pattern string) error {
	result, err := rc.CleanupWithOptions(ctx, pattern, DefaultCleanupOptions())
	if err != nil {
		return err
	}

	// For backward compatibility, we don't return the result in the basic Cleanup method
	_ = result
	return nil
}

// CleanupWithOptions removes cache entries matching a pattern with configurable options
func (rc *RedisCache) CleanupWithOptions(ctx context.Context, pattern string, options *CleanupOptions) (*CleanupResult, error) {
	// Validate context
	if err := rc.validator.ValidateContext(ctx); err != nil {
		return nil, err
	}

	// Validate and sanitize pattern
	if err := rc.validator.ValidateCleanupPattern(pattern); err != nil {
		return nil, err
	}

	if options == nil {
		options = DefaultCleanupOptions()
	}

	// Validate options using enhanced validation logic
	if err := rc.validateCleanupOptionsEnhanced(options); err != nil {
		return nil, err
	}

	startTime := time.Now()
	result := &CleanupResult{}

	// Create a timeout context if specified
	cleanupCtx := ctx
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		cleanupCtx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// Use Redis SCAN to find keys matching the pattern
	cursor := uint64(0)
	var allKeys []string
	var totalBytesScanned int64

	// Scan for keys in batches
	for {
		// Check for context cancellation
		select {
		case <-cleanupCtx.Done():
			return nil, NewTimeoutError("", "cleanup operation timed out during scan", cleanupCtx.Err())
		default:
		}

		keys, nextCursor, err := rc.client.ScanWithRetry(cleanupCtx, cursor, pattern, options.ScanCount)
		if err != nil {
			result.ErrorsOccurred++
			if isTimeoutError(err) {
				return result, NewTimeoutError("", "timeout during cleanup scan", err)
			}
			if isConnectionError(err) {
				return result, NewConnectionError("failed to scan for cleanup", err)
			}
			return result, fmt.Errorf("failed to scan keys for cleanup: %w", err)
		}

		// Check if we need to limit the keys before processing
		keysToProcess := keys
		if options.MaxKeys > 0 {
			remainingSlots := options.MaxKeys - int64(len(allKeys))
			if remainingSlots <= 0 {
				break // We've already reached the limit
			}
			if int64(len(keys)) > remainingSlots {
				keysToProcess = keys[:remainingSlots]
			}
		}

		result.KeysScanned += int64(len(keysToProcess))
		allKeys = append(allKeys, keysToProcess...)

		// If collecting metrics, estimate bytes scanned
		if options.CollectMetrics {
			for _, key := range keysToProcess {
				totalBytesScanned += int64(len(key)) + 8 // Estimate key overhead
			}
		}

		// Check if we've hit the max keys limit
		if options.MaxKeys > 0 && int64(len(allKeys)) >= options.MaxKeys {
			break
		}

		cursor = nextCursor
		if cursor == 0 {
			break // Scan complete
		}
	}

	// If dry run, just return scan results
	if options.DryRun {
		result.Duration = time.Since(startTime)
		result.KeysDeleted = 0
		result.BytesFreed = totalBytesScanned // Estimate what would be freed
		return result, nil
	}

	// Delete keys in batches for efficiency
	if len(allKeys) > 0 {
		deletedCount, bytesFreed, batchCount, err := rc.deleteKeysInBatches(cleanupCtx, allKeys, options)
		result.KeysDeleted = deletedCount
		result.BytesFreed = bytesFreed
		result.BatchesUsed = batchCount

		if err != nil {
			result.ErrorsOccurred++
			result.Duration = time.Since(startTime)
			return result, err
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// validateCleanupOptions validates cleanup options (legacy method)
func (rc *RedisCache) validateCleanupOptions(options *CleanupOptions) error {
	return rc.validateCleanupOptionsEnhanced(options)
}

// validateCleanupOptionsEnhanced validates cleanup options with enhanced security checks
func (rc *RedisCache) validateCleanupOptionsEnhanced(options *CleanupOptions) error {
	if options == nil {
		return NewValidationError("cleanup options cannot be nil", nil)
	}

	if options.BatchSize <= 0 {
		return NewValidationError(fmt.Sprintf("batch size must be positive, got %d", options.BatchSize), nil)
	}

	if options.BatchSize > 1000 {
		return NewValidationError(fmt.Sprintf("batch size too large (max 1000), got %d", options.BatchSize), nil)
	}

	if options.ScanCount <= 0 {
		return NewValidationError(fmt.Sprintf("scan count must be positive, got %d", options.ScanCount), nil)
	}

	if options.ScanCount > 10000 {
		return NewValidationError(fmt.Sprintf("scan count too large (max 10000), got %d", options.ScanCount), nil)
	}

	if options.MaxKeys < 0 {
		return NewValidationError(fmt.Sprintf("max keys cannot be negative, got %d", options.MaxKeys), nil)
	}

	if options.Timeout < 0 {
		return NewValidationError(fmt.Sprintf("timeout cannot be negative, got %v", options.Timeout), nil)
	}

	// Check for reasonable timeout limits
	maxTimeout := 30 * time.Minute
	if options.Timeout > maxTimeout {
		return NewValidationError(fmt.Sprintf("timeout exceeds maximum allowed duration of %v", options.Timeout), nil)
	}

	return nil
}

// deleteKeysInBatches deletes keys in batches and returns metrics
func (rc *RedisCache) deleteKeysInBatches(ctx context.Context, keys []string, options *CleanupOptions) (int64, int64, int, error) {
	var totalDeleted int64
	var totalBytesFreed int64
	batchCount := 0

	// Process keys in batches
	for i := 0; i < len(keys); i += options.BatchSize {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return totalDeleted, totalBytesFreed, batchCount, NewTimeoutError("", "cleanup operation timed out during deletion", ctx.Err())
		default:
		}

		end := i + options.BatchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		batchCount++

		// If collecting metrics, get key sizes before deletion
		var batchBytesFreed int64
		if options.CollectMetrics {
			for _, key := range batch {
				// Estimate memory usage: key name + value + overhead
				if size, err := rc.client.MemoryUsageWithRetry(ctx, key); err == nil {
					batchBytesFreed += size
				} else {
					// Fallback estimation if MEMORY USAGE fails
					batchBytesFreed += int64(len(key)) + 100 // Rough estimate
				}
			}
		}

		// Delete the batch
		deletedInBatch, err := rc.client.DelBatchWithRetry(ctx, batch...)
		if err != nil {
			if isTimeoutError(err) {
				return totalDeleted, totalBytesFreed, batchCount, NewTimeoutError("", "timeout during batch deletion", err)
			}
			if isConnectionError(err) {
				return totalDeleted, totalBytesFreed, batchCount, NewConnectionError("failed to delete batch during cleanup", err)
			}
			return totalDeleted, totalBytesFreed, batchCount, fmt.Errorf("failed to delete batch during cleanup: %w", err)
		}

		totalDeleted += deletedInBatch
		if options.CollectMetrics {
			// Only count bytes for actually deleted keys
			totalBytesFreed += (batchBytesFreed * deletedInBatch) / int64(len(batch))
		}
	}

	return totalDeleted, totalBytesFreed, batchCount, nil
}

// GetCacheSize returns information about cache size and memory usage
func (rc *RedisCache) GetCacheSize(ctx context.Context) (*CacheSizeInfo, error) {
	info := &CacheSizeInfo{}

	// Get database size (number of keys)
	dbSize, err := rc.client.DBSizeWithRetry(ctx)
	if err != nil {
		if isTimeoutError(err) {
			return nil, NewTimeoutError("", "timeout getting database size", err)
		}
		if isConnectionError(err) {
			return nil, NewConnectionError("failed to get database size", err)
		}
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}
	info.TotalKeys = dbSize

	// Get memory usage information
	memInfo, err := rc.client.InfoWithRetry(ctx, "memory")
	if err != nil {
		if isTimeoutError(err) {
			return nil, NewTimeoutError("", "timeout getting memory info", err)
		}
		if isConnectionError(err) {
			return nil, NewConnectionError("failed to get memory info", err)
		}
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	// Parse memory information
	memoryMetrics, err := rc.parseMemoryInfo(memInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse memory info: %w", err)
	}
	info.MemoryUsed = memoryMetrics.UsedMemory
	info.MemoryPeak = memoryMetrics.UsedMemoryPeak
	info.MemoryOverhead = 0 // Not available in new format

	// Get cache-specific statistics by scanning for our key patterns
	info.DiagramCount = rc.countKeysByPattern(ctx, "/diagrams/puml/*")
	info.StateMachineCount = rc.countKeysByPattern(ctx, "/machines/*")
	info.EntityCount = rc.countKeysByPattern(ctx, "/machines/*/entities/*")

	return info, nil
}

// countKeysByPattern counts keys matching a specific pattern
func (rc *RedisCache) countKeysByPattern(ctx context.Context, pattern string) int64 {
	var count int64
	cursor := uint64(0)

	for {
		keys, nextCursor, err := rc.client.ScanWithRetry(ctx, cursor, pattern, 100)
		if err != nil {
			return count // Return partial count on error
		}

		count += int64(len(keys))
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return count
}

// splitString splits a string by delimiter (simple implementation)
func splitString(s, delimiter string) []string {
	if s == "" {
		return []string{}
	}

	var result []string
	start := 0
	delimLen := len(delimiter)

	for i := 0; i <= len(s)-delimLen; i++ {
		if s[i:i+delimLen] == delimiter {
			result = append(result, s[start:i])
			start = i + delimLen
			i += delimLen - 1 // Skip the delimiter
		}
	}

	// Add the last part (even if it's empty)
	result = append(result, s[start:])

	return result
}

// Health performs a basic health check on the cache
func (rc *RedisCache) Health(ctx context.Context) error {
	return rc.client.HealthWithRetry(ctx)
}

// HealthDetailed performs a comprehensive health check and returns detailed status information
func (rc *RedisCache) HealthDetailed(ctx context.Context) (*HealthStatus, error) {
	startTime := time.Now()

	healthStatus := &HealthStatus{
		Timestamp: startTime,
		Status:    "healthy",
		Errors:    []string{},
		Warnings:  []string{},
	}

	// Check basic connectivity
	connectionHealth, err := rc.GetConnectionHealth(ctx)
	if err != nil {
		healthStatus.Errors = append(healthStatus.Errors, fmt.Sprintf("Connection health check failed: %v", err))
		healthStatus.Status = "unhealthy"
	} else if !connectionHealth.Connected {
		healthStatus.Errors = append(healthStatus.Errors, fmt.Sprintf("Redis connection failed: %s", connectionHealth.LastError))
		healthStatus.Status = "unhealthy"
	}
	healthStatus.Connection = *connectionHealth

	// Get performance metrics
	performanceMetrics, err := rc.GetPerformanceMetrics(ctx)
	if err != nil {
		healthStatus.Warnings = append(healthStatus.Warnings, fmt.Sprintf("Performance metrics unavailable: %v", err))
		if healthStatus.Status == "healthy" {
			healthStatus.Status = "degraded"
		}
	} else {
		healthStatus.Performance = *performanceMetrics

		// Check for performance issues
		if performanceMetrics.MemoryUsage.MemoryFragmentation > 1.5 {
			healthStatus.Warnings = append(healthStatus.Warnings, "High memory fragmentation detected")
			if healthStatus.Status == "healthy" {
				healthStatus.Status = "degraded"
			}
		}

		if performanceMetrics.KeyspaceInfo.HitRate < 80.0 {
			healthStatus.Warnings = append(healthStatus.Warnings, "Low cache hit rate detected")
			if healthStatus.Status == "healthy" {
				healthStatus.Status = "degraded"
			}
		}
	}

	// Run diagnostics
	diagnostics, err := rc.RunDiagnostics(ctx)
	if err != nil {
		healthStatus.Warnings = append(healthStatus.Warnings, fmt.Sprintf("Diagnostics failed: %v", err))
		if healthStatus.Status == "healthy" {
			healthStatus.Status = "degraded"
		}
	} else {
		healthStatus.Diagnostics = *diagnostics

		// Check diagnostic results
		if !diagnostics.ConfigurationCheck.Valid {
			healthStatus.Status = "unhealthy"
			healthStatus.Errors = append(healthStatus.Errors, "Configuration validation failed")
		}

		if !diagnostics.NetworkCheck.Reachable {
			healthStatus.Status = "unhealthy"
			healthStatus.Errors = append(healthStatus.Errors, "Network connectivity issues detected")
		}

		if !diagnostics.PerformanceCheck.Acceptable {
			if healthStatus.Status == "healthy" {
				healthStatus.Status = "degraded"
			}
			healthStatus.Warnings = append(healthStatus.Warnings, "Performance issues detected")
		}

		if !diagnostics.DataIntegrityCheck.Consistent {
			if healthStatus.Status == "healthy" {
				healthStatus.Status = "degraded"
			}
			healthStatus.Warnings = append(healthStatus.Warnings, "Data integrity issues detected")
		}
	}

	healthStatus.ResponseTime = time.Since(startTime)
	return healthStatus, nil
}

// GetConnectionHealth returns detailed connection health information
func (rc *RedisCache) GetConnectionHealth(ctx context.Context) (*ConnectionHealth, error) {
	connectionHealth := &ConnectionHealth{
		Address:  rc.config.RedisAddr,
		Database: rc.config.RedisDB,
	}

	// Test basic connectivity with ping
	pingStart := time.Now()
	err := rc.client.HealthWithRetry(ctx)
	pingLatency := time.Since(pingStart)

	if err != nil {
		connectionHealth.Connected = false
		connectionHealth.LastError = err.Error()
		return connectionHealth, nil
	}

	connectionHealth.Connected = true
	connectionHealth.PingLatency = pingLatency

	// Get connection pool statistics
	client := rc.client.Client()
	if client != nil {
		poolStats := client.PoolStats()
		connectionHealth.PoolStats = ConnectionPoolStats{
			TotalConnections:  int(poolStats.TotalConns),
			IdleConnections:   int(poolStats.IdleConns),
			ActiveConnections: int(poolStats.TotalConns - poolStats.IdleConns),
			Hits:              int(poolStats.Hits),
			Misses:            int(poolStats.Misses),
			Timeouts:          int(poolStats.Timeouts),
			StaleConnections:  int(poolStats.StaleConns),
		}
	}

	return connectionHealth, nil
}

// GetPerformanceMetrics returns detailed performance metrics
func (rc *RedisCache) GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error) {
	metrics := &PerformanceMetrics{}

	// Get memory information
	memoryInfo, err := rc.client.InfoWithRetry(ctx, "memory")
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	memoryMetrics, err := rc.parseMemoryInfo(memoryInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse memory info: %w", err)
	}
	metrics.MemoryUsage = *memoryMetrics

	// Get keyspace information
	keyspaceInfo, err := rc.client.InfoWithRetry(ctx, "keyspace")
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace info: %w", err)
	}

	keyspaceMetrics, err := rc.parseKeyspaceInfo(keyspaceInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse keyspace info: %w", err)
	}
	metrics.KeyspaceInfo = *keyspaceMetrics

	// Get stats information
	statsInfo, err := rc.client.InfoWithRetry(ctx, "stats")
	if err != nil {
		return nil, fmt.Errorf("failed to get stats info: %w", err)
	}

	operationMetrics, err := rc.parseStatsInfo(statsInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse stats info: %w", err)
	}
	metrics.OperationStats = *operationMetrics

	// Parse keyspace hits/misses from stats info and update keyspace metrics
	rc.parseKeyspaceStatsFromInfo(statsInfo, &metrics.KeyspaceInfo)

	// Get server information
	serverInfo, err := rc.client.InfoWithRetry(ctx, "server")
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	serverMetrics, err := rc.parseServerInfo(serverInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server info: %w", err)
	}
	metrics.ServerInfo = *serverMetrics

	return metrics, nil
}

// RunDiagnostics performs comprehensive diagnostic checks
func (rc *RedisCache) RunDiagnostics(ctx context.Context) (*DiagnosticInfo, error) {
	diagnostics := &DiagnosticInfo{}

	// Configuration check
	configCheck := rc.validateConfiguration()
	diagnostics.ConfigurationCheck = configCheck

	// Network check
	networkCheck, err := rc.performNetworkCheck(ctx)
	if err != nil {
		return nil, fmt.Errorf("network check failed: %w", err)
	}
	diagnostics.NetworkCheck = *networkCheck

	// Performance check
	performanceCheck, err := rc.performPerformanceCheck(ctx)
	if err != nil {
		return nil, fmt.Errorf("performance check failed: %w", err)
	}
	diagnostics.PerformanceCheck = *performanceCheck

	// Data integrity check
	dataIntegrityCheck, err := rc.performDataIntegrityCheck(ctx)
	if err != nil {
		return nil, fmt.Errorf("data integrity check failed: %w", err)
	}
	diagnostics.DataIntegrityCheck = *dataIntegrityCheck

	// Generate recommendations
	diagnostics.Recommendations = rc.generateRecommendations(diagnostics)

	return diagnostics, nil
}

// Close closes the cache connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// Helper functions to identify error types

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	// Check for context timeout or Redis timeout errors
	return err == context.DeadlineExceeded ||
		err == context.Canceled ||
		contains(err.Error(), "timeout") ||
		contains(err.Error(), "i/o timeout")
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := err.Error()
	return contains(errorStr, "connection refused") ||
		contains(errorStr, "connection reset") ||
		contains(errorStr, "network is unreachable") ||
		contains(errorStr, "no route to host") ||
		contains(errorStr, "broken pipe")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				func() bool {
					for i := 0; i <= len(s)-len(substr); i++ {
						match := true
						for j := 0; j < len(substr); j++ {
							if s[i+j] != substr[j] &&
								(s[i+j] < 'A' || s[i+j] > 'Z' || s[i+j]+32 != substr[j]) &&
								(s[i+j] < 'a' || s[i+j] > 'z' || s[i+j]-32 != substr[j]) {
								match = false
								break
							}
						}
						if match {
							return true
						}
					}
					return false
				}()))
}

// parseMemoryInfo parses Redis memory information from INFO memory output
func (rc *RedisCache) parseMemoryInfo(info string) (*MemoryMetrics, error) {
	metrics := &MemoryMetrics{}

	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "used_memory":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				metrics.UsedMemory = val
			}
		case "used_memory_human":
			metrics.UsedMemoryHuman = value
		case "used_memory_peak":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				metrics.UsedMemoryPeak = val
			}
		case "mem_fragmentation_ratio":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				metrics.MemoryFragmentation = val
			}
		case "maxmemory":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				metrics.MaxMemory = val
			}
		case "maxmemory_policy":
			metrics.MaxMemoryPolicy = value
		}
	}

	return metrics, nil
}

// parseKeyspaceInfo parses Redis keyspace information from INFO keyspace output
func (rc *RedisCache) parseKeyspaceInfo(info string) (*KeyspaceMetrics, error) {
	metrics := &KeyspaceMetrics{}

	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Parse db0:keys=X,expires=Y,avg_ttl=Z format
		if strings.HasPrefix(line, fmt.Sprintf("db%d:", rc.config.RedisDB)) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				dbInfo := parts[1]
				dbParts := strings.Split(dbInfo, ",")

				for _, part := range dbParts {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) == 2 {
						key := strings.TrimSpace(kv[0])
						value := strings.TrimSpace(kv[1])

						switch key {
						case "keys":
							if val, err := strconv.ParseInt(value, 10, 64); err == nil {
								metrics.TotalKeys = val
							}
						case "expires":
							if val, err := strconv.ParseInt(value, 10, 64); err == nil {
								metrics.ExpiringKeys = val
							}
						}
					}
				}
			}
		}
	}

	// Calculate average key size (estimate)
	if metrics.TotalKeys > 0 {
		// This is a rough estimate - in a real implementation you might want to sample keys
		metrics.AverageKeySize = 50.0 // Default estimate
	}

	return metrics, nil
}

// parseStatsInfo parses Redis stats information from INFO stats output
func (rc *RedisCache) parseStatsInfo(info string) (*OperationMetrics, error) {
	metrics := &OperationMetrics{}

	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "total_commands_processed":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				metrics.TotalCommands = val
			}
		case "instantaneous_ops_per_sec":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				metrics.CommandsPerSecond = val
			}
		case "keyspace_hits":
			// Skip keyspace hits - will be handled in keyspace info parsing
		case "keyspace_misses":
			// Skip keyspace misses - will be handled in keyspace info parsing
		case "rejected_connections":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				metrics.RejectedConnections = val
			}
		}
	}

	return metrics, nil
}

// parseServerInfo parses Redis server information from INFO server output
func (rc *RedisCache) parseServerInfo(info string) (*ServerMetrics, error) {
	metrics := &ServerMetrics{}

	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "redis_version":
			metrics.RedisVersion = value
		case "uptime_in_seconds":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				metrics.UptimeSeconds = val
				metrics.UptimeDays = val / 86400 // Convert seconds to days
			}
		case "redis_mode":
			metrics.ServerMode = value
		case "role":
			metrics.Role = value
		case "connected_slaves":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				metrics.ConnectedSlaves = val
			}
		}
	}

	return metrics, nil
}

// validateConfiguration performs configuration validation
func (rc *RedisCache) validateConfiguration() ConfigCheck {
	check := ConfigCheck{
		Valid:           true,
		Issues:          []string{},
		Recommendations: []string{},
	}

	// Check connection timeouts
	if rc.config.DialTimeout < time.Second {
		check.Issues = append(check.Issues, "Dial timeout is very low (< 1s)")
		check.Recommendations = append(check.Recommendations, "Consider increasing dial timeout to at least 1 second")
	}

	if rc.config.ReadTimeout < 500*time.Millisecond {
		check.Issues = append(check.Issues, "Read timeout is very low (< 500ms)")
		check.Recommendations = append(check.Recommendations, "Consider increasing read timeout to at least 500ms")
	}

	if rc.config.WriteTimeout < 500*time.Millisecond {
		check.Issues = append(check.Issues, "Write timeout is very low (< 500ms)")
		check.Recommendations = append(check.Recommendations, "Consider increasing write timeout to at least 500ms")
	}

	// Check pool size
	if rc.config.PoolSize < 5 {
		check.Recommendations = append(check.Recommendations, "Consider increasing pool size for better concurrency")
	}

	if rc.config.PoolSize > 100 {
		check.Recommendations = append(check.Recommendations, "Pool size is very high, consider reducing to save resources")
	}

	// Check TTL
	if rc.config.DefaultTTL < time.Hour {
		check.Recommendations = append(check.Recommendations, "Default TTL is very low, consider increasing for better cache efficiency")
	}

	// Check retry configuration
	if rc.config.RetryConfig != nil {
		if rc.config.RetryConfig.MaxAttempts > 10 {
			check.Issues = append(check.Issues, "Max retry attempts is very high")
			check.Recommendations = append(check.Recommendations, "Consider reducing max retry attempts to avoid long delays")
		}

		if rc.config.RetryConfig.MaxDelay > 30*time.Second {
			check.Issues = append(check.Issues, "Max retry delay is very high")
			check.Recommendations = append(check.Recommendations, "Consider reducing max retry delay")
		}
	}

	if len(check.Issues) > 0 {
		check.Valid = false
	}

	return check
}

// performNetworkCheck performs network connectivity validation
func (rc *RedisCache) performNetworkCheck(ctx context.Context) (*NetworkCheck, error) {
	check := &NetworkCheck{
		Issues: []string{},
	}

	// Test basic connectivity with multiple pings
	var latencies []time.Duration
	var failures int

	for i := 0; i < 5; i++ {
		start := time.Now()
		err := rc.client.HealthWithRetry(ctx)
		latency := time.Since(start)

		if err != nil {
			failures++
			check.Issues = append(check.Issues, fmt.Sprintf("Ping %d failed: %v", i+1, err))
		} else {
			latencies = append(latencies, latency)
		}
	}

	// Calculate average latency
	if len(latencies) > 0 {
		var totalLatency time.Duration
		for _, lat := range latencies {
			totalLatency += lat
		}
		check.Latency = totalLatency / time.Duration(len(latencies))
		check.Reachable = true
	} else {
		check.Reachable = false
	}

	// Calculate packet loss
	check.PacketLoss = (float64(failures) / 5.0) * 100.0

	// Add warnings for high latency or packet loss
	if check.Latency > 100*time.Millisecond {
		check.Issues = append(check.Issues, "High network latency detected")
	}

	if check.PacketLoss > 0 {
		check.Issues = append(check.Issues, fmt.Sprintf("Packet loss detected: %.1f%%", check.PacketLoss))
	}

	return check, nil
}

// performPerformanceCheck performs performance validation
func (rc *RedisCache) performPerformanceCheck(ctx context.Context) (*PerformanceCheck, error) {
	check := &PerformanceCheck{
		Acceptable:      true,
		Bottlenecks:     []string{},
		Recommendations: []string{},
	}

	// Get performance metrics for analysis
	metrics, err := rc.GetPerformanceMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance metrics: %w", err)
	}

	// Check memory fragmentation
	if metrics.MemoryUsage.MemoryFragmentation > 2.0 {
		check.Acceptable = false
		check.Bottlenecks = append(check.Bottlenecks, "High memory fragmentation")
		check.Recommendations = append(check.Recommendations, "Consider restarting Redis or using MEMORY PURGE command")
	} else if metrics.MemoryUsage.MemoryFragmentation > 1.5 {
		check.Bottlenecks = append(check.Bottlenecks, "Moderate memory fragmentation")
		check.Recommendations = append(check.Recommendations, "Monitor memory fragmentation and consider cleanup")
	}

	// Check hit rate
	if metrics.KeyspaceInfo.HitRate < 70.0 {
		check.Acceptable = false
		check.Bottlenecks = append(check.Bottlenecks, "Low cache hit rate")
		check.Recommendations = append(check.Recommendations, "Review caching strategy and TTL settings")
	} else if metrics.KeyspaceInfo.HitRate < 85.0 {
		check.Recommendations = append(check.Recommendations, "Cache hit rate could be improved")
	}

	// Check memory usage
	if metrics.MemoryUsage.MaxMemory > 0 {
		usagePercent := (float64(metrics.MemoryUsage.UsedMemory) / float64(metrics.MemoryUsage.MaxMemory)) * 100.0
		if usagePercent > 90.0 {
			check.Acceptable = false
			check.Bottlenecks = append(check.Bottlenecks, "High memory usage")
			check.Recommendations = append(check.Recommendations, "Consider increasing memory limit or implementing cleanup")
		} else if usagePercent > 80.0 {
			check.Recommendations = append(check.Recommendations, "Memory usage is getting high, monitor closely")
		}
	}

	// Check connection pool utilization
	connectionHealth, err := rc.GetConnectionHealth(ctx)
	if err == nil {
		if connectionHealth.PoolStats.TotalConnections > 0 {
			utilizationPercent := (float64(connectionHealth.PoolStats.ActiveConnections) / float64(connectionHealth.PoolStats.TotalConnections)) * 100.0
			if utilizationPercent > 90.0 {
				check.Bottlenecks = append(check.Bottlenecks, "High connection pool utilization")
				check.Recommendations = append(check.Recommendations, "Consider increasing pool size")
			}
		}

		if connectionHealth.PoolStats.Timeouts > 0 {
			check.Bottlenecks = append(check.Bottlenecks, "Connection pool timeouts detected")
			check.Recommendations = append(check.Recommendations, "Consider increasing pool size or connection timeouts")
		}
	}

	return check, nil
}

// performDataIntegrityCheck performs data integrity validation
func (rc *RedisCache) performDataIntegrityCheck(ctx context.Context) (*DataIntegrityCheck, error) {
	check := &DataIntegrityCheck{
		Consistent:   true,
		Issues:       []string{},
		OrphanedKeys: 0,
	}

	// Check for orphaned entity keys
	// Scan for all entity keys and verify they have corresponding state machines
	entityPattern := "/machines/*/entities/*"

	client := rc.client.Client()
	if client == nil {
		return check, nil
	}

	iter := client.Scan(ctx, 0, entityPattern, 100).Iterator()
	var entityKeys []string

	for iter.Next(ctx) {
		entityKeys = append(entityKeys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan for entity keys: %w", err)
	}

	// Check each entity key to see if its parent state machine exists
	for _, entityKey := range entityKeys {
		// Extract state machine key from entity key
		// Entity key format: /machines/<version>/<diagram>/entities/<entityID>
		parts := strings.Split(entityKey, "/")
		if len(parts) >= 5 {
			stateMachineKey := fmt.Sprintf("/machines/%s/%s", parts[2], parts[3])

			// Check if state machine exists
			exists, err := rc.client.ExistsWithRetry(ctx, stateMachineKey)
			if err != nil {
				check.Issues = append(check.Issues, fmt.Sprintf("Failed to check state machine for entity %s: %v", entityKey, err))
				check.Consistent = false
				continue
			}

			if exists == 0 {
				check.OrphanedKeys++
				check.Issues = append(check.Issues, fmt.Sprintf("Orphaned entity key found: %s", entityKey))
				check.Consistent = false
			}
		}
	}

	// Additional integrity checks could be added here:
	// - Verify JSON structure of cached objects
	// - Check for expired keys that should have been cleaned up
	// - Validate referential integrity between diagrams and state machines

	return check, nil
}

// generateRecommendations generates performance and configuration recommendations
func (rc *RedisCache) generateRecommendations(diagnostics *DiagnosticInfo) []string {
	var recommendations []string

	// Collect all recommendations from sub-checks
	recommendations = append(recommendations, diagnostics.ConfigurationCheck.Recommendations...)
	recommendations = append(recommendations, diagnostics.PerformanceCheck.Recommendations...)

	// Add general recommendations based on overall health
	if !diagnostics.NetworkCheck.Reachable {
		recommendations = append(recommendations, "Check network connectivity to Redis server")
		recommendations = append(recommendations, "Verify Redis server is running and accessible")
	}

	if diagnostics.DataIntegrityCheck.OrphanedKeys > 0 {
		recommendations = append(recommendations, "Run cleanup operation to remove orphaned keys")
		recommendations = append(recommendations, "Review cache cleanup policies")
	}

	if len(diagnostics.ConfigurationCheck.Issues) > 0 {
		recommendations = append(recommendations, "Review and update Redis configuration")
	}

	if len(diagnostics.PerformanceCheck.Bottlenecks) > 0 {
		recommendations = append(recommendations, "Address identified performance bottlenecks")
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueRecommendations []string
	for _, rec := range recommendations {
		if !seen[rec] {
			seen[rec] = true
			uniqueRecommendations = append(uniqueRecommendations, rec)
		}
	}

	// Ensure we always return a non-nil slice
	if uniqueRecommendations == nil {
		return []string{}
	}
	return uniqueRecommendations
}

// parseKeyspaceStatsFromInfo parses keyspace statistics from Redis stats info
func (rc *RedisCache) parseKeyspaceStatsFromInfo(info string, metrics *KeyspaceMetrics) {
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "keyspace_hits":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				metrics.KeyspaceHits = val
			}
		case "keyspace_misses":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				metrics.KeyspaceMisses = val
			}
		}
	}

	// Calculate hit rate
	totalRequests := metrics.KeyspaceHits + metrics.KeyspaceMisses
	if totalRequests > 0 {
		metrics.HitRate = (float64(metrics.KeyspaceHits) / float64(totalRequests)) * 100.0
	}
}
