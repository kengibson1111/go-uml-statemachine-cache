package integration

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-cache/cache"
	"github.com/kengibson1111/go-uml-statemachine-cache/internal"
	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BenchmarkRedisCache_DiagramOperations benchmarks diagram operations
func BenchmarkRedisCache_DiagramOperations(b *testing.B) {
	cache, cleanup := setupBenchmarkCache(b)
	defer cleanup()

	ctx := context.Background()
	content := generateTestContent(1000) // 1KB content

	b.ResetTimer()

	b.Run("StoreDiagram", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			diagramName := fmt.Sprintf("benchmark-store-%d", i)
			err := cache.StoreDiagram(ctx, diagramName, content, time.Hour)
			if err != nil {
				b.Fatalf("StoreDiagram failed: %v", err)
			}
		}
	})

	// Store some diagrams for retrieval benchmark
	for i := 0; i < 1000; i++ {
		diagramName := fmt.Sprintf("benchmark-get-%d", i)
		_ = cache.StoreDiagram(ctx, diagramName, content, time.Hour)
	}

	b.Run("GetDiagram", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			diagramName := fmt.Sprintf("benchmark-get-%d", i%1000)
			_, err := cache.GetDiagram(ctx, diagramName)
			if err != nil {
				b.Fatalf("GetDiagram failed: %v", err)
			}
		}
	})
}

// BenchmarkRedisCache_StateMachineOperations benchmarks state machine operations
func BenchmarkRedisCache_StateMachineOperations(b *testing.B) {
	cache, cleanup := setupBenchmarkCache(b)
	defer cleanup()

	ctx := context.Background()
	umlVersion := "2.0"

	// Pre-store diagrams
	for i := 0; i < 1000; i++ {
		diagramName := fmt.Sprintf("benchmark-sm-%d", i)
		content := fmt.Sprintf("@startuml\nstate S%d\n@enduml", i)
		_ = cache.StoreDiagram(ctx, diagramName, content, time.Hour)
	}

	// Create a test state machine
	createStateMachine := func(id int) *models.StateMachine {
		return &models.StateMachine{
			ID:      fmt.Sprintf("benchmark-sm-%d", id),
			Name:    fmt.Sprintf("benchmark-sm-%d", id),
			Version: umlVersion,
			Regions: []*models.Region{
				{
					ID:   fmt.Sprintf("region-%d", id),
					Name: fmt.Sprintf("Region%d", id),
					States: []*models.State{
						{
							Vertex: models.Vertex{
								ID:   fmt.Sprintf("state-%d", id),
								Name: fmt.Sprintf("State%d", id),
								Type: "state",
							},
							IsSimple: true,
						},
					},
					Transitions: []*models.Transition{
						{
							ID:   fmt.Sprintf("transition-%d", id),
							Name: fmt.Sprintf("Transition%d", id),
							Kind: models.TransitionKindExternal,
						},
					},
				},
			},
		}
	}

	b.ResetTimer()

	b.Run("StoreStateMachine", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			diagramName := fmt.Sprintf("benchmark-sm-%d", i%1000)
			stateMachine := createStateMachine(i)
			err := cache.StoreStateMachine(ctx, umlVersion, diagramName, stateMachine, time.Hour)
			if err != nil {
				b.Fatalf("StoreStateMachine failed: %v", err)
			}
		}
	})

	// Store some state machines for retrieval benchmark
	for i := 0; i < 100; i++ {
		diagramName := fmt.Sprintf("benchmark-sm-%d", i)
		stateMachine := createStateMachine(i)
		_ = cache.StoreStateMachine(ctx, umlVersion, diagramName, stateMachine, time.Hour)
	}

	b.Run("GetStateMachine", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			diagramName := fmt.Sprintf("benchmark-sm-%d", i%100)
			_, err := cache.GetStateMachine(ctx, umlVersion, diagramName)
			if err != nil {
				b.Fatalf("GetStateMachine failed: %v", err)
			}
		}
	})
}

// TestRedisCache_HighThroughputOperations tests high throughput scenarios
func TestRedisCache_HighThroughputOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high throughput test in short mode")
	}

	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	duration := 5 * time.Second
	numWorkers := 20

	var totalOperations int64
	var totalErrors int64
	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	// Start timer
	go func() {
		time.Sleep(duration)
		close(stopChan)
	}()

	// Launch workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			var operations int64
			var errors int64

			for {
				select {
				case <-stopChan:
					atomic.AddInt64(&totalOperations, operations)
					atomic.AddInt64(&totalErrors, errors)
					return
				default:
					// Perform a batch of operations
					batchSize := 10
					for j := 0; j < batchSize; j++ {
						diagramName := fmt.Sprintf("throughput-test-%d-%d", workerID, operations+int64(j))
						content := fmt.Sprintf("@startuml\nstate Throughput%d_%d\n@enduml", workerID, operations+int64(j))

						// Store
						err := cache.StoreDiagram(ctx, diagramName, content, time.Minute)
						if err != nil {
							errors++
							continue
						}

						// Get
						_, err = cache.GetDiagram(ctx, diagramName)
						if err != nil {
							errors++
							continue
						}
					}
					operations += int64(batchSize * 2) // Count both store and get operations
				}
			}
		}(i)
	}

	wg.Wait()

	totalOps := atomic.LoadInt64(&totalOperations)
	totalErrs := atomic.LoadInt64(&totalErrors)

	throughput := float64(totalOps) / duration.Seconds()
	errorRate := float64(totalErrs) / float64(totalOps) * 100

	t.Logf("High throughput test results:")
	t.Logf("  Duration: %v", duration)
	t.Logf("  Workers: %d", numWorkers)
	t.Logf("  Total operations: %d", totalOps)
	t.Logf("  Total errors: %d", totalErrs)
	t.Logf("  Throughput: %.2f ops/sec", throughput)
	t.Logf("  Error rate: %.2f%%", errorRate)

	// Assert reasonable performance
	assert.Greater(t, throughput, 100.0, "Throughput should be at least 100 ops/sec")
	assert.Less(t, errorRate, 5.0, "Error rate should be less than 5%")
}

// TestRedisCache_LargePayloadHandling tests handling of large payloads
func TestRedisCache_LargePayloadHandling(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	testCases := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content := generateTestContent(tc.size)
			diagramName := fmt.Sprintf("large-payload-%s", tc.name)

			start := time.Now()

			// Store large payload
			err := cache.StoreDiagram(ctx, diagramName, content, time.Hour)
			require.NoError(t, err)
			storeTime := time.Since(start)

			// Retrieve large payload
			start = time.Now()
			retrieved, err := cache.GetDiagram(ctx, diagramName)
			require.NoError(t, err)
			getTime := time.Since(start)

			// Verify content
			assert.Equal(t, len(content), len(retrieved))
			assert.Equal(t, content, retrieved)

			t.Logf("Large payload %s: Store=%v, Get=%v", tc.name, storeTime, getTime)

			// Performance expectations (adjust based on your requirements)
			if tc.size <= 100*1024 { // Up to 100KB
				assert.Less(t, storeTime, 500*time.Millisecond, "Store should be fast for %s", tc.name)
				assert.Less(t, getTime, 200*time.Millisecond, "Get should be fast for %s", tc.name)
			} else { // 1MB
				assert.Less(t, storeTime, 2*time.Second, "Store should complete within 2s for %s", tc.name)
				assert.Less(t, getTime, 1*time.Second, "Get should complete within 1s for %s", tc.name)
			}
		})
	}
}

// TestRedisCache_TTLPerformance tests TTL-related performance
func TestRedisCache_TTLPerformance(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	numItems := 1000

	// Test storing items with various TTL values
	ttlValues := []time.Duration{
		time.Second,
		time.Minute,
		time.Hour,
		24 * time.Hour,
	}

	for _, ttl := range ttlValues {
		t.Run(fmt.Sprintf("TTL_%v", ttl), func(t *testing.T) {
			start := time.Now()

			// Store items with specific TTL
			for i := 0; i < numItems; i++ {
				diagramName := fmt.Sprintf("ttl-test-%v-%d", ttl, i)
				content := fmt.Sprintf("@startuml\nstate TTL%d\n@enduml", i)

				err := cache.StoreDiagram(ctx, diagramName, content, ttl)
				require.NoError(t, err)
			}

			storeTime := time.Since(start)
			t.Logf("Stored %d items with TTL %v in %v", numItems, ttl, storeTime)

			// Verify items exist
			start = time.Now()
			for i := 0; i < numItems; i++ {
				diagramName := fmt.Sprintf("ttl-test-%v-%d", ttl, i)
				_, err := cache.GetDiagram(ctx, diagramName)
				require.NoError(t, err)
			}
			getTime := time.Since(start)
			t.Logf("Retrieved %d items with TTL %v in %v", numItems, ttl, getTime)
		})
	}

	// Test TTL expiration for short TTL
	t.Run("TTL_Expiration", func(t *testing.T) {
		shortTTL := 200 * time.Millisecond
		diagramName := "ttl-expiration-test"
		content := "@startuml\nstate Expiring\n@enduml"

		// Store with short TTL
		err := cache.StoreDiagram(ctx, diagramName, content, shortTTL)
		require.NoError(t, err)

		// Verify it exists
		_, err = cache.GetDiagram(ctx, diagramName)
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(shortTTL + 100*time.Millisecond)

		// Verify it's gone
		_, err = cache.GetDiagram(ctx, diagramName)
		require.Error(t, err)
		assert.True(t, internal.IsNotFoundError(err))
	})
}

// TestRedisCache_ConcurrentTTLOperations tests concurrent operations with TTL
func TestRedisCache_ConcurrentTTLOperations(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	numGoroutines := 10
	itemsPerGoroutine := 50

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*itemsPerGoroutine)

	// Different TTL values for different goroutines
	ttlValues := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		time.Second,
		2 * time.Second,
		5 * time.Second,
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			ttl := ttlValues[goroutineID%len(ttlValues)]

			for j := 0; j < itemsPerGoroutine; j++ {
				diagramName := fmt.Sprintf("concurrent-ttl-%d-%d", goroutineID, j)
				content := fmt.Sprintf("@startuml\nstate ConcurrentTTL%d_%d\n@enduml", goroutineID, j)

				// Store with specific TTL
				err := cache.StoreDiagram(ctx, diagramName, content, ttl)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, item %d: store failed: %w", goroutineID, j, err)
					continue
				}

				// Immediately try to retrieve
				_, err = cache.GetDiagram(ctx, diagramName)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, item %d: immediate get failed: %w", goroutineID, j, err)
					continue
				}

				// For short TTLs, test expiration
				if ttl <= 500*time.Millisecond {
					time.Sleep(ttl + 100*time.Millisecond)
					_, err = cache.GetDiagram(ctx, diagramName)
					if err == nil || !internal.IsNotFoundError(err) {
						errors <- fmt.Errorf("goroutine %d, item %d: expected expiration but item still exists", goroutineID, j)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		t.Errorf("Concurrent TTL operations failed with %d errors:", len(errorList))
		for i, err := range errorList {
			if i < 10 {
				t.Errorf("  Error %d: %v", i+1, err)
			}
		}
		if len(errorList) > 10 {
			t.Errorf("  ... and %d more errors", len(errorList)-10)
		}
	}
}

// TestRedisCache_CleanupPerformance tests cleanup operation performance
func TestRedisCache_CleanupPerformance(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	numItems := 1000

	// Store many items
	for i := 0; i < numItems; i++ {
		diagramName := fmt.Sprintf("cleanup-test-%d", i)
		content := fmt.Sprintf("@startuml\nstate Cleanup%d\n@enduml", i)

		err := cache.StoreDiagram(ctx, diagramName, content, time.Hour)
		require.NoError(t, err)
	}

	// Test cleanup performance
	start := time.Now()
	err := cache.Cleanup(ctx, "/diagrams/puml/cleanup-test-*")
	require.NoError(t, err)
	cleanupTime := time.Since(start)

	t.Logf("Cleaned up %d items in %v", numItems, cleanupTime)

	// Verify items are gone
	for i := 0; i < 10; i++ { // Check first 10 items
		diagramName := fmt.Sprintf("cleanup-test-%d", i)
		_, err := cache.GetDiagram(ctx, diagramName)
		assert.Error(t, err)
		assert.True(t, internal.IsNotFoundError(err))
	}

	// Performance expectation
	assert.Less(t, cleanupTime, 5*time.Second, "Cleanup should complete within 5 seconds")
}

// setupBenchmarkCache creates a cache for benchmarking
func setupBenchmarkCache(b *testing.B) (*cache.RedisCache, func()) {
	config := internal.DefaultConfig()
	config.RedisDB = 15 // Use test database
	config.DefaultTTL = time.Hour

	cache, err := cache.NewRedisCache(config)
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}

	// Test Redis connection
	ctx := context.Background()
	err = cache.Health(ctx)
	if err != nil {
		b.Skip("Redis not available for benchmarking:", err)
	}

	// Cleanup function
	cleanup := func() {
		_ = cache.Cleanup(context.Background(), "/diagrams/puml/*")
		_ = cache.Cleanup(context.Background(), "/machines/*")
		_ = cache.Close()
	}

	return cache, cleanup
}
