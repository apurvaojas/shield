// Package instrumentation config provides Viper-based configuration for logging and OTEL middlewares.
// This replaces custom environment variable parsing with centralized Viper configuration.
package instrumentation

import (
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the instrumentation package
type Config struct {
	Logging ViperLoggingConfig `mapstructure:"logging"`
	OTEL    OTELConfig         `mapstructure:"otel"`
}

// ViperLoggingConfig holds configuration for the logging middleware from Viper
type ViperLoggingConfig struct {
	WithRequestBody    bool     `mapstructure:"with_request_body"`
	WithResponseBody   bool     `mapstructure:"with_response_body"`
	WithRequestHeader  bool     `mapstructure:"with_request_header"`
	WithResponseHeader bool     `mapstructure:"with_response_header"`
	WithUserAgent      bool     `mapstructure:"with_user_agent"`
	WithRequestID      bool     `mapstructure:"with_request_id"`
	WithSpanID         bool     `mapstructure:"with_span_id"`
	WithTraceID        bool     `mapstructure:"with_trace_id"`
	SkipPaths          []string `mapstructure:"skip_paths"`
	DefaultLevel       string   `mapstructure:"default_level"`
	ClientErrorLevel   string   `mapstructure:"client_error_level"`
	ServerErrorLevel   string   `mapstructure:"server_error_level"`
}

// OTELConfig holds configuration for OTEL middleware
type OTELConfig struct {
	ServiceName        string   `mapstructure:"service_name"`
	EnableTracing      bool     `mapstructure:"enable_tracing"`
	EnableMetrics      bool     `mapstructure:"enable_metrics"`
	WithSpanID         bool     `mapstructure:"with_span_id"`
	WithTraceID        bool     `mapstructure:"with_trace_id"`
	WithUserAgent      bool     `mapstructure:"with_user_agent"`
	WithRequestBody    bool     `mapstructure:"with_request_body"`
	WithResponseBody   bool     `mapstructure:"with_response_body"`
	WithRequestHeader  bool     `mapstructure:"with_request_header"`
	WithResponseHeader bool     `mapstructure:"with_response_header"`
	FilterPaths        []string `mapstructure:"filter_paths"`
	FilterMethods      []string `mapstructure:"filter_methods"`
}

// DefaultConfig returns default configuration values
func DefaultConfig() Config {
	return Config{
		Logging: ViperLoggingConfig{
			WithRequestBody:    false,
			WithResponseBody:   false,
			WithRequestHeader:  false,
			WithResponseHeader: false,
			WithUserAgent:      true,
			WithRequestID:      true,
			WithSpanID:         true,
			WithTraceID:        true,
			SkipPaths:          []string{"/health", "/metrics", "/ping"},
			DefaultLevel:       "info",
			ClientErrorLevel:   "warn",
			ServerErrorLevel:   "error",
		},
		OTEL: OTELConfig{
			ServiceName:        "gin-service",
			EnableTracing:      true,
			EnableMetrics:      true,
			WithSpanID:         true,
			WithTraceID:        true,
			WithUserAgent:      true,
			WithRequestBody:    false,
			WithResponseBody:   false,
			WithRequestHeader:  false,
			WithResponseHeader: false,
			FilterPaths:        []string{"/health", "/metrics", "/ping"},
			FilterMethods:      []string{},
		},
	}
}

// LoadConfig loads configuration from Viper with environment variable fallbacks
func LoadConfig() Config {
	config := DefaultConfig()

	// Set up Viper to read from environment variables with prefix
	viper.SetEnvPrefix("INSTRUMENTATION")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Logging configuration from environment variables
	if viper.IsSet("LOG_WITH_REQUEST_BODY") {
		config.Logging.WithRequestBody = viper.GetBool("LOG_WITH_REQUEST_BODY")
	}
	if viper.IsSet("LOG_WITH_RESPONSE_BODY") {
		config.Logging.WithResponseBody = viper.GetBool("LOG_WITH_RESPONSE_BODY")
	}
	if viper.IsSet("LOG_WITH_REQUEST_HEADER") {
		config.Logging.WithRequestHeader = viper.GetBool("LOG_WITH_REQUEST_HEADER")
	}
	if viper.IsSet("LOG_WITH_RESPONSE_HEADER") {
		config.Logging.WithResponseHeader = viper.GetBool("LOG_WITH_RESPONSE_HEADER")
	}
	if viper.IsSet("LOG_WITH_USER_AGENT") {
		config.Logging.WithUserAgent = viper.GetBool("LOG_WITH_USER_AGENT")
	}
	if viper.IsSet("LOG_WITH_REQUEST_ID") {
		config.Logging.WithRequestID = viper.GetBool("LOG_WITH_REQUEST_ID")
	}
	if viper.IsSet("LOG_WITH_SPAN_ID") {
		config.Logging.WithSpanID = viper.GetBool("LOG_WITH_SPAN_ID")
	}
	if viper.IsSet("LOG_WITH_TRACE_ID") {
		config.Logging.WithTraceID = viper.GetBool("LOG_WITH_TRACE_ID")
	}
	if viper.IsSet("LOG_SKIP_PATHS") {
		config.Logging.SkipPaths = viper.GetStringSlice("LOG_SKIP_PATHS")
	}
	if viper.IsSet("LOG_DEFAULT_LEVEL") {
		config.Logging.DefaultLevel = viper.GetString("LOG_DEFAULT_LEVEL")
	}
	if viper.IsSet("LOG_CLIENT_ERROR_LEVEL") {
		config.Logging.ClientErrorLevel = viper.GetString("LOG_CLIENT_ERROR_LEVEL")
	}
	if viper.IsSet("LOG_SERVER_ERROR_LEVEL") {
		config.Logging.ServerErrorLevel = viper.GetString("LOG_SERVER_ERROR_LEVEL")
	}

	// OTEL configuration from environment variables
	if viper.IsSet("OTEL_SERVICE_NAME") {
		config.OTEL.ServiceName = viper.GetString("OTEL_SERVICE_NAME")
	}
	if viper.IsSet("OTEL_ENABLE_TRACING") {
		config.OTEL.EnableTracing = viper.GetBool("OTEL_ENABLE_TRACING")
	}
	if viper.IsSet("OTEL_ENABLE_METRICS") {
		config.OTEL.EnableMetrics = viper.GetBool("OTEL_ENABLE_METRICS")
	}
	if viper.IsSet("OTEL_WITH_SPAN_ID") {
		config.OTEL.WithSpanID = viper.GetBool("OTEL_WITH_SPAN_ID")
	}
	if viper.IsSet("OTEL_WITH_TRACE_ID") {
		config.OTEL.WithTraceID = viper.GetBool("OTEL_WITH_TRACE_ID")
	}
	if viper.IsSet("OTEL_WITH_USER_AGENT") {
		config.OTEL.WithUserAgent = viper.GetBool("OTEL_WITH_USER_AGENT")
	}
	if viper.IsSet("OTEL_WITH_REQUEST_BODY") {
		config.OTEL.WithRequestBody = viper.GetBool("OTEL_WITH_REQUEST_BODY")
	}
	if viper.IsSet("OTEL_WITH_RESPONSE_BODY") {
		config.OTEL.WithResponseBody = viper.GetBool("OTEL_WITH_RESPONSE_BODY")
	}
	if viper.IsSet("OTEL_WITH_REQUEST_HEADER") {
		config.OTEL.WithRequestHeader = viper.GetBool("OTEL_WITH_REQUEST_HEADER")
	}
	if viper.IsSet("OTEL_WITH_RESPONSE_HEADER") {
		config.OTEL.WithResponseHeader = viper.GetBool("OTEL_WITH_RESPONSE_HEADER")
	}
	if viper.IsSet("OTEL_FILTER_PATHS") {
		config.OTEL.FilterPaths = viper.GetStringSlice("OTEL_FILTER_PATHS")
	}
	if viper.IsSet("OTEL_FILTER_METHODS") {
		config.OTEL.FilterMethods = viper.GetStringSlice("OTEL_FILTER_METHODS")
	}

	return config
}

// GetLoggingConfig returns the logging configuration
func GetLoggingConfig() ViperLoggingConfig {
	return LoadConfig().Logging
}

// GetOTELConfig returns the OTEL configuration
func GetOTELConfig() OTELConfig {
	return LoadConfig().OTEL
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

// ToSlogLevel converts config level string to slog.Level
func (c ViperLoggingConfig) DefaultSlogLevel() slog.Level {
	return parseLogLevel(c.DefaultLevel)
}

// ClientErrorSlogLevel converts config level string to slog.Level
func (c ViperLoggingConfig) ClientErrorSlogLevel() slog.Level {
	return parseLogLevel(c.ClientErrorLevel)
}

// ServerErrorSlogLevel converts config level string to slog.Level
func (c ViperLoggingConfig) ServerErrorSlogLevel() slog.Level {
	return parseLogLevel(c.ServerErrorLevel)
}
