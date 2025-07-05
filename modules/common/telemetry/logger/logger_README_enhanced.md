# Logger Package

A comprehensive logging package for the Shield project that provides structured logging with multiple output destinations, file rotation, custom formatting, and OpenTelemetry integration.

## Features

- **Multiple Log Levels**: Debug, Info, Warning, Error
- **Multiple Outputs**: Console, File, OpenTelemetry (OTEL)
- **Environment-based Configuration**: Uses environment variables for all settings
- **File Rotation**: Configurable file rotation by size, time, or both
- **Custom Formatting**: 
  - Automatic timestamps
  - Trace ID and Span ID integration
  - Stack traces for error-level logs
  - Environment-specific formatting (Text for dev, JSON for prod)
- **Structured Logging**: Full support for slog structured logging
- **Error Recovery**: Built-in error recovery for logging failures

## Environment Variables

### Basic Configuration
- `LOG_LEVEL`: Log level (debug, info, warn, error) - Default: `info`
- `ENVIRONMENT`: Environment (development, staging, production) - Default: `development`
- `LOG_FILE_ENABLED`: Enable file logging (`true`/`false`) - Default: `false`
- `LOG_FILE_DIR`: Directory for log files - Default: `./logs`

### OpenTelemetry Configuration
- `OTEL_ENABLED`: Enable OTEL logging (`true`/`false`) - Default: `false`
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTEL endpoint URL
- `OTEL_SERVICE_NAME`: Service name for OTEL - Default: `shield-api`

### File Rotation Configuration
- `LOG_ROTATION_ENABLED`: Enable file rotation (`true`/`false`) - Default: `true`
- `LOG_MAX_FILE_SIZE_MB`: Maximum file size in MB before rotation - Default: `100`
- `LOG_MAX_FILES`: Maximum number of log files to keep - Default: `5`
- `LOG_MAX_AGE_DAYS`: Maximum age of log files in days - Default: `30`
- `LOG_ROTATION_INTERVAL`: Rotation interval (`hourly`, `daily`, `size-based`) - Default: `daily`

## Usage

### Basic Setup

```go
package main

import (
    "log/slog"
    "your-project/modules/common"
)

func main() {
    // Initialize the logger (reads configuration from environment variables)
    if err := common.InitLogger(); err != nil {
        panic("Failed to initialize logger: " + err.Error())
    }

    // Use slog anywhere in your application
    slog.Info("Application started")
    slog.Debug("Debug information", slog.String("component", "main"))
    slog.Error("An error occurred", slog.String("error", "connection failed"))
}
```

### Environment-Specific Configuration

#### Development Environment
```bash
export LOG_LEVEL=debug
export ENVIRONMENT=development
export LOG_FILE_ENABLED=true
export LOG_FILE_DIR=./logs
```

#### Production Environment
```bash
export LOG_LEVEL=info
export ENVIRONMENT=production
export LOG_FILE_ENABLED=true
export LOG_FILE_DIR=/var/log/shield
export LOG_ROTATION_ENABLED=true
export LOG_MAX_FILE_SIZE_MB=100
export LOG_MAX_FILES=10
export LOG_MAX_AGE_DAYS=30
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
export OTEL_SERVICE_NAME=shield-api
```

### Advanced Usage

#### Structured Logging with Context
```go
import (
    "context"
    "log/slog"
)

func handleRequest(ctx context.Context, userID string) {
    // Logs will automatically include trace_id and span_id if available in context
    slog.InfoContext(ctx, "Processing user request",
        slog.String("user_id", userID),
        slog.String("action", "login"),
    )
}
```

#### Error Logging with Stack Traces
```go
// Error logs automatically include stack traces
slog.Error("Database connection failed",
    slog.String("database", "postgresql"),
    slog.String("error", "timeout"),
    slog.Int("retry_count", 3),
)
```

#### Contextual Logging
```go
// Create logger with persistent attributes
logger := slog.With(
    slog.String("module", "auth"),
    slog.String("version", "1.0.0"),
)

logger.Info("Module initialized")
logger.Error("Authentication failed", slog.String("reason", "invalid_token"))
```

#### Grouped Attributes
```go
// Group related attributes
logger := slog.Default().WithGroup("database")
logger.Info("Query executed",
    slog.Duration("execution_time", 150*time.Millisecond),
    slog.Int("rows_affected", 5),
)
```

## Log Output Format

### Development (Text Format)
```
time=2024-07-04T10:15:30.123Z level=INFO msg="User logged in" user_id=12345 username=john_doe
time=2024-07-04T10:15:31.456Z level=ERROR msg="Database error" error="connection timeout" stack_trace="goroutine 1..."
```

### Production (JSON Format)
```json
{
  "time": "2024-07-04T10:15:30.123Z",
  "level": "INFO",
  "msg": "User logged in",
  "user_id": "12345",
  "username": "john_doe",
  "trace_id": "abc123def456",
  "span_id": "789ghi012",
  "environment": "production",
  "service": "shield-api"
}
```

### Error Logs with Stack Trace
```json
{
  "time": "2024-07-04T10:15:31.456Z",
  "level": "ERROR",
  "msg": "Database connection failed",
  "error": "connection timeout",
  "database": "postgresql",
  "retry_count": 3,
  "stack_trace": "goroutine 1 [running]:\nruntime/debug.Stack()\n\t/usr/local/go/src/runtime/debug/stack.go:24 +0x64...",
  "trace_id": "abc123def456",
  "span_id": "789ghi012",
  "environment": "production",
  "service": "shield-api"
}
```

## File Rotation

The logger supports intelligent file rotation with multiple strategies:

- **Size-based**: Rotate when file reaches maximum size
- **Time-based**: Rotate daily or hourly
- **Age-based**: Automatically clean up old files
- **Count-based**: Keep only a specified number of recent files

### File Naming Convention
- Format: `app_YYYY-MM-DD_HH-MM-SS.log`
- Example: `app_2024-07-04_10-15-30.log`

## OpenTelemetry Integration

When OTEL is enabled, the logger:
- Extracts trace_id and span_id from context
- Includes them in all log entries
- Provides placeholder for OTEL exporter integration

## Error Handling

The logger includes robust error handling:
- Graceful fallback if file logging fails
- Error recovery middleware
- Non-blocking initialization (logs errors to stderr)

## Performance Considerations

- Uses `slog-multi` for efficient fanout to multiple handlers
- Minimal overhead for disabled log levels
- Efficient file rotation with configurable retention policies
- Stack trace capture only for error-level logs

## Thread Safety

All components are thread-safe and can be used concurrently from multiple goroutines.

## Dependencies

- `log/slog`: Standard Go structured logging
- `github.com/samber/slog-multi`: Multi-handler support
- `go.opentelemetry.io/otel/trace`: OpenTelemetry tracing integration
