// Project renamed from shield to shield.
// All import paths and module names have been updated accordingly.

// Package main is the entry point for the shield service
//
// @title           Org Forms Config Management API
// @version         1.0
// @description     API documentation for Organic Forms Configuration Management
// @description     This API provides endpoints for:
// @description     - User Authentication & Authorization
// @description     - Organization Management
// @description     - SSO Configuration
// @description     - Form Configuration Management
//
// @contact.name   API Support
// @contact.url    https://github.com/yourusername/shield/issues
// @contact.email  support@example.com
//
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host      localhost:8001
// @BasePath  /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token

package main

import (
	"log"
	"shield/cmd/app/routers"
	"shield/modules/common/config"
	"shield/modules/common/infra/database"
	"shield/modules/common/infra/logger"
	"time"

	_ "shield/docs" // This line is needed for swagger

	"github.com/spf13/viper"
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
	//set timezone
	viper.SetDefault("SERVER_TIMEZONE", "Asia/kolkata")
	loc, _ := time.LoadLocation(viper.GetString("SERVER_TIMEZONE"))
	time.Local = loc

	if err := config.SetupConfig(); err != nil {
		logger.Fatalf("config SetupConfig() error: %s", err)
	}
	masterDSN, replicaDSN := config.DbConfiguration()

	// Retry connection if error
	maxRetries := 5
	retryInterval := time.Second

	for i := 0; i < maxRetries; i++ {
		if err := database.DBConnection(masterDSN, replicaDSN); err != nil {
			log.Printf("database DbConnection error: %s. Retrying in %v...", err, retryInterval)
			time.Sleep(retryInterval)
			continue
		}
		break
	}

	// Initialize router with swagger docs
	router := routers.InitRoutes()

	// Start server
	if err := router.Run(config.ServerConfig()); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}

}
