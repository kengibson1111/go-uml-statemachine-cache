# Health Monitoring Example

This example demonstrates the comprehensive health monitoring capabilities of the Redis cache system.

## Features Demonstrated

1. **Basic Health Check** - Simple connectivity test
2. **Detailed Health Status** - Comprehensive system health with status categorization
3. **Connection Health** - Connection-specific metrics and pool statistics
4. **Performance Metrics** - Memory usage, hit rates, and server information
5. **System Diagnostics** - Configuration validation, network checks, and data integrity
6. **Real-time Monitoring** - Health status before and after cache operations

## Prerequisites

- Redis server running on localhost:6379
- Go 1.21 or later

## Running the Example

```bash
# Start Redis server (if not already running)
redis-server

# Run the example
go run examples/health_monitoring_example/main.go
```

## Expected Output

The example will show:

- ✅/❌ Status indicators for each health check
- 📊 Performance metrics (memory, hit rates, etc.)
- 🔗 Connection details and pool statistics
- 💡 Recommendations for optimization
- 📋 Full JSON health status for detailed inspection

## Health Status Categories

- **Healthy**: All systems operating normally
- **Degraded**: Minor issues detected (high fragmentation, low hit rate)
- **Unhealthy**: Critical issues (connection failures, configuration problems)

## Monitoring in Production

This health monitoring system can be integrated with:

- Prometheus/Grafana for metrics collection
- Health check endpoints in web services
- Automated alerting systems
- Load balancer health checks

## Troubleshooting

If the example fails to connect to Redis:

1. Ensure Redis is running: `redis-cli ping`
2. Check the Redis address in the configuration
3. Verify firewall settings allow connections to port 6379
4. Check Redis logs for any error messages