package cache

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRedisCache_HealthDetailed(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockRedisClient)
		expectedStatus string
		expectError    bool
	}{
		{
			name: "healthy system",
			setupMocks: func(mockClient *MockRedisClient) {
				// Mock successful health check
				mockClient.On("HealthWithRetry", mock.Anything).Return(nil)

				// Mock Redis client - return nil to avoid pool stats access
				mockClient.On("Client").Return((*redis.Client)(nil))

				// Mock INFO commands for performance metrics
				mockClient.On("InfoWithRetry", mock.Anything, "memory").Return(
					"used_memory:1048576\r\nused_memory_human:1.00M\r\nused_memory_peak:2097152\r\nmem_fragmentation_ratio:1.2\r\nmaxmemory:0\r\nmaxmemory_policy:noeviction\r\n", nil)
				mockClient.On("InfoWithRetry", mock.Anything, "keyspace").Return(
					"db0:keys=100,expires=50,avg_ttl=3600000\r\n", nil)
				mockClient.On("InfoWithRetry", mock.Anything, "stats").Return(
					"total_commands_processed:1000\r\ninstantaneous_ops_per_sec:10.5\r\nkeyspace_hits:900\r\nkeyspace_misses:100\r\nrejected_connections:0\r\n", nil)
				mockClient.On("InfoWithRetry", mock.Anything, "server").Return(
					"redis_version:7.0.0\r\nuptime_in_seconds:86400\r\nredis_mode:standalone\r\nrole:master\r\nconnected_slaves:0\r\n", nil)

				// Mock diagnostics - Client() returns nil so data integrity check returns early
			},
			expectedStatus: "healthy",
			expectError:    false,
		},
		{
			name: "degraded system - high fragmentation",
			setupMocks: func(mockClient *MockRedisClient) {
				mockClient.On("HealthWithRetry", mock.Anything).Return(nil)
				mockClient.On("Client").Return((*redis.Client)(nil))

				// High memory fragmentation
				mockClient.On("InfoWithRetry", mock.Anything, "memory").Return(
					"used_memory:1048576\r\nused_memory_human:1.00M\r\nused_memory_peak:2097152\r\nmem_fragmentation_ratio:2.5\r\nmaxmemory:0\r\nmaxmemory_policy:noeviction\r\n", nil)
				mockClient.On("InfoWithRetry", mock.Anything, "keyspace").Return(
					"db0:keys=100,expires=50,avg_ttl=3600000\r\n", nil)
				mockClient.On("InfoWithRetry", mock.Anything, "stats").Return(
					"total_commands_processed:1000\r\ninstantaneous_ops_per_sec:10.5\r\nkeyspace_hits:700\r\nkeyspace_misses:300\r\nrejected_connections:0\r\n", nil)
				mockClient.On("InfoWithRetry", mock.Anything, "server").Return(
					"redis_version:7.0.0\r\nuptime_in_seconds:86400\r\nredis_mode:standalone\r\nrole:master\r\nconnected_slaves:0\r\n", nil)

				// Mock diagnostics - Client() returns nil so data integrity check returns early
			},
			expectedStatus: "degraded",
			expectError:    false,
		},
		{
			name: "unhealthy system - connection failed",
			setupMocks: func(mockClient *MockRedisClient) {
				// Mock connection health check failure
				mockClient.On("HealthWithRetry", mock.Anything).Return(fmt.Errorf("connection refused"))
				mockClient.On("Client").Return((*redis.Client)(nil))

				// Mock performance metrics failure
				mockClient.On("InfoWithRetry", mock.Anything, "memory").Return("", fmt.Errorf("connection failed"))
				mockClient.On("InfoWithRetry", mock.Anything, "keyspace").Return("", fmt.Errorf("connection failed"))
				mockClient.On("InfoWithRetry", mock.Anything, "stats").Return("", fmt.Errorf("connection failed"))
				mockClient.On("InfoWithRetry", mock.Anything, "server").Return("", fmt.Errorf("connection failed"))

				// Mock diagnostics - network check (5 pings for network check)
				mockClient.On("HealthWithRetry", mock.Anything).Return(fmt.Errorf("connection refused")).Times(5)

				// Mock diagnostics - performance check (calls GetPerformanceMetrics again)
				mockClient.On("InfoWithRetry", mock.Anything, "memory").Return("", fmt.Errorf("connection failed"))
				mockClient.On("InfoWithRetry", mock.Anything, "keyspace").Return("", fmt.Errorf("connection failed"))
				mockClient.On("InfoWithRetry", mock.Anything, "stats").Return("", fmt.Errorf("connection failed"))
				mockClient.On("InfoWithRetry", mock.Anything, "server").Return("", fmt.Errorf("connection failed"))

				// Mock diagnostics - performance check calls GetConnectionHealth again
				mockClient.On("HealthWithRetry", mock.Anything).Return(fmt.Errorf("connection refused"))
				mockClient.On("Client").Return((*redis.Client)(nil))
			},
			expectedStatus: "unhealthy",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()

			tt.setupMocks(mockClient)

			config := DefaultRedisConfig()
			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

			ctx := context.Background()
			health, err := cache.HealthDetailed(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, health)
				assert.Equal(t, tt.expectedStatus, health.Status)
				assert.False(t, health.Timestamp.IsZero())
				assert.GreaterOrEqual(t, health.ResponseTime, time.Duration(0))
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRedisCache_GetConnectionHealth(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockRedisClient)
		expectError bool
		connected   bool
	}{
		{
			name: "successful connection health check",
			setupMocks: func(mockClient *MockRedisClient) {
				mockClient.On("HealthWithRetry", mock.Anything).Return(nil)

				// Mock Redis client - return nil to avoid pool stats access
				mockClient.On("Client").Return((*redis.Client)(nil))
			},
			expectError: false,
			connected:   true,
		},
		{
			name: "failed connection health check",
			setupMocks: func(mockClient *MockRedisClient) {
				mockClient.On("HealthWithRetry", mock.Anything).Return(fmt.Errorf("connection timeout"))
				mockClient.On("Client").Return((*redis.Client)(nil))
			},
			expectError: false,
			connected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()

			tt.setupMocks(mockClient)

			config := DefaultRedisConfig()
			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

			ctx := context.Background()
			health, err := cache.GetConnectionHealth(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, health)
				assert.Equal(t, tt.connected, health.Connected)
				assert.Equal(t, config.RedisAddr, health.Address)
				assert.Equal(t, config.RedisDB, health.Database)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRedisCache_GetPerformanceMetrics(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockRedisClient)
		expectError bool
	}{
		{
			name: "successful performance metrics retrieval",
			setupMocks: func(mockClient *MockRedisClient) {
				mockClient.On("InfoWithRetry", mock.Anything, "memory").Return(
					"used_memory:1048576\r\nused_memory_human:1.00M\r\nused_memory_peak:2097152\r\nmem_fragmentation_ratio:1.2\r\nmaxmemory:0\r\nmaxmemory_policy:noeviction\r\n", nil)
				mockClient.On("InfoWithRetry", mock.Anything, "keyspace").Return(
					"db0:keys=100,expires=50,avg_ttl=3600000\r\n", nil)
				mockClient.On("InfoWithRetry", mock.Anything, "stats").Return(
					"total_commands_processed:1000\r\ninstantaneous_ops_per_sec:10.5\r\nkeyspace_hits:900\r\nkeyspace_misses:100\r\nrejected_connections:0\r\n", nil)
				mockClient.On("InfoWithRetry", mock.Anything, "server").Return(
					"redis_version:7.0.0\r\nuptime_in_seconds:86400\r\nredis_mode:standalone\r\nrole:master\r\nconnected_slaves:0\r\n", nil)
			},
			expectError: false,
		},
		{
			name: "failed to get memory info",
			setupMocks: func(mockClient *MockRedisClient) {
				mockClient.On("InfoWithRetry", mock.Anything, "memory").Return("", fmt.Errorf("redis error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()

			tt.setupMocks(mockClient)

			config := DefaultRedisConfig()
			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

			ctx := context.Background()
			metrics, err := cache.GetPerformanceMetrics(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, metrics)

				// Verify memory metrics
				assert.Equal(t, int64(1048576), metrics.MemoryUsage.UsedMemory)
				assert.Equal(t, "1.00M", metrics.MemoryUsage.UsedMemoryHuman)
				assert.Equal(t, int64(2097152), metrics.MemoryUsage.UsedMemoryPeak)
				assert.Equal(t, 1.2, metrics.MemoryUsage.MemoryFragmentation)

				// Verify keyspace metrics
				assert.Equal(t, int64(100), metrics.KeyspaceInfo.TotalKeys)
				assert.Equal(t, int64(50), metrics.KeyspaceInfo.ExpiringKeys)

				// Verify keyspace metrics (hits/misses)
				assert.Equal(t, int64(900), metrics.KeyspaceInfo.KeyspaceHits)
				assert.Equal(t, int64(100), metrics.KeyspaceInfo.KeyspaceMisses)
				assert.Equal(t, 90.0, metrics.KeyspaceInfo.HitRate) // 900/(900+100) * 100

				// Verify operation metrics
				assert.Equal(t, int64(1000), metrics.OperationStats.TotalCommands)
				assert.Equal(t, 10.5, metrics.OperationStats.CommandsPerSecond)

				// Verify server metrics
				assert.Equal(t, "7.0.0", metrics.ServerInfo.RedisVersion)
				assert.Equal(t, int64(86400), metrics.ServerInfo.UptimeSeconds)
				assert.Equal(t, int64(1), metrics.ServerInfo.UptimeDays) // 86400/86400
				assert.Equal(t, "standalone", metrics.ServerInfo.ServerMode)
				assert.Equal(t, "master", metrics.ServerInfo.Role)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRedisCache_RunDiagnostics(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockRedisClient)
		expectError bool
	}{
		{
			name: "successful diagnostics",
			setupMocks: func(mockClient *MockRedisClient) {
				// Mock network check (multiple pings)
				mockClient.On("HealthWithRetry", mock.Anything).Return(nil).Times(5)

				// Mock performance metrics
				mockClient.On("InfoWithRetry", mock.Anything, "memory").Return(
					"used_memory:1048576\r\nused_memory_human:1.00M\r\nused_memory_peak:2097152\r\nmem_fragmentation_ratio:1.2\r\nmaxmemory:0\r\nmaxmemory_policy:noeviction\r\n", nil)
				mockClient.On("InfoWithRetry", mock.Anything, "keyspace").Return(
					"db0:keys=100,expires=50,avg_ttl=3600000\r\n", nil)
				mockClient.On("InfoWithRetry", mock.Anything, "stats").Return(
					"total_commands_processed:1000\r\ninstantaneous_ops_per_sec:10.5\r\nkeyspace_hits:900\r\nkeyspace_misses:100\r\nrejected_connections:0\r\n", nil)
				mockClient.On("InfoWithRetry", mock.Anything, "server").Return(
					"redis_version:7.0.0\r\nuptime_in_seconds:86400\r\nredis_mode:standalone\r\nrole:master\r\nconnected_slaves:0\r\n", nil)

				// Mock connection health
				mockClient.On("Client").Return((*redis.Client)(nil))

				// Mock data integrity check - Client() returns nil so data integrity check returns early
			},
			expectError: false,
		},
		{
			name: "network check failure",
			setupMocks: func(mockClient *MockRedisClient) {
				mockClient.On("HealthWithRetry", mock.Anything).Return(fmt.Errorf("network error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()

			tt.setupMocks(mockClient)

			config := DefaultRedisConfig()
			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

			ctx := context.Background()
			diagnostics, err := cache.RunDiagnostics(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, diagnostics)

				// Verify configuration check
				assert.NotNil(t, diagnostics.ConfigurationCheck)

				// Verify network check
				assert.NotNil(t, diagnostics.NetworkCheck)

				// Verify performance check
				assert.NotNil(t, diagnostics.PerformanceCheck)

				// Verify data integrity check
				assert.NotNil(t, diagnostics.DataIntegrityCheck)

				// Verify recommendations are generated
				assert.NotNil(t, diagnostics.Recommendations)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestParseMemoryInfo(t *testing.T) {
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := DefaultRedisConfig()
	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	memoryInfo := `# Memory
used_memory:1048576
used_memory_human:1.00M
used_memory_peak:2097152
mem_fragmentation_ratio:1.25
maxmemory:0
maxmemory_policy:noeviction
`

	metrics, err := cache.parseMemoryInfo(memoryInfo)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	assert.Equal(t, int64(1048576), metrics.UsedMemory)
	assert.Equal(t, "1.00M", metrics.UsedMemoryHuman)
	assert.Equal(t, int64(2097152), metrics.UsedMemoryPeak)
	assert.Equal(t, 1.25, metrics.MemoryFragmentation)
	assert.Equal(t, int64(0), metrics.MaxMemory)
	assert.Equal(t, "noeviction", metrics.MaxMemoryPolicy)
}

func TestParseKeyspaceInfo(t *testing.T) {
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := DefaultRedisConfig()
	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	keyspaceInfo := `# Keyspace
db0:keys=150,expires=75,avg_ttl=3600000
`

	metrics, err := cache.parseKeyspaceInfo(keyspaceInfo)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	assert.Equal(t, int64(150), metrics.TotalKeys)
	assert.Equal(t, int64(75), metrics.ExpiringKeys)
	assert.Equal(t, 50.0, metrics.AverageKeySize) // Default estimate
}

func TestParseStatsInfo(t *testing.T) {
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := DefaultRedisConfig()
	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	statsInfo := `# Stats
total_commands_processed:5000
instantaneous_ops_per_sec:25.5
keyspace_hits:4500
keyspace_misses:500
rejected_connections:2
`

	metrics, err := cache.parseStatsInfo(statsInfo)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	assert.Equal(t, int64(5000), metrics.TotalCommands)
	assert.Equal(t, 25.5, metrics.CommandsPerSecond)
	assert.Equal(t, int64(2), metrics.RejectedConnections)

	// Test keyspace stats parsing separately
	keyspaceMetrics := &KeyspaceMetrics{}
	cache.parseKeyspaceStatsFromInfo(statsInfo, keyspaceMetrics)
	assert.Equal(t, int64(4500), keyspaceMetrics.KeyspaceHits)
	assert.Equal(t, int64(500), keyspaceMetrics.KeyspaceMisses)
	assert.Equal(t, 90.0, keyspaceMetrics.HitRate) // 4500/(4500+500) * 100
}

func TestParseServerInfo(t *testing.T) {
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := DefaultRedisConfig()
	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	serverInfo := `# Server
redis_version:7.0.5
uptime_in_seconds:172800
redis_mode:standalone
role:master
connected_slaves:2
`

	metrics, err := cache.parseServerInfo(serverInfo)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	assert.Equal(t, "7.0.5", metrics.RedisVersion)
	assert.Equal(t, int64(172800), metrics.UptimeSeconds)
	assert.Equal(t, int64(2), metrics.UptimeDays) // 172800/86400
	assert.Equal(t, "standalone", metrics.ServerMode)
	assert.Equal(t, "master", metrics.Role)
	assert.Equal(t, int64(2), metrics.ConnectedSlaves)
}

func TestValidateConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		config         *RedisConfig
		expectValid    bool
		expectedIssues int
	}{
		{
			name:           "valid configuration",
			config:         DefaultRedisConfig(),
			expectValid:    true,
			expectedIssues: 0,
		},
		{
			name: "low timeouts",
			config: &RedisConfig{
				RedisAddr:    "localhost:6379",
				DialTimeout:  500 * time.Millisecond,
				ReadTimeout:  200 * time.Millisecond,
				WriteTimeout: 200 * time.Millisecond,
				PoolSize:     10,
				DefaultTTL:   24 * time.Hour,
			},
			expectValid:    false,
			expectedIssues: 3, // All three timeouts are too low
		},
		{
			name: "high retry attempts",
			config: &RedisConfig{
				RedisAddr:    "localhost:6379",
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10,
				DefaultTTL:   24 * time.Hour,
				RetryConfig: &RedisRetryConfig{
					MaxAttempts: 15,
					MaxDelay:    45 * time.Second,
				},
			},
			expectValid:    false,
			expectedIssues: 2, // High retry attempts and delay
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()

			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, tt.config)

			check := cache.validateConfiguration()

			assert.Equal(t, tt.expectValid, check.Valid)
			assert.Equal(t, tt.expectedIssues, len(check.Issues))

			if !tt.expectValid {
				assert.Greater(t, len(check.Recommendations), 0)
			}
		})
	}
}

func TestGenerateRecommendations(t *testing.T) {
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := DefaultRedisConfig()
	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	diagnostics := &DiagnosticInfo{
		ConfigurationCheck: ConfigCheck{
			Valid:           false,
			Issues:          []string{"Low timeout"},
			Recommendations: []string{"Increase timeout"},
		},
		NetworkCheck: NetworkCheck{
			Reachable: false,
		},
		PerformanceCheck: PerformanceCheck{
			Acceptable:      false,
			Bottlenecks:     []string{"High memory usage"},
			Recommendations: []string{"Optimize memory"},
		},
		DataIntegrityCheck: DataIntegrityCheck{
			Consistent:   false,
			OrphanedKeys: 5,
		},
	}

	recommendations := cache.generateRecommendations(diagnostics)

	assert.Greater(t, len(recommendations), 0)

	// Check that recommendations include expected items
	recommendationText := strings.Join(recommendations, " ")
	assert.Contains(t, recommendationText, "Increase timeout")
	assert.Contains(t, recommendationText, "Optimize memory")
	assert.Contains(t, recommendationText, "network connectivity")
	assert.Contains(t, recommendationText, "cleanup operation")
}
