# Examples Guide

This guide provides comprehensive examples for using the go-uml-statemachine-cache library across different platforms.

## Available Examples

The `examples/` directory contains complete working examples:

- `examples/diagram_cache_example/` - Basic diagram caching operations
- `examples/state_machine_cache_example/` - State machine and entity caching
- `examples/health_monitoring_example/` - Health monitoring and diagnostics
- `examples/cleanup_example/` - Cache cleanup and management
- `examples/error_handling_example/` - Comprehensive error handling

## Running Examples

### Windows

#### Command Prompt
```cmd
cd go-uml-statemachine-cache
go run examples\diagram_cache_example\main.go
go run examples\state_machine_cache_example\main.go
go run examples\health_monitoring_example\main.go
go run examples\cleanup_example\main.go
go run examples\error_handling_example\main.go
```

#### PowerShell
```powershell
cd go-uml-statemachine-cache
go run examples\diagram_cache_example\main.go
go run examples\state_machine_cache_example\main.go
go run examples\health_monitoring_example\main.go
go run examples\cleanup_example\main.go
go run examples\error_handling_example\main.go
```

### Linux/macOS

#### Bash/Zsh
```bash
cd go-uml-statemachine-cache
go run examples/diagram_cache_example/main.go
go run examples/state_machine_cache_example/main.go
go run examples/health_monitoring_example/main.go
go run examples/cleanup_example/main.go
go run examples/error_handling_example/main.go
```

## Building and Running Examples

### Windows
```cmd
# Build all examples
go build -o diagram_example.exe .\examples\diagram_cache_example
go build -o statemachine_example.exe .\examples\state_machine_cache_example
go build -o health_example.exe .\examples\health_monitoring_example
go build -o cleanup_example.exe .\examples\cleanup_example
go build -o error_example.exe .\examples\error_handling_example

# Run built examples
.\diagram_example.exe
.\statemachine_example.exe
.\health_example.exe
.\cleanup_example.exe
.\error_example.exe
```

### Linux/macOS
```bash
# Build all examples
go build -o diagram_example ./examples/diagram_cache_example
go build -o statemachine_example ./examples/state_machine_cache_example
go build -o health_example ./examples/health_monitoring_example
go build -o cleanup_example ./examples/cleanup_example
go build -o error_example ./examples/error_handling_example

# Run built examples
./diagram_example
./statemachine_example
./health_example
./cleanup_example
./error_example
```

## Prerequisites by Platform

### Windows
- Go 1.24.4 or later
- Redis server (install via Chocolatey, WSL, or Docker)
- Windows 10/11 or Windows Server 2019/2022

### Linux
- Go 1.24.4 or later
- Redis server (install via package manager or from source)
- Modern Linux distribution (Ubuntu 18.04+, CentOS 7+, etc.)

### macOS
- Go 1.24.4 or later
- Redis server (install via Homebrew, MacPorts, or Docker)
- macOS 10.15+ (Catalina or later)

## Environment Setup

### Setting Redis Connection (Cross-Platform)

#### Using Environment Variables

**Windows (Command Prompt):**
```cmd
set REDIS_ADDR=localhost:6379
set REDIS_PASSWORD=your-password
set REDIS_DB=0
```

**Windows (PowerShell):**
```powershell
$env:REDIS_ADDR="localhost:6379"
$env:REDIS_PASSWORD="your-password"
$env:REDIS_DB="0"
```

**Linux/macOS (Bash):**
```bash
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=your-password
export REDIS_DB=0
```

#### Using Configuration Files

Create a `config.json` file in your project root:
```json
{
  "redis": {
    "addr": "localhost:6379",
    "password": "",
    "db": 0,
    "poolSize": 10,
    "dialTimeout": "5s",
    "readTimeout": "3s",
    "writeTimeout": "3s"
  }
}
```

Then load it in your application:
```go
import (
    "encoding/json"
    "os"
)

type Config struct {
    Redis struct {
        Addr         string `json:"addr"`
        Password     string `json:"password"`
        DB           int    `json:"db"`
        PoolSize     int    `json:"poolSize"`
        DialTimeout  string `json:"dialTimeout"`
        ReadTimeout  string `json:"readTimeout"`
        WriteTimeout string `json:"writeTimeout"`
    } `json:"redis"`
}

func loadConfig() (*Config, error) {
    file, err := os.Open("config.json")
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var config Config
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&config)
    return &config, err
}
```

## Platform-Specific Example Modifications

### File Path Handling

The examples work across platforms, but if you need to modify file paths:

```go
import (
    "path/filepath"
    "runtime"
)

func getExamplePath(filename string) string {
    if runtime.GOOS == "windows" {
        return filepath.Join("examples", "data", filename)
    }
    return filepath.Join("examples", "data", filename)
}
```

### Redis Connection Examples

#### Local Development (All Platforms)
```go
config := &cache.RedisConfig{
    RedisAddr:    "localhost:6379",
    RedisDB:      0,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    PoolSize:     10,
}
```

#### Docker Redis (All Platforms)
```go
config := &cache.RedisConfig{
    RedisAddr:    "localhost:6379", // Docker port mapping
    RedisDB:      0,
    DialTimeout:  10 * time.Second, // Longer timeout for container startup
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 5 * time.Second,
    PoolSize:     10,
}
```

#### Remote Redis (Production)
```go
config := &cache.RedisConfig{
    RedisAddr:     "redis.example.com:6379",
    RedisPassword: os.Getenv("REDIS_PASSWORD"),
    RedisDB:       0,
    DialTimeout:   15 * time.Second,
    ReadTimeout:   10 * time.Second,
    WriteTimeout:  10 * time.Second,
    PoolSize:      20,
}
```

## Testing Examples

### Unit Tests (No Redis Required)

**Windows:**
```cmd
go test .\examples\... -v
```

**Linux/macOS:**
```bash
go test ./examples/... -v
```

### Integration Tests (Redis Required)

**Windows:**
```cmd
# Start Redis first
redis-server
# In another terminal
go test .\examples\... -tags=integration -v
```

**Linux:**
```bash
# Start Redis first
sudo systemctl start redis-server
# Run tests
go test ./examples/... -tags=integration -v
```

**macOS:**
```bash
# Start Redis first
brew services start redis
# Run tests
go test ./examples/... -tags=integration -v
```

## Troubleshooting by Platform

### Windows
- **Redis not found**: Install via Chocolatey or use WSL
- **Path issues**: Use forward slashes or double backslashes in strings
- **Permission errors**: Run Command Prompt as Administrator if needed

### Linux
- **Redis connection refused**: Check if Redis service is running with `systemctl status redis-server`
- **Permission denied**: Ensure Redis is configured to accept connections
- **Port conflicts**: Check if port 6379 is available with `netstat -tlnp | grep 6379`

### macOS
- **Redis not found**: Install via Homebrew: `brew install redis`
- **Connection issues**: Check Redis is running with `brew services list | grep redis`
- **Firewall issues**: Ensure local connections are allowed

## Performance Considerations

### Platform-Specific Optimizations

#### Windows
- Use SSD storage for Redis data directory
- Configure Windows Defender exclusions for Redis directory
- Consider using WSL2 for better Linux compatibility

#### Linux
- Tune kernel parameters for Redis:
  ```bash
  echo 'vm.overcommit_memory = 1' >> /etc/sysctl.conf
  echo 'net.core.somaxconn = 65535' >> /etc/sysctl.conf
  sysctl -p
  ```
- Use systemd for Redis service management
- Consider using Redis with persistent storage

#### macOS
- Use Homebrew services for automatic Redis startup
- Configure Redis to use appropriate memory limits
- Consider Docker for consistent behavior across environments

For detailed usage patterns and API documentation, see the README.md and API.md files.