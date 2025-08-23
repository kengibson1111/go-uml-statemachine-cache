#!/bin/bash

# Integration Test Runner for Linux/Unix
# This script runs Redis integration tests with proper setup

echo -e "\033[32mStarting Redis Cache Integration Tests...\033[0m"
echo ""

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo -e "\033[31mERROR: Go is not installed or not in PATH\033[0m"
    echo "Please install Go and add it to your PATH"
    read -p "Press Enter to exit"
    exit 1
fi

GO_VERSION=$(go version)
echo -e "\033[34mGo version: $GO_VERSION\033[0m"

# Set test environment variables
export REDIS_ADDR="localhost:6379"
export REDIS_DB="15"

echo -e "\033[33mChecking Redis connection...\033[0m"

# Create a simple Redis connection test
TEMP_FILE=$(mktemp --suffix=.go)
cat > "$TEMP_FILE" << 'EOF'
package main

import (
    "context"
    "fmt"
    "github.com/redis/go-redis/v9"
)

func main() {
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   15,
    })
    
    _, err := client.Ping(context.Background()).Result()
    if err != nil {
        fmt.Printf("Redis connection failed: %v\n", err)
        return
    }
    fmt.Println("Redis connection successful")
}
EOF

# Test Redis connection
if go run "$TEMP_FILE" 2>&1; then
    echo -e "\033[32mRedis connection successful\033[0m"
else
    echo -e "\033[33mWARNING: Redis connection test failed\033[0m"
    echo -e "\033[33mMake sure Redis is running on localhost:6379\033[0m"
    echo -e "\033[33mYou can start Redis using Docker: docker run -d -p 6379:6379 redis:latest\033[0m"
    echo ""
    read -p "Continue anyway? (y/n): " continue_choice
    if [[ "$continue_choice" != "y" && "$continue_choice" != "Y" ]]; then
        echo -e "\033[31mExiting...\033[0m"
        rm -f "$TEMP_FILE"
        exit 1
    fi
fi

# Clean up temp file
rm -f "$TEMP_FILE"

echo ""
echo -e "\033[32mRunning integration tests...\033[0m"
echo -e "\033[36m================================\033[0m"

# Function to run tests and check results
run_test_suite() {
    local test_name="$1"
    local pattern="$2"
    local timeout_seconds="${3:-30}"
    
    echo -e "\033[33mRunning $test_name...\033[0m"
    
    if go test -v ./test/integration -run "$pattern" -timeout "${timeout_seconds}s" 2>&1; then
        echo -e "\033[32m$test_name passed\033[0m"
        return 0
    else
        echo -e "\033[31m$test_name failed\033[0m"
        return 1
    fi
}

all_passed=true

# Define test suites
declare -a test_suites=(
    "Basic Integration Tests|TestRedisCache_.*Integration|30"
    "Concurrent Integration Tests|TestRedisCache_Concurrent.*|60"
    "Thread Safety Tests|TestRedisCache_ThreadSafety|30"
    "Performance Tests|TestRedisCache_Performance.*|120"
    "High Throughput Tests|TestRedisCache_HighThroughput.*|60"
    "TTL Tests|TestRedisCache_.*TTL.*|60"
    "Cleanup Tests|TestRedisCache_Cleanup.*|30"
    "Memory Usage Tests|TestRedisCache_MemoryUsage|30"
    "Connection Resilience Tests|TestRedisCache_ConnectionResilience|30"
)

# Run test suites
for suite in "${test_suites[@]}"; do
    IFS='|' read -r name pattern timeout <<< "$suite"
    echo ""
    if ! run_test_suite "$name" "$pattern" "$timeout"; then
        all_passed=false
    fi
done

echo ""
echo -e "\033[36m================================\033[0m"

if [ "$all_passed" = true ]; then
    echo -e "\033[32mAll integration tests completed successfully!\033[0m"
else
    echo -e "\033[31mSome integration tests failed!\033[0m"
    echo -e "\033[31mCheck the output above for details.\033[0m"
fi

echo ""

# Optional: Run benchmarks
read -p "Run benchmarks? (y/n): " run_bench
if [[ "$run_bench" == "y" || "$run_bench" == "Y" ]]; then
    echo ""
    echo -e "\033[33mRunning benchmarks...\033[0m"
    go test -v ./test/integration -bench="BenchmarkRedisCache_.*" -benchtime=5s -timeout=300s
fi

# Optional: Run stress tests
read -p "Run stress tests? (y/n): " run_stress
if [[ "$run_stress" == "y" || "$run_stress" == "Y" ]]; then
    echo ""
    echo -e "\033[33mRunning stress tests...\033[0m"
    go test -v ./test/integration -run="TestRedisCache_StressTest" -timeout=120s
fi

echo ""
echo -e "\033[32mIntegration test suite completed.\033[0m"

if [ "$all_passed" = true ]; then
    exit 0
else
    exit 1
fi