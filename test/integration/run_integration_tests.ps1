# Integration Test Runner for Windows PowerShell
# This script runs Redis integration tests with proper setup

Write-Host "Starting Redis Cache Integration Tests..." -ForegroundColor Green
Write-Host ""

# Check if Go is available
try {
    $goVersion = go version
    Write-Host "Go version: $goVersion" -ForegroundColor Blue
} catch {
    Write-Host "ERROR: Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go and add it to your PATH"
    Read-Host "Press Enter to exit"
    exit 1
}

# Set test environment variables
$env:REDIS_ADDR = "localhost:6379"
$env:REDIS_DB = "15"

Write-Host "Checking Redis connection..." -ForegroundColor Yellow

# Create a simple Redis connection test
$testScript = @"
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
"@

# Write test script to temporary file
$tempFile = [System.IO.Path]::GetTempFileName() + ".go"
$testScript | Out-File -FilePath $tempFile -Encoding UTF8

try {
    $result = go run $tempFile 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Redis connection successful" -ForegroundColor Green
    } else {
        Write-Host "WARNING: Redis connection test failed" -ForegroundColor Yellow
        Write-Host "Make sure Redis is running on localhost:6379" -ForegroundColor Yellow
        Write-Host "You can start Redis using Docker: docker run -d -p 6379:6379 redis:latest" -ForegroundColor Yellow
        Write-Host ""
        $continue = Read-Host "Continue anyway? (y/n)"
        if ($continue -ne "y" -and $continue -ne "Y") {
            Write-Host "Exiting..." -ForegroundColor Red
            exit 1
        }
    }
} finally {
    Remove-Item $tempFile -ErrorAction SilentlyContinue
}

Write-Host ""
Write-Host "Running integration tests..." -ForegroundColor Green
Write-Host "================================" -ForegroundColor Cyan

# Function to run tests and check results
function Run-TestSuite {
    param(
        [string]$TestName,
        [string]$Pattern,
        [int]$TimeoutSeconds = 30
    )
    
    Write-Host "Running $TestName..." -ForegroundColor Yellow
    $result = go test -v ./test/integration -run $Pattern -timeout "${TimeoutSeconds}s" 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "$TestName passed" -ForegroundColor Green
        return $true
    } else {
        Write-Host "$TestName failed" -ForegroundColor Red
        Write-Host $result -ForegroundColor Red
        return $false
    }
}

$allPassed = $true

# Run test suites
$testSuites = @(
    @{Name="Basic Integration Tests"; Pattern="TestRedisCache_.*Integration"; Timeout=30},
    @{Name="Concurrent Integration Tests"; Pattern="TestRedisCache_Concurrent.*"; Timeout=60},
    @{Name="Thread Safety Tests"; Pattern="TestRedisCache_ThreadSafety"; Timeout=30},
    @{Name="Performance Tests"; Pattern="TestRedisCache_Performance.*"; Timeout=120},
    @{Name="High Throughput Tests"; Pattern="TestRedisCache_HighThroughput.*"; Timeout=60},
    @{Name="TTL Tests"; Pattern="TestRedisCache_.*TTL.*"; Timeout=60},
    @{Name="Cleanup Tests"; Pattern="TestRedisCache_Cleanup.*"; Timeout=30},
    @{Name="Memory Usage Tests"; Pattern="TestRedisCache_MemoryUsage"; Timeout=30},
    @{Name="Connection Resilience Tests"; Pattern="TestRedisCache_ConnectionResilience"; Timeout=30}
)

foreach ($suite in $testSuites) {
    Write-Host ""
    $passed = Run-TestSuite -TestName $suite.Name -Pattern $suite.Pattern -TimeoutSeconds $suite.Timeout
    if (-not $passed) {
        $allPassed = $false
    }
}

Write-Host ""
Write-Host "================================" -ForegroundColor Cyan

if ($allPassed) {
    Write-Host "All integration tests completed successfully!" -ForegroundColor Green
} else {
    Write-Host "Some integration tests failed!" -ForegroundColor Red
    Write-Host "Check the output above for details." -ForegroundColor Red
}

Write-Host ""

# Optional: Run benchmarks
$runBench = Read-Host "Run benchmarks? (y/n)"
if ($runBench -eq "y" -or $runBench -eq "Y") {
    Write-Host ""
    Write-Host "Running benchmarks..." -ForegroundColor Yellow
    go test -v ./test/integration -bench="BenchmarkRedisCache_.*" -benchtime=5s -timeout=300s
}

# Optional: Run stress tests
$runStress = Read-Host "Run stress tests? (y/n)"
if ($runStress -eq "y" -or $runStress -eq "Y") {
    Write-Host ""
    Write-Host "Running stress tests..." -ForegroundColor Yellow
    go test -v ./test/integration -run="TestRedisCache_StressTest" -timeout=120s
}

Write-Host ""
Write-Host "Integration test suite completed." -ForegroundColor Green

if ($allPassed) {
    exit 0
} else {
    exit 1
}