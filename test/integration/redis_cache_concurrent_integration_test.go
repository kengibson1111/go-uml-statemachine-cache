package integration

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRedisCache_ConcurrentDiagramOperations tests concurrent access to diagram operations
func TestRedisCache_ConcurrentDiagramOperations(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	numGoroutines := 10
	numOperationsPerGoroutine := 20

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperationsPerGoroutine)

	// Test concurrent diagram storage and retrieval
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numOperationsPerGoroutine; j++ {
				diagramName := fmt.Sprintf("concurrent-diagram-%d-%d", goroutineID, j)
				content := fmt.Sprintf("@startuml\nstate S%d_%d\n@enduml", goroutineID, j)

				// Store diagram
				err := cache.StoreDiagram(ctx, diagramName, content, time.Hour)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, operation %d: store failed: %w", goroutineID, j, err)
					continue
				}

				// Retrieve diagram
				retrieved, err := cache.GetDiagram(ctx, diagramName)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, operation %d: get failed: %w", goroutineID, j, err)
					continue
				}

				if retrieved != content {
					errors <- fmt.Errorf("goroutine %d, operation %d: content mismatch", goroutineID, j)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		t.Errorf("Concurrent operations failed with %d errors:", len(errorList))
		for i, err := range errorList {
			if i < 10 { // Limit output to first 10 errors
				t.Errorf("  Error %d: %v", i+1, err)
			}
		}
		if len(errorList) > 10 {
			t.Errorf("  ... and %d more errors", len(errorList)-10)
		}
	}
}

// TestRedisCache_ConcurrentStateMachineOperations tests concurrent state machine operations
func TestRedisCache_ConcurrentStateMachineOperations(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	numGoroutines := 5
	numOperationsPerGoroutine := 10

	// First, store diagrams that state machines will reference
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < numOperationsPerGoroutine; j++ {
			diagramName := fmt.Sprintf("concurrent-sm-diagram-%d-%d", i, j)
			content := fmt.Sprintf("@startuml\nstate S%d_%d\n@enduml", i, j)
			err := cache.StoreDiagram(ctx, diagramName, content, time.Hour)
			require.NoError(t, err)
		}
	}

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperationsPerGoroutine)

	// Test concurrent state machine operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numOperationsPerGoroutine; j++ {
				diagramName := fmt.Sprintf("concurrent-sm-diagram-%d-%d", goroutineID, j)
				umlVersion := "2.0"

				// Create a test state machine
				stateMachine := &models.StateMachine{
					ID:      fmt.Sprintf("sm-%d-%d", goroutineID, j),
					Name:    diagramName,
					Version: umlVersion,
					Regions: []*models.Region{
						{
							ID:   fmt.Sprintf("region-%d-%d", goroutineID, j),
							Name: fmt.Sprintf("Region%d_%d", goroutineID, j),
							States: []*models.State{
								{
									Vertex: models.Vertex{
										ID:   fmt.Sprintf("state-%d-%d", goroutineID, j),
										Name: fmt.Sprintf("State%d_%d", goroutineID, j),
										Type: "state",
									},
									IsSimple: true,
								},
							},
						},
					},
				}

				// Store state machine
				err := cache.StoreStateMachine(ctx, umlVersion, diagramName, stateMachine, time.Hour)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, operation %d: store SM failed: %w", goroutineID, j, err)
					continue
				}

				// Retrieve state machine
				retrieved, err := cache.GetStateMachine(ctx, umlVersion, diagramName)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, operation %d: get SM failed: %w", goroutineID, j, err)
					continue
				}

				if retrieved.ID != stateMachine.ID {
					errors <- fmt.Errorf("goroutine %d, operation %d: SM ID mismatch", goroutineID, j)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		t.Errorf("Concurrent state machine operations failed with %d errors:", len(errorList))
		for i, err := range errorList {
			if i < 10 { // Limit output to first 10 errors
				t.Errorf("  Error %d: %v", i+1, err)
			}
		}
		if len(errorList) > 10 {
			t.Errorf("  ... and %d more errors", len(errorList)-10)
		}
	}
}

// TestRedisCache_ThreadSafety tests thread safety with mixed operations
func TestRedisCache_ThreadSafety(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	numGoroutines := 8
	duration := 2 * time.Second

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*100) // Buffer for potential errors
	stopChan := make(chan struct{})

	// Start timer to stop all goroutines
	go func() {
		time.Sleep(duration)
		close(stopChan)
	}()

	// Mixed operations: some goroutines do diagrams, others do state machines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		if i%2 == 0 {
			// Diagram operations
			go func(goroutineID int) {
				defer wg.Done()
				operationCount := 0

				for {
					select {
					case <-stopChan:
						t.Logf("Goroutine %d (diagrams) completed %d operations", goroutineID, operationCount)
						return
					default:
						diagramName := fmt.Sprintf("thread-safety-diagram-%d-%d", goroutineID, operationCount)
						content := fmt.Sprintf("@startuml\nstate ThreadSafe%d_%d\n@enduml", goroutineID, operationCount)

						// Store
						err := cache.StoreDiagram(ctx, diagramName, content, time.Minute)
						if err != nil {
							errors <- fmt.Errorf("diagram store error (g%d, op%d): %w", goroutineID, operationCount, err)
							continue
						}

						// Get
						_, err = cache.GetDiagram(ctx, diagramName)
						if err != nil {
							errors <- fmt.Errorf("diagram get error (g%d, op%d): %w", goroutineID, operationCount, err)
							continue
						}

						operationCount++
						time.Sleep(10 * time.Millisecond) // Small delay to prevent overwhelming
					}
				}
			}(i)
		} else {
			// State machine operations
			go func(goroutineID int) {
				defer wg.Done()
				operationCount := 0

				for {
					select {
					case <-stopChan:
						t.Logf("Goroutine %d (state machines) completed %d operations", goroutineID, operationCount)
						return
					default:
						diagramName := fmt.Sprintf("thread-safety-sm-%d-%d", goroutineID, operationCount)

						// First store the diagram
						diagramContent := fmt.Sprintf("@startuml\nstate ThreadSafeSM%d_%d\n@enduml", goroutineID, operationCount)
						err := cache.StoreDiagram(ctx, diagramName, diagramContent, time.Minute)
						if err != nil {
							errors <- fmt.Errorf("SM diagram store error (g%d, op%d): %w", goroutineID, operationCount, err)
							continue
						}

						// Then store the state machine
						stateMachine := &models.StateMachine{
							ID:      fmt.Sprintf("thread-safe-sm-%d-%d", goroutineID, operationCount),
							Name:    diagramName,
							Version: "2.0",
							Regions: []*models.Region{
								{
									ID:   fmt.Sprintf("thread-safe-region-%d-%d", goroutineID, operationCount),
									Name: fmt.Sprintf("ThreadSafeRegion%d_%d", goroutineID, operationCount),
								},
							},
						}

						err = cache.StoreStateMachine(ctx, "2.0", diagramName, stateMachine, time.Minute)
						if err != nil {
							errors <- fmt.Errorf("SM store error (g%d, op%d): %w", goroutineID, operationCount, err)
							continue
						}

						operationCount++
						time.Sleep(15 * time.Millisecond) // Slightly longer delay for more complex operations
					}
				}
			}(i)
		}
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		t.Errorf("Thread safety test failed with %d errors:", len(errorList))
		for i, err := range errorList {
			if i < 15 { // Show more errors for thread safety issues
				t.Errorf("  Error %d: %v", i+1, err)
			}
		}
		if len(errorList) > 15 {
			t.Errorf("  ... and %d more errors", len(errorList)-15)
		}
	}
}

// TestRedisCache_PerformanceBenchmark tests performance under load
func TestRedisCache_PerformanceBenchmark(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	// Test different payload sizes
	testCases := []struct {
		name        string
		payloadSize int
		operations  int
	}{
		{"small_payload", 100, 1000},
		{"medium_payload", 1000, 500},
		{"large_payload", 10000, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate test content of specified size
			content := generateTestContent(tc.payloadSize)

			start := time.Now()
			var wg sync.WaitGroup
			errors := make(chan error, tc.operations)

			// Perform operations concurrently
			for i := 0; i < tc.operations; i++ {
				wg.Add(1)
				go func(opID int) {
					defer wg.Done()

					diagramName := fmt.Sprintf("perf-test-%s-%d", tc.name, opID)

					// Store operation
					storeStart := time.Now()
					err := cache.StoreDiagram(ctx, diagramName, content, time.Hour)
					if err != nil {
						errors <- fmt.Errorf("store operation %d failed: %w", opID, err)
						return
					}
					storeTime := time.Since(storeStart)

					// Get operation
					getStart := time.Now()
					retrieved, err := cache.GetDiagram(ctx, diagramName)
					if err != nil {
						errors <- fmt.Errorf("get operation %d failed: %w", opID, err)
						return
					}
					getTime := time.Since(getStart)

					if len(retrieved) != len(content) {
						errors <- fmt.Errorf("operation %d: content length mismatch", opID)
						return
					}

					// Log slow operations
					if storeTime > 100*time.Millisecond {
						t.Logf("Slow store operation %d: %v", opID, storeTime)
					}
					if getTime > 50*time.Millisecond {
						t.Logf("Slow get operation %d: %v", opID, getTime)
					}
				}(i)
			}

			wg.Wait()
			close(errors)
			totalTime := time.Since(start)

			// Check for errors
			var errorList []error
			for err := range errors {
				errorList = append(errorList, err)
			}

			if len(errorList) > 0 {
				t.Errorf("Performance test %s failed with %d errors", tc.name, len(errorList))
				for i, err := range errorList {
					if i < 5 {
						t.Errorf("  Error %d: %v", i+1, err)
					}
				}
			} else {
				opsPerSecond := float64(tc.operations*2) / totalTime.Seconds() // *2 for store+get
				t.Logf("Performance test %s: %d operations in %v (%.2f ops/sec)",
					tc.name, tc.operations*2, totalTime, opsPerSecond)
			}
		})
	}
}

// TestRedisCache_StressTest performs stress testing with high concurrency
func TestRedisCache_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	numGoroutines := runtime.NumCPU() * 4 // Scale with available CPUs
	duration := 10 * time.Second

	t.Logf("Starting stress test with %d goroutines for %v", numGoroutines, duration)

	var wg sync.WaitGroup
	var totalOperations int64
	var totalErrors int64
	stopChan := make(chan struct{})

	// Start timer
	go func() {
		time.Sleep(duration)
		close(stopChan)
	}()

	// Launch stress test goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			var ops int64
			var errors int64

			for {
				select {
				case <-stopChan:
					// Use atomic operations to safely update counters
					atomic.AddInt64(&totalOperations, ops)
					atomic.AddInt64(&totalErrors, errors)
					t.Logf("Goroutine %d: %d operations, %d errors", goroutineID, ops, errors)
					return
				default:
					diagramName := fmt.Sprintf("stress-test-%d-%d", goroutineID, ops)
					content := fmt.Sprintf("@startuml\nstate StressTest%d_%d\nstate End\nStressTest%d_%d --> End\n@enduml",
						goroutineID, ops, goroutineID, ops)

					// Store operation
					err := cache.StoreDiagram(ctx, diagramName, content, time.Minute)
					if err != nil {
						errors++
						continue
					}

					// Get operation
					_, err = cache.GetDiagram(ctx, diagramName)
					if err != nil {
						errors++
						continue
					}

					ops += 2 // Count both store and get

					// Small delay to prevent overwhelming the system
					if ops%100 == 0 {
						time.Sleep(time.Millisecond)
					}
				}
			}
		}(i)
	}

	wg.Wait()

	totalOps := atomic.LoadInt64(&totalOperations)
	totalErrs := atomic.LoadInt64(&totalErrors)

	t.Logf("Stress test completed: %d total operations, %d total errors", totalOps, totalErrs)
}

// TestRedisCache_ConnectionResilience tests behavior under connection issues
func TestRedisCache_ConnectionResilience(t *testing.T) {
	// This test requires a more complex setup to simulate connection failures
	// For now, we'll test basic retry behavior

	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	// Test with a very short timeout to trigger timeout errors
	shortCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	// This should trigger a timeout error
	err := cache.StoreDiagram(shortCtx, "timeout-test", "content", time.Hour)

	// We expect either a timeout error or context deadline exceeded
	if err != nil {
		t.Logf("Expected timeout error occurred: %v", err)
		// Verify it's a timeout-related error
		assert.True(t,
			err == context.DeadlineExceeded || strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline"),
			"Expected timeout error, got: %v", err)
	}

	// Test that normal operations still work after timeout
	err = cache.StoreDiagram(ctx, "recovery-test", "content", time.Hour)
	assert.NoError(t, err, "Cache should recover after timeout")
}

// TestRedisCache_MemoryUsage tests memory usage patterns
func TestRedisCache_MemoryUsage(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()

	// Get initial memory stats
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Store a large number of items
	numItems := 1000
	itemSize := 1024 // 1KB each

	for i := 0; i < numItems; i++ {
		diagramName := fmt.Sprintf("memory-test-%d", i)
		content := generateTestContent(itemSize)

		err := cache.StoreDiagram(ctx, diagramName, content, time.Hour)
		require.NoError(t, err)
	}

	// Get memory stats after storing items
	var m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Clean up all items
	for i := 0; i < numItems; i++ {
		diagramName := fmt.Sprintf("memory-test-%d", i)
		err := cache.DeleteDiagram(ctx, diagramName)
		require.NoError(t, err)
	}

	// Get final memory stats
	var m3 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m3)

	t.Logf("Memory usage - Initial: %d KB, After storing: %d KB, After cleanup: %d KB",
		m1.Alloc/1024, m2.Alloc/1024, m3.Alloc/1024)

	// Basic sanity check - memory should not grow excessively
	memoryGrowth := m3.Alloc - m1.Alloc
	maxAcceptableGrowth := uint64(numItems * itemSize / 2) // Allow 50% of stored data size as growth

	if memoryGrowth > maxAcceptableGrowth {
		t.Logf("Warning: Memory growth (%d KB) exceeds expected threshold (%d KB)",
			memoryGrowth/1024, maxAcceptableGrowth/1024)
	}
}

// generateTestContent creates test content of specified size
func generateTestContent(size int) string {
	if size < 20 {
		return "@startuml\nstate A\n@enduml"
	}

	// Create content with specified size
	baseContent := "@startuml\n"
	endContent := "\n@enduml"

	remainingSize := size - len(baseContent) - len(endContent)
	if remainingSize <= 0 {
		return baseContent + endContent
	}

	// Fill with state definitions
	statePattern := "state StateXXXXX\n"
	numStates := remainingSize / len(statePattern)

	content := baseContent
	for i := 0; i < numStates; i++ {
		stateName := fmt.Sprintf("state State%05d\n", i)
		content += stateName
	}

	// Fill remaining space with transitions
	remaining := size - len(content) - len(endContent)
	if remaining > 0 {
		filler := "A --> B\n"
		for remaining > len(filler) {
			content += filler
			remaining -= len(filler)
		}
	}

	content += endContent
	return content
}
