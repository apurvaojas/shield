package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server          ServerConfig
	Database        DatabaseConfig
	Redis           RedisConfig
	Cognito         CognitoConfig
	JWT             JWTConfig
	OPA             OPAConfig
	Observability   ObservabilityConfig
	RateLimiting    RateLimitingConfig
	Security        SecurityConfig
	Features        FeaturesConfig
	Logger          LoggerConfig
	Instrumentation InstrumentationConfig
}

// ServerConfig holds server-specific configuration.
type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
	Timezone    string `mapstructure:"timezone"`
}

// DatabaseConfig holds database connection details.
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"sslMode"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
}

// RedisConfig holds Redis connection details.
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// CognitoConfig holds AWS Cognito specific configuration.
type CognitoConfig struct {
	UserPoolID      string `mapstructure:"userPoolId"`
	AppClientID     string `mapstructure:"appClientId"`
	AppClientSecret string `mapstructure:"appClientSecret"`
	Region          string `mapstructure:"region"`
	Domain          string `mapstructure:"domain"`
}

// JWTConfig holds JWT token configuration.
type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	Expiry        time.Duration `mapstructure:"expiry"`
	RefreshExpiry time.Duration `mapstructure:"refreshExpiry"`
}

// OPAConfig holds Open Policy Agent configuration.
type OPAConfig struct {
	ServerURL  string `mapstructure:"serverUrl"`
	PolicyPath string `mapstructure:"policyPath"`
}

// ObservabilityConfig holds observability configuration.
type ObservabilityConfig struct {
	JaegerEndpoint     string `mapstructure:"jaegerEndpoint"`
	PrometheusEndpoint string `mapstructure:"prometheusEndpoint"`
	EnableMetrics      bool   `mapstructure:"enableMetrics"`
	EnableTracing      bool   `mapstructure:"enableTracing"`
}

// RateLimitingConfig holds rate limiting configuration.
type RateLimitingConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requestsPerMinute"`
	Burst             int  `mapstructure:"burst"`
}

// SecurityConfig holds security-related configuration.
type SecurityConfig struct {
	CORS           CORSConfig `mapstructure:"cors"`
	TrustedProxies []string   `mapstructure:"trustedProxies"`
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowedOrigins"`
	AllowedMethods []string `mapstructure:"allowedMethods"`
	AllowedHeaders []string `mapstructure:"allowedHeaders"`
}

// FeaturesConfig holds feature flag configuration.
type FeaturesConfig struct {
	MultiFactorAuth bool `mapstructure:"multiFactorAuth"`
	DeviceTracking  bool `mapstructure:"deviceTracking"`
	SessionRotation bool `mapstructure:"sessionRotation"`
}

// LoggerConfig holds logger configuration.
type LoggerConfig struct {
	Level         string `mapstructure:"level"`
	FileEnabled   bool   `mapstructure:"fileEnabled"`
	FileDir       string `mapstructure:"fileDir"`
	MaxFileSizeMB int    `mapstructure:"maxFileSizeMB"`
	MaxFiles      int    `mapstructure:"maxFiles"`
	MaxAgeDays    int    `mapstructure:"maxAgeDays"`
	Compress      bool   `mapstructure:"compress"`
	EnableMasking bool   `mapstructure:"enableMasking"`
}

// InstrumentationConfig holds instrumentation configuration.
type InstrumentationConfig struct {
	Logging       LoggingInstrumentationConfig `mapstructure:"logging"`
	OpenTelemetry OTELInstrumentationConfig    `mapstructure:"openTelemetry"`
}

// LoggingInstrumentationConfig holds logging instrumentation configuration.
type LoggingInstrumentationConfig struct {
	WithRequestBody    bool     `mapstructure:"withRequestBody"`
	WithResponseBody   bool     `mapstructure:"withResponseBody"`
	WithRequestHeader  bool     `mapstructure:"withRequestHeader"`
	WithResponseHeader bool     `mapstructure:"withResponseHeader"`
	WithUserAgent      bool     `mapstructure:"withUserAgent"`
	WithRequestID      bool     `mapstructure:"withRequestId"`
	WithSpanID         bool     `mapstructure:"withSpanId"`
	WithTraceID        bool     `mapstructure:"withTraceId"`
	SkipPaths          []string `mapstructure:"skipPaths"`
	DefaultLevel       string   `mapstructure:"defaultLevel"`
	ClientErrorLevel   string   `mapstructure:"clientErrorLevel"`
	ServerErrorLevel   string   `mapstructure:"serverErrorLevel"`
}

// OTELInstrumentationConfig holds OpenTelemetry instrumentation configuration.
type OTELInstrumentationConfig struct {
	ServiceName        string   `mapstructure:"serviceName"`
	EnableTracing      bool     `mapstructure:"enableTracing"`
	EnableMetrics      bool     `mapstructure:"enableMetrics"`
	WithSpanID         bool     `mapstructure:"withSpanId"`
	WithTraceID        bool     `mapstructure:"withTraceId"`
	WithUserAgent      bool     `mapstructure:"withUserAgent"`
	WithRequestBody    bool     `mapstructure:"withRequestBody"`
	WithResponseBody   bool     `mapstructure:"withResponseBody"`
	WithRequestHeader  bool     `mapstructure:"withRequestHeader"`
	WithResponseHeader bool     `mapstructure:"withResponseHeader"`
	FilterPaths        []string `mapstructure:"filterPaths"`
	FilterMethods      []string `mapstructure:"filterMethods"`
}

// Global configuration instance
var AppConfig *Config

// LoadConfig loads configuration from YAML files based on environment.
func LoadConfig() error {
	// Determine the environment
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "dev" // Default to dev
	}

	// Configure Viper to read from YAML files
	viper.SetConfigName("application-" + env)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	// Enable environment variable override
	viper.AutomaticEnv()
	// Replace dots with underscores for environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Unmarshal into the config struct
	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// Set the global config
	AppConfig = config

	log.Printf("Configuration loaded successfully for environment: %s", config.Server.Environment)
	return nil
}

// GetConfig returns the global configuration instance.
func GetConfig() *Config {
	if AppConfig == nil {
		log.Fatal("Configuration not loaded. Call LoadConfig() first.")
	}
	return AppConfig
}

// GetServerConfig returns the server configuration.
func GetServerConfig() ServerConfig {
	return GetConfig().Server
}

// GetDatabaseConfig returns the database configuration.
func GetDatabaseConfig() DatabaseConfig {
	return GetConfig().Database
}

// GetRedisConfig returns the Redis configuration.
func GetRedisConfig() RedisConfig {
	return GetConfig().Redis
}

// GetCognitoConfig returns the Cognito configuration.
func GetCognitoConfig() CognitoConfig {
	return GetConfig().Cognito
}

// GetJWTConfig returns the JWT configuration.
func GetJWTConfig() JWTConfig {
	return GetConfig().JWT
}

// GetOPAConfig returns the OPA configuration.
func GetOPAConfig() OPAConfig {
	return GetConfig().OPA
}

// GetSecurityConfig returns the security configuration.
func GetSecurityConfig() SecurityConfig {
	return GetConfig().Security
}

// GetLoggerConfig returns the logger configuration.
func GetLoggerConfig() LoggerConfig {
	return GetConfig().Logger
}

// GetInstrumentationConfig returns the instrumentation configuration.
func GetInstrumentationConfig() InstrumentationConfig {
	return GetConfig().Instrumentation
}

// Environment helpers
func IsProduction() bool {
	return GetConfig().Server.Environment == "production"
}

func IsDevelopment() bool {
	return GetConfig().Server.Environment == "development" || GetConfig().Server.Environment == "dev"
}

func IsStaging() bool {
	return GetConfig().Server.Environment == "staging"
}

// GetServerAddress returns the full server address with port.
func GetServerAddress() string {
	config := GetServerConfig()
	return fmt.Sprintf(":%d", config.Port)
}
