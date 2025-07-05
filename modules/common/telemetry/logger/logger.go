/*
* Write a logger that can be used in the whole project.
* It should be able to log messages with different levels (info, warning, error).
* Use the standard log package for simplicity. https://github.com/samber/slog-multi
* It should be able to log to different outputs (console, file, OTEL).
* use Environment variable to set the log level.
* use environment variable for OTEL configuration.
* use environment variable for file logging configuration.
* foe local development, use console logging.
* for production and staging, use file logging and OTEL both.
* expose only InitLogger function to initialize the logger.
* This will initialize the logger based on the environment. and set the log/slog setDefault to the logger.
 */
package common

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"shield/modules/common/telemetry/instrumentation"

	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogConfig holds configuration for the logger
type LogConfig struct {
	Level           string
	Environment     string
	EnableFile      bool
	FileDir         string
	EnableOTEL      bool
	OTELEndpoint    string
	OTELServiceName string
	// File rotation settings
	MaxFileSize int // in MB
	MaxFiles    int // max number of log files to keep
	MaxAge      int // max days to keep log files
	Compress    bool
}

// getLogConfig reads configuration from environment variables
func getLogConfig() LogConfig {
	config := LogConfig{
		Level:           getEnvWithDefault("LOG_LEVEL", "info"),
		Environment:     getEnvWithDefault("ENVIRONMENT", "development"),
		EnableFile:      getEnvWithDefault("LOG_FILE_ENABLED", "false") == "true",
		FileDir:         getEnvWithDefault("LOG_FILE_DIR", "./logs"),
		EnableOTEL:      getEnvWithDefault("OTEL_ENABLED", "false") == "true",
		OTELEndpoint:    getEnvWithDefault("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		OTELServiceName: getEnvWithDefault("OTEL_SERVICE_NAME", "shield-api"),
		MaxFileSize:     parseIntWithDefault("LOG_MAX_FILE_SIZE_MB", 100),
		MaxFiles:        parseIntWithDefault("LOG_MAX_FILES", 5),
		MaxAge:          parseIntWithDefault("LOG_MAX_AGE_DAYS", 30),
		Compress:        getEnvWithDefault("LOG_COMPRESS", "true") == "true",
	}
	return config
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseIntWithDefault parses environment variable as int with default value
func parseIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// createConsoleHandler creates a console handler with formatting
func createConsoleHandler(level slog.Level, environment string) slog.Handler {
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: environment == "development",
	}

	var baseHandler slog.Handler
	if environment == "development" {
		// Use text handler for better readability in development
		baseHandler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// Use JSON handler for production environments
		baseHandler = slog.NewJSONHandler(os.Stdout, opts)
	}

	// Apply formatters using slog-formatter
	formatters := []slogformatter.Formatter{
		// Add timestamp formatting
		slogformatter.TimeFormatter(time.RFC3339, time.UTC),
		// Error formatting with stack traces
		slogformatter.ErrorFormatter("error"),
		// Format trace information
		slogformatter.FormatByKey("trace_id", func(v slog.Value) slog.Value {
			return v // Keep trace_id as is
		}),
		slogformatter.FormatByKey("span_id", func(v slog.Value) slog.Value {
			return v // Keep span_id as is
		}),
	}

	return slogformatter.NewFormatterHandler(formatters...)(baseHandler)
}

// createFileHandler creates a file handler with rotation using lumberjack
func createFileHandler(level slog.Level, config LogConfig) (slog.Handler, io.Closer, error) {
	// Create the log directory
	if err := os.MkdirAll(config.FileDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Use lumberjack for file rotation
	rotatingWriter := &lumberjack.Logger{
		Filename:   filepath.Join(config.FileDir, "app.log"),
		MaxSize:    config.MaxFileSize, // megabytes
		MaxBackups: config.MaxFiles,
		MaxAge:     config.MaxAge, // days
		Compress:   config.Compress,
		LocalTime:  true, // Use local time for log file names
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	baseHandler := slog.NewJSONHandler(rotatingWriter, opts)

	// Apply formatters using slog-formatter
	formatters := []slogformatter.Formatter{
		// Add timestamp formatting
		slogformatter.TimeFormatter(time.RFC3339, time.UTC),
		// Error formatting with stack traces
		slogformatter.ErrorFormatter("error"),
		// Format trace information
		slogformatter.FormatByKey("trace_id", func(v slog.Value) slog.Value {
			return v
		}),
		slogformatter.FormatByKey("span_id", func(v slog.Value) slog.Value {
			return v
		}),
	}

	formattedHandler := slogformatter.NewFormatterHandler(formatters...)(baseHandler)
	return formattedHandler, rotatingWriter, nil
}

// createOTELHandler creates an OpenTelemetry log handler
func createOTELHandler(endpoint, serviceName string) (slog.Handler, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("OTEL endpoint is required")
	}

	// Create OTLP log exporter
	ctx := context.Background()
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithHeaders(map[string]string{
			"service.name": serviceName,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create log processor
	processor := sdklog.NewBatchProcessor(exporter)

	// Create logger provider
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(processor),
		sdklog.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	// Set global logger provider
	global.SetLoggerProvider(provider)

	// Create a bridge handler that converts slog records to OTEL log records
	return &otelHandler{
		logger: provider.Logger("slog-bridge"),
	}, nil
}

// otelHandler bridges slog to OpenTelemetry logs
type otelHandler struct {
	logger otellog.Logger
	attrs  []slog.Attr
	groups []string
}

func (h *otelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Convert slog level to OTEL severity
	severity := slogLevelToOTELSeverity(level)
	return h.logger.Enabled(ctx, otellog.EnabledParameters{
		Severity: severity,
	})
}

func (h *otelHandler) Handle(ctx context.Context, record slog.Record) error {
	// Create OTEL log record
	var logRecord otellog.Record
	logRecord.SetTimestamp(record.Time)
	logRecord.SetSeverity(slogLevelToOTELSeverity(record.Level))
	logRecord.SetSeverityText(record.Level.String())
	logRecord.SetBody(otellog.StringValue(record.Message))

	// Add trace information if available
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		spanCtx := span.SpanContext()
		logRecord.AddAttributes(
			otellog.String("trace_id", spanCtx.TraceID().String()),
			otellog.String("span_id", spanCtx.SpanID().String()),
		)
	}

	// Add attributes
	record.Attrs(func(attr slog.Attr) bool {
		logRecord.AddAttributes(slogAttrToOTELKeyValue(attr))
		return true
	})

	// Add handler attributes
	for _, attr := range h.attrs {
		logRecord.AddAttributes(slogAttrToOTELKeyValue(attr))
	}

	h.logger.Emit(ctx, logRecord)
	return nil
}

func (h *otelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &otelHandler{
		logger: h.logger,
		attrs:  append(h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *otelHandler) WithGroup(name string) slog.Handler {
	return &otelHandler{
		logger: h.logger,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

// Helper functions for OTEL conversion
func slogLevelToOTELSeverity(level slog.Level) otellog.Severity {
	switch level {
	case slog.LevelDebug:
		return otellog.SeverityDebug
	case slog.LevelInfo:
		return otellog.SeverityInfo
	case slog.LevelWarn:
		return otellog.SeverityWarn
	case slog.LevelError:
		return otellog.SeverityError
	default:
		return otellog.SeverityInfo
	}
}

func slogAttrToOTELKeyValue(attr slog.Attr) otellog.KeyValue {
	switch attr.Value.Kind() {
	case slog.KindString:
		return otellog.String(attr.Key, attr.Value.String())
	case slog.KindInt64:
		return otellog.Int64(attr.Key, attr.Value.Int64())
	case slog.KindFloat64:
		return otellog.Float64(attr.Key, attr.Value.Float64())
	case slog.KindBool:
		return otellog.Bool(attr.Key, attr.Value.Bool())
	default:
		return otellog.String(attr.Key, attr.Value.String())
	}
}

// maskingHandler implements slog.Handler to mask sensitive data in log messages
type maskingHandler struct {
	next   slog.Handler
	masker *instrumentation.SensitiveDataMasker
}

// newMaskingHandler creates a new masking handler
func newMaskingHandler(next slog.Handler) *maskingHandler {
	return &maskingHandler{
		next:   next,
		masker: instrumentation.GetDefaultMasker(),
	}
}

// Enabled returns whether the handler is enabled for the given level
func (h *maskingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle processes the log record, masks sensitive data, and passes it to the next handler
func (h *maskingHandler) Handle(ctx context.Context, record slog.Record) error {
	// Create a new record with masked message
	maskedMessage := h.masker.MaskPII(record.Message)

	// Create a new record with the masked message
	newRecord := slog.NewRecord(record.Time, record.Level, maskedMessage, record.PC)

	// Process attributes and mask sensitive data
	record.Attrs(func(attr slog.Attr) bool {
		maskedAttr := h.maskAttribute(attr)
		newRecord.AddAttrs(maskedAttr)
		return true
	})

	return h.next.Handle(ctx, newRecord)
}

// maskAttribute recursively masks sensitive data in slog attributes
func (h *maskingHandler) maskAttribute(attr slog.Attr) slog.Attr {
	key := strings.ToLower(attr.Key)

	// Check if this is a sensitive field that should be fully masked
	sensitiveFields := map[string]bool{
		"password":           true,
		"passwd":             true,
		"pwd":                true,
		"secret":             true,
		"token":              true,
		"api_key":            true,
		"apikey":             true,
		"private_key":        true,
		"privatekey":         true,
		"access_token":       true,
		"refresh_token":      true,
		"client_secret":      true,
		"authorization_code": true,
		"pin":                true,
		"otp":                true,
		"cvv":                true,
		"cvc":                true,
		"security_code":      true,
		"authorization":      true,
		"cookie":             true,
		"set-cookie":         true,
		"x-auth-token":       true,
		"x-api-key":          true,
		"bearer":             true,
	}

	if sensitiveFields[key] {
		return slog.String(attr.Key, "[MASKED]")
	}

	// For other attributes, mask PII in the value
	switch attr.Value.Kind() {
	case slog.KindString:
		maskedValue := h.masker.MaskPII(attr.Value.String())
		return slog.String(attr.Key, maskedValue)
	case slog.KindGroup:
		// Handle grouped attributes recursively
		var maskedAttrs []any
		for _, groupAttr := range attr.Value.Group() {
			maskedAttr := h.maskAttribute(groupAttr)
			maskedAttrs = append(maskedAttrs, maskedAttr)
		}
		return slog.Group(attr.Key, maskedAttrs...)
	default:
		// For non-string values, return as-is (numbers, booleans, etc.)
		return attr
	}
}

// WithAttrs returns a new handler with additional attributes (also masked)
func (h *maskingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var maskedAttrs []slog.Attr
	for _, attr := range attrs {
		maskedAttrs = append(maskedAttrs, h.maskAttribute(attr))
	}

	return &maskingHandler{
		next:   h.next.WithAttrs(maskedAttrs),
		masker: h.masker,
	}
}

// WithGroup returns a new handler with a group
func (h *maskingHandler) WithGroup(name string) slog.Handler {
	return &maskingHandler{
		next:   h.next.WithGroup(name),
		masker: h.masker,
	}
}

// InitLogger initializes the logger based on environment configuration
func InitLogger() error {
	config := getLogConfig()
	level := parseLogLevel(config.Level)

	var handlers []slog.Handler

	// Always add console handler with masking (masking is mandatory)
	consoleHandler := createConsoleHandler(level, config.Environment)
	maskedConsoleHandler := newMaskingHandler(consoleHandler)
	handlers = append(handlers, maskedConsoleHandler)

	// Add file handler for production and staging with masking
	if config.Environment != "development" || config.EnableFile {
		fileHandler, closer, err := createFileHandler(level, config)
		if err != nil {
			// Log error but don't fail initialization
			fmt.Fprintf(os.Stderr, "Failed to create file handler: %v\n", err)
		} else {
			maskedFileHandler := newMaskingHandler(fileHandler)
			handlers = append(handlers, maskedFileHandler)
			// Note: In a real implementation, you should store the closer to clean up on shutdown
			_ = closer
		}
	}

	// Add OTEL handler for production and staging with masking
	if (config.Environment == "production" || config.Environment == "staging") && config.EnableOTEL {
		otelHandler, err := createOTELHandler(config.OTELEndpoint, config.OTELServiceName)
		if err != nil {
			// Log error but don't fail initialization
			fmt.Fprintf(os.Stderr, "Failed to create OTEL handler: %v\n", err)
		} else {
			maskedOTELHandler := newMaskingHandler(otelHandler)
			handlers = append(handlers, maskedOTELHandler)
		}
	}

	// Create multi-handler based on number of handlers
	var multiHandler slog.Handler
	if len(handlers) == 1 {
		multiHandler = handlers[0]
	} else {
		// Use fanout to distribute logs to all handlers
		multiHandler = slogmulti.Fanout(handlers...)
	}

	// Add middleware to enrich logs with environment information and trace context
	enrichmentHandler := slogmulti.Pipe(
		slogmulti.NewHandleInlineMiddleware(func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
			// Add environment and service information for non-development environments
			if config.Environment != "development" {
				record.AddAttrs(
					slog.String("environment", config.Environment),
					slog.String("service", config.OTELServiceName),
				)
			}

			// Add trace information if available
			if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
				spanCtx := span.SpanContext()
				record.AddAttrs(
					slog.String("trace_id", spanCtx.TraceID().String()),
					slog.String("span_id", spanCtx.SpanID().String()),
				)
			}

			return next(ctx, record)
		}),
	).Handler(multiHandler)

	// Wrap with error recovery
	finalHandler := slogmulti.Pipe(
		slogmulti.RecoverHandlerError(func(ctx context.Context, record slog.Record, err error) {
			fmt.Fprintf(os.Stderr, "Logger error: %v\n", err)
		}),
	).Handler(enrichmentHandler)

	// Create logger (masking is now mandatory and applied to each handler)
	logger := slog.New(finalHandler)

	// Set as default logger
	slog.SetDefault(logger)

	// Log initialization success
	slog.Info("Logger initialized successfully",
		slog.String("level", config.Level),
		slog.String("environment", config.Environment),
		slog.Bool("file_enabled", config.EnableFile || config.Environment != "development"),
		slog.Bool("otel_enabled", config.EnableOTEL && (config.Environment == "production" || config.Environment == "staging")),
		slog.Bool("masking_enabled", true), // Always true since masking is now mandatory
	)

	return nil
}
