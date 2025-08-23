package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_Cleanup_Basic(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		pattern     string
		expectError bool
		errorType   CacheErrorType
		setupMocks  func(*MockRedisClient, *MockKeyGenerator)
	}{
		{
			name:        "successful cleanup with pattern",
			pattern:     "/diagrams/puml/*",
			expectError: false,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// Mock the scan operation
				mockClient.On("ScanWithRetry", mock.Anything, uint64(0), "/diagrams/puml/*", int64(100)).Return(
					[]string{"/diagrams/puml/test1", "/diagrams/puml/test2"}, uint64(0), nil)

				// Mock memory usage for metrics (default cleanup options have CollectMetrics=true)
				mockClient.On("MemoryUsageWithRetry", mock.Anything, "/diagrams/puml/test1").Return(int64(100), nil)
				mockClient.On("MemoryUsageWithRetry", mock.Anything, "/diagrams/puml/test2").Return(int64(150), nil)

				// Mock the delete operation
				mockClient.On("DelBatchWithRetry", mock.Anything, []string{"/diagrams/puml/test1", "/diagrams/puml/test2"}).Return(int64(2), nil)
			},
		},
		{
			name:        "empty pattern validation error",
			pattern:     "",
			expectError: true,
			errorType:   CacheErrorTypeValidation,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// No mocks needed for validation errors
			},
		},
		{
			name:        "scan timeout error",
			pattern:     "/machines/*",
			expectError: true,
			errorType:   CacheErrorTypeTimeout,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// Mock timeout error during scan
				mockClient.On("ScanWithRetry", mock.Anything, uint64(0), "/machines/*", int64(100)).Return(
					[]string{}, uint64(0), context.DeadlineExceeded)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()
			config := &RedisConfig{DefaultTTL: time.Hour}

			tt.setupMocks(mockClient, mockKeyGen)

			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

			err := cache.Cleanup(ctx, tt.pattern)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != 0 {
					cacheErr, ok := err.(*CacheError)
					require.True(t, ok, "expected CacheError, got %T", err)
					assert.Equal(t, tt.errorType, cacheErr.Type)
				}
			} else {
				require.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
			mockKeyGen.AssertExpectations(t)
		})
	}
}

func TestRedisCache_CleanupWithOptions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		pattern        string
		options        *CleanupOptions
		expectError    bool
		expectedResult *CleanupResult
		setupMocks     func(*MockRedisClient, *MockKeyGenerator)
	}{
		{
			name:    "successful cleanup with custom options",
			pattern: "/test/*",
			options: &CleanupOptions{
				BatchSize:      50,
				ScanCount:      50,
				MaxKeys:        100,
				DryRun:         false,
				Timeout:        30 * time.Second,
				CollectMetrics: true,
			},
			expectError: false,
			expectedResult: &CleanupResult{
				KeysScanned: 3,
				KeysDeleted: 3,
				BatchesUsed: 1,
			},
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// Mock scan operation
				mockClient.On("ScanWithRetry", mock.Anything, uint64(0), "/test/*", int64(50)).Return(
					[]string{"/test/key1", "/test/key2", "/test/key3"}, uint64(0), nil)

				// Mock memory usage for metrics
				mockClient.On("MemoryUsageWithRetry", mock.Anything, "/test/key1").Return(int64(100), nil)
				mockClient.On("MemoryUsageWithRetry", mock.Anything, "/test/key2").Return(int64(150), nil)
				mockClient.On("MemoryUsageWithRetry", mock.Anything, "/test/key3").Return(int64(200), nil)

				// Mock delete operation
				mockClient.On("DelBatchWithRetry", mock.Anything, []string{"/test/key1", "/test/key2", "/test/key3"}).Return(int64(3), nil)
			},
		},
		{
			name:    "dry run mode",
			pattern: "/test/*",
			options: &CleanupOptions{
				BatchSize:      100,
				ScanCount:      100,
				MaxKeys:        0,
				DryRun:         true,
				Timeout:        30 * time.Second,
				CollectMetrics: true,
			},
			expectError: false,
			expectedResult: &CleanupResult{
				KeysScanned: 2,
				KeysDeleted: 0, // No deletion in dry run
				BatchesUsed: 0,
			},
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// Mock scan operation
				mockClient.On("ScanWithRetry", mock.Anything, uint64(0), "/test/*", int64(100)).Return(
					[]string{"/test/key1", "/test/key2"}, uint64(0), nil)

				// No delete mocks needed for dry run
			},
		},
		{
			name:    "invalid options - negative batch size",
			pattern: "/test/*",
			options: &CleanupOptions{
				BatchSize: -1,
			},
			expectError: true,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// No mocks needed for validation errors
			},
		},
		{
			name:    "max keys limit",
			pattern: "/test/*",
			options: &CleanupOptions{
				BatchSize:      100,
				ScanCount:      100,
				MaxKeys:        2, // Limit to 2 keys
				DryRun:         false,
				Timeout:        30 * time.Second,
				CollectMetrics: false,
			},
			expectError: false,
			expectedResult: &CleanupResult{
				KeysScanned: 2, // Should stop at max keys
				KeysDeleted: 2,
				BatchesUsed: 1,
			},
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// Mock scan operation - returns more keys than max
				mockClient.On("ScanWithRetry", mock.Anything, uint64(0), "/test/*", int64(100)).Return(
					[]string{"/test/key1", "/test/key2", "/test/key3", "/test/key4"}, uint64(0), nil)

				// Mock delete operation - should only delete first 2 keys
				mockClient.On("DelBatchWithRetry", mock.Anything, []string{"/test/key1", "/test/key2"}).Return(int64(2), nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()
			config := &RedisConfig{DefaultTTL: time.Hour}

			tt.setupMocks(mockClient, mockKeyGen)

			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

			result, err := cache.CleanupWithOptions(ctx, tt.pattern, tt.options)

			if tt.expectError {
				require.Error(t, err)
				// Result might still be returned with partial information
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				if tt.expectedResult != nil {
					assert.Equal(t, tt.expectedResult.KeysScanned, result.KeysScanned)
					assert.Equal(t, tt.expectedResult.KeysDeleted, result.KeysDeleted)
					assert.Equal(t, tt.expectedResult.BatchesUsed, result.BatchesUsed)
					assert.True(t, result.Duration >= 0) // Should have non-negative duration
				}
			}

			mockClient.AssertExpectations(t)
			mockKeyGen.AssertExpectations(t)
		})
	}
}

func TestRedisCache_GetCacheSize(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		expectError    bool
		expectedResult *CacheSizeInfo
		setupMocks     func(*MockRedisClient, *MockKeyGenerator)
	}{
		{
			name:        "successful cache size retrieval",
			expectError: false,
			expectedResult: &CacheSizeInfo{
				TotalKeys:         100,
				DiagramCount:      25,
				StateMachineCount: 15,
				EntityCount:       60,
				MemoryUsed:        1048576, // 1MB
				MemoryPeak:        2097152, // 2MB
				MemoryOverhead:    524288,  // 512KB
			},
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// Mock DBSize
				mockClient.On("DBSizeWithRetry", mock.Anything).Return(int64(100), nil)

				// Mock Info for memory statistics
				memoryInfo := "used_memory:1048576\r\nused_memory_peak:2097152\r\nused_memory_overhead:524288\r\n"
				mockClient.On("InfoWithRetry", mock.Anything, "memory").Return(memoryInfo, nil)

				// Mock scan operations for counting different key types
				// Diagrams
				mockClient.On("ScanWithRetry", mock.Anything, uint64(0), "/diagrams/puml/*", int64(100)).Return(
					make([]string, 25), uint64(0), nil)

				// State machines
				mockClient.On("ScanWithRetry", mock.Anything, uint64(0), "/machines/*", int64(100)).Return(
					make([]string, 15), uint64(0), nil)

				// Entities
				mockClient.On("ScanWithRetry", mock.Anything, uint64(0), "/machines/*/entities/*", int64(100)).Return(
					make([]string, 60), uint64(0), nil)
			},
		},

		{
			name:        "dbsize error",
			expectError: true,
			setupMocks: func(mockClient *MockRedisClient, mockKeyGen *MockKeyGenerator) {
				// Mock DBSize error
				mockClient.On("DBSizeWithRetry", mock.Anything).Return(int64(0), errors.New("connection error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockRedisClient()
			mockKeyGen := NewMockKeyGenerator()
			config := &RedisConfig{DefaultTTL: time.Hour}

			tt.setupMocks(mockClient, mockKeyGen)

			cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

			result, err := cache.GetCacheSize(ctx)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				if tt.expectedResult != nil {
					assert.Equal(t, tt.expectedResult.TotalKeys, result.TotalKeys)
					assert.Equal(t, tt.expectedResult.DiagramCount, result.DiagramCount)
					assert.Equal(t, tt.expectedResult.StateMachineCount, result.StateMachineCount)
					assert.Equal(t, tt.expectedResult.EntityCount, result.EntityCount)
					assert.Equal(t, tt.expectedResult.MemoryUsed, result.MemoryUsed)
					assert.Equal(t, tt.expectedResult.MemoryPeak, result.MemoryPeak)
					assert.Equal(t, tt.expectedResult.MemoryOverhead, result.MemoryOverhead)
				}
			}

			mockClient.AssertExpectations(t)
			mockKeyGen.AssertExpectations(t)
		})
	}
}

func TestDefaultCleanupOptions(t *testing.T) {
	options := DefaultCleanupOptions()

	require.NotNil(t, options)
	assert.Equal(t, 100, options.BatchSize)
	assert.Equal(t, int64(100), options.ScanCount)
	assert.Equal(t, int64(0), options.MaxKeys)
	assert.False(t, options.DryRun)
	assert.Equal(t, 5*time.Minute, options.Timeout)
	assert.True(t, options.CollectMetrics)
}

func TestRedisCache_validateCleanupOptions(t *testing.T) {
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := &RedisConfig{DefaultTTL: time.Hour}
	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	tests := []struct {
		name        string
		options     *CleanupOptions
		expectError bool
	}{
		{
			name: "valid options",
			options: &CleanupOptions{
				BatchSize:      100,
				ScanCount:      100,
				MaxKeys:        1000,
				DryRun:         false,
				Timeout:        time.Minute,
				CollectMetrics: true,
			},
			expectError: false,
		},
		{
			name: "zero batch size",
			options: &CleanupOptions{
				BatchSize: 0,
				ScanCount: 100,
			},
			expectError: true,
		},
		{
			name: "negative batch size",
			options: &CleanupOptions{
				BatchSize: -1,
				ScanCount: 100,
			},
			expectError: true,
		},
		{
			name: "batch size too large",
			options: &CleanupOptions{
				BatchSize: 1001,
				ScanCount: 100,
			},
			expectError: true,
		},
		{
			name: "zero scan count",
			options: &CleanupOptions{
				BatchSize: 100,
				ScanCount: 0,
			},
			expectError: true,
		},
		{
			name: "negative max keys",
			options: &CleanupOptions{
				BatchSize: 100,
				ScanCount: 100,
				MaxKeys:   -1,
			},
			expectError: true,
		},
		{
			name: "negative timeout",
			options: &CleanupOptions{
				BatchSize: 100,
				ScanCount: 100,
				MaxKeys:   0,
				Timeout:   -time.Second,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.validateCleanupOptions(tt.options)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRedisCache_parseMemoryInfo(t *testing.T) {
	mockClient := NewMockRedisClient()
	mockKeyGen := NewMockKeyGenerator()
	config := &RedisConfig{DefaultTTL: time.Hour}
	cache := NewRedisCacheWithDependencies(mockClient, mockKeyGen, config)

	tests := []struct {
		name     string
		info     string
		key      string
		expected int64
	}{
		{
			name:     "parse used_memory",
			info:     "used_memory:1048576\r\nused_memory_peak:2097152\r\n",
			key:      "used_memory:",
			expected: 1048576,
		},
		{
			name:     "parse used_memory_peak",
			info:     "used_memory:1048576\r\nused_memory_peak:2097152\r\n",
			key:      "used_memory_peak:",
			expected: 2097152,
		},
		{
			name:     "key not found",
			info:     "used_memory:1048576\r\nused_memory_peak:2097152\r\n",
			key:      "nonexistent:",
			expected: 0,
		},
		{
			name:     "empty info",
			info:     "",
			key:      "used_memory:",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.parseMemoryInfo(tt.info, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		delimiter string
		expected  []string
	}{
		{
			name:      "normal split",
			input:     "a,b,c",
			delimiter: ",",
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "split with CRLF",
			input:     "line1\r\nline2\r\nline3",
			delimiter: "\r\n",
			expected:  []string{"line1", "line2", "line3"},
		},
		{
			name:      "empty string",
			input:     "",
			delimiter: ",",
			expected:  []string{},
		},
		{
			name:      "no delimiter found",
			input:     "nodelmiter",
			delimiter: ",",
			expected:  []string{"nodelmiter"},
		},
		{
			name:      "delimiter at end",
			input:     "a,b,",
			delimiter: ",",
			expected:  []string{"a", "b", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitString(tt.input, tt.delimiter)
			assert.Equal(t, tt.expected, result)
		})
	}
}
