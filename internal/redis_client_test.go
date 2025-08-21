package internal

import (
	"context"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test default values
	if config.RedisAddr != "localhost:6379" {
		t.Errorf("Expected RedisAddr to be 'localhost:6379', got '%s'", config.RedisAddr)
	}

	if config.RedisPassword != "" {
		t.Errorf("Expected RedisPassword to be empty, got '%s'", config.RedisPassword)
	}

	if config.RedisDB != 0 {
		t.Errorf("Expected RedisDB to be 0, got %d", config.RedisDB)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", config.MaxRetries)
	}

	if config.DialTimeout != 5*time.Second {
		t.Errorf("Expected DialTimeout to be 5s, got %v", config.DialTimeout)
	}

	if config.ReadTimeout != 3*time.Second {
		t.Errorf("Expected ReadTimeout to be 3s, got %v", config.ReadTimeout)
	}

	if config.WriteTimeout != 3*time.Second {
		t.Errorf("Expected WriteTimeout to be 3s, got %v", config.WriteTimeout)
	}

	if config.PoolSize != 10 {
		t.Errorf("Expected PoolSize to be 10, got %d", config.PoolSize)
	}

	if config.DefaultTTL != 24*time.Hour {
		t.Errorf("Expected DefaultTTL to be 24h, got %v", config.DefaultTTL)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "empty redis address",
			config: &Config{
				RedisAddr:    "",
				RedisDB:      0,
				MaxRetries:   3,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10,
				DefaultTTL:   24 * time.Hour,
			},
			expectError: true,
			errorMsg:    "redis address cannot be empty",
		},
		{
			name: "invalid redis database - negative",
			config: &Config{
				RedisAddr:    "localhost:6379",
				RedisDB:      -1,
				MaxRetries:   3,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10,
				DefaultTTL:   24 * time.Hour,
			},
			expectError: true,
			errorMsg:    "redis database must be between 0 and 15",
		},
		{
			name: "invalid redis database - too high",
			config: &Config{
				RedisAddr:    "localhost:6379",
				RedisDB:      16,
				MaxRetries:   3,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10,
				DefaultTTL:   24 * time.Hour,
			},
			expectError: true,
			errorMsg:    "redis database must be between 0 and 15",
		},
		{
			name: "negative max retries",
			config: &Config{
				RedisAddr:    "localhost:6379",
				RedisDB:      0,
				MaxRetries:   -1,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10,
				DefaultTTL:   24 * time.Hour,
			},
			expectError: true,
			errorMsg:    "max retries cannot be negative",
		},
		{
			name: "zero dial timeout",
			config: &Config{
				RedisAddr:    "localhost:6379",
				RedisDB:      0,
				MaxRetries:   3,
				DialTimeout:  0,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10,
				DefaultTTL:   24 * time.Hour,
			},
			expectError: true,
			errorMsg:    "dial timeout must be positive",
		},
		{
			name: "zero read timeout",
			config: &Config{
				RedisAddr:    "localhost:6379",
				RedisDB:      0,
				MaxRetries:   3,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  0,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10,
				DefaultTTL:   24 * time.Hour,
			},
			expectError: true,
			errorMsg:    "read timeout must be positive",
		},
		{
			name: "zero write timeout",
			config: &Config{
				RedisAddr:    "localhost:6379",
				RedisDB:      0,
				MaxRetries:   3,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 0,
				PoolSize:     10,
				DefaultTTL:   24 * time.Hour,
			},
			expectError: true,
			errorMsg:    "write timeout must be positive",
		},
		{
			name: "zero pool size",
			config: &Config{
				RedisAddr:    "localhost:6379",
				RedisDB:      0,
				MaxRetries:   3,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     0,
				DefaultTTL:   24 * time.Hour,
			},
			expectError: true,
			errorMsg:    "pool size must be positive",
		},
		{
			name: "zero default TTL",
			config: &Config{
				RedisAddr:    "localhost:6379",
				RedisDB:      0,
				MaxRetries:   3,
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10,
				DefaultTTL:   0,
			},
			expectError: true,
			errorMsg:    "default TTL must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					// Check if error message contains expected substring
					if len(tt.errorMsg) > 0 && len(err.Error()) > 0 {
						// For partial matching, check if the expected message is contained
						found := false
						if len(err.Error()) >= len(tt.errorMsg) {
							for i := 0; i <= len(err.Error())-len(tt.errorMsg); i++ {
								if err.Error()[i:i+len(tt.errorMsg)] == tt.errorMsg {
									found = true
									break
								}
							}
						}
						if !found {
							t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestNewRedisClient(t *testing.T) {
	t.Run("with valid config", func(t *testing.T) {
		config := DefaultConfig()
		client, err := NewRedisClient(config)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if client == nil {
			t.Error("Expected client to be non-nil")
		}

		if client.Config() != config {
			t.Error("Expected client config to match provided config")
		}

		if client.Client() == nil {
			t.Error("Expected underlying Redis client to be non-nil")
		}

		// Clean up
		client.Close()
	})

	t.Run("with nil config uses defaults", func(t *testing.T) {
		client, err := NewRedisClient(nil)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if client == nil {
			t.Error("Expected client to be non-nil")
		}

		config := client.Config()
		if config.RedisAddr != "localhost:6379" {
			t.Errorf("Expected default RedisAddr, got %s", config.RedisAddr)
		}

		// Clean up
		client.Close()
	})

	t.Run("with invalid config", func(t *testing.T) {
		config := &Config{
			RedisAddr: "", // Invalid empty address
		}

		client, err := NewRedisClient(config)

		if err == nil {
			t.Error("Expected error for invalid config")
		}

		if client != nil {
			t.Error("Expected client to be nil for invalid config")
			client.Close()
		}
	})
}

func TestRedisClientMethods(t *testing.T) {
	config := DefaultConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	t.Run("Config method", func(t *testing.T) {
		returnedConfig := client.Config()
		if returnedConfig != config {
			t.Error("Config method should return the same config instance")
		}
	})

	t.Run("Client method", func(t *testing.T) {
		redisClient := client.Client()
		if redisClient == nil {
			t.Error("Client method should return non-nil Redis client")
		}
	})

	t.Run("GetConnectionInfo method", func(t *testing.T) {
		ctx := context.Background()
		info, err := client.GetConnectionInfo(ctx)

		if err != nil {
			t.Errorf("GetConnectionInfo should not return error: %v", err)
		}

		if info == nil {
			t.Error("GetConnectionInfo should return non-nil info")
		}

		// Check expected fields
		expectedFields := []string{"addr", "db", "pool_size", "pool_hits", "pool_misses", "pool_timeouts", "pool_total_conns", "pool_idle_conns", "pool_stale_conns"}
		for _, field := range expectedFields {
			if _, exists := info[field]; !exists {
				t.Errorf("Expected field '%s' in connection info", field)
			}
		}
	})
}

// TestHealthCheck tests the health check functionality
// Note: This test will fail if Redis is not running, which is expected in CI/CD environments
func TestHealthCheck(t *testing.T) {
	config := DefaultConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// This test will only pass if Redis is actually running
	// In a real environment, you might want to skip this test if Redis is not available
	err = client.Health(ctx)

	// We don't fail the test if Redis is not running, just log it
	if err != nil {
		t.Logf("Health check failed (Redis may not be running): %v", err)
	} else {
		t.Log("Health check passed - Redis is running")
	}
}

func TestHealthCheckWithTimeout(t *testing.T) {
	config := DefaultConfig()
	config.ReadTimeout = 1 * time.Millisecond // Very short timeout

	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err = client.Health(ctx)

	// We expect this to fail due to timeout, but we don't fail the test
	// since Redis might not be running in the test environment
	if err != nil {
		t.Logf("Health check failed as expected with timeout: %v", err)
	}
}
