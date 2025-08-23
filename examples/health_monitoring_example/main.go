package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/cache"
)

func main() {
	fmt.Println("Redis Cache Health Monitoring Example")
	fmt.Println("=====================================")

	// Create Redis cache configuration
	config := cache.DefaultRedisConfig()
	config.RedisAddr = "localhost:6379"
	config.RedisDB = 0

	// Create cache instance
	redisCache, err := cache.NewRedisCache(config)
	if err != nil {
		log.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer redisCache.Close()

	ctx := context.Background()

	// Demonstrate basic health check
	fmt.Println("\n1. Basic Health Check")
	fmt.Println("---------------------")
	err = redisCache.Health(ctx)
	if err != nil {
		fmt.Printf("❌ Basic health check failed: %v\n", err)
	} else {
		fmt.Println("✅ Basic health check passed")
	}

	// Demonstrate detailed health status
	fmt.Println("\n2. Detailed Health Status")
	fmt.Println("-------------------------")
	healthStatus, err := redisCache.HealthDetailed(ctx)
	if err != nil {
		fmt.Printf("❌ Detailed health check failed: %v\n", err)
	} else {
		fmt.Printf("✅ Overall Status: %s\n", healthStatus.Status)
		fmt.Printf("📊 Response Time: %v\n", healthStatus.ResponseTime)
		fmt.Printf("🔗 Connected: %t\n", healthStatus.Connection.Connected)
		fmt.Printf("⚡ Ping Latency: %v\n", healthStatus.Connection.PingLatency)

		if len(healthStatus.Errors) > 0 {
			fmt.Println("❌ Errors:")
			for _, err := range healthStatus.Errors {
				fmt.Printf("   - %s\n", err)
			}
		}

		if len(healthStatus.Warnings) > 0 {
			fmt.Println("⚠️  Warnings:")
			for _, warning := range healthStatus.Warnings {
				fmt.Printf("   - %s\n", warning)
			}
		}
	}

	// Demonstrate connection health
	fmt.Println("\n3. Connection Health Details")
	fmt.Println("----------------------------")
	connHealth, err := redisCache.GetConnectionHealth(ctx)
	if err != nil {
		fmt.Printf("❌ Connection health check failed: %v\n", err)
	} else {
		fmt.Printf("🌐 Address: %s\n", connHealth.Address)
		fmt.Printf("🗄️  Database: %d\n", connHealth.Database)
		fmt.Printf("🔗 Connected: %t\n", connHealth.Connected)
		fmt.Printf("⚡ Ping Latency: %v\n", connHealth.PingLatency)
		fmt.Printf("🏊 Pool - Total: %d, Active: %d, Idle: %d\n",
			connHealth.PoolStats.TotalConnections,
			connHealth.PoolStats.ActiveConnections,
			connHealth.PoolStats.IdleConnections)
	}

	// Demonstrate performance metrics
	fmt.Println("\n4. Performance Metrics")
	fmt.Println("----------------------")
	perfMetrics, err := redisCache.GetPerformanceMetrics(ctx)
	if err != nil {
		fmt.Printf("❌ Performance metrics failed: %v\n", err)
	} else {
		fmt.Printf("💾 Memory Used: %s\n", perfMetrics.MemoryUsage.UsedMemoryHuman)
		fmt.Printf("📈 Memory Fragmentation: %.2f\n", perfMetrics.MemoryUsage.MemoryFragmentation)
		fmt.Printf("🔑 Total Keys: %d\n", perfMetrics.KeyspaceInfo.TotalKeys)
		fmt.Printf("🎯 Hit Rate: %.1f%%\n", perfMetrics.KeyspaceInfo.HitRate)
		fmt.Printf("⚙️  Commands/sec: %.1f\n", perfMetrics.OperationStats.CommandsPerSecond)
		fmt.Printf("🏷️  Redis Version: %s\n", perfMetrics.ServerInfo.RedisVersion)
		fmt.Printf("⏰ Uptime: %d days\n", perfMetrics.ServerInfo.UptimeDays)
	}

	// Demonstrate diagnostics
	fmt.Println("\n5. System Diagnostics")
	fmt.Println("---------------------")
	diagnostics, err := redisCache.RunDiagnostics(ctx)
	if err != nil {
		fmt.Printf("❌ Diagnostics failed: %v\n", err)
	} else {
		fmt.Printf("⚙️  Configuration Valid: %t\n", diagnostics.ConfigurationCheck.Valid)
		fmt.Printf("🌐 Network Reachable: %t\n", diagnostics.NetworkCheck.Reachable)
		fmt.Printf("📊 Performance Acceptable: %t\n", diagnostics.PerformanceCheck.Acceptable)
		fmt.Printf("🔍 Data Consistent: %t\n", diagnostics.DataIntegrityCheck.Consistent)

		if diagnostics.DataIntegrityCheck.OrphanedKeys > 0 {
			fmt.Printf("🗑️  Orphaned Keys: %d\n", diagnostics.DataIntegrityCheck.OrphanedKeys)
		}

		if len(diagnostics.Recommendations) > 0 {
			fmt.Println("💡 Recommendations:")
			for _, rec := range diagnostics.Recommendations {
				fmt.Printf("   - %s\n", rec)
			}
		}
	}

	// Store some test data to demonstrate cache functionality
	fmt.Println("\n6. Testing Cache Operations")
	fmt.Println("---------------------------")

	// Store a test diagram
	err = redisCache.StoreDiagram(ctx, "test-diagram", "@startuml\nstate A\nstate B\nA --> B\n@enduml", time.Hour)
	if err != nil {
		fmt.Printf("❌ Failed to store diagram: %v\n", err)
	} else {
		fmt.Println("✅ Test diagram stored successfully")
	}

	// Get updated health status after operations
	fmt.Println("\n7. Health Status After Operations")
	fmt.Println("---------------------------------")
	healthStatus, err = redisCache.HealthDetailed(ctx)
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
	} else {
		fmt.Printf("✅ Status: %s\n", healthStatus.Status)
		fmt.Printf("🔑 Total Keys: %d\n", healthStatus.Performance.KeyspaceInfo.TotalKeys)

		// Pretty print the full health status as JSON for detailed inspection
		if healthJSON, err := json.MarshalIndent(healthStatus, "", "  "); err == nil {
			fmt.Println("\n📋 Full Health Status (JSON):")
			fmt.Println(string(healthJSON))
		}
	}

	fmt.Println("\n🎉 Health monitoring example completed!")
}
