# gatus-sdk

[![test](https://github.com/TwiN/gatus-sdk/workflows/test/badge.svg)](https://github.com/TwiN/gatus-sdk/actions?query=workflow%3Atest)
[![Go Report Card](https://goreportcard.com/badge/github.com/TwiN/gatus-sdk)](https://goreportcard.com/report/github.com/TwiN/gatus-sdk)
[![Go version](https://img.shields.io/github/go-mod/go-version/TwiN/gatus-sdk.svg)](https://github.com/TwiN/gatus-sdk)
[![License](https://img.shields.io/github/license/TwiN/gatus-sdk.svg)](LICENSE)

A lightweight, zero-dependency Go SDK for interacting with Gatus status page APIs.

Lost? The CLI can be found at [gatus-cli](https://github.com/TwiN/gatus-cli), while the main Gatus project is at [gatus](https://github.com/TwiN/gatus).


## Installation

```bash
go get github.com/TwiN/gatus-sdk
```

No additional dependencies required! This SDK uses only the Go standard library.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    gatus "github.com/TwiN/gatus-sdk"
)

func main() {
    // Create a new client
    client := gatus.NewClient("https://status.example.org")
    
    // Get all endpoint statuses
    statuses, err := client.GetAllEndpointStatuses(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    for _, status := range statuses {
        fmt.Printf("Endpoint: %s (Group: %s) - Key: %s\n", status.Name, status.Group, status.Key)
        
        if len(status.Results) > 0 {
            lastResult := status.Results[0]
            fmt.Printf("  Status: %d, Success: %v\n", lastResult.Status, lastResult.Success)
        }
    }
}
```

## API Reference

### Client Configuration

```go
// Create client with default settings
client := gatus.NewClient("https://status.example.com")

// Create client with custom timeout
client := gatus.NewClient("https://status.example.com", gatus.WithTimeout(10 * time.Second))

// Create client with custom user agent
client := gatus.NewClient("https://status.example.com", gatus.WithUserAgent("MyApp/1.0"))

// Create client with custom HTTP client
httpClient := &http.Client{
    Timeout: 15 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns: 50,
    },
}
client := gatus.NewClient("https://status.example.com", gatus.WithHTTPClient(httpClient))
```

### Key Generation

The SDK provides a utility function to generate endpoint keys in the format expected by Gatus:

```go
// Generate a key for an endpoint
key := gatus.GenerateEndpointKey("core", "blog-home")
fmt.Println(key) // Output: core_blog-home

// Special characters are replaced with hyphens
key = gatus.GenerateEndpointKey("api/v1", "health_check.test")
fmt.Println(key) // Output: api-v1_health-check-test

// Empty group is handled
key = gatus.GenerateEndpointKey("", "standalone")
fmt.Println(key) // Output: _standalone
```

### Getting Endpoint Statuses

```go
ctx := context.Background()

// Get all endpoint statuses
statuses, err := client.GetAllEndpointStatuses(ctx)
if err != nil {
    log.Fatal(err)
}

// Get status by key
status, err := client.GetEndpointStatusByKey(ctx, "core_blog-home")
if err != nil {
    log.Fatal(err)
}

// Get status by group and name (key is generated automatically)
status, err := client.GetEndpointStatus(ctx, "core", "blog-home")
if err != nil {
    log.Fatal(err)
}

// Check if endpoint is healthy
if len(status.Results) > 0 && status.Results[0].Success {
    fmt.Println("Endpoint is healthy")
}
```

### Uptime Information

```go
// Get uptime percentage
uptime, err := client.GetEndpointUptime(ctx, "core_blog-home", "24h")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Uptime: %.2f%%\n", uptime)

// Get detailed uptime data
uptimeData, err := client.GetEndpointUptimeData(ctx, "core_blog-home", "7d")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Uptime: %.2f%% over %s\n", uptimeData.Uptime, uptimeData.Duration)

// Valid durations: 1h, 24h, 7d, 30d
err = gatus.ValidateDuration("24h") // Returns nil
err = gatus.ValidateDuration("48h") // Returns validation error
```

### Response Time Metrics

```go
// Get response time statistics
respTimes, err := client.GetEndpointResponseTimes(ctx, "core_blog-home", "24h")
if err != nil {
    log.Fatal(err)
}

// Convert nanoseconds to milliseconds for display
fmt.Printf("Response Times:\n")
fmt.Printf("  Average: %dms\n", respTimes.Average/1000000)
fmt.Printf("  Min: %dms\n", respTimes.Min/1000000)
fmt.Printf("  Max: %dms\n", respTimes.Max/1000000)
```

### Badge URLs

Generate badge URLs for embedding in documentation or dashboards:

```go
key := "core_blog-home"

// Get uptime badge URL
uptimeBadgeURL := client.GetEndpointUptimeBadgeURL(key, "24h")
fmt.Printf("![Uptime](%s)\n", uptimeBadgeURL)

// Get health badge URL
healthBadgeURL := client.GetEndpointHealthBadgeURL(key)
fmt.Printf("![Health](%s)\n", healthBadgeURL)

// Get response time badge URL
respTimeBadgeURL := client.GetEndpointResponseTimeBadgeURL(key, "24h")
fmt.Printf("![Response Time](%s)\n", respTimeBadgeURL)
```

## Complete Examples

### Example 1: Monitor Multiple Endpoints

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    gatus "github.com/TwiN/gatus-sdk"
)

func main() {
    client := gatus.NewClient("https://status.example.org", gatus.WithTimeout(5 * time.Second))
    ctx := context.Background()
    // Define endpoints to monitor
    endpoints := []struct {
        Group string
        Name  string
    }{
        {"core", "blog-home"},
        {"services", "api"},
        {"databases", "postgres"},
    }
    for _, ep := range endpoints {
        key := gatus.GenerateEndpointKey(ep.Group, ep.Name)
        // Get status
        status, err := client.GetEndpointStatusByKey(ctx, key)
        if err != nil {
            log.Printf("Error getting status for %s: %v", key, err)
            continue
        }
        // Get uptime
        uptime, err := client.GetEndpointUptime(ctx, key, "24h")
        if err != nil {
            log.Printf("Error getting uptime for %s: %v", key, err)
            continue
        }
        // Get response times
        respTimes, err := client.GetEndpointResponseTimes(ctx, key, "24h")
        if err != nil {
            log.Printf("Error getting response times for %s: %v", key, err)
            continue
        }
        fmt.Printf("\nEndpoint: %s/%s\n", ep.Group, ep.Name)
        fmt.Printf("  Key: %s\n", key)
        fmt.Printf("  Uptime (24h): %.2f%%\n", uptime)
        fmt.Printf("  Avg Response: %dms\n", respTimes.Average/1000000)
        if len(status.Results) > 0 {
            lastResult := status.Results[0]
            fmt.Printf("  Last Check: %s\n", lastResult.Timestamp.Format(time.RFC3339))
            fmt.Printf("  Status: %d\n", lastResult.Status)
            fmt.Printf("  Success: %v\n", lastResult.Success)
            if len(lastResult.Errors) > 0 {
                fmt.Printf("  Errors: %v\n", lastResult.Errors)
            }
        }
    }
}
```

### Example 2: Generate Status Report

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    "time"
    
    gatus "github.com/TwiN/gatus-sdk"
)

func main() {
    client := gatus.NewClient("https://status.example.org")
    ctx := context.Background()
    
    // Get all endpoints
    statuses, err := client.GetAllEndpointStatuses(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // Generate markdown report
    var report strings.Builder
    report.WriteString("# Status Report\n\n")
    report.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))
    
    // Group endpoints by group
    groups := make(map[string][]gatus.EndpointStatus)
    for _, status := range statuses {
        groups[status.Group] = append(groups[status.Group], status)
    }
    
    for group, endpoints := range groups {
        if group == "" {
            report.WriteString("## Ungrouped\n\n")
        } else {
            report.WriteString(fmt.Sprintf("## %s\n\n", group))
        }
        
        report.WriteString("| Endpoint | Status | Uptime (24h) | Health |\n")
        report.WriteString("|----------|--------|--------------|--------|\n")
        
        for _, ep := range endpoints {
            // Get uptime
            uptime, _ := client.GetEndpointUptime(ctx, ep.Key, "24h")
            
            // Determine health status
            health := "🔴 Down"
            if len(ep.Results) > 0 && ep.Results[0].Success {
                if uptime >= 99.9 {
                    health = "🟢 Healthy"
                } else if uptime >= 95.0 {
                    health = "🟡 Degraded"
                } else {
                    health = "🟠 Issues"
                }
            }
            
            // Get badge URLs
            healthBadge := client.GetEndpointHealthBadgeURL(ep.Key)
            uptimeBadge := client.GetEndpointUptimeBadgeURL(ep.Key, "24h")
            
            report.WriteString(fmt.Sprintf("| %s | ![Health](%s) | ![Uptime](%s) | %s |\n", ep.Name, healthBadge, uptimeBadge, health))
        }
        
        report.WriteString("\n")
    }
    
    fmt.Println(report.String())
}
```

### Example 3: Context with Timeout

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    gatus "github.com/TwiN/gatus-sdk"
)

func main() {
    client := gatus.NewClient("https://status.example.org")
    // Create context with 2 second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    // This will timeout if the request takes more than 2 seconds
    statuses, err := client.GetAllEndpointStatuses(ctx)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            log.Fatal("Request timed out")
        }
        log.Fatal(err)
    }
    fmt.Printf("Retrieved %d endpoint statuses\n", len(statuses))
}
```

### Example 4: Error Handling

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    
    gatus "github.com/TwiN/gatus-sdk"
)

func main() {
    client := gatus.NewClient("https://status.example.org")
    ctx := context.Background()
    
    // Try to get a non-existent endpoint
    status, err := client.GetEndpointStatusByKey(ctx, "nonexistent_endpoint")
    if err != nil {
        // Check for specific error types
        var apiErr *gatus.APIError
        if errors.As(err, &apiErr) {
            fmt.Printf("API Error: Status %d - %s\n", apiErr.StatusCode, apiErr.Message)
            if apiErr.Body != "" {
                fmt.Printf("Response body: %s\n", apiErr.Body)
            }
            return
        }
        
        var valErr *gatus.ValidationError
        if errors.As(err, &valErr) {
            fmt.Printf("Validation Error: Field '%s' - %s\n", valErr.Field, valErr.Message)
            return
        }
        
        // Other error
        log.Fatal(err)
    }
    
    fmt.Printf("Endpoint: %s\n", status.Name)
}
```

## Testing

Run tests with coverage:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detection
go test -race ./...

# Run specific tests
go test -run TestGenerateEndpointKey ./...

# Verbose output
go test -v ./...
```

## Performance

The SDK is designed for performance:

- HTTP client connection pooling
- Support for gzip compression
- Efficient JSON parsing
- Minimal memory allocations
- Zero external dependencies

Benchmark results:

```bash
go test -bench=. ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Ensure tests pass with 100% coverage
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

### Development Guidelines

- Maintain 100% test coverage
- Use only Go standard library (no external dependencies)
- Follow Go idioms and best practices
- Document all exported functions
- Add examples for new features
- Run `go fmt` and `go vet` before committing

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/TwiN/gatus-sdk).

