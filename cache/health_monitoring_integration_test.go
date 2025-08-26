package cache

import (
	"context"
	"testing"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthMonitoringIntegration tests the health monitoring functionality with a real Redis instance
// This test requires a Redis server running on localhost:6379
func TestHealthMonitoringIntegration(t *testing.T) {
	// Skip if running in CI or if Redis is not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create Redis cache configuration
	config := DefaultRedisConfig()
	config.RedisAddr = "localhost:6379"
	config.RedisDB = 0

	// Create cache instance
	redisCache, err := NewRedisCache(config)
	require.NoError(t, err, "Failed to create Redis cache")
	defer redisCache.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("basic health check", func(t *testing.T) {
		err := redisCache.Health(ctx)
		assert.NoError(t, err, "Basic health check should pass")
	})

	t.Run("detailed health status", func(t *testing.T) {
		healthStatus, err := redisCache.HealthDetailed(ctx)
		require.NoError(t, err, "Detailed health check should not fail")
		require.NotNil(t, healthStatus, "Health status should not be nil")

		// Verify basic structure
		assert.NotEmpty(t, healthStatus.Status, "Status should not be empty")
		assert.Contains(t, []string{"healthy", "degraded", "unhealthy"}, healthStatus.Status, "Status should be valid")
		assert.False(t, healthStatus.Timestamp.IsZero(), "Timestamp should be set")
		assert.GreaterOrEqual(t, healthStatus.ResponseTime, time.Duration(0), "Response time should be non-negative")

		// Verify connection health
		assert.True(t, healthStatus.Connection.Connected, "Should be connected to Redis")
		assert.Equal(t, config.RedisAddr, healthStatus.Connection.Address, "Address should match config")
		assert.Equal(t, config.RedisDB, healthStatus.Connection.Database, "Database should match config")
		assert.GreaterOrEqual(t, healthStatus.Connection.PingLatency, time.Duration(0), "Ping latency should be non-negative")

		// Verify performance metrics are populated
		assert.NotNil(t, healthStatus.Performance, "Performance metrics should be present")
		assert.GreaterOrEqual(t, healthStatus.Performance.MemoryUsage.UsedMemory, int64(0), "Used memory should be non-negative")
		assert.NotEmpty(t, healthStatus.Performance.ServerInfo.RedisVersion, "Redis version should be present")

		// Verify diagnostics are populated
		assert.NotNil(t, healthStatus.Diagnostics, "Diagnostics should be present")
		assert.NotNil(t, healthStatus.Diagnostics.ConfigurationCheck, "Configuration check should be present")
		assert.NotNil(t, healthStatus.Diagnostics.NetworkCheck, "Network check should be present")
		assert.NotNil(t, healthStatus.Diagnostics.PerformanceCheck, "Performance check should be present")
		assert.NotNil(t, healthStatus.Diagnostics.DataIntegrityCheck, "Data integrity check should be present")
	})

	t.Run("connection health", func(t *testing.T) {
		connHealth, err := redisCache.GetConnectionHealth(ctx)
		require.NoError(t, err, "Connection health check should not fail")
		require.NotNil(t, connHealth, "Connection health should not be nil")

		assert.True(t, connHealth.Connected, "Should be connected")
		assert.Equal(t, config.RedisAddr, connHealth.Address, "Address should match")
		assert.Equal(t, config.RedisDB, connHealth.Database, "Database should match")
		assert.GreaterOrEqual(t, connHealth.PingLatency, time.Duration(0), "Ping latency should be non-negative")
		assert.Empty(t, connHealth.LastError, "Should not have connection errors")

		// Pool stats should be reasonable
		assert.GreaterOrEqual(t, connHealth.PoolStats.TotalConnections, 0, "Total connections should be non-negative")
		assert.GreaterOrEqual(t, connHealth.PoolStats.IdleConnections, 0, "Idle connections should be non-negative")
	})

	t.Run("performance metrics", func(t *testing.T) {
		perfMetrics, err := redisCache.GetPerformanceMetrics(ctx)
		require.NoError(t, err, "Performance metrics should not fail")
		require.NotNil(t, perfMetrics, "Performance metrics should not be nil")

		// Memory metrics
		assert.GreaterOrEqual(t, perfMetrics.MemoryUsage.UsedMemory, int64(0), "Used memory should be non-negative")
		assert.NotEmpty(t, perfMetrics.MemoryUsage.UsedMemoryHuman, "Human readable memory should be present")
		assert.Greater(t, perfMetrics.MemoryUsage.MemoryFragmentation, 0.0, "Memory fragmentation should be positive")

		// Keyspace metrics
		assert.GreaterOrEqual(t, perfMetrics.KeyspaceInfo.TotalKeys, int64(0), "Total keys should be non-negative")
		assert.GreaterOrEqual(t, perfMetrics.KeyspaceInfo.ExpiringKeys, int64(0), "Expiring keys should be non-negative")
		assert.GreaterOrEqual(t, perfMetrics.KeyspaceInfo.HitRate, 0.0, "Hit rate should be non-negative")

		// Operation metrics
		assert.GreaterOrEqual(t, perfMetrics.OperationStats.TotalCommands, int64(0), "Total commands should be non-negative")
		assert.GreaterOrEqual(t, perfMetrics.OperationStats.CommandsPerSecond, 0.0, "Commands per second should be non-negative")

		// Server metrics
		assert.NotEmpty(t, perfMetrics.ServerInfo.RedisVersion, "Redis version should be present")
		assert.GreaterOrEqual(t, perfMetrics.ServerInfo.UptimeSeconds, int64(0), "Uptime should be non-negative")
	})

	t.Run("diagnostics", func(t *testing.T) {
		diagnostics, err := redisCache.RunDiagnostics(ctx)
		require.NoError(t, err, "Diagnostics should not fail")
		require.NotNil(t, diagnostics, "Diagnostics should not be nil")

		// Configuration check
		assert.NotNil(t, diagnostics.ConfigurationCheck, "Configuration check should be present")
		// Note: Configuration might be valid or invalid depending on settings

		// Network check
		assert.NotNil(t, diagnostics.NetworkCheck, "Network check should be present")
		assert.True(t, diagnostics.NetworkCheck.Reachable, "Network should be reachable")
		assert.Greater(t, diagnostics.NetworkCheck.Latency, time.Duration(0), "Network latency should be positive")
		assert.Equal(t, 0.0, diagnostics.NetworkCheck.PacketLoss, "Should have no packet loss")

		// Performance check
		assert.NotNil(t, diagnostics.PerformanceCheck, "Performance check should be present")
		// Note: Performance might be acceptable or not depending on Redis state

		// Data integrity check
		assert.NotNil(t, diagnostics.DataIntegrityCheck, "Data integrity check should be present")
		assert.GreaterOrEqual(t, diagnostics.DataIntegrityCheck.OrphanedKeys, int64(0), "Orphaned keys should be non-negative")

		// Recommendations should be present (might be empty)
		assert.NotNil(t, diagnostics.Recommendations, "Recommendations should be present")
	})

	t.Run("health monitoring with cache operations", func(t *testing.T) {
		// Store some test data
		err := redisCache.StoreDiagram(ctx, models.DiagramTypePUML, "health-test-diagram", "@startuml\nstate A\nstate B\nA --> B\n@enduml", time.Hour)
		require.NoError(t, err, "Should be able to store test diagram")

		// Get health status after operations
		healthStatus, err := redisCache.HealthDetailed(ctx)
		require.NoError(t, err, "Health check should work after operations")

		// Should still be connected and functional
		assert.True(t, healthStatus.Connection.Connected, "Should still be connected")
		assert.GreaterOrEqual(t, healthStatus.Performance.KeyspaceInfo.TotalKeys, int64(1), "Should have at least one key")

		// Clean up
		err = redisCache.DeleteDiagram(ctx, models.DiagramTypePUML, "health-test-diagram")
		assert.NoError(t, err, "Should be able to clean up test data")
	})
}

// TestHealthMonitoringWithInvalidRedis tests health monitoring behavior when Redis is not available
func TestHealthMonitoringWithInvalidRedis(t *testing.T) {
	// Create Redis cache configuration with invalid address
	config := DefaultRedisConfig()
	config.RedisAddr = "localhost:9999" // Invalid port
	config.DialTimeout = 1 * time.Second
	config.ReadTimeout = 1 * time.Second
	config.WriteTimeout = 1 * time.Second

	// Create cache instance
	redisCache, err := NewRedisCache(config)
	require.NoError(t, err, "Should be able to create cache with invalid config")
	defer redisCache.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("basic health check fails", func(t *testing.T) {
		err := redisCache.Health(ctx)
		assert.Error(t, err, "Health check should fail with invalid Redis")
	})

	t.Run("detailed health shows unhealthy", func(t *testing.T) {
		healthStatus, err := redisCache.HealthDetailed(ctx)
		require.NoError(t, err, "HealthDetailed should not fail even with invalid Redis")
		require.NotNil(t, healthStatus, "Health status should not be nil")

		assert.Equal(t, "unhealthy", healthStatus.Status, "Status should be unhealthy")
		assert.False(t, healthStatus.Connection.Connected, "Should not be connected")
		assert.NotEmpty(t, healthStatus.Connection.LastError, "Should have connection error")
		assert.Greater(t, len(healthStatus.Errors), 0, "Should have errors")
	})

	t.Run("connection health shows disconnected", func(t *testing.T) {
		connHealth, err := redisCache.GetConnectionHealth(ctx)
		require.NoError(t, err, "GetConnectionHealth should not fail")
		require.NotNil(t, connHealth, "Connection health should not be nil")

		assert.False(t, connHealth.Connected, "Should not be connected")
		assert.NotEmpty(t, connHealth.LastError, "Should have error message")
		assert.Equal(t, config.RedisAddr, connHealth.Address, "Address should match config")
	})
}
