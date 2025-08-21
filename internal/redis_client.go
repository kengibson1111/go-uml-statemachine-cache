package internal

import (
	"context"
	"fmt"
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
	}
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
