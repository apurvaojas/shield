package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	authn "github.com/tentackles/shield/modules/authn"
	"github.com/tentackles/shield/pkg/database"

	_ "github.com/tentackles/shield/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Shield Platform API
// @version 1.0
// @description This is the main API for the Shield Identity and Access Management Platform.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /api/v1
// @schemes http
func main() {
	log.Println("Starting Shield Platform API...")

	// Initialize configuration
	// This will load from .env file in the current directory (project root) or environment variables
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	_ = viper.ReadInConfig() // ignore error if .env does not exist, environment variables will still be used
	log.Println("Configuration loaded.")

	router := gin.Default()

	// --- Initialize AuthN Module Dependencies ---
	// Initialize Database connection
	db, err := database.NewConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection established.")

	// Auto-migrate the database schema
	models := authn.GetModelsForMigration()
	err = db.AutoMigrate(models...)
	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}
	log.Println("Database schema migrated successfully.")

	// Initialize AuthN Service with dependencies
	authnSvc := authn.NewAuthService(db)
	log.Println("AuthN Service initialized.")

	// Register AuthN routes under /api/v1/auth
	authn.RegisterAuthRoutes(router.Group("/api/v1/auth"), authnSvc)
	log.Println("AuthN API routes registered under /api/v1/auth.")

	// --- Swagger for the main application ---
	// URL: http://localhost:8081/swagger/index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Println("Swagger UI registered at /swagger/index.html")

	port := viper.GetString("SERVER_PORT")
	if port == "" {
		log.Println("Server port not configured, using default 8081")
		port = "8081"
	}

	// Use HTTPS if SSL_CERT_FILE and SSL_KEY_FILE env vars are set and files exist
	certFile := viper.GetString("SSL_CERT_FILE")
	keyFile := viper.GetString("SSL_KEY_FILE")
	if certFile != "" && keyFile != "" {
		if _, errCert := os.Stat(certFile); errCert == nil {
			if _, errKey := os.Stat(keyFile); errKey == nil {
				log.Printf("Server starting with HTTPS on port %s (cert: %s, key: %s)...", port, certFile, keyFile)
				if err := router.RunTLS(":"+port, certFile, keyFile); err != nil {
					log.Fatalf("Failed to start HTTPS server: %v", err)
				}
				return
			} else {
				log.Printf("SSL key file not found at %s, falling back to HTTP on port %s...", keyFile, port)
			}
		} else {
			log.Printf("SSL cert file not found at %s, falling back to HTTP on port %s...", certFile, port)
		}
	}

	log.Printf("Server starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
