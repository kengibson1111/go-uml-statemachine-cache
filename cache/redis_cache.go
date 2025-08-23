package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/redis/go-redis/v9"

	"github.com/kengibson1111/go-uml-statemachine-cache/internal"
)

// RedisCache implements the Cache interface using Redis as the backend
type RedisCache struct {
	client internal.RedisClientInterface
	keyGen internal.KeyGenerator
	config *internal.Config
}

// NewRedisCache creates a new Redis-backed cache implementation
func NewRedisCache(config *internal.Config) (*RedisCache, error) {
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
func NewRedisCacheWithDependencies(client internal.RedisClientInterface, keyGen internal.KeyGenerator, config *internal.Config) *RedisCache {
	return &RedisCache{
		client: client,
		keyGen: keyGen,
		config: config,
	}
}

// StoreDiagram stores a PlantUML diagram with TTL support
func (rc *RedisCache) StoreDiagram(ctx context.Context, name string, pumlContent string, ttl time.Duration) error {
	if name == "" {
		return internal.NewValidationError("diagram name cannot be empty", nil)
	}

	if pumlContent == "" {
		return internal.NewValidationError("diagram content cannot be empty", nil)
	}

	// Generate cache key
	key := rc.keyGen.DiagramKey(name)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return internal.NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Use default TTL if not specified
	if ttl <= 0 {
		ttl = rc.config.DefaultTTL
	}

	// Store the diagram content in Redis
	err := rc.client.SetWithRetry(ctx, key, pumlContent, ttl)
	if err != nil {
		if isTimeoutError(err) {
			return internal.NewTimeoutError(key, "timeout storing diagram", err)
		}
		if isConnectionError(err) {
			return internal.NewConnectionError("failed to store diagram", err)
		}
		return fmt.Errorf("failed to store diagram '%s': %w", name, err)
	}

	return nil
}

// GetDiagram retrieves a PlantUML diagram with error handling
func (rc *RedisCache) GetDiagram(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", internal.NewValidationError("diagram name cannot be empty", nil)
	}

	// Generate cache key
	key := rc.keyGen.DiagramKey(name)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return "", internal.NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Retrieve the diagram content from Redis
	content, err := rc.client.GetWithRetry(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return "", internal.NewNotFoundError(key)
		}
		if isTimeoutError(err) {
			return "", internal.NewTimeoutError(key, "timeout retrieving diagram", err)
		}
		if isConnectionError(err) {
			return "", internal.NewConnectionError("failed to retrieve diagram", err)
		}
		return "", fmt.Errorf("failed to retrieve diagram '%s': %w", name, err)
	}

	return content, nil
}

// DeleteDiagram removes a diagram from the cache for cleanup
func (rc *RedisCache) DeleteDiagram(ctx context.Context, name string) error {
	if name == "" {
		return internal.NewValidationError("diagram name cannot be empty", nil)
	}

	// Generate cache key
	key := rc.keyGen.DiagramKey(name)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return internal.NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Delete the diagram from Redis
	err := rc.client.DelWithRetry(ctx, key)
	if err != nil {
		if isTimeoutError(err) {
			return internal.NewTimeoutError(key, "timeout deleting diagram", err)
		}
		if isConnectionError(err) {
			return internal.NewConnectionError("failed to delete diagram", err)
		}
		return fmt.Errorf("failed to delete diagram '%s': %w", name, err)
	}

	return nil
}

// StoreStateMachine stores a parsed state machine with TTL support
func (rc *RedisCache) StoreStateMachine(ctx context.Context, umlVersion, name string, machine *models.StateMachine, ttl time.Duration) error {
	if umlVersion == "" {
		return internal.NewValidationError("UML version cannot be empty", nil)
	}

	if name == "" {
		return internal.NewValidationError("state machine name cannot be empty", nil)
	}

	if machine == nil {
		return internal.NewValidationError("state machine cannot be nil", nil)
	}

	// Generate cache key
	key := rc.keyGen.StateMachineKey(umlVersion, name)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return internal.NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Serialize the state machine to JSON
	data, err := json.Marshal(machine)
	if err != nil {
		return internal.NewSerializationError(key, "failed to marshal state machine", err)
	}

	// Use default TTL if not specified
	if ttl <= 0 {
		ttl = rc.config.DefaultTTL
	}

	// Store the serialized state machine in Redis
	err = rc.client.SetWithRetry(ctx, key, data, ttl)
	if err != nil {
		if isTimeoutError(err) {
			return internal.NewTimeoutError(key, "timeout storing state machine", err)
		}
		if isConnectionError(err) {
			return internal.NewConnectionError("failed to store state machine", err)
		}
		return fmt.Errorf("failed to store state machine '%s': %w", name, err)
	}

	return nil
}

// GetStateMachine retrieves a parsed state machine
func (rc *RedisCache) GetStateMachine(ctx context.Context, umlVersion, name string) (*models.StateMachine, error) {
	if umlVersion == "" {
		return nil, internal.NewValidationError("UML version cannot be empty", nil)
	}

	if name == "" {
		return nil, internal.NewValidationError("state machine name cannot be empty", nil)
	}

	// Generate cache key
	key := rc.keyGen.StateMachineKey(umlVersion, name)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return nil, internal.NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Retrieve the serialized state machine from Redis
	data, err := rc.client.GetWithRetry(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, internal.NewNotFoundError(key)
		}
		if isTimeoutError(err) {
			return nil, internal.NewTimeoutError(key, "timeout retrieving state machine", err)
		}
		if isConnectionError(err) {
			return nil, internal.NewConnectionError("failed to retrieve state machine", err)
		}
		return nil, fmt.Errorf("failed to retrieve state machine '%s': %w", name, err)
	}

	// Deserialize the state machine from JSON
	var machine models.StateMachine
	err = json.Unmarshal([]byte(data), &machine)
	if err != nil {
		return nil, internal.NewSerializationError(key, "failed to unmarshal state machine", err)
	}

	return &machine, nil
}

// DeleteStateMachine removes a state machine from the cache
func (rc *RedisCache) DeleteStateMachine(ctx context.Context, umlVersion, name string) error {
	if umlVersion == "" {
		return internal.NewValidationError("UML version cannot be empty", nil)
	}

	if name == "" {
		return internal.NewValidationError("state machine name cannot be empty", nil)
	}

	// Generate cache key
	key := rc.keyGen.StateMachineKey(umlVersion, name)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return internal.NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Delete the state machine from Redis
	err := rc.client.DelWithRetry(ctx, key)
	if err != nil {
		if isTimeoutError(err) {
			return internal.NewTimeoutError(key, "timeout deleting state machine", err)
		}
		if isConnectionError(err) {
			return internal.NewConnectionError("failed to delete state machine", err)
		}
		return fmt.Errorf("failed to delete state machine '%s': %w", name, err)
	}

	return nil
}

// StoreEntity stores a state machine entity with TTL support
func (rc *RedisCache) StoreEntity(ctx context.Context, umlVersion, diagramName, entityID string, entity interface{}, ttl time.Duration) error {
	if umlVersion == "" {
		return internal.NewValidationError("UML version cannot be empty", nil)
	}

	if diagramName == "" {
		return internal.NewValidationError("diagram name cannot be empty", nil)
	}

	if entityID == "" {
		return internal.NewValidationError("entity ID cannot be empty", nil)
	}

	if entity == nil {
		return internal.NewValidationError("entity cannot be nil", nil)
	}

	// Generate cache key
	key := rc.keyGen.EntityKey(umlVersion, diagramName, entityID)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return internal.NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Serialize the entity to JSON
	data, err := json.Marshal(entity)
	if err != nil {
		return internal.NewSerializationError(key, "failed to marshal entity", err)
	}

	// Use default TTL if not specified
	if ttl <= 0 {
		ttl = rc.config.DefaultTTL
	}

	// Store the serialized entity in Redis
	err = rc.client.SetWithRetry(ctx, key, data, ttl)
	if err != nil {
		if isTimeoutError(err) {
			return internal.NewTimeoutError(key, "timeout storing entity", err)
		}
		if isConnectionError(err) {
			return internal.NewConnectionError("failed to store entity", err)
		}
		return fmt.Errorf("failed to store entity '%s': %w", entityID, err)
	}

	return nil
}

// GetEntity retrieves a state machine entity
func (rc *RedisCache) GetEntity(ctx context.Context, umlVersion, diagramName, entityID string) (interface{}, error) {
	if umlVersion == "" {
		return nil, internal.NewValidationError("UML version cannot be empty", nil)
	}

	if diagramName == "" {
		return nil, internal.NewValidationError("diagram name cannot be empty", nil)
	}

	if entityID == "" {
		return nil, internal.NewValidationError("entity ID cannot be empty", nil)
	}

	// Generate cache key
	key := rc.keyGen.EntityKey(umlVersion, diagramName, entityID)

	// Validate the generated key
	if err := rc.keyGen.ValidateKey(key); err != nil {
		return nil, internal.NewKeyInvalidError(key, fmt.Sprintf("invalid key generated: %v", err))
	}

	// Retrieve the serialized entity from Redis
	data, err := rc.client.GetWithRetry(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, internal.NewNotFoundError(key)
		}
		if isTimeoutError(err) {
			return nil, internal.NewTimeoutError(key, "timeout retrieving entity", err)
		}
		if isConnectionError(err) {
			return nil, internal.NewConnectionError("failed to retrieve entity", err)
		}
		return nil, fmt.Errorf("failed to retrieve entity '%s': %w", entityID, err)
	}

	// Deserialize the entity from JSON
	var entity interface{}
	err = json.Unmarshal([]byte(data), &entity)
	if err != nil {
		return nil, internal.NewSerializationError(key, "failed to unmarshal entity", err)
	}

	return entity, nil
}

// Cleanup removes cache entries matching a pattern
func (rc *RedisCache) Cleanup(ctx context.Context, pattern string) error {
	if pattern == "" {
		return internal.NewValidationError("cleanup pattern cannot be empty", nil)
	}

	// Use Redis SCAN to find keys matching the pattern
	client := rc.client.Client()
	iter := client.Scan(ctx, 0, pattern, 0).Iterator()

	var keysToDelete []string
	for iter.Next(ctx) {
		keysToDelete = append(keysToDelete, iter.Val())
	}

	if err := iter.Err(); err != nil {
		if isTimeoutError(err) {
			return internal.NewTimeoutError("", "timeout during cleanup scan", err)
		}
		if isConnectionError(err) {
			return internal.NewConnectionError("failed to scan for cleanup", err)
		}
		return fmt.Errorf("failed to scan keys for cleanup: %w", err)
	}

	// Delete found keys in batches
	if len(keysToDelete) > 0 {
		err := rc.client.DelWithRetry(ctx, keysToDelete...)
		if err != nil {
			if isTimeoutError(err) {
				return internal.NewTimeoutError("", "timeout during cleanup deletion", err)
			}
			if isConnectionError(err) {
				return internal.NewConnectionError("failed to delete keys during cleanup", err)
			}
			return fmt.Errorf("failed to delete keys during cleanup: %w", err)
		}
	}

	return nil
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
