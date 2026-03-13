package main

import (
	log "log/slog"
	"os"
	"shield/cmd/app/config"
	"shield/cmd/app/router"
	"shield/modules/common/database"
	"shield/modules/common/swagger"
	common "shield/modules/common/telemetry/logger"
	"time"

	"gorm.io/gorm"

	authn "shield/modules/authn"

	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
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

// @host      localhost:8081
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

	// Run automigrations for authn models so required tables (users, organizations, sessions, etc.) exist.
	if db != nil {
		models := authn.GetModelsForMigration()
		if err := db.AutoMigrate(models...); err != nil {
			log.Error("Database automigrate failed", "err", err)
		} else {
			log.Info("Database automigrate completed successfully")
		}
		// Ensure default organization exists for individual users
		if org, err := authn.EnsureDefaultOrganization(db); err != nil {
			log.Error("Failed to ensure default organization", "err", err)
		} else if org != nil {
			log.Info("Default organization ensured", "id", org.ID, "name", org.Name)
		}
	}

	// Configure the API base path in swagger documentation
	swagger.ConfigureApiBasePath("/api/v1")

	// Initialize router with database connection
	routerInstance := router.InitRoutes(db)
	if routerInstance == nil {
		log.Error("Failed to initialize router")
	}

	// Start server: either as a normal HTTP server (for Fargate / local) or as AWS Lambda
	serverAddr := config.GetServerAddress()
	log.Info("Server starting", "address", serverAddr)

	// Detect Lambda environment by common env vars and start adapter if present
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" || os.Getenv("LAMBDA_TASK_ROOT") != "" {
		log.Info("Running in AWS Lambda mode - starting lambda adapter")
		adapter := ginadapter.New(routerInstance)
		// This will block and run the Lambda handler
		lambda.Start(adapter.ProxyWithContext)
		return
	}

	// Normal HTTP server (local / Fargate / Docker)
	if err := routerInstance.Run(serverAddr); err != nil {
		log.Error("Failed to start server", "err", err)
	}
}
