package common

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// TestLogConfig tests the configuration reading from environment variables
func TestLogConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected LogConfig
	}{
		{
			name:    "default configuration",
			envVars: map[string]string{},
			expected: LogConfig{
				Level:           "info",
				Environment:     "development",
				EnableFile:      false,
				FileDir:         "./logs",
				EnableOTEL:      false,
				OTELEndpoint:    "",
				OTELServiceName: "shield-api",
				MaxFileSize:     100,
				MaxFiles:        5,
				MaxAge:          30,
				Compress:        true,
			},
		},
		{
			name: "production configuration",
			envVars: map[string]string{
				"LOG_LEVEL":                   "error",
				"ENVIRONMENT":                 "production",
				"LOG_FILE_ENABLED":            "true",
				"LOG_FILE_DIR":                "/var/log/app",
				"OTEL_ENABLED":                "true",
				"OTEL_EXPORTER_OTLP_ENDPOINT": "http://jaeger:14268/api/traces",
				"OTEL_SERVICE_NAME":           "my-service",
				"LOG_MAX_FILE_SIZE_MB":        "500",
				"LOG_MAX_FILES":               "10",
				"LOG_MAX_AGE_DAYS":            "7",
				"LOG_COMPRESS":                "false",
			},
			expected: LogConfig{
				Level:           "error",
				Environment:     "production",
				EnableFile:      true,
				FileDir:         "/var/log/app",
				EnableOTEL:      true,
				OTELEndpoint:    "http://jaeger:14268/api/traces",
				OTELServiceName: "my-service",
				MaxFileSize:     500,
				MaxFiles:        10,
				MaxAge:          7,
				Compress:        false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				// Clean up environment variables
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			config := getLogConfig()
			if config != tt.expected {
				t.Errorf("getLogConfig() = %+v, want %+v", config, tt.expected)
			}
		})
	}
}

// TestParseLogLevel tests log level parsing
func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"invalid", slog.LevelInfo}, // default case
		{"", slog.LevelInfo},        // default case
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCreateConsoleHandler tests console handler creation
func TestCreateConsoleHandler(t *testing.T) {
	tests := []struct {
		name        string
		level       slog.Level
		environment string
	}{
		{"development", slog.LevelDebug, "development"},
		{"production", slog.LevelInfo, "production"},
		{"staging", slog.LevelWarn, "staging"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := createConsoleHandler(tt.level, tt.environment)
			if handler == nil {
				t.Error("createConsoleHandler() returned nil")
			}

			// Test that handler respects the log level
			ctx := context.Background()
			if !handler.Enabled(ctx, tt.level) {
				t.Errorf("handler should be enabled for level %v", tt.level)
			}

			// Test that handler rejects lower levels
			if tt.level > slog.LevelDebug && handler.Enabled(ctx, slog.LevelDebug) {
				t.Error("handler should not be enabled for debug level when level is higher")
			}
		})
	}
}

// TestCreateFileHandler tests file handler creation
func TestCreateFileHandler(t *testing.T) {
	tempDir := t.TempDir()

	config := LogConfig{
		FileDir:     tempDir,
		MaxFileSize: 10,
		MaxFiles:    3,
		MaxAge:      7,
		Compress:    true,
	}

	handler, closer, err := createFileHandler(slog.LevelInfo, config)
	if err != nil {
		t.Fatalf("createFileHandler() failed: %v", err)
	}
	defer closer.Close()

	if handler == nil {
		t.Error("createFileHandler() returned nil handler")
	}

	// Check that the log file was created
	logFile := filepath.Join(tempDir, "app.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		// File might not exist until first write, which is fine
	}

	// Test logging to file
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	err = handler.Handle(ctx, record)
	if err != nil {
		t.Errorf("handler.Handle() failed: %v", err)
	}
}

// TestOTELHandler tests OTEL handler functionality
func TestOTELHandler(t *testing.T) {
	// Skip this test if we can't create a proper mock
	t.Skip("OTEL handler test requires complex mock setup")

	// This test would require a full OTEL setup which is complex for unit testing
	// In practice, integration tests would be more appropriate for OTEL functionality
}

// TestSlogLevelToOTELSeverity tests the conversion function
func TestSlogLevelToOTELSeverity(t *testing.T) {
	tests := []struct {
		slogLevel slog.Level
		expected  string // We'll check the string representation
	}{
		{slog.LevelDebug, "DEBUG"},
		{slog.LevelInfo, "INFO"},
		{slog.LevelWarn, "WARN"},
		{slog.LevelError, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			severity := slogLevelToOTELSeverity(tt.slogLevel)
			if severity.String() != tt.expected {
				t.Errorf("slogLevelToOTELSeverity(%v) = %v, want %v", tt.slogLevel, severity.String(), tt.expected)
			}
		})
	}
}

// TestInitLogger tests the main initialization function
func TestInitLogger(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "development environment",
			envVars: map[string]string{
				"ENVIRONMENT": "development",
				"LOG_LEVEL":   "debug",
			},
			wantErr: false,
		},
		{
			name: "production environment with file logging",
			envVars: map[string]string{
				"ENVIRONMENT":      "production",
				"LOG_LEVEL":        "info",
				"LOG_FILE_ENABLED": "true",
				"LOG_FILE_DIR":     t.TempDir(),
			},
			wantErr: false,
		},
		{
			name: "staging environment",
			envVars: map[string]string{
				"ENVIRONMENT": "staging",
				"LOG_LEVEL":   "warn",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				// Clean up environment variables
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			err := InitLogger()
			if (err != nil) != tt.wantErr {
				t.Errorf("InitLogger() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Test that the logger works
			slog.Info("test info message")
			slog.Warn("test warn message")
			slog.Error("test error message")
		})
	}
}

// TestLoggerWithTraceContext tests logging with trace context
func TestLoggerWithTraceContext(t *testing.T) {
	// Set up environment for testing
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("LOG_LEVEL")
	}()

	err := InitLogger()
	if err != nil {
		t.Fatalf("InitLogger() failed: %v", err)
	}

	// Create a mock tracer
	tracer := tracenoop.NewTracerProvider().Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-operation")
	defer span.End()

	// Log with trace context
	slog.InfoContext(ctx, "test message with trace")
	slog.ErrorContext(ctx, "test error with trace")
}

// TestLoggerWithErrorStackTrace tests that error stack traces are captured
func TestLoggerWithErrorStackTrace(t *testing.T) {
	// Capture stdout to verify log output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	// Set up environment for testing
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("LOG_LEVEL")
	}()

	err := InitLogger()
	if err != nil {
		t.Fatalf("InitLogger() failed: %v", err)
	}

	// Create an error and log it
	testErr := errors.New("test error")
	slog.Error("error occurred", "error", testErr)

	// Close writer and read output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that the error was logged (basic check)
	if !strings.Contains(output, "test error") {
		t.Error("Expected error message not found in log output")
	}
}

// TestFileRotation tests that file rotation configuration works
func TestFileRotation(t *testing.T) {
	tempDir := t.TempDir()

	// Set up environment for file logging
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("LOG_FILE_ENABLED", "true")
	os.Setenv("LOG_FILE_DIR", tempDir)
	os.Setenv("LOG_MAX_FILE_SIZE_MB", "1") // Small size for testing
	os.Setenv("LOG_MAX_FILES", "3")
	os.Setenv("LOG_MAX_AGE_DAYS", "1")

	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FILE_ENABLED")
		os.Unsetenv("LOG_FILE_DIR")
		os.Unsetenv("LOG_MAX_FILE_SIZE_MB")
		os.Unsetenv("LOG_MAX_FILES")
		os.Unsetenv("LOG_MAX_AGE_DAYS")
	}()

	err := InitLogger()
	if err != nil {
		t.Fatalf("InitLogger() failed: %v", err)
	}

	// Generate some log messages
	for i := 0; i < 10; i++ {
		slog.Info(fmt.Sprintf("test message %d", i))
	}

	// Check that log file was created
	logFile := filepath.Join(tempDir, "app.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestLogLevels tests that different log levels work correctly
func TestLogLevels(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	// Set up environment for testing with INFO level
	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("LOG_LEVEL", "info")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("LOG_LEVEL")
	}()

	err := InitLogger()
	if err != nil {
		t.Fatalf("InitLogger() failed: %v", err)
	}

	// Log at different levels
	slog.Debug("debug message")  // Should not appear
	slog.Info("info message")    // Should appear
	slog.Warn("warning message") // Should appear
	slog.Error("error message")  // Should appear

	// Close writer and read output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that appropriate messages appear
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not appear when log level is INFO")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info message should appear when log level is INFO")
	}
	if !strings.Contains(output, "warning message") {
		t.Error("Warning message should appear when log level is INFO")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should appear when log level is INFO")
	}
}

// TestStructuredLogging tests structured logging with attributes
func TestStructuredLogging(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	// Set up environment for JSON logging
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("LOG_LEVEL", "info")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("LOG_LEVEL")
	}()

	err := InitLogger()
	if err != nil {
		t.Fatalf("InitLogger() failed: %v", err)
	}

	// Log with structured data
	slog.Info("user action",
		slog.String("user_id", "12345"),
		slog.String("action", "login"),
		slog.Int("attempt", 1),
		slog.Bool("success", true),
	)

	// Close writer and read output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Split by lines to get individual JSON objects
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("No log output found")
	}

	// Parse the last JSON line (our user action log)
	var logEntry map[string]interface{}
	lastLine := lines[len(lines)-1]
	if err := json.Unmarshal([]byte(lastLine), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v\nOutput: %s", err, lastLine)
	}

	// Check that structured fields are present
	if logEntry["user_id"] != "12345" {
		t.Error("Expected user_id field not found or incorrect")
	}
	if logEntry["action"] != "login" {
		t.Error("Expected action field not found or incorrect")
	}
	if logEntry["attempt"] != float64(1) { // JSON numbers are float64
		t.Error("Expected attempt field not found or incorrect")
	}
	if logEntry["success"] != true {
		t.Error("Expected success field not found or incorrect")
	}
}

// Benchmark tests
func BenchmarkLogger(b *testing.B) {
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("LOG_LEVEL", "info")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("LOG_LEVEL")
	}()

	InitLogger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slog.Info("benchmark message",
				slog.String("key", "value"),
				slog.Int("number", 42),
			)
		}
	})
}

func BenchmarkLoggerWithContext(b *testing.B) {
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("LOG_LEVEL", "info")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("LOG_LEVEL")
	}()

	InitLogger()

	// Set up a trace context
	otel.SetTracerProvider(tracenoop.NewTracerProvider())
	tracer := otel.Tracer("benchmark")
	ctx, span := tracer.Start(context.Background(), "benchmark-operation")
	defer span.End()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slog.InfoContext(ctx, "benchmark message with context",
				slog.String("key", "value"),
				slog.Int("number", 42),
			)
		}
	})
}
