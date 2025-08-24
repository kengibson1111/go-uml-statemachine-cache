# Configuration Guide

This guide provides detailed information about configuring the go-uml-statemachine-cache library for optimal performance and reliability.

## Table of Contents

- [Basic Configuration](#basic-configuration)
- [Redis Connection Settings](#redis-connection-settings)
- [Performance Tuning](#performance-tuning)
- [Retry and Resilience](#retry-and-resilience)
- [Security Configuration](#security-configuration)
- [Monitoring and Health Checks](#monitoring-and-health-checks)
- [Environment-Specific Settings](#environment-specific-settings)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Basic Configuration

### Default Configuration

The library provides sensible defaults for most use cases:

```go
config := cache.DefaultRedisConfig()
// Default values:
// RedisAddr: "localhost:6379"
// RedisDB: 0
// MaxRetries: 3
// DialTimeout: 5 seconds
// ReadTimeout: 3 seconds
// WriteTimeout: 3 seconds
// PoolSize: 10
// DefaultTTL: 24 hours
```

### Minimal Configuration

For basic usage, you only need to specify the Redis address:

```go
config := cache.DefaultRedisConfig()
config.RedisAddr = "your-redis-server:6379"

redisCache, err := cache.NewRedisCache(config)
if err != nil {
    log.Fatal("Failed to create cache:", err)
}
```

### Complete Configuration Example

```go
config := &cache.RedisConfig{
    // Redis connection settings
    RedisAddr:     "redis.example.com:6379",
    RedisPassword: "your-secure-password",
    RedisDB:       1,

    // Connection timeouts
    DialTimeout:  10 * time.Second,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 5 * time.Second,

    // Connection pool
    PoolSize:   20,
    MaxRetries: 5,

    // Cache behavior
    DefaultTTL: 12 * time.Hour,

    // Retry configuration
    RetryConfig: &cache.RedisRetryConfig{
        MaxAttempts:  3,
        InitialDelay: 200 * time.Millisecond,
        MaxDelay:     10 * time.Second,
        Multiplier:   2.5,
        Jitter:       true,
        RetryableOps: []string{"get", "set", "del", "exists", "scan"},
    },
}
```

## Redis Connection Settings

### Basic Connection Parameters

```go
config := &cache.RedisConfig{
    RedisAddr:     "localhost:6379",  // Redis server address
    RedisPassword: "",                // Password (leave empty if no auth)
    RedisDB:       0,                 // Database number (0-15)
}
```

### Connection Timeouts

Configure timeouts based on your network conditions and requirements:

```go
config.DialTimeout = 5 * time.Second   // Time to establish connection
config.ReadTimeout = 3 * time.Second   // Time to read response
config.WriteTimeout = 3 * time.Second  // Time to write request
```

**Recommendations:**
- **Local Redis**: 1-2 seconds for all timeouts
- **Same datacenter**: 3-5 seconds
- **Cross-region**: 10-15 seconds
- **High-latency networks**: 20-30 seconds

### Connection Pool Configuration

The connection pool manages Redis connections efficiently:

```go
config.PoolSize = 10    // Maximum connections in pool
config.MaxRetries = 3   // Retries for failed operations
```

**Pool Size Guidelines:**
- **Low concurrency** (1-10 goroutines): 5-10 connections
- **Medium concurrency** (10-100 goroutines): 10-20 connections
- **High concurrency** (100+ goroutines): 20-50 connections
- **Very high concurrency** (1000+ goroutines): 50-100 connections

**Note**: More connections aren't always better. Monitor pool statistics to find the optimal size.

## Performance Tuning

### TTL Configuration

Configure appropriate TTL values for different data types:

```go
config.DefaultTTL = 24 * time.Hour  // Default for all operations

// In your application, use specific TTLs:
// Short-lived data
redisCache.StoreDiagram(ctx, name, content, 1*time.Hour)

// Long-lived data
redisCache.StoreDiagram(ctx, name, content, 7*24*time.Hour)

// Permanent data (use with caution)
redisCache.StoreDiagram(ctx, name, content, 0) // No expiration
```

**TTL Recommendations:**
- **Development/Testing**: 1-6 hours
- **Staging**: 6-24 hours
- **Production**: 24-168 hours (1-7 days)
- **Archive data**: 30-90 days

### Memory Optimization

Monitor and optimize memory usage:

```go
// Get cache size information
sizeInfo, err := redisCache.GetCacheSize(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Memory used: %d bytes\n", sizeInfo.MemoryUsed)
fmt.Printf("Total keys: %d\n", sizeInfo.TotalKeys)

// Implement cleanup strategy
if sizeInfo.MemoryUsed > 1024*1024*1024 { // 1GB
    // Cleanup old entries
    err = redisCache.Cleanup(ctx, "/diagrams/puml/old-*")
}
```

### Batch Operations

Use batch operations for better performance:

```go
// Cleanup with batching
options := &cache.CleanupOptions{
    BatchSize:      100,              // Process 100 keys at a time
    ScanCount:      1000,             // Scan 1000 keys per operation
    CollectMetrics: true,             // Collect performance metrics
}

result, err := redisCache.CleanupWithOptions(ctx, pattern, options)
```

## Retry and Resilience

### Retry Configuration

Configure retry behavior for different scenarios:

```go
retryConfig := &cache.RedisRetryConfig{
    MaxAttempts:  3,                    // Maximum retry attempts
    InitialDelay: 100 * time.Millisecond, // Initial delay
    MaxDelay:     5 * time.Second,      // Maximum delay
    Multiplier:   2.0,                  // Exponential backoff multiplier
    Jitter:       true,                 // Add randomness to delays
    RetryableOps: []string{             // Operations to retry
        "get", "set", "del", "exists", "scan", "ping",
    },
}

config.RetryConfig = retryConfig
```

### Circuit Breaker Configuration

Configure circuit breaker for fault tolerance:

```go
circuitConfig := cache.DefaultCircuitBreakerConfig()
circuitConfig.FailureThreshold = 5              // Failures before opening
circuitConfig.RecoveryTimeout = 30 * time.Second // Time before trying again
circuitConfig.SuccessThreshold = 3              // Successes to close circuit

recoveryManager := cache.NewErrorRecoveryManager(circuitConfig)
```

### Error Handling Strategy

Implement comprehensive error handling:

```go
_, err := redisCache.GetDiagram(ctx, "diagram-name")
if err != nil {
    switch {
    case cache.IsNotFoundError(err):
        // Cache miss - load from source
        content := loadFromSource("diagram-name")
        redisCache.StoreDiagram(ctx, "diagram-name", content, time.Hour)
        
    case cache.IsConnectionError(err):
        // Connection issue - use fallback or retry
        log.Printf("Redis connection error: %v", err)
        // Implement fallback logic
        
    case cache.IsRetryExhaustedError(err):
        // All retries failed - escalate or use circuit breaker
        log.Printf("All retries exhausted: %v", err)
        
    case cache.IsValidationError(err):
        // Input validation failed - fix input
        log.Printf("Validation error: %v", err)
        
    default:
        // Unexpected error
        log.Printf("Unexpected cache error: %v", err)
    }
}
```

## Security Configuration

### Authentication

Configure Redis authentication:

```go
config := &cache.RedisConfig{
    RedisAddr:     "secure-redis.example.com:6379",
    RedisPassword: "your-strong-password",
    RedisDB:       0,
}
```

### TLS/SSL Configuration

For secure connections (requires Redis with TLS support):

```go
import (
    "crypto/tls"
    "github.com/redis/go-redis/v9"
)

// Note: This requires direct Redis client configuration
// The cache library uses the go-redis client internally
```

### Input Validation

The library automatically validates and sanitizes inputs:

```go
// These inputs are automatically validated:
// - Diagram names: length, characters, encoding
// - Content: size, encoding, security patterns
// - Keys: format, length, special characters
// - TTL values: range, validity

// Validation errors are returned as CacheErrorTypeValidation
err := redisCache.StoreDiagram(ctx, "invalid\x00name", content, time.Hour)
if cache.IsValidationError(err) {
    log.Printf("Input validation failed: %v", err)
}
```

### Network Security

Configure network-level security:

```go
config := &cache.RedisConfig{
    RedisAddr: "127.0.0.1:6379",  // Bind to localhost only
    // Or use private network addresses
    RedisAddr: "10.0.1.100:6379", // Private network
}
```

## Monitoring and Health Checks

### Basic Health Monitoring

```go
// Simple health check
err := redisCache.Health(ctx)
if err != nil {
    log.Printf("Cache is unhealthy: %v", err)
    // Implement alerting or fallback
}
```

### Detailed Health Monitoring

```go
health, err := redisCache.HealthDetailed(ctx)
if err != nil {
    log.Fatal(err)
}

// Check overall status
if health.Status != "healthy" {
    log.Printf("Cache status: %s", health.Status)
    
    // Check specific issues
    if len(health.Errors) > 0 {
        log.Printf("Errors: %v", health.Errors)
    }
    
    if len(health.Warnings) > 0 {
        log.Printf("Warnings: %v", health.Warnings)
    }
}

// Monitor performance metrics
perf := health.Performance
log.Printf("Memory usage: %s", perf.MemoryUsage.UsedMemoryHuman)
log.Printf("Hit rate: %.2f%%", perf.KeyspaceInfo.HitRate)
log.Printf("Commands/sec: %.2f", perf.OperationStats.CommandsPerSecond)
```

### Connection Health Monitoring

```go
connHealth, err := redisCache.GetConnectionHealth(ctx)
if err != nil {
    log.Fatal(err)
}

log.Printf("Connected: %v", connHealth.Connected)
log.Printf("Ping latency: %v", connHealth.PingLatency)
log.Printf("Pool stats - Total: %d, Active: %d, Idle: %d",
    connHealth.PoolStats.TotalConnections,
    connHealth.PoolStats.ActiveConnections,
    connHealth.PoolStats.IdleConnections)
```

### Diagnostic Information

```go
diagnostics, err := redisCache.RunDiagnostics(ctx)
if err != nil {
    log.Fatal(err)
}

// Check configuration
if !diagnostics.ConfigurationCheck.Valid {
    log.Println("Configuration issues:")
    for _, issue := range diagnostics.ConfigurationCheck.Issues {
        log.Printf("  - %s", issue)
    }
}

// Performance recommendations
if len(diagnostics.Recommendations) > 0 {
    log.Println("Performance recommendations:")
    for _, rec := range diagnostics.Recommendations {
        log.Printf("  - %s", rec)
    }
}
```

## Cross-Platform Configuration

### Platform-Specific Redis Setup

#### Windows Configuration

**Redis Configuration File Location:**
- Chocolatey install: `C:\ProgramData\chocolatey\lib\redis-64\tools\redis.windows.conf`
- Manual install: `C:\Program Files\Redis\redis.windows.conf`

**Windows Service Management:**
```cmd
# Start Redis service
net start Redis

# Stop Redis service
net stop Redis

# Check service status
sc query Redis
```

**Windows-Specific Redis Config:**
```conf
# redis.windows.conf
bind 127.0.0.1
port 6379
timeout 0
tcp-keepalive 60
loglevel notice
logfile "C:/Program Files/Redis/Logs/redis_log.txt"
databases 16
save 900 1
save 300 10
save 60 10000
dir "C:/Program Files/Redis/Data/"
```

#### Linux Configuration

**Redis Configuration File Location:**
- Ubuntu/Debian: `/etc/redis/redis.conf`
- CentOS/RHEL: `/etc/redis.conf`
- Arch Linux: `/etc/redis/redis.conf`

**Linux Service Management:**
```bash
# SystemD (modern distributions)
sudo systemctl start redis-server
sudo systemctl stop redis-server
sudo systemctl restart redis-server
sudo systemctl status redis-server
sudo systemctl enable redis-server  # Auto-start on boot

# SysV Init (older distributions)
sudo service redis-server start
sudo service redis-server stop
sudo service redis-server restart
sudo service redis-server status
```

**Linux-Specific Redis Config:**
```conf
# /etc/redis/redis.conf
bind 127.0.0.1
port 6379
timeout 0
tcp-keepalive 300
daemonize yes
supervised systemd
pidfile /var/run/redis/redis-server.pid
loglevel notice
logfile /var/log/redis/redis-server.log
databases 16
dir /var/lib/redis
```

#### macOS Configuration

**Redis Configuration File Location:**
- Homebrew: `/usr/local/etc/redis.conf` or `/opt/homebrew/etc/redis.conf` (Apple Silicon)
- MacPorts: `/opt/local/etc/redis/redis.conf`

**macOS Service Management:**
```bash
# Homebrew services
brew services start redis
brew services stop redis
brew services restart redis
brew services list | grep redis

# Manual start (Homebrew)
redis-server /usr/local/etc/redis.conf

# MacPorts
sudo port load redis
sudo port unload redis
```

**macOS-Specific Redis Config:**
```conf
# /usr/local/etc/redis.conf (Homebrew)
bind 127.0.0.1
port 6379
timeout 0
tcp-keepalive 300
daemonize no  # Homebrew manages this
loglevel notice
logfile /usr/local/var/log/redis.log
databases 16
dir /usr/local/var/db/redis/
```

### Environment Variables by Platform

#### Windows Environment Variables

**Command Prompt:**
```cmd
set REDIS_ADDR=localhost:6379
set REDIS_PASSWORD=your-password
set REDIS_DB=0
set REDIS_POOL_SIZE=10
set REDIS_TIMEOUT=5s
```

**PowerShell:**
```powershell
$env:REDIS_ADDR="localhost:6379"
$env:REDIS_PASSWORD="your-password"
$env:REDIS_DB="0"
$env:REDIS_POOL_SIZE="10"
$env:REDIS_TIMEOUT="5s"
```

**Windows Registry (System-wide):**
```cmd
setx REDIS_ADDR "localhost:6379" /M
setx REDIS_PASSWORD "your-password" /M
```

#### Linux/macOS Environment Variables

**Bash/Zsh:**
```bash
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=your-password
export REDIS_DB=0
export REDIS_POOL_SIZE=10
export REDIS_TIMEOUT=5s
```

**System-wide (Linux):**
```bash
# Add to /etc/environment
echo 'REDIS_ADDR=localhost:6379' | sudo tee -a /etc/environment
echo 'REDIS_PASSWORD=your-password' | sudo tee -a /etc/environment
```

**User Profile (macOS):**
```bash
# Add to ~/.zshrc or ~/.bash_profile
echo 'export REDIS_ADDR=localhost:6379' >> ~/.zshrc
echo 'export REDIS_PASSWORD=your-password' >> ~/.zshrc
```

### Docker Configuration (Cross-Platform)

#### Basic Docker Setup
```bash
# Run Redis container
docker run --name redis-cache -p 6379:6379 -d redis:latest

# Run with password
docker run --name redis-cache -p 6379:6379 -d redis:latest redis-server --requirepass your-password

# Run with custom config
docker run --name redis-cache -p 6379:6379 -v /path/to/redis.conf:/usr/local/etc/redis/redis.conf -d redis:latest redis-server /usr/local/etc/redis/redis.conf
```

#### Docker Compose Configuration
```yaml
# docker-compose.yml
version: '3.8'
services:
  redis:
    image: redis:latest
    container_name: redis-cache
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
      - ./redis.conf:/usr/local/etc/redis/redis.conf
    command: redis-server /usr/local/etc/redis/redis.conf
    environment:
      - REDIS_PASSWORD=your-password
    restart: unless-stopped

volumes:
  redis_data:
```

### Go Application Configuration by Platform

#### Cross-Platform Configuration Loading
```go
package main

import (
    "os"
    "runtime"
    "path/filepath"
    "time"
    "github.com/kengibson1111/go-uml-statemachine-cache/cache"
)

func getConfigPath() string {
    switch runtime.GOOS {
    case "windows":
        return filepath.Join(os.Getenv("APPDATA"), "myapp", "config.json")
    case "darwin":
        return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "myapp", "config.json")
    default: // Linux and others
        configDir := os.Getenv("XDG_CONFIG_HOME")
        if configDir == "" {
            configDir = filepath.Join(os.Getenv("HOME"), ".config")
        }
        return filepath.Join(configDir, "myapp", "config.json")
    }
}

func createPlatformOptimizedConfig() *cache.RedisConfig {
    config := cache.DefaultRedisConfig()
    
    // Platform-specific optimizations
    switch runtime.GOOS {
    case "windows":
        // Windows-specific settings
        config.DialTimeout = 10 * time.Second  // Longer timeout for Windows
        config.PoolSize = 15                   // Slightly larger pool
        
    case "darwin":
        // macOS-specific settings
        config.DialTimeout = 5 * time.Second
        config.PoolSize = 10
        
    default: // Linux
        // Linux-specific settings
        config.DialTimeout = 3 * time.Second   // Faster on Linux
        config.PoolSize = 20                   // Can handle more connections
    }
    
    // Load from environment variables
    if addr := os.Getenv("REDIS_ADDR"); addr != "" {
        config.RedisAddr = addr
    }
    if password := os.Getenv("REDIS_PASSWORD"); password != "" {
        config.RedisPassword = password
    }
    
    return config
}
```

## Environment-Specific Settings

### Development Environment

```go
config := &cache.RedisConfig{
    RedisAddr:    "localhost:6379",
    RedisDB:      0,
    DialTimeout:  2 * time.Second,
    ReadTimeout:  1 * time.Second,
    WriteTimeout: 1 * time.Second,
    PoolSize:     5,
    DefaultTTL:   1 * time.Hour,  // Short TTL for development
    RetryConfig: &cache.RedisRetryConfig{
        MaxAttempts: 2,  // Fewer retries for faster feedback
        InitialDelay: 50 * time.Millisecond,
        MaxDelay:     1 * time.Second,
    },
}
```

### Staging Environment

```go
config := &cache.RedisConfig{
    RedisAddr:    "staging-redis.internal:6379",
    RedisDB:      1,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    PoolSize:     10,
    DefaultTTL:   6 * time.Hour,
    RetryConfig:  cache.DefaultRedisRetryConfig(),
}
```

### Production Environment

```go
config := &cache.RedisConfig{
    RedisAddr:     "prod-redis-cluster.internal:6379",
    RedisPassword: os.Getenv("REDIS_PASSWORD"),
    RedisDB:       0,
    DialTimeout:   10 * time.Second,
    ReadTimeout:   5 * time.Second,
    WriteTimeout:  5 * time.Second,
    PoolSize:      20,
    MaxRetries:    5,
    DefaultTTL:    24 * time.Hour,
    RetryConfig: &cache.RedisRetryConfig{
        MaxAttempts:  5,
        InitialDelay: 200 * time.Millisecond,
        MaxDelay:     10 * time.Second,
        Multiplier:   2.0,
        Jitter:       true,
        RetryableOps: []string{"get", "set", "del", "exists", "scan", "ping"},
    },
}
```

### High-Availability Environment

```go
config := &cache.RedisConfig{
    RedisAddr:     "redis-ha-proxy.internal:6379",
    RedisPassword: os.Getenv("REDIS_PASSWORD"),
    RedisDB:       0,
    DialTimeout:   15 * time.Second,  // Longer for failover
    ReadTimeout:   10 * time.Second,
    WriteTimeout:  10 * time.Second,
    PoolSize:      50,                // Larger pool for high load
    MaxRetries:    10,                // More retries for resilience
    DefaultTTL:    48 * time.Hour,    // Longer TTL for stability
    RetryConfig: &cache.RedisRetryConfig{
        MaxAttempts:  10,
        InitialDelay: 500 * time.Millisecond,
        MaxDelay:     30 * time.Second,
        Multiplier:   1.5,  // Gentler backoff
        Jitter:       true,
        RetryableOps: []string{"get", "set", "del", "exists", "scan", "ping"},
    },
}
```

## Best Practices

### Configuration Management

1. **Use environment variables** for sensitive configuration:
   ```go
   config.RedisAddr = os.Getenv("REDIS_ADDR")
   config.RedisPassword = os.Getenv("REDIS_PASSWORD")
   ```

2. **Validate configuration** before creating cache:
   ```go
   if config.RedisAddr == "" {
       log.Fatal("REDIS_ADDR environment variable is required")
   }
   ```

3. **Use configuration files** for complex setups:
   ```go
   // Load from JSON, YAML, or TOML configuration files
   ```

### Performance Optimization

1. **Monitor key metrics**:
   - Memory usage
   - Hit rate
   - Connection pool utilization
   - Response times

2. **Implement cleanup strategies**:
   ```go
   // Regular cleanup of old entries
   go func() {
       ticker := time.NewTicker(1 * time.Hour)
       for range ticker.C {
           redisCache.Cleanup(ctx, "/diagrams/puml/temp-*")
       }
   }()
   ```

3. **Use appropriate TTL values**:
   - Short TTL for frequently changing data
   - Long TTL for stable data
   - No TTL only for permanent data

### Error Handling

1. **Implement graceful degradation**:
   ```go
   content, err := redisCache.GetDiagram(ctx, name)
   if err != nil {
       // Fall back to loading from source
       content = loadFromDatabase(name)
   }
   ```

2. **Use circuit breakers** for critical operations
3. **Log errors with context** for debugging
4. **Implement alerting** for critical errors

### Security

1. **Use strong passwords** for Redis authentication
2. **Limit network access** to Redis servers
3. **Monitor for suspicious patterns** in logs
4. **Regularly update** Redis and client libraries

## Troubleshooting

### Common Issues

#### Connection Timeouts

**Symptoms**: Frequent timeout errors, slow responses

**Solutions**:
```go
// Increase timeouts
config.DialTimeout = 10 * time.Second
config.ReadTimeout = 5 * time.Second
config.WriteTimeout = 5 * time.Second

// Check network connectivity
// Verify Redis server performance
```

#### Pool Exhaustion

**Symptoms**: "connection pool timeout" errors

**Solutions**:
```go
// Increase pool size
config.PoolSize = 20

// Monitor pool usage
connHealth, _ := redisCache.GetConnectionHealth(ctx)
fmt.Printf("Pool utilization: %d/%d",
    connHealth.PoolStats.ActiveConnections,
    connHealth.PoolStats.TotalConnections)
```

#### Memory Issues

**Symptoms**: Redis out of memory, slow performance

**Solutions**:
```go
// Implement regular cleanup
options := &cache.CleanupOptions{
    MaxKeys: 10000,
    CollectMetrics: true,
}
redisCache.CleanupWithOptions(ctx, "/old-data/*", options)

// Reduce TTL values
config.DefaultTTL = 6 * time.Hour

// Monitor memory usage
sizeInfo, _ := redisCache.GetCacheSize(ctx)
```

#### High Error Rates

**Symptoms**: Many cache errors, degraded performance

**Solutions**:
```go
// Check error patterns
_, err := redisCache.GetDiagram(ctx, name)
if err != nil {
    if cacheErr, ok := err.(*cache.CacheError); ok {
        log.Printf("Error type: %v, Severity: %v",
            cacheErr.Type, cacheErr.Severity)
    }
}

// Run diagnostics
diagnostics, _ := redisCache.RunDiagnostics(ctx)
// Check diagnostics.Recommendations
```

### Debugging Tools

#### Enable Detailed Logging

```go
// Log all cache operations (development only)
import "log"

originalGet := redisCache.GetDiagram
redisCache.GetDiagram = func(ctx context.Context, name string) (string, error) {
    start := time.Now()
    result, err := originalGet(ctx, name)
    log.Printf("GetDiagram(%s) took %v, error: %v", name, time.Since(start), err)
    return result, err
}
```

#### Monitor Redis Directly

#### Cross-Platform Redis CLI Access

**Windows:**
```cmd
# Using Redis CLI (if installed via Chocolatey)
redis-cli -h localhost -p 6379

# Using Docker
docker exec -it redis-cache redis-cli

# Using WSL
wsl redis-cli -h localhost -p 6379
```

**Linux:**
```bash
# Using system Redis CLI
redis-cli -h localhost -p 6379

# Using Docker
docker exec -it redis-cache redis-cli

# Remote connection
redis-cli -h redis.example.com -p 6379 -a your-password
```

**macOS:**
```bash
# Using Homebrew Redis CLI
redis-cli -h localhost -p 6379

# Using Docker
docker exec -it redis-cache redis-cli

# Using MacPorts
/opt/local/bin/redis-cli -h localhost -p 6379
```

#### Common Redis CLI Commands

```bash
# Monitor all commands
MONITOR

# Check memory usage
INFO memory

# List keys by pattern
KEYS /diagrams/puml/*

# Check key TTL
TTL /diagrams/puml/my-diagram

# Get Redis configuration
CONFIG GET *

# Check Redis version
INFO server
```

#### Performance Profiling

```go
import _ "net/http/pprof"
import "net/http"

// Enable pprof endpoint
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Access profiling at http://localhost:6060/debug/pprof/
```

### Configuration Validation

```go
func validateConfig(config *cache.RedisConfig) error {
    if config.RedisAddr == "" {
        return fmt.Errorf("RedisAddr is required")
    }
    
    if config.PoolSize <= 0 {
        return fmt.Errorf("PoolSize must be positive")
    }
    
    if config.DefaultTTL < 0 {
        return fmt.Errorf("DefaultTTL cannot be negative")
    }
    
    // Add more validation as needed
    return nil
}

// Use before creating cache
if err := validateConfig(config); err != nil {
    log.Fatal("Invalid configuration:", err)
}
```

This configuration guide should help you optimize the cache library for your specific use case and environment. Remember to monitor performance and adjust settings based on your actual usage patterns.