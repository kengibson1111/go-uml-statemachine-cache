package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/cache"
)

func main() {
	// Command line flags
	var (
		redisAddr = flag.String("addr", "localhost:6379", "Redis server address")
		redisDB   = flag.Int("db", 0, "Redis database number")
		password  = flag.String("password", "", "Redis password")
		timeout   = flag.Duration("timeout", 10*time.Second, "Connection timeout")
		jsonOut   = flag.Bool("json", false, "Output in JSON format")
		verbose   = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	// Create Redis cache configuration
	config := cache.DefaultRedisConfig()
	config.RedisAddr = *redisAddr
	config.RedisDB = *redisDB
	config.RedisPassword = *password
	config.DialTimeout = *timeout
	config.ReadTimeout = *timeout
	config.WriteTimeout = *timeout

	// Create cache instance
	redisCache, err := cache.NewRedisCache(config)
	if err != nil {
		log.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer redisCache.Close()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Perform health check
	healthStatus, err := redisCache.HealthDetailed(ctx)
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}

	if *jsonOut {
		// Output JSON format
		output, err := json.MarshalIndent(healthStatus, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		// Output human-readable format
		printHealthStatus(healthStatus, *verbose)
	}

	// Set exit code based on health status
	switch healthStatus.Status {
	case "healthy":
		os.Exit(0)
	case "degraded":
		os.Exit(1)
	case "unhealthy":
		os.Exit(2)
	default:
		os.Exit(3)
	}
}

func printHealthStatus(health *cache.HealthStatus, verbose bool) {
	// Status header
	statusIcon := getStatusIcon(health.Status)
	fmt.Printf("%s Redis Cache Health Status: %s\n", statusIcon, health.Status)
	fmt.Printf("📅 Timestamp: %s\n", health.Timestamp.Format(time.RFC3339))
	fmt.Printf("⏱️  Response Time: %v\n", health.ResponseTime)
	fmt.Println()

	// Connection status
	fmt.Println("🔗 Connection Status:")
	if health.Connection.Connected {
		fmt.Printf("   ✅ Connected to %s (DB %d)\n", health.Connection.Address, health.Connection.Database)
		fmt.Printf("   ⚡ Ping Latency: %v\n", health.Connection.PingLatency)
	} else {
		fmt.Printf("   ❌ Disconnected from %s\n", health.Connection.Address)
		if health.Connection.LastError != "" {
			fmt.Printf("   🚨 Error: %s\n", health.Connection.LastError)
		}
	}

	if verbose {
		fmt.Printf("   🏊 Pool Stats: Total=%d, Active=%d, Idle=%d\n",
			health.Connection.PoolStats.TotalConnections,
			health.Connection.PoolStats.ActiveConnections,
			health.Connection.PoolStats.IdleConnections)
	}
	fmt.Println()

	// Performance metrics
	fmt.Println("📊 Performance Metrics:")
	fmt.Printf("   💾 Memory: %s (Fragmentation: %.2f)\n",
		health.Performance.MemoryUsage.UsedMemoryHuman,
		health.Performance.MemoryUsage.MemoryFragmentation)
	fmt.Printf("   🔑 Keys: %d (Expiring: %d)\n",
		health.Performance.KeyspaceInfo.TotalKeys,
		health.Performance.KeyspaceInfo.ExpiringKeys)
	fmt.Printf("   🎯 Hit Rate: %.1f%%\n", health.Performance.KeyspaceInfo.HitRate)
	fmt.Printf("   ⚙️  Commands/sec: %.1f\n", health.Performance.OperationStats.CommandsPerSecond)

	if verbose {
		fmt.Printf("   🏷️  Redis Version: %s\n", health.Performance.ServerInfo.RedisVersion)
		fmt.Printf("   ⏰ Uptime: %d days\n", health.Performance.ServerInfo.UptimeDays)
	}
	fmt.Println()

	// Diagnostics summary
	fmt.Println("🔍 Diagnostics:")
	fmt.Printf("   ⚙️  Configuration: %s\n", getBoolIcon(health.Diagnostics.ConfigurationCheck.Valid))
	fmt.Printf("   🌐 Network: %s\n", getBoolIcon(health.Diagnostics.NetworkCheck.Reachable))
	fmt.Printf("   📈 Performance: %s\n", getBoolIcon(health.Diagnostics.PerformanceCheck.Acceptable))
	fmt.Printf("   🔒 Data Integrity: %s\n", getBoolIcon(health.Diagnostics.DataIntegrityCheck.Consistent))

	if health.Diagnostics.DataIntegrityCheck.OrphanedKeys > 0 {
		fmt.Printf("   🗑️  Orphaned Keys: %d\n", health.Diagnostics.DataIntegrityCheck.OrphanedKeys)
	}
	fmt.Println()

	// Errors
	if len(health.Errors) > 0 {
		fmt.Println("❌ Errors:")
		for _, err := range health.Errors {
			fmt.Printf("   • %s\n", err)
		}
		fmt.Println()
	}

	// Warnings
	if len(health.Warnings) > 0 {
		fmt.Println("⚠️  Warnings:")
		for _, warning := range health.Warnings {
			fmt.Printf("   • %s\n", warning)
		}
		fmt.Println()
	}

	// Recommendations
	if len(health.Diagnostics.Recommendations) > 0 {
		fmt.Println("💡 Recommendations:")
		for _, rec := range health.Diagnostics.Recommendations {
			fmt.Printf("   • %s\n", rec)
		}
		fmt.Println()
	}

	// Exit code information
	fmt.Println("📋 Exit Codes:")
	fmt.Println("   0 = Healthy")
	fmt.Println("   1 = Degraded")
	fmt.Println("   2 = Unhealthy")
	fmt.Println("   3 = Unknown")
}

func getStatusIcon(status string) string {
	switch status {
	case "healthy":
		return "✅"
	case "degraded":
		return "⚠️"
	case "unhealthy":
		return "❌"
	default:
		return "❓"
	}
}

func getBoolIcon(value bool) string {
	if value {
		return "✅ OK"
	}
	return "❌ Failed"
}
