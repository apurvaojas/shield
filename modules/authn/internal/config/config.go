`package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server      ServerConfig
	Cognito     CognitoConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	OPA         OPAConfig
	Monitoring  MonitoringConfig
	RateLimit   RateLimitConfig
	Security    SecurityConfig
	Features    FeatureConfig
}

// ServerConfig holds server-specific configuration.
type ServerConfig struct {
	Port        string `mapstructure:"PORT"`
	Environment string `mapstructure:"ENVIRONMENT"` // e.g., development, staging, production
	LogLevel    string `mapstructure:"LOG_LEVEL"`
}

// CognitoConfig holds AWS Cognito specific configuration.
type CognitoConfig struct {
	UserPoolID     string `mapstructure:"COGNITO_USER_POOL_ID"`
	AppClientID    string `mapstructure:"COGNITO_APP_CLIENT_ID"`
	AppClientSecret string `mapstructure:"COGNITO_APP_CLIENT_SECRET"` // Optional, if client secret is enabled
	Region         string `mapstructure:"COGNITO_REGION"`
	Domain         string `mapstructure:"COGNITO_DOMAIN"` // For federated sign-in if using Cognito Hosted UI
}

// DatabaseConfig holds database connection details.
type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSLMODE"` // e.g., disable, require, verify-full
}

// RedisConfig holds Redis connection details.
type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

// JWTConfig holds JWT token configuration.
type JWTConfig struct {
	Secret        string `mapstructure:"JWT_SECRET"`
	Expiry        string `mapstructure:"JWT_EXPIRY"`
	RefreshExpiry string `mapstructure:"JWT_REFRESH_EXPIRY"`
}

// OPAConfig holds Open Policy Agent configuration.
type OPAConfig struct {
	ServerURL  string `mapstructure:"OPA_SERVER_URL"`
	PolicyPath string `mapstructure:"OPA_POLICY_PATH"`
}

// MonitoringConfig holds observability configuration.
type MonitoringConfig struct {
	JaegerEndpoint     string `mapstructure:"JAEGER_ENDPOINT"`
	PrometheusEndpoint string `mapstructure:"PROMETHEUS_ENDPOINT"`
	EnableMetrics      bool   `mapstructure:"ENABLE_METRICS"`
	EnableTracing      bool   `mapstructure:"ENABLE_TRACING"`
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Enabled            bool `mapstructure:"RATE_LIMIT_ENABLED"`
	RequestsPerMinute  int  `mapstructure:"RATE_LIMIT_REQUESTS_PER_MINUTE"`
	Burst              int  `mapstructure:"RATE_LIMIT_BURST"`
}

// SecurityConfig holds security-related configuration.
type SecurityConfig struct {
	CORSAllowedOrigins []string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	CORSAllowedMethods []string `mapstructure:"CORS_ALLOWED_METHODS"`
	CORSAllowedHeaders []string `mapstructure:"CORS_ALLOWED_HEADERS"`
	TrustedProxies     []string `mapstructure:"TRUSTED_PROXIES"`
}

// FeatureConfig holds feature flag configuration.
type FeatureConfig struct {
	MultiFactorAuth  bool `mapstructure:"FEATURE_MULTI_FACTOR_AUTH"`
	DeviceTracking   bool `mapstructure:"FEATURE_DEVICE_TRACKING"`
	SessionRotation  bool `mapstructure:"FEATURE_SESSION_ROTATION"`
}

// LoadConfig loads configuration from environment variables and .env file.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path) // Path to look for the config file in
	
	// Determine which environment file to load
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}
	
	// Try to load environment-specific file first
	envFile := fmt.Sprintf(".env.%s", env)
	viper.SetConfigName(envFile)
	viper.SetConfigType("env")
	
	viper.AutomaticEnv() // Read in environment variables that match

	// Set default values
	setDefaults()

	// Try to read environment-specific file
	if err = viper.ReadInConfig(); err != nil {
		// If environment-specific file not found, try generic .env
		viper.SetConfigName(".env")
		if err = viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Printf("Configuration file not found for environment '%s', relying on environment variables and defaults.", env)
			} else {
				log.Printf("Error reading config file: %s\n", err)
				return
			}
		}
	}

	// Replace dots with underscores for environment variables for nested structs
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Handle comma-separated values for slices
	handleSliceValues()

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Printf("Unable to decode into struct, %v", err)
		return
	}

	log.Printf("Configuration loaded successfully for environment: %s", config.Server.Environment)
	return
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("PORT", "8081")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("LOG_LEVEL", "debug")
	
	// Database defaults
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSLMODE", "disable")
	
	// Redis defaults
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", 0)
	
	// JWT defaults
	viper.SetDefault("JWT_EXPIRY", "24h")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "168h")
	
	// Monitoring defaults
	viper.SetDefault("ENABLE_METRICS", true)
	viper.SetDefault("ENABLE_TRACING", true)
	
	// Rate limiting defaults
	viper.SetDefault("RATE_LIMIT_ENABLED", true)
	viper.SetDefault("RATE_LIMIT_REQUESTS_PER_MINUTE", 100)
	viper.SetDefault("RATE_LIMIT_BURST", 20)
	
	// Feature flags defaults
	viper.SetDefault("FEATURE_MULTI_FACTOR_AUTH", false)
	viper.SetDefault("FEATURE_DEVICE_TRACKING", true)
	viper.SetDefault("FEATURE_SESSION_ROTATION", true)
}

// handleSliceValues processes comma-separated values for slice fields
func handleSliceValues() {
	// Handle CORS origins
	if origins := viper.GetString("CORS_ALLOWED_ORIGINS"); origins != "" {
		viper.Set("CORS_ALLOWED_ORIGINS", strings.Split(origins, ","))
	}
	
	// Handle CORS methods
	if methods := viper.GetString("CORS_ALLOWED_METHODS"); methods != "" {
		viper.Set("CORS_ALLOWED_METHODS", strings.Split(methods, ","))
	}
	
	// Handle CORS headers
	if headers := viper.GetString("CORS_ALLOWED_HEADERS"); headers != "" {
		viper.Set("CORS_ALLOWED_HEADERS", strings.Split(headers, ","))
	}
	
	// Handle trusted proxies
	if proxies := viper.GetString("TRUSTED_PROXIES"); proxies != "" {
		viper.Set("TRUSTED_PROXIES", strings.Split(proxies, ","))
	}
}

// GetConfig provides a global way to access the loaded configuration.
// This is a simple approach; for larger apps, consider dependency injection.
var AppConfig Config

// InitConfig initializes the AppConfig global variable.
// Call this function once at the start of your application.
func InitConfig(path string) {
	var err error
	AppConfig, err = LoadConfig(path)
	if err != nil {
		log.Fatalf("Failed to load application configuration: %s", err)
	}
}

// IsProduction returns true if running in production environment
func IsProduction() bool {
	return AppConfig.Server.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func IsDevelopment() bool {
	return AppConfig.Server.Environment == "development"
}

// IsStaging returns true if running in staging environment
func IsStaging() bool {
	return AppConfig.Server.Environment == "staging"
}

// IsTest returns true if running in test environment
func IsTest() bool {
	return AppConfig.Server.Environment == "test"
}

	



