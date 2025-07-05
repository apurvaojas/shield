/*
logging middleware for gin
This middleware captures HTTP request logs using slog.
It uses the slog package for structured logging and supports various log levels.
* https://github.com/samber/slog-gin
* exposes InitLoggingMiddleware function that initializes the slog middleware for Gin.
*/
package instrumentation

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
)

// InitLoggingMiddleware initializes the slog middleware for Gin using Viper configuration
func InitLoggingMiddleware() gin.HandlerFunc {
	config := GetLoggingConfig()

	// Configure slog-gin middleware with Viper-loaded config
	// Note: Sensitive data masking is now handled directly in the logger
	slogginConfig := sloggin.Config{
		WithRequestBody:    config.WithRequestBody,
		WithResponseBody:   config.WithResponseBody,
		WithRequestHeader:  config.WithRequestHeader,
		WithResponseHeader: config.WithResponseHeader,
		WithUserAgent:      config.WithUserAgent,
		WithRequestID:      config.WithRequestID,
		WithSpanID:         config.WithSpanID,
		WithTraceID:        config.WithTraceID,
		DefaultLevel:       config.DefaultSlogLevel(),
		ClientErrorLevel:   config.ClientErrorSlogLevel(),
		ServerErrorLevel:   config.ServerErrorSlogLevel(),
		Filters: []sloggin.Filter{
			// Skip specified paths
			func(c *gin.Context) bool {
				path := c.Request.URL.Path
				for _, skipPath := range config.SkipPaths {
					if path == skipPath {
						return false // Skip this path
					}
				}
				return true // Allow this path
			},
		},
	}

	// Return the middleware directly - masking is handled by the logger itself
	return sloggin.NewWithConfig(slog.Default(), slogginConfig)
}

// AddCustomAttributes adds custom attributes to a single log entry
// This function can be used within request handlers to add context-specific logging
// Note: Sensitive data masking is handled automatically by the logger
func AddCustomAttributes(c *gin.Context, attrs ...slog.Attr) {
	for _, attr := range attrs {
		sloggin.AddCustomAttributes(c, attr)
	}
}

// AddCustomAttributesUnsafe is now the same as AddCustomAttributes since masking is handled in the logger
// This function is kept for backward compatibility
func AddCustomAttributesUnsafe(c *gin.Context, attrs ...slog.Attr) {
	AddCustomAttributes(c, attrs...)
}

// InitRequestIDMiddleware adds a request ID to each request if not present
// This should be used before the logging middleware to ensure request IDs are captured
func InitRequestIDMiddleware() gin.HandlerFunc {
	return gin.Recovery() // Using gin.Recovery() which already handles request ID generation
}
