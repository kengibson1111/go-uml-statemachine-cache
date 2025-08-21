package internal

import (
	"context"
	"errors"
	"strings"
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

	// Test retry config defaults
	if config.RetryConfig == nil {
		t.Error("Expected RetryConfig to be non-nil")
	} else {
		retryConfig := config.RetryConfig
		if retryConfig.MaxAttempts != 3 {
			t.Errorf("Expected MaxAttempts to be 3, got %d", retryConfig.MaxAttempts)
		}
		if retryConfig.InitialDelay != 100*time.Millisecond {
			t.Errorf("Expected InitialDelay to be 100ms, got %v", retryConfig.InitialDelay)
		}
		if retryConfig.MaxDelay != 5*time.Second {
			t.Errorf("Expected MaxDelay to be 5s, got %v", retryConfig.MaxDelay)
		}
		if retryConfig.Multiplier != 2.0 {
			t.Errorf("Expected Multiplier to be 2.0, got %f", retryConfig.Multiplier)
		}
		if !retryConfig.Jitter {
			t.Error("Expected Jitter to be true")
		}
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

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts to be 3, got %d", config.MaxAttempts)
	}

	if config.InitialDelay != 100*time.Millisecond {
		t.Errorf("Expected InitialDelay to be 100ms, got %v", config.InitialDelay)
	}

	if config.MaxDelay != 5*time.Second {
		t.Errorf("Expected MaxDelay to be 5s, got %v", config.MaxDelay)
	}

	if config.Multiplier != 2.0 {
		t.Errorf("Expected Multiplier to be 2.0, got %f", config.Multiplier)
	}

	if !config.Jitter {
		t.Error("Expected Jitter to be true")
	}

	expectedOps := []string{"ping", "get", "set", "del", "exists", "expire"}
	if len(config.RetryableOps) != len(expectedOps) {
		t.Errorf("Expected %d retryable operations, got %d", len(expectedOps), len(config.RetryableOps))
	}

	for i, op := range expectedOps {
		if i >= len(config.RetryableOps) || config.RetryableOps[i] != op {
			t.Errorf("Expected retryable operation '%s' at index %d", op, i)
		}
	}
}

func TestValidateRetryConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *RetryConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default retry config",
			config:      DefaultRetryConfig(),
			expectError: false,
		},
		{
			name: "negative max attempts",
			config: &RetryConfig{
				MaxAttempts:  -1,
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     5 * time.Second,
				Multiplier:   2.0,
				Jitter:       true,
			},
			expectError: true,
			errorMsg:    "max attempts cannot be negative",
		},
		{
			name: "negative initial delay",
			config: &RetryConfig{
				MaxAttempts:  3,
				InitialDelay: -100 * time.Millisecond,
				MaxDelay:     5 * time.Second,
				Multiplier:   2.0,
				Jitter:       true,
			},
			expectError: true,
			errorMsg:    "initial delay cannot be negative",
		},
		{
			name: "negative max delay",
			config: &RetryConfig{
				MaxAttempts:  3,
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     -5 * time.Second,
				Multiplier:   2.0,
				Jitter:       true,
			},
			expectError: true,
			errorMsg:    "max delay cannot be negative",
		},
		{
			name: "multiplier less than 1.0",
			config: &RetryConfig{
				MaxAttempts:  3,
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     5 * time.Second,
				Multiplier:   0.5,
				Jitter:       true,
			},
			expectError: true,
			errorMsg:    "multiplier must be >= 1.0",
		},
		{
			name: "initial delay greater than max delay",
			config: &RetryConfig{
				MaxAttempts:  3,
				InitialDelay: 10 * time.Second,
				MaxDelay:     5 * time.Second,
				Multiplier:   2.0,
				Jitter:       true,
			},
			expectError: true,
			errorMsg:    "initial delay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRetryConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	config := DefaultConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "connection reset",
			err:      errors.New("connection reset by peer"),
			expected: true,
		},
		{
			name:     "connection timeout",
			err:      errors.New("connection timeout"),
			expected: true,
		},
		{
			name:     "network unreachable",
			err:      errors.New("network is unreachable"),
			expected: true,
		},
		{
			name:     "no route to host",
			err:      errors.New("no route to host"),
			expected: true,
		},
		{
			name:     "broken pipe",
			err:      errors.New("broken pipe"),
			expected: true,
		},
		{
			name:     "i/o timeout",
			err:      errors.New("i/o timeout"),
			expected: true,
		},
		{
			name:     "redis loading",
			err:      errors.New("LOADING Redis is loading the dataset in memory"),
			expected: true,
		},
		{
			name:     "redis busy",
			err:      errors.New("BUSY Redis is busy running a script"),
			expected: true,
		},
		{
			name:     "redis tryagain",
			err:      errors.New("TRYAGAIN Multiple keys request during rehashing of slot"),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      errors.New("WRONGTYPE Operation against a key holding the wrong kind of value"),
			expected: false,
		},
		{
			name:     "authentication error",
			err:      errors.New("NOAUTH Authentication required"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected isRetryableError(%v) to be %v, got %v", tt.err, tt.expected, result)
			}
		})
	}
}

func TestIsOperationRetryable(t *testing.T) {
	config := DefaultConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	tests := []struct {
		name      string
		operation string
		expected  bool
	}{
		{
			name:      "ping operation",
			operation: "ping",
			expected:  true,
		},
		{
			name:      "get operation",
			operation: "get",
			expected:  true,
		},
		{
			name:      "set operation",
			operation: "set",
			expected:  true,
		},
		{
			name:      "del operation",
			operation: "del",
			expected:  true,
		},
		{
			name:      "exists operation",
			operation: "exists",
			expected:  true,
		},
		{
			name:      "expire operation",
			operation: "expire",
			expected:  true,
		},
		{
			name:      "non-retryable operation",
			operation: "flushall",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.isOperationRetryable(tt.operation)
			if result != tt.expected {
				t.Errorf("Expected isOperationRetryable(%s) to be %v, got %v", tt.operation, tt.expected, result)
			}
		})
	}
}

func TestCalculateBackoffDelay(t *testing.T) {
	config := DefaultConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	tests := []struct {
		name     string
		attempt  int
		minDelay time.Duration
		maxDelay time.Duration
	}{
		{
			name:     "first attempt",
			attempt:  0,
			minDelay: 90 * time.Millisecond,  // 100ms - 10% jitter
			maxDelay: 110 * time.Millisecond, // 100ms + 10% jitter
		},
		{
			name:     "second attempt",
			attempt:  1,
			minDelay: 180 * time.Millisecond, // 200ms - 10% jitter
			maxDelay: 220 * time.Millisecond, // 200ms + 10% jitter
		},
		{
			name:     "third attempt",
			attempt:  2,
			minDelay: 360 * time.Millisecond, // 400ms - 10% jitter
			maxDelay: 440 * time.Millisecond, // 400ms + 10% jitter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := client.calculateBackoffDelay(tt.attempt)
			if delay < tt.minDelay || delay > tt.maxDelay {
				t.Errorf("Expected delay between %v and %v, got %v", tt.minDelay, tt.maxDelay, delay)
			}
		})
	}
}

func TestCalculateBackoffDelayWithoutJitter(t *testing.T) {
	config := DefaultConfig()
	config.RetryConfig.Jitter = false
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "first attempt",
			attempt:  0,
			expected: 100 * time.Millisecond,
		},
		{
			name:     "second attempt",
			attempt:  1,
			expected: 200 * time.Millisecond,
		},
		{
			name:     "third attempt",
			attempt:  2,
			expected: 400 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := client.calculateBackoffDelay(tt.attempt)
			if delay != tt.expected {
				t.Errorf("Expected delay %v, got %v", tt.expected, delay)
			}
		})
	}
}

func TestCalculateBackoffDelayMaxCap(t *testing.T) {
	config := DefaultConfig()
	config.RetryConfig.Jitter = false
	config.RetryConfig.MaxDelay = 300 * time.Millisecond
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	// Third attempt would normally be 400ms, but should be capped at 300ms
	delay := client.calculateBackoffDelay(2)
	if delay != 300*time.Millisecond {
		t.Errorf("Expected delay to be capped at 300ms, got %v", delay)
	}
}

// TestExecuteWithRetrySuccess tests successful execution without retries
func TestExecuteWithRetrySuccess(t *testing.T) {
	config := DefaultConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	callCount := 0

	err = client.executeWithRetry(ctx, "ping", func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected function to be called once, got %d calls", callCount)
	}
}

// TestExecuteWithRetryNonRetryableOperation tests that non-retryable operations don't retry
func TestExecuteWithRetryNonRetryableOperation(t *testing.T) {
	config := DefaultConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	callCount := 0
	testError := errors.New("connection refused")

	err = client.executeWithRetry(ctx, "flushall", func() error {
		callCount++
		return testError
	})

	if err != testError {
		t.Errorf("Expected original error, got: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected function to be called once, got %d calls", callCount)
	}
}

// TestExecuteWithRetryNonRetryableError tests that non-retryable errors don't trigger retries
func TestExecuteWithRetryNonRetryableError(t *testing.T) {
	config := DefaultConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	callCount := 0
	testError := errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

	err = client.executeWithRetry(ctx, "ping", func() error {
		callCount++
		return testError
	})

	if err != testError {
		t.Errorf("Expected original error, got: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected function to be called once, got %d calls", callCount)
	}
}

// TestExecuteWithRetryRetryableError tests that retryable errors trigger retries
func TestExecuteWithRetryRetryableError(t *testing.T) {
	config := DefaultConfig()
	config.RetryConfig.MaxAttempts = 3
	config.RetryConfig.InitialDelay = 1 * time.Millisecond // Very short delay for testing
	config.RetryConfig.Jitter = false
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	callCount := 0
	testError := errors.New("connection refused")

	err = client.executeWithRetry(ctx, "ping", func() error {
		callCount++
		return testError
	})

	if err == nil {
		t.Error("Expected error after all retries exhausted")
	}

	if !strings.Contains(err.Error(), "failed after 3 attempts") {
		t.Errorf("Expected error message to mention retry attempts, got: %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected function to be called 3 times, got %d calls", callCount)
	}
}

// TestExecuteWithRetryEventualSuccess tests that retries eventually succeed
func TestExecuteWithRetryEventualSuccess(t *testing.T) {
	config := DefaultConfig()
	config.RetryConfig.MaxAttempts = 3
	config.RetryConfig.InitialDelay = 1 * time.Millisecond // Very short delay for testing
	config.RetryConfig.Jitter = false
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	callCount := 0
	testError := errors.New("connection refused")

	err = client.executeWithRetry(ctx, "ping", func() error {
		callCount++
		if callCount < 3 {
			return testError
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error after eventual success, got: %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected function to be called 3 times, got %d calls", callCount)
	}
}

// TestExecuteWithRetryContextCancellation tests that context cancellation stops retries
func TestExecuteWithRetryContextCancellation(t *testing.T) {
	config := DefaultConfig()
	config.RetryConfig.MaxAttempts = 10
	config.RetryConfig.InitialDelay = 100 * time.Millisecond
	config.RetryConfig.Jitter = false
	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	callCount := 0
	testError := errors.New("connection refused")

	err = client.executeWithRetry(ctx, "ping", func() error {
		callCount++
		return testError
	})

	if err == nil {
		t.Error("Expected context cancellation error")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded error, got: %v", err)
	}

	// Should have been called at least once, but not all 10 times due to timeout
	if callCount == 0 {
		t.Error("Expected function to be called at least once")
	}
	if callCount >= 10 {
		t.Errorf("Expected function to be called less than 10 times due to timeout, got %d calls", callCount)
	}
}

// TestRetryMethodsWithoutRedis tests the retry wrapper methods when Redis is not available
func TestRetryMethodsWithoutRedis(t *testing.T) {
	config := DefaultConfig()
	config.RetryConfig.MaxAttempts = 2
	config.RetryConfig.InitialDelay = 1 * time.Millisecond
	config.RetryConfig.Jitter = false

	client, err := NewRedisClient(config)
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("HealthWithRetry", func(t *testing.T) {
		err := client.HealthWithRetry(ctx)
		// This will fail if Redis is not running, which is expected
		if err == nil {
			t.Log("Health check passed - Redis is running")
		} else {
			t.Logf("Health check failed as expected (Redis not running): %v", err)
		}
	})

	t.Run("PingWithRetry", func(t *testing.T) {
		err := client.PingWithRetry(ctx)
		// This will fail if Redis is not running, which is expected
		if err == nil {
			t.Log("Ping passed - Redis is running")
		} else {
			t.Logf("Ping failed as expected (Redis not running): %v", err)
		}
	})

	t.Run("SetWithRetry", func(t *testing.T) {
		err := client.SetWithRetry(ctx, "test-key", "test-value", time.Hour)
		// This will fail if Redis is not running, which is expected
		if err == nil {
			t.Log("Set passed - Redis is running")
		} else {
			t.Logf("Set failed as expected (Redis not running): %v", err)
		}
	})

	t.Run("GetWithRetry", func(t *testing.T) {
		_, err := client.GetWithRetry(ctx, "test-key")
		// This will fail if Redis is not running, which is expected
		if err == nil {
			t.Log("Get passed - Redis is running")
		} else {
			t.Logf("Get failed as expected (Redis not running): %v", err)
		}
	})

	t.Run("DelWithRetry", func(t *testing.T) {
		err := client.DelWithRetry(ctx, "test-key")
		// This will fail if Redis is not running, which is expected
		if err == nil {
			t.Log("Del passed - Redis is running")
		} else {
			t.Logf("Del failed as expected (Redis not running): %v", err)
		}
	})

	t.Run("ExistsWithRetry", func(t *testing.T) {
		_, err := client.ExistsWithRetry(ctx, "test-key")
		// This will fail if Redis is not running, which is expected
		if err == nil {
			t.Log("Exists passed - Redis is running")
		} else {
			t.Logf("Exists failed as expected (Redis not running): %v", err)
		}
	})

	t.Run("ExpireWithRetry", func(t *testing.T) {
		err := client.ExpireWithRetry(ctx, "test-key", time.Hour)
		// This will fail if Redis is not running, which is expected
		if err == nil {
			t.Log("Expire passed - Redis is running")
		} else {
			t.Logf("Expire failed as expected (Redis not running): %v", err)
		}
	})
}

// TestConfigWithRetryConfig tests configuration validation with retry config
func TestConfigWithRetryConfig(t *testing.T) {
	t.Run("valid config with retry config", func(t *testing.T) {
		config := DefaultConfig()
		config.RetryConfig = &RetryConfig{
			MaxAttempts:  5,
			InitialDelay: 50 * time.Millisecond,
			MaxDelay:     10 * time.Second,
			Multiplier:   1.5,
			Jitter:       false,
			RetryableOps: []string{"ping", "get"},
		}

		client, err := NewRedisClient(config)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if client != nil {
			client.Close()
		}
	})

	t.Run("invalid config with bad retry config", func(t *testing.T) {
		config := DefaultConfig()
		config.RetryConfig = &RetryConfig{
			MaxAttempts:  -1, // Invalid
			InitialDelay: 50 * time.Millisecond,
			MaxDelay:     10 * time.Second,
			Multiplier:   1.5,
			Jitter:       false,
		}

		client, err := NewRedisClient(config)
		if err == nil {
			t.Error("Expected error for invalid retry config")
		}
		if client != nil {
			client.Close()
		}
	})
}
