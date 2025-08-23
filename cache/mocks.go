package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock implementation of the RedisClientInterface for testing
type MockRedisClient struct {
	mock.Mock
}

// NewMockRedisClient creates a new mock Redis client
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{}
}

// Health mocks the Health method
func (m *MockRedisClient) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// HealthWithRetry mocks the HealthWithRetry method
func (m *MockRedisClient) HealthWithRetry(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// SetWithRetry mocks the SetWithRetry method
func (m *MockRedisClient) SetWithRetry(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

// GetWithRetry mocks the GetWithRetry method
func (m *MockRedisClient) GetWithRetry(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

// DelWithRetry mocks the DelWithRetry method
func (m *MockRedisClient) DelWithRetry(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

// Client mocks the Client method
func (m *MockRedisClient) Client() *redis.Client {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*redis.Client)
}

// Config mocks the Config method
func (m *MockRedisClient) Config() *RedisConfig {
	args := m.Called()
	return args.Get(0).(*RedisConfig)
}

// Close mocks the Close method
func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Additional methods needed for enhanced cleanup functionality

// DBSize mocks the DBSize method
func (m *MockRedisClient) DBSize(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// Info mocks the Info method
func (m *MockRedisClient) Info(ctx context.Context, section string) (string, error) {
	args := m.Called(ctx, section)
	return args.String(0), args.Error(1)
}

// Scan mocks the Scan method
func (m *MockRedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	args := m.Called(ctx, cursor, match, count)
	return args.Get(0).([]string), args.Get(1).(uint64), args.Error(2)
}

// MemoryUsage mocks the MemoryUsage method
func (m *MockRedisClient) MemoryUsage(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

// ScanWithRetry mocks the ScanWithRetry method
func (m *MockRedisClient) ScanWithRetry(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	args := m.Called(ctx, cursor, match, count)
	return args.Get(0).([]string), args.Get(1).(uint64), args.Error(2)
}

// DelBatchWithRetry mocks the DelBatchWithRetry method
func (m *MockRedisClient) DelBatchWithRetry(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

// DBSizeWithRetry mocks the DBSizeWithRetry method
func (m *MockRedisClient) DBSizeWithRetry(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// InfoWithRetry mocks the InfoWithRetry method
func (m *MockRedisClient) InfoWithRetry(ctx context.Context, section string) (string, error) {
	args := m.Called(ctx, section)
	return args.String(0), args.Error(1)
}

// MemoryUsageWithRetry mocks the MemoryUsageWithRetry method
func (m *MockRedisClient) MemoryUsageWithRetry(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

// Legacy methods for backward compatibility
// Del mocks the Del method
func (m *MockRedisClient) Del(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

// MockKeyGenerator is a mock implementation of the KeyGenerator for testing
type MockKeyGenerator struct {
	mock.Mock
}

// NewMockKeyGenerator creates a new mock key generator
func NewMockKeyGenerator() *MockKeyGenerator {
	return &MockKeyGenerator{}
}

// DiagramKey mocks the DiagramKey method
func (m *MockKeyGenerator) DiagramKey(name string) string {
	args := m.Called(name)
	return args.String(0)
}

// StateMachineKey mocks the StateMachineKey method
func (m *MockKeyGenerator) StateMachineKey(umlVersion, name string) string {
	args := m.Called(umlVersion, name)
	return args.String(0)
}

// EntityKey mocks the EntityKey method
func (m *MockKeyGenerator) EntityKey(umlVersion, diagramName, entityID string) string {
	args := m.Called(umlVersion, diagramName, entityID)
	return args.String(0)
}

// ValidateKey mocks the ValidateKey method
func (m *MockKeyGenerator) ValidateKey(key string) error {
	args := m.Called(key)
	return args.Error(0)
}
