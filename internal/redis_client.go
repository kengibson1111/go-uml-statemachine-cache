package internal

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis connection configuration parameters
type Config struct {
	// Redis connection settings
	RedisAddr     string `json:"redis_addr"`     // Redis server address (host:port)
	RedisPassword string `json:"redis_password"` // Redis password (optional)
	RedisDB       int    `json:"redis_db"`       // Redis database number

	// Connection pool settings
	MaxRetries   int           `json:"max_retries"`   // Maximum number of retries
	DialTimeout  time.Duration `json:"dial_timeout"`  // Timeout for establishing connection
	ReadTimeout  time.Duration `json:"read_timeout"`  // Timeout for socket reads
	WriteTimeout time.Duration `json:"write_timeout"` // Timeout for socket writes
	PoolSize     int           `json:"pool_size"`     // Maximum number of socket connections

	// Cache settings
	DefaultTTL time.Duration `json:"default_ttl"` // Default time-to-live for cache entries

	// Resilience settings
	RetryConfig *RetryConfig `json:"retry_config"` // Retry configuration for operations
}

// RetryConfig defines retry behavior with exponential backoff
type RetryConfig struct {
	MaxAttempts  int           `json:"max_attempts"`  // Maximum number of retry attempts
	InitialDelay time.Duration `json:"initial_delay"` // Initial delay before first retry
	MaxDelay     time.Duration `json:"max_delay"`     // Maximum delay between retries
	Multiplier   float64       `json:"multiplier"`    // Backoff multiplier
	Jitter       bool          `json:"jitter"`        // Whether to add random jitter
	RetryableOps []string      `json:"retryable_ops"` // Operations that should be retried
}

// DefaultRetryConfig returns a RetryConfig with sensible default values
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
		RetryableOps: []string{"ping", "get", "set", "del", "exists", "expire"},
	}
}

// DefaultConfig returns a Config with sensible default values
func DefaultConfig() *Config {
	return &Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		MaxRetries:    3,
		DialTimeout:   5 * time.Second,
		ReadTimeout:   3 * time.Second,
		WriteTimeout:  3 * time.Second,
		PoolSize:      10,
		DefaultTTL:    24 * time.Hour,
		RetryConfig:   DefaultRetryConfig(),
	}
}

// RedisClientInterface defines the interface for Redis client operations
type RedisClientInterface interface {
	Health(ctx context.Context) error
	HealthWithRetry(ctx context.Context) error
	SetWithRetry(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	GetWithRetry(ctx context.Context, key string) (string, error)
	DelWithRetry(ctx context.Context, keys ...string) error
	Client() *redis.Client
	Config() *Config
	Close() error
}

// RedisClient wraps the go-redis client with additional functionality
type RedisClient struct {
	client *redis.Client
	config *Config
}

// NewRedisClient creates a new Redis client with the provided configuration
func NewRedisClient(config *Config) (*RedisClient, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create Redis client options
	opts := &redis.Options{
		Addr:         config.RedisAddr,
		Password:     config.RedisPassword,
		DB:           config.RedisDB,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolSize:     config.PoolSize,
	}

	// Create Redis client
	client := redis.NewClient(opts)

	redisClient := &RedisClient{
		client: client,
		config: config,
	}

	return redisClient, nil
}

// validateConfig validates the Redis configuration parameters
func validateConfig(config *Config) error {
	if config.RedisAddr == "" {
		return fmt.Errorf("redis address cannot be empty")
	}

	if config.RedisDB < 0 || config.RedisDB > 15 {
		return fmt.Errorf("redis database must be between 0 and 15, got %d", config.RedisDB)
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative, got %d", config.MaxRetries)
	}

	if config.DialTimeout <= 0 {
		return fmt.Errorf("dial timeout must be positive, got %v", config.DialTimeout)
	}

	if config.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive, got %v", config.ReadTimeout)
	}

	if config.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive, got %v", config.WriteTimeout)
	}

	if config.PoolSize <= 0 {
		return fmt.Errorf("pool size must be positive, got %d", config.PoolSize)
	}

	if config.DefaultTTL <= 0 {
		return fmt.Errorf("default TTL must be positive, got %v", config.DefaultTTL)
	}

	// Validate retry configuration if provided
	if config.RetryConfig != nil {
		if err := validateRetryConfig(config.RetryConfig); err != nil {
			return fmt.Errorf("invalid retry configuration: %w", err)
		}
	}

	return nil
}

// validateRetryConfig validates the retry configuration parameters
func validateRetryConfig(config *RetryConfig) error {
	if config.MaxAttempts < 0 {
		return fmt.Errorf("max attempts cannot be negative, got %d", config.MaxAttempts)
	}

	if config.InitialDelay < 0 {
		return fmt.Errorf("initial delay cannot be negative, got %v", config.InitialDelay)
	}

	if config.MaxDelay < 0 {
		return fmt.Errorf("max delay cannot be negative, got %v", config.MaxDelay)
	}

	if config.Multiplier < 1.0 {
		return fmt.Errorf("multiplier must be >= 1.0, got %f", config.Multiplier)
	}

	if config.InitialDelay > config.MaxDelay {
		return fmt.Errorf("initial delay (%v) cannot be greater than max delay (%v)", config.InitialDelay, config.MaxDelay)
	}

	return nil
}

// Health performs a health check on the Redis connection
func (rc *RedisClient) Health(ctx context.Context) error {
	// Use PING command to check if Redis is responsive
	pong, err := rc.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	if pong != "PONG" {
		return fmt.Errorf("unexpected ping response: %s", pong)
	}

	return nil
}

// Client returns the underlying Redis client for direct access
func (rc *RedisClient) Client() *redis.Client {
	return rc.client
}

// Config returns the Redis client configuration
func (rc *RedisClient) Config() *Config {
	return rc.config
}

// Close closes the Redis client connection
func (rc *RedisClient) Close() error {
	return rc.client.Close()
}

// GetConnectionInfo returns information about the current Redis connection
func (rc *RedisClient) GetConnectionInfo(ctx context.Context) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Get basic connection info
	info["addr"] = rc.config.RedisAddr
	info["db"] = rc.config.RedisDB
	info["pool_size"] = rc.config.PoolSize

	// Get pool stats
	poolStats := rc.client.PoolStats()
	info["pool_hits"] = poolStats.Hits
	info["pool_misses"] = poolStats.Misses
	info["pool_timeouts"] = poolStats.Timeouts
	info["pool_total_conns"] = poolStats.TotalConns
	info["pool_idle_conns"] = poolStats.IdleConns
	info["pool_stale_conns"] = poolStats.StaleConns

	return info, nil
}

// isRetryableError determines if an error should trigger a retry
func (rc *RedisClient) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for network-related errors that are typically retryable
	errorStr := err.Error()

	// Connection errors
	if contains(errorStr, "connection refused") ||
		contains(errorStr, "connection reset") ||
		contains(errorStr, "connection timeout") ||
		contains(errorStr, "network is unreachable") ||
		contains(errorStr, "no route to host") ||
		contains(errorStr, "broken pipe") ||
		contains(errorStr, "i/o timeout") {
		return true
	}

	// Redis-specific errors that might be retryable
	if contains(errorStr, "LOADING") ||
		contains(errorStr, "BUSY") ||
		contains(errorStr, "TRYAGAIN") {
		return true
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
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

// isOperationRetryable checks if the given operation should be retried
func (rc *RedisClient) isOperationRetryable(operation string) bool {
	if rc.config.RetryConfig == nil {
		return false
	}

	for _, op := range rc.config.RetryConfig.RetryableOps {
		if op == operation {
			return true
		}
	}
	return false
}

// calculateBackoffDelay calculates the delay for the next retry attempt
func (rc *RedisClient) calculateBackoffDelay(attempt int) time.Duration {
	if rc.config.RetryConfig == nil {
		return time.Second
	}

	config := rc.config.RetryConfig

	// Calculate exponential backoff
	delay := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt))

	// Cap at max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// Add jitter if enabled
	if config.Jitter {
		jitter := rand.Float64() * 0.1 * delay // 10% jitter
		delay += jitter
	}

	return time.Duration(delay)
}

// executeWithRetry executes a function with retry logic
func (rc *RedisClient) executeWithRetry(ctx context.Context, operation string, fn func() error) error {
	if !rc.isOperationRetryable(operation) || rc.config.RetryConfig == nil {
		return fn()
	}

	var lastErr error
	maxAttempts := rc.config.RetryConfig.MaxAttempts

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Execute the operation
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if !rc.isRetryableError(err) {
			return err
		}

		// Don't wait after the last attempt
		if attempt == maxAttempts-1 {
			break
		}

		// Calculate and wait for backoff delay
		delay := rc.calculateBackoffDelay(attempt)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("operation '%s' failed after %d attempts: %w", operation, maxAttempts, lastErr)
}

// HealthWithRetry performs a health check with retry logic
func (rc *RedisClient) HealthWithRetry(ctx context.Context) error {
	return rc.executeWithRetry(ctx, "ping", func() error {
		return rc.Health(ctx)
	})
}

// PingWithRetry performs a ping operation with retry logic
func (rc *RedisClient) PingWithRetry(ctx context.Context) error {
	return rc.executeWithRetry(ctx, "ping", func() error {
		pong, err := rc.client.Ping(ctx).Result()
		if err != nil {
			return fmt.Errorf("ping failed: %w", err)
		}
		if pong != "PONG" {
			return fmt.Errorf("unexpected ping response: %s", pong)
		}
		return nil
	})
}

// SetWithRetry performs a SET operation with retry logic
func (rc *RedisClient) SetWithRetry(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return rc.executeWithRetry(ctx, "set", func() error {
		return rc.client.Set(ctx, key, value, expiration).Err()
	})
}

// GetWithRetry performs a GET operation with retry logic
func (rc *RedisClient) GetWithRetry(ctx context.Context, key string) (string, error) {
	var result string
	err := rc.executeWithRetry(ctx, "get", func() error {
		val, err := rc.client.Get(ctx, key).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	})
	return result, err
}

// DelWithRetry performs a DEL operation with retry logic
func (rc *RedisClient) DelWithRetry(ctx context.Context, keys ...string) error {
	return rc.executeWithRetry(ctx, "del", func() error {
		return rc.client.Del(ctx, keys...).Err()
	})
}

// ExistsWithRetry performs an EXISTS operation with retry logic
func (rc *RedisClient) ExistsWithRetry(ctx context.Context, keys ...string) (int64, error) {
	var result int64
	err := rc.executeWithRetry(ctx, "exists", func() error {
		val, err := rc.client.Exists(ctx, keys...).Result()
		if err != nil {
			return err
		}
		result = val
		return nil
	})
	return result, err
}

// ExpireWithRetry performs an EXPIRE operation with retry logic
func (rc *RedisClient) ExpireWithRetry(ctx context.Context, key string, expiration time.Duration) error {
	return rc.executeWithRetry(ctx, "expire", func() error {
		return rc.client.Expire(ctx, key, expiration).Err()
	})
}
