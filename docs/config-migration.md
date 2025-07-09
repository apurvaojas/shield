# Configuration Migration Guide

## Overview
This document shows the migration from environment variable-based configuration to structured YAML configuration.

## Migration Mapping

### Server Configuration
| Old Environment Variable | New YAML Path | Description |
|--------------------------|---------------|-------------|
| `PORT` | `server.port` | Server port |
| `ENVIRONMENT` | `server.environment` | Environment name |
| `DEBUG` | `server.debug` | Debug mode flag |
| `ALLOWED_HOSTS` | `security.trustedProxies` | Trusted proxy addresses |
| `SERVER_TIMEZONE` | `server.timezone` | Server timezone |

### Database Configuration
| Old Environment Variable | New YAML Path | Description |
|--------------------------|---------------|-------------|
| `DB_HOST` | `database.host` | Database host |
| `DB_PORT` | `database.port` | Database port |
| `DB_USER` | `database.user` | Database username |
| `DB_PASSWORD` | `database.password` | Database password |
| `DB_NAME` | `database.name` | Database name |
| `DB_SSLMODE` | `database.sslMode` | SSL mode |

### Logging Instrumentation
| Old Environment Variable | New YAML Path | Description |
|--------------------------|---------------|-------------|
| `LOG_WITH_REQUEST_BODY` | `instrumentation.logging.withRequestBody` | Log request body |
| `LOG_WITH_RESPONSE_BODY` | `instrumentation.logging.withResponseBody` | Log response body |
| `LOG_WITH_REQUEST_HEADER` | `instrumentation.logging.withRequestHeader` | Log request headers |
| `LOG_WITH_RESPONSE_HEADER` | `instrumentation.logging.withResponseHeader` | Log response headers |
| `LOG_WITH_USER_AGENT` | `instrumentation.logging.withUserAgent` | Log user agent |
| `LOG_WITH_REQUEST_ID` | `instrumentation.logging.withRequestId` | Log request ID |
| `LOG_WITH_SPAN_ID` | `instrumentation.logging.withSpanId` | Log span ID |
| `LOG_WITH_TRACE_ID` | `instrumentation.logging.withTraceId` | Log trace ID |
| `LOG_SKIP_PATHS` | `instrumentation.logging.skipPaths` | Paths to skip logging |
| `LOG_DEFAULT_LEVEL` | `instrumentation.logging.defaultLevel` | Default log level |
| `LOG_CLIENT_ERROR_LEVEL` | `instrumentation.logging.clientErrorLevel` | Client error log level |
| `LOG_SERVER_ERROR_LEVEL` | `instrumentation.logging.serverErrorLevel` | Server error log level |

### OpenTelemetry Configuration
| Old Environment Variable | New YAML Path | Description |
|--------------------------|---------------|-------------|
| `OTEL_SERVICE_NAME` | `instrumentation.openTelemetry.serviceName` | Service name |
| `OTEL_ENABLE_TRACING` | `instrumentation.openTelemetry.enableTracing` | Enable tracing |
| `OTEL_ENABLE_METRICS` | `instrumentation.openTelemetry.enableMetrics` | Enable metrics |
| `OTEL_WITH_SPAN_ID` | `instrumentation.openTelemetry.withSpanId` | Include span ID |
| `OTEL_WITH_TRACE_ID` | `instrumentation.openTelemetry.withTraceId` | Include trace ID |
| `OTEL_WITH_USER_AGENT` | `instrumentation.openTelemetry.withUserAgent` | Include user agent |
| `OTEL_WITH_REQUEST_BODY` | `instrumentation.openTelemetry.withRequestBody` | Include request body |
| `OTEL_WITH_RESPONSE_BODY` | `instrumentation.openTelemetry.withResponseBody` | Include response body |
| `OTEL_WITH_REQUEST_HEADER` | `instrumentation.openTelemetry.withRequestHeader` | Include request headers |
| `OTEL_WITH_RESPONSE_HEADER` | `instrumentation.openTelemetry.withResponseHeader` | Include response headers |
| `OTEL_FILTER_PATHS` | `instrumentation.openTelemetry.filterPaths` | Paths to filter |
| `OTEL_FILTER_METHODS` | `instrumentation.openTelemetry.filterMethods` | Methods to filter |

## Environment-Specific Configuration Files

### Development (`application-dev.yaml`)
- Debug mode enabled
- Relaxed security settings
- Detailed logging
- Local database and services

### Staging (`application-staging.yaml`)
- Production-like settings
- Moderate security
- Environment variable substitution for secrets
- AWS services

### Production (`application-prod.yaml`)
- Maximum security
- Minimal logging
- All secrets from environment variables
- Production-grade connection pools

## Usage

Set the `ENVIRONMENT` variable to load the appropriate configuration:

```bash
export ENVIRONMENT=dev     # Loads application-dev.yaml
export ENVIRONMENT=staging # Loads application-staging.yaml
export ENVIRONMENT=prod    # Loads application-prod.yaml
```

## Benefits

1. **Hierarchical Structure**: Better organization of configuration values
2. **Type Safety**: Proper types (int, bool, arrays) instead of string parsing
3. **Environment-Specific**: Easy management of different environments
4. **Validation**: Built-in YAML validation
5. **Documentation**: Self-documenting configuration structure
6. **Environment Variables**: Still supports ${VAR} substitution for secrets

## Code Changes

### Router (cmd/app/router/index.go)
- `viper.GetBool("DEBUG")` → `viper.GetBool("server.debug")`
- `viper.GetString("ALLOWED_HOSTS")` → `viper.GetStringSlice("security.trustedProxies")`

### Main Application (cmd/app/main.go)
- `viper.GetString("SERVER_TIMEZONE")` → `viper.GetString("server.timezone")`

### Instrumentation Config (modules/common/telemetry/instrumentation/config.go)
- All `LOG_*` variables moved to `instrumentation.logging.*`
- All `OTEL_*` variables moved to `instrumentation.openTelemetry.*`

### Database Config (modules/common/database/database.go)
- All `DB_*` variables moved to `database.*`
- Added connection pool configuration support
