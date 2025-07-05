// Example usage of the logger
// This file demonstrates how to use the logger package
package common

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// ExampleUsage demonstrates how to use the logger with enhanced features
func ExampleUsage() {
	// Set environment variables for configuration
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("LOG_FILE_ENABLED", "true")
	os.Setenv("LOG_FILE_DIR", "./logs")
	os.Setenv("LOG_ROTATION_ENABLED", "true")
	os.Setenv("LOG_MAX_FILE_SIZE_MB", "10")
	os.Setenv("LOG_MAX_FILES", "3")
	os.Setenv("LOG_MAX_AGE_DAYS", "7")
	os.Setenv("LOG_ROTATION_INTERVAL", "daily")

	// Initialize the logger
	if err := InitLogger(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// Create a context with trace information (simulated)
	ctx := context.Background()

	// Basic logging - timestamps are automatically added
	slog.Debug("This is a debug message")
	slog.Info("Application started successfully")
	slog.Warn("This is a warning message")

	// Structured logging with attributes
	slog.InfoContext(ctx, "User logged in",
		slog.String("user_id", "12345"),
		slog.String("username", "john_doe"),
		slog.Time("login_time", time.Now()),
	)

	// Error logging - stack trace is automatically added for error level
	slog.ErrorContext(ctx, "Database connection failed",
		slog.String("error", "connection timeout"),
		slog.String("database", "postgresql"),
		slog.Int("retry_count", 3),
	)

	// With attributes for contextual logging
	logger := slog.With(
		slog.String("module", "auth"),
		slog.String("version", "1.0.0"),
	)

	logger.Info("Authentication module initialized")
	logger.Error("Failed to validate token",
		slog.String("token_id", "abc123"),
		slog.String("reason", "expired"),
	)

	// Group related attributes
	groupedLogger := logger.WithGroup("database")
	groupedLogger.Info("Database query executed",
		slog.Duration("execution_time", 150*time.Millisecond),
		slog.Int("rows_affected", 5),
	)
}

// ProductionExample shows configuration for production environment
func ProductionExample() {
	// Production environment variables
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("LOG_FILE_ENABLED", "true")
	os.Setenv("LOG_FILE_DIR", "/var/log/shield")
	os.Setenv("LOG_ROTATION_ENABLED", "true")
	os.Setenv("LOG_MAX_FILE_SIZE_MB", "100")
	os.Setenv("LOG_MAX_FILES", "10")
	os.Setenv("LOG_MAX_AGE_DAYS", "30")
	os.Setenv("LOG_ROTATION_INTERVAL", "daily")
	os.Setenv("OTEL_ENABLED", "true")
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4317")
	os.Setenv("OTEL_SERVICE_NAME", "shield-api")

	// Initialize logger
	if err := InitLogger(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// Log some events - will include environment and service info automatically
	slog.Info("Service started in production mode",
		slog.String("version", "1.2.3"),
		slog.String("commit", "abc123def"),
	)
}
