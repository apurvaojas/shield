package main

import (
	log "log/slog"
	"shield/cmd/app/config"
	"shield/cmd/app/router"
	"shield/modules/common/database"
	common "shield/modules/common/telemetry/logger"
	"time"

	_ "shield/docs" // This line is needed for swagger

	"gorm.io/gorm"
)

// @version         1.0
// @description     API documentation for Organic Forms Configuration Management
// @description     This API provides endpoints for:
// @description     - User Authentication & Authorization
// @description     - Organization Management
// @description     - SSO Configuration
// @description     - Form Configuration Management

// @contact.name   API Support
// @contact.url    https://github.com/yourusername/shield/issues
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8001
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
func main() {
	// Initialize logger first
	if err := common.InitLogger(); err != nil {
		log.Error("Failed to initialize logger", "err", err)
	}

	log.Info("Starting Shield Platform API...")

	// Load centralized configuration
	if err := config.LoadConfig(); err != nil {
		log.Error("Failed to load configuration", "err", err)
	}

	cfg := config.GetConfig()
	log.Info("Loaded configuration for environment", "environment", cfg.Server.Environment)

	// Set timezone
	if cfg.Server.Timezone != "" {
		loc, err := time.LoadLocation(cfg.Server.Timezone)
		if err != nil {
			log.Info("Invalid timezone, using UTC", "timezone", cfg.Server.Timezone)
			loc, _ = time.LoadLocation("UTC")
		}
		time.Local = loc
	}

	// Initialize database connection (optional for demo)
	maxRetries := 3
	retryInterval := time.Second
	var db *gorm.DB
	dbConnected := false

	for i := 0; i < maxRetries; i++ {
		var err error
		db, err = database.NewConnection()
		if err != nil {
			log.Info("database connection error. Retrying...", "err", err, "retryInterval", retryInterval)
			time.Sleep(retryInterval)
			continue
		}
		log.Info("Database connected successfully")
		dbConnected = true
		break
	}

	if !dbConnected {
		log.Info("Warning: Failed to connect to database after retries. Continuing without database...")
		db = nil // Explicitly set to nil for clarity
	}

	// Initialize router with database connection
	routerInstance := router.InitRoutes(db)
	if routerInstance == nil {
		log.Error("Failed to initialize router")
	}

	// Start server
	serverAddr := config.GetServerAddress()
	log.Info("Server starting", "address", serverAddr)

	if err := routerInstance.Run(serverAddr); err != nil {
		log.Error("Failed to start server", "err", err)
	}
}
