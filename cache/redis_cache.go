package cache

import (
	"context"
	"encoding/json"
	"fmt"
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

func DefaultRedisConfig() *RedisConfig {
	return internal.DefaultConfig()
}

func DefaultRedisRetryConfig() *RedisRetryConfig {
	return internal.DefaultRetryConfig()
}

// RedisCache implements the Cache interface using Redis as the backend
type RedisCache struct {
	client internal.RedisClientInterface
	keyGen internal.KeyGenerator
	config *RedisConfig
}

// NewRedisCache creates a new Redis-backed cache implementation
func NewRedisCache(config *RedisConfig) (*RedisCache, error) {
	client, err := internal.NewRedisClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	return &RedisCache{
		client: client,
		keyGen: internal.NewKeyGenerator(),
		config: config,
	}, nil
}

// NewRedisCacheWithDependencies creates a new Redis cache with injected dependencies for testing
func NewRedisCacheWithDependencies(client internal.RedisClientInterface, keyGen internal.KeyGenerator, config *RedisConfig) *RedisCache {
	return &RedisCache{
		client: client,
		keyGen: keyGen,
		config: config,
	}
}

// StoreDiagram stores a PlantUML diagram with TTL support
func (rc *RedisCache) StoreDiagram(ctx context.Context, name string, pumlContent string, ttl time.Duration) error {
	if name == "" {
		return NewValidationError("diagram name cannot be empty", nil)
	}

	if pumlContent == "" {
		return NewValidationError("diagram content cannot be empty", nil)
	}

	// Generate cache key
	key := rc.keyGen.DiagramKey(name)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Use default TTL if not specified
	if ttl <= 0 {
		ttl = rc.config.DefaultTTL
	}

	// Store the diagram content in Redis
	err := rc.client.SetWithRetry(ctx, key, pumlContent, ttl)
	if err != nil {
		if isTimeoutError(err) {
			return NewTimeoutError(key, "timeout storing diagram", err)
		}
		if isConnectionError(err) {
			return NewConnectionError("failed to store diagram", err)
		}
		return fmt.Errorf("failed to store diagram '%s': %w", name, err)
	}

	return nil
}

// GetDiagram retrieves a PlantUML diagram with error handling
func (rc *RedisCache) GetDiagram(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", NewValidationError("diagram name cannot be empty", nil)
	}

	// Generate cache key
	key := rc.keyGen.DiagramKey(name)

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
		return "", fmt.Errorf("failed to retrieve diagram '%s': %w", name, err)
	}

	return content, nil
}

// DeleteDiagram removes a diagram from the cache for cleanup
func (rc *RedisCache) DeleteDiagram(ctx context.Context, name string) error {
	if name == "" {
		return NewValidationError("diagram name cannot be empty", nil)
	}

	// Generate cache key
	key := rc.keyGen.DiagramKey(name)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Delete the diagram from Redis
	err := rc.client.DelWithRetry(ctx, key)
	if err != nil {
		if isTimeoutError(err) {
			return NewTimeoutError(key, "timeout deleting diagram", err)
		}
		if isConnectionError(err) {
			return NewConnectionError("failed to delete diagram", err)
		}
		return fmt.Errorf("failed to delete diagram '%s': %w", name, err)
	}

	return nil
}

// StoreStateMachine stores a parsed state machine with TTL support and creates entity cache paths
func (rc *RedisCache) StoreStateMachine(ctx context.Context, umlVersion, name string, machine *models.StateMachine, ttl time.Duration) error {
	if umlVersion == "" {
		return NewValidationError("UML version cannot be empty", nil)
	}

	if name == "" {
		return NewValidationError("state machine name cannot be empty", nil)
	}

	if machine == nil {
		return NewValidationError("state machine cannot be nil", nil)
	}

	// Validate that the corresponding diagram exists before storing state machine
	_, err := rc.GetDiagram(ctx, name)
	if err != nil {
		if IsNotFoundError(err) {
			return NewValidationError(fmt.Sprintf("cannot store state machine: corresponding diagram '%s' does not exist", name), err)
		}
		return fmt.Errorf("failed to validate diagram existence for state machine '%s': %w", name, err)
	}

	// Generate cache key
	key := rc.keyGen.StateMachineKey(umlVersion, name)

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
		entityKey := rc.keyGen.EntityKey(umlVersion, name, entityID)

		// Validate entity key before storing
		if err := rc.keyGen.ValidateKey(entityKey); err != nil {
			return NewKeyInvalidError(entityKey, fmt.Sprintf("invalid entity key generated for '%s': %v", entityID, err))
		}

		// Update the Entities mapping for referential integrity
		machine.Entities[entityID] = entityKey

		// Store the entity
		err := rc.StoreEntity(ctx, umlVersion, name, entityID, entity, ttl)
		if err != nil {
			// If entity storage fails, clean up any previously stored entities to maintain consistency
			rc.cleanupPartialEntityStorage(ctx, umlVersion, name, machine.Entities)
			return fmt.Errorf("failed to store entity '%s': %w", entityID, err)
		}
	}

	// Serialize the state machine to JSON (now with populated Entities map)
	data, err := json.Marshal(machine)
	if err != nil {
		// Clean up entities if state machine serialization fails
		rc.cleanupPartialEntityStorage(ctx, umlVersion, name, machine.Entities)
		return NewSerializationError(key, "failed to marshal state machine", err)
	}

	// Store the serialized state machine in Redis
	err = rc.client.SetWithRetry(ctx, key, data, ttl)
	if err != nil {
		// Clean up entities if state machine storage fails
		rc.cleanupPartialEntityStorage(ctx, umlVersion, name, machine.Entities)

		if isTimeoutError(err) {
			return NewTimeoutError(key, "timeout storing state machine", err)
		}
		if isConnectionError(err) {
			return NewConnectionError("failed to store state machine", err)
		}
		return fmt.Errorf("failed to store state machine '%s': %w", name, err)
	}

	return nil
}

// GetStateMachine retrieves a parsed state machine
func (rc *RedisCache) GetStateMachine(ctx context.Context, umlVersion, name string) (*models.StateMachine, error) {
	if umlVersion == "" {
		return nil, NewValidationError("UML version cannot be empty", nil)
	}

	if name == "" {
		return nil, NewValidationError("state machine name cannot be empty", nil)
	}

	// Generate cache key
	key := rc.keyGen.StateMachineKey(umlVersion, name)

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
		return nil, fmt.Errorf("failed to retrieve state machine '%s': %w", name, err)
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
	if umlVersion == "" {
		return NewValidationError("UML version cannot be empty", nil)
	}

	if name == "" {
		return NewValidationError("state machine name cannot be empty", nil)
	}

	// Generate cache key for the state machine
	key := rc.keyGen.StateMachineKey(umlVersion, name)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// First, try to retrieve the state machine to get entity information for cascade deletion
	machine, err := rc.GetStateMachine(ctx, umlVersion, name)
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
			entityKey := rc.keyGen.EntityKey(umlVersion, name, entityID)
			keysToDelete = append(keysToDelete, entityKey)
		}
	}

	// Additional safety check: scan for any orphaned entity keys that might exist
	// but aren't tracked in the Entities map to ensure complete cleanup
	entityPattern := rc.keyGen.EntityKey(umlVersion, name, "*")
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
			return fmt.Errorf("failed to delete state machine '%s' and its entities: %w", name, err)
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
	if umlVersion == "" {
		return NewValidationError("UML version cannot be empty", nil)
	}

	if diagramName == "" {
		return NewValidationError("diagram name cannot be empty", nil)
	}

	if entityID == "" {
		return NewValidationError("entity ID cannot be empty", nil)
	}

	if entity == nil {
		return NewValidationError("entity cannot be nil", nil)
	}

	// Generate cache key
	key := rc.keyGen.EntityKey(umlVersion, diagramName, entityID)

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
		return fmt.Errorf("failed to store entity '%s': %w", entityID, err)
	}

	return nil
}

// UpdateStateMachineEntityMapping updates the Entities mapping in a stored state machine
// This ensures referential integrity when entities are added or removed independently
func (rc *RedisCache) UpdateStateMachineEntityMapping(ctx context.Context, umlVersion, name string, entityID, entityKey string, operation string) error {
	if umlVersion == "" {
		return NewValidationError("UML version cannot be empty", nil)
	}

	if name == "" {
		return NewValidationError("state machine name cannot be empty", nil)
	}

	if entityID == "" {
		return NewValidationError("entity ID cannot be empty", nil)
	}

	if operation != "add" && operation != "remove" {
		return NewValidationError("operation must be 'add' or 'remove'", nil)
	}

	if operation == "add" && entityKey == "" {
		return NewValidationError("entity key cannot be empty for add operation", nil)
	}

	// Get the current state machine
	machine, err := rc.GetStateMachine(ctx, umlVersion, name)
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
		machine.Entities[entityID] = entityKey
	case "remove":
		delete(machine.Entities, entityID)
	}

	// Store the updated state machine
	key := rc.keyGen.StateMachineKey(umlVersion, name)
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
	if umlVersion == "" {
		return nil, NewValidationError("UML version cannot be empty", nil)
	}

	if diagramName == "" {
		return nil, NewValidationError("diagram name cannot be empty", nil)
	}

	if entityID == "" {
		return nil, NewValidationError("entity ID cannot be empty", nil)
	}

	// Generate cache key
	key := rc.keyGen.EntityKey(umlVersion, diagramName, entityID)

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
		return nil, fmt.Errorf("failed to retrieve entity '%s': %w", entityID, err)
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
	if pattern == "" {
		return nil, NewValidationError("cleanup pattern cannot be empty", nil)
	}

	if options == nil {
		options = DefaultCleanupOptions()
	}

	// Validate options
	if err := rc.validateCleanupOptions(options); err != nil {
		return nil, NewValidationError(fmt.Sprintf("invalid cleanup options: %v", err), err)
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

// validateCleanupOptions validates cleanup options
func (rc *RedisCache) validateCleanupOptions(options *CleanupOptions) error {
	if options.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive, got %d", options.BatchSize)
	}
	if options.BatchSize > 1000 {
		return fmt.Errorf("batch size too large (max 1000), got %d", options.BatchSize)
	}
	if options.ScanCount <= 0 {
		return fmt.Errorf("scan count must be positive, got %d", options.ScanCount)
	}
	if options.MaxKeys < 0 {
		return fmt.Errorf("max keys cannot be negative, got %d", options.MaxKeys)
	}
	if options.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative, got %v", options.Timeout)
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
	info.MemoryUsed = rc.parseMemoryInfo(memInfo, "used_memory:")
	info.MemoryPeak = rc.parseMemoryInfo(memInfo, "used_memory_peak:")
	info.MemoryOverhead = rc.parseMemoryInfo(memInfo, "used_memory_overhead:")

	// Get cache-specific statistics by scanning for our key patterns
	info.DiagramCount = rc.countKeysByPattern(ctx, "/diagrams/puml/*")
	info.StateMachineCount = rc.countKeysByPattern(ctx, "/machines/*")
	info.EntityCount = rc.countKeysByPattern(ctx, "/machines/*/entities/*")

	return info, nil
}

// parseMemoryInfo extracts memory values from Redis INFO output
func (rc *RedisCache) parseMemoryInfo(info, key string) int64 {
	lines := splitString(info, "\r\n")
	for _, line := range lines {
		if len(line) > len(key) && line[:len(key)] == key {
			valueStr := line[len(key):]
			// Parse the numeric value
			var value int64
			for _, char := range valueStr {
				if char >= '0' && char <= '9' {
					value = value*10 + int64(char-'0')
				} else {
					break
				}
			}
			return value
		}
	}
	return 0
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

// Health performs a health check on the cache
func (rc *RedisCache) Health(ctx context.Context) error {
	return rc.client.HealthWithRetry(ctx)
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
