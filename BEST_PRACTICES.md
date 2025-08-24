# Best Practices Guide

This guide provides best practices for using the go-uml-statemachine-cache library effectively and safely.

## Configuration Best Practices

### Environment-Specific Configuration

Use different configurations for different environments:

```go
func createCacheConfig(env string) *cache.RedisConfig {
    config := cache.DefaultRedisConfig()
    
    switch env {
    case "development":
        config.RedisAddr = "localhost:6379"
        config.DefaultTTL = 1 * time.Hour
        config.PoolSize = 5
    case "production":
        config.RedisAddr = os.Getenv("REDIS_ADDR")
        config.RedisPassword = os.Getenv("REDIS_PASSWORD")
        config.DefaultTTL = 24 * time.Hour
        config.PoolSize = 20
    }
    
    return config
}
```

### Security Configuration

1. **Always use authentication in production**:
   ```go
   config.RedisPassword = os.Getenv("REDIS_PASSWORD")
   ```

2. **Use environment variables for sensitive data**:
   ```go
   config.RedisAddr = os.Getenv("REDIS_ADDR")
   ```

3. **Validate configuration before use**:
   ```go
   if config.RedisAddr == "" {
       log.Fatal("Redis address is required")
   }
   ```

## Error Handling Best Practices

### Comprehensive Error Checking

Always check for specific error types:

```go
content, err := redisCache.GetDiagram(ctx, name)
if err != nil {
    switch {
    case cache.IsNotFoundError(err):
        // Load from source and cache
        content = loadFromSource(name)
        redisCache.StoreDiagram(ctx, name, content, time.Hour)
    case cache.IsConnectionError(err):
        // Use fallback or retry
        log.Printf("Redis connection error: %v", err)
        return fallbackContent, nil
    case cache.IsValidationError(err):
        // Fix input and retry
        log.Printf("Validation error: %v", err)
        return "", err
    default:
        log.Printf("Unexpected cache error: %v", err)
        return "", err
    }
}
```

### Graceful Degradation

Implement fallback mechanisms:

```go
func getDiagramWithFallback(ctx context.Context, name string) (string, error) {
    // Try cache first
    content, err := redisCache.GetDiagram(ctx, name)
    if err == nil {
        return content, nil
    }
    
    // Fall back to database on cache miss or error
    if cache.IsNotFoundError(err) || cache.IsConnectionError(err) {
        content, err := loadFromDatabase(name)
        if err != nil {
            return "", err
        }
        
        // Cache for future use (best effort)
        go func() {
            redisCache.StoreDiagram(context.Background(), name, content, time.Hour)
        }()
        
        return content, nil
    }
    
    return "", err
}
```

## Performance Best Practices

### Connection Pool Sizing

Size your connection pool appropriately:

```go
// For low concurrency (< 10 goroutines)
config.PoolSize = 5

// For medium concurrency (10-100 goroutines)
config.PoolSize = 15

// For high concurrency (100+ goroutines)
config.PoolSize = 30
```

### TTL Management

Use appropriate TTL values:

```go
// Short-lived development data
redisCache.StoreDiagram(ctx, name, content, 30*time.Minute)

// Production data
redisCache.StoreDiagram(ctx, name, content, 24*time.Hour)

// Long-term stable data
redisCache.StoreDiagram(ctx, name, content, 7*24*time.Hour)
```

### Batch Operations

Use batch operations for cleanup:

```go
options := &cache.CleanupOptions{
    BatchSize:      100,
    ScanCount:      1000,
    CollectMetrics: true,
}

result, err := redisCache.CleanupWithOptions(ctx, pattern, options)
```

## Monitoring Best Practices

### Health Check Implementation

Implement regular health checks:

```go
func monitorCacheHealth(redisCache cache.Cache) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        
        health, err := redisCache.HealthDetailed(ctx)
        if err != nil {
            log.Printf("Health check failed: %v", err)
            cancel()
            continue
        }
        
        if health.Status != "healthy" {
            log.Printf("Cache unhealthy: %s", health.Status)
            // Implement alerting
        }
        
        // Monitor key metrics
        if health.Performance.KeyspaceInfo.HitRate < 80.0 {
            log.Printf("Low hit rate: %.2f%%", health.Performance.KeyspaceInfo.HitRate)
        }
        
        cancel()
    }
}
```

### Memory Monitoring

Monitor cache memory usage:

```go
func monitorMemoryUsage(redisCache cache.Cache) {
    sizeInfo, err := redisCache.GetCacheSize(context.Background())
    if err != nil {
        log.Printf("Failed to get cache size: %v", err)
        return
    }
    
    // Alert if memory usage is high
    if sizeInfo.MemoryUsed > 1024*1024*1024 { // 1GB
        log.Printf("High memory usage: %d bytes", sizeInfo.MemoryUsed)
        
        // Implement cleanup strategy
        cleanupOldEntries(redisCache)
    }
}
```

## Windows-Specific Best Practices

### Path Handling

Use proper path separators:

```go
// Use Go's filepath package for cross-platform compatibility
import "path/filepath"

examplePath := filepath.Join("examples", "diagram_cache_example", "main.go")
```

### Command Execution

Use Windows-compatible commands:

```cmd
REM Use Windows commands
go run examples\diagram_cache_example\main.go

REM Or use forward slashes (Go handles both)
go run examples/diagram_cache_example/main.go
```

### Redis Installation

Prefer containerized Redis on Windows:

```cmd
docker run --name redis-cache -p 6379:6379 -d redis:latest
```

## Testing Best Practices

### Unit Testing

Write comprehensive unit tests:

```go
func TestCacheOperations(t *testing.T) {
    config := cache.DefaultRedisConfig()
    config.RedisAddr = "localhost:6379"
    
    redisCache, err := cache.NewRedisCache(config)
    require.NoError(t, err)
    defer redisCache.Close()
    
    ctx := context.Background()
    
    // Test store and retrieve
    err = redisCache.StoreDiagram(ctx, "test", "content", time.Hour)
    require.NoError(t, err)
    
    content, err := redisCache.GetDiagram(ctx, "test")
    require.NoError(t, err)
    assert.Equal(t, "content", content)
}
```

### Integration Testing

Test with real Redis:

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Requires Redis running on localhost:6379
    config := cache.DefaultRedisConfig()
    redisCache, err := cache.NewRedisCache(config)
    require.NoError(t, err)
    defer redisCache.Close()
    
    // Test health check
    err = redisCache.Health(context.Background())
    require.NoError(t, err)
}
```

## Deployment Best Practices

### Configuration Management

Use configuration files or environment variables:

```go
type Config struct {
    Redis cache.RedisConfig `json:"redis"`
    App   AppConfig         `json:"app"`
}

func loadConfig() (*Config, error) {
    data, err := os.ReadFile("config.json")
    if err != nil {
        return nil, err
    }
    
    var config Config
    err = json.Unmarshal(data, &config)
    return &config, err
}
```

### Graceful Shutdown

Implement proper shutdown:

```go
func main() {
    redisCache, err := cache.NewRedisCache(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Handle shutdown signals
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        log.Println("Shutting down...")
        redisCache.Close()
        os.Exit(0)
    }()
    
    // Your application logic here
}
```

### Logging

Implement structured logging:

```go
import "github.com/sirupsen/logrus"

func setupLogging() {
    logrus.SetFormatter(&logrus.JSONFormatter{})
    logrus.SetLevel(logrus.InfoLevel)
}

func logCacheOperation(operation string, key string, duration time.Duration, err error) {
    fields := logrus.Fields{
        "operation": operation,
        "key":       key,
        "duration":  duration,
    }
    
    if err != nil {
        fields["error"] = err.Error()
        logrus.WithFields(fields).Error("Cache operation failed")
    } else {
        logrus.WithFields(fields).Info("Cache operation completed")
    }
}
```

## Security Best Practices

### Input Validation

The library automatically validates inputs, but you should also validate at the application level:

```go
func validateDiagramName(name string) error {
    if len(name) == 0 {
        return fmt.Errorf("diagram name cannot be empty")
    }
    
    if len(name) > 100 {
        return fmt.Errorf("diagram name too long")
    }
    
    // Check for valid characters
    if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name) {
        return fmt.Errorf("diagram name contains invalid characters")
    }
    
    return nil
}
```

### Network Security

Secure your Redis connections:

```go
config := &cache.RedisConfig{
    RedisAddr:     "redis.internal:6379", // Use internal networks
    RedisPassword: os.Getenv("REDIS_PASSWORD"),
    // Consider TLS for sensitive data
}
```

### Access Control

Implement proper access control:

```go
func (s *Service) GetDiagram(ctx context.Context, userID, diagramName string) (string, error) {
    // Check user permissions
    if !s.hasAccess(userID, diagramName) {
        return "", fmt.Errorf("access denied")
    }
    
    return s.cache.GetDiagram(ctx, diagramName)
}
```

## Maintenance Best Practices

### Regular Cleanup

Implement regular cleanup:

```go
func scheduleCleanup(redisCache cache.Cache) {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
        
        // Clean up old test data
        err := redisCache.Cleanup(ctx, "/diagrams/puml/test-*")
        if err != nil {
            log.Printf("Cleanup failed: %v", err)
        }
        
        cancel()
    }
}
```

### Monitoring and Alerting

Set up monitoring:

```go
func setupMonitoring(redisCache cache.Cache) {
    go func() {
        for {
            time.Sleep(5 * time.Minute)
            
            health, err := redisCache.HealthDetailed(context.Background())
            if err != nil {
                sendAlert("Cache health check failed: " + err.Error())
                continue
            }
            
            if health.Status != "healthy" {
                sendAlert("Cache is unhealthy: " + health.Status)
            }
            
            // Monitor performance metrics
            if health.Performance.KeyspaceInfo.HitRate < 70.0 {
                sendAlert(fmt.Sprintf("Low cache hit rate: %.2f%%", 
                    health.Performance.KeyspaceInfo.HitRate))
            }
        }
    }()
}
```

### Backup and Recovery

Consider backup strategies for critical data:

```go
func backupCriticalData(redisCache cache.Cache) error {
    // Export critical diagrams
    criticalDiagrams := []string{"main-workflow", "core-process"}
    
    for _, name := range criticalDiagrams {
        content, err := redisCache.GetDiagram(context.Background(), name)
        if err != nil {
            if cache.IsNotFoundError(err) {
                continue
            }
            return err
        }
        
        // Save to backup storage
        err = saveToBackup(name, content)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

Following these best practices will help you build robust, secure, and performant applications using the go-uml-statemachine-cache library.