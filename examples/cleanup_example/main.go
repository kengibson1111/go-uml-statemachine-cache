package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/cache"
	"github.com/kengibson1111/go-uml-statemachine-models/models"
)

func main() {
	fmt.Println("Redis Cache Cleanup Example")
	fmt.Println("===========================")

	// Create Redis cache configuration
	config := cache.DefaultRedisConfig()
	config.RedisAddr = "localhost:6379"
	config.RedisDB = 1 // Use database 1 for testing

	// Create cache instance
	redisCache, err := cache.NewRedisCache(config)
	if err != nil {
		log.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer redisCache.Close()

	ctx := context.Background()

	// Test Redis connection
	if err := redisCache.Health(ctx); err != nil {
		log.Fatalf("Redis health check failed: %v", err)
	}
	fmt.Println("✓ Connected to Redis successfully")

	// Populate cache with test data
	fmt.Println("\n1. Populating cache with test data...")
	if err := populateTestData(ctx, redisCache); err != nil {
		log.Fatalf("Failed to populate test data: %v", err)
	}

	// Get initial cache size
	fmt.Println("\n2. Getting initial cache size...")
	initialSize, err := redisCache.GetCacheSize(ctx)
	if err != nil {
		log.Fatalf("Failed to get cache size: %v", err)
	}
	printCacheSize("Initial", initialSize)

	// Demonstrate basic cleanup
	fmt.Println("\n3. Performing basic cleanup of test diagrams...")
	err = redisCache.Cleanup(ctx, fmt.Sprintf("/diagrams/%s/test-*", models.DiagramTypePUML.String()))
	if err != nil {
		log.Fatalf("Basic cleanup failed: %v", err)
	}
	fmt.Println("✓ Basic cleanup completed")

	// Demonstrate advanced cleanup with options
	fmt.Println("\n4. Performing advanced cleanup with options...")
	cleanupOptions := &cache.CleanupOptions{
		BatchSize:      50,
		ScanCount:      100,
		MaxKeys:        10, // Limit to 10 keys for demonstration
		DryRun:         false,
		Timeout:        30 * time.Second,
		CollectMetrics: true,
	}

	result, err := redisCache.CleanupWithOptions(ctx, "/machines/test-*", cleanupOptions)
	if err != nil {
		log.Fatalf("Advanced cleanup failed: %v", err)
	}
	printCleanupResult("Advanced cleanup", result)

	// Demonstrate dry run
	fmt.Println("\n5. Performing dry run cleanup...")
	dryRunOptions := &cache.CleanupOptions{
		BatchSize:      100,
		ScanCount:      100,
		MaxKeys:        0,
		DryRun:         true,
		Timeout:        30 * time.Second,
		CollectMetrics: true,
	}

	dryRunResult, err := redisCache.CleanupWithOptions(ctx, "/test-*", dryRunOptions)
	if err != nil {
		log.Fatalf("Dry run cleanup failed: %v", err)
	}
	printCleanupResult("Dry run", dryRunResult)

	// Get final cache size
	fmt.Println("\n6. Getting final cache size...")
	finalSize, err := redisCache.GetCacheSize(ctx)
	if err != nil {
		log.Fatalf("Failed to get final cache size: %v", err)
	}
	printCacheSize("Final", finalSize)

	// Demonstrate bulk cleanup
	fmt.Println("\n7. Performing bulk cleanup of remaining test data...")
	bulkOptions := cache.DefaultCleanupOptions()
	bulkOptions.BatchSize = 200 // Larger batch size for efficiency
	bulkOptions.CollectMetrics = true

	bulkResult, err := redisCache.CleanupWithOptions(ctx, "/test-*", bulkOptions)
	if err != nil {
		log.Fatalf("Bulk cleanup failed: %v", err)
	}
	printCleanupResult("Bulk cleanup", bulkResult)

	fmt.Println("\n✓ Cleanup example completed successfully!")
}

func populateTestData(ctx context.Context, cache cache.Cache) error {
	// Create test diagrams
	for i := 0; i < 20; i++ {
		diagramName := fmt.Sprintf("test-diagram-%d", i)
		content := fmt.Sprintf("@startuml\nstate S%d\nstate T%d\nS%d --> T%d\n@enduml", i, i, i, i)

		if err := cache.StoreDiagram(ctx, models.DiagramTypePUML, diagramName, content, time.Hour); err != nil {
			return fmt.Errorf("failed to store diagram %s: %w", diagramName, err)
		}
	}

	// Create some additional test keys with different patterns
	for i := 0; i < 15; i++ {
		key := fmt.Sprintf("test-key-%d", i)
		if err := cache.StoreDiagram(ctx, models.DiagramTypePUML, key, "test content", time.Hour); err != nil {
			return fmt.Errorf("failed to store test key %s: %w", key, err)
		}
	}

	fmt.Printf("✓ Created 35 test entries in cache")
	return nil
}

func printCacheSize(label string, size *cache.CacheSizeInfo) {
	fmt.Printf("%s Cache Size:\n", label)
	fmt.Printf("  Total Keys: %d\n", size.TotalKeys)
	fmt.Printf("  Diagrams: %d\n", size.DiagramCount)
	fmt.Printf("  State Machines: %d\n", size.StateMachineCount)
	fmt.Printf("  Entities: %d\n", size.EntityCount)
	fmt.Printf("  Memory Used: %s\n", formatBytes(size.MemoryUsed))
	fmt.Printf("  Memory Peak: %s\n", formatBytes(size.MemoryPeak))
	fmt.Printf("  Memory Overhead: %s\n", formatBytes(size.MemoryOverhead))
}

func printCleanupResult(label string, result *cache.CleanupResult) {
	fmt.Printf("%s Result:\n", label)
	fmt.Printf("  Keys Scanned: %d\n", result.KeysScanned)
	fmt.Printf("  Keys Deleted: %d\n", result.KeysDeleted)
	fmt.Printf("  Bytes Freed: %s\n", formatBytes(result.BytesFreed))
	fmt.Printf("  Duration: %v\n", result.Duration)
	fmt.Printf("  Batches Used: %d\n", result.BatchesUsed)
	fmt.Printf("  Errors: %d\n", result.ErrorsOccurred)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
