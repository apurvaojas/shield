# Logger Package

This package provides a comprehensive logging solution for the Shield project using Go's structured logging (`slog`) with advanced handler composition via `slog-multi`.

## Features

- **Multiple Log Levels**: Debug, Info, Warning, Error
- **Multiple Outputs**: Console, File, OpenTelemetry (OTEL)
- **Environment-aware**: Different configurations for development, staging, and production
- **Structured Logging**: JSON format with attributes and groups
- **Error Recovery**: Graceful handling of logging failures
- **Log Enrichment**: Automatic addition of environment and service metadata

## Environment Variables

Configure the logger using these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `ENVIRONMENT` | `development` | Environment (development, staging, production) |
| `LOG_FILE_ENABLED` | `false` | Enable file logging |
| `LOG_FILE_DIR` | `./logs` | Directory for log files |
| `OTEL_ENABLED` | `false` | Enable OpenTelemetry logging |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `` | OTEL endpoint URL |
| `OTEL_SERVICE_NAME` | `shield-api` | Service name for OTEL |

## Environment-specific Behavior

### Development
- **Console**: Text format for better readability
- **File**: Disabled by default (can be enabled with `LOG_FILE_ENABLED=true`)
- **OTEL**: Disabled
- **Source**: Include source file information

### Staging/Production
- **Console**: JSON format
- **File**: Enabled automatically with daily rotation
- **OTEL**: Enabled if configured
- **Enrichment**: Automatic addition of environment and service metadata

## Usage

### Initialize the Logger

```go
package main

import (
    "shield/modules/common"
)

func main() {
    // Initialize the logger - call this once at application startup
    if err := common.InitLogger(); err != nil {
        panic(err)
    }
    
    // Logger is now available globally via slog package
    // Your application code here...
}
```

### Basic Logging

```go
import "log/slog"

// Simple messages
slog.Info("Application started")
slog.Debug("Debug information")
slog.Warn("Warning message")
slog.Error("Error occurred")
```

### Structured Logging

```go
import "log/slog"

// With attributes
slog.Info("User login",
    slog.String("user_id", "123"),
    slog.String("ip", "192.168.1.1"),
    slog.Int("attempt", 1),
)

// With groups
slog.Info("Database operation",
    slog.Group("db",
        slog.String("operation", "SELECT"),
        slog.String("table", "users"),
        slog.Duration("duration", time.Millisecond*150),
    ),
)
```

### Contextual Logging

```go
import "log/slog"

// Create a logger with persistent attributes
logger := slog.With(
    slog.String("request_id", "abc-123"),
    slog.String("user_id", "456"),
)

logger.Info("Request started")
logger.Info("Processing data")
logger.Info("Request completed")
```

## Example Environment Configurations

### Development (.env)
```bash
LOG_LEVEL=debug
ENVIRONMENT=development
LOG_FILE_ENABLED=false
```

### Staging (.env)
```bash
LOG_LEVEL=info
ENVIRONMENT=staging
LOG_FILE_ENABLED=true
LOG_FILE_DIR=/var/log/shield
OTEL_ENABLED=true
OTEL_EXPORTER_OTLP_ENDPOINT=https://otel-collector.staging.com
OTEL_SERVICE_NAME=shield-api
```

### Production (.env)
```bash
LOG_LEVEL=warn
ENVIRONMENT=production
LOG_FILE_ENABLED=true
LOG_FILE_DIR=/var/log/shield
OTEL_ENABLED=true
OTEL_EXPORTER_OTLP_ENDPOINT=https://otel-collector.prod.com
OTEL_SERVICE_NAME=shield-api
```

## Log File Management

- Files are created daily with format: `app_YYYY-MM-DD.log`
- Files are stored in the configured `LOG_FILE_DIR`
- JSON format for easy parsing by log aggregation tools
- Automatic directory creation if it doesn't exist

## Error Handling

The logger includes built-in error recovery:
- Failed handlers don't crash the application
- Errors are logged to stderr
- Fallback to console logging if other handlers fail

## Future Enhancements

- **OTEL Integration**: Full OpenTelemetry support will be added when dependencies are available
- **Log Rotation**: Automatic log file rotation and cleanup
- **Performance Monitoring**: Built-in metrics for logging performance
- **Sampling**: Configurable log sampling for high-volume scenarios

## Dependencies

- `github.com/samber/slog-multi`: Handler composition and middleware
- `log/slog`: Go's standard structured logging (Go 1.21+)

Note: OpenTelemetry dependencies will be added in future versions for full OTEL support.
