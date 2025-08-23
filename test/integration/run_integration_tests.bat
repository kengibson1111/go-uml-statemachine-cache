@echo off
REM Integration Test Runner for Windows
REM This script runs Redis integration tests with proper setup

echo Starting Redis Cache Integration Tests...
echo.

REM Check if Go is available
go version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go and add it to your PATH
    pause
    exit /b 1
)

REM Set test environment variables
set REDIS_ADDR=localhost:6379
set REDIS_DB=15

echo Checking Redis connection...
REM Try to connect to Redis using a simple Go program
go run -c "package main; import (\"github.com/redis/go-redis/v9\", \"context\", \"fmt\"); func main() { client := redis.NewClient(&redis.Options{Addr: \"localhost:6379\", DB: 15}); _, err := client.Ping(context.Background()).Result(); if err != nil { fmt.Printf(\"Redis connection failed: %v\n\", err); return }; fmt.Println(\"Redis connection successful\") }" 2>nul
if errorlevel 1 (
    echo WARNING: Redis connection test failed
    echo Make sure Redis is running on localhost:6379
    echo You can start Redis using Docker: docker run -d -p 6379:6379 redis:latest
    echo.
    echo Continue anyway? (y/n)
    set /p continue=
    if /i not "%continue%"=="y" (
        echo Exiting...
        pause
        exit /b 1
    )
)

echo.
echo Running integration tests...
echo ================================

REM Run basic integration tests
echo Running basic integration tests...
go test -v ./test/integration -run "TestRedisCache_.*Integration" -timeout 30s
if errorlevel 1 (
    echo Basic integration tests failed
    goto :error
)

echo.
echo Running concurrent integration tests...
go test -v ./test/integration -run "TestRedisCache_Concurrent.*" -timeout 60s
if errorlevel 1 (
    echo Concurrent integration tests failed
    goto :error
)

echo.
echo Running thread safety tests...
go test -v ./test/integration -run "TestRedisCache_ThreadSafety" -timeout 30s
if errorlevel 1 (
    echo Thread safety tests failed
    goto :error
)

echo.
echo Running performance tests...
go test -v ./test/integration -run "TestRedisCache_Performance.*" -timeout 120s
if errorlevel 1 (
    echo Performance tests failed
    goto :error
)

echo.
echo Running high throughput tests...
go test -v ./test/integration -run "TestRedisCache_HighThroughput.*" -timeout 60s
if errorlevel 1 (
    echo High throughput tests failed
    goto :error
)

echo.
echo Running TTL tests...
go test -v ./test/integration -run "TestRedisCache_.*TTL.*" -timeout 60s
if errorlevel 1 (
    echo TTL tests failed
    goto :error
)

echo.
echo Running cleanup tests...
go test -v ./test/integration -run "TestRedisCache_Cleanup.*" -timeout 30s
if errorlevel 1 (
    echo Cleanup tests failed
    goto :error
)

echo.
echo Running memory usage tests...
go test -v ./test/integration -run "TestRedisCache_MemoryUsage" -timeout 30s
if errorlevel 1 (
    echo Memory usage tests failed
    goto :error
)

echo.
echo ================================
echo All integration tests completed successfully!
echo.

REM Optional: Run benchmarks
echo Run benchmarks? (y/n)
set /p runbench=
if /i "%runbench%"=="y" (
    echo.
    echo Running benchmarks...
    go test -v ./test/integration -bench=BenchmarkRedisCache_.* -benchtime=5s -timeout 300s
)

echo.
echo Integration test suite completed.
pause
exit /b 0

:error
echo.
echo ================================
echo Integration tests failed!
echo Check the output above for details.
echo.
pause
exit /b 1