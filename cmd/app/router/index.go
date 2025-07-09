package router

import (
	"shield/cmd/app/config"
	"shield/modules/authn"
	"shield/modules/common/telemetry/instrumentation"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// initAuthRoutes initializes authentication and authorization routes using the authn module
func initAuthRoutes(router gin.IRouter, db *gorm.DB) {
	// Check if database is available
	if db == nil {
		// Create placeholder routes when database is not available
		router.Group("/auth").GET("/*any", func(c *gin.Context) {
			c.JSON(503, gin.H{"error": "Database not available - authn module unavailable"})
		})
		router.Group("/org").GET("/*any", func(c *gin.Context) {
			c.JSON(503, gin.H{"error": "Database not available - authn module unavailable"})
		})
		return
	}

	// Initialize authn service with the provided database connection
	authService := authn.NewAuthService(db)

	// Register authn routes using the public API
	v1RouterGroup, ok := router.(*gin.RouterGroup)
	if !ok {
		// Fallback: create placeholder routes
		router.Group("/auth").GET("/*any", func(c *gin.Context) {
			c.JSON(500, gin.H{"error": "Router initialization failed"})
		})
		return
	}

	authn.RegisterAuthRoutes(v1RouterGroup, authService)
}

// InitRoutes initializes all modules routes
func InitRoutes(db *gorm.DB) *gin.Engine {
	cfg := config.GetConfig()

	if cfg.Server.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	allowedHosts := cfg.Security.TrustedProxies
	// InitSwagger()

	router := gin.New()

	router.SetTrustedProxies(allowedHosts)
	router.Use(instrumentation.InitLoggingMiddleware())
	router.Use(instrumentation.InitOTELMiddleware(nil))
	router.Use(gin.Recovery())
	// router.Use(middleware.CORSMiddleware())

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Initialize API routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "API is running"})
		})

		// Initialize authn module routes
		initAuthRoutes(v1, db)
	}

	return router
}
